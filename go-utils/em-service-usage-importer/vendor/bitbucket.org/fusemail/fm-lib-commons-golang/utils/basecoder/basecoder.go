// Package basecoder provides encode and decode as per alphabet, including base62 via New62 func.
package basecoder

import (
	"math"
	"strings"

	log "github.com/sirupsen/logrus"
)

// BaseCoder provides (en|de)coding based on Alphabet.
type BaseCoder struct {
	Alphabet string
}

// New62 constructs BaseCoder instance for base62.
func New62() *BaseCoder {
	return &BaseCoder{
		Alphabet: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
	}
}

// IsBase62 returns true if input 'str' is composed by base62 characters;
// returns false otherwise (including empty string).
func IsBase62(str string) bool {
	if len(str) == 0 {
		return false
	}
	for _, c := range str {
		if (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			continue
		}
		return false
	}
	return true
}

// Encode encodes int to string as per alphabet.
func (b *BaseCoder) Encode(number int) string {
	base := len(b.Alphabet)
	log.WithFields(log.Fields{"base": base, "number": number}).Debug("encode")

	if base == 0 {
		return ""
	}

	if number < 1 {
		return string(b.Alphabet[0])
	}

	result := []byte{}

	for number > 0 {
		remainder := number % base
		number /= base
		result = append(result, b.Alphabet[remainder])
	}

	// Reverse result.
	for i := 0; i < len(result)/2; i++ {
		j := len(result) - i - 1
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

// Decode decodes string to int as per alphabet.
func (b *BaseCoder) Decode(str string) int {
	base := len(b.Alphabet)
	log.WithFields(log.Fields{"base": base, "str": str}).Debug("decode")

	if base == 0 || len(str) == 0 {
		return 0
	}

	result := 0

	for i, char := range str {
		idx := strings.IndexByte(b.Alphabet, byte(char))
		if idx < 0 {
			return 0
		}

		exp := len(str) - i - 1
		plus := idx * int(math.Pow(float64(base), float64(exp)))
		result += plus
	}

	return result
}
