package buildcache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// A store as Loom's serves it: blobs by sha256 and refs/build/<key>, read with no credential. buildcache never writes
// to a store, so any other method is refused and kept in writes for a test to see.
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
	if request.Method != http.MethodGet {
		s.writes = append(s.writes, request.Method+" "+request.URL.Path)
		http.Error(writer, "read only", http.StatusMethodNotAllowed)
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

func (s *fakeStore) blob(content []byte) string {
	sum := sha256.Sum256(content)
	hash := hex.EncodeToString(sum[:])
	s.objects["/blobs/"+hash] = content
	return hash
}

// put stores a product for key as Loom's store holds one, and returns its manifest.
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
	evaluated, err := evaluate(root, cache, nameKey, set, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	set.Key = evaluated.key
	if err = recordReads(cache, nameKey, thisPackage.Name, set); err != nil {
		t.Fatal(err)
	}
	return evaluated.key
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

// A product the store doesn't hold, or can't serve, is built here, for this machine alone, and nothing is written to
// the store.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
func TestAnUnstoredOrUnreachableProductIsBuiltHere(t *testing.T) {
	for _, address := range []string{"", "http://127.0.0.1:1", "failing mid-fetch"} {
		t.Run(address, func(t *testing.T) {
			var store *fakeStore
			if address == "" {
				store, _ = shared(t)
				recorded(t)
			} else if address == "failing mid-fetch" {
				// The ref and manifest read, then a blob's read fails: the store's trouble, not poisoning.
				store, _ = shared(t)
				product := store.put(t, recorded(t), map[string]string{"product": "stored"})
				store.broken = "/blobs/" + product.Files[0].SHA256
			} else {
				cached(t)
				t.Setenv("ADAMIC_BUILD_STORE", address)
				recorded(t)
			}
			var built bool
			Product(t, thisPackage, func(directory string) error {
				built = true
				return os.WriteFile(filepath.Join(directory, "product"), []byte("built"), 0o644)
			})
			if !built {
				t.Fatal("not built")
			}
			if store != nil && len(store.writes) != 0 {
				t.Fatalf("an untraced build wrote to the store: %v", store.writes)
			}
		})
	}
}

// The tree builder's settled product stays in its cache, read sets beside it, and nothing reaches the store: no write,
// and no read set asked for. Another machine, given those read sets as Loom's tree index ships them, fetches the
// product from the store by the key they give.
// Not parallel: newRig.
func TestTheTreeBuilderWritesNothingAndARunnerFetchesByTheShippedReadSets(t *testing.T) {
	r := newRig(t)
	store, _ := shared(t)
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", r.cache)
	t.Setenv("ADAMIC_BUILD_STORE", "traced="+os.Getenv("ADAMIC_BUILD_STORE"))
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	r.window(build.Build, func() { r.open(r.pid, "source/a.c") })
	if settlement := r.settle(t); len(settlement.Settled) != 1 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	Product(t, port, func(string) error { t.Fatal("the tree builder rebuilt a settled product"); return nil })
	if len(store.writes) != 0 || len(store.reads) != 0 {
		t.Fatalf("the tree builder reached the store: wrote %q, read %q", store.writes, store.reads)
	}
	file, _ := loadReads(r.cache, build.NameKey)
	key := file.Sets[0].Key
	store.put(t, key, map[string]string{"product": "stored"})
	// Another machine: an empty cache holding only the shipped read sets, untraced, reading the store.
	elsewhere := t.TempDir()
	shipped, err := os.ReadFile(readsPath(r.cache, build.NameKey))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(readsPath(elsewhere, build.NameKey), shipped, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_BUILD_STORE", strings.TrimPrefix(os.Getenv("ADAMIC_BUILD_STORE"), "traced="))
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", elsewhere)
	t.Setenv("ADAMIC_BUILD_TRACE", "")
	forgetHeldForLife()
	directory := Product(t, port, func(string) error { t.Fatal("rebuilt what the store holds"); return nil })
	if content, err := os.ReadFile(filepath.Join(directory, "product")); err != nil || string(content) != "stored" || directory != filepath.Join(elsewhere, key) {
		t.Fatalf("the fetched product at %s reads %q, %v", directory, content, err)
	}
	if len(store.writes) != 0 {
		t.Fatalf("a fetch wrote to the store: %q", store.writes)
	}
}

// A lookup the cache's read sets miss asks the store nothing: the store holds products by key, never read sets, and a
// name key the cache holds no read set for has no key to ask it for (Oct 14: probing refs/reads for one, 18 GETs a
// miss, was 96.7% of the store's billed reads, every one a 404). Find answers missing, and get builds here.
// Mutant: the store's read sets joined on a miss, as they were.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
func TestALookupTheCachesReadSetsMissAsksTheStoreNothing(t *testing.T) {
	store, _ := shared(t)
	nameKey := NameKey(repositoryRootForTest(t), thisPackage)
	findings, err := Find([]string{nameKey})
	if err != nil || len(findings) != 1 || findings[0].Outcome != FindingMissing {
		t.Fatalf("a name key the cache doesn't know: %+v, %v", findings, err)
	}
	if len(store.reads)+len(store.writes) != 0 {
		t.Fatalf("Find asked the store for what the cache missed: read %q, wrote %q", store.reads, store.writes)
	}
	var built bool
	Product(t, thisPackage, func(directory string) error {
		built = true
		return os.WriteFile(filepath.Join(directory, "product"), []byte("built"), 0o644)
	})
	if !built {
		t.Fatal("not built")
	}
	if len(store.reads)+len(store.writes) != 0 {
		t.Fatalf("a lookup asked the store for what the cache missed: read %q, wrote %q", store.reads, store.writes)
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

// Refs are split by trust: a candidate reads main's refs, then candidates'; a trusted reader reads only main's.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
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
	// A trusted reader never takes it: with only a candidate's ref stored, it builds its own.
	t.Setenv("ADAMIC_BUILD_STORE_TRUST", "main")
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
	recorded(t)
	if directory = Product(t, thisPackage, building("its own")); strings.Contains(directory, key) {
		t.Fatalf("a trusted reader took a candidate's product at %s", directory)
	}
}

// The shared tier is off unless ADAMIC_BUILD_STORE turns it on (Oct 9 20:38Z): unset, nothing is fetched, and a miss is
// built here.
// Not parallel: points the build cache and the store at this test through t.Setenv.
func TestTheSharedTierIsOffUnlessTurnedOn(t *testing.T) {
	for value, want := range map[string]string{"": "", "off": "", "on": defaultStore, "traced": defaultStore, "traced=http://127.0.0.1:9": "http://127.0.0.1:9", "http://127.0.0.1:9/": "http://127.0.0.1:9"} {
		t.Setenv("ADAMIC_BUILD_STORE", value)
		if got := storeAddress(); got != want {
			t.Fatalf("ADAMIC_BUILD_STORE=%q: the store is %q, want %q", value, got, want)
		}
	}
	cached(t)
	t.Setenv("ADAMIC_BUILD_STORE", "")
	// The read set gives a key the store would be asked for, were it on.
	recorded(t)
	if content, _ := os.ReadFile(filepath.Join(Product(t, thisPackage, building("built here")), "product")); string(content) != "built here" {
		t.Fatalf("an unset store's product reads %q", content)
	}
}
