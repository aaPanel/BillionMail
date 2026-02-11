package compress

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressBytesRoundTrip(t *testing.T) {
	z := NewGZipper(-1)

	tests := []struct {
		name  string
		input []byte
	}{
		{"simple text", []byte("hello world")},
		{"empty", []byte{}},
		{"binary data", []byte{0x00, 0xFF, 0x01, 0xFE}},
		{"large input", make([]byte, 10000)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := z.CompressBytes(tt.input)
			require.NoError(t, err)

			if len(tt.input) > 0 {
				assert.NotEmpty(t, compressed)
			}

			decompressed, err := z.DecompressBytes(compressed)
			require.NoError(t, err)
			assert.Equal(t, tt.input, decompressed)
		})
	}
}

func TestDecompressBytesInvalid(t *testing.T) {
	z := NewGZipper(-1)

	_, err := z.DecompressBytes([]byte("not gzip data"))
	assert.Error(t, err)
}

func TestPackageLevelCompressDecompress(t *testing.T) {
	data := []byte("package level round trip test")
	compressed, err := Compress(data)
	require.NoError(t, err)

	decompressed, err := Decompress(compressed)
	require.NoError(t, err)
	assert.Equal(t, data, decompressed)
}
