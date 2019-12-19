package preload

import (
	"github.com/foxcpp/go-mtasts"
	"time"
)

type PreloadedCache struct {
	L     *List
	Inner mtasts.Store
}

func (pc PreloadedCache) List() ([]string, error) {
	return pc.Inner.List()
}

func (pc PreloadedCache) Store(key string, id string, fetchTime time.Time, policy *mtasts.Policy) error {
	return pc.Inner.Store(key, id, fetchTime, policy)
}

func (pc PreloadedCache) Load(key string) (string, time.Time, *mtasts.Policy, error) {
	id, fetchTime, policy, err := pc.Inner.Load(key)
	if err != nil {
		if err == mtasts.ErrNoPolicy {
			ent, ok := pc.L.Lookup(key)
			if ok {
				sts := ent.STS(pc.L)

				// Use of non-sensical policy ID will ensure it will be always
				// replaced when domain publishes an actual policy.
				return "\x00PRELOADED", time.Now(), &sts, nil
			}
		}

		return "", time.Time{}, nil, err
	}
	return id, fetchTime, policy, nil
}

// WrapCache wraps the mtasts.Store to use the preload list as a second source
// to fetch policies from.
func WrapCache(c mtasts.Store, l *List) PreloadedCache {
	return PreloadedCache{L: l, Inner: c}
}
