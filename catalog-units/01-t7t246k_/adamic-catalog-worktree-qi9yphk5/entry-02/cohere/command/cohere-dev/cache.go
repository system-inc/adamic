package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/system-inc/cohere/internal/release/gocache"
)

// The Go build cache holds itself under a cap (#jc6ca7r).
//
// Go trims its cache only of entries unused for five days. The house writes far faster than that: on
// 2026-10-05 the cache went from cleared at 09:01 to 200 GB by 11:30, and had been 347 GB with the disk at
// 99% before. The launcher's trim, since removed (#cpxc2d7), ran every few minutes all morning and
// removed almost nothing, because it spares every entry used in the last 90 minutes, and nearly all of a
// cache that grows by a hundred gigabytes an hour was used in the last 90 minutes. It has to spare them:
// Go reads an entry some time after finding it, and refreshes an entry's time at most once an hour, so a
// trim beside running builds cannot tell an entry a build is reading from one nobody wants. Removing one a
// build was reading failed it with "could not import ... no such file" (#3sgjy0h).
//
// So this trim keeps the builds out instead. When a token's run ends, at most once every ten minutes for
// the machine, cohere-dev starts a trim in the background. It measures the cache, and only when it is over
// the cap does it take every token in the pool, so no build that goes through cohere-dev is running, and
// trim the least recently used entries down to three quarters of the cap, sparing only the last
// cacheTrimSpare, which covers a build outside the pool that just wrote. Then it gives the tokens back.
// It waits for the pool to be empty holding nothing, and takes every token in one step only then, so it
// never keeps a token idle while another runs: taken one at a time with a wait for each, it held half the
// pool idle for minutes behind a land gate (2026-10-05 15:55). The trim itself takes seconds. A pool that
// is never empty for cacheTrimPatience is left alone, and the next look tries again. A quiet window, opened
// with `cohere-dev quiet on`, switches the trim off while it holds (quiet.go).

// cacheCapVariable sets the Go build cache's cap, in gigabytes, a fraction allowed.
const cacheCapVariable = "COHERE_DEV_CACHE_GB"

// defaultCacheGigabytes is the cap: a cold whole-module run writes about 11 GB, and a run after a change
// about 10 more, so 40 holds a few members' work since the last trim.
const defaultCacheGigabytes = 40

// cacheTrimInterval is how often the machine's cache is looked at, at most.
const cacheTrimInterval = 10 * time.Minute

// cacheTrimSpare is how recently written an entry may be and never be trimmed, while the pool is held:
// what a build outside the pool has just written and may be about to read.
const cacheTrimSpare = 10 * time.Minute

// cacheTrimPatience is how long the trim waits for an empty pool before it leaves this look to the next.
const cacheTrimPatience = 5 * time.Minute

// cacheTrimPoll is how often the trim looks for an empty pool while it waits.
const cacheTrimPoll = 250 * time.Millisecond

// trimCacheVerb runs cohere-dev as the background trim. It is what startCacheTrim starts.
const trimCacheVerb = "trim-cache"

func cacheCapBytes() (int64, error) {
	value := os.Getenv(cacheCapVariable)
	if value == "" {
		return defaultCacheGigabytes << 30, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s is %q, want a number of gigabytes above 0", cacheCapVariable, value)
	}
	return int64(parsed * (1 << 30)), nil
}

// startCacheTrim starts the background trim when the machine's cache is due a look, and does not wait for
// it. A run pays a stat; the trim is another process, its output nowhere, so no run waits on it.
func startCacheTrim(poolDirectory string) {
	if !filesLock || !claimCacheLook(poolDirectory, time.Now()) {
		return
	}
	executable, err := os.Executable()
	if err != nil {
		return
	}
	trim := exec.Command(executable, trimCacheVerb)
	trim.Env = outsideAToken(os.Environ())
	if err := trim.Start(); err != nil {
		return
	}
	trim.Process.Release()
}

// claimCacheLook reports whether the cache is due a look, and claims it, so runs ending together start one
// trim between them.
func claimCacheLook(poolDirectory string, now time.Time) bool {
	stamp := filepath.Join(poolDirectory, "gocache.checked")
	if information, err := os.Stat(stamp); err == nil && now.Sub(information.ModTime()) < cacheTrimInterval {
		return false
	}
	if os.WriteFile(stamp, nil, 0o644) != nil || os.Chtimes(stamp, now, now) != nil {
		return false
	}
	return true
}

// outsideAToken is an environment with no token held and Go unbudgeted.
func outsideAToken(environment []string) []string {
	kept := make([]string, 0, len(environment))
	for _, variable := range environment {
		if !strings.HasPrefix(variable, heldTokenVariable+"=") && !strings.HasPrefix(variable, "GOMAXPROCS=") {
			kept = append(kept, variable)
		}
	}
	return kept
}

// trimCache is the background trim's whole run: measure, and when over the cap, hold the pool and trim.
func trimCache() int {
	directory, err := poolDirectory()
	if err != nil {
		return 1
	}
	capBytes, err := cacheCapBytes()
	if err != nil {
		recordCacheLook(directory, err.Error())
		return 2
	}
	// One trim at a time, however many were started.
	lock, err := os.OpenFile(filepath.Join(directory, "gocache-trim.lock"), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return 1
	}
	defer lock.Close()
	if taken, err := tryLockFile(lock); err != nil || !taken {
		return 0
	}
	defer unlockFile(lock)

	// A quiet window holds off even the measuring, which walks every entry (quiet.go).
	holds, quietLine := quietHolds(directory, time.Now())
	if holds {
		recordCacheLook(directory, "not looked at: "+quietLine)
		return 0
	}
	if quietLine != "" {
		recordCacheLook(directory, quietLine)
	}

	cache, err := goCacheDirectory()
	if err != nil {
		recordCacheLook(directory, err.Error())
		return 1
	}
	if err := poolGuards(cache); err != nil {
		recordCacheLook(directory, err.Error())
		return 2
	}
	// Measured first, removing nothing, so a cache under its cap costs no one a token.
	measured, err := gocache.Trim(cache, gocache.Limit{Cap: math.MaxInt64}, time.Now())
	if err != nil {
		recordCacheLook(directory, err.Error())
		return 1
	}
	if measured.Before <= capBytes {
		recordCacheLook(directory, fmt.Sprintf("%s under the cap, nothing trimmed", gigabytes(measured.Before)))
		return 0
	}

	shape, err := readPoolShape()
	if err != nil {
		recordCacheLook(directory, err.Error())
		return 2
	}
	release, err := holdThePool(directory, shape.tokens, fmt.Sprintf("trimming the Go build cache, %s over the %s cap",
		gigabytes(measured.Before), gigabytes(capBytes)), cacheTrimPatience)
	if errors.Is(err, errPoolNeverEmpty) {
		recordCacheLook(directory, fmt.Sprintf("%s over the %s cap, not trimmed: the pool was never empty for %s, "+
			"and the next look tries again", gigabytes(measured.Before), gigabytes(capBytes), cacheTrimPatience))
		return 0
	}
	if err != nil {
		recordCacheLook(directory, "holding the pool to trim: "+err.Error())
		return 1
	}
	// A window opened while the trim waited for an empty pool holds too: the wait can be minutes.
	if holds, quietLine := quietHolds(directory, time.Now()); holds {
		release()
		recordCacheLook(directory, fmt.Sprintf("%s over the %s cap, not trimmed: %s", gigabytes(measured.Before), gigabytes(capBytes), quietLine))
		return 0
	}
	started := time.Now()
	trimmed, err := gocache.Trim(cache, gocache.Limit{Cap: capBytes, Target: capBytes / 4 * 3, Spare: cacheTrimSpare}, started)
	release()
	if err != nil {
		recordCacheLook(directory, err.Error())
		return 1
	}
	recordCacheLook(directory, fmt.Sprintf("trimmed from %s to %s, %d entries removed, %s written in the last %s kept; "+
		"the pool was held %s", gigabytes(trimmed.Before), gigabytes(trimmed.After), trimmed.Removed, gigabytes(trimmed.InUse),
		cacheTrimSpare, time.Since(started).Round(time.Millisecond)))
	return 0
}

// errPoolNeverEmpty is holdThePool's answer when every look found a token in use.
var errPoolNeverEmpty = errors.New("the pool was never empty")

// holdThePool takes every token in one step, once none is in use, and returns what gives them back. While
// any token is in use it holds none, so it never keeps a token idle while another runs; it looks again every
// cacheTrimPoll, for up to patience. Its holder lines say why, so a run waiting behind it knows.
func holdThePool(directory string, tokens int, why string, patience time.Duration) (func(), error) {
	deadline := time.Now().Add(patience)
	for {
		held, err := takeEveryToken(directory, tokens)
		if err != nil {
			return nil, err
		}
		if held != nil {
			line := fmt.Sprintf("pid %d since %s: %s", os.Getpid(), time.Now().Format("15:04:05"), why)
			for token := 1; token <= tokens; token++ {
				os.WriteFile(holderPath(directory, token), []byte(line+"\n"), 0o644)
			}
			return func() {
				for index, lock := range held {
					os.Remove(holderPath(directory, index+1))
					unlockFile(lock)
					lock.Close()
				}
			}, nil
		}
		if time.Now().After(deadline) {
			return nil, errPoolNeverEmpty
		}
		time.Sleep(cacheTrimPoll)
	}
}

// takeEveryToken takes every token, or none: a token found in use gives back the ones already taken at once,
// and it returns nil.
func takeEveryToken(directory string, tokens int) ([]*os.File, error) {
	held := []*os.File{}
	letGo := func() {
		for _, lock := range held {
			unlockFile(lock)
			lock.Close()
		}
	}
	for token := 1; token <= tokens; token++ {
		lock, err := os.OpenFile(lockPath(directory, token), os.O_CREATE|os.O_RDWR, 0o644)
		if err != nil {
			letGo()
			return nil, err
		}
		taken, err := tryLockFile(lock)
		if err != nil || !taken {
			lock.Close()
			letGo()
			return nil, err
		}
		held = append(held, lock)
	}
	return held, nil
}

// poolGuards refuses a cache outside the user cache directory the pool lives in. Holding the pool keeps out
// the builds that draw from it, and those are the builds whose cache sits beside it: Go puts its default
// cache in the same user cache directory. A run with `HOME` pointed at a scratch directory and the live
// `GOCACHE` passed through held a scratch pool, empty, at once, and trimmed the live cache from 112.9 GB to
// 30 GB under every real gate, its log in the scratch home (2026-10-06 04:44, #d0x1fhp). The pool no longer
// follows `HOME` (poolHome), and this is the guard behind that.
func poolGuards(cache string) error {
	home, err := poolHome()
	if err != nil {
		return fmt.Errorf("not trimmed: %w", err)
	}
	root := userCacheDirectoryIn(home)
	resolvedRoot, rootErr := filepath.EvalSymlinks(root)
	resolvedCache, cacheErr := filepath.EvalSymlinks(cache)
	if rootErr == nil && cacheErr == nil {
		if relative, err := filepath.Rel(resolvedRoot, resolvedCache); err == nil && relative != ".." &&
			!strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return nil
		}
	}
	return fmt.Errorf("not trimmed: the Go build cache %s is outside %s, the user cache directory this pool lives in, "+
		"so holding this pool would not keep out the builds reading that cache; trim it from a run whose home holds it", cache, root)
}

// goCacheDirectory is the cache a plain go command here uses: GOCACHE when it is set, Go's default
// otherwise.
func goCacheDirectory() (string, error) {
	output, err := exec.Command("go", "env", "GOCACHE").Output()
	if err != nil {
		return "", fmt.Errorf("asking go for its build cache: %w", err)
	}
	directory := strings.TrimSpace(string(output))
	if directory == "" || directory == "off" {
		return "", fmt.Errorf("go has no build cache here (GOCACHE=%q)", directory)
	}
	return directory, nil
}

// recordCacheLook keeps the last look at the cache, for cohere-dev status, and appends it to the trim log.
func recordCacheLook(directory string, what string) {
	line := time.Now().Format("2006-01-02 15:04:05") + " " + what
	temporary := filepath.Join(directory, ".gocache.status")
	if os.WriteFile(temporary, []byte(line+"\n"), 0o644) == nil {
		os.Rename(temporary, filepath.Join(directory, "gocache.status"))
	}
	if log, err := os.OpenFile(filepath.Join(directory, "gocache-trim.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
		fmt.Fprintln(log, line)
		log.Close()
	}
}

// cacheStatus is the cache's line in cohere-dev status.
func cacheStatus(directory string) string {
	capBytes, err := cacheCapBytes()
	if err != nil {
		return "Go build cache: " + err.Error()
	}
	last := "never looked at"
	if contents, err := os.ReadFile(filepath.Join(directory, "gocache.status")); err == nil {
		last = "last look " + strings.TrimSpace(string(contents))
	}
	return fmt.Sprintf("Go build cache: cap %s, looked at every %s at most; %s", gigabytes(capBytes), cacheTrimInterval, last)
}

func gigabytes(bytes int64) string {
	return strconv.FormatFloat(float64(bytes)/(1<<30), 'f', 1, 64) + " GB"
}
