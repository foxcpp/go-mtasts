//+build integration

package preload

import (
	"net/http"
	"testing"
	"time"
)

// This test checks whether go-mtasts/preload can properly consume the actual
// EFF list as it is deployed.

func TestEFFDownload(t *testing.T) {
	oldNow := now
	now = time.Now
	defer func() {
		now = oldNow
	}()

	list, err := Download(http.DefaultClient, STARTTLSEverywhere)
	if err != nil {
		t.Fatal(err)
	}

	ent, ok := list.Lookup("protonmail.com")
	if !ok {
		t.Fatal("No entry for protonmail.com?")
	}

	t.Logf("List entry: %+v", ent)
	sts := ent.STS(list)
	t.Logf("MTA-STS form: %+v", sts)

	if !sts.Match("mail.protonmail.ch") {
		t.Fatal("Main MX does not match?")
	}
}
