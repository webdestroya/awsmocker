package awsmocker

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func encodeAsXml(obj interface{}) string {
	out, err := xml.MarshalIndent(obj, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(out)
}

func EncodeAsJson(obj interface{}) string {
	out, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return string(out)
}

func inferContentType(value string) string {
	switch {
	case strings.HasPrefix(value, "<"):
		return ContentTypeXML
	case strings.HasPrefix(value, "{"):
		return ContentTypeJSON
	default:
		return ContentTypeText
	}
}

func isEof(r *bufio.Reader) bool {
	_, err := r.Peek(1)
	return errors.Is(err, io.EOF)
}

/*
func copyOrWarn(dst io.Writer, src io.Reader, wg *sync.WaitGroup) {
	if _, err := io.Copy(dst, src); err != nil {
		panic(fmt.Errorf("Error copying to client: %w", err))
	}
	wg.Done()
}

func copyAndClose(dst, src halfClosable) {
	if _, err := io.Copy(dst, src); err != nil {
		panic(fmt.Errorf("Error copying to client: %w", err))
	}

	_ = dst.CloseWrite()
	_ = src.CloseRead()
}
*/

func httpError(w io.WriteCloser, srcErr error) {
	if _, err := io.WriteString(w, "HTTP/1.1 502 Bad Gateway\r\n\r\n"); err != nil {
		panic(fmt.Errorf("Error responding to client: %w", err))
	}
	if err := w.Close(); err != nil {
		panic(fmt.Errorf("Error closing client connection: %w", err))
	}
}

func httpErrorCode(w io.WriteCloser, code int) {

	errString := fmt.Sprintf("HTTP/1.1 %d %s\r\n\r\n", code, http.StatusText(code))

	if _, err := io.WriteString(w, errString); err != nil {
		panic(fmt.Errorf("Error responding to client: %w", err))
	}
	if err := w.Close(); err != nil {
		panic(fmt.Errorf("Error closing client connection: %w", err))
	}
}
