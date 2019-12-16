package mtasts

import (
	"strings"

	"golang.org/x/net/idna"
	"golang.org/x/text/unicode/norm"
)

// forLookup converts the domain into a canonical form suitable for table
// lookups and other comparisons.
//
// TL;DR Use this instead of strings.ToLower to prepare domain for lookups.
//
// Domains that contain invalid UTF-8 or invalid A-label
// domains are simply converted to local-case using strings.ToLower, but the
// error is also returned.
func forLookup(domain string) (string, error) {
	uDomain, err := idna.ToUnicode(domain)
	if err != nil {
		return strings.ToLower(domain), err
	}

	// Side note: strings.ToLower does not support full case-folding, so it is
	// important to apply NFC normalization first.
	uDomain = norm.NFC.String(uDomain)
	uDomain = strings.ToLower(uDomain)
	uDomain = strings.TrimSuffix(uDomain, ".")
	return uDomain, nil
}
