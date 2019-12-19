package preload

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type mockFile struct {
	contentType string
	body        string
}

func mockHTTP(fs map[string]mockFile) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		file, ok := fs[r.URL.Path]
		if !ok {
			http.NotFound(rw, r)
			return
		}

		if file.contentType != "" {
			rw.Header().Add("Content-Type", file.contentType)
		}
		io.WriteString(rw, file.body)
	})
}

func TestDownload(t *testing.T) {
	handler := mockHTTP(map[string]mockFile{
		"/policies.json": {
			contentType: "application/json",
			body:        sampleList,
		},
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()
	c := srv.Client()

	list, err := Download(c, Source{
		ListURI: srv.URL + "/policies.json",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(list, sampleListParsed) {
		t.Fatalf("Wrong structure downloaded:\nWant %+v\nGot : %+v", sampleListParsed, list)
	}
}

func TestDownload_WrongContentType(t *testing.T) {
	handler := mockHTTP(map[string]mockFile{
		"/policies.json": {
			contentType: "text/plain",
			body:        sampleList,
		},
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()
	c := srv.Client()

	_, err := Download(c, Source{
		ListURI: srv.URL + "/policies.json",
	})
	if err == nil {
		t.Fatal("Expected an error, got none")
	}
}

const testPGPKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EXfpYkAEEAKqcUT/a+rUJKREDaexcW515rt4Xfq4sLpOgBmOnZtszs0rUoTLG
crpHWsSIl+dv3SLdNxC6LpEIWu19MMqtvjs12qAA+CigE3luS4DD+04YxxxOdFFU
n3Z5xJDQYTejRPhEzdCU7fAN/8hhByPcSZApU+3Fo56UaQWBExp944XtABEBAAG0
CFRlc3QgS2V5iM4EEwEKADgWIQQm0On7jdiwqjEvTPFmcewxesogAgUCXfpYkAIb
AwULCQgHAwUVCgkICwUWAgMBAAIeAQIXgAAKCRBmcewxesogAhrfA/97rrbAStFT
YV/FOvxSmw6knja24+0m5lxaH+TRnZt0GEcrf7PH8WlRg9bx4h1p+3B2tcpJzU26
6LGgDjWxKfdL0oHwz8zba81wqVmUcTr0kdwAL5jQ8Dm3wcmObIZre8gLxGldw0ob
xvjhm5hf9KpgW8OAXo2Xo6QsjJHFMMqbJQ==
=6UWm
-----END PGP PUBLIC KEY BLOCK-----`

const testPGPKeySec = `-----BEGIN PGP PRIVATE KEY BLOCK-----

lQHYBF36WJABBACqnFE/2vq1CSkRA2nsXFudea7eF36uLC6ToAZjp2bbM7NK1KEy
xnK6R1rEiJfnb90i3TcQui6RCFrtfTDKrb47NdqgAPgooBN5bkuAw/tOGMccTnRR
VJ92ecSQ0GE3o0T4RM3QlO3wDf/IYQcj3EmQKVPtxaOelGkFgRMafeOF7QARAQAB
AAP9H2DnqqRmRuSb6najxSaJbRGjwVI16OfUWy9r7WktCDTejW1FBpcsI6ma/pmW
wqi21cI07f0oMmGEg7hqQGSrH2DctGCKQLGkvO+DQV3ovnRtAD07kgAmnosgLA6D
cVjKpAeFxHt7wUwYZryMo93VpR0nfhH4twa/mUloj6JGATsCANCjoLHz7Fc3+QFF
bEpUD/QBQxPfmIICLyIyB6lfmKd8B+L4o8DIDyiXDMUO1PvYdRAVCBVw+/eUPV6J
Fsxuox8CANFWyVYIfSZNuK6FCVpw1Q0dKvCGnVMYUhVo0bXTWwlAH5pSOmfN+fX+
BdmYtEKK+cIZ7TUf2YPqZC+KV7zu4XMB/280cEvKqMvED+BpZjjEYacrq/OF6EL9
rnpTSZ92TS6Xi5oVoWpoFn7+ut3DP4BWyl2VX2fwMvd5usEXLqU9HnmdsrQIVGVz
dCBLZXmIzgQTAQoAOBYhBCbQ6fuN2LCqMS9M8WZx7DF6yiACBQJd+liQAhsDBQsJ
CAcDBRUKCQgLBRYCAwEAAh4BAheAAAoJEGZx7DF6yiACGt8D/3uutsBK0VNhX8U6
/FKbDqSeNrbj7SbmXFof5NGdm3QYRyt/s8fxaVGD1vHiHWn7cHa1yknNTbrosaAO
NbEp90vSgfDPzNtrzXCpWZRxOvSR3AAvmNDwObfByY5shmt7yAvEaV3DShvG+OGb
mF/0qmBbw4BejZejpCyMkcUwypsl
=W1Pt
-----END PGP PRIVATE KEY BLOCK-----`

const sampleListSig = `-----BEGIN PGP SIGNATURE-----

iLMEAAEKAB0WIQQm0On7jdiwqjEvTPFmcewxesogAgUCXfpb/wAKCRBmcewxesog
AnkjA/0awPQeKEhFtcFc3Fh4OoToIIRhsQgonrFEKh9+VGM3/Le2uswGwhKwCcBr
Uw5qbMGnKtNavooTmuCxktlzFhfz+jOwA7YJ3B5Ep3SiWKQi2x4fW8BCvkIHNlD2
Sw23jShEvdEJZCPyBbVyapmwq3/YyReIlzYli7Ec3aRot4Dvpw==
=izFf
-----END PGP SIGNATURE-----`

func TestDownload_VerifyPGP(t *testing.T) {
	handler := mockHTTP(map[string]mockFile{
		"/policies.json": {
			contentType: "application/json",
			body:        sampleList,
		},
		"/policies.json.asc": {
			body: sampleListSig,
		},
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()
	c := srv.Client()

	list, err := Download(c, Source{
		ListURI: srv.URL + "/policies.json",
		SigURI:  srv.URL + "/policies.json.asc",
		SigKey:  testPGPKey,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(list, sampleListParsed) {
		t.Fatalf("Wrong structure downloaded:\nWant %+v\nGot : %+v", sampleListParsed, list)
	}
}

func TestDownload_VerifyPGP_BrokenSig(t *testing.T) {
	const listStr = `{
  "timestamp": "2014-06-06T14:30:16.000000+00:00",
  "author": "Electronic Frontier Foundation https://eff.org",
  "expires": "2014-06-06T14:30:16.000000+00:00",
  "version": "0.1",
  "policy-aliases": {},
  "policies": {}
}`
	handler := mockHTTP(map[string]mockFile{
		"/policies.json": {
			contentType: "application/json",
			body:        listStr,
		},
		"/policies.json.asc": {
			body: sampleListSig,
		},
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()
	c := srv.Client()

	_, err := Download(c, Source{
		ListURI: srv.URL + "/policies.json",
		SigURI:  srv.URL + "/policies.json.asc",
		SigKey:  testPGPKey,
	})
	if err == nil {
		t.Fatal("Expected an error, got none")
	}
	if _, ok := err.(PGPError); !ok {
		t.Fatal("Returned error should be PGPError")
	}
}
