package buildcache

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// The shared tier (@system_adamic's ruling, Oct 9: one store, Loom's): a product another machine built is fetched by
// its key from Loom's artifacts store, read directly and unauthenticated, so a test anywhere starts after a fetch instead of a
// build. A product is a manifest blob naming one blob per file; refs/build/<key> holds the manifest's sha256. Every
// blob is checked against its own hash and the manifest against the key it was asked for, so a stored product can't
// be wrong without failing loudly: a mismatch is an error, never a quiet rebuild. An unreachable store is a miss.
//
// The shared tier is off unless ADAMIC_BUILD_STORE turns it on: "on" for Loom's artifacts store, or another store's address
// (@system_adamic_release, Oct 9 20:38Z: native products embed their build directory, so a stored one never matches a
// rebuild and every audited hit reds a gate, and a test that asserts a cold build fetched one instead; it stays off
// until those are fixed, and gocacheprog is untouched). Workshop's tree builder opts in with ADAMIC_BUILD_STORE=traced
// (traced=<address> for another store) and runs under cmd/traced.
//
// buildcache only reads the store, and only products: it never writes one, and never asks it for a read set. A product
// reaches a runner through Loom, whose tree index carries each tree's read sets into the unit's cache and whose held
// products answer refs/build and blobs/. The read sets the tree builder once published as refs/reads went with the
// publishing (Oct 14: Workshop's tree builder was stood down, so nothing wrote them, yet every lookup the cache's own
// read sets missed asked the store for 18 of them, 96.7% of the bucket's billed reads, every one a 404).
const defaultStore = "https://artifacts.loom.system.inc"

// A fetched product is rebuilt and compared file by file at this rate (ADAMIC_BUILD_AUDIT overrides it): a store
// that served a wrong product fails the test as poisoned (the ruling on #x2651cf: audit every gate's hits).
const defaultAudit = 0.05

var errNotStored = errors.New("not in the store")

type manifest struct {
	Version int            `json:"version"`
	Key     string         `json:"key"`
	Name    string         `json:"name"`
	Files   []manifestFile `json:"files"`
}

type manifestFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
	Mode   uint32 `json:"mode"`
}

// storeAddress is the shared store's address, or "" when the shared tier is off.
func storeAddress() string {
	address := os.Getenv("ADAMIC_BUILD_STORE")
	if rest, ok := strings.CutPrefix(address, "traced="); ok {
		address = rest
	}
	switch address {
	case "", "off":
		return ""
	case "on", "traced":
		address = defaultStore
	}
	return strings.TrimSuffix(address, "/")
}

// treeBuilder says whether this run is Workshop's tree builder, the one opted in to building only under a trace
// (ADAMIC_BUILD_STORE=traced, or traced=<address>).
func treeBuilder() bool {
	setting := os.Getenv("ADAMIC_BUILD_STORE")
	return setting == "traced" || strings.HasPrefix(setting, "traced=")
}

// storeClient is the store's client, made at its first use: its transport is storeTransport's.
var storeClient = sync.OnceValue(func() *http.Client {
	return &http.Client{Timeout: 5 * time.Minute, Transport: storeTransport()}
})

// systemRoots loads the machine's certificate roots: a variable so a test can see where it runs.
var systemRoots = x509.SystemCertPool

// storeResolved is where the caller already resolved the store's hosts, ADAMIC_BUILD_STORE_RESOLVED: entries
// "<host>=<address>,<address>…" separated by spaces. Workshop's tree builder resolves them outside the trace (build-tree)
// so a traced product test's store lookups read no DNS configuration: /etc/resolv.conf, /etc/hosts and
// /etc/nsswitch.conf were in every product a lookup's trace settled, each box's own (its nameserver), so no runner on
// another box matched a native product Workshop built (Fabric on Cloud, Oct 12, #vdha67a). With it a lookup never asks
// the resolver, so a DNS file in a read set is one the product's own build read, keyed as before.
const storeResolved = "ADAMIC_BUILD_STORE_RESOLVED"

// storeTransport is http.DefaultTransport's, dialing a host storeResolved names at its addresses, each in turn, rather
// than resolving it: an address that doesn't answer is a miss like any other, never a lookup. Any other host resolves
// as usual. Its certificate roots are the machine's, loaded once here as keying work: a handshake left to load them read
// /etc/ssl/certs inside a product's trace, each box's own bundle, so the product missed on every other box (Loom and
// Tests, Oct 12, the smoke's css and yaml reds). Roots that don't load leave the handshake to load them, as before.
func storeTransport() http.RoundTripper {
	resolved := map[string][]string{}
	for _, entry := range strings.Fields(os.Getenv(storeResolved)) {
		host, addresses, ok := strings.Cut(entry, "=")
		if ok && host != "" {
			resolved[host] = strings.Split(addresses, ",")
		}
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	var roots *x509.CertPool
	keying(func() { roots, _ = systemRoots() })
	if roots != nil {
		transport.TLSClientConfig = &tls.Config{RootCAs: roots}
	}
	if len(resolved) == 0 {
		return transport
	}
	dialer := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	transport.DialContext = func(dialContext context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil || len(resolved[host]) == 0 {
			return dialer.DialContext(dialContext, network, address)
		}
		var errs []error
		for _, ip := range resolved[host] {
			if net.ParseIP(ip) == nil {
				errs = append(errs, fmt.Errorf("%s=%s isn't an address", host, ip))
				continue
			}
			connection, err := dialer.DialContext(dialContext, network, net.JoinHostPort(ip, port))
			if err == nil {
				return connection, nil
			}
			errs = append(errs, err)
		}
		return nil, fmt.Errorf("%s at %s, as %s names it: %w", host, strings.Join(resolved[host], ","), storeResolved, errors.Join(errs...))
	}
	return transport
}

// download reads one object, or errNotStored for a 404.
func download(address string) ([]byte, error) {
	response, err := storeClient().Get(address)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return nil, errNotStored
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s answered %s", address, response.Status)
	}
	return io.ReadAll(response.Body)
}

// blob reads the blob with this sha256 and checks it is that blob.
func blob(store, sum string) ([]byte, error) {
	content, err := download(store + "/blobs/" + sum)
	if err != nil {
		return nil, err
	}
	if actual := sha256.Sum256(content); hex.EncodeToString(actual[:]) != sum {
		return nil, poisoned(fmt.Errorf("blob %s has sha256 %x", sum, actual))
	}
	return content, nil
}

// poisonedError is a stored product that isn't what its key or hashes say: it fails the caller, never quietly rebuilt.
type poisonedError struct{ err error }

func (e poisonedError) Error() string { return "cache poisoning: " + e.err.Error() }
func (e poisonedError) Unwrap() error { return e.err }

func poisoned(err error) error { return poisonedError{err} }

// Refs are split by trust (Loom, Oct 9: a key covers a product's inputs, not the caller's build function, so a candidate
// could define a different build under main's key): refs/build holds main's products, refs/build-candidate everyone
// else's. A reader trusted as main's (ADAMIC_BUILD_STORE_TRUST=main) reads refs/build alone; a candidate reads
// refs/build, then refs/build-candidate, so a candidate can only ever mislead candidates.
func trusted() bool { return os.Getenv("ADAMIC_BUILD_STORE_TRUST") == "main" }

func readNamespaces() []string {
	if trusted() {
		return []string{"build"}
	}
	return []string{"build", "build-candidate"}
}

// fetch places the stored product for key into directory. It returns errNotStored when the store has none or can't
// be reached (the caller builds), and a poisonedError when what it holds doesn't check.
func fetch(key, directory string) error {
	store := storeAddress()
	var reference []byte
	var namespace string
	err := errNotStored
	for _, namespace = range readNamespaces() {
		if reference, err = download(store + "/refs/" + namespace + "/" + key); !errors.Is(err, errNotStored) {
			break
		}
	}
	if err != nil {
		if errors.Is(err, errNotStored) {
			return errNotStored
		}
		return fmt.Errorf("%w: %v", errNotStored, err)
	}
	// Whatever the ref holds, the manifest is checked against it as a hash, so a ref that isn't one names nothing.
	sum := strings.TrimSpace(string(reference))
	content, err := blob(store, sum)
	if err != nil {
		return named(err, fmt.Sprintf("refs/%s/%s names manifest %s", namespace, key, sum))
	}
	var product manifest
	if err = json.Unmarshal(content, &product); err != nil {
		return poisoned(fmt.Errorf("manifest %s: %v", sum, err))
	}
	if product.Version != 1 || product.Key != key {
		return poisoned(fmt.Errorf("manifest %s is version %d for key %s, asked for %s", sum, product.Version, product.Key, key))
	}
	for _, file := range product.Files {
		clean := path.Clean(file.Path)
		if file.Path != clean || path.IsAbs(clean) || clean == "." || strings.HasPrefix(clean, "../") || clean == ".." {
			return poisoned(fmt.Errorf("manifest %s names the path %q", sum, file.Path))
		}
		content, err := blob(store, file.SHA256)
		if err != nil {
			return named(err, fmt.Sprintf("manifest %s names blob %s for %s", sum, file.SHA256, file.Path))
		}
		if int64(len(content)) != file.Size {
			return poisoned(fmt.Errorf("%s is %d bytes, its manifest says %d", file.Path, len(content), file.Size))
		}
		target := filepath.Join(directory, filepath.FromSlash(clean))
		if err = os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err = os.WriteFile(target, content, fs.FileMode(file.Mode&0o755)); err != nil {
			return err
		}
	}
	return nil
}

// named sorts a failure to read a blob something named: a wrong blob is poisoned, a named blob the store doesn't hold
// is poisoned too (a ref or manifest pointing at nothing), and anything else (the network) is a miss.
func named(err error, what string) error {
	var wrong poisonedError
	switch {
	case errors.As(err, &wrong):
		return err
	case errors.Is(err, errNotStored):
		return poisoned(fmt.Errorf("%s, which the store doesn't hold", what))
	default:
		return fmt.Errorf("%w: reading what %s: %v", errNotStored, what, err)
	}
}

// describeProduct is a product directory's manifest: every regular file by path, hash, size and mode, sorted.
// Anything else in a product (a symlink, a device) can't be stored, so it is an error.
func describeProduct(key, name, directory string) (manifest, error) {
	product := manifest{Version: 1, Key: key, Name: name}
	root, err := ProductRoot(directory)
	if err != nil {
		return product, err
	}
	err = filepath.WalkDir(root, func(file string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file, so the product can't be stored", file)
		}
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, file)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(content)
		mode := uint32(0o644)
		if info.Mode()&0o111 != 0 {
			mode = 0o755
		}
		product.Files = append(product.Files, manifestFile{Path: filepath.ToSlash(relative), SHA256: hex.EncodeToString(sum[:]), Size: int64(len(content)), Mode: mode})
		return nil
	})
	sort.Slice(product.Files, func(i, j int) bool { return product.Files[i].Path < product.Files[j].Path })
	return product, err
}

// auditing says whether this fetched product is rebuilt and compared.
func auditing() bool {
	rate := defaultAudit
	if value := os.Getenv("ADAMIC_BUILD_AUDIT"); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			rate = parsed
		}
	}
	return rand.Float64() < rate
}

// auditListed says whether nameKey is one ADAMIC_BUILD_AUDIT_KEYS's file lists, one name key a line: a product a tree
// built and a check rebuilds on its local hit (#earlycut's sampled key check). Unset lists nothing. A file named that
// can't be read is the lookup's error, naming it, since a check that audits nothing would read clean (Tests' must). The
// file is read once a process, by its path (auditLists).
func auditListed(nameKey string) (bool, error) {
	path := os.Getenv("ADAMIC_BUILD_AUDIT_KEYS")
	if path == "" {
		return false, nil
	}
	read, _ := auditLists.LoadOrStore(path, sync.OnceValues(func() ([]string, error) {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("the audit's list of name keys (ADAMIC_BUILD_AUDIT_KEYS): %w", err)
		}
		return strings.Fields(string(content)), nil
	}))
	listed, err := read.(func() ([]string, error))()
	return slices.Contains(listed, nameKey), err
}

// auditLists holds each ADAMIC_BUILD_AUDIT_KEYS file this process read, by its path, as its read.
var auditLists sync.Map

// auditHit rebuilds a product found in the cache, directory under key, into a new directory under the process's
// temporary directory, never beside the product, so a build that bakes in a path its key doesn't name differs too
// (Tests' change 2), and appends the audit line Loom reads (builder.AuditTree's auditLine) to $ADAMIC_BUILD_LOG: clean
// with the digest, or mismatch with the stored digest, the rebuilt one and where the rebuilt copy is kept. A rebuild
// that fails is the lookup's error, since nothing was compared.
func auditHit(key, name, directory string, build func(directory string) error) error {
	rebuilt, err := os.MkdirTemp("", ".audit-"+key[:12]+"-")
	if err != nil {
		return err
	}
	if err = build(rebuilt); err != nil {
		os.RemoveAll(rebuilt)
		return fmt.Errorf("audit of %s (%s): its rebuild failed: %w", key[:12], name, err)
	}
	stored, err := productDigest(directory)
	if err != nil {
		os.RemoveAll(rebuilt)
		return err
	}
	fresh, err := productDigest(rebuilt)
	if err != nil {
		os.RemoveAll(rebuilt)
		return err
	}
	name = strings.ReplaceAll(name, " ", "_")
	if fresh == stored {
		os.RemoveAll(rebuilt)
		note("audit %s %s clean %s", name, key, stored)
		return nil
	}
	note("audit %s %s mismatch %s %s %s", name, key, stored, fresh, rebuilt)
	return nil
}

// audit rebuilds a fetched product beside it and compares the two file by file.
func audit(key, name, fetched string, build func(directory string) error) error {
	rebuilt, err := os.MkdirTemp(filepath.Dir(fetched), ".audit-"+key[:12]+"-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(rebuilt)
	if err = build(rebuilt); err != nil {
		return err
	}
	want, err := describeProduct(key, name, rebuilt)
	if err != nil {
		return err
	}
	got, err := describeProduct(key, name, fetched)
	if err != nil {
		return err
	}
	wantJSON, _ := json.Marshal(want.Files)
	gotJSON, _ := json.Marshal(got.Files)
	if !bytes.Equal(wantJSON, gotJSON) {
		return poisoned(fmt.Errorf("the stored product for %s (%s) differs from a rebuild: stored %s, rebuilt %s (or the build isn't reproducible)", key[:12], name, gotJSON, wantJSON))
	}
	return nil
}
