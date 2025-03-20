// tar+gzip compression module
// @author Zhj<2023-08-25>

package compress

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	// defaultGZipper default tar+gzip compressor
	defaultGZipper = NewGZipper(-1)
)

// GZipper tar+gzip compressor
type GZipper struct {
	mutex   sync.Mutex // mutex
	quota   int64      // compressed file size limit, -1 for unlimited
	written int64      // size of already compressed file
}

// NewGZipper creates a tar+gzip compressor
func NewGZipper(quota int64) *GZipper {
	return &GZipper{quota: quota}
}

// SetQuota sets the compressed file size limit
func (z *GZipper) SetQuota(quota int64) {
	z.mutex.Lock()
	defer z.mutex.Unlock()
	z.quota = quota
}

// incWritten increases the size of already compressed file
func (z *GZipper) incWritten(n int64) (err error) {
	// if no limit is set, do nothing
	if z.quota < 0 {
		return nil
	}

	z.mutex.Lock()
	defer z.mutex.Unlock()

	z.written += n

	// check if exceeds the limit
	if z.quota > -1 && z.written > z.quota {
		err = errors.New("disk quota exceeded")
		return
	}

	return nil
}

// compressHelper packs a file or directory into tar
func (z *GZipper) compressHelper(tw *tar.Writer, path string, fi os.FileInfo, root string) error {
	header, err := tar.FileInfoHeader(fi, "")

	if err != nil {
		return err
	}

	// compatible with Linux path rules
	header.Name = strings.TrimPrefix(strings.TrimPrefix(filepath.ToSlash(path), "./"), "/")
	header.Name = strings.TrimPrefix(strings.TrimPrefix(header.Name, root), "/")

	if header.Name == "" {
		return err
	}

	if fi.IsDir() {
		header.Name += "/"

		err = tw.WriteHeader(header)

		return err
	}

	file, err := os.Open(path)

	if err != nil {
		return err
	}

	defer file.Close()

	err = tw.WriteHeader(header)

	if err != nil {
		return err
	}

	var n int64

	n, err = io.Copy(tw, file)

	if err != nil {
		return err
	}

	return z.incWritten(n)
}

// decompressHelper decompresses a file or directory
func (z *GZipper) decompressHelper(zr *gzip.Reader, dst string) error {
	// get absolute path of decompression target
	dstAbs, err := filepath.Abs(dst)

	if err != nil {
		return err
	}

	tr := tar.NewReader(zr)

	for {
		header, err := tr.Next()

		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// remove ../ from filename
		arcName := filepath.ToSlash(filepath.Clean(header.Name))

		if strings.Contains(arcName, "../") {
			arcName = strings.Replace(arcName, "../", "", -1)
		}

		filename := filepath.Join(dst, arcName)

		// get absolute path of the file
		filenameAbs, err := filepath.Abs(filename)

		if err != nil {
			return err
		}

		// check if the file is under the decompression target path
		if !strings.HasPrefix(filenameAbs, dstAbs) {
			return errors.New("illegal file path: " + filename)
		}

		// check if it's a directory
		// if it's a directory, create it and skip
		if header.FileInfo().IsDir() {
			err = os.MkdirAll(filename, 0755)

			if err != nil {
				return err
			}

			continue
		}

		// create directory
		err = os.MkdirAll(filepath.Dir(filename), 0755)

		if err != nil {
			return err
		}

		file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)

		if err != nil {
			return err
		}

		err = func() error {
			defer file.Close()

			var n int64

			n, err = io.Copy(file, tr)

			if err != nil {
				return err
			}

			return z.incWritten(n)
		}()

		if err != nil {
			return err
		}
	}

	return nil
}

// Compress tar+gzip compression
func (z *GZipper) Compress(dst string, srcList ...string) error {
	// create file and open file handle
	fp, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer fp.Close()

	zw := gzip.NewWriter(fp)

	defer zw.Close()

	tw := tar.NewWriter(zw)

	defer tw.Close()

	for _, src := range srcList {
		// read source file attributes
		srcFi, err := os.Stat(src)

		if err != nil {
			return err
		}

		root := src

		if !srcFi.IsDir() {
			root = filepath.Dir(root)
		}

		root = strings.TrimPrefix(strings.TrimPrefix(filepath.ToSlash(root), "./"), "/")

		err = filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			return z.compressHelper(tw, path, fi, root)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// Decompress tar+gzip decompression
func (z *GZipper) Decompress(src, dst string) error {
	fp, err := os.Open(src)

	if err != nil {
		return err
	}

	defer fp.Close()

	zr, err := gzip.NewReader(fp)

	if err != nil {
		return err
	}

	defer zr.Close()

	return z.decompressHelper(zr, dst)
}

// DecompressReader decompresses file stream and outputs to specified file
func (z *GZipper) DecompressReader(reader io.Reader, dst string) error {
	zr, err := gzip.NewReader(reader)

	if err != nil {
		return err
	}

	defer zr.Close()

	fp, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer fp.Close()

	_, err = io.Copy(fp, zr)

	return err
}

// DecompressContent decompresses file content
func (z *GZipper) DecompressContent(data []byte, dst string) error {
	return z.DecompressReader(bytes.NewBuffer(data), dst)
}

// CompressBytes tar+gzip compresses byte stream
func (z *GZipper) CompressBytes(data []byte) (bs []byte, err error) {
	buf := &bytes.Buffer{}

	func() {
		zw := gzip.NewWriter(buf)

		defer zw.Close()

		if _, err = zw.Write(data); err != nil {
			return
		}
	}()

	return buf.Bytes(), err
}

// DecompressBytes tar+gzip decompresses byte stream
func (z *GZipper) DecompressBytes(data []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewBuffer(data))

	if err != nil {
		return []byte{}, err
	}

	defer zr.Close()

	return io.ReadAll(zr)
}

// Gzip tar+gzip compression
// @author Zhj<2023-08-24>
func Gzip(dst string, srcList ...string) error {
	return defaultGZipper.Compress(dst, srcList...)
}

// Ungzip tar+gzip decompression
// @author Zhj<2023-08-24>
func Ungzip(src, dst string) error {
	return defaultGZipper.Decompress(src, dst)
}

// UngzipContent decompresses file content (internal function)
func UngzipContent(data []byte, dst string) error {
	return defaultGZipper.DecompressContent(data, dst)
}

// UngzipReader decompresses file stream (internal function)
func UngzipReader(reader io.Reader, dst string) error {
	return defaultGZipper.DecompressReader(reader, dst)
}

// Compress compresses data using gzip
// @author Zhj<2024-01-04>
func Compress(data []byte) (bs []byte, err error) {
	return defaultGZipper.CompressBytes(data)
}

// Decompress decompresses data using gzip
// @author Zhj<2024-01-04>
func Decompress(data []byte) ([]byte, error) {
	return defaultGZipper.DecompressBytes(data)
}
