package preload

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/foxcpp/go-mtasts"
)

// From https://github.com/EFForg/starttls-everywhere/blob/master/RULES.md
const sampleList = `{
  "timestamp": "2014-06-06T14:30:16.000000+00:00",
  "author": "Electronic Frontier Foundation https://eff.org",
  "expires": "2014-06-06T15:30:16.000000+00:00",
  "version": "0.1",
  "policy-aliases": {
    "gmail": {
      "mode": "testing",
      "mxs": [".mail.google.com"]
    }
  },
  "policies": {
    "yahoo.com": {
      "mode": "enforce",
      "mxs": [".yahoodns.net"]
     },
    "eff.org": {
      "mode": "enforce",
      "mxs": [".eff.org"]
    },
    "gmail.com": {
      "policy-alias": "gmail"
    },
    "example.com": {
      "mode": "testing",
      "mxs": ["mail.example.com", ".example.net"]
    }
  }
}`

var sampleListParsed = &List{
	Timestamp: ListTime(time.Date(2014, time.June, 6, 14, 30, 16, 0, time.UTC)),
	Author:    "Electronic Frontier Foundation https://eff.org",
	Expires:   ListTime(time.Date(2014, time.June, 6, 15, 30, 16, 0, time.UTC)),
	Version:   "0.1",
	PolicyAliases: map[string]Entry{
		"gmail": {
			Mode: mtasts.ModeTesting,
			MXs:  []string{".mail.google.com"},
		},
	},
	Policies: map[string]Entry{
		"yahoo.com": {
			Mode: mtasts.ModeEnforce,
			MXs:  []string{".yahoodns.net"},
		},
		"eff.org": {
			Mode: mtasts.ModeEnforce,
			MXs:  []string{".eff.org"},
		},
		"gmail.com": {
			PolicyAlias: "gmail",
		},
		"example.com": {
			Mode: mtasts.ModeTesting,
			MXs:  []string{"mail.example.com", ".example.net"},
		},
	},
}

func TestLoadList(t *testing.T) {
	l, err := LoadList(strings.NewReader(sampleList))
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(l, sampleListParsed) {
		t.Fatalf("Wrong structure output:\nWant %+v\nGot : %+v", sampleListParsed, l)
	}
}

func TestList_LookupEntry(t *testing.T) {
	l, err := LoadList(strings.NewReader(sampleList))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("trivial", func(t *testing.T) {
		ent, ok := l.LookupEntry("example.com")
		if !ok {
			t.Fatal("Entry not found but should")
		}

		entExpected := Entry{
			Domain: "example.com",
			Mode:   mtasts.ModeTesting,
			MXs:    []string{"mail.example.com", ".example.net"},
		}
		if !reflect.DeepEqual(ent, entExpected) {
			t.Fatalf("Wrong structure output:\nWant %+v\nGot : %+v", sampleListParsed, l)
		}
	})
	t.Run("aliased", func(t *testing.T) {
		ent, ok := l.LookupEntry("gmail.com")
		if !ok {
			t.Fatal("Entry not found but should")
		}

		entExpected := Entry{
			Domain: "gmail.com",
			Mode:   mtasts.ModeTesting,
			MXs:    []string{".mail.google.com"},
		}
		if !reflect.DeepEqual(ent, entExpected) {
			t.Fatalf("Wrong structure output:\nWant %+v\nGot : %+v", sampleListParsed, l)
		}
	})
	t.Run("non-existent", func(t *testing.T) {
		_, ok := l.LookupEntry("non-existent.com")
		if ok {
			t.Fatal("Entry found but should not")
		}
	})
	t.Run("non-existent alias", func(t *testing.T) {
		l.Policies["gmail.com"] = Entry{PolicyAlias: "wtf"}
		_, ok := l.LookupEntry("gmail.com")
		if ok {
			t.Fatal("Entry found but should not")
		}
	})
}

func TestEntry_STS(t *testing.T) {
	l, err := LoadList(strings.NewReader(sampleList))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("trivial", func(t *testing.T) {
		ent, ok := l.LookupEntry("example.com")
		if !ok {
			t.Fatal("Entry not found but should")
		}

		sts := ent.STS(l)
		stsExpected := mtasts.Policy{
			Mode:   mtasts.ModeTesting,
			MaxAge: int(now().Sub(time.Time(l.Expires)).Seconds()),
			MX:     []string{"mail.example.com", "*.example.net"},
		}
		if !reflect.DeepEqual(sts, stsExpected) {
			t.Fatalf("Wrong structure output:\nWant %+v\nGot : %+v", sampleListParsed, l)
		}
	})
}

func TestLoadList_UnixTimestamp(t *testing.T) {
	const listStr = `{
  "timestamp": 1576604526,
  "author": "Electronic Frontier Foundation https://eff.org",
  "expires": "2014-06-06T14:30:16.000000+00:00",
  "version": "0.1",
  "policy-aliases": {},
  "policies": {}
}`

	l, err := LoadList(strings.NewReader(listStr))
	if err != nil {
		t.Fatal(err)
	}

	expectDate := time.Date(2019, time.December, 17, 17, 42, 06, 0, time.UTC)
	if !time.Time(l.Timestamp).Equal(expectDate) {
		t.Fatal("Wrong time:", l.Timestamp, "is not", expectDate)
	}
}

func init() {
	now = func() time.Time {
		return time.Date(2014, time.June, 6, 14, 30, 16, 0, time.UTC)
	}
}
