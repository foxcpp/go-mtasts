package mtasts

import (
	"context"
	"errors"
	"mime"
	"net"
	"net/http"
	"runtime/trace"
	"time"
)

var httpClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return errors.New("mtasts: HTTP redirects are forbidden")
	},
	Timeout: time.Minute,
}

func downloadPolicy(ctx context.Context, domain string) (*Policy, error) {
	// TODO: Consult OCSP/CRL to detect revoked certificates?

	req, err := newRequestWithContext(ctx, "GET", "https://mta-sts."+domain+"/.well-known/mta-sts.txt", nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Policies fetched via HTTPS are only valid if the HTTP response code is
	// 200 (OK).  HTTP 3xx redirects MUST NOT be followed.
	if resp.StatusCode != 200 {
		return nil, errors.New("mtasts: HTTP " + resp.Status)
	}

	contentType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	if contentType != "text/plain" {
		return nil, errors.New("mtasts: unexpected content type")
	}

	return readPolicy(resp.Body)
}

type Resolver interface {
	LookupTXT(ctx context.Context, domain string) ([]string, error)
}

type Store interface {
	// List method is used by Cache.Refresh to clean policy data.
	List() ([]string, error)

	// Store method is used by Cache to store policy data.
	Store(key string, id string, fetchTime time.Time, policy *Policy) error

	// Load method is used by Cache to load policy data previously stored
	// using Store.
	//
	// If there is no cached policy, Load should return ErrNoPolicy.
	Load(key string) (id string, fetchTime time.Time, policy *Policy, err error)
}

// Cache structure implements transparent MTA-STS policy caching using provided
// Store implementation.
//
// It is the only way to fetch policies as caching is important to prevent
// downgrade attacks.
//
// goroutine-safety is solely defined by safety of the underlying Store and
// Resolver objects.
type Cache struct {
	Store    Store
	Resolver Resolver

	// If non-nil replaces the function used to download policy texts.
	DownloadPolicy func(domain string) (*Policy, error)
}

func IsNoPolicy(err error) bool {
	return err == ErrNoPolicy
}

// ErrNoPolicy indicates that remote domain does not offer a MTA-STS policy or
// it was ignored due to errors.
//
// Callers should not check for this directly and use IsNoPolicy
// function to decide actual handling strategy.
var ErrNoPolicy = errors.New("mtasts: no policy")

// Get reads policy from cache or tries to fetch it from Policy Host.
//
// The domain is assumed to be normalized, as done by dns.ForLookup.
func (c *Cache) Get(ctx context.Context, domain string) (*Policy, error) {
	_, p, err := c.fetch(ctx, false, time.Now(), domain)
	return p, err
}

func (c *Cache) Refresh() error {
	refreshCtx, refreshTask := trace.NewTask(context.Background(), "mtasts.Cache/Refresh")
	defer refreshTask.End()

	list, err := c.Store.List()
	if err != nil {
		return err
	}

	for _, ent := range list {
		// If policy is going to expire in next 6 hours (half of our refresh
		// period) - we still want to refresh it.
		// Since otherwise we are going to have expired policy for another 6 hours,
		// which makes it useless.
		// See https://tools.ietf.org/html/rfc8461#section-10.2.
		_, _, _ = c.fetch(refreshCtx, false, time.Now().Add(6*time.Hour), ent)

		// TODO: figure out how to clean stale entires from cache
		// and if this is really necessary.
	}

	return nil
}

func (c *Cache) fetch(ctx context.Context, ignoreDns bool, now time.Time, domain string) (cacheHit bool, p *Policy, err error) {
	defer trace.StartRegion(ctx, "mtasts.Cache/fetch").End()

	validCache := true
	cachedId, fetchTime, cachedPolicy, err := c.Store.Load(domain)
	if err != nil {
		validCache = false
	} else if fetchTime.Add(time.Duration(cachedPolicy.MaxAge) * time.Second).Before(now) {
		validCache = false
	}

	var dnsId string
	if !ignoreDns {
		records, err := c.Resolver.LookupTXT(ctx, "_mta-sts."+domain)
		if err != nil {
			if validCache {
				return true, cachedPolicy, nil
			}

			if derr, ok := err.(*net.DNSError); ok && !derr.IsTemporary {
				return false, nil, ErrNoPolicy
			}
			return false, nil, err
		}

		// RFC says:
		//   If the number of resulting records is not one, or if the resulting
		//   record is syntactically invalid, senders MUST assume the recipient
		//   domain does not have an available MTA-STS Policy. ...
		//   (Note that the absence of a usable TXT record is not by itself
		//   sufficient to remove a sender's previously cached policy for the Policy
		//   Domain, as discussed in Section 5.1, "Policy Application Control Flow".)
		if len(records) != 1 {
			if validCache {
				return true, cachedPolicy, nil
			}
			return false, nil, ErrNoPolicy
		}
		dnsId, err = readDNSRecord(records[0])
		if err != nil {
			if validCache {
				return true, cachedPolicy, nil
			}
			return false, nil, ErrNoPolicy
		}
	}

	if !validCache || dnsId != cachedId {
		var (
			policy *Policy
			err    error
		)
		if c.DownloadPolicy != nil {
			policy, err = c.DownloadPolicy(domain)
		} else {
			policy, err = downloadPolicy(ctx, domain)
		}
		if err != nil {
			if validCache {
				return true, cachedPolicy, nil
			}
			return false, nil, ErrNoPolicy
		}

		if err := c.Store.Store(domain, dnsId, time.Now(), policy); err != nil {
			// We still got up-to-date policy, cache is not critcial.
			return false, cachedPolicy, nil
		}
		return false, policy, nil
	}

	return true, cachedPolicy, nil
}
