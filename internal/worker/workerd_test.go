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
func requireWorkerd(test *testing.T) string {
	test.Helper()
	if os.Getenv("ADAMIC_WORKER_WORKERD") != "1" {
		test.Skip(
			"set ADAMIC_WORKER_WORKERD=1 to run the workerd corpus; Workers maintainers run this lane by hand with cloud/setup.sh --workerd",
		)
	}
	binary, lookPathError := exec.LookPath("workerd")
	if lookPathError != nil {
		test.Fatal(lookPathError)
	}
	test.Logf("workerd resolved by PATH: %s", binary)
	return binary
}

func workerdLoopback(address string) error {
	host, _, splitHostPortError := net.SplitHostPort(address)
	if splitHostPortError != nil || host != "127.0.0.1" {
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
func startWorkerd(test *testing.T, directory, address string) (*workerdServer, error) {
	test.Helper()
	return startWorkerdWithPortHook(test, directory, address, nil)
}

// released is a per-call test seam: no shared port picker or environment mutation.
func startWorkerdWithPortHook(
	test *testing.T,
	directory, address string,
	released func(string),
) (*workerdServer, error) {
	test.Helper()
	if workerdLoopbackError := workerdLoopback(address); workerdLoopbackError != nil {
		return nil, workerdLoopbackError
	}
	var previous error
	oldAddress := address
	for attempt := 0; attempt < 3; attempt++ {
		listener, err := net.Listen("tcp4", address)
		if err == nil {
			address = listener.Addr().String()
			listener.Close()
			if previous != nil {
				test.Logf("workerd bind retry: %s -> %s: %v", oldAddress, address, previous)
			}
			if released != nil {
				released(address)
			}
			server, launchError := startWorkerdAttempt(test, directory, address)
			if launchError == nil {
				return server, nil
			}
			err = launchError
		}
		if !strings.Contains(strings.ToLower(err.Error()), "address already in use") &&
			!strings.Contains(err.Error(), "EADDRINUSE") {
			return nil, err
		}
		oldAddress, previous = address, err
		address = "127.0.0.1:0"
	}
	return nil, fmt.Errorf("workerd bind failed after three ports: %w", previous)
}

func startWorkerdAttempt(test *testing.T, directory, address string) (*workerdServer, error) {
	test.Helper()
	binary, lookPathError := exec.LookPath("workerd")
	if lookPathError != nil {
		return nil, lookPathError
	}
	scratch := test.TempDir()
	entries, readDirectoryError := os.ReadDir(directory)
	if readDirectoryError != nil {
		return nil, readDirectoryError
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
		contents, readError := os.ReadFile(filepath.Join(directory, name))
		if readError != nil {
			return nil, readError
		}
		if writeError := os.WriteFile(filepath.Join(scratch, name), contents, 0600); writeError != nil {
			return nil, writeError
		}
		modules = append(
			modules,
			fmt.Sprintf("(name = %s, %s = embed %s)", strconv.Quote(name), kind, strconv.Quote(name)),
		)
	}
	configuration := `using Workerd = import "/workerd/workerd.capnp";
const config :Workerd.Config = (services = [(name = "compute", worker = (modules = [` + strings.Join(modules, ",") + `], compatibilityDate = "2026-10-07"))], sockets = [(name = "http", address = ` + strconv.Quote(address) + `, http = (), service = "compute")]);`
	configurationPath := filepath.Join(scratch, "config.capnp")
	if writeError := os.WriteFile(configurationPath, []byte(configuration), 0600); writeError != nil {
		return nil, writeError
	}
	logPath := filepath.Join(scratch, "server.log")
	log, createError := os.Create(logPath)
	if createError != nil {
		return nil, createError
	}
	controlRead, controlWrite, pipeError := os.Pipe()
	if pipeError != nil {
		log.Close()
		return nil, pipeError
	}
	command := exec.Command(binary, "serve", configurationPath, "--control-fd", "3")
	command.ExtraFiles = []*os.File{controlWrite}
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Stdout, command.Stderr = log, log
	if startError := command.Start(); startError != nil {
		controlRead.Close()
		controlWrite.Close()
		log.Close()
		return nil, startError
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
			test.Error("workerd survived Cleanup: wait timed out")
		}
	}
	test.Cleanup(stop)
	ready := make(chan error, 1)
	go func() {
		var event struct {
			Event, Socket string
			Port          int
		}
		err := json.NewDecoder(controlRead).Decode(&event)
		if err == nil && (event.Event != "listen" ||
			event.Socket != "http" ||
			net.JoinHostPort("127.0.0.1", strconv.Itoa(event.Port)) != address) {
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

func TestWorkerdRetriesAnOccupiedPort(test *testing.T) {
	test.Parallel()
	requireWorkerd(test)
	directory := test.TempDir()
	write(test, filepath.Join(directory, "worker.mjs"), `export default {fetch(){return new Response("ok")}};`)
	calls := 0
	occupiedAddress := ""
	server, startWorkerdWithPortHookError := startWorkerdWithPortHook(test, directory, "127.0.0.1:0", func(
		address string,
	) {
		calls++
		if calls != 1 {
			return
		}
		listener, listenError := net.Listen("tcp4", address)
		if listenError != nil {
			test.Fatal(listenError)
		}
		occupiedAddress = address
		test.Cleanup(func() { listener.Close() })
	})
	if startWorkerdWithPortHookError != nil {
		test.Fatal(startWorkerdWithPortHookError)
	}
	if calls != 2 || server.address == occupiedAddress {
		test.Fatalf(
			"retry did not select a fresh port: calls=%d, occupied=%s, server=%s",
			calls,
			occupiedAddress,
			server.address,
		)
	}
	command := exec.CommandContext(
		test.Context(),
		"node",
		"--input-type=module",
		"-e",
		`const response = await fetch(process.argv[1]); if (response.status !== 200 || await response.text() !== "ok") process.exit(1);`,
		"http://"+server.address,
	)
	if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
		test.Fatalf("retried workerd did not serve: %v\n%s", combinedOutputError, output)
	}
}

func TestWorkerdRefusesNonLoopbackAddresses(test *testing.T) {
	test.Parallel()
	for _, address := range []string{"0.0.0.0:0", "192.0.2.1:0", "[::1]:0", "localhost:0", ":0"} {
		if _, startWorkerdError := startWorkerd(
			test,
			"missing",
			address,
		); startWorkerdError == nil || !strings.Contains(startWorkerdError.Error(), "refuses non-loopback") {
			test.Errorf("%s: %v", address, startWorkerdError)
		}
	}
	if workerdLoopbackError := workerdLoopback("127.0.0.1:0"); workerdLoopbackError != nil {
		test.Fatal(workerdLoopbackError)
	}
}

func TestWorkerdCleanupLeavesNoServer(test *testing.T) {
	test.Parallel()
	requireWorkerd(test)
	directory := test.TempDir()
	write(test, filepath.Join(directory, "worker.mjs"), `export default {fetch(){return new Response("ok")}};`)
	var process *os.Process
	// Backstop makes the kill-removal mutant safe even when the assertion fails.
	test.Cleanup(func() {
		if process != nil {
			_ = syscall.Kill(-process.Pid, syscall.SIGKILL)
		}
	})
	test.Run("server", func(test *testing.T) {
		server, startWorkerdError := startWorkerd(test, directory, "127.0.0.1:0")
		if startWorkerdError != nil {
			test.Fatal(startWorkerdError)
		}
		process = server.process
	})
	if process == nil {
		return
	}
	if killError := syscall.Kill(process.Pid, 0); killError != syscall.ESRCH {
		test.Fatalf("workerd survives its test: pid %d, probe %v", process.Pid, killError)
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
func runWorkerdCorpusShard(test *testing.T, shard int, wasm bool) {
	requireWorkerd(test)
	family := "TestWorkerdJavaScriptCorpus"
	if wasm {
		family = "TestWorkerdWebAssemblyCorpus"
	}
	directory := workerdComputeProduct(test, wasm, false)
	server, startWorkerdError := startWorkerd(test, directory, "127.0.0.1:0")
	if startWorkerdError != nil {
		test.Fatal(startWorkerdError)
	}
	root, absolutePathError := filepath.Abs("../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	corpus := filepath.Join(root, "workers/compute/corpus")
	requests, readError := os.ReadFile(filepath.Join(corpus, "requests.jsonl"))
	if readError != nil {
		test.Fatal(readError)
	}
	responses, readError := os.ReadFile(filepath.Join(corpus, "responses.jsonl"))
	if readError != nil {
		test.Fatal(readError)
	}
	requestRows := strings.Split(strings.TrimSuffix(string(requests), "\n"), "\n")
	responseRows := strings.Split(strings.TrimSuffix(string(responses), "\n"), "\n")
	if len(requestRows) != 600 || len(responseRows) != 600 {
		test.Fatal("compute corpus must contain 600 requests and responses")
	}
	var selectedRequests, selectedResponses []string
	for i, key := range workerdCorpusKeys() {
		if familyOwns(family, shard, key) {
			if !json.Valid([]byte(requestRows[i])) || !json.Valid([]byte(responseRows[i])) {
				test.Fatal("invalid corpus JSON")
			}
			selectedRequests = append(selectedRequests, requestRows[i])
			selectedResponses = append(selectedResponses, responseRows[i])
		}
	}
	scratch := test.TempDir()
	requestPath := filepath.Join(scratch, "requests.jsonl")
	responsePath := filepath.Join(scratch, "responses.jsonl")
	write(test, requestPath, strings.Join(selectedRequests, "\n")+"\n")
	write(test, responsePath, strings.Join(selectedResponses, "\n")+"\n")
	replay := filepath.Join(root, "workers/replay.mjs")
	for _, args := range [][]string{
		{filepath.Join(root, "workers/compute/twin/worker.ts"), requestPath, "--compare", responsePath},
		{"--url", "http://" + server.address, requestPath, "--compare", responsePath},
	} {
		command := exec.CommandContext(
			test.Context(),
			"node",
			append([]string{"--disable-warning=ExperimentalWarning", replay}, args...)...,
		)
		if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
			test.Fatalf("corpus: %v\n%s", combinedOutputError, output)
		} else {
			test.Logf("%s", output)
		}
	}
	test.Logf("%d requests matched twin and workerd", len(selectedRequests))
}
func runWorkerdJavaScriptCorpusShard(test *testing.T, shard int) {
	runWorkerdCorpusShard(test, shard, false)
}
func runWorkerdWebAssemblyCorpusShard(test *testing.T, shard int) {
	runWorkerdCorpusShard(test, shard, true)
}
