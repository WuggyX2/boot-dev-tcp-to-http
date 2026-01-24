package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParsing(t *testing.T) {
	// 1. Valid single header
	t.Run("Valid single header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host: localhost:42069\r\n") // Removed extra \r\n to test just the header line
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Equal(t, len(data), n)
		assert.False(t, done)
	})

	// 2. Valid single header with extra whitespace
	t.Run("Valid single header with extra whitespace", func(t *testing.T) {
		headers := NewHeaders()
		// RFC 7230 allows OWS (Optional Whitespace) around values
		// "Key:   Value   \r\n" should parse to "Value" (trimmed)
		data := []byte("Content-Length:   123   \r\n")
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		assert.Equal(t, "123", headers["content-length"])
		assert.Equal(t, len(data), n)
		assert.False(t, done)
	})

	// 3. Valid 2 headers with existing headers
	t.Run("Valid 2 headers with existing headers", func(t *testing.T) {
		headers := NewHeaders()

		// 1. Setup existing state
		initialData := []byte("X-Existing: true\r\n")
		_, _, err := headers.Parse(initialData)
		require.NoError(t, err)

		// 2. Parse chunk containing 2 new headers
		// Assuming Parse consumes one header at a time, we loop twice
		multiData := []byte("X-Header-One: 1\r\nX-Header-Two: 2\r\n")

		// Parse first line of the chunk
		n1, done1, err1 := headers.Parse(multiData)
		require.NoError(t, err1)
		assert.False(t, done1)

		// Parse second line of the chunk
		n2, done2, err2 := headers.Parse(multiData[n1:])
		require.NoError(t, err2)
		assert.False(t, done2)

		// Validate all state is preserved
		assert.Equal(t, "true", headers["x-existing"])
		assert.Equal(t, "1", headers["x-header-one"])
		assert.Equal(t, "2", headers["x-header-two"])

		// Ensure bytes consumed matches length
		assert.Equal(t, len(multiData), n1+n2)
	})

	// 4. Valid done
	t.Run("Valid done", func(t *testing.T) {
		headers := NewHeaders()
		// The empty CRLF line signals the end of the headers section
		data := []byte("\r\n")
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		assert.Equal(t, 2, n) // Should consume the \r\n
		assert.True(t, done, "Done should be true when empty line is parsed")
	})

	// 5. Invalid spacing header
	t.Run("Invalid spacing header", func(t *testing.T) {
		headers := NewHeaders()
		// Leading spaces in the Key are generally invalid
		data := []byte("       Host : localhost:42069       \r\n\r\n")
		n, done, err := headers.Parse(data)

		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Invalid char in header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("test-he@der: test-value\r\n")

		n, done, err := headers.Parse(data)
		require.Error(t, err)
		assert.Equal(t, len(data), n)
		assert.False(t, done)
	})

	t.Run("Multiple valid same headers", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Set-Person: lane-loves-go\r\n")
		_, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.False(t, done)

		data = []byte("Set-Person: prime-loves-zig\r\n")
		_, done, err = headers.Parse(data)
		require.NoError(t, err)
		assert.False(t, done)

		data = []byte("Set-Person: tj-loves-ocaml\r\n")
		_, done, err = headers.Parse(data)
		require.NoError(t, err)
		assert.False(t, done)

		assert.Equal(t, "lane-loves-go, prime-loves-zig, tj-loves-ocaml", headers["set-person"])
	})
}
