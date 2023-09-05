package utils

import (
	"bytes"
	"io"
	"net/http"
)

func UnmarshalRequestBody(req *http.Request) string {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return ""
	}
	body = bytes.TrimSpace(body)
	req.Body = io.NopCloser(bytes.NewReader(body))

	return string(body)
}
