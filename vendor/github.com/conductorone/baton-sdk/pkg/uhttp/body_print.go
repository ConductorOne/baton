package uhttp

// Implements a debugging facility for request responses. This changes
// the behavior of `BaseHttpClient` with an unexported flag.
//
// IMPORTANT: This feature is intended for development and debugging purposes only.
// Do not enable in production as it may expose sensitive information in logs.
//
// Usage:
//   client := uhttp.NewBaseHttpClient(
//     httpClient,
//     uhttp.WithPrintBody(true), // Enable response body printing
//   )

import (
	"errors"
	"fmt"
	"io"
	"os"
)

type printReader struct {
	reader io.Reader
}

func (pr *printReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		_, merr := fmt.Fprint(os.Stdout, string(p[:n]))
		if merr != nil {
			return -1, errors.Join(err, merr)
		}
	}

	return n, err
}

func wrapPrintBody(body io.Reader) io.Reader {
	return &printReader{reader: body}
}
