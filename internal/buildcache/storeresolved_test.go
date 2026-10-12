package buildcache

import (
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
