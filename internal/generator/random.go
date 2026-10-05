package generator

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
)

const (
	alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	alphaUpperNum = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	alphaLowerNum = "abcdefghijklmnopqrstuvwxyz0123456789"
	base64url    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
	base64std    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/="
)



// RandomString returns a random string of length n chosen from charset.
func RandomString(charset string, n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	charsetLen := big.NewInt(int64(len(charset)))
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			// Fallback in catastrophic rand error
			b[i] = charset[i%len(charset)]
			continue
		}
		b[i] = charset[num.Int64()]
	}
	return string(b)
}

// RandomAlphanumeric returns an alphanumeric string of length n.
func RandomAlphanumeric(n int) string {
	return RandomString(alphanumeric, n)
}

// RandomUpperAlphanumeric returns an uppercase alphanumeric string of length n.
func RandomUpperAlphanumeric(n int) string {
	return RandomString(alphaUpperNum, n)
}

// RandomHex returns a hex-encoded string of specified byte length (result is 2*bytesLen chars).
func RandomHex(bytesLen int) string {
	b := make([]byte, bytesLen)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// RandomBase64URL returns a Base64URL-safe string of length n.
func RandomBase64URL(n int) string {
	return RandomString(base64url, n)
}

// RandomInt returns a random integer in [0, maxVal).
func RandomInt(maxVal int) int {
	if maxVal <= 0 {
		return 0
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(maxVal)))
	if err != nil {
		return 0
	}
	return int(n.Int64())
}

// RandomChoice returns a random element from a string slice.
func RandomChoice(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[RandomInt(len(items))]
}
