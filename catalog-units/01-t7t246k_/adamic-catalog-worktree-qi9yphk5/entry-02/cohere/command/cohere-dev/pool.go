package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// The machine's budget for heavy Go work (#qhg0ntb).
//
// Every go command started through cohere-dev, of any size, and every command run with `cohere-dev exec`,
// takes one of a few tokens and runs with Go held to a share of the cores. Before, only whole-module test
// runs were held to one at a time and everything else ran unbounded, each go command on all 16 cores, so a
// dozen members meant a dozen times sixteen threads: load sat at 110 to 180 all morning on 2026-10-05, a
// 27s gate took 7 to 10 minutes, a package's CPU inflated about nineteen times, wall-clock deadlines in
// tests failed, and Kirk's typing lagged.
//
// A token is a slot lock (slot-N.lock in the pool directory), taken in arrival order (queueTicket). Go is
// held to the token's share twice over: GOFLAGS' -p bounds how many packages it compiles and how many test
// binaries it runs at once, and GOMAXPROCS bounds the threads each of those processes uses, so a token
// costs about -p times GOMAXPROCS threads. Both are inherited by every go command a test or script starts
// beneath it. GOMAXPROCS alone is not a budget: go's -p defaults to it, so 8 meant up to eight test binaries
// at eight threads each, and the measured load stayed at 107 to 140 on sixteen cores (#qhg0ntb).
// A command that already runs under a token, a test that runs go build, say, runs inside its parent's share
// and takes no second token, which would also deadlock a pool its parent has already filled.

// tokensVariable sets how many tokens the machine has.
const tokensVariable = "COHERE_DEV_TOKENS"

// packagesVariable sets how many packages each token's run builds or tests at once: go's -p.
const packagesVariable = "COHERE_DEV_PACKAGES"

// threadsVariable sets how many threads each of those processes may use: GOMAXPROCS.
const threadsVariable = "COHERE_DEV_THREADS"

// heldTokenVariable is set, to the token's number, in the environment of everything a token's run starts.
const heldTokenVariable = "COHERE_DEV_TOKEN"

// coresVariable is set, in the environment of everything a token's run starts, to the cores the token is
// worth: its packages times its threads. A runner that is not Go (Jest's workers, node's test concurrency)
// sizes itself by it, through `cohere-dev exec --`, without knowing Go's split (#wrawfdb).
const coresVariable = "COHERE_DEV_CORES"

// The pool's shape: four runs at once, each two packages at a time on two threads, sixteen threads on a
// sixteen-core machine. Measured in a quiet window on 2026-10-05 replaying the morning's mix, four gates and
// two builds started at once: this shape finished in 112s at a mean load of 28, against 113s at 119 for
// four tokens bounded by GOMAXPROCS alone, 130s at 140 for two of eight, and 125s at 191 with no pool, where
// all four gates failed on a test's deadline (#qhg0ntb).
const (
	defaultTokens   = 4
	defaultPackages = 2
	defaultThreads  = 2
)

// poolShape is the pool's tokens, and how many packages and threads each token's run may use.
type poolShape struct {
	tokens, packages, threads int
}

// readPoolShape is the pool's shape, from the defaults and whatever the environment changes.
func readPoolShape() (poolShape, error) {
	var shape poolShape
	var err error
	if shape.tokens, err = positiveVariable(tokensVariable, defaultTokens); err != nil {
		return shape, err
	}
	if shape.packages, err = positiveVariable(packagesVariable, defaultPackages); err != nil {
		return shape, err
	}
	if shape.threads, err = positiveVariable(threadsVariable, defaultThreads); err != nil {
		return shape, err
	}
	return shape, nil
}

func positiveVariable(name string, fallback int) (int, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("%s is %q, want a whole number of at least 1", name, value)
	}
	return parsed, nil
}

// homeVariable names the home the pool hangs from in place of the account's own, so a test can have a pool of
// its own. Nothing else moves the pool: it never follows `HOME` (poolHome).
const homeVariable = "COHERE_DEV_HOME"

// poolHome is the home the machine's pool hangs from: the account's own, from the system's user database,
// never the `HOME` variable. A gate run with `HOME` set to a scratch directory, to leave its corpora unset,
// and the go cache passed through, opened a pool of its own and so ran outside the machine's, and its trim
// held that empty pool and trimmed the live cache under every real gate (2026-10-06 04:44, #d0x1fhp).
func poolHome() (string, error) {
	if home := os.Getenv(homeVariable); home != "" {
		return home, nil
	}
	account, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("finding this account's home for the machine's pool: %w", err)
	}
	return account.HomeDir, nil
}

// userCacheDirectoryIn is the user cache directory under a home, where Go puts its default build cache. On
// Linux it is ~/.cache whatever `XDG_CACHE_HOME` says, which moves Go's cache but not the pool, so a cache
// moved that way is refused by the trim rather than trimmed (poolGuards).
func userCacheDirectoryIn(home string) string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Caches")
	case "windows":
		return filepath.Join(home, "AppData", "Local")
	}
	return filepath.Join(home, ".cache")
}

// poolDirectory is where the tokens, their holders and the line live, one per machine user.
func poolDirectory() (string, error) {
	home, err := poolHome()
	if err != nil {
		return "", err
	}
	// The directory the whole-module slot used, so a checkout still on that wrapper and one on this pool take
	// the same lock for slot 1 and never run at once on it.
	directory := filepath.Join(userCacheDirectoryIn(home), "cohere", "test-slots")
	return directory, os.MkdirAll(directory, 0o755)
}

// withToken runs work holding a token, in arrival order, with the environment its commands must run in.
// what describes the work for the line other waiters read. It returns work's exit code.
func withToken(what string, work func(environment []string) int) int {
	if held := os.Getenv(heldTokenVariable); held != "" {
		return work(os.Environ())
	}
	shape, err := readPoolShape()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cohere-dev: %v\n", err)
		return 2
	}
	tokens := shape.tokens
	directory, err := poolDirectory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cohere-dev: %v\n", err)
		return 1
	}

	held, err := waitInLine(directory, tokens, what, "token", func() {
		fmt.Fprintf(os.Stderr, "cohere-dev: waiting for a token; the machine runs %d heavy Go runs at once, each %d packages "+
			"at a time on %d threads, and these hold them (cohere-dev status shows the line):\n", tokens, shape.packages, shape.threads)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "cohere-dev: %v\n", err)
		return 1
	}
	code := work(budgetEnvironment(os.Environ(), held.number, shape))
	held.release()
	// The run is over and its token free, so this is when the cache is looked at (cache.go).
	startCacheTrim(directory)
	return code
}

// waitInLine takes one of a directory's slots in arrival order, waiting its turn, and returns it held. The
// slots are the pool's tokens, or land's one lock (land.go). noun names a slot in what a waiting run prints,
// and explain opens what it prints when it first finds every slot held, before the holders' lines.
func waitInLine(directory string, slots int, what string, noun string, explain func()) (*heldSlot, error) {
	// In line, in arrival order, where a lock holds; see queueTicket.
	var ticket *queueTicket
	var err error
	if filesLock {
		if ticket, err = joinQueue(directory); err != nil {
			return nil, fmt.Errorf("joining the line for a %s: %w", noun, err)
		}
		// Left as soon as a slot is taken, so the next in line moves up while this one runs; this covers
		// every way out before that.
		defer func() {
			if ticket != nil {
				ticket.leave()
			}
		}()
	}

	waitingSince := time.Now()
	announced := false
	lastAhead := -1
	// The holders as last printed, so a change of holder is printed again: a waiter that named only the first
	// holder it saw went on naming it after it exited (lint_rules' land, a dead pid named while another held
	// the lock, 2026-10-06).
	lastHolders := ""
	for {
		ahead, waiting := 0, 0
		if ticket != nil {
			if ahead, waiting, err = ticket.place(); err != nil {
				return nil, fmt.Errorf("reading the line for a %s: %w", noun, err)
			}
		}
		// Only the first in line, as many as there are slots, may take one, so no later arrival passes them.
		if ahead < slots {
			held, err := takeSlot(directory, slots, what)
			if err != nil {
				return nil, fmt.Errorf("taking a %s: %w", noun, err)
			}
			if held != nil {
				if ticket != nil {
					ticket.leave()
					ticket = nil
				}
				if announced {
					fmt.Fprintf(os.Stderr, "cohere-dev: took %s %d after waiting %s\n", noun, held.number, time.Since(waitingSince).Round(time.Second))
				}
				return held, nil
			}
		}
		current := holders(directory, slots)
		if !announced {
			explain()
			for _, holder := range current {
				fmt.Fprintf(os.Stderr, "  %s\n", holder)
			}
			announced = true
			lastHolders = strings.Join(current, "\n")
		} else if joined := strings.Join(current, "\n"); joined != lastHolders && len(current) > 0 {
			fmt.Fprintf(os.Stderr, "cohere-dev: the %s holders changed; now:\n", noun)
			for _, holder := range current {
				fmt.Fprintf(os.Stderr, "  %s\n", holder)
			}
			lastHolders = joined
		}
		if ticket != nil && ahead != lastAhead {
			fmt.Fprintf(os.Stderr, "cohere-dev: %s in line of %d waiting\n", ordinal(ahead+1), waiting)
			lastAhead = ahead
		}
		time.Sleep(time.Second)
	}
}

// budgetEnvironment is the environment a token's run starts its commands in: Go held to the token's packages
// and threads, the caller's other GOFLAGS kept, and the token named, so a command started beneath it runs in
// this share rather than taking another.
func budgetEnvironment(base []string, token int, shape poolShape) []string {
	environment := make([]string, 0, len(base)+3)
	flags := []string{}
	for _, variable := range base {
		name, value, _ := strings.Cut(variable, "=")
		switch name {
		case "GOMAXPROCS", heldTokenVariable, coresVariable:
			continue
		case "GOFLAGS":
			for _, flag := range strings.Fields(value) {
				if flag != "-p" && !strings.HasPrefix(flag, "-p=") {
					flags = append(flags, flag)
				}
			}
			continue
		}
		environment = append(environment, variable)
	}
	flags = append(flags, "-p="+strconv.Itoa(shape.packages))
	return append(environment, "GOFLAGS="+strings.Join(flags, " "), "GOMAXPROCS="+strconv.Itoa(shape.threads),
		heldTokenVariable+"="+strconv.Itoa(token), coresVariable+"="+strconv.Itoa(shape.packages*shape.threads))
}

// status prints the pool: its shape, who holds each token, and who waits in line, oldest first.
func status() int {
	shape, err := readPoolShape()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cohere-dev: %v\n", err)
		return 2
	}
	directory, err := poolDirectory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cohere-dev: %v\n", err)
		return 1
	}
	fmt.Printf("pool: %d tokens, each %d packages at a time on %d threads, nice %d (%s)\n", shape.tokens, shape.packages,
		shape.threads, testNiceness, directory)
	for token := 1; token <= shape.tokens; token++ {
		holder := "free"
		if held := heldBy(directory, token); held != "" {
			holder = held
		}
		fmt.Printf("  token %d: %s\n", token, holder)
	}
	fmt.Println(cacheStatus(directory))
	fmt.Println(quietStatus(directory, time.Now()))
	waiting := waiters(directory)
	fmt.Printf("line: %d waiting\n", len(waiting))
	for place, waiter := range waiting {
		fmt.Printf("  %s: %s\n", ordinal(place+1), waiter)
	}
	// Land's one lock and its own line (land.go), so a stalled landing shows here beside the tokens.
	land, err := landDirectory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cohere-dev: %v\n", err)
		return 1
	}
	holder := "free"
	if held := heldBy(land, 1); held != "" {
		holder = held
	}
	fmt.Printf("land lock: %s\n", holder)
	landing := waiters(land)
	fmt.Printf("land line: %d waiting\n", len(landing))
	for place, waiter := range landing {
		fmt.Printf("  %s: %s\n", ordinal(place+1), waiter)
	}
	return 0
}

// heldBy says who holds a token, or nothing when it is free: its lock is what says so, not its holder file,
// which a run killed outright leaves behind.
func heldBy(directory string, token int) string {
	lock, err := os.OpenFile(lockPath(directory, token), os.O_RDWR, 0)
	if err != nil {
		return ""
	}
	defer lock.Close()
	if taken, _ := tryLockFile(lock); taken {
		unlockFile(lock)
		return ""
	}
	contents, _ := os.ReadFile(holderPath(directory, token))
	if line := strings.TrimSpace(string(contents)); line != "" {
		return line
	}
	return "held, by a run that has not named itself yet"
}

// waiters names every live ticket in line, oldest first, by its pid.
func waiters(directory string) []string {
	entries, err := os.ReadDir(queueDirectory(directory))
	if err != nil {
		return nil
	}
	lines := []string{}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		ticket, err := os.OpenFile(filepath.Join(queueDirectory(directory), name), os.O_RDWR, 0)
		if err != nil {
			continue
		}
		if taken, _ := tryLockFile(ticket); taken {
			// A waiter that is gone; the next waiter's look at the line removes it.
			unlockFile(ticket)
			ticket.Close()
			continue
		}
		ticket.Close()
		arrived, pid, _ := strings.Cut(name, "-")
		since := ""
		if nanoseconds, err := strconv.ParseInt(arrived, 10, 64); err == nil {
			since = " since " + time.Unix(0, nanoseconds).Format("15:04:05")
		}
		lines = append(lines, "pid "+pid+since)
	}
	return lines
}
