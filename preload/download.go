package preload

import (
	"bytes"
	"fmt"
	"net/http"
)

type Source struct {
	// HTTP(S) URI of the list blob.
	ListURI string

	// HTTP(S) URI of the ASCII-armored PGP signature taht is valid for data fetched
	// from ListURI.
	SigURI string

	// ASCII-armored PGP key to use when verifying the signature fetched from
	// SigURI.
	SigKey string
}

// PGPError is returned when Download fails due to the problem with PGP
// signature verification.
type PGPError struct {
	Err error
}

func (err PGPError) Error() string {
	return "mtasts: cannot verify the PGP signature: " + err.Err.Error()
}

func (err PGPError) Unwrap() error {
	return err.Err
}

// Download downloads the list and verifies the PGP signature for it using
// source URIs provided in the Source structure.
//
// SigURI can be set to an empty string to disable PGP verification.
func Download(h *http.Client, s Source) (*List, error) {
	resp, err := h.Get(s.ListURI)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("mtasts: unexpected HTTP status: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	if resp.Header.Get("Content-Type") != "application/json" {
		return nil, fmt.Errorf("mtasts: unexpected Content-Type: %s", resp.Header.Get("Content-Type"))
	}

	// Dump the body into RAM, we need it multiple times.
	buf := bytes.NewBuffer(make([]byte, 0, resp.ContentLength))
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}

	if s.SigURI != "" {
		if s.SigKey == "" {
			return nil, PGPError{Err: fmt.Errorf("empty SigKey")}
		}

		sigResp, err := h.Get(s.SigURI)
		if err != nil {
			return nil, PGPError{Err: err}
		}
		if sigResp.StatusCode != 200 {
			return nil, PGPError{Err: fmt.Errorf("unexpected HTTP status: %d", resp.StatusCode)}
		}
		defer sigResp.Body.Close()

		if err := verifyPGP(s.SigKey, sigResp.Body, bytes.NewReader(buf.Bytes())); err != nil {
			return nil, err
		}
	}
	return Read(bytes.NewReader(buf.Bytes()))
}
