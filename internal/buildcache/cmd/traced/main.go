// traced runs a command whose tests build products, under strace, and keys each product by what its build read
// (#vt46geg): only Workshop builds products, and only a traced build is placed under its key.
//
//	go run ./internal/buildcache/cmd/traced -- go test -count=1 -run '^TestProduct_' ./stage1/cohere/...
//
// Run it inside the adamic tree, with the environment the command builds in (ADAMIC_BUILD_CACHE_DIR, GOFLAGS).
// Workshop's tree builder sets ADAMIC_BUILD_STORE=traced, which refuses any product built outside a trace; what settles
// stays in the cache, which Loom ships, read sets and all, and nothing is written to the store. The command runs under
// strace -f with ADAMIC_BUILD_TRACE naming a trace directory under the build cache, so every process that builds a
// product marks it in the trace and journals it; the trace streams through a pipe into buildcache.Settle, never to disk.
// GOENV is off unless set, so go never reads a configuration file in the home directory, which no key can name.
//
// The tree is taken as git sees it (HEAD, status, each submodule) before the run and after: a read set is valued when
// the run ends, so a run whose tree changed meanwhile refuses every product, as does one that wrote into the tree.
// traced -tree-state <file> takes it and writes it to the file, nothing more; a run with ADAMIC_TRACE_TREE_STATE naming
// that file starts from it rather than taking its own (Loom's build-tree, once for each wave of product tests that start
// together, so a wave's start never forks git in the hundreds, #2q48btg).
//
// traced -machinery reads paths on stdin, one a line, slash-separated from the tree's root, and prints those that are
// the build machinery (buildcache.IsMachinery), nothing more, on any system: Loom seeds a tree from another only when
// the diff between them names none (#kjgmzt8), asking this tree's traced and main's, so the machinery is defined once,
// here, and grows with the split's exclusions. A traced without the flag exits 2, and Loom then seeds strictly.
//
// It prints one line per settled product (its key, how many reads, and how far its declared Files drift from them) and
// one per refused product, naming what it read that no key can name; refused products are Loom's to fix and are never
// placed. It exits with the command's status, or 3 when the command passed and a product was refused. Linux only:
// strace 6 with --seccomp-bpf and --decode-fds=path.
//
// With ADAMIC_TRACE_REPORT naming a file, it also writes there one JSON line per product (buildcache.Report: name,
// nameKey, key and reads when settled, reason when refused), for Loom's tree builder. A run that can't be keyed at all
// writes one line more, outcome lost, first; one that can't trace or settle (exit 2) writes only that line; one strace
// failed under before the command exited (exit 2) writes it first, then whatever settled before. The file is
// emptied before the command runs, so a report left by an earlier run never reads as this one's.
package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

// The calls strace reports: every call that names a path (open, stat, access, readlink, exec, chdir, ...), directory
// listings, fchdir and forks, failed calls too (a missing file the build looked for is a read). Filtered in the
// kernel (--seccomp-bpf), so untraced calls cost nothing; --decode-fds=path prints descriptors as their paths.
var straceArguments = []string{"-f", "--seccomp-bpf", "--decode-fds=path", "-q", "-e", "signal=none", "-e",
	"trace=%file,getdents64,fchdir,clone,clone3,fork,vfork"}

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: traced -- <command> [arguments...]\n       traced -tree-state <file>\n       traced -machinery < paths")
		flag.PrintDefaults()
	}
	treeState := flag.String("tree-state", "", "write the tree's state to this file and exit, for every run of a wave to start from (ADAMIC_TRACE_TREE_STATE)")
	machinery := flag.Bool("machinery", false, "read paths on stdin, one a line, and print those that are the build machinery (buildcache.IsMachinery), for Loom's seed rule")
	flag.Parse()
	if *machinery {
		if flag.NArg() != 0 || *treeState != "" {
			flag.Usage()
			os.Exit(2)
		}
		if err := printMachinery(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "traced:", err)
			os.Exit(2)
		}
		return
	}
	if *treeState != "" && flag.NArg() == 0 {
		if err := writeTreeState(*treeState); err != nil {
			fmt.Fprintln(os.Stderr, "traced:", err)
			os.Exit(2)
		}
		return
	}
	if flag.NArg() == 0 || *treeState != "" {
		flag.Usage()
		os.Exit(2)
	}
	report := os.Getenv("ADAMIC_TRACE_REPORT")
	if report != "" {
		if absolute, err := filepath.Abs(report); err == nil {
			report = absolute
		}
		if err := os.WriteFile(report, nil, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "traced: ADAMIC_TRACE_REPORT:", err)
			os.Exit(2)
		}
	}
	status, err := run(flag.Args(), report)
	if err != nil {
		fmt.Fprintln(os.Stderr, "traced:", err)
		if report != "" {
			buildcache.WriteReport(report, []buildcache.Report{{Outcome: buildcache.ReportLost, Reason: "traced: " + err.Error()}})
		}
		os.Exit(2)
	}
	os.Exit(status)
}

// printMachinery writes each line of paths that names the build machinery, in order, to out.
func printMachinery(paths io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(paths)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	writer := bufio.NewWriter(out)
	for scanner.Scan() {
		if buildcache.IsMachinery(scanner.Text()) {
			fmt.Fprintln(writer, scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return writer.Flush()
}

func run(command []string, report string) (int, error) {
	if runtime.GOOS != "linux" {
		return 0, errors.New("tracing is strace's, on Linux (Workshop); here a product builds only with ADAMIC_BUILD_STORE=off, for this machine alone")
	}
	probe := exec.Command("strace", append(append([]string{}, straceArguments...), "-o", "/dev/null", "true")...)
	if output, err := probe.CombinedOutput(); err != nil {
		return 0, fmt.Errorf("this strace can't trace as Settle needs (strace 6 with --seccomp-bpf and --decode-fds=path): %v\n%s", err, output)
	}
	cache := os.Getenv("ADAMIC_BUILD_CACHE_DIR")
	if cache == "" {
		user, err := os.UserCacheDir()
		if err != nil {
			return 0, err
		}
		cache = filepath.Join(user, "adamic-build")
	}
	if err := os.MkdirAll(filepath.Join(cache, "traces"), 0o755); err != nil {
		return 0, err
	}
	directory, err := os.MkdirTemp(filepath.Join(cache, "traces"), time.Now().UTC().Format("20060102T150405Z")+"-")
	if err != nil {
		return 0, err
	}
	pipe := filepath.Join(directory, "trace.fifo")
	if err = syscall.Mkfifo(pipe, 0o600); err != nil {
		return 0, err
	}
	before, err := treeStateBefore(buildcache.TreeState)
	if err != nil {
		return 0, fmt.Errorf("the tree's state before the run: %v", err)
	}
	// Settle sees the environment the command runs in: the same cache, the same store, the same trace.
	os.Setenv("ADAMIC_BUILD_TRACE", directory)
	if os.Getenv("GOENV") == "" {
		os.Setenv("GOENV", "off")
	}
	type settled struct {
		settlement buildcache.Settlement
		err        error
	}
	done := make(chan settled, 1)
	go func() {
		// Opening blocks until strace opens the pipe to write; the trace ends when strace exits.
		reader, err := os.Open(pipe)
		if err != nil {
			done <- settled{err: err}
			return
		}
		defer reader.Close()
		settlement, err := buildcache.Settle(directory, reader, before)
		done <- settled{settlement, err}
	}()
	strace := exec.Command("strace", append(append(append([]string{}, straceArguments...), "-o", pipe, "--"), command...)...)
	complaints := &straceComplaints{}
	strace.Stdin, strace.Stdout, strace.Stderr = os.Stdin, os.Stdout, io.MultiWriter(os.Stderr, complaints)
	status := 0
	if err = strace.Run(); err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			// strace never opened the pipe: open it ourselves so the reader ends.
			if writer, openErr := os.OpenFile(pipe, os.O_WRONLY, 0); openErr == nil {
				writer.Close()
			}
			<-done
			return 0, err
		}
		status = exit.ExitCode()
	}
	result := <-done
	if result.err != nil {
		return 0, fmt.Errorf("settling %s: %v", directory, result.err)
	}
	// strace exits with the command's status, but exits 1 when ptrace(PTRACE_LISTEN) fails on a tracee already gone, which
	// loses nothing: a tracee still running would never have been resumed, and the run couldn't have ended (home-7's dry
	// run, Oct 11: 3 of 454 tests exited 1 though go test passed and their products settled). Then the run's status is
	// the command's own exit, as the trace holds it. After any other complaint of strace's, the trace may be missing
	// events, so strace's status stands and the run reads as failed.
	if result.settlement.CommandExited && status != result.settlement.CommandStatus && complaints.onlyListen() {
		fmt.Fprintf(os.Stderr, "traced: strace exited %d after %d failed PTRACE_LISTEN on tracees already gone; the command exited %d, the run's status\n",
			status, len(complaints.lines), result.settlement.CommandStatus)
		status = result.settlement.CommandStatus
	}
	lost := tracerLost(result.settlement.CommandExited, status, complaints)
	if lost != "" {
		fmt.Fprintln(os.Stderr, "traced:", lost)
		result.settlement.Reports = append([]buildcache.Report{{Outcome: buildcache.ReportLost, Reason: "traced: " + lost}}, result.settlement.Reports...)
	}
	if report != "" {
		if err = buildcache.WriteReport(report, result.settlement.Reports); err != nil {
			return 0, fmt.Errorf("ADAMIC_TRACE_REPORT: %v", err)
		}
	}
	lines := append(append([]string{}, result.settlement.Settled...), result.settlement.Refused...)
	if result.settlement.TestReads != "" {
		lines = append(lines, result.settlement.TestReads)
	}
	text := strings.Join(lines, "\n") + "\n"
	os.WriteFile(filepath.Join(directory, "settled.txt"), []byte(text), 0o644)
	fmt.Print(text)
	fmt.Printf("traced: %d settled, %d refused; journal and report in %s\n", len(result.settlement.Settled), len(result.settlement.Refused), directory)
	if lost != "" {
		return 2, nil
	}
	if status == 0 && len(result.settlement.Refused) > 0 {
		status = 3
	}
	return status, nil
}

// writeTreeState takes the tree's state and writes it to path whole: it lands under its own name only once written, so a
// run starting from it never reads half of one.
func writeTreeState(path string) error {
	state, err := buildcache.TreeState()
	if err != nil {
		return fmt.Errorf("the tree's state: %v", err)
	}
	partial := path + ".partial"
	if err = os.WriteFile(partial, []byte(state), 0o644); err != nil {
		return err
	}
	return os.Rename(partial, path)
}

// treeStateBefore is the tree's state before the run: the one ADAMIC_TRACE_TREE_STATE's file holds, taken by -tree-state
// before the wave this run starts in, or one take takes now. A state taken earlier only widens what must hold still
// until the run ends, so a change since it was taken still refuses every product. The variable is unset either way, so
// nothing the command runs sees it, or keys the file it names.
func treeStateBefore(take func() (string, error)) (string, error) {
	path := os.Getenv("ADAMIC_TRACE_TREE_STATE")
	os.Unsetenv("ADAMIC_TRACE_TREE_STATE")
	if path == "" {
		return take()
	}
	state, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("ADAMIC_TRACE_TREE_STATE: %v", err)
	}
	if len(state) == 0 {
		return "", fmt.Errorf("ADAMIC_TRACE_TREE_STATE: %s holds no state", path)
	}
	return string(state), nil
}

// tracerLost is why a run was lost to strace itself, or "": strace complained and failed before the trace held the
// command's exit, so the trace is cut short and the command died with it or ran on untraced (Workshop, Oct 14, #j4hwdks:
// strace 6.8's PTRACE_LISTEN failing on WSL2 mid-fill, the fill reading 0 settled, 0 refused, its product never recorded
// and a read-mode unit failing on it). It's the tracer's failure, never the command's: the report's first line says the
// run was lost and traced exits 2, as for a run it couldn't trace, so the caller traces it again. What settled before
// strace failed stands, each product's build traced whole.
func tracerLost(commandExited bool, status int, complaints *straceComplaints) string {
	if commandExited || status == 0 || len(complaints.lines) == 0 {
		return ""
	}
	return fmt.Sprintf("strace exited %d before the command did, after %s, so the trace is cut short", status, strings.Join(complaints.lines, "; "))
}

// straceComplaints keeps strace's own lines from the stderr it shares with the command: each one it prints as
// "strace: ...".
type straceComplaints struct {
	partial []byte
	lines   []string
}

func (complaints *straceComplaints) Write(data []byte) (int, error) {
	complaints.partial = append(complaints.partial, data...)
	for {
		end := bytes.IndexByte(complaints.partial, '\n')
		if end < 0 {
			return len(data), nil
		}
		if line := string(complaints.partial[:end]); strings.HasPrefix(line, "strace: ") {
			complaints.lines = append(complaints.lines, line)
		}
		complaints.partial = complaints.partial[end+1:]
	}
}

var listenFailed = regexp.MustCompile(`^strace: ptrace\(PTRACE_LISTEN,pid:\d+,sig:\d+\): Input/output error$`)

// onlyListen says strace complained, and only of PTRACE_LISTEN failing with EIO.
func (complaints *straceComplaints) onlyListen() bool {
	for _, line := range complaints.lines {
		if !listenFailed.MatchString(line) {
			return false
		}
	}
	return len(complaints.lines) > 0
}
