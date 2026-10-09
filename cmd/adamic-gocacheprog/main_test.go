package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// Not parallel: this subprocess entry point owns stdin/stdout and exits the process.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("ADAMIC_CACHE_HELPER") != "1" {
		return
	}
	main()
	os.Exit(0)
}

type store struct {
	mu       sync.Mutex
	data     map[string][]byte
	requests []string
}

func fakeStore(t *testing.T) (*store, *httptest.Server) {
	t.Helper()
	s := &store{data: map[string][]byte{}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.requests = append(s.requests, r.Method+" "+r.URL.Path)
		switch r.Method {
		case "PUT":
			if r.Header.Get("Authorization") != "Bearer test-token" {
				w.WriteHeader(401)
				return
			}
			b, _ := io.ReadAll(r.Body)
			if old, ok := s.data[r.URL.Path]; ok && strings.HasPrefix(r.URL.Path, "/refs/") && !bytes.Equal(old, b) {
				w.WriteHeader(409)
				return
			}
			s.data[r.URL.Path] = b
			w.WriteHeader(201)
		case "GET":
			b, ok := s.data[r.URL.Path]
			if !ok {
				w.WriteHeader(404)
				return
			}
			_, _ = w.Write(b)
		}
	}))
	t.Cleanup(server.Close)
	return s, server
}

func helperEnv(dir, url, trust, token string) []string {
	overrides := map[string]string{"ADAMIC_CACHE_HELPER": "1", "ADAMIC_GOCACHE_DIR": dir, "ADAMIC_GOCACHE_STORE": url, "ADAMIC_GOCACHE_WRITE": url, "ADAMIC_GOCACHE_TRUST": trust, "ADAMIC_GOCACHE_TOKEN": token, "GOWORK": "off", "GOCACHEPROG": ""}
	result := []string{}
	for _, v := range os.Environ() {
		key, _, _ := strings.Cut(v, "=")
		if _, ok := overrides[key]; !ok {
			result = append(result, v)
		}
	}
	for k, v := range overrides {
		result = append(result, k+"="+v)
	}
	return result
}

func tokenFile(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "token")
	if err := os.WriteFile(p, []byte("test-token\n"), 0600); err != nil {
		t.Fatal(err)
	}
	return p
}

func transact(t *testing.T, dir, url, trust, token string, req request, body []byte) (response, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestHelperProcess$")
	cmd.Env = helperEnv(dir, url, trust, token)
	var input, stderr, stdout bytes.Buffer
	enc := json.NewEncoder(&input)
	_ = enc.Encode(req)
	if req.Command == "put" && req.BodySize > 0 {
		_ = enc.Encode(body)
	}
	_ = enc.Encode(request{ID: 99, Command: "close"})
	cmd.Stdin = &input
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("helper: %v: %s", err, stderr.String())
	}
	dec := json.NewDecoder(&stdout)
	var hello, out, close response
	if err := dec.Decode(&hello); err != nil {
		t.Fatal(err)
	}
	if hello.ID != 0 || len(hello.KnownCommands) != 3 {
		t.Fatalf("capabilities: %+v", hello)
	}
	if err := dec.Decode(&out); err != nil {
		t.Fatal(err)
	}
	if err := dec.Decode(&close); err != nil {
		t.Fatal(err)
	}
	if out.ID != req.ID || close.ID != 99 || out.Err != "" {
		t.Fatalf("responses: %+v %+v", out, close)
	}
	return out, stderr.String()
}

func putRequest(body []byte) request {
	a := sha256.Sum256([]byte("action"))
	o := sha256.Sum256(body)
	return request{ID: 1, Command: "put", ActionID: a[:], OutputID: o[:], BodySize: int64(len(body))}
}

func getRequest() request {
	a := sha256.Sum256([]byte("action"))
	return request{ID: 2, Command: "get", ActionID: a[:]}
}

func TestSharedCache(t *testing.T) {
	t.Parallel()
	_, srv := fakeStore(t)
	body := []byte("compiled output")
	token := tokenFile(t)
	transact(t, t.TempDir(), srv.URL, "main", token, putRequest(body), body)
	out, _ := transact(t, t.TempDir(), srv.URL, "main", "", getRequest(), nil)
	if out.Miss || out.Size != int64(len(body)) {
		t.Fatalf("get: %+v", out)
	}
	got, err := os.ReadFile(out.DiskPath)
	if err != nil || !bytes.Equal(got, body) {
		t.Fatalf("body: %q, %v", got, err)
	}
}

func TestPoisoning(t *testing.T) {
	t.Parallel()
	for _, tampered := range []string{"poisoned output", "compiled output plus extra bytes"} {
		t.Run(tampered, func(t *testing.T) {
			s, srv := fakeStore(t)
			body := []byte("compiled output")
			transact(t, t.TempDir(), srv.URL, "main", tokenFile(t), putRequest(body), body)
			s.mu.Lock()
			s.data["/blobs/"+digest(body)] = []byte(tampered)
			s.mu.Unlock()
			out, log := transact(t, t.TempDir(), srv.URL, "main", "", getRequest(), nil)
			if !out.Miss || !strings.Contains(log, "cache poisoning") {
				t.Fatalf("response %+v; log %q", out, log)
			}
		})
	}
}

func TestUnavailable(t *testing.T) {
	t.Parallel()
	_, srv := fakeStore(t)
	srv.Close()
	out, _ := transact(t, t.TempDir(), srv.URL, "main", "", getRequest(), nil)
	if !out.Miss {
		t.Fatal("expected miss")
	}
}

func TestNoToken(t *testing.T) {
	t.Parallel()
	s, srv := fakeStore(t)
	body := []byte("compiled output")
	dir := t.TempDir()
	transact(t, dir, srv.URL, "main", "", putRequest(body), body)
	out, _ := transact(t, dir, srv.URL, "main", "", getRequest(), nil)
	if out.Miss {
		t.Fatal("local miss")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.requests) != 0 {
		t.Fatalf("unexpected network: %v", s.requests)
	}
}

func TestTrustSplit(t *testing.T) {
	t.Parallel()
	s, srv := fakeStore(t)
	body := []byte("compiled output")
	transact(t, t.TempDir(), srv.URL, "", tokenFile(t), putRequest(body), body)
	out, _ := transact(t, t.TempDir(), srv.URL, "main", "", getRequest(), nil)
	if !out.Miss {
		t.Fatal("trusted reader used candidate")
	}
	s.mu.Lock()
	for _, r := range s.requests {
		if strings.HasPrefix(r, "PUT /refs/gocache/") || strings.HasPrefix(r, "GET /refs/gocache-candidate/") {
			t.Fatalf("wrong namespace: %s", r)
		}
	}
	s.mu.Unlock()
	out, _ = transact(t, t.TempDir(), srv.URL, "", "", getRequest(), nil)
	if out.Miss {
		t.Fatal("candidate reader missed")
	}
}

func TestRefConflict(t *testing.T) {
	t.Parallel()
	_, srv := fakeStore(t)
	token := tokenFile(t)
	body := []byte("first")
	transact(t, t.TempDir(), srv.URL, "main", token, putRequest(body), body)
	body = []byte("second")
	dir := t.TempDir()
	_, log := transact(t, dir, srv.URL, "main", token, putRequest(body), body)
	if !strings.Contains(log, "ref conflict") {
		t.Fatal(log)
	}
	out, _ := transact(t, dir, srv.URL, "main", "", getRequest(), nil)
	got, _ := os.ReadFile(out.DiskPath)
	if !bytes.Equal(got, body) {
		t.Fatalf("local entry: %q", got)
	}
}

func TestEmptyBody(t *testing.T) {
	t.Parallel()
	_, srv := fakeStore(t)
	transact(t, t.TempDir(), srv.URL, "main", tokenFile(t), putRequest(nil), nil)
	out, _ := transact(t, t.TempDir(), srv.URL, "main", "", getRequest(), nil)
	if out.Miss || out.Size != 0 {
		t.Fatalf("get: %+v", out)
	}
}

func runGo(t *testing.T, dir, url, token string, args ...string) (string, time.Duration) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), args...)
	cmd.Dir = dir
	cmd.Env = append(helperEnv(t.TempDir(), url, "main", token), "GOCACHE="+t.TempDir(), "GOCACHEPROG="+fmt.Sprintf("%q -test.run=^TestHelperProcess$", os.Args[0]), "GOMAXPROCS=4")
	start := time.Now()
	b, err := cmd.CombinedOutput()
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("go %v: %v\n%s", args, err, b)
	}
	return string(b), elapsed
}

func TestGoBuildEndToEnd(t *testing.T) {
	t.Parallel()
	_, srv := fakeStore(t)
	dir := t.TempDir()
	token := tokenFile(t)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module cachetest\n\ngo 1.24\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tiny.go"), []byte("package cachetest\nfunc Add(a,b int) int { return a+b }\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cold, _ := runGo(t, dir, srv.URL, token, "build", "-x", ".")
	warm, _ := runGo(t, dir, srv.URL, "", "build", "-x", ".")
	if !strings.Contains(cold, "/compile ") {
		t.Fatalf("cold build did not compile: %s", cold)
	}
	if strings.Contains(warm, "/compile ") {
		t.Fatalf("warm build compiled: %s", warm)
	}
}

func TestRecordPoisoning(t *testing.T) {
	t.Parallel()
	s, srv := fakeStore(t)
	body := []byte("compiled output")
	transact(t, t.TempDir(), srv.URL, "main", tokenFile(t), putRequest(body), body)
	s.mu.Lock()
	for p, b := range s.data {
		if strings.HasPrefix(p, "/blobs/") && !bytes.Equal(b, body) {
			s.data[p] = []byte(`{"size":0}`)
		}
	}
	s.mu.Unlock()
	out, log := transact(t, t.TempDir(), srv.URL, "main", "", getRequest(), nil)
	if !out.Miss || !strings.Contains(log, "cache poisoning") {
		t.Fatalf("response %+v; log %q", out, log)
	}
}

func TestTrustedLocalIsolation(t *testing.T) {
	t.Parallel()
	_, srv := fakeStore(t)
	body := []byte("compiled output")
	dir := t.TempDir()
	transact(t, dir, srv.URL, "", "", putRequest(body), body)
	out, _ := transact(t, dir, "off", "main", "", getRequest(), nil)
	if !out.Miss {
		t.Fatal("trusted reader used local candidate")
	}
}

func TestLocalPoisoning(t *testing.T) {
	t.Parallel()
	body := []byte("compiled output")
	dir := t.TempDir()
	out, _ := transact(t, dir, "off", "main", "", putRequest(body), body)
	if err := os.WriteFile(out.DiskPath, []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	out, log := transact(t, dir, "off", "main", "", getRequest(), nil)
	if !out.Miss || !strings.Contains(log, "cache poisoning") {
		t.Fatalf("response %+v; log %q", out, log)
	}
}

func TestFailedWritesKeepLocal(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }))
	defer srv.Close()
	body := []byte("compiled output")
	dir := t.TempDir()
	out, log := transact(t, dir, srv.URL, "main", tokenFile(t), putRequest(body), body)
	if out.DiskPath == "" || !strings.Contains(log, "HTTP 500") {
		t.Fatalf("response %+v; log %q", out, log)
	}
	out, _ = transact(t, dir, "off", "main", "", getRequest(), nil)
	if out.Miss {
		t.Fatal("remote failure lost local entry")
	}
}

func TestMissingTokenWritesNothing(t *testing.T) {
	t.Parallel()
	s, srv := fakeStore(t)
	body := []byte("compiled output")
	_, log := transact(t, t.TempDir(), srv.URL, "main", filepath.Join(t.TempDir(), "missing-token"), putRequest(body), body)
	if !strings.Contains(log, "token unavailable") {
		t.Fatal(log)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.requests) != 0 {
		t.Fatalf("unexpected network: %v", s.requests)
	}
}

func TestCandidateReadsMainFirst(t *testing.T) {
	t.Parallel()
	s, srv := fakeStore(t)
	token := tokenFile(t)
	transact(t, t.TempDir(), srv.URL, "", token, putRequest([]byte("candidate")), []byte("candidate"))
	transact(t, t.TempDir(), srv.URL, "main", token, putRequest([]byte("main")), []byte("main"))
	s.mu.Lock()
	s.requests = nil
	s.mu.Unlock()
	out, _ := transact(t, t.TempDir(), srv.URL, "", "", getRequest(), nil)
	b, _ := os.ReadFile(out.DiskPath)
	if out.Miss || string(b) != "main" {
		t.Fatalf("expected main: %+v %q", out, b)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.requests {
		if strings.HasPrefix(r, "GET /refs/gocache-candidate/") {
			t.Fatalf("unnecessary candidate read: %s", r)
		}
	}
}

func TestPoisoningStopsFallback(t *testing.T) {
	t.Parallel()
	s, srv := fakeStore(t)
	token := tokenFile(t)
	transact(t, t.TempDir(), srv.URL, "", token, putRequest([]byte("candidate")), []byte("candidate"))
	transact(t, t.TempDir(), srv.URL, "main", token, putRequest([]byte("main")), []byte("main"))
	s.mu.Lock()
	s.data["/blobs/"+digest([]byte("main"))] = []byte("evil")
	s.mu.Unlock()
	out, log := transact(t, t.TempDir(), srv.URL, "", "", getRequest(), nil)
	if !out.Miss || !strings.Contains(log, "cache poisoning") {
		t.Fatalf("response %+v; log %q", out, log)
	}
}
