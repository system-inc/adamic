package worker

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Like oracle's requireWasmtime lane: a recipe opts in; a claimed tool must exist.
func requireWorkerd(t *testing.T) string {
	t.Helper()
	if os.Getenv("ADAMIC_WORKER_WORKERD") != "1" {
		t.Skip("set ADAMIC_WORKER_WORKERD=1 to run the workerd corpus; Workers maintainers run this lane by hand with cloud/setup.sh --workerd")
	}
	binary, err := exec.LookPath("workerd")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("workerd resolved by PATH: %s", binary)
	return binary
}

func workerdLoopback(address string) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil || host != "127.0.0.1" {
		return fmt.Errorf("workerd refuses non-loopback address %q; use 127.0.0.1", address)
	}
	return nil
}

type workerdServer struct {
	address string
	process *os.Process
}

// startWorkerd owns the server's whole process group until Cleanup has waited.
// Tests run against whatever cloud/setup.sh pins, found through the unit's PATH.
func startWorkerd(t *testing.T, directory, address string) (*workerdServer, error) {
	t.Helper()
	return startWorkerdWithPortHook(t, directory, address, nil)
}

// released is a per-call test seam: no shared port picker or environment mutation.
func startWorkerdWithPortHook(t *testing.T, directory, address string, released func(string)) (*workerdServer, error) {
	t.Helper()
	if err := workerdLoopback(address); err != nil {
		return nil, err
	}
	var previous error
	oldAddress := address
	for attempt := 0; attempt < 3; attempt++ {
		listener, err := net.Listen("tcp4", address)
		if err == nil {
			address = listener.Addr().String()
			listener.Close()
			if previous != nil {
				t.Logf("workerd bind retry: %s -> %s: %v", oldAddress, address, previous)
			}
			if released != nil {
				released(address)
			}
			server, launchError := startWorkerdAttempt(t, directory, address)
			if launchError == nil {
				return server, nil
			}
			err = launchError
		}
		if !strings.Contains(strings.ToLower(err.Error()), "address already in use") && !strings.Contains(err.Error(), "EADDRINUSE") {
			return nil, err
		}
		oldAddress, previous = address, err
		address = "127.0.0.1:0"
	}
	return nil, fmt.Errorf("workerd bind failed after three ports: %w", previous)
}

func startWorkerdAttempt(t *testing.T, directory, address string) (*workerdServer, error) {
	t.Helper()
	binary, err := exec.LookPath("workerd")
	if err != nil {
		return nil, err
	}
	scratch := t.TempDir()
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Name() == "worker.mjs" {
			return entries[j].Name() != "worker.mjs"
		}
		if entries[j].Name() == "worker.mjs" {
			return false
		}
		return entries[i].Name() < entries[j].Name()
	})
	var modules []string
	for _, entry := range entries {
		name := entry.Name()
		kind := "esModule"
		if strings.HasSuffix(name, ".wasm") {
			kind = "wasm"
		} else if !strings.HasSuffix(name, ".mjs") {
			continue
		}
		contents, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(scratch, name), contents, 0600); err != nil {
			return nil, err
		}
		modules = append(modules, fmt.Sprintf("(name = %s, %s = embed %s)", strconv.Quote(name), kind, strconv.Quote(name)))
	}
	config := `using Workerd = import "/workerd/workerd.capnp";
const config :Workerd.Config = (services = [(name = "compute", worker = (modules = [` + strings.Join(modules, ",") + `], compatibilityDate = "2026-10-07"))], sockets = [(name = "http", address = ` + strconv.Quote(address) + `, http = (), service = "compute")]);`
	configPath := filepath.Join(scratch, "config.capnp")
	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		return nil, err
	}
	logPath := filepath.Join(scratch, "server.log")
	log, err := os.Create(logPath)
	if err != nil {
		return nil, err
	}
	controlRead, controlWrite, err := os.Pipe()
	if err != nil {
		log.Close()
		return nil, err
	}
	command := exec.Command(binary, "serve", configPath, "--control-fd", "3")
	command.ExtraFiles = []*os.File{controlWrite}
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Stdout, command.Stderr = log, log
	if err := command.Start(); err != nil {
		controlRead.Close()
		controlWrite.Close()
		log.Close()
		return nil, err
	}
	controlWrite.Close()
	done := make(chan struct{})
	var waitError error
	go func() { waitError = command.Wait(); log.Close(); close(done) }()
	stopped := false
	stop := func() {
		if stopped {
			return
		}
		stopped = true
		controlRead.Close()
		// Kill the group even if the direct child has already exited.
		_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("workerd survived Cleanup: wait timed out")
		}
	}
	t.Cleanup(stop)
	ready := make(chan error, 1)
	go func() {
		var event struct {
			Event, Socket string
			Port          int
		}
		err := json.NewDecoder(controlRead).Decode(&event)
		if err == nil && (event.Event != "listen" || event.Socket != "http" || net.JoinHostPort("127.0.0.1", strconv.Itoa(event.Port)) != address) {
			err = fmt.Errorf("unexpected workerd readiness event: %+v", event)
		}
		ready <- err
	}()
	// Only workerd's own listen event proves readiness. A TCP probe can connect
	// to the process that stole our port and falsely declare this server ready.
	select {
	case err := <-ready:
		if err == nil {
			return &workerdServer{address, command.Process}, nil
		}
		stop()
		report, _ := os.ReadFile(logPath)
		return nil, fmt.Errorf("workerd readiness: %v\n%s", err, report)
	case <-done:
		stop()
		report, _ := os.ReadFile(logPath)
		return nil, fmt.Errorf("workerd exited: %v\n%s", waitError, report)
	case <-time.After(10 * time.Second):
		stop()
		report, _ := os.ReadFile(logPath)
		return nil, fmt.Errorf("workerd readiness timed out\n%s", report)
	}
}

func TestWorkerdRetriesAnOccupiedPort(t *testing.T) {
	t.Parallel()
	requireWorkerd(t)
	directory := t.TempDir()
	write(t, filepath.Join(directory, "worker.mjs"), `export default {fetch(){return new Response("ok")}};`)
	calls := 0
	occupiedAddress := ""
	server, err := startWorkerdWithPortHook(t, directory, "127.0.0.1:0", func(address string) {
		calls++
		if calls != 1 {
			return
		}
		listener, err := net.Listen("tcp4", address)
		if err != nil {
			t.Fatal(err)
		}
		occupiedAddress = address
		t.Cleanup(func() { listener.Close() })
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || server.address == occupiedAddress {
		t.Fatalf("retry did not select a fresh port: calls=%d, occupied=%s, server=%s", calls, occupiedAddress, server.address)
	}
	command := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", `const response = await fetch(process.argv[1]); if (response.status !== 200 || await response.text() !== "ok") process.exit(1);`, "http://"+server.address)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("retried workerd did not serve: %v\n%s", err, output)
	}
}

func TestWorkerdRefusesNonLoopbackAddresses(t *testing.T) {
	t.Parallel()
	for _, address := range []string{"0.0.0.0:0", "192.0.2.1:0", "[::1]:0", "localhost:0", ":0"} {
		if _, err := startWorkerd(t, "missing", address); err == nil || !strings.Contains(err.Error(), "refuses non-loopback") {
			t.Errorf("%s: %v", address, err)
		}
	}
	if err := workerdLoopback("127.0.0.1:0"); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerdCleanupLeavesNoServer(t *testing.T) {
	t.Parallel()
	requireWorkerd(t)
	directory := t.TempDir()
	write(t, filepath.Join(directory, "worker.mjs"), `export default {fetch(){return new Response("ok")}};`)
	var process *os.Process
	// Backstop makes the kill-removal mutant safe even when the assertion fails.
	t.Cleanup(func() {
		if process != nil {
			_ = syscall.Kill(-process.Pid, syscall.SIGKILL)
		}
	})
	t.Run("server", func(t *testing.T) {
		server, err := startWorkerd(t, directory, "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		process = server.process
	})
	if process == nil {
		return
	}
	if err := syscall.Kill(process.Pid, 0); err != syscall.ESRCH {
		t.Fatalf("workerd survives its test: pid %d, probe %v", process.Pid, err)
	}
}

func workerdCorpusKeys() []string {
	keys := make([]string, 600)
	for i := range keys {
		keys[i] = strconv.Itoa(i)
	}
	return keys
}

// Each corpus unit runs against whatever cloud/setup.sh pins, resolved on PATH;
// it never uses a private npm copy or caches a server or replay verdict.
func runWorkerdCorpusShard(t *testing.T, shard int, wasm bool) {
	requireWorkerd(t)
	family := "TestWorkerdJavaScriptCorpus"
	if wasm {
		family = "TestWorkerdWebAssemblyCorpus"
	}
	directory := workerdComputeProduct(t, wasm, false)
	server, err := startWorkerd(t, directory, "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	corpus := filepath.Join(root, "workers/compute/corpus")
	requests, err := os.ReadFile(filepath.Join(corpus, "requests.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	responses, err := os.ReadFile(filepath.Join(corpus, "responses.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	requestRows := strings.Split(strings.TrimSuffix(string(requests), "\n"), "\n")
	responseRows := strings.Split(strings.TrimSuffix(string(responses), "\n"), "\n")
	if len(requestRows) != 600 || len(responseRows) != 600 {
		t.Fatal("compute corpus must contain 600 requests and responses")
	}
	var selectedRequests, selectedResponses []string
	for i, key := range workerdCorpusKeys() {
		if familyOwns(family, shard, key) {
			if !json.Valid([]byte(requestRows[i])) || !json.Valid([]byte(responseRows[i])) {
				t.Fatal("invalid corpus JSON")
			}
			selectedRequests = append(selectedRequests, requestRows[i])
			selectedResponses = append(selectedResponses, responseRows[i])
		}
	}
	scratch := t.TempDir()
	requestPath := filepath.Join(scratch, "requests.jsonl")
	responsePath := filepath.Join(scratch, "responses.jsonl")
	write(t, requestPath, strings.Join(selectedRequests, "\n")+"\n")
	write(t, responsePath, strings.Join(selectedResponses, "\n")+"\n")
	replay := filepath.Join(root, "workers/replay.mjs")
	for _, args := range [][]string{
		{filepath.Join(root, "workers/compute/twin/worker.ts"), requestPath, "--compare", responsePath},
		{"--url", "http://" + server.address, requestPath, "--compare", responsePath},
	} {
		command := exec.CommandContext(t.Context(), "node", append([]string{"--disable-warning=ExperimentalWarning", replay}, args...)...)
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("corpus: %v\n%s", err, output)
		} else {
			t.Logf("%s", output)
		}
	}
	t.Logf("%d requests matched twin and workerd", len(selectedRequests))
}
func runWorkerdJavaScriptCorpusShard(t *testing.T, shard int)  { runWorkerdCorpusShard(t, shard, false) }
func runWorkerdWebAssemblyCorpusShard(t *testing.T, shard int) { runWorkerdCorpusShard(t, shard, true) }
