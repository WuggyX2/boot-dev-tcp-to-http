package headers

import "bytes"

const crlf = "\r\n"

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
		return 0, true, nil
	}

	return bytesRead, false, nil
}
