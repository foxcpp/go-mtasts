package preload

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/foxcpp/go-mockdns"
	"github.com/foxcpp/go-mtasts"
)

func mockDownloadPolicy(policy *mtasts.Policy, err error) func(string) (*mtasts.Policy, error) {
	return func(string) (*mtasts.Policy, error) {
		return policy, err
	}
}

func TestPreloadedCache(t *testing.T) {
	list, err := Read(strings.NewReader(sampleList))
	if err != nil {
		t.Fatal(err)
	}

	c := mtasts.NewRAMCache()
	c.Resolver = &mockdns.Resolver{}
	c.DownloadPolicy = mockDownloadPolicy(nil, errors.New("no"))
	c.Store = WrapCache(c.Store, list)

	expectedPolicy := &mtasts.Policy{
		Mode:   mtasts.ModeTesting,
		MaxAge: int(time.Time(list.Expires).Sub(now()).Seconds()),
		MX:     []string{"*.mail.google.com"},
	}

	policy, err := c.Get(context.Background(), "gmail.com")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(policy, expectedPolicy) {
		t.Fatalf("Wrong structure output:\nWant %+v\nGot : %+v", expectedPolicy, policy)
	}
}

func TestPreloadedCache_MTASTSPresent(t *testing.T) {
	list, err := Read(strings.NewReader(sampleList))
	if err != nil {
		t.Fatal(err)
	}

	expectedPolicy := &mtasts.Policy{
		Mode:   mtasts.ModeTesting,
		MaxAge: int(time.Time(list.Expires).Sub(now()).Seconds()),
		MX:     []string{"*.override.google.com"},
	}

	c := mtasts.NewRAMCache()
	c.Resolver = &mockdns.Resolver{
		Zones: map[string]mockdns.Zone{
			"_mta-sts.gmail.com.": {
				TXT: []string{"v=STSv1;id=123"},
			},
		},
	}
	c.DownloadPolicy = mockDownloadPolicy(expectedPolicy, nil)
	c.Store = WrapCache(c.Store, list)

	policy, err := c.Get(context.Background(), "gmail.com")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(policy, expectedPolicy) {
		t.Fatalf("Wrong structure output:\nWant %+v\nGot: %+v", expectedPolicy, policy)
	}
}

func TestPreloadedCache_ListExpired(t *testing.T) {
	list, err := Read(strings.NewReader(sampleList))
	if err != nil {
		t.Fatal(err)
	}

	c := mtasts.NewRAMCache()
	c.Resolver = &mockdns.Resolver{}

	list.Expires = ListTime(now().Add(-10 * time.Second))
	c.Store = WrapCache(c.Store, list)

	_, err = c.Get(context.Background(), "gmail.com")
	if err == nil {
		t.Fatal("Expected an error, got none")
	}
}
