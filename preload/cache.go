package preload

import (
	"errors"
	"sync"
	"time"

	"github.com/foxcpp/go-mtasts"
)

type PreloadedCache struct {
	lLock sync.RWMutex
	l     *List
	inner mtasts.Store
}

func (pc *PreloadedCache) List() ([]string, error) {
	return pc.inner.List()
}

func (pc *PreloadedCache) Store(key string, id string, fetchTime time.Time, policy *mtasts.Policy) error {
	return pc.inner.Store(key, id, fetchTime, policy)
}

// Update replaces the List object used by PreloadedCache in the
// goroutine-safe way.
//
// Additionally, it implements downgrade protection by returning an error when
// the current list is newer than newList or when the newList is already
// expired.
func (pc *PreloadedCache) Update(newList *List) error {
	if newList.Expired() {
		return errors.New("mtasts/preload: the new list is expired")
	}

	pc.lLock.Lock()
	defer pc.lLock.Unlock()

	if time.Time(newList.Timestamp).Before(time.Time(pc.l.Timestamp)) {
		return errors.New("mtasts/preload: the new list is older than the currently used one")
	}

	pc.l = newList

	return nil
}

func (pc *PreloadedCache) Load(key string) (string, time.Time, *mtasts.Policy, error) {
	id, fetchTime, policy, err := pc.inner.Load(key)
	if err == nil {
		return id, fetchTime, policy, nil
	}
	if err != mtasts.ErrNoPolicy {
		return "", time.Time{}, nil, err
	}
	pc.lLock.RLock()
	defer pc.lLock.RUnlock()

	if pc.l.Expired() {
		return "", time.Time{}, nil, mtasts.ErrNoPolicy
	}

	ent, ok := pc.l.Lookup(key)
	if !ok {
		return "", time.Time{}, nil, mtasts.ErrNoPolicy
	}

	sts := ent.STS(pc.l)

	// Use of non-sensical policy ID will ensure it will be always
	// replaced when domain publishes an actual policy.
	return "\x00PRELOADED", time.Now(), &sts, nil
}

// WrapCache wraps the mtasts.Store to use the preload list as a second source
// to fetch policies from.
func WrapCache(c mtasts.Store, l *List) *PreloadedCache {
	return &PreloadedCache{l: l, inner: c}
}
