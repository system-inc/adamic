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
// Not parallel: test.Setenv.
func TestAStoreHostResolvedByTheCallerIsNeverLookedUp(test *testing.T) {
	served := httptest.NewServer(
		http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { io.WriteString(writer, "held") }),
	)
	defer served.Close()
	_, port, _ := net.SplitHostPort(strings.TrimPrefix(served.URL, "http://"))
	address := "http://store.invalid:" + port + "/blobs/x"
	get := func() (string, error) {
		response, getError := (&http.Client{Transport: storeTransport()}).Get(address)
		if getError != nil {
			return "", getError
		}
		defer response.Body.Close()
		body, readError := io.ReadAll(response.Body)
		return string(body), readError
	}
	test.Setenv(storeResolved, "")
	if _, getError := get(); getError == nil {
		test.Fatal("store.invalid answered with nothing naming its address: the test can't tell a lookup from none")
	}
	test.Setenv(storeResolved, "elsewhere.invalid=192.0.2.1 store.invalid=not-an-address,127.0.0.1")
	if body, getError := get(); getError != nil || body != "held" {
		test.Fatalf("a host the variable names, dialed at its address: %q, %v", body, getError)
	}
	test.Setenv(storeResolved, "store.invalid=not-an-address")
	if _, getError := get(); getError == nil || !strings.Contains(getError.Error(), "isn't an address") {
		test.Fatalf("an address that isn't one: %v", getError)
	}
}

// A traced lookup's certificate roots are loaded as keying work, once, and pinned on the store client, so a handshake
// inside a product's trace never reads the machine's certificate bundle (the smoke's css and yaml reds, Oct 12).
// Mutant: the roots loaded outside the keying window.
// Not parallel: test.Setenv and systemRoots.
func TestTheStoresCertificateRootsAreLoadedAsKeying(test *testing.T) {
	test.Setenv("ADAMIC_BUILD_TRACE", test.TempDir())
	pool := x509.NewCertPool()
	saved := systemRoots
	test.Cleanup(func() { systemRoots = saved })
	keyingWhenLoaded := int64(-1)
	systemRoots = func() (*x509.CertPool, error) {
		keyingWhenLoaded = keyingNow.Load()
		return pool, nil
	}
	transport := storeTransport().(*http.Transport)
	if keyingWhenLoaded < 1 {
		test.Fatalf("the roots were loaded with %d keyings running", keyingWhenLoaded)
	}
	if transport.TLSClientConfig == nil || transport.TLSClientConfig.RootCAs != pool {
		test.Fatal("the roots loaded aren't the store client's")
	}
}
