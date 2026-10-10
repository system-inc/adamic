package buildcache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/rand/v2"
	"net/http"
	"os"
	"path"
	"path/filepath"
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
// until those are fixed, and gocacheprog is untouched). Only Workshop's tree builder publishes: it opts in with
// ADAMIC_BUILD_STORE=traced (traced=<address> for another store) and runs under cmd/traced, and only a settled traced
// build is published (#vt46geg), by a machine holding the write credential, through Loom's Worker: the gate boxes hold a publish-scoped token at
// ~/.loom/publish-token (ADAMIC_BUILD_STORE_TOKEN names another file). Every file's blob goes first, then the
// manifest's, then the ref, so a ref never names a blob the store doesn't hold. A ref never changes: the store
// refuses one naming a different manifest, which means a key that isn't honest, and that is said loudly.
const defaultStore = "https://artifacts.loom.system.inc"

// Where writes go (ADAMIC_BUILD_STORE_WRITE overrides it): /blobs/<sha256> and /refs/build/<key> under it.
const defaultWriter = "https://runs.loom.system.inc/public"

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

// publishing says whether this run is Workshop's tree builder, the one opted in to building traced and publishing what
// settles (ADAMIC_BUILD_STORE=traced, or traced=<address>).
func publishing() bool {
	setting := os.Getenv("ADAMIC_BUILD_STORE")
	return setting == "traced" || strings.HasPrefix(setting, "traced=")
}

var storeClient = &http.Client{Timeout: 5 * time.Minute}

// download reads one object, or errNotStored for a 404.
func download(address string) ([]byte, error) {
	response, err := storeClient.Get(address)
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
// could define a different build under main's key). refs/build is written only by main's own gate, which runs uncached
// and publishes what it built, with ADAMIC_BUILD_STORE_TRUST=main; everything else writes refs/build-candidate. A
// trusted reader reads refs/build alone; a candidate reads refs/build, then refs/build-candidate, so a candidate can
// only ever mislead candidates.
func trusted() bool { return os.Getenv("ADAMIC_BUILD_STORE_TRUST") == "main" }

func readNamespaces() []string {
	if trusted() {
		return []string{"build"}
	}
	return []string{"build", "build-candidate"}
}

func writeNamespace() string {
	if trusted() {
		return "build"
	}
	return "build-candidate"
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
	err := filepath.WalkDir(directory, func(file string, entry fs.DirEntry, err error) error {
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
		relative, err := filepath.Rel(directory, file)
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

// publishToken is the write credential, or "" on a machine that doesn't hold one.
func publishToken() string {
	path := os.Getenv("ADAMIC_BUILD_STORE_TOKEN")
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		path = filepath.Join(home, ".loom", "publish-token")
	}
	token, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(token))
}

// publish stores a product this machine built, when it holds the write credential: every file's blob, the
// manifest's, then the ref.
func publish(key, name, directory string) error {
	token := publishToken()
	if token == "" || storeAddress() == "" {
		return nil
	}
	writer := os.Getenv("ADAMIC_BUILD_STORE_WRITE")
	if writer == "" {
		writer = defaultWriter
	}
	writer = strings.TrimSuffix(writer, "/")
	product, err := describeProduct(key, name, directory)
	if err != nil {
		return err
	}
	for _, file := range product.Files {
		content, err := os.ReadFile(filepath.Join(directory, filepath.FromSlash(file.Path)))
		if err != nil {
			return err
		}
		if err = upload(writer+"/blobs/"+file.SHA256, token, content); err != nil {
			return err
		}
	}
	encoded, err := json.Marshal(product)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(encoded)
	manifestHash := hex.EncodeToString(sum[:])
	if err = upload(writer+"/blobs/"+manifestHash, token, encoded); err != nil {
		return err
	}
	if err = upload(writer+"/refs/"+writeNamespace()+"/"+key, token, []byte(manifestHash)); err != nil {
		if strings.Contains(err.Error(), "409") {
			return fmt.Errorf("the store already holds a different product for key %s (%s): the key isn't honest or the build isn't reproducible: %v", key[:12], name, err)
		}
		return err
	}
	return nil
}

// upload PUTs one object through Loom's Worker: 201 stored, 200 already held.
func upload(address, token string, content []byte) error {
	request, err := http.NewRequest(http.MethodPut, address, bytes.NewReader(content))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := storeClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 512))
		return fmt.Errorf("PUT %s answered %s: %s", address, response.Status, bytes.TrimSpace(body))
	}
	return nil
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

// Read sets travel with the products (#vt46geg): a machine with no read sets of its own finds a product in the store
// by its name key. R2 expires refs/ and blobs/ seven days after they were written, and the Worker never rewrites a
// ref or a blob it holds, so a read set stays findable only by being written again, new: the tree builder publishes
// nameKey's whole read set file as refs/<namespace>/<name key>.<day>.<n> (day since the Unix epoch, n from 0 within the
// day), naming a blob that holds the file and the day it was published, so each day's blob is new bytes and fresh. It
// publishes on every settlement and, on any use of a product, when the newest ref is more than readsFresh days old
// (refresh on use, as Loom's builder keeps its own refs; no bucket lifecycle exemption). A reader looks from tomorrow
// back over the lifetime and takes the newest day that has one, so it never depends on the oldest surviving ref.
// The namespaces split by trust as products' do: reads (main's; the tree builder runs with ADAMIC_BUILD_STORE_TRUST=main)
// and reads-candidate.
const (
	readsLifetime = 7
	readsFresh    = 5
)

type publishedReads struct {
	NameKey string    `json:"nameKey"`
	Day     int       `json:"day"`
	File    readsFile `json:"file"`
}

// readsClock is the clock read set refs are named by; a test moves it across days.
var readsClock = time.Now

func readsDay() int { return int(readsClock().Unix() / 86400) }

func readsNamespaces() []string {
	if trusted() {
		return []string{"reads"}
	}
	return []string{"reads", "reads-candidate"}
}

func readsWriteNamespace() string {
	if trusted() {
		return "reads"
	}
	return "reads-candidate"
}

func readsRef(store, namespace, nameKey string, day, index int) string {
	return fmt.Sprintf("%s/refs/%s/%s.%d.%d", store, namespace, nameKey, day, index)
}

// newestReads is the newest read set ref in namespace: the latest day from tomorrow (a writer's clock ahead of this
// one) back over the lifetime that has one, and the highest index that day; ok is false when there is none.
func newestReads(store, namespace, nameKey string) (day, index int, ok bool) {
	for day = readsDay() + 1; day >= readsDay()-readsLifetime; day-- {
		if _, err := download(readsRef(store, namespace, nameKey, day, 0)); err != nil {
			continue
		}
		for index = 0; ; index++ {
			if _, err := download(readsRef(store, namespace, nameKey, day, index+1)); err != nil {
				return day, index, true
			}
		}
	}
	return 0, 0, false
}

// fetchReads is nameKey's read sets in the store, newest first. An unreachable store, or a file that doesn't check, is
// no sets: a lookup then misses, never trusts it.
func fetchReads(nameKey string) []readSet {
	store := storeAddress()
	var sets []readSet
	for _, namespace := range readsNamespaces() {
		day, index, ok := newestReads(store, namespace, nameKey)
		if !ok {
			continue
		}
		reference, err := download(readsRef(store, namespace, nameKey, day, index))
		if err != nil {
			continue
		}
		content, err := blob(store, strings.TrimSpace(string(reference)))
		if err != nil {
			continue
		}
		var published publishedReads
		if json.Unmarshal(content, &published) == nil && published.NameKey == nameKey && published.File.Version == readsVersion {
			sets = append(sets, published.File.Sets...)
		}
	}
	return sets
}

// publishReads stores nameKey's read set file as today's next ref, after its blob.
func publishReads(cache, nameKey string) error {
	token := publishToken()
	if token == "" || storeAddress() == "" {
		return nil
	}
	file, err := loadReads(cache, nameKey)
	if err != nil || len(file.Sets) == 0 {
		return err
	}
	writer := os.Getenv("ADAMIC_BUILD_STORE_WRITE")
	if writer == "" {
		writer = defaultWriter
	}
	writer = strings.TrimSuffix(writer, "/")
	day := readsDay()
	encoded, err := json.Marshal(publishedReads{NameKey: nameKey, Day: day, File: file})
	if err != nil {
		return err
	}
	sum := sha256.Sum256(encoded)
	hash := hex.EncodeToString(sum[:])
	if err = upload(writer+"/blobs/"+hash, token, encoded); err != nil {
		return err
	}
	namespace := readsWriteNamespace()
	index := 0
	if newest, highest, ok := newestReads(storeAddress(), namespace, nameKey); ok && newest == day {
		index = highest + 1
	}
	for ; index < 1<<10; index++ {
		err = upload(readsRef(writer, namespace, nameKey, day, index), token, []byte(hash))
		if err == nil || !strings.Contains(err.Error(), "409") {
			return err
		}
	}
	return errors.New("no free read set index today")
}

var refreshedReads sync.Map

// refreshReads, on the tree builder's use of a product, publishes its read sets again when the store's newest is more
// than readsFresh days old or gone, once per process, so a product in use never outlives its read sets.
func refreshReads(cache, nameKey string) {
	if !publishing() || publishToken() == "" {
		return
	}
	if _, done := refreshedReads.LoadOrStore(nameKey, true); done {
		return
	}
	if day, _, ok := newestReads(storeAddress(), readsWriteNamespace(), nameKey); ok && day > readsDay()-readsFresh {
		return
	}
	if err := publishReads(cache, nameKey); err != nil {
		note("refresh read sets %s: %v", nameKey[:12], err)
	}
}
