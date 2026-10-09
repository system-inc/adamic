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

func thisKey(t *testing.T) string {
	t.Helper()
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	return key(t, root, thisPackage)
}

// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
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

// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
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
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
func TestABuildPublishesBlobsThenItsRefAndAnotherMachineFetchesIt(t *testing.T) {
	store, _ := shared(t)
	build := func(directory string) error {
		os.MkdirAll(filepath.Join(directory, "bin"), 0o755)
		os.WriteFile(filepath.Join(directory, "bin", "checker"), []byte("the checker"), 0o755)
		return os.WriteFile(filepath.Join(directory, "notes"), []byte("notes"), 0o644)
	}
	// Without the token nothing is written.
	Product(t, thisPackage, build)
	if len(store.writes) != 0 {
		t.Fatalf("a machine without the token wrote %v", store.writes)
	}
	token := filepath.Join(t.TempDir(), "publish-token")
	os.WriteFile(token, []byte("gate-box-token\n"), 0o600)
	t.Setenv("ADAMIC_BUILD_STORE_TOKEN", token)
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
	Product(t, thisPackage, build)
	if len(store.writes) != 0 {
		t.Fatalf("a test uploaded inline: %v", store.writes)
	}
	entries, err := spoolEntries()
	if err != nil || len(entries) != 1 {
		t.Fatalf("spool entries: %v, %v", entries, err)
	}
	// The recorded trust survives a change in the publisher's environment.
	t.Setenv("ADAMIC_BUILD_STORE_TRUST", "main")
	runPublisher(t)
	t.Setenv("ADAMIC_BUILD_STORE_TRUST", "")
	entries, err = spoolEntries()
	if err != nil || len(entries) != 0 {
		t.Fatalf("spool after publish: %v, %v", entries, err)
	}
	// A candidate's build writes the candidate ref, never main's.
	if len(store.writes) != 4 || !strings.HasPrefix(store.writes[0], "/blobs/") || !strings.HasPrefix(store.writes[2], "/blobs/") || store.writes[3] != "/refs/build-candidate/"+thisKey(t) {
		t.Fatalf("writes: %q", store.writes)
	}
	// Another machine, an empty local cache: fetched, never built, exec bit kept.
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
	directory := Product(t, thisPackage, func(string) error { t.Fatal("rebuilt what the store holds"); return nil })
	info, err := os.Stat(filepath.Join(directory, "bin", "checker"))
	if err != nil || info.Mode()&0o111 == 0 {
		t.Fatalf("the fetched checker: %v, mode %v", err, info)
	}
}

// A ref never changes: a store already holding a different product for the key refuses, and the build says the
// key isn't honest without failing the product it built.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
func TestADifferentProductForAStoredKeyIsSaidLoudly(t *testing.T) {
	store, log := shared(t)
	token := filepath.Join(t.TempDir(), "publish-token")
	os.WriteFile(token, []byte("gate-box-token"), 0o600)
	t.Setenv("ADAMIC_BUILD_STORE_TOKEN", token)
	store.put(t, thisKey(t), map[string]string{"product": "another machine's"})
	store.objects["/refs/build-candidate/"+thisKey(t)] = store.objects["/refs/build/"+thisKey(t)]
	// Reads can't reach the store, so this machine builds its own, then tries to publish it under the same key.
	t.Setenv("ADAMIC_BUILD_STORE", "http://127.0.0.1:1")
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
	directory := Product(t, thisPackage, func(directory string) error {
		return os.WriteFile(filepath.Join(directory, "product"), []byte("this machine's"), 0o644)
	})
	if content, _ := os.ReadFile(filepath.Join(directory, "product")); string(content) != "this machine's" {
		t.Fatalf("the built product reads %q", content)
	}
	if err := PublishSpool(); err == nil || !strings.Contains(err.Error(), "isn't honest") {
		t.Fatalf("publish: %v", err)
	} else {
		note("%v", err)
	}
	if lines, _ := os.ReadFile(log); !strings.Contains(string(lines), "isn't honest") {
		t.Fatalf("log: %q", lines)
	}
}

// Fetched products stay usable while their queued rebuilds run as a separate audit unit.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
func TestAnAuditedFetchThatDiffersFromARebuildIsPoisoning(t *testing.T) {
	for _, content := range []string{"built", "a wrong product"} {
		t.Run(content, func(t *testing.T) {
			store, log := shared(t)
			t.Setenv("ADAMIC_BUILD_AUDIT", "1")
			store.put(t, thisKey(t), map[string]string{"product": content})
			var builds int
			directory := Product(t, thisPackage, func(directory string) error {
				builds++
				return os.WriteFile(filepath.Join(directory, "product"), []byte("built"), 0o644)
			})
			if builds != 0 {
				t.Fatalf("a fetched product was rebuilt inline: %d builds", builds)
			}
			if got, err := os.ReadFile(filepath.Join(directory, "product")); err != nil || string(got) != content {
				t.Fatalf("the test lost its fetched product: %q, %v", got, err)
			}
			entries, err := auditEntries()
			if err != nil || len(entries) != 1 {
				t.Fatalf("audit entries: %v, %v", entries, err)
			}
			if lines, _ := os.ReadFile(log); !strings.Contains(string(lines), " audit-queued ") {
				t.Fatalf("census: %q", lines)
			}
			output, err := publisherOutput(t, "-audit")
			if content == "built" {
				if err != nil {
					t.Fatalf("honest audit: %v\n%s", err, output)
				}
				entries, err = auditEntries()
				if err != nil || len(entries) != 0 {
					t.Fatalf("successful audit left entries: %v, %v", entries, err)
				}
			} else {
				if err == nil || !strings.Contains(string(output), "cache poisoning:") ||
					!strings.Contains(string(output), "differs from a rebuild") || !strings.Contains(string(output), thisKey(t)[:12]) {
					t.Fatalf("wrong product's audit: %v\n%s", err, output)
				}
				entries, err = auditEntries()
				if err != nil || len(entries) != 1 {
					t.Fatalf("poisoning witness was lost: %v, %v", entries, err)
				}
			}
		})
	}
}

// Refs are split by trust: main's gate reads only main's refs and publishes there from its uncached build; a candidate
// reads main's, then candidates'.
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
func TestRefsAreSplitByTrust(t *testing.T) {
	store, _ := shared(t)
	token := filepath.Join(t.TempDir(), "publish-token")
	os.WriteFile(token, []byte("gate-box-token"), 0o600)
	t.Setenv("ADAMIC_BUILD_STORE_TOKEN", token)
	store.put(t, thisKey(t), map[string]string{"product": "a candidate's"})
	store.objects["/refs/build-candidate/"+thisKey(t)] = store.objects["/refs/build/"+thisKey(t)]
	delete(store.objects, "/refs/build/"+thisKey(t))
	build := func(directory string) error {
		return os.WriteFile(filepath.Join(directory, "product"), []byte("main's"), 0o644)
	}
	// A candidate reads the candidate ref.
	directory := Product(t, thisPackage, func(string) error { t.Fatal("a candidate built what a candidate ref holds"); return nil })
	if content, _ := os.ReadFile(filepath.Join(directory, "product")); string(content) != "a candidate's" {
		t.Fatalf("a candidate read %q", content)
	}
	// A trusted reader never does: with only a candidate's ref stored, it builds.
	t.Setenv("ADAMIC_BUILD_STORE_TRUST", "main")
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
	var built bool
	Product(t, thisPackage, func(directory string) error { built = true; return build(directory) })
	if !built {
		t.Fatal("a trusted reader took a candidate's product")
	}
	// Main's gate, uncached, builds and publishes to main's ref.
	t.Setenv("ADAMIC_BUILD_CACHE", "off")
	directory = Product(t, thisPackage, build)
	if content, _ := os.ReadFile(filepath.Join(directory, "product")); string(content) != "main's" {
		t.Fatalf("main's gate read %q", content)
	}
	if last := store.writes[len(store.writes)-1]; last != "/refs/build/"+thisKey(t) {
		t.Fatalf("main's gate wrote %v", store.writes)
	}
	// With the cache on, a trusted reader reads main's ref and ignores candidates'.
	t.Setenv("ADAMIC_BUILD_CACHE", "")
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
	directory = Product(t, thisPackage, func(string) error { t.Fatal("rebuilt what main's ref holds"); return nil })
	if content, _ := os.ReadFile(filepath.Join(directory, "product")); string(content) != "main's" {
		t.Fatalf("a trusted reader read %q", content)
	}
	// Uncached and untrusted, nothing is written.
	t.Setenv("ADAMIC_BUILD_STORE_TRUST", "")
	t.Setenv("ADAMIC_BUILD_CACHE", "off")
	writes := len(store.writes)
	Product(t, thisPackage, build)
	if len(store.writes) != writes {
		t.Fatalf("an uncached candidate wrote %v", store.writes[writes:])
	}
}
