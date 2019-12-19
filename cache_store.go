package mtasts

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type fsStore struct {
	Dir string
}

func (s fsStore) List() ([]string, error) {
	info, err := ioutil.ReadDir(s.Dir)
	if err != nil {
		return nil, err
	}
	domains := make([]string, 0, len(info))
	for _, ent := range info {
		if ent.IsDir() {
			continue
		}
		domains = append(domains, ent.Name())
	}
	return domains, nil
}

func (s fsStore) Store(domain, id string, fetchTime time.Time, p *Policy) error {
	path := filepath.Join(s.Dir, domain)

	f, err := os.Create(path + ".tmp")
	if err != nil {
		return err
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(map[string]interface{}{
		"ID":        id,
		"FetchTime": fetchTime,
		"Policy":    p,
	})
	if err != nil {
		return err
	}

	return os.Rename(f.Name(), path)
}

func (s fsStore) Load(domain string) (id string, fetchTime time.Time, p *Policy, err error) {
	f, err := os.Open(filepath.Join(s.Dir, domain))
	if err != nil {
		if os.IsNotExist(err) {
			return "", time.Time{}, nil, ErrNoPolicy
		}
		return "", time.Time{}, nil, err
	}
	defer f.Close()

	data := struct {
		ID        string
		FetchTime time.Time
		Policy    *Policy
	}{}
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return "", time.Time{}, nil, err
	}
	return data.ID, data.FetchTime, data.Policy, nil
}

// NewFSCache creates the Cache object using FS directory to store cached
// policies.
//
// The specified directory should exist and be writtable.
func NewFSCache(directory string) *Cache {
	return &Cache{
		Store:    fsStore{Dir: directory},
		Resolver: net.DefaultResolver,
	}
}

type ramStore struct {
	lock sync.RWMutex
	m    map[string]struct {
		id        string
		fetchtime time.Time
		policy    *Policy
	}
}

func (s *ramStore) List() ([]string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	keys := make([]string, 0, len(s.m))
	for k := range s.m {
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *ramStore) Store(key string, id string, fetchTime time.Time, policy *Policy) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.m[key] = struct {
		id        string
		fetchtime time.Time
		policy    *Policy
	}{
		id, fetchTime, policy,
	}
	return nil
}

func (s *ramStore) Load(key string) (id string, fetchTime time.Time, policy *Policy, err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	data, ok := s.m[key]
	if !ok {
		return "", time.Time{}, nil, ErrNoPolicy
	}
	return data.id, data.fetchtime, data.policy, nil
}

func newRAMStore() *ramStore {
	return &ramStore{m: make(map[string]struct {
		id        string
		fetchtime time.Time
		policy    *Policy
	})}
}

// NewRAMCache creates the Cache object using RAM map to store cached policies.
//
// The underlying Store implementation is goroutine-safe.
func NewRAMCache() *Cache {
	return &Cache{
		Store:    newRAMStore(),
		Resolver: net.DefaultResolver,
	}
}

type nopStore struct{}

func (nopStore) List() ([]string, error) {
	return nil, nil
}

func (nopStore) Store(key string, id string, fetchTime time.Time, policy *Policy) error {
	return nil
}

func (nopStore) Load(key string) (id string, fetchTime time.Time, policy *Policy, err error) {
	return "", time.Time{}, nil, ErrNoPolicy
}

// NewNopCache creates the Cache object that never stores fetched policies and
// always repeats the lookup.
//
// It should be used only for tests, caching is criticial for the MTA-STS
// security model.
func NewNopCache() *Cache {
	return &Cache{
		Store:    nopStore{},
		Resolver: net.DefaultResolver,
	}
}
