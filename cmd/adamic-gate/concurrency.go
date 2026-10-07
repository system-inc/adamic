package main

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

const defaultConcurrency = "auto"

type concurrency struct {
	Setting           string
	Jobs              int
	Parallel          int // Zero leaves Go's -parallel flag unset.
	EffectiveParallel int
	Budget            int
}

var testParallel int
var concurrencySetting = defaultConcurrency

func init() {
	c, err := resolveConcurrency(defaultConcurrency, 0)
	if err != nil {
		panic(err)
	}
	packageJobs, testParallel = c.Jobs, c.Parallel
}

func resolveConcurrency(setting string, jobs int) (concurrency, error) {
	c := concurrency{Setting: setting, Jobs: runtime.GOMAXPROCS(0), EffectiveParallel: runtime.GOMAXPROCS(0)}
	if jobs < 0 {
		return c, fmt.Errorf("jobs must be positive")
	}
	if setting != "auto" {
		left, right, ok := strings.Cut(setting, "x")
		var err error
		c.Jobs, err = strconv.Atoi(left)
		if err != nil || !ok || c.Jobs < 1 || c.Jobs > 1024 {
			return c, fmt.Errorf("concurrency must be auto or positive JOBSxPARALLEL (each at most 1024)")
		}
		c.Parallel, err = strconv.Atoi(right)
		if err != nil || c.Parallel < 1 || c.Parallel > 1024 {
			return c, fmt.Errorf("concurrency must be auto or positive JOBSxPARALLEL (each at most 1024)")
		}
		c.EffectiveParallel = c.Parallel
		if jobs != 0 && jobs != c.Jobs {
			return c, fmt.Errorf("-jobs conflicts with -concurrency %s; use a single concurrency setting", setting)
		}
	} else if jobs != 0 {
		c.Jobs = jobs
	}
	c.Budget = c.Jobs * c.EffectiveParallel
	return c, nil
}

func currentConcurrency() concurrency {
	effective := testParallel
	if effective == 0 {
		effective = runtime.GOMAXPROCS(0)
	}
	return concurrency{concurrencySetting, packageJobs, testParallel, effective, packageJobs * effective}
}

func concurrencyArgs(args []string, parallel int) []string {
	if parallel == 0 {
		return args
	}
	out := append([]string{}, args[:len(args)-1]...)
	out = append(out, "-parallel="+strconv.Itoa(parallel), args[len(args)-1])
	return out
}

func validateConcurrency(c concurrency) error {
	if c.Jobs < 1 || c.Parallel < 0 || c.EffectiveParallel < 1 || c.Budget != c.Jobs*c.EffectiveParallel {
		return fmt.Errorf("invalid concurrency evidence")
	}
	if c.Setting == "auto" {
		if c.Parallel != 0 {
			return fmt.Errorf("auto concurrency must leave -parallel unset")
		}
		return nil
	}
	want, err := resolveConcurrency(c.Setting, c.Jobs)
	if err != nil || want != c {
		return fmt.Errorf("concurrency evidence differs from setting %q", c.Setting)
	}
	return nil
}
