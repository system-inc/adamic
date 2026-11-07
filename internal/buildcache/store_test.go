package buildcache

import (
	"bytes"
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

func (store *fakeStore) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if request.Method != http.MethodGet {
		store.writes = append(store.writes, request.Method+" "+request.URL.Path)
		http.Error(writer, "read only", http.StatusMethodNotAllowed)
		return
	}
	store.reads = append(store.reads, request.URL.Path)
	if request.URL.Path == store.broken {
		http.Error(writer, "upstream timeout", http.StatusBadGateway)
		return
	}
	content, ok := store.objects[request.URL.Path]
	if !ok {
		http.NotFound(writer, request)
		return
	}
	writer.Write(content)
}

func (store *fakeStore) blob(content []byte) string {
	sum := sha256.Sum256(content)
	hash := hex.EncodeToString(sum[:])
	store.objects["/blobs/"+hash] = content
	return hash
}

// put stores a product for key as Loom's store holds one, and returns its manifest.
func (store *fakeStore) put(test *testing.T, key string, files map[string]string) manifest {
	test.Helper()
	product := manifest{Version: 1, Key: key, Name: "buildcache test"}
	for name, content := range files {
		product.Files = append(
			product.Files,
			manifestFile{Path: name, SHA256: store.blob([]byte(content)), Size: int64(len(content)), Mode: 0o644},
		)
	}
	store.ref(test, key, product)
	return product
}

func (store *fakeStore) ref(test *testing.T, key string, product manifest) {
	test.Helper()
	encoded, encodeError := json.Marshal(product)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	store.objects["/refs/build/"+key] = []byte(store.blob(encoded) + "\n")
}

func shared(test *testing.T) (*fakeStore, string) {
	test.Helper()
	_, log := cached(test)
	store := &fakeStore{objects: map[string][]byte{}}
	server := httptest.NewServer(store)
	test.Cleanup(server.Close)
	test.Setenv("ADAMIC_BUILD_STORE", server.URL)
	test.Setenv("ADAMIC_BUILD_AUDIT", "0")
	return store, log
}

// recorded gives thisPackage a read set in the current cache, as a traced build of it on Workshop would have (it read
// buildcache.go), and returns the key it gives on this tree: the key a store holds its product under.
func recorded(test *testing.T) string {
	test.Helper()
	root := repositoryRootForTest(test)
	cache, cacheDirectoryError := cacheDirectory()
	if cacheDirectoryError != nil {
		test.Fatal(cacheDirectoryError)
	}
	nameKey := NameKey(root, thisPackage)
	set := readSet{Entries: []readEntry{{Kind: "content", Path: "internal/buildcache/buildcache.go"}}}
	evaluated, evaluateError := evaluate(root, cache, nameKey, set, 0, nil)
	if evaluateError != nil {
		test.Fatal(evaluateError)
	}
	set.Key = evaluated.key
	if recordReadsError := recordReads(cache, nameKey, thisPackage.Name, set); recordReadsError != nil {
		test.Fatal(recordReadsError)
	}
	return evaluated.key
}

// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through test.Setenv.
func TestAStoredProductIsFetchedNotBuilt(test *testing.T) {
	store, log := shared(test)
	store.put(test, recorded(test), map[string]string{"bin/checker": "the checker", "lib/a.o": "object"})
	directory := Product(
		test,
		thisPackage,
		func(string) error { test.Fatal("built a product the store holds"); return nil },
	)
	for name, want := range map[string]string{"bin/checker": "the checker", "lib/a.o": "object"} {
		if content, readError := os.ReadFile(
			filepath.Join(directory, name),
		); readError != nil || string(content) != want {
			test.Fatalf("%s reads %q, %v", name, content, readError)
		}
	}
	if lines, _ := os.ReadFile(log); !strings.Contains(string(lines), " fetched ") {
		test.Fatalf("census: %q", lines)
	}
	// Fetched once, then a local hit: the store isn't asked again.
	reads := len(store.reads)
	Product(test, thisPackage, func(string) error { test.Fatal("rebuilt"); return nil })
	if len(store.reads) != reads {
		test.Fatalf("a local hit read the store again: %v", store.reads[reads:])
	}
}

// Anything stored that doesn't check fails loudly as poisoning, and nothing reaches the local cache.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through test.Setenv.
func TestAStoredProductThatDoesntCheckIsPoisoning(test *testing.T) {
	cases := map[string]func(test *testing.T, store *fakeStore, key string){
		"a blob that isn't its hash": func(test *testing.T, store *fakeStore, key string) {
			product := store.put(test, key, map[string]string{"bin/checker": "the checker"})
			// The same length, so only the blob's hash can tell.
			store.objects["/blobs/"+product.Files[0].SHA256] = []byte("the checkeR")
		},
		"a manifest for another key": func(test *testing.T, store *fakeStore, key string) {
			product := store.put(test, key, map[string]string{"bin/checker": "the checker"})
			product.Key = strings.Repeat("0", 64)
			store.ref(test, key, product)
		},
		"a path outside the product": func(test *testing.T, store *fakeStore, key string) {
			product := store.put(test, key, map[string]string{"bin/checker": "the checker"})
			product.Files[0].Path = "../escape"
			store.ref(test, key, product)
		},
		"a size that isn't the blob's": func(test *testing.T, store *fakeStore, key string) {
			product := store.put(test, key, map[string]string{"bin/checker": "the checker"})
			product.Files[0].Size++
			store.ref(test, key, product)
		},
		"a ref naming a missing manifest": func(test *testing.T, store *fakeStore, key string) {
			store.objects["/refs/build/"+key] = []byte(strings.Repeat("a", 64))
		},
		"a ref that isn't a hash": func(test *testing.T, store *fakeStore, key string) {
			store.objects["/refs/build/"+key] = []byte("latest")
		},
	}
	for name, plant := range cases {
		test.Run(name, func(test *testing.T) {
			store, _ := shared(test)
			key := recorded(test)
			plant(test, store, key)
			_, getError := Get(thisPackage, func(string) error { test.Fatal("built instead of failing"); return nil })
			var wrong poisonedError
			if !errors.As(getError, &wrong) {
				test.Fatalf("got %v, want cache poisoning", getError)
			}
			if _, statError := os.Stat(filepath.Join(os.Getenv("ADAMIC_BUILD_CACHE_DIR"), key)); statError == nil {
				test.Fatal("a poisoned product reached the local cache")
			}
		})
	}
}

// A product the store doesn't hold, or can't serve, is built here, for this machine alone, and nothing is written to
// the store.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through test.Setenv.
func TestAnUnstoredOrUnreachableProductIsBuiltHere(test *testing.T) {
	for _, address := range []string{"", "http://127.0.0.1:1", "failing mid-fetch"} {
		test.Run(address, func(test *testing.T) {
			var store *fakeStore
			if address == "" {
				store, _ = shared(test)
				recorded(test)
			} else if address == "failing mid-fetch" {
				// The ref and manifest read, then a blob's read fails: the store's trouble, not poisoning.
				store, _ = shared(test)
				product := store.put(test, recorded(test), map[string]string{"product": "stored"})
				store.broken = "/blobs/" + product.Files[0].SHA256
			} else {
				cached(test)
				test.Setenv("ADAMIC_BUILD_STORE", address)
				recorded(test)
			}
			var built bool
			Product(test, thisPackage, func(directory string) error {
				built = true
				return os.WriteFile(filepath.Join(directory, "product"), []byte("built"), 0o644)
			})
			if !built {
				test.Fatal("not built")
			}
			if store != nil && len(store.writes) != 0 {
				test.Fatalf("an untraced build wrote to the store: %v", store.writes)
			}
		})
	}
}

// The tree builder's settled product stays in its cache, read sets beside it, and nothing reaches the store: no write,
// and no read set asked for. Another machine, given those read sets as Loom's tree index ships them, fetches the
// product from the store by the key they give.
// Not parallel: newRig.
func TestTheTreeBuilderWritesNothingAndARunnerFetchesByTheShippedReadSets(test *testing.T) {
	r := newRig(test)
	store, _ := shared(test)
	test.Setenv("ADAMIC_BUILD_CACHE_DIR", r.cache)
	test.Setenv("ADAMIC_BUILD_STORE", "traced="+os.Getenv("ADAMIC_BUILD_STORE"))
	Product(test, port, building("built"))
	build := r.built(test)["port"]
	r.window(build.Build, func() { r.open(r.pid, "source/a.c") })
	if settlement := r.settle(test); len(settlement.Settled) != 1 {
		test.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	Product(test, port, func(string) error { test.Fatal("the tree builder rebuilt a settled product"); return nil })
	if len(store.writes) != 0 || len(store.reads) != 0 {
		test.Fatalf("the tree builder reached the store: wrote %q, read %q", store.writes, store.reads)
	}
	file, _ := loadReads(r.cache, build.NameKey)
	key := file.Sets[0].Key
	store.put(test, key, map[string]string{"product": "stored"})
	// Another machine: an empty cache holding only the shipped read sets, untraced, reading the store.
	elsewhere := test.TempDir()
	shipped, readError := os.ReadFile(readsPath(r.cache, build.NameKey))
	if readError != nil {
		test.Fatal(readError)
	}
	if writeError := os.WriteFile(readsPath(elsewhere, build.NameKey), shipped, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	test.Setenv("ADAMIC_BUILD_STORE", strings.TrimPrefix(os.Getenv("ADAMIC_BUILD_STORE"), "traced="))
	test.Setenv("ADAMIC_BUILD_CACHE_DIR", elsewhere)
	test.Setenv("ADAMIC_BUILD_TRACE", "")
	forgetHeldForLife()
	directory := Product(test, port, func(string) error { test.Fatal("rebuilt what the store holds"); return nil })
	if content, readError := os.ReadFile(
		filepath.Join(directory, "product"),
	); readError != nil || string(content) != "stored" || directory != filepath.Join(elsewhere, key) {
		test.Fatalf("the fetched product at %s reads %q, %v", directory, content, readError)
	}
	if len(store.writes) != 0 {
		test.Fatalf("a fetch wrote to the store: %q", store.writes)
	}
}

// A lookup the cache's read sets miss asks the store nothing: the store holds products by key, never read sets, and a
// name key the cache holds no read set for has no key to ask it for (Oct 14: probing refs/reads for one, 18 GETs a
// miss, was 96.7% of the store's billed reads, every one a 404). Find answers missing, and get builds here.
// Mutant: the store's read sets joined on a miss, as they were.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through test.Setenv.
func TestALookupTheCachesReadSetsMissAsksTheStoreNothing(test *testing.T) {
	store, _ := shared(test)
	nameKey := NameKey(repositoryRootForTest(test), thisPackage)
	findings, findError := Find([]string{nameKey})
	if findError != nil || len(findings) != 1 || findings[0].Outcome != FindingMissing {
		test.Fatalf("a name key the cache doesn't know: %+v, %v", findings, findError)
	}
	if len(store.reads)+len(store.writes) != 0 {
		test.Fatalf("Find asked the store for what the cache missed: read %q, wrote %q", store.reads, store.writes)
	}
	var built bool
	Product(test, thisPackage, func(directory string) error {
		built = true
		return os.WriteFile(filepath.Join(directory, "product"), []byte("built"), 0o644)
	})
	if !built {
		test.Fatal("not built")
	}
	if len(store.reads)+len(store.writes) != 0 {
		test.Fatalf("a lookup asked the store for what the cache missed: read %q, wrote %q", store.reads, store.writes)
	}
}

// Audited, a fetched product is rebuilt and compared, where building may (under a trace): the same product passes, a
// different one is poisoning.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through test.Setenv.
func TestAnAuditedFetchThatDiffersFromARebuildIsPoisoning(test *testing.T) {
	store, log := shared(test)
	test.Setenv("ADAMIC_BUILD_AUDIT", "1")
	test.Setenv("ADAMIC_BUILD_TRACE", test.TempDir())
	store.put(test, recorded(test), map[string]string{"product": "built"})
	build := func(directory string) error {
		return os.WriteFile(filepath.Join(directory, "product"), []byte("built"), 0o644)
	}
	Product(test, thisPackage, build)
	if lines, _ := os.ReadFile(log); !strings.Contains(string(lines), " audited ") {
		test.Fatalf("census: %q", lines)
	}
	test.Setenv("ADAMIC_BUILD_CACHE_DIR", test.TempDir())
	store.put(test, recorded(test), map[string]string{"product": "a wrong product"})
	_, getError := Get(thisPackage, build)
	var wrong poisonedError
	if !errors.As(getError, &wrong) || !strings.Contains(getError.Error(), "differs from a rebuild") {
		test.Fatalf("got %v, want poisoning", getError)
	}
}

// Refs are split by trust: a candidate reads main's refs, then candidates'; a trusted reader reads only main's.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through test.Setenv.
func TestRefsAreSplitByTrust(test *testing.T) {
	store, _ := shared(test)
	key := recorded(test)
	store.put(test, key, map[string]string{"product": "a candidate's"})
	store.objects["/refs/build-candidate/"+key] = store.objects["/refs/build/"+key]
	delete(store.objects, "/refs/build/"+key)
	directory := Product(
		test,
		thisPackage,
		func(string) error { test.Fatal("a candidate built what a candidate ref holds"); return nil },
	)
	if content, _ := os.ReadFile(filepath.Join(directory, "product")); string(content) != "a candidate's" {
		test.Fatalf("a candidate read %q", content)
	}
	// A trusted reader never takes it: with only a candidate's ref stored, it builds its own.
	test.Setenv("ADAMIC_BUILD_STORE_TRUST", "main")
	test.Setenv("ADAMIC_BUILD_CACHE_DIR", test.TempDir())
	recorded(test)
	if directory = Product(test, thisPackage, building("its own")); strings.Contains(directory, key) {
		test.Fatalf("a trusted reader took a candidate's product at %s", directory)
	}
}

// The shared tier is off unless ADAMIC_BUILD_STORE turns it on (Oct 9 20:38Z): unset, nothing is fetched, and a miss is
// built here.
// Not parallel: points the build cache and the store at this test through test.Setenv.
func TestTheSharedTierIsOffUnlessTurnedOn(test *testing.T) {
	for value, want := range map[string]string{
		"":                          "",
		"off":                       "",
		"on":                        defaultStore,
		"traced":                    defaultStore,
		"traced=http://127.0.0.1:9": "http://127.0.0.1:9",
		"http://127.0.0.1:9/":       "http://127.0.0.1:9",
	} {
		test.Setenv("ADAMIC_BUILD_STORE", value)
		if got := storeAddress(); got != want {
			test.Fatalf("ADAMIC_BUILD_STORE=%q: the store is %q, want %q", value, got, want)
		}
	}
	cached(test)
	test.Setenv("ADAMIC_BUILD_STORE", "")
	// The read set gives a key the store would be asked for, were it on.
	recorded(test)
	if content, _ := os.ReadFile(
		filepath.Join(Product(test, thisPackage, building("built here")), "product"),
	); string(content) != "built here" {
		test.Fatalf("an unset store's product reads %q", content)
	}
}

// staleFinder fills the cache's read sets for thisPackage as a file whose oldest set (the last of keptSets) is the one
// that values to a product on this tree, though it stored another key, the shape #2dkn3fb left: three newer sets that
// find nothing here, each reading a different file. It returns the key the oldest set gives here.
func staleFinder(test *testing.T) (string, string, string) {
	test.Helper()
	root := repositoryRootForTest(test)
	cache, cacheDirectoryError := cacheDirectory()
	if cacheDirectoryError != nil {
		test.Fatal(cacheDirectoryError)
	}
	nameKey := NameKey(root, thisPackage)
	finder := readSet{
		Key:     strings.Repeat("0", 64),
		Entries: []readEntry{{Kind: "content", Path: "internal/buildcache/buildcache.go"}},
	}
	evaluated, evaluateError := evaluate(root, cache, nameKey, finder, 0, nil)
	if evaluateError != nil {
		test.Fatal(evaluateError)
	}
	for _, set := range []readSet{
		finder,
		readOnly(test, root, cache, nameKey, "internal/buildcache/store.go"),
		readOnly(test, root, cache, nameKey, "internal/buildcache/find.go"),
		readOnly(test, root, cache, nameKey, "internal/buildcache/reads.go"),
	} {
		if recordReadsError := recordReads(cache, nameKey, thisPackage.Name, set); recordReadsError != nil {
			test.Fatal(recordReadsError)
		}
	}
	return cache, nameKey, evaluated.key
}

// readOnly is a set that reads path alone, under the key it gives here, which names no product in the cache.
func readOnly(test *testing.T, root, cache, nameKey, path string) readSet {
	test.Helper()
	set := readSet{Entries: []readEntry{{Kind: "content", Path: path}}}
	evaluated, evaluateError := evaluate(root, cache, nameKey, set, 0, nil)
	if evaluateError != nil {
		test.Fatal(evaluateError)
	}
	set.Key = evaluated.key
	return set
}

// pushOut records one set newer than every set in nameKey's file, as a later settlement of the product would.
func pushOut(test *testing.T, cache, nameKey string) {
	test.Helper()
	if recordReadsError := recordReads(
		cache,
		nameKey,
		thisPackage.Name,
		readOnly(test, repositoryRootForTest(test), cache, nameKey, "internal/buildcache/settle.go"),
	); recordReadsError != nil {
		test.Fatal(recordReadsError)
	}
}

// A lookup that finds or fetches a product by a set that stored another key records that set again, first, under the
// key it found (#xaz1487, system_adamic's yes on #2dkn3fb), so a newer set that would have pushed it past the file's
// keptSets leaves it, and the product is still found. Found in the cache, fetched from the store and found by Find
// alike. Mutants that each fail it: the set never recorded again (get's hit, get's fetch, Find's found); recorded again
// under the key it stored.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through test.Setenv.
func TestTheSetThatFoundAProductUnderAnotherKeyIsRecordedAgainFirst(test *testing.T) {
	for _, how := range []string{"found in the cache", "fetched from the store", "found by Find"} {
		test.Run(how, func(test *testing.T) {
			store, _ := shared(test)
			rerecorded.Clear()
			cache, nameKey, key := staleFinder(test)
			switch how {
			case "fetched from the store":
				store.put(test, key, map[string]string{"product": "stored"})
				Product(test, thisPackage, func(string) error { test.Fatal("built what the store holds"); return nil })
			case "found in the cache":
				if createDirectoryError := os.MkdirAll(filepath.Join(cache, key), 0o755); createDirectoryError != nil {
					test.Fatal(createDirectoryError)
				}
				os.WriteFile(filepath.Join(cache, key, "product"), []byte("cached"), 0o644)
				Product(test, thisPackage, func(string) error { test.Fatal("built what the cache holds"); return nil })
			case "found by Find":
				os.MkdirAll(filepath.Join(cache, key), 0o755)
				os.WriteFile(filepath.Join(cache, key, "product"), []byte("cached"), 0o644)
				if findings, findError := Find(
					[]string{nameKey},
				); findError != nil || findings[0].Outcome != FindingFound || findings[0].Key != key {
					test.Fatalf("Find: %+v, %v", findings, findError)
				}
			}
			file, loadReadsError := loadReads(cache, nameKey)
			if loadReadsError != nil ||
				len(file.Sets) != keptSets ||
				file.Sets[0].Key != key ||
				file.Sets[0].Entries[0].Path != "internal/buildcache/buildcache.go" {
				test.Fatalf(
					"the file's sets after the lookup: %+v, %v; want the finding set first under %s",
					file.Sets,
					loadReadsError,
					key[:12],
				)
			}
			if file.Sets[0].Entries[0].Value == "" {
				test.Fatal("the set recorded again holds no value to name a miss by")
			}
			pushOut(test, cache, nameKey)
			directory := Product(
				test,
				thisPackage,
				func(string) error { test.Fatal("the product was lost once a newer set was recorded"); return nil },
			)
			if directory != filepath.Join(cache, key) {
				test.Fatalf("found at %s, want %s", directory, filepath.Join(cache, key))
			}
		})
	}
}

// A set that found its product under the key it stored, and any lookup in read mode, leave the file as it was: read
// mode never writes the cache's read sets, and a set that names its own product needs nothing recorded. Mutants that
// each fail it: read mode not asked; a set recorded again whatever key it gave.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through test.Setenv.
func TestALookupThatFoundItsStoredKeyOrReadsOnlyLeavesTheFile(test *testing.T) {
	test.Run("read mode", func(test *testing.T) {
		shared(test)
		rerecorded.Clear()
		cache, nameKey, key := staleFinder(test)
		os.MkdirAll(filepath.Join(cache, key), 0o755)
		os.WriteFile(filepath.Join(cache, key, "product"), []byte("cached"), 0o644)
		before, _ := os.ReadFile(readsPath(cache, nameKey))
		test.Setenv("ADAMIC_BUILD_CACHE", "read")
		forgetHeldForLife()
		test.Cleanup(forgetHeldForLife)
		Product(test, thisPackage, func(string) error { test.Fatal("built in read mode"); return nil })
		if after, _ := os.ReadFile(readsPath(cache, nameKey)); !bytes.Equal(before, after) {
			test.Fatal("read mode wrote the read sets")
		}
	})
	test.Run("its own key", func(test *testing.T) {
		shared(test)
		rerecorded.Clear()
		key := recorded(test)
		cache, _ := cacheDirectory()
		nameKey := NameKey(repositoryRootForTest(test), thisPackage)
		os.MkdirAll(filepath.Join(cache, key), 0o755)
		os.WriteFile(filepath.Join(cache, key, "product"), []byte("cached"), 0o644)
		before, _ := os.ReadFile(readsPath(cache, nameKey))
		Product(test, thisPackage, func(string) error { test.Fatal("built what the cache holds"); return nil })
		if after, _ := os.ReadFile(readsPath(cache, nameKey)); !bytes.Equal(before, after) {
			test.Fatal("a set that found its own key was recorded again")
		}
	})
}
