package preload

import (
	"io"
	"strings"

	"golang.org/x/crypto/openpgp"
)

func verifyPGP(key string, sig, blob io.Reader) error {
	entList, err := openpgp.ReadArmoredKeyRing(strings.NewReader(key))
	if err != nil {
		return PGPError{Err: err}
	}

	_, err = openpgp.CheckArmoredDetachedSignature(entList, blob, sig)
	if err != nil {
		return PGPError{Err: err}
	}
	return nil
}
