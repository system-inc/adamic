// skipcensus scans source or checks a go test -json log without caching results.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/system-inc/adamic/internal/skipcensus"
)

func main() {
	root := flag.String("root", ".", "repository root")
	table := flag.String("table", "internal/skipcensus/testdata/skips.json", "declarations")
	scan := flag.Bool("scan", false, "print AST census instead of checking a log")
	extra := flag.String("extra", "", "more declarations for checking the log only, for skips in the gated tree that the -root tree doesn't have")
	repository := flag.String("git", "", "a checkout whose origin answers whether a pending skip's awaited branch is on main")
	platform := flag.String("goos", runtime.GOOS, "the platform the log's tests ran on, which rows' platforms are held to (default: this machine's)")
	flag.Parse()
	if *scan {
		rows, err := skipcensus.Scan(*root)
		if err != nil {
			fatal(err)
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(rows); err != nil {
			fatal(err)
		}
		return
	}
	if flag.NArg() != 1 {
		fatal(fmt.Errorf("usage: skipcensus [-root directory] [-table file] log.jsonl"))
	}
	f, err := os.Open(*table)
	if err != nil {
		fatal(err)
	}
	rows, err := skipcensus.Load(f)
	f.Close()
	if err != nil {
		fatal(err)
	}
	actual, err := skipcensus.Scan(*root)
	if err != nil {
		fatal(err)
	}
	if err := skipcensus.Validate(actual, rows); err != nil {
		fatal(err)
	}
	log, err := os.Open(flag.Arg(0))
	if err != nil {
		fatal(err)
	}
	defer log.Close()
	if *extra != "" {
		f, err := os.Open(*extra)
		if err != nil {
			fatal(err)
		}
		more, err := skipcensus.Load(f)
		f.Close()
		if err != nil {
			fatal(err)
		}
		rows = append(rows, more...)
	}
	var landed skipcensus.Landed
	if *repository != "" {
		landed = landedOnMain(*repository)
	}
	if err := skipcensus.CheckLogOn(log, os.Stdout, rows, landed, *platform); err != nil {
		fatal(err)
	}
}

// landedOnMain asks origin for main's tip and the branch's, fetches both, and says whether main contains the
// branch's tip. A branch origin no longer has is an error: its reason can't be checked any more.
func landedOnMain(repository string) skipcensus.Landed {
	answers := map[string]bool{}
	return func(branch string) (bool, error) {
		if answer, ok := answers[branch]; ok {
			return answer, nil
		}
		main, err := remoteTip(repository, "main")
		if err != nil {
			return false, err
		}
		tip, err := remoteTip(repository, branch)
		if err != nil {
			return false, err
		}
		if output, err := exec.Command("git", "-C", repository, "fetch", "-q", "origin", main, tip).CombinedOutput(); err != nil {
			return false, fmt.Errorf("fetching main and %s: %v: %s", branch, err, strings.TrimSpace(string(output)))
		}
		err = exec.Command("git", "-C", repository, "merge-base", "--is-ancestor", tip, main).Run()
		var exit *exec.ExitError
		switch {
		case err == nil:
			answers[branch] = true
		case errors.As(err, &exit) && exit.ExitCode() == 1:
			answers[branch] = false
		default:
			return false, fmt.Errorf("is %s on main: %v", branch, err)
		}
		return answers[branch], nil
	}
}

func remoteTip(repository, branch string) (string, error) {
	output, err := exec.Command("git", "-C", repository, "ls-remote", "origin", "refs/heads/"+branch).Output()
	if err != nil {
		return "", fmt.Errorf("asking origin for %s: %v", branch, err)
	}
	fields := strings.Fields(string(output))
	if len(fields) < 2 || fields[1] != "refs/heads/"+branch {
		return "", fmt.Errorf("origin has no branch %s", branch)
	}
	return fields[0], nil
}
func fatal(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
