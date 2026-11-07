// adamic-gocacheprog implements Go's GOCACHEPROG protocol (Go 1.24+).
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// These wire fields follow cmd/go/internal/cacheprog. Body is a separate
// base64 JSON string, present only when BodySize > 0.
type request struct {
	ID                 int64
	Command            string
	ActionID, OutputID []byte
	BodySize           int64
}

type response struct {
	ID            int64
	Err           string   `json:",omitempty"`
	KnownCommands []string `json:",omitempty"`
	Miss          bool     `json:",omitempty"`
	OutputID      []byte   `json:",omitempty"`
	Size          int64    `json:",omitempty"`
	DiskPath      string   `json:",omitempty"`
}

type record struct {
	OutputID []byte `json:"output_id"`
	Size     int64  `json:"size"`
	SHA256   string `json:"sha256"`
}

type cache struct {
	dir, store, writer, token, namespace string
	client                               *http.Client
	stderr                               io.Writer
	logMu                                sync.Mutex
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func newCache() (*cache, error) {
	root, userCacheDirectoryError := os.UserCacheDir()
	if userCacheDirectoryError != nil {
		root = os.TempDir()
	}
	dir, absolutePathError := filepath.Abs(env("ADAMIC_GOCACHE_DIR", filepath.Join(root, "adamic-gocache")))
	if absolutePathError != nil {
		return nil, absolutePathError
	}
	if createDirectoryError := os.MkdirAll(dir, 0700); createDirectoryError != nil {
		return nil, createDirectoryError
	}
	c := &cache{
		dir:       dir,
		store:     strings.TrimRight(env("ADAMIC_GOCACHE_STORE", "https://artifacts.loom.system.inc"), "/"),
		writer:    strings.TrimRight(env("ADAMIC_GOCACHE_WRITE", "https://runs.loom.system.inc/public"), "/"),
		namespace: "gocache-candidate",
		stderr:    os.Stderr,
		// The timeout covers reading the body, and a linked test binary is tens of megabytes: at 3 s a cold
		// instance on a slow link missed every large output. A miss is always safe, only slower.
		client: &http.Client{
			Timeout:       30 * time.Second,
			CheckRedirect: func(r *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
	if os.Getenv("ADAMIC_GOCACHE_TRUST") == "main" {
		c.namespace = "gocache"
	}
	if c.store == "off" {
		c.store = ""
	}
	if path := os.Getenv("ADAMIC_GOCACHE_TOKEN"); path != "" {
		b, readError := os.ReadFile(path)
		if readError == nil {
			c.token = strings.TrimSpace(string(b))
		} else {
			c.log("token unavailable: %v", readError)
		}
	}
	return c, nil
}

func (cache *cache) log(format string, args ...any) {
	cache.logMu.Lock()
	defer cache.logMu.Unlock()
	fmt.Fprintf(cache.stderr, "adamic-gocacheprog: "+format+"\n", args...)
}

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func validHash(s string) bool {
	b, decodeStringError := hex.DecodeString(s)
	return decodeStringError == nil && len(b) == sha256.Size && s == strings.ToLower(s)
}

func atomicFile(path string, b []byte) error {
	if createDirectoryError := os.MkdirAll(filepath.Dir(path), 0700); createDirectoryError != nil {
		return createDirectoryError
	}
	f, temporaryFileError := os.CreateTemp(filepath.Dir(path), ".tmp-")
	if temporaryFileError != nil {
		return temporaryFileError
	}
	defer os.Remove(f.Name())
	if _, writeError := f.Write(b); writeError != nil {
		f.Close()
		return writeError
	}
	if closeError := f.Close(); closeError != nil {
		return closeError
	}
	return os.Rename(f.Name(), path)
}

func (cache *cache) blobPath(hash string) string { return filepath.Join(cache.dir, "blobs", hash) }
func (cache *cache) refPath(ns, action string) string {
	return filepath.Join(cache.dir, "refs", ns, action)
}
func (cache *cache) check(b []byte, hash string) bool {
	if actual := digest(b); actual != hash {
		cache.log("cache poisoning: blob %s has sha256 %s", hash, actual)
		return false
	}
	return true
}

func (cache *cache) readRemote(path string, limit int64) ([]byte, bool) {
	response, getError := cache.client.Get(cache.store + path)
	if getError != nil {
		return nil, false
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, false
	}
	b, readError := io.ReadAll(io.LimitReader(response.Body, limit+1))
	return b, readError == nil && int64(len(b)) <= limit
}

// poisoned distinguishes corruption from an ordinary miss: once corruption is
// detected, the entire get misses instead of falling through to another ref.
func (cache *cache) blob(hash string, limit int64) (body []byte, ok, poisoned bool) {
	if !validHash(hash) {
		return nil, false, false
	}
	if b, readError := os.ReadFile(cache.blobPath(hash)); readError == nil {
		if !cache.check(b, hash) {
			return nil, false, true
		}
		if int64(len(b)) <= limit {
			return b, true, false
		}
		return nil, false, false
	}
	if cache.store == "" {
		return nil, false, false
	}
	b, ok := cache.readRemote("/blobs/"+hash, limit)
	if int64(len(b)) > limit {
		cache.log("cache poisoning: blob %s exceeds size limit %d", hash, limit)
		return nil, false, true
	}
	if !ok {
		return nil, false, false
	}
	if !cache.check(b, hash) {
		return nil, false, true
	}
	return b, true, false
}

func (cache *cache) get(req request) response {
	out := response{ID: req.ID, Miss: true}
	if len(req.ActionID) != sha256.Size {
		return out
	}
	action := hex.EncodeToString(req.ActionID)
	namespaces := []string{"gocache"}
	if cache.namespace != "gocache" {
		namespaces = append(namespaces, "gocache-candidate")
	}
	// Check all permitted local entries before making any network request.
	for _, remote := range []bool{false, true} {
		if remote && cache.store == "" {
			break
		}
		for _, ns := range namespaces {
			var reference []byte
			if remote {
				var ok bool
				reference, ok = cache.readRemote("/refs/"+ns+"/"+action, 256)
				if !ok {
					continue
				}
			} else {
				var readError error
				reference, readError = os.ReadFile(cache.refPath(ns, action))
				if readError != nil {
					continue
				}
			}
			hash := strings.TrimSpace(string(reference))
			if !validHash(hash) {
				continue
			}
			var rb []byte
			var ok bool
			if remote {
				var poisoned bool
				rb, ok, poisoned = cache.blob(hash, 4096)
				if poisoned {
					return out
				}
			} else {
				rb, _ = os.ReadFile(cache.blobPath(hash))
				if rb != nil && !cache.check(rb, hash) {
					return out
				}
				ok = rb != nil
			}
			if !ok {
				continue
			}
			var r record
			if json.Unmarshal(rb, &r) != nil ||
				len(r.OutputID) != sha256.Size ||
				r.Size < 0 ||
				r.Size == int64(^uint64(0)>>1) ||
				!validHash(r.SHA256) ||
				hex.EncodeToString(r.OutputID) != r.SHA256 {
				continue
			}
			var body []byte
			if remote {
				var poisoned bool
				body, ok, poisoned = cache.blob(r.SHA256, r.Size)
				if poisoned {
					return out
				}
			} else {
				body, _ = os.ReadFile(cache.blobPath(r.SHA256))
				if body != nil && !cache.check(body, r.SHA256) {
					return out
				}
				ok = body != nil
			}
			if !ok || int64(len(body)) != r.Size {
				continue
			}
			if remote {
				if atomicFileError := atomicFile(cache.blobPath(r.SHA256), body); atomicFileError != nil {
					continue
				}
				if atomicFileError := atomicFile(cache.blobPath(hash), rb); atomicFileError == nil {
					_ = atomicFile(cache.refPath(ns, action), []byte(hash))
				}
			}
			return response{ID: req.ID, OutputID: r.OutputID, Size: r.Size, DiskPath: cache.blobPath(r.SHA256)}
		}
	}
	return out
}

func (cache *cache) upload(path string, b []byte) bool {
	request, newRequestError := http.NewRequest(http.MethodPut, cache.writer+path, bytes.NewReader(b))
	if newRequestError != nil {
		cache.log("write: %v", newRequestError)
		return false
	}
	request.Header.Set("Authorization", "Bearer "+cache.token)
	response, requestError := cache.client.Do(request)
	if requestError != nil {
		cache.log("write: %v", requestError)
		return false
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode == 200 || response.StatusCode == 201 {
		return true
	}
	if response.StatusCode == 409 && strings.HasPrefix(path, "/refs/") {
		cache.log("ref conflict: %s already holds a different record; keeping local entry", path)
	} else {
		cache.log("write %s: HTTP %d", path, response.StatusCode)
	}
	return false
}

func (cache *cache) put(req request, body []byte) response {
	out := response{ID: req.ID}
	hash := digest(body)
	if len(req.ActionID) != sha256.Size ||
		hex.EncodeToString(req.OutputID) != hash ||
		req.BodySize != int64(len(body)) {
		out.Err = "invalid put IDs or body size"
		return out
	}
	r := record{OutputID: req.OutputID, Size: int64(len(body)), SHA256: hash}
	rb, _ := json.Marshal(r)
	rh := digest(rb)
	action := hex.EncodeToString(req.ActionID)
	if atomicFileError := atomicFile(cache.blobPath(hash), body); atomicFileError != nil {
		out.Err = atomicFileError.Error()
		return out
	}
	out.DiskPath = cache.blobPath(hash)
	if atomicFileError := atomicFile(cache.blobPath(rh), rb); atomicFileError != nil {
		cache.log("local record: %v", atomicFileError)
	} else if atomicFileError = atomicFile(cache.refPath(cache.namespace, action), []byte(rh)); atomicFileError != nil {
		cache.log("local ref: %v", atomicFileError)
	}
	if cache.store != "" && cache.token != "" && cache.upload("/blobs/"+hash, body) && cache.upload("/blobs/"+rh, rb) {
		cache.upload("/refs/"+cache.namespace+"/"+action, []byte(rh))
	}
	return out
}

// serve frames the stream by lines rather than with one json.Decoder. Go writes
// each request as a JSON line plus a blank line, and a put body as a quoted
// base64 line. If the body file changes size or fails to read mid-copy, go
// abandons the put without writing the closing quote, forgets its ID, and
// writes the next request straight after the truncated base64. That put must
// never be answered: go treats a response for an unknown ID as fatal.
func serve(c *cache, input io.Reader, output io.Writer) error {
	in, enc := bufio.NewReader(input), json.NewEncoder(output)
	if encodeError := enc.Encode(response{KnownCommands: []string{"get", "put", "close"}}); encodeError != nil {
		return encodeError
	}
	var waitGroup sync.WaitGroup
	var mutex sync.Mutex
	send := func(r response) {
		mutex.Lock()
		defer mutex.Unlock()
		if encodeError := enc.Encode(r); encodeError != nil {
			c.log("protocol output: %v", encodeError)
		}
	}
	// Bodies run to tens of megabytes, so lines are read whole, never through a
	// bufio.Scanner token limit. carry holds a request that followed an
	// abandoned body on the same line.
	var carry []byte
	next := func() ([]byte, error) {
		if line := carry; line != nil {
			carry = nil
			return line, nil
		}
		for {
			line, readBytesError := in.ReadBytes('\n')
			if line = bytes.TrimSpace(line); len(line) > 0 {
				return line, nil
			}
			if readBytesError != nil {
				return nil, readBytesError
			}
		}
	}
	slots := make(chan struct{}, 16)
	for {
		line, nextError := next()
		if nextError != nil {
			waitGroup.Wait()
			if nextError == io.EOF {
				return nil
			}
			return nextError
		}
		var req request
		if decodeError := json.Unmarshal(line, &req); decodeError != nil {
			waitGroup.Wait()
			return decodeError
		}
		var body []byte
		if req.Command == "put" && req.BodySize > 0 {
			line, nextError := next()
			if nextError != nil && nextError != io.EOF {
				waitGroup.Wait()
				return nextError
			}
			// Base64 never holds a quote or a brace, so a body line that is not one
			// closed string is a put go abandoned, possibly with EOF right behind it.
			brace := bytes.IndexByte(line, '{')
			if len(line) >= 2 && line[0] == '"' && line[len(line)-1] == '"' && brace < 0 {
				if decodeError := json.Unmarshal(line, &body); decodeError != nil {
					waitGroup.Wait()
					return decodeError
				}
			} else if len(line) == 0 || line[0] == '"' {
				arrived := max(len(line)-1, 0)
				if brace >= 0 {
					arrived, carry = brace-1, line[brace:]
				}
				c.log(
					"put %d abandoned mid-body by go: action %x, BodySize %d, %d base64 bytes arrived; dropped without a response",
					req.ID,
					req.ActionID,
					req.BodySize,
					arrived,
				)
				continue
			} else {
				waitGroup.Wait()
				return fmt.Errorf("put %d: body is not a JSON string", req.ID)
			}
		}
		if req.Command == "close" {
			waitGroup.Wait()
			send(response{ID: req.ID})
			return nil
		}
		slots <- struct{}{}
		waitGroup.Add(1)
		go func(req request, body []byte) {
			defer waitGroup.Done()
			defer func() { <-slots }()
			switch req.Command {
			case "get":
				send(c.get(req))
			case "put":
				send(c.put(req, body))
			default:
				send(response{ID: req.ID, Err: "unknown command"})
			}
		}(req, body)
	}
}

func main() {
	c, err := newCache()
	if err == nil {
		err = serve(c, os.Stdin, os.Stdout)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "adamic-gocacheprog:", err)
		os.Exit(1)
	}
}
