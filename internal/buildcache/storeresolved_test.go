package buildcache

import (
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A store host ADAMIC_BUILD_STORE_RESOLVED names is dialed at its addresses, never resolved, so a traced store lookup
// reads no DNS configuration (#vdha67a): a name no resolver knows answers through the address given, and the same name
// without it fails to resolve. An address that isn't one is said, and another host resolves as usual. Mutant: the
// dialer resolving the host whatever the variable names.
// Not parallel: t.Setenv.
func TestAStoreHostResolvedByTheCallerIsNeverLookedUp(t *testing.T) {
	served := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { io.WriteString(writer, "held") }))
	defer served.Close()
	_, port, _ := net.SplitHostPort(strings.TrimPrefix(served.URL, "http://"))
	address := "http://store.invalid:" + port + "/blobs/x"
	get := func() (string, error) {
		response, err := (&http.Client{Transport: storeTransport()}).Get(address)
		if err != nil {
			return "", err
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		return string(body), err
	}
	t.Setenv(storeResolved, "")
	if _, err := get(); err == nil {
		t.Fatal("store.invalid answered with nothing naming its address: the test can't tell a lookup from none")
	}
	t.Setenv(storeResolved, "elsewhere.invalid=192.0.2.1 store.invalid=not-an-address,127.0.0.1")
	if body, err := get(); err != nil || body != "held" {
		t.Fatalf("a host the variable names, dialed at its address: %q, %v", body, err)
	}
	t.Setenv(storeResolved, "store.invalid=not-an-address")
	if _, err := get(); err == nil || !strings.Contains(err.Error(), "isn't an address") {
		t.Fatalf("an address that isn't one: %v", err)
	}
}

// A traced lookup's certificate roots are loaded as keying work, once, and pinned on the store client, so a handshake
// inside a product's trace never reads the machine's certificate bundle (the smoke's css and yaml reds, Oct 12). Mutant:
// the roots loaded outside the keying window.
// Not parallel: t.Setenv and systemRoots.
func TestTheStoresCertificateRootsAreLoadedAsKeying(t *testing.T) {
	t.Setenv("ADAMIC_BUILD_TRACE", t.TempDir())
	pool := x509.NewCertPool()
	saved := systemRoots
	t.Cleanup(func() { systemRoots = saved })
	keyingWhenLoaded := int64(-1)
	systemRoots = func() (*x509.CertPool, error) {
		keyingWhenLoaded = keyingNow.Load()
		return pool, nil
	}
	transport := storeTransport().(*http.Transport)
	if keyingWhenLoaded < 1 {
		t.Fatalf("the roots were loaded with %d keyings running", keyingWhenLoaded)
	}
	if transport.TLSClientConfig == nil || transport.TLSClientConfig.RootCAs != pool {
		t.Fatal("the roots loaded aren't the store client's")
	}
}
