// Package preload implements parsing, updating and lookups for EFF STARTTLS
// Everywhere preload list. It can be used to prime MTA-STS cache with useful
// data to decrease the chance of downgrade attacks being possible.
package preload

import (
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/foxcpp/go-mtasts"
	"golang.org/x/net/idna"
)

var now = time.Now

type ListTime time.Time

func (t *ListTime) MarshalJSON() ([]byte, error) {
	return json.Marshal((*time.Time)(t).Format("2019-12-26T16:04:22.974364-08:00"))
}

func (t *ListTime) UnmarshalJSON(b []byte) error {
	s := ""
	if err := json.Unmarshal(b, &s); err != nil {
		// RULES.md says there can be either Unix timestamp or fixed-length
		// string format.
		i := int64(1)
		err := json.Unmarshal(b, &i)
		if err != nil {
			return err
		}
		*t = ListTime(time.Unix(i, 0).UTC())
		return nil
	}

	timeVal, err := time.ParseInLocation("2006-01-02T15:04:05.000000-07:00", s, time.UTC)
	if err != nil {
		return err
	}

	*t = ListTime(timeVal)
	return nil
}

func (t ListTime) String() string {
	return time.Time(t).String()
}

type Entry struct {
	// Set to the normalized domain name by LookupEntry.
	Domain string `json:"-"`

	PolicyAlias string      `json:"policy-alias"`
	Mode        mtasts.Mode `json:"mode"`
	MXs         []string    `json:"mxs"`
}

type List struct {
	Timestamp     ListTime         `json:"timestamp"`
	Author        string           `json:"author"`
	Version       string           `json:"version"`
	Expires       ListTime         `json:"expires"`
	PolicyAliases map[string]Entry `json:"policy-aliases"`
	Policies      map[string]Entry `json:"policies"`
}

func LoadList(r io.Reader) (*List, error) {
	l := List{}
	err := json.NewDecoder(r).Decode(&l)
	return &l, err
}

// LookupEntry extracts the corresponding entry from the list.
//
// Ths specified domain is case-folded and converted to A-labels form before
// lookup.
//
// PolicyAlias field is always empty, LookupEntry resolves aliases. If there is
// no such alias - ok = false is returned without an entry.
func (l *List) LookupEntry(domain string) (e Entry, ok bool) {
	// STARTTLS List spec does not specify the behavior in regards of IDNA
	// domains. As a sanity check, we refuse to lookup non-IDNA conforming
	// domains and convert them to A-labels form as it is consistent with
	// MTA-STS.
	//
	// https://github.com/EFForg/starttls-everywhere/issues/156
	domainACE, err := idna.ToASCII(domain)
	if err != nil {
		return Entry{}, false
	}
	domainACE = strings.ToLower(domainACE)

	e, ok = l.Policies[domainACE]
	if !ok {
		return Entry{}, false
	}

	if e.PolicyAlias != "" {
		e, ok = l.PolicyAliases[e.PolicyAlias]
		if !ok {
			return Entry{}, false
		}
	}

	e.Domain = domainACE

	return e, true
}

// STS converts the Entry into the equivalent MTA-STS policy.
func (e *Entry) STS(l *List) mtasts.Policy {
	policy := mtasts.Policy{
		Mode: e.Mode,

		// Set MaxAge so that policy will expire when the list expires.
		MaxAge: int(now().Sub(time.Time(l.Expires)).Seconds()),

		MX: make([]string, 0, len(e.MXs)),
	}

	for _, mx := range e.MXs {
		if strings.HasPrefix(mx, ".") {
			mx = "*" + mx
		}

		// https://github.com/EFForg/starttls-everywhere/issues/156
		mxACE, err := idna.ToASCII(mx)
		if err == nil {
			mx = mxACE
		}

		policy.MX = append(policy.MX, mx)
	}

	return policy
}
