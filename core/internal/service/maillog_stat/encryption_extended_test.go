package maillog_stat

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Unicode strings
// ---------------------------------------------------------------------------

func TestEncryptDecrypt_Unicode(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"Chinese", "你好世界"},
		{"Japanese", "こんにちは世界"},
		{"Korean", "안녕하세요 세계"},
		{"Arabic", "مرحبا بالعالم"},
		{"Emoji", "Hello 🌍🎉🚀"},
		{"Mixed scripts", "Hello 你好 مرحبا こんにちは"},
		{"Accented Latin", "cafe\u0301 re\u0301sume\u0301 nai\u0308ve"},
		{"Cyrillic", "Привет мир"},
		{"Thai", "สวัสดีชาวโลก"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted := Encrypt(tt.data)
			require.NotEmpty(t, encrypted)

			var decrypted string
			err := Decrypt(encrypted, &decrypted)
			require.NoError(t, err)
			assert.Equal(t, tt.data, decrypted)
		})
	}
}

// ---------------------------------------------------------------------------
// Large payloads
// ---------------------------------------------------------------------------

func TestEncryptDecrypt_LargePayloads(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"1KB", 1024},
		{"10KB", 10 * 1024},
		{"100KB", 100 * 1024},
		{"1MB", 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := strings.Repeat("A", tt.size)
			encrypted := Encrypt(data)
			require.NotEmpty(t, encrypted)

			var decrypted string
			err := Decrypt(encrypted, &decrypted)
			require.NoError(t, err)
			assert.Equal(t, data, decrypted)
		})
	}
}

func TestEncryptDecrypt_LargeMap(t *testing.T) {
	data := make(map[string]interface{})
	for i := 0; i < 500; i++ {
		data[string(rune('a'+i%26))+strings.Repeat("x", i%10)] = i
	}

	encrypted := Encrypt(data)
	require.NotEmpty(t, encrypted)

	var decrypted map[string]interface{}
	err := Decrypt(encrypted, &decrypted)
	require.NoError(t, err)
	// verify a sampling of keys
	assert.NotEmpty(t, decrypted)
}

// ---------------------------------------------------------------------------
// Nil values
// ---------------------------------------------------------------------------

func TestEncrypt_NilValue(t *testing.T) {
	encrypted := Encrypt(nil)
	require.NotEmpty(t, encrypted, "nil should marshal to JSON null and encrypt")

	var decrypted interface{}
	err := Decrypt(encrypted, &decrypted)
	require.NoError(t, err)
	assert.Nil(t, decrypted)
}

func TestEncrypt_NilPointer(t *testing.T) {
	var ptr *string
	encrypted := Encrypt(ptr)
	require.NotEmpty(t, encrypted)

	var decrypted interface{}
	err := Decrypt(encrypted, &decrypted)
	require.NoError(t, err)
	assert.Nil(t, decrypted)
}

// ---------------------------------------------------------------------------
// Arrays / slices
// ---------------------------------------------------------------------------

func TestEncryptDecrypt_Slices(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
	}{
		{"int slice", []int{1, 2, 3, 4, 5}},
		{"string slice", []string{"alpha", "beta", "gamma"}},
		{"empty slice", []string{}},
		{"nested slice", [][]int{{1, 2}, {3, 4}}},
		{"mixed slice", []interface{}{1, "two", 3.0, true, nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted := Encrypt(tt.data)
			require.NotEmpty(t, encrypted)

			var decrypted interface{}
			err := Decrypt(encrypted, &decrypted)
			require.NoError(t, err)

			origJSON, _ := json.Marshal(tt.data)
			decJSON, _ := json.Marshal(decrypted)
			assert.JSONEq(t, string(origJSON), string(decJSON))
		})
	}
}

// ---------------------------------------------------------------------------
// Boolean values
// ---------------------------------------------------------------------------

func TestEncryptDecrypt_Booleans(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
	}{
		{"true", true},
		{"false", false},
		{"bool in map", map[string]bool{"active": true, "deleted": false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted := Encrypt(tt.data)
			require.NotEmpty(t, encrypted)

			var decrypted interface{}
			err := Decrypt(encrypted, &decrypted)
			require.NoError(t, err)

			origJSON, _ := json.Marshal(tt.data)
			decJSON, _ := json.Marshal(decrypted)
			assert.JSONEq(t, string(origJSON), string(decJSON))
		})
	}
}

// ---------------------------------------------------------------------------
// Concurrent encrypt/decrypt safety
// ---------------------------------------------------------------------------

func TestEncryptDecrypt_ConcurrentSafety(t *testing.T) {
	var wg sync.WaitGroup
	data := map[string]interface{}{
		"id":    42,
		"email": "test@example.com",
		"tags":  []string{"a", "b"},
	}

	origJSON, _ := json.Marshal(data)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			encrypted := Encrypt(data)
			assert.NotEmpty(t, encrypted)

			var decrypted map[string]interface{}
			err := Decrypt(encrypted, &decrypted)
			assert.NoError(t, err)

			decJSON, _ := json.Marshal(decrypted)
			assert.JSONEq(t, string(origJSON), string(decJSON))
		}()
	}

	wg.Wait()
}

func TestEncryptDecrypt_ConcurrentDifferentTypes(t *testing.T) {
	var wg sync.WaitGroup
	inputs := []interface{}{
		"string value",
		42,
		true,
		[]int{1, 2, 3},
		map[string]string{"k": "v"},
		nil,
		3.14,
	}

	for _, input := range inputs {
		wg.Add(1)
		go func(d interface{}) {
			defer wg.Done()
			origJSON, _ := json.Marshal(d)

			encrypted := Encrypt(d)
			assert.NotEmpty(t, encrypted)

			var decrypted interface{}
			err := Decrypt(encrypted, &decrypted)
			assert.NoError(t, err)

			decJSON, _ := json.Marshal(decrypted)
			assert.JSONEq(t, string(origJSON), string(decJSON))
		}(input)
	}

	wg.Wait()
}

// ---------------------------------------------------------------------------
// Numeric types
// ---------------------------------------------------------------------------

func TestEncryptDecrypt_NumericTypes(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
	}{
		{"zero", 0},
		{"negative int", -999},
		{"float", 3.14159},
		{"negative float", -0.001},
		{"large int", 9999999999},
		{"scientific notation float", 1.23e10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted := Encrypt(tt.data)
			require.NotEmpty(t, encrypted)

			var decrypted interface{}
			err := Decrypt(encrypted, &decrypted)
			require.NoError(t, err)

			origJSON, _ := json.Marshal(tt.data)
			decJSON, _ := json.Marshal(decrypted)
			assert.JSONEq(t, string(origJSON), string(decJSON))
		})
	}
}

// ---------------------------------------------------------------------------
// Decrypt error cases
// ---------------------------------------------------------------------------

func TestDecrypt_InvalidBase64(t *testing.T) {
	var result interface{}
	err := Decrypt("!!!invalid-base64!!!", &result)
	assert.Error(t, err)
}

func TestDecrypt_TooShort(t *testing.T) {
	var result interface{}
	// valid base64 but decodes to < 32 bytes
	err := Decrypt("AAAA", &result)
	assert.Error(t, err)
}

func TestDecrypt_EmptyString(t *testing.T) {
	var result interface{}
	err := Decrypt("", &result)
	assert.Error(t, err)
}

func TestDecrypt_CorruptedCiphertext(t *testing.T) {
	encrypted := Encrypt("test data")

	// flip some characters in the middle to corrupt ciphertext
	runes := []rune(encrypted)
	mid := len(runes) / 2
	if runes[mid] == 'A' {
		runes[mid] = 'B'
	} else {
		runes[mid] = 'A'
	}
	corrupted := string(runes)

	var result interface{}
	err := Decrypt(corrupted, &result)
	// may error on decryption or JSON unmarshal
	// we just assert it doesn't produce the original value
	if err == nil {
		origJSON, _ := json.Marshal("test data")
		decJSON, _ := json.Marshal(result)
		assert.NotEqual(t, string(origJSON), string(decJSON),
			"corrupted ciphertext should not produce original data")
	}
}

// ---------------------------------------------------------------------------
// Encrypt produces unique outputs (random key/iv)
// ---------------------------------------------------------------------------

func TestEncrypt_UniqueOutputs(t *testing.T) {
	data := "same input"
	seen := make(map[string]bool)

	for i := 0; i < 20; i++ {
		enc := Encrypt(data)
		seen[enc] = true
	}

	// random key/iv -> each encryption should produce different ciphertext
	assert.Greater(t, len(seen), 1,
		"multiple encryptions of same data should produce different ciphertexts")
}

// ---------------------------------------------------------------------------
// PKCS7 Padding / Unpadding
// ---------------------------------------------------------------------------

func TestPKCS7Padding(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		blockSize int
		wantLen   int // expected length after padding
	}{
		{"exact block", []byte("1234567890123456"), 16, 32}, // full block of padding added
		{"one short", []byte("123456789012345"), 16, 16},
		{"empty", []byte{}, 16, 16},
		{"single byte", []byte{0x41}, 16, 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			padded := PKCS7Padding(tt.data, tt.blockSize)
			assert.Len(t, padded, tt.wantLen)
			// padding byte value should equal number of padding bytes added
			padByte := padded[len(padded)-1]
			assert.Equal(t, tt.wantLen-len(tt.data), int(padByte))
		})
	}
}

func TestPKCS7Unpadding(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		original := []byte("Hello, World!")
		padded := PKCS7Padding(original, 16)
		unpadded := PKCS7UnPadding(padded, 16)
		assert.Equal(t, original, unpadded)
	})

	t.Run("full block padding round trip", func(t *testing.T) {
		original := []byte("1234567890123456") // exactly 16 bytes
		padded := PKCS7Padding(original, 16)
		assert.Len(t, padded, 32) // extra full block
		unpadded := PKCS7UnPadding(padded, 16)
		assert.Equal(t, original, unpadded)
	})
}

// ---------------------------------------------------------------------------
// Struct encrypt/decrypt
// ---------------------------------------------------------------------------

func TestEncryptDecrypt_Struct(t *testing.T) {
	type TrackingEvent struct {
		CampaignID int    `json:"campaign_id"`
		Email      string `json:"email"`
		Action     string `json:"action"`
	}

	original := TrackingEvent{
		CampaignID: 123,
		Email:      "user@example.com",
		Action:     "open",
	}

	encrypted := Encrypt(original)
	require.NotEmpty(t, encrypted)

	var decrypted TrackingEvent
	err := Decrypt(encrypted, &decrypted)
	require.NoError(t, err)
	assert.Equal(t, original, decrypted)
}

// ---------------------------------------------------------------------------
// Deeply nested structures
// ---------------------------------------------------------------------------

func TestEncryptDecrypt_DeepNesting(t *testing.T) {
	data := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"value": "deep",
				},
			},
		},
	}

	encrypted := Encrypt(data)
	require.NotEmpty(t, encrypted)

	var decrypted map[string]interface{}
	err := Decrypt(encrypted, &decrypted)
	require.NoError(t, err)

	origJSON, _ := json.Marshal(data)
	decJSON, _ := json.Marshal(decrypted)
	assert.JSONEq(t, string(origJSON), string(decJSON))
}

// ---------------------------------------------------------------------------
// Special string values
// ---------------------------------------------------------------------------

func TestEncryptDecrypt_SpecialStrings(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"empty string", ""},
		{"whitespace only", "   \t\n  "},
		{"JSON-like string", `{"key": "value"}`},
		{"HTML string", `<script>alert("xss")</script>`},
		{"null bytes", "before\x00after"},
		{"backslashes", `path\to\file`},
		{"quotes", `she said "hello"`},
		{"newlines", "line1\nline2\nline3"},
		{"URL", "https://example.com/path?q=1&r=2#anchor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted := Encrypt(tt.data)
			require.NotEmpty(t, encrypted)

			var decrypted string
			err := Decrypt(encrypted, &decrypted)
			require.NoError(t, err)
			assert.Equal(t, tt.data, decrypted)
		})
	}
}
