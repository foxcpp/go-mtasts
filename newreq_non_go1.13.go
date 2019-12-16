//+build !go1.13

package mtasts

import (
	"context"
	"io"
	"net/http"
)

func newRequestWithContext(_ context.Context, method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}
