package preload

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/openpgp"
)

func verifyPGP(key string, sig, blob io.Reader) error {
	entList, err := openpgp.ReadArmoredKeyRing(strings.NewReader(key))
	if err != nil {
		return PGPError{Err: fmt.Errorf("key read: %w", err)}
	}

	_, err = openpgp.CheckArmoredDetachedSignature(entList, blob, sig)
	if err != nil {
		return PGPError{Err: fmt.Errorf("sig check: %w", err)}
	}
	return nil
}
