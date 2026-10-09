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

// A store as Loom's serves it: blobs by sha256 and refs/build/<key>, read with no credential.
type fakeStore struct {
	mutex   sync.Mutex
	objects map[string][]byte
	reads   []string
	broken  string
}

func (s *fakeStore) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
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
	t.Setenv("ADAMIC_BUILD_AUDIT", "0")
	return store, log
}

func thisKey(t *testing.T) string {
	t.Helper()
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	return key(t, root, thisPackage)
}

func TestAStoredProductIsFetchedNotBuilt(t *testing.T) {
	store, log := shared(t)
	store.put(t, thisKey(t), map[string]string{"bin/checker": "the checker", "lib/a.o": "object"})
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
			plant(t, store, thisKey(t))
			_, err := Get(thisPackage, func(string) error { t.Fatal("built instead of failing"); return nil })
			var wrong poisonedError
			if !errors.As(err, &wrong) {
				t.Fatalf("got %v, want cache poisoning", err)
			}
			cache := os.Getenv("ADAMIC_BUILD_CACHE_DIR")
			if _, err := os.Stat(filepath.Join(cache, thisKey(t))); err == nil {
				t.Fatal("a poisoned product reached the local cache")
			}
		})
	}
}

func TestAnUnstoredOrUnreachableProductIsBuilt(t *testing.T) {
	for _, address := range []string{"", "http://127.0.0.1:1", "failing mid-fetch"} {
		t.Run(address, func(t *testing.T) {
			if address == "" {
				shared(t)
			} else if address == "failing mid-fetch" {
				// The ref and manifest read, then a blob's read fails: the store's trouble, not poisoning.
				store, _ := shared(t)
				product := store.put(t, thisKey(t), map[string]string{"product": "stored"})
				store.broken = "/blobs/" + product.Files[0].SHA256
			} else {
				cached(t)
				t.Setenv("ADAMIC_BUILD_STORE", address)
			}
			var built bool
			Product(t, thisPackage, func(directory string) error {
				built = true
				return os.WriteFile(filepath.Join(directory, "product"), []byte("built"), 0o644)
			})
			if !built {
				t.Fatal("not built")
			}
		})
	}
}

// A machine holding the write credential publishes what it built: every blob, then the manifest, then the ref, so
// the ref never names a blob the store doesn't hold. Another machine then fetches it.
func TestABuildPublishesBlobsThenItsRefAndAnotherMachineFetchesIt(t *testing.T) {
	store, _ := shared(t)
	calls := filepath.Join(t.TempDir(), "calls")
	program := filepath.Join(t.TempDir(), "put")
	// The stand-in writer stores what it's given in the fake store's directory, through a file the test reads back.
	if err := os.WriteFile(program, []byte("#!/bin/sh\nprintf '%s %s\\n' \"$1\" \"$2\" >> "+calls+"\nif [ \"$1\" = blob ]; then cp \"$3\" \""+calls+".$2\"; else printf '%s' \"$3\" > \""+calls+".ref\"; fi\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_BUILD_STORE_PUT", program)
	Product(t, thisPackage, func(directory string) error {
		os.MkdirAll(filepath.Join(directory, "bin"), 0o755)
		os.WriteFile(filepath.Join(directory, "bin", "checker"), []byte("the checker"), 0o755)
		return os.WriteFile(filepath.Join(directory, "notes"), []byte("notes"), 0o644)
	})
	lines, err := os.ReadFile(calls)
	if err != nil {
		t.Fatal(err)
	}
	order := strings.Split(strings.TrimSpace(string(lines)), "\n")
	if len(order) != 4 || !strings.HasPrefix(order[0], "blob ") || !strings.HasPrefix(order[2], "blob ") || order[3] != "ref build/"+thisKey(t) {
		t.Fatalf("writes: %q", order)
	}
	// Serve what was written, empty the local cache, and fetch it as another machine would.
	for _, line := range order[:3] {
		hash := strings.Fields(line)[1]
		content, err := os.ReadFile(calls + "." + hash)
		if err != nil {
			t.Fatal(err)
		}
		store.objects["/blobs/"+hash] = content
	}
	reference, _ := os.ReadFile(calls + ".ref")
	store.objects["/refs/build/"+thisKey(t)] = reference
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
	directory := Product(t, thisPackage, func(string) error { t.Fatal("rebuilt what the store holds"); return nil })
	info, err := os.Stat(filepath.Join(directory, "bin", "checker"))
	if err != nil || info.Mode()&0o111 == 0 {
		t.Fatalf("the fetched checker: %v, mode %v", err, info)
	}
}

// Audited, a fetched product is rebuilt and compared: the same product passes, a different one is poisoning.
func TestAnAuditedFetchThatDiffersFromARebuildIsPoisoning(t *testing.T) {
	store, log := shared(t)
	t.Setenv("ADAMIC_BUILD_AUDIT", "1")
	store.put(t, thisKey(t), map[string]string{"product": "built"})
	build := func(directory string) error {
		return os.WriteFile(filepath.Join(directory, "product"), []byte("built"), 0o644)
	}
	Product(t, thisPackage, build)
	if lines, _ := os.ReadFile(log); !strings.Contains(string(lines), " audited ") {
		t.Fatalf("census: %q", lines)
	}
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
	store.put(t, thisKey(t), map[string]string{"product": "a wrong product"})
	_, err := Get(thisPackage, build)
	var wrong poisonedError
	if !errors.As(err, &wrong) || !strings.Contains(err.Error(), "differs from a rebuild") {
		t.Fatalf("got %v, want poisoning", err)
	}
}
