//+build go1.13

package mtasts

import (
	"context"
	"io"
	"net/http"
)

func newRequestWithContext(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, url, body)
}
