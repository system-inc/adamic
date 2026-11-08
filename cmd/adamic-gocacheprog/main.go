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
	OutputId []byte `json:"output_id"`
	Size     int64  `json:"size"`
	Sha256   string `json:"sha256"`
}

type cache struct {
	dir, store, writer, token, namespace string
	client                               *http.Client
	stderr                               io.Writer
	logMu                                sync.Mutex
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
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
	cache := &cache{
		dir:       dir,
		store:     strings.TrimRight(env("ADAMIC_GOCACHE_STORE", "https://artifacts.loom.system.inc"), "/"),
		writer:    strings.TrimRight(env("ADAMIC_GOCACHE_WRITE", "https://runs.loom.system.inc/public"), "/"),
		namespace: "gocache-candidate",
		stderr:    os.Stderr,
		// The timeout covers reading the body, and a linked test binary is tens of megabytes: at 3 s a cold
		// instance on a slow link missed every large output. A miss is always safe, only slower.
		client: &http.Client{
			Timeout:       30 * time.Second,
			CheckRedirect: func(request *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
	if os.Getenv("ADAMIC_GOCACHE_TRUST") == "main" {
		cache.namespace = "gocache"
	}
	if cache.store == "off" {
		cache.store = ""
	}
	if path := os.Getenv("ADAMIC_GOCACHE_TOKEN"); path != "" {
		content, readError := os.ReadFile(path)
		if readError == nil {
			cache.token = strings.TrimSpace(string(content))
		} else {
			cache.log("token unavailable: %v", readError)
		}
	}
	return cache, nil
}

func (cache *cache) log(format string, args ...any) {
	cache.logMu.Lock()
	defer cache.logMu.Unlock()
	fmt.Fprintf(cache.stderr, "adamic-gocacheprog: "+format+"\n", args...)
}

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func validHash(text string) bool {
	b, decodeStringError := hex.DecodeString(text)
	return decodeStringError == nil && len(b) == sha256.Size && text == strings.ToLower(text)
}

func atomicFile(path string, b []byte) error {
	if createDirectoryError := os.MkdirAll(filepath.Dir(path), 0700); createDirectoryError != nil {
		return createDirectoryError
	}
	file, temporaryFileError := os.CreateTemp(filepath.Dir(path), ".tmp-")
	if temporaryFileError != nil {
		return temporaryFileError
	}
	defer os.Remove(file.Name())
	if _, writeError := file.Write(b); writeError != nil {
		file.Close()
		return writeError
	}
	if closeError := file.Close(); closeError != nil {
		return closeError
	}
	return os.Rename(file.Name(), path)
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
	content, readError := io.ReadAll(io.LimitReader(response.Body, limit+1))
	return content, readError == nil && int64(len(content)) <= limit
}

// poisoned distinguishes corruption from an ordinary miss: once corruption is
// detected, the entire get misses instead of falling through to another ref.
func (cache *cache) blob(hash string, limit int64) (body []byte, ok, poisoned bool) {
	if !validHash(hash) {
		return nil, false, false
	}
	if content, readError := os.ReadFile(cache.blobPath(hash)); readError == nil {
		if !cache.check(content, hash) {
			return nil, false, true
		}
		if int64(len(content)) <= limit {
			return content, true, false
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
			var record record
			if json.Unmarshal(rb, &record) != nil ||
				len(record.OutputId) != sha256.Size ||
				record.Size < 0 ||
				record.Size == int64(^uint64(0)>>1) ||
				!validHash(record.Sha256) ||
				hex.EncodeToString(record.OutputId) != record.Sha256 {
				continue
			}
			var body []byte
			if remote {
				var poisoned bool
				body, ok, poisoned = cache.blob(record.Sha256, record.Size)
				if poisoned {
					return out
				}
			} else {
				body, _ = os.ReadFile(cache.blobPath(record.Sha256))
				if body != nil && !cache.check(body, record.Sha256) {
					return out
				}
				ok = body != nil
			}
			if !ok || int64(len(body)) != record.Size {
				continue
			}
			if remote {
				if atomicFileError := atomicFile(cache.blobPath(record.Sha256), body); atomicFileError != nil {
					continue
				}
				if atomicFileError := atomicFile(cache.blobPath(hash), rb); atomicFileError == nil {
					_ = atomicFile(cache.refPath(ns, action), []byte(hash))
				}
			}
			return response{
				ID:       req.ID,
				OutputID: record.OutputId,
				Size:     record.Size,
				DiskPath: cache.blobPath(record.Sha256),
			}
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
	record := record{OutputId: req.OutputID, Size: int64(len(body)), Sha256: hash}
	rb, _ := json.Marshal(record)
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
func serve(cache *cache, input io.Reader, output io.Writer) error {
	in, enc := bufio.NewReader(input), json.NewEncoder(output)
	if encodeError := enc.Encode(response{KnownCommands: []string{"get", "put", "close"}}); encodeError != nil {
		return encodeError
	}
	var waitGroup sync.WaitGroup
	var mutex sync.Mutex
	send := func(response response) {
		mutex.Lock()
		defer mutex.Unlock()
		if encodeError := enc.Encode(response); encodeError != nil {
			cache.log("protocol output: %v", encodeError)
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
				cache.log(
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
				send(cache.get(req))
			case "put":
				send(cache.put(req, body))
			default:
				send(response{ID: req.ID, Err: "unknown command"})
			}
		}(req, body)
	}
}

func main() {
	cache, err := newCache()
	if err == nil {
		err = serve(cache, os.Stdin, os.Stdout)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "adamic-gocacheprog:", err)
		os.Exit(1)
	}
}
