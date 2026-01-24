package headers

import (
	"bytes"
	"errors"
	"strings"
)

const crlf = "\r\n"

// RFC 7320 valid header tokens
var isTokenTable = [256]uint8{
	'!': 1, '#': 1, '$': 1, '%': 1, '&': 1, '\'': 1, '*': 1, '+': 1, '-': 1, '.': 1,
	'0': 1, '1': 1, '2': 1, '3': 1, '4': 1, '5': 1, '6': 1, '7': 1, '8': 1, '9': 1,
	'A': 1, 'B': 1, 'C': 1, 'D': 1, 'E': 1, 'F': 1, 'G': 1, 'H': 1, 'I': 1, 'J': 1,
	'K': 1, 'L': 1, 'M': 1, 'N': 1, 'O': 1, 'P': 1, 'Q': 1, 'R': 1, 'S': 1, 'T': 1,
	'U': 1, 'V': 1, 'W': 1, 'X': 1, 'Y': 1, 'Z': 1,
	'^': 1, '_': 1, '`': 1,
	'a': 1, 'b': 1, 'c': 1, 'd': 1, 'e': 1, 'f': 1, 'g': 1, 'h': 1, 'i': 1, 'j': 1,
	'k': 1, 'l': 1, 'm': 1, 'n': 1, 'o': 1, 'p': 1, 'q': 1, 'r': 1, 's': 1, 't': 1,
	'u': 1, 'v': 1, 'w': 1, 'x': 1, 'y': 1, 'z': 1,
	'|': 1, '~': 1,
}

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	bytesRead := idx + len(crlf)

	if idx == -1 {
		return 0, false, nil
	}

	if idx == 0 {
		return bytesRead, true, nil
	}

	headerText := string(data[:idx])
	key, val, found := strings.Cut(headerText, ":")

	if found == false {
		return 0, false, errors.New("Header is not valid")
	}

	if key != strings.TrimRight(key, " ") {
		return 0, false, errors.New("invalid header name, contains space before colon")
	}

	isValidKey := validateTokens(key)

	if isValidKey == false {
		return bytesRead, false, errors.New("Header key or value has invalid characters")
	}

	key = strings.TrimSpace(key)
	val = strings.TrimSpace(val)

	h.Set(key, val)

	return bytesRead, false, nil
}

func (h Headers) Set(key, val string) {
	key = strings.ToLower(key)
	existingVal, ok := h[key]

	if ok {
		val = strings.Join([]string{existingVal, val}, ", ")
	}

	h[key] = val
}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)
	val, ok := h[key]
	return val, ok
}

func validateTokens(tokens string) bool {
	for _, token := range tokens {
		if isTokenTable[token] == 0 {
			return false
		}
	}

	return true
}
