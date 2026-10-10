// traced runs a command whose tests build products, under strace, and keys each product by what its build read
// (#vt46geg): only Workshop builds products, and only a traced build is placed under its key or published.
//
//	go run ./internal/buildcache/cmd/traced -- go test -count=1 -run '^TestProduct_' ./stage1/cohere/...
//
// Run it inside the adamic tree, with the environment the command builds in (ADAMIC_BUILD_CACHE_DIR, GOFLAGS,
// ADAMIC_BUILD_STORE for publishing). The command runs under strace -f with ADAMIC_BUILD_TRACE naming a trace
// directory under the build cache, so every process that builds a product marks it in the trace and journals it;
// the trace streams through a pipe into buildcache.Settle, never to disk. GOENV is off unless set, so go never reads
// a configuration file in the home directory, which no key can name.
//
// It prints one line per settled product (its key, how many reads, and how far its declared Files drift from them)
// and one per refused product, naming what it read that no key can name; refused products are Loom's to fix and are
// neither placed nor published. It exits with the command's status, or 3 when the command passed and a product was
// refused. Linux only: strace 6 with --seccomp-bpf and --decode-fds=path.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
		fmt.Fprintln(os.Stderr, "usage: traced -- <command> [arguments...]")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}
	status, err := run(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "traced:", err)
		os.Exit(2)
	}
	os.Exit(status)
}

func run(command []string) (int, error) {
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
		settlement, err := buildcache.Settle(directory, reader)
		done <- settled{settlement, err}
	}()
	strace := exec.Command("strace", append(append(append([]string{}, straceArguments...), "-o", pipe, "--"), command...)...)
	strace.Stdin, strace.Stdout, strace.Stderr = os.Stdin, os.Stdout, os.Stderr
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
	report := strings.Join(append(append([]string{}, result.settlement.Settled...), result.settlement.Refused...), "\n") + "\n"
	os.WriteFile(filepath.Join(directory, "settled.txt"), []byte(report), 0o644)
	fmt.Print(report)
	fmt.Printf("traced: %d settled, %d refused; journal and report in %s\n", len(result.settlement.Settled), len(result.settlement.Refused), directory)
	if status == 0 && len(result.settlement.Refused) > 0 {
		status = 3
	}
	return status, nil
}
