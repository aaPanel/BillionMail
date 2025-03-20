// zip compression module
// @author Zhj<2023-08-25>

package compress

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	// defaultZipper default zip compressor
	defaultZipper = NewZipper(-1)
)

// Zipper zip compressor
type Zipper struct {
	mutex   sync.Mutex // mutex
	quota   int64      // compressed file size limit, -1 for unlimited
	written int64      // size of already compressed file
}

// NewZipper creates a zip compressor
func NewZipper(quota int64) *Zipper {
	return &Zipper{quota: quota}
}

// SetQuota sets the compressed file size limit
func (z *Zipper) SetQuota(quota int64) {
	z.mutex.Lock()
	defer z.mutex.Unlock()
	z.quota = quota
}

// incWritten increases the size of already compressed file
func (z *Zipper) incWritten(n int64) (err error) {
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

// compressHelper compresses a file or directory
func (z *Zipper) compressHelper(zw *zip.Writer, path string, fi os.FileInfo, root string) error {
	header, err := zip.FileInfoHeader(fi)

	if err != nil {
		return err
	}

	header.Name = strings.TrimPrefix(strings.TrimPrefix(filepath.ToSlash(path), "./"), "/")
	header.Name = strings.TrimPrefix(strings.TrimPrefix(header.Name, root), "/")
	header.Method = zip.Deflate

	if header.Name == "" {
		return err
	}

	if fi.IsDir() {
		header.Name += "/"

		_, err := zw.CreateHeader(header)

		return err
	}

	w, err := zw.CreateHeader(header)

	if err != nil {
		return err
	}

	file, err := os.Open(path)

	if err != nil {
		return err
	}

	defer file.Close()

	var n int64

	n, err = io.Copy(w, file)

	if err != nil {
		return err
	}

	// increase the size of already compressed file
	return z.incWritten(n)
}

// Compress compresses a file or directory
func (z *Zipper) Compress(dst string, srcList ...string) error {
	// create file and open file handle
	fp, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer fp.Close()

	zw := zip.NewWriter(fp)

	defer zw.Close()

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

			return z.compressHelper(zw, path, fi, root)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// Decompress decompresses a zip file
func (z *Zipper) Decompress(src, dst string) error {
	// get absolute path of decompression target
	dstAbs, err := filepath.Abs(dst)

	if err != nil {
		return err
	}

	zr, err := zip.OpenReader(src)

	if err != nil {
		return err
	}

	defer zr.Close()

	for _, fz := range zr.File {
		// remove ../ from filename
		arcName := filepath.ToSlash(filepath.Clean(fz.Name))

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
		if fz.FileInfo().IsDir() {
			// create directory
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

			fzp, err := fz.Open()

			if err != nil {
				return err
			}

			defer fzp.Close()

			var n int64

			n, err = io.Copy(file, fzp)

			if err != nil {
				return err
			}

			// increase the size of already decompressed file
			err = z.incWritten(n)

			return err
		}()

		if err != nil {
			return err
		}
	}

	return nil
}

// Zip zip compression
// @author Zhj<2023-08-24>
func Zip(dst string, srcList ...string) error {
	return defaultZipper.Compress(dst, srcList...)
}

// Unzip zip decompression
// @author Zhj<2023-08-24>
func Unzip(src, dst string) error {
	return defaultZipper.Decompress(src, dst)
}
