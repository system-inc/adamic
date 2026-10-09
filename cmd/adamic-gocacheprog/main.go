// adamic-gocacheprog implements Go's GOCACHEPROG protocol (Go 1.24+).
package main

import (
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
	root, err := os.UserCacheDir()
	if err != nil {
		root = os.TempDir()
	}
	dir, err := filepath.Abs(env("ADAMIC_GOCACHE_DIR", filepath.Join(root, "adamic-gocache")))
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	c := &cache{
		dir:       dir,
		store:     strings.TrimRight(env("ADAMIC_GOCACHE_STORE", "https://adamic-store.kirkouimet.com"), "/"),
		writer:    strings.TrimRight(env("ADAMIC_GOCACHE_WRITE", "https://loom.kirkouimet.com/public"), "/"),
		namespace: "gocache-candidate",
		stderr:    os.Stderr,
		client: &http.Client{
			Timeout:       3 * time.Second,
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
		b, err := os.ReadFile(path)
		if err == nil {
			c.token = strings.TrimSpace(string(b))
		} else {
			c.log("token unavailable: %v", err)
		}
	}
	return c, nil
}

func (c *cache) log(format string, args ...any) {
	c.logMu.Lock()
	defer c.logMu.Unlock()
	fmt.Fprintf(c.stderr, "adamic-gocacheprog: "+format+"\n", args...)
}

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func validHash(s string) bool {
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == sha256.Size && s == strings.ToLower(s)
}

func atomicFile(path string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".tmp-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

func (c *cache) blobPath(hash string) string      { return filepath.Join(c.dir, "blobs", hash) }
func (c *cache) refPath(ns, action string) string { return filepath.Join(c.dir, "refs", ns, action) }
func (c *cache) check(b []byte, hash string) bool {
	if actual := digest(b); actual != hash {
		c.log("cache poisoning: blob %s has sha256 %s", hash, actual)
		return false
	}
	return true
}

func (c *cache) readRemote(path string, limit int64) ([]byte, bool) {
	res, err := c.client.Get(c.store + path)
	if err != nil {
		return nil, false
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, false
	}
	b, err := io.ReadAll(io.LimitReader(res.Body, limit+1))
	return b, err == nil && int64(len(b)) <= limit
}

// poisoned distinguishes corruption from an ordinary miss: once corruption is
// detected, the entire get misses instead of falling through to another ref.
func (c *cache) blob(hash string, limit int64) (body []byte, ok, poisoned bool) {
	if !validHash(hash) {
		return nil, false, false
	}
	if b, err := os.ReadFile(c.blobPath(hash)); err == nil {
		if !c.check(b, hash) {
			return nil, false, true
		}
		if int64(len(b)) <= limit {
			return b, true, false
		}
		return nil, false, false
	}
	if c.store == "" {
		return nil, false, false
	}
	b, ok := c.readRemote("/blobs/"+hash, limit)
	if int64(len(b)) > limit {
		c.log("cache poisoning: blob %s exceeds size limit %d", hash, limit)
		return nil, false, true
	}
	if !ok {
		return nil, false, false
	}
	if !c.check(b, hash) {
		return nil, false, true
	}
	return b, true, false
}

func (c *cache) get(req request) response {
	out := response{ID: req.ID, Miss: true}
	if len(req.ActionID) != sha256.Size {
		return out
	}
	action := hex.EncodeToString(req.ActionID)
	namespaces := []string{"gocache"}
	if c.namespace != "gocache" {
		namespaces = append(namespaces, "gocache-candidate")
	}
	// Check all permitted local entries before making any network request.
	for _, remote := range []bool{false, true} {
		if remote && c.store == "" {
			break
		}
		for _, ns := range namespaces {
			var ref []byte
			if remote {
				var ok bool
				ref, ok = c.readRemote("/refs/"+ns+"/"+action, 256)
				if !ok {
					continue
				}
			} else {
				var err error
				ref, err = os.ReadFile(c.refPath(ns, action))
				if err != nil {
					continue
				}
			}
			hash := strings.TrimSpace(string(ref))
			if !validHash(hash) {
				continue
			}
			var rb []byte
			var ok bool
			if remote {
				var poisoned bool
				rb, ok, poisoned = c.blob(hash, 4096)
				if poisoned {
					return out
				}
			} else {
				rb, _ = os.ReadFile(c.blobPath(hash))
				if rb != nil && !c.check(rb, hash) {
					return out
				}
				ok = rb != nil
			}
			if !ok {
				continue
			}
			var r record
			if json.Unmarshal(rb, &r) != nil || len(r.OutputID) != sha256.Size || r.Size < 0 || r.Size == int64(^uint64(0)>>1) || !validHash(r.SHA256) || hex.EncodeToString(r.OutputID) != r.SHA256 {
				continue
			}
			var body []byte
			if remote {
				var poisoned bool
				body, ok, poisoned = c.blob(r.SHA256, r.Size)
				if poisoned {
					return out
				}
			} else {
				body, _ = os.ReadFile(c.blobPath(r.SHA256))
				if body != nil && !c.check(body, r.SHA256) {
					return out
				}
				ok = body != nil
			}
			if !ok || int64(len(body)) != r.Size {
				continue
			}
			if remote {
				if err := atomicFile(c.blobPath(r.SHA256), body); err != nil {
					continue
				}
				if err := atomicFile(c.blobPath(hash), rb); err == nil {
					_ = atomicFile(c.refPath(ns, action), []byte(hash))
				}
			}
			return response{ID: req.ID, OutputID: r.OutputID, Size: r.Size, DiskPath: c.blobPath(r.SHA256)}
		}
	}
	return out
}

func (c *cache) upload(path string, b []byte) bool {
	req, err := http.NewRequest(http.MethodPut, c.writer+path, bytes.NewReader(b))
	if err != nil {
		c.log("write: %v", err)
		return false
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	res, err := c.client.Do(req)
	if err != nil {
		c.log("write: %v", err)
		return false
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 4096))
	if res.StatusCode == 200 || res.StatusCode == 201 {
		return true
	}
	if res.StatusCode == 409 && strings.HasPrefix(path, "/refs/") {
		c.log("ref conflict: %s already holds a different record; keeping local entry", path)
	} else {
		c.log("write %s: HTTP %d", path, res.StatusCode)
	}
	return false
}

func (c *cache) put(req request, body []byte) response {
	out := response{ID: req.ID}
	hash := digest(body)
	if len(req.ActionID) != sha256.Size || hex.EncodeToString(req.OutputID) != hash || req.BodySize != int64(len(body)) {
		out.Err = "invalid put IDs or body size"
		return out
	}
	r := record{OutputID: req.OutputID, Size: int64(len(body)), SHA256: hash}
	rb, _ := json.Marshal(r)
	rh := digest(rb)
	action := hex.EncodeToString(req.ActionID)
	if err := atomicFile(c.blobPath(hash), body); err != nil {
		out.Err = err.Error()
		return out
	}
	out.DiskPath = c.blobPath(hash)
	if err := atomicFile(c.blobPath(rh), rb); err != nil {
		c.log("local record: %v", err)
	} else if err = atomicFile(c.refPath(c.namespace, action), []byte(rh)); err != nil {
		c.log("local ref: %v", err)
	}
	if c.store != "" && c.token != "" && c.upload("/blobs/"+hash, body) && c.upload("/blobs/"+rh, rb) {
		c.upload("/refs/"+c.namespace+"/"+action, []byte(rh))
	}
	return out
}

func serve(c *cache, input io.Reader, output io.Writer) error {
	dec, enc := json.NewDecoder(input), json.NewEncoder(output)
	if err := enc.Encode(response{KnownCommands: []string{"get", "put", "close"}}); err != nil {
		return err
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	send := func(r response) {
		mu.Lock()
		defer mu.Unlock()
		if err := enc.Encode(r); err != nil {
			c.log("protocol output: %v", err)
		}
	}
	slots := make(chan struct{}, 16)
	for {
		var req request
		if err := dec.Decode(&req); err != nil {
			wg.Wait()
			if err == io.EOF {
				return nil
			}
			return err
		}
		var body []byte
		if req.Command == "put" && req.BodySize > 0 {
			if err := dec.Decode(&body); err != nil {
				wg.Wait()
				return err
			}
		}
		if req.Command == "close" {
			wg.Wait()
			send(response{ID: req.ID})
			return nil
		}
		slots <- struct{}{}
		wg.Add(1)
		go func(req request, body []byte) {
			defer wg.Done()
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
