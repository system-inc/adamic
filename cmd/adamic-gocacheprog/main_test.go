package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

// Not parallel: this subprocess entry point owns stdin/stdout and exits the process.
func TestHelperProcess(test *testing.T) {
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

func fakeStore(test *testing.T) (*store, *httptest.Server) {
	test.Helper()
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
	test.Cleanup(server.Close)
	return s, server
}

// helperSettings are what the helper process reads, pointed at a test's own store and directory.
func helperSettings(dir, url, trust, token string) map[string]string {
	return map[string]string{
		"ADAMIC_CACHE_HELPER":  "1",
		"ADAMIC_GOCACHE_DIR":   dir,
		"ADAMIC_GOCACHE_STORE": url,
		"ADAMIC_GOCACHE_WRITE": url,
		"ADAMIC_GOCACHE_TRUST": trust,
		"ADAMIC_GOCACHE_TOKEN": token,
		"GOWORK":               "off",
		"GOCACHEPROG":          "",
	}
}

// helperEnv is this process's environment with helperSettings over it, for the helper, a copy of this test binary.
func helperEnv(dir, url, trust, token string) []string {
	overrides := helperSettings(dir, url, trust, token)
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

func tokenFile(test *testing.T) string {
	test.Helper()
	p := filepath.Join(test.TempDir(), "token")
	if err := os.WriteFile(p, []byte("test-token\n"), 0600); err != nil {
		test.Fatal(err)
	}
	return p
}

func transact(test *testing.T, dir, url, trust, token string, req request, body []byte) (response, string) {
	test.Helper()
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
		test.Fatalf("helper: %v: %s", err, stderr.String())
	}
	dec := json.NewDecoder(&stdout)
	var hello, out, close response
	if err := dec.Decode(&hello); err != nil {
		test.Fatal(err)
	}
	if hello.ID != 0 || len(hello.KnownCommands) != 3 {
		test.Fatalf("capabilities: %+v", hello)
	}
	if err := dec.Decode(&out); err != nil {
		test.Fatal(err)
	}
	if err := dec.Decode(&close); err != nil {
		test.Fatal(err)
	}
	if out.ID != req.ID || close.ID != 99 || out.Err != "" {
		test.Fatalf("responses: %+v %+v", out, close)
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

func TestSharedCache(test *testing.T) {
	test.Parallel()
	_, srv := fakeStore(test)
	body := []byte("compiled output")
	token := tokenFile(test)
	transact(test, test.TempDir(), srv.URL, "main", token, putRequest(body), body)
	out, _ := transact(test, test.TempDir(), srv.URL, "main", "", getRequest(), nil)
	if out.Miss || out.Size != int64(len(body)) {
		test.Fatalf("get: %+v", out)
	}
	got, err := os.ReadFile(out.DiskPath)
	if err != nil || !bytes.Equal(got, body) {
		test.Fatalf("body: %q, %v", got, err)
	}
}

func TestPoisoning(test *testing.T) {
	test.Parallel()
	for _, tampered := range []string{"poisoned output", "compiled output plus extra bytes"} {
		test.Run(tampered, func(test *testing.T) {
			s, srv := fakeStore(test)
			body := []byte("compiled output")
			transact(test, test.TempDir(), srv.URL, "main", tokenFile(test), putRequest(body), body)
			s.mu.Lock()
			s.data["/blobs/"+digest(body)] = []byte(tampered)
			s.mu.Unlock()
			out, log := transact(test, test.TempDir(), srv.URL, "main", "", getRequest(), nil)
			if !out.Miss || !strings.Contains(log, "cache poisoning") {
				test.Fatalf("response %+v; log %q", out, log)
			}
		})
	}
}

func TestUnavailable(test *testing.T) {
	test.Parallel()
	_, srv := fakeStore(test)
	srv.Close()
	out, _ := transact(test, test.TempDir(), srv.URL, "main", "", getRequest(), nil)
	if !out.Miss {
		test.Fatal("expected miss")
	}
}

func TestNoToken(test *testing.T) {
	test.Parallel()
	s, srv := fakeStore(test)
	body := []byte("compiled output")
	dir := test.TempDir()
	transact(test, dir, srv.URL, "main", "", putRequest(body), body)
	out, _ := transact(test, dir, srv.URL, "main", "", getRequest(), nil)
	if out.Miss {
		test.Fatal("local miss")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.requests) != 0 {
		test.Fatalf("unexpected network: %v", s.requests)
	}
}

func TestTrustSplit(test *testing.T) {
	test.Parallel()
	s, srv := fakeStore(test)
	body := []byte("compiled output")
	transact(test, test.TempDir(), srv.URL, "", tokenFile(test), putRequest(body), body)
	out, _ := transact(test, test.TempDir(), srv.URL, "main", "", getRequest(), nil)
	if !out.Miss {
		test.Fatal("trusted reader used candidate")
	}
	s.mu.Lock()
	for _, r := range s.requests {
		if strings.HasPrefix(r, "PUT /refs/gocache/") || strings.HasPrefix(r, "GET /refs/gocache-candidate/") {
			test.Fatalf("wrong namespace: %s", r)
		}
	}
	s.mu.Unlock()
	out, _ = transact(test, test.TempDir(), srv.URL, "", "", getRequest(), nil)
	if out.Miss {
		test.Fatal("candidate reader missed")
	}
}

func TestRefConflict(test *testing.T) {
	test.Parallel()
	_, srv := fakeStore(test)
	token := tokenFile(test)
	body := []byte("first")
	transact(test, test.TempDir(), srv.URL, "main", token, putRequest(body), body)
	body = []byte("second")
	dir := test.TempDir()
	_, log := transact(test, dir, srv.URL, "main", token, putRequest(body), body)
	if !strings.Contains(log, "ref conflict") {
		test.Fatal(log)
	}
	out, _ := transact(test, dir, srv.URL, "main", "", getRequest(), nil)
	got, _ := os.ReadFile(out.DiskPath)
	if !bytes.Equal(got, body) {
		test.Fatalf("local entry: %q", got)
	}
}

func TestEmptyBody(test *testing.T) {
	test.Parallel()
	_, srv := fakeStore(test)
	transact(test, test.TempDir(), srv.URL, "main", tokenFile(test), putRequest(nil), nil)
	out, _ := transact(test, test.TempDir(), srv.URL, "main", "", getRequest(), nil)
	if out.Miss || out.Size != 0 {
		test.Fatalf("get: %+v", out)
	}
}

func runGo(test *testing.T, dir, url, token string, args ...string) (string, time.Duration) {
	test.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	// runtime.GOROOT is empty in a test binary built with -trimpath, so the go command is toolchain's, and it runs with
	// toolchain's environment, the helper's settings over it: go hands them to the helper it starts as GOCACHEPROG.
	goCommand, err := toolchain.Path("go")
	if err != nil {
		test.Skipf("no go command: %v", err)
	}
	cmd := exec.CommandContext(ctx, goCommand, args...)
	cmd.Dir = dir
	var settings []string
	for key, value := range helperSettings(test.TempDir(), url, "main", token) {
		settings = append(settings, key+"="+value)
	}
	cmd.Env = toolchain.Env(append(
		settings,
		"GOCACHE="+test.TempDir(),
		"GOCACHEPROG="+fmt.Sprintf("%q -test.run=^TestHelperProcess$", os.Args[0]),
		"GOMAXPROCS=4",
	)...)
	start := time.Now()
	b, err := cmd.CombinedOutput()
	elapsed := time.Since(start)
	if err != nil {
		test.Fatalf("go %v: %v\n%s", args, err, b)
	}
	return string(b), elapsed
}

func TestGoBuildEndToEnd(test *testing.T) {
	test.Parallel()
	_, srv := fakeStore(test)
	dir := test.TempDir()
	token := tokenFile(test)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module cachetest\n\ngo 1.24\n"), 0600); err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(dir, "tiny.go"),
		[]byte("package cachetest\nfunc Add(a,b int) int { return a+b }\n"),
		0600,
	); err != nil {
		test.Fatal(err)
	}
	cold, _ := runGo(test, dir, srv.URL, token, "build", "-x", ".")
	warm, _ := runGo(test, dir, srv.URL, "", "build", "-x", ".")
	if !strings.Contains(cold, "/compile ") {
		test.Fatalf("cold build did not compile: %s", cold)
	}
	if strings.Contains(warm, "/compile ") {
		test.Fatalf("warm build compiled: %s", warm)
	}
}

func TestRecordPoisoning(test *testing.T) {
	test.Parallel()
	s, srv := fakeStore(test)
	body := []byte("compiled output")
	transact(test, test.TempDir(), srv.URL, "main", tokenFile(test), putRequest(body), body)
	s.mu.Lock()
	for p, b := range s.data {
		if strings.HasPrefix(p, "/blobs/") && !bytes.Equal(b, body) {
			s.data[p] = []byte(`{"size":0}`)
		}
	}
	s.mu.Unlock()
	out, log := transact(test, test.TempDir(), srv.URL, "main", "", getRequest(), nil)
	if !out.Miss || !strings.Contains(log, "cache poisoning") {
		test.Fatalf("response %+v; log %q", out, log)
	}
}

func TestTrustedLocalIsolation(test *testing.T) {
	test.Parallel()
	_, srv := fakeStore(test)
	body := []byte("compiled output")
	dir := test.TempDir()
	transact(test, dir, srv.URL, "", "", putRequest(body), body)
	out, _ := transact(test, dir, "off", "main", "", getRequest(), nil)
	if !out.Miss {
		test.Fatal("trusted reader used local candidate")
	}
}

func TestLocalPoisoning(test *testing.T) {
	test.Parallel()
	body := []byte("compiled output")
	dir := test.TempDir()
	out, _ := transact(test, dir, "off", "main", "", putRequest(body), body)
	if err := os.WriteFile(out.DiskPath, []byte("tampered"), 0600); err != nil {
		test.Fatal(err)
	}
	out, log := transact(test, dir, "off", "main", "", getRequest(), nil)
	if !out.Miss || !strings.Contains(log, "cache poisoning") {
		test.Fatalf("response %+v; log %q", out, log)
	}
}

func TestFailedWritesKeepLocal(test *testing.T) {
	test.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }))
	defer srv.Close()
	body := []byte("compiled output")
	dir := test.TempDir()
	out, log := transact(test, dir, srv.URL, "main", tokenFile(test), putRequest(body), body)
	if out.DiskPath == "" || !strings.Contains(log, "HTTP 500") {
		test.Fatalf("response %+v; log %q", out, log)
	}
	out, _ = transact(test, dir, "off", "main", "", getRequest(), nil)
	if out.Miss {
		test.Fatal("remote failure lost local entry")
	}
}

func TestMissingTokenWritesNothing(test *testing.T) {
	test.Parallel()
	s, srv := fakeStore(test)
	body := []byte("compiled output")
	_, log := transact(
		test,
		test.TempDir(),
		srv.URL,
		"main",
		filepath.Join(test.TempDir(), "missing-token"),
		putRequest(body),
		body,
	)
	if !strings.Contains(log, "token unavailable") {
		test.Fatal(log)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.requests) != 0 {
		test.Fatalf("unexpected network: %v", s.requests)
	}
}

func TestCandidateReadsMainFirst(test *testing.T) {
	test.Parallel()
	s, srv := fakeStore(test)
	token := tokenFile(test)
	transact(test, test.TempDir(), srv.URL, "", token, putRequest([]byte("candidate")), []byte("candidate"))
	transact(test, test.TempDir(), srv.URL, "main", token, putRequest([]byte("main")), []byte("main"))
	s.mu.Lock()
	s.requests = nil
	s.mu.Unlock()
	out, _ := transact(test, test.TempDir(), srv.URL, "", "", getRequest(), nil)
	b, _ := os.ReadFile(out.DiskPath)
	if out.Miss || string(b) != "main" {
		test.Fatalf("expected main: %+v %q", out, b)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.requests {
		if strings.HasPrefix(r, "GET /refs/gocache-candidate/") {
			test.Fatalf("unnecessary candidate read: %s", r)
		}
	}
}

func TestPoisoningStopsFallback(test *testing.T) {
	test.Parallel()
	s, srv := fakeStore(test)
	token := tokenFile(test)
	transact(test, test.TempDir(), srv.URL, "", token, putRequest([]byte("candidate")), []byte("candidate"))
	transact(test, test.TempDir(), srv.URL, "main", token, putRequest([]byte("main")), []byte("main"))
	s.mu.Lock()
	s.data["/blobs/"+digest([]byte("main"))] = []byte("evil")
	s.mu.Unlock()
	out, log := transact(test, test.TempDir(), srv.URL, "", "", getRequest(), nil)
	if !out.Miss || !strings.Contains(log, "cache poisoning") {
		test.Fatalf("response %+v; log %q", out, log)
	}
}

// writeLikeGo frames a request the way cmd/go's writeToChild does: JSON, a
// blank line, then for a put the base64 body as a quoted line.
func writeLikeGo(b *bytes.Buffer, req request, body []byte) {
	j, _ := json.Marshal(req)
	b.Write(j)
	b.WriteString("\n\n")
	if req.Command == "put" && req.BodySize > 0 {
		b.WriteString(`"` + base64.StdEncoding.EncodeToString(body) + "\"\n")
	}
}

func serveStream(test *testing.T, c *cache, input []byte) (map[int64]response, error) {
	test.Helper()
	var output bytes.Buffer
	err := serve(c, bytes.NewReader(input), &output)
	responses := map[int64]response{}
	dec := json.NewDecoder(&output)
	for {
		var r response
		if dec.Decode(&r) != nil {
			break
		}
		if _, ok := responses[r.ID]; ok {
			test.Fatalf("duplicate response for ID %d", r.ID)
		}
		responses[r.ID] = r
	}
	return responses, err
}

func localCache(test *testing.T) (*cache, *bytes.Buffer) {
	var log bytes.Buffer
	return &cache{dir: test.TempDir(), namespace: "gocache", stderr: &log, client: http.DefaultClient}, &log
}

// Go abandons a put whose body changed size mid-copy and writes the next
// request straight after the truncated base64, on the same line.
func TestAbandonedPutResyncs(test *testing.T) {
	test.Parallel()
	c, log := localCache(test)
	body := []byte("compiled output")
	var input bytes.Buffer
	writeLikeGo(&input, putRequest(body), body)
	a := sha256.Sum256([]byte("abandoned"))
	o := sha256.Sum256([]byte("changed"))
	abandoned := request{ID: 2, Command: "put", ActionID: a[:], OutputID: o[:], BodySize: 4096}
	j, _ := json.Marshal(abandoned)
	input.Write(j)
	input.WriteString("\n\n\"" + base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("x"), 300))[:100])
	writeLikeGo(&input, request{ID: 3, Command: "get", ActionID: a[:]}, nil)
	get := getRequest()
	get.ID = 4
	writeLikeGo(&input, get, nil)
	writeLikeGo(&input, request{ID: 5, Command: "close"}, nil)
	responses, err := serveStream(test, c, input.Bytes())
	if err != nil {
		test.Fatalf("serve: %v", err)
	}
	if r, ok := responses[2]; ok {
		test.Fatalf("answered the abandoned put: %+v", r)
	}
	if r, ok := responses[1]; !ok || r.Err != "" || r.DiskPath == "" {
		test.Fatalf("valid put: %+v", r)
	}
	if r, ok := responses[3]; !ok || !r.Miss {
		test.Fatalf("get after the abandoned body: %+v", r)
	}
	if _, ok := responses[4]; !ok {
		test.Fatal("following get unanswered")
	}
	if _, ok := responses[5]; !ok || len(responses) != 5 {
		test.Fatalf("responses: %+v", responses)
	}
	want := fmt.Sprintf(
		"put 2 abandoned mid-body by go: action %s, BodySize 4096, 100 base64 bytes arrived",
		hex.EncodeToString(a[:]),
	)
	if !strings.Contains(log.String(), want) {
		test.Fatalf("log %q", log.String())
	}
	input.Reset()
	writeLikeGo(&input, getRequest(), nil)
	writeLikeGo(&input, request{ID: 3, Command: "close"}, nil)
	responses, err = serveStream(test, c, input.Bytes())
	got, _ := os.ReadFile(responses[2].DiskPath)
	if err != nil || responses[2].Miss || !bytes.Equal(got, body) {
		test.Fatalf("valid put not served: %+v %q %v", responses[2], got, err)
	}
}

func TestAbandonedPutAtEOF(test *testing.T) {
	test.Parallel()
	for _, tail := range []string{"", "\"QUJD", "\"QUJD\n"} {
		test.Run(fmt.Sprintf("%q", tail), func(test *testing.T) {
			c, log := localCache(test)
			j, _ := json.Marshal(putRequest([]byte("never arrives")))
			responses, err := serveStream(test, c, append(j, "\n\n"+tail...))
			if err != nil {
				test.Fatalf("serve: %v", err)
			}
			if _, ok := responses[1]; ok || len(responses) != 1 {
				test.Fatalf("responses: %+v", responses)
			}
			if !strings.Contains(log.String(), "put 1 abandoned mid-body") {
				test.Fatalf("log %q", log.String())
			}
		})
	}
}
