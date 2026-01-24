package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := min(cr.pos+cr.numBytesPerRead, len(cr.data))
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestRequestLineParse(t *testing.T) {

	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	// Test: Good GET Request line
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}

	// Test: Good GET Request line with path
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	reader = &chunkReader{
		data:            "/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}

	// Test: Invalid number of parts in request line
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	reader = &chunkReader{
		data:            "GET / HTTP/1.2\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	// Test: Invalid/Unsupported http version
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	reader = &chunkReader{
		data:            "get / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}

	// Test: valid http method 1
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	reader = &chunkReader{
		data:            "USE / HTTP/1.2\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 2,
	}

	// Test: valid http method 2
	_, err = RequestFromReader(reader)
	require.Error(t, err)
}

func TestHeaderParsing(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedHeaders map[string]string
		expectError     bool
	}{
		{
			name:  "Standard Headers",
			input: "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			expectedHeaders: map[string]string{
				"host":       "localhost:42069",
				"user-agent": "curl/7.81.0", // Fixed version number from your original code to match input
				"accept":     "*/*",
			},
			expectError: false,
		},
		{
			name:            "Empty Headers",
			input:           "GET / HTTP/1.1\r\n\r\n",
			expectedHeaders: map[string]string{}, // Should be empty, not nil
			expectError:     false,
		},
		{
			name:            "Malformed Header",
			input:           "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n", // Missing colon
			expectedHeaders: nil,
			expectError:     true,
		},
		{
			// Note: Behavior depends on implementation.
			// Simple maps overwrite (Last-Write-Wins).
			// Compliant HTTP parsers should merge with commas: "val1, val2".
			// Assuming Overwrite behavior for this test based on map[string]string usage.
			name:  "Duplicate Headers",
			input: "GET / HTTP/1.1\r\nHost: localhost\r\nX-Custom: value1\r\nX-Custom: value2\r\n\r\n",
			expectedHeaders: map[string]string{
				"host":     "localhost",
				"x-custom": "value1, value2",
			},
			expectError: false,
		},
		{
			name:  "Case Insensitive Headers",
			input: "GET / HTTP/1.1\r\nHOST: localhost\r\nconTENT-tYPe: text/plain\r\n\r\n",
			expectedHeaders: map[string]string{
				"host":         "localhost",  // Keys should be normalized (usually lowercase)
				"content-type": "text/plain", // Keys should be normalized
			},
			expectError: false,
		},
		{
			name:            "Missing End of Headers",
			input:           "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: curl/7.81.0", // No \r\n\r\n
			expectedHeaders: nil,
			expectError:     true,
		},
		{
			name:  "Whitespace Handling",
			input: "GET / HTTP/1.1\r\nHost:   localhost   \r\n\r\n",
			expectedHeaders: map[string]string{
				"host": "localhost", // Values should be trimmed
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize your chunk reader
			reader := &chunkReader{
				data:            tc.input,
				numBytesPerRead: 3, // Simulate fragmented reading
			}

			r, err := RequestFromReader(reader)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, r)

				// Iterate over expected headers and verify they exist in the result
				for k, v := range tc.expectedHeaders {
					assert.Equal(t, v, r.Headers[k], "Header mismatch for key: "+k)
				}

				// specific check for Empty Headers case
				if len(tc.expectedHeaders) == 0 {
					assert.Empty(t, r.Headers)
				}
			}
		})
	}
}
