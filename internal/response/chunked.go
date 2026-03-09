package response

import (
	"fmt"
	"strconv"
)

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	chunkLenHex := strconv.FormatInt(int64(len(p)), 16)
	chunkToWrite := fmt.Sprintf("%s\r\n%s\r\n", chunkLenHex, p)

	return w.writer.Write([]byte(chunkToWrite))
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	chunkToWrite := "0\r\n"
	return w.writer.Write([]byte(chunkToWrite))
}
