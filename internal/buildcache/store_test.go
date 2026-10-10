package buildcache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// A store as Loom's serves it: blobs by sha256 and refs/build/<key>, read with no credential.
type fakeStore struct {
	mutex   sync.Mutex
	objects map[string][]byte
	reads   []string
	writes  []string
	broken  string
}

func (s *fakeStore) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if request.Method == http.MethodPut {
		s.write(writer, request)
		return
	}
	s.reads = append(s.reads, request.URL.Path)
	if request.URL.Path == s.broken {
		http.Error(writer, "upstream timeout", http.StatusBadGateway)
		return
	}
	content, ok := s.objects[request.URL.Path]
	if !ok {
		http.NotFound(writer, request)
		return
	}
	writer.Write(content)
}

// write is Loom's Worker: a bearer token, blobs that hash to their name, a ref only to a held blob, never changed.
func (s *fakeStore) write(writer http.ResponseWriter, request *http.Request) {
	name := strings.TrimPrefix(request.URL.Path, "/public")
	s.writes = append(s.writes, name)
	if request.Header.Get("Authorization") != "Bearer gate-box-token" {
		http.Error(writer, "forbidden", http.StatusForbidden)
		return
	}
	content, _ := io.ReadAll(request.Body)
	switch {
	case strings.HasPrefix(name, "/blobs/"):
		if sum := sha256.Sum256(content); "/blobs/"+hex.EncodeToString(sum[:]) != name {
			http.Error(writer, "hash", http.StatusBadRequest)
			return
		}
		s.objects[name] = content
		writer.WriteHeader(http.StatusCreated)
	case strings.HasPrefix(name, "/refs/"):
		held, ok := s.objects[name]
		if _, blob := s.objects["/blobs/"+string(content)]; !blob || (ok && string(held) != string(content)) {
			http.Error(writer, "conflict", http.StatusConflict)
			return
		}
		s.objects[name] = content
		writer.WriteHeader(http.StatusCreated)
	}
}

func (s *fakeStore) blob(content []byte) string {
	sum := sha256.Sum256(content)
	hash := hex.EncodeToString(sum[:])
	s.objects["/blobs/"+hash] = content
	return hash
}

// put stores a product for key the way publish does, and returns its manifest.
func (s *fakeStore) put(t *testing.T, key string, files map[string]string) manifest {
	t.Helper()
	product := manifest{Version: 1, Key: key, Name: "buildcache test"}
	for name, content := range files {
		product.Files = append(product.Files, manifestFile{Path: name, SHA256: s.blob([]byte(content)), Size: int64(len(content)), Mode: 0o644})
	}
	s.ref(t, key, product)
	return product
}

func (s *fakeStore) ref(t *testing.T, key string, product manifest) {
	t.Helper()
	encoded, err := json.Marshal(product)
	if err != nil {
		t.Fatal(err)
	}
	s.objects["/refs/build/"+key] = []byte(s.blob(encoded) + "\n")
}

func shared(t *testing.T) (*fakeStore, string) {
	t.Helper()
	_, log := cached(t)
	store := &fakeStore{objects: map[string][]byte{}}
	server := httptest.NewServer(store)
	t.Cleanup(server.Close)
	t.Setenv("ADAMIC_BUILD_STORE", server.URL)
	t.Setenv("ADAMIC_BUILD_STORE_WRITE", server.URL+"/public")
	t.Setenv("ADAMIC_BUILD_AUDIT", "0")
	return store, log
}

// recorded gives thisPackage a read set in the current cache, as a traced build of it on Workshop would have (it read
// buildcache.go), and returns the key it gives on this tree: the key a store holds its product under.
func recorded(t *testing.T) string {
	t.Helper()
	root := repositoryRootForTest(t)
	cache, err := cacheDirectory()
	if err != nil {
		t.Fatal(err)
	}
	nameKey := NameKey(root, thisPackage)
	set := readSet{Entries: []readEntry{{Kind: "content", Path: "internal/buildcache/buildcache.go"}}}
	evaluated, err := evaluate(root, cache, nameKey, set, 0)
	if err != nil {
		t.Fatal(err)
	}
	set.Key = evaluated.key
	if err = recordReads(cache, nameKey, thisPackage.Name, set); err != nil {
		t.Fatal(err)
	}
	return evaluated.key
}

func token(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "publish-token")
	if err := os.WriteFile(path, []byte("gate-box-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_BUILD_STORE_TOKEN", path)
}

// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
func TestAStoredProductIsFetchedNotBuilt(t *testing.T) {
	store, log := shared(t)
	store.put(t, recorded(t), map[string]string{"bin/checker": "the checker", "lib/a.o": "object"})
	directory := Product(t, thisPackage, func(string) error { t.Fatal("built a product the store holds"); return nil })
	for name, want := range map[string]string{"bin/checker": "the checker", "lib/a.o": "object"} {
		if content, err := os.ReadFile(filepath.Join(directory, name)); err != nil || string(content) != want {
			t.Fatalf("%s reads %q, %v", name, content, err)
		}
	}
	if lines, _ := os.ReadFile(log); !strings.Contains(string(lines), " fetched ") {
		t.Fatalf("census: %q", lines)
	}
	// Fetched once, then a local hit: the store isn't asked again.
	reads := len(store.reads)
	Product(t, thisPackage, func(string) error { t.Fatal("rebuilt"); return nil })
	if len(store.reads) != reads {
		t.Fatalf("a local hit read the store again: %v", store.reads[reads:])
	}
}

// Anything stored that doesn't check fails loudly as poisoning, and nothing reaches the local cache.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
func TestAStoredProductThatDoesntCheckIsPoisoning(t *testing.T) {
	cases := map[string]func(t *testing.T, store *fakeStore, key string){
		"a blob that isn't its hash": func(t *testing.T, store *fakeStore, key string) {
			product := store.put(t, key, map[string]string{"bin/checker": "the checker"})
			// The same length, so only the blob's hash can tell.
			store.objects["/blobs/"+product.Files[0].SHA256] = []byte("the checkeR")
		},
		"a manifest for another key": func(t *testing.T, store *fakeStore, key string) {
			product := store.put(t, key, map[string]string{"bin/checker": "the checker"})
			product.Key = strings.Repeat("0", 64)
			store.ref(t, key, product)
		},
		"a path outside the product": func(t *testing.T, store *fakeStore, key string) {
			product := store.put(t, key, map[string]string{"bin/checker": "the checker"})
			product.Files[0].Path = "../escape"
			store.ref(t, key, product)
		},
		"a size that isn't the blob's": func(t *testing.T, store *fakeStore, key string) {
			product := store.put(t, key, map[string]string{"bin/checker": "the checker"})
			product.Files[0].Size++
			store.ref(t, key, product)
		},
		"a ref naming a missing manifest": func(t *testing.T, store *fakeStore, key string) {
			store.objects["/refs/build/"+key] = []byte(strings.Repeat("a", 64))
		},
		"a ref that isn't a hash": func(t *testing.T, store *fakeStore, key string) {
			store.objects["/refs/build/"+key] = []byte("latest")
		},
	}
	for name, plant := range cases {
		t.Run(name, func(t *testing.T) {
			store, _ := shared(t)
			key := recorded(t)
			plant(t, store, key)
			_, err := Get(thisPackage, func(string) error { t.Fatal("built instead of failing"); return nil })
			var wrong poisonedError
			if !errors.As(err, &wrong) {
				t.Fatalf("got %v, want cache poisoning", err)
			}
			if _, err := os.Stat(filepath.Join(os.Getenv("ADAMIC_BUILD_CACHE_DIR"), key)); err == nil {
				t.Fatal("a poisoned product reached the local cache")
			}
		})
	}
}

// A product the store doesn't hold, or can't serve, is built only under a trace; untraced, the miss is refused and
// nothing is built.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
func TestAnUnstoredOrUnreachableProductIsBuiltOnlyUnderATrace(t *testing.T) {
	for _, address := range []string{"", "http://127.0.0.1:1", "failing mid-fetch"} {
		t.Run(address, func(t *testing.T) {
			if address == "" {
				shared(t)
				recorded(t)
			} else if address == "failing mid-fetch" {
				// The ref and manifest read, then a blob's read fails: the store's trouble, not poisoning.
				store, _ := shared(t)
				product := store.put(t, recorded(t), map[string]string{"product": "stored"})
				store.broken = "/blobs/" + product.Files[0].SHA256
			} else {
				cached(t)
				t.Setenv("ADAMIC_BUILD_STORE", address)
				recorded(t)
			}
			var built bool
			build := func(directory string) error {
				built = true
				return os.WriteFile(filepath.Join(directory, "product"), []byte("built"), 0o644)
			}
			if _, err := Get(thisPackage, build); !errors.Is(err, ErrUntraced) || built {
				t.Fatalf("an untraced miss: %v, built %t", err, built)
			}
			t.Setenv("ADAMIC_BUILD_TRACE", t.TempDir())
			Product(t, thisPackage, build)
			if !built {
				t.Fatal("not built under the trace")
			}
		})
	}
}

// Only a settled traced build publishes: every blob, then the manifest, then the ref, so the ref never names a blob
// the store doesn't hold. Another machine, given the read sets, then fetches it by the key they give.
// Not parallel: newRig.
func TestOnlyASettledBuildIsPublishedAndAnotherMachineFetchesIt(t *testing.T) {
	r := newRig(t)
	store, _ := shared(t)
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", r.cache)
	token(t)
	Product(t, port, func(directory string) error {
		os.MkdirAll(filepath.Join(directory, "bin"), 0o755)
		os.WriteFile(filepath.Join(directory, "bin", "checker"), []byte("the checker"), 0o755)
		return os.WriteFile(filepath.Join(directory, "notes"), []byte("notes"), 0o644)
	})
	if len(store.writes) != 0 {
		t.Fatalf("a build published before it was settled: %v", store.writes)
	}
	build := r.built(t)["port"]
	r.window(build.Build, func() { r.open(r.pid, "source/a.c") })
	settlement := r.settle(t)
	if len(settlement.Settled) != 1 || strings.Contains(settlement.Settled[0], "publish failed") {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	file, _ := loadReads(r.cache, build.NameKey)
	key := file.Sets[0].Key
	// A candidate's settlement writes the candidate ref, never main's.
	if len(store.writes) != 4 || !strings.HasPrefix(store.writes[0], "/blobs/") || !strings.HasPrefix(store.writes[2], "/blobs/") || store.writes[3] != "/refs/build-candidate/"+key {
		t.Fatalf("writes: %q", store.writes)
	}
	// Another machine: an empty cache but for the read sets, untraced. Fetched, never built, exec bit kept.
	elsewhere := t.TempDir()
	content, _ := os.ReadFile(readsPath(r.cache, build.NameKey))
	os.WriteFile(readsPath(elsewhere, build.NameKey), content, 0o644)
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", elsewhere)
	t.Setenv("ADAMIC_BUILD_TRACE", "")
	directory := Product(t, port, func(string) error { t.Fatal("rebuilt what the store holds"); return nil })
	info, err := os.Stat(filepath.Join(directory, "bin", "checker"))
	if err != nil || info.Mode()&0o111 == 0 || directory != filepath.Join(elsewhere, key) {
		t.Fatalf("the fetched checker at %s: %v, mode %v", directory, err, info)
	}
}

// A ref never changes: a store already holding a different product for the key refuses, and the settlement says the
// key isn't honest without failing the product it placed.
// Not parallel: newRig.
func TestADifferentProductForAStoredKeyIsSaidLoudly(t *testing.T) {
	r := newRig(t)
	shared(t)
	token(t)
	trace := r.trace
	for _, content := range []string{"this machine's", "another build's"} {
		cache := t.TempDir()
		t.Setenv("ADAMIC_BUILD_CACHE_DIR", cache)
		r.cache = cache
		r.trace = filepath.Join(filepath.Dir(trace), filepath.Base(trace)+"-"+strings.Fields(content)[0])
		os.MkdirAll(r.trace, 0o755)
		t.Setenv("ADAMIC_BUILD_TRACE", r.trace)
		Product(t, port, building(content))
		build := r.built(t)["port"]
		r.window(build.Build, func() { r.open(r.pid, "source/a.c") })
		settlement := r.settle(t)
		if len(settlement.Settled) != 1 {
			t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
		}
		if honest := !strings.Contains(settlement.Settled[0], "isn't honest"); honest != (content == "this machine's") {
			t.Fatalf("%s: %s", content, settlement.Settled[0])
		}
	}
}

// Audited, a fetched product is rebuilt and compared, where building may (under a trace): the same product passes, a
// different one is poisoning.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
func TestAnAuditedFetchThatDiffersFromARebuildIsPoisoning(t *testing.T) {
	store, log := shared(t)
	t.Setenv("ADAMIC_BUILD_AUDIT", "1")
	t.Setenv("ADAMIC_BUILD_TRACE", t.TempDir())
	store.put(t, recorded(t), map[string]string{"product": "built"})
	build := func(directory string) error {
		return os.WriteFile(filepath.Join(directory, "product"), []byte("built"), 0o644)
	}
	Product(t, thisPackage, build)
	if lines, _ := os.ReadFile(log); !strings.Contains(string(lines), " audited ") {
		t.Fatalf("census: %q", lines)
	}
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
	store.put(t, recorded(t), map[string]string{"product": "a wrong product"})
	_, err := Get(thisPackage, build)
	var wrong poisonedError
	if !errors.As(err, &wrong) || !strings.Contains(err.Error(), "differs from a rebuild") {
		t.Fatalf("got %v, want poisoning", err)
	}
}

// Refs are split by trust: a candidate reads main's refs, then candidates'; a trusted reader reads only main's; and
// main's gate, uncached and traced, publishes to main's ref when its build settles.
// Not parallel: newRig.
func TestRefsAreSplitByTrust(t *testing.T) {
	store, _ := shared(t)
	key := recorded(t)
	store.put(t, key, map[string]string{"product": "a candidate's"})
	store.objects["/refs/build-candidate/"+key] = store.objects["/refs/build/"+key]
	delete(store.objects, "/refs/build/"+key)
	directory := Product(t, thisPackage, func(string) error { t.Fatal("a candidate built what a candidate ref holds"); return nil })
	if content, _ := os.ReadFile(filepath.Join(directory, "product")); string(content) != "a candidate's" {
		t.Fatalf("a candidate read %q", content)
	}
	// A trusted reader never takes it: with only a candidate's ref stored, the product is missing to it.
	t.Setenv("ADAMIC_BUILD_STORE_TRUST", "main")
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
	recorded(t)
	if _, err := Get(thisPackage, func(string) error { t.Fatal("built untraced"); return nil }); !errors.Is(err, ErrUntraced) {
		t.Fatalf("a trusted reader with only a candidate's ref: %v", err)
	}
	// Main's gate: uncached, traced, settled, published to main's ref.
	address := os.Getenv("ADAMIC_BUILD_STORE")
	r := newRig(t)
	t.Setenv("ADAMIC_BUILD_STORE", address)
	token(t)
	t.Setenv("ADAMIC_BUILD_CACHE", "off")
	Product(t, port, building("main's"))
	build := r.built(t)["port"]
	r.window(build.Build, func() { r.open(r.pid, "source/a.c") })
	t.Setenv("ADAMIC_BUILD_CACHE", "")
	if settlement := r.settle(t); len(settlement.Settled) != 1 || strings.Contains(settlement.Settled[0], "publish failed") {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	file, _ := loadReads(r.cache, build.NameKey)
	if last := store.writes[len(store.writes)-1]; last != "/refs/build/"+file.Sets[0].Key {
		t.Fatalf("main's gate wrote %v", store.writes)
	}
}

// The shared tier is off unless ADAMIC_BUILD_STORE turns it on (Oct 9 20:38Z): unset, nothing is fetched or published,
// even on a machine holding the write credential, and an untraced miss isn't built either.
// Not parallel: points the build cache, the store's writer and the token at this test through t.Setenv.
func TestTheSharedTierIsOffUnlessTurnedOn(t *testing.T) {
	for value, want := range map[string]string{"": "", "off": "", "on": defaultStore, "http://127.0.0.1:9/": "http://127.0.0.1:9"} {
		t.Setenv("ADAMIC_BUILD_STORE", value)
		if got := storeAddress(); got != want {
			t.Fatalf("ADAMIC_BUILD_STORE=%q: the store is %q, want %q", value, got, want)
		}
	}
	cached(t)
	store := &fakeStore{objects: map[string][]byte{}}
	server := httptest.NewServer(store)
	t.Cleanup(server.Close)
	t.Setenv("ADAMIC_BUILD_STORE", "")
	t.Setenv("ADAMIC_BUILD_STORE_WRITE", server.URL+"/public")
	token(t)
	store.put(t, recorded(t), map[string]string{"product": "stored"})
	if _, err := Get(thisPackage, func(string) error { t.Fatal("built untraced"); return nil }); !errors.Is(err, ErrUntraced) {
		t.Fatalf("an unset store: %v", err)
	}
	if len(store.reads)+len(store.writes) != 0 {
		t.Fatalf("an unset store reached the store: reads %v, writes %v", store.reads, store.writes)
	}
}
