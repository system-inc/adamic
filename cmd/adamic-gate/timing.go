package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type timingObservation struct {
	Log          string
	Seconds      float64
	SetupSeconds float64
	Children     map[string]float64
}
type timingSource struct {
	Path, SHA256       string
	Commit, BuildFlags string
}
type timingCalibration struct {
	Sources      []timingSource
	Method       string
	Observations map[string][]timingObservation
	Missing      []string
}

func median(values []float64) float64 {
	sort.Float64s(values)
	n := len(values)
	if n == 0 {
		return 0
	}
	if n%2 == 1 {
		return values[n/2]
	}
	return (values[n/2-1] + values[n/2]) / 2
}

// Each input is one shard invocation log. Repeated parents must not overwrite one another.
func calibrateTimings(paths []string) (map[string]float64, timingCalibration, error) {
	audit := timingCalibration{Method: "median shard observations; split parents: median residual plus union of child medians; layout: sum member medians less per-member fixture span plus one median fixture span", Observations: map[string][]timingObservation{}}
	weights := map[string]float64{}
	for _, path := range paths {
		digest, err := fileDigest(path)
		if err != nil {
			return nil, audit, err
		}
		source := timingSource{Path: path, SHA256: digest}
		var recorded summary
		if err := loadJSON(filepath.Join(filepath.Dir(path), "summary.json"), &recorded); err == nil {
			source.Commit = recorded.Plan.Commit
			source.BuildFlags = recorded.BuildFlags
		} else if !os.IsNotExist(err) {
			return nil, audit, err
		}
		audit.Sources = append(audit.Sources, source)
		f, err := os.Open(path)
		if err != nil {
			return nil, audit, err
		}
		scan := bufio.NewScanner(f)
		scan.Buffer(make([]byte, 65536), 16*1024*1024)
		observations := map[string]*timingObservation{}
		starts := map[string]time.Time{}
		setups := map[string]float64{}
		children := map[string]map[string]float64{}
		flush := func(pkg string) {
			for key, o := range observations {
				if pkg != "" && !strings.HasPrefix(key, pkg+"::") {
					continue
				}
				o.SetupSeconds = setups[key]
				o.Children = children[key]
				audit.Observations[key] = append(audit.Observations[key], *o)
				delete(observations, key)
				delete(children, key)
				delete(setups, key)
				delete(starts, key)
			}
		}
		for scan.Scan() {
			var e event
			if err = json.Unmarshal(scan.Bytes(), &e); err != nil {
				f.Close()
				return nil, audit, fmt.Errorf("%s: %w", path, err)
			}
			key := e.Package + "::" + e.Test
			if e.Action == "fail" {
				f.Close()
				return nil, audit, fmt.Errorf("%s: failed invocation %s", path, key)
			}
			if e.Test == "" {
				if e.Action == "start" || e.Action == "pass" || e.Action == "skip" {
					flush(e.Package)
				}
				continue
			}
			if e.Action == "cont" {
				starts[key] = e.Time
			}
			if e.Action == "output" && strings.Contains(e.Output, "original fork full Markdown parsing/layout off agrees on all") {
				if start, ok := starts[key]; ok && !e.Time.IsZero() {
					setups[key] = e.Time.Sub(start).Seconds()
				}
			}
			if e.Test == "" || (e.Action != "pass" && e.Action != "skip") {
				continue
			}
			observations[key] = &timingObservation{Log: path, Seconds: e.Elapsed}
			if strings.Contains(e.Test, "/") {
				parent := e.Package + "::" + strings.Split(e.Test, "/")[0]
				if children[parent] == nil {
					children[parent] = map[string]float64{}
				}
				children[parent][key] = e.Elapsed
			}
		}
		err = scan.Err()
		f.Close()
		if err != nil {
			return nil, audit, err
		}
		flush("")
	}
	for key, observations := range audit.Observations {
		values := []float64{}
		hasChildren := false
		for _, o := range observations {
			values = append(values, o.Seconds)
			hasChildren = hasChildren || len(o.Children) > 0
		}
		if !hasChildren || len(observations) == 1 {
			weights[key] = median(values)
			continue
		}
		residuals := []float64{}
		parallel := false
		union := map[string][]float64{}
		for _, o := range observations {
			sum := 0.0
			childKeys := []string{}
			for child := range o.Children {
				childKeys = append(childKeys, child)
			}
			sort.Strings(childKeys)
			for _, child := range childKeys {
				seconds := o.Children[child]
				sum += seconds
				union[child] = append(union[child], seconds)
			}
			residual := o.Seconds - sum
			if residual < 0 {
				parallel = true
				residual = 0
			}
			residuals = append(residuals, residual)
		}
		if parallel {
			weights[key] = median(values)
			continue
		}
		seconds := median(residuals)
		childKeys := []string{}
		for child := range union {
			childKeys = append(childKeys, child)
		}
		sort.Strings(childKeys)
		for _, child := range childKeys {
			seconds += median(union[child])
		}
		weights[key] = seconds
	}
	groups := knownAffinities()
	for _, g := range planningAffinities() {
		if g.Split != "" {
			groups = append(groups, g)
		}
	}
	for _, g := range groups {
		if g.WholePackage {
			for key := range weights {
				if strings.HasPrefix(key, g.Package+"::") {
					test := strings.TrimPrefix(key, g.Package+"::")
					if !strings.Contains(test, "/") {
						g.Tests = append(g.Tests, test)
					}
				}
			}
			sort.Strings(g.Tests)
		}
		sum := 0.0
		observed := 0
		setups := []float64{}
		for _, name := range g.Tests {
			key := g.Package + "::" + name
			seconds, ok := weights[key]
			if !ok {
				audit.Missing = append(audit.Missing, key)
				continue
			}
			observed++
			sum += seconds
			if g.Fixture == "layoutOnce" {
				values := []float64{}
				for _, o := range audit.Observations[key] {
					if o.SetupSeconds > 0 {
						values = append(values, o.SetupSeconds)
					}
				}
				setup := median(values)
				sum -= setup
				if setup > 0 {
					setups = append(setups, setup)
				}
			}
		}
		if len(setups) > 0 {
			sum += median(setups)
		}
		if observed > 0 {
			weights[g.key()] = sum
		}
	}
	sort.Strings(audit.Missing)
	return weights, audit, nil
}
