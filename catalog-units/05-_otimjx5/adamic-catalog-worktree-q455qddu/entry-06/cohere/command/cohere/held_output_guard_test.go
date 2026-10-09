package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// A command a test ran has exited, but its captured output never closes: exec's Wait sits in
// awaitGoroutines until the test binary's -timeout, which by default is ten minutes of silence before a
// dump nobody reads as the answer. Two things cause it. Something the command started outlives it and
// still holds the pipe's write end, or the pipe's read end was closed out from under the poller by a
// stray Close, a second *os.File on one descriptor finalized after the first was closed, so the reader
// waits for a readiness that never comes. The second hung command/cohere for 20 minutes on 2026-10-03
// (#zqsdzbq): verdict_test.go wrapped another file's descriptor and dropped the wrapper.
//
// So the package's tests run under a watch. A goroutine still waiting on a command's output past
// heldOutputDeadline after the command exited ends the test binary with a report: the stuck goroutine,
// which names its test, and every process holding the other end of a pipe this process reads. When
// none does, the report says the read end was most likely closed behind the poller.

// heldOutputDeadline is how long a command's output may stay open after the command has exited. Copying
// what is left in a pipe takes microseconds, so anything near this is a defect, never a slow machine.
const heldOutputDeadline = 2 * time.Minute

// watchForHeldOutput checks every interval for a goroutine waiting on an exited command's output, and
// calls fail with a report once one has waited past deadline. It returns a function that stops the watch.
func watchForHeldOutput(deadline time.Duration, interval time.Duration, fail func(report string)) func() {
	stop := make(chan struct{})
	go func() {
		firstSeen := map[string]time.Time{}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
			}
			now := time.Now()
			waiting := goroutinesWaitingOnOutput()
			for identifier := range firstSeen {
				if _, still := waiting[identifier]; !still {
					delete(firstSeen, identifier)
				}
			}
			for identifier, stack := range waiting {
				if _, seen := firstSeen[identifier]; !seen {
					firstSeen[identifier] = now
					continue
				}
				if now.Sub(firstSeen[identifier]) >= deadline {
					fail(heldOutputReport(stack, now.Sub(firstSeen[identifier])))
					return
				}
			}
		}
	}()
	return func() { close(stop) }
}

// goroutinesWaitingOnOutput is every goroutine in exec's wait for a command's output after the command
// exited, by goroutine id, with its stack.
func goroutinesWaitingOnOutput() map[string]string {
	buffer := make([]byte, 1<<20)
	for {
		written := runtime.Stack(buffer, true)
		if written < len(buffer) {
			buffer = buffer[:written]
			break
		}
		buffer = make([]byte, 2*len(buffer))
	}
	waiting := map[string]string{}
	for _, stack := range strings.Split(string(buffer), "\n\n") {
		if !strings.Contains(stack, "os/exec.(*Cmd).awaitGoroutines") {
			continue
		}
		if fields := strings.Fields(stack); len(fields) > 1 && fields[0] == "goroutine" {
			waiting[fields[1]] = stack
		}
	}
	return waiting
}

// heldOutputReport says which goroutine is stuck, for how long, and who holds the other end of the pipes
// this process reads.
func heldOutputReport(stack string, waited time.Duration) string {
	var report strings.Builder
	fmt.Fprintf(&report, "a command exited and its output has stayed open at least %s since:\n\n%s\n\n",
		waited.Round(time.Second), stack)
	holders, err := pipeHolders()
	switch {
	case err != nil:
		fmt.Fprintf(&report, "the holders could not be listed: %v\n", err)
	case len(holders) == 0:
		report.WriteString("no other process holds the other end of any pipe this process reads, so the read end was most " +
			"likely closed out from under the poller: look for a second *os.File made on one descriptor (os.NewFile on a " +
			"descriptor something else owns), whose finalizer closes it again after its number was reused\n")
	default:
		report.WriteString("processes, other than this one's running children, holding the other end of a pipe this process reads:\n")
		for _, holder := range holders {
			fmt.Fprintf(&report, "  %s\n", holder)
		}
	}
	return report.String()
}

// pipeHolders lists, as "pid <pid> <command>: <parent, start and command line>", every process holding the
// far end of a pipe this process has open past its standard streams, leaving out this process's own running children: a command
// another test is still running holds its pipe as it should, and so does the lsof asking.
func pipeHolders() ([]string, error) {
	own, err := exec.Command("lsof", "-n", "-P", "-F", "pftdn", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		return nil, fmt.Errorf("lsof on this process: %w", err)
	}
	farEnds := map[string]bool{}
	for _, open := range lsofFiles(own) {
		// Not this process's own standard streams, which are pipes to whatever ran it. lsof leaves a pipe's
		// access blank on macOS, so the direction can't be read; every pipe exec opens to capture a command's
		// output is past these three.
		if open.kind == "PIPE" && open.descriptor != "0" && open.descriptor != "1" && open.descriptor != "2" &&
			strings.HasPrefix(open.name, "->") {
			farEnds[strings.TrimPrefix(open.name, "->")] = true
		}
	}
	if len(farEnds) == 0 {
		return nil, nil
	}
	everyone, err := exec.Command("lsof", "-n", "-P", "-F", "pcftdn").Output()
	if err != nil && len(everyone) == 0 {
		return nil, fmt.Errorf("lsof on every process: %w", err)
	}
	self := strconv.Itoa(os.Getpid())
	seen := map[string]bool{}
	var holders []string
	for _, open := range lsofFiles(everyone) {
		if open.kind != "PIPE" || !farEnds[open.device] || open.pid == self || seen[open.pid] {
			continue
		}
		seen[open.pid] = true
		started, _ := exec.Command("ps", "-o", "ppid=,lstart=,command=", "-p", open.pid).Output()
		// A process ps no longer finds has exited and holds nothing now, the lsof that listed it included.
		if fields := strings.Fields(string(started)); len(fields) == 0 || fields[0] == self {
			continue
		}
		holders = append(holders, fmt.Sprintf("pid %s %s: %s", open.pid, open.command, strings.TrimSpace(string(started))))
	}
	return holders, nil
}

// lsofFile is one open file from lsof's field output.
type lsofFile struct {
	pid, command, descriptor, kind, device, name string
}

// lsofFiles parses `lsof -F` output, where each line is one field: p and c begin a process, f begins a
// file, and t, d and n describe it.
func lsofFiles(output []byte) []lsofFile {
	var files []lsofFile
	var pid, command string
	for _, line := range bytes.Split(output, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		value := string(line[1:])
		switch line[0] {
		case 'p':
			pid, command = value, ""
		case 'c':
			command = value
		case 'f':
			files = append(files, lsofFile{pid: pid, command: command, descriptor: value})
		case 't', 'd', 'n':
			if len(files) == 0 {
				continue
			}
			last := &files[len(files)-1]
			switch line[0] {
			case 't':
				last.kind = value
			case 'd':
				last.device = value
			case 'n':
				last.name = value
			}
		}
	}
	return files
}

// heldOutputFiringBound is how long the watch may take to fire on its 2s deadline before the test calls it
// hung. It is not a speed check: load decides how long a healthy watch takes.
const heldOutputFiringBound = 2 * time.Minute

// failHeldOutput ends the test binary with the report, since no test can recover from a goroutine that
// never returns and waiting for -timeout only delays the same failure with less said.
func failHeldOutput(report string) {
	fmt.Fprintf(os.Stderr, "\nheld output: %s", report)
	os.Exit(2)
}

// The watch names the process that holds a command's output open after the command has exited. The
// command here is a shell that starts a sleep in the background and exits: the sleep inherits the pipe, so
// the test's wait for output hangs exactly as a leaked descendant would make it. Run in a child test
// binary, since the watch ends the binary it runs in.
//
// No bound here is a speed check, since load stretches the watch (#wwhrpm0). It took 25.27s at the load of
// the cold rebuilds after a cache clear, which outlasted a 25s sleep and a 20s bound, and failed a healthy
// run. The sleep is five minutes, so it outlasts the watch at any load seen, and the test kills it by the
// pid the report names. The bound on firing is two minutes, a hang detector, with go test's -timeout behind
// it.
func TestAHeldOutputNamesItsHolder(t *testing.T) {
	t.Parallel()
	if os.Getenv("COHERE_HELD_OUTPUT_HELPER") == "1" {
		watchForHeldOutput(2*time.Second, 200*time.Millisecond, failHeldOutput)
		exec.Command("/bin/sh", "-c", "sleep 300 & echo started").CombinedOutput()
		t.Fatal("the command's output closed while its background sleep still held it")
	}
	if runtime.GOOS == "windows" {
		t.Skip("lsof names a pipe's holders, and Windows has no lsof")
	}

	helper := exec.Command(os.Args[0], "-test.run=^TestAHeldOutputNamesItsHolder$", "-test.count=1")
	helper.Env = append(os.Environ(), "COHERE_HELD_OUTPUT_HELPER=1")
	outputFile, err := os.Create(t.TempDir() + "/output")
	if err != nil {
		t.Fatal(err)
	}
	defer outputFile.Close()
	// A file, not a buffer: the helper's sleep inherits whatever the helper writes to, and a buffer's pipe
	// would hang this test on the same sleep.
	helper.Stdout, helper.Stderr = outputFile, outputFile
	started := time.Now()
	err = helper.Run()
	output, _ := os.ReadFile(outputFile.Name())
	if sleeper := regexp.MustCompile(`pid (\d+) sleep`).FindSubmatch(output); sleeper != nil {
		if pid, convertError := strconv.Atoi(string(sleeper[1])); convertError == nil {
			if process, findError := os.FindProcess(pid); findError == nil {
				process.Kill()
			}
		}
	}
	if err == nil {
		t.Fatalf("the helper passed, so the watch never fired:\n%s", output)
	}
	if elapsed := time.Since(started); elapsed > heldOutputFiringBound {
		t.Errorf("the watch took %s to fire on a 2s deadline, past the %s that says it hung rather than ran slow",
			elapsed, heldOutputFiringBound)
	}
	for _, want := range []string{"held output: a command exited", "TestAHeldOutputNamesItsHolder", "sleep 300"} {
		if !bytes.Contains(output, []byte(want)) {
			t.Errorf("the report does not say %q:\n%s", want, output)
		}
	}
	if bytes.Contains(output, []byte(" lsof:")) {
		t.Errorf("the report names the lsof that asked, which is this process's own running child:\n%s", output)
	}
}
