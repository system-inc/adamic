package main

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"
)

type complement struct {
	Package, Parent string
	Shard           int
}
type selection struct{ Run, Skip string }

func complements(p plan) []complement {
	seen := map[string]bool{}
	var out []complement
	for _, u := range p.Units {
		parent, _, split := strings.Cut(u.Test, "/")
		if !split {
			continue
		}
		key := u.Package + "::" + parent
		if seen[key] {
			continue
		}
		seen[key] = true
		owner := p.Count - 1
		if (p.WASI != nil || p.Environment != nil) && !u.WASI && len(u.RequiredEnvironment) == 0 && p.Count > 1 {
			owner--
		}
		for _, group := range p.Affinity {
			if group.Package != u.Package {
				continue
			}
			for _, test := range group.Tests {
				if test == u.Test {
					owner = group.Shard
				}
			}
		}
		out = append(out, complement{u.Package, parent, owner})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Package == out[j].Package {
			return out[i].Parent < out[j].Parent
		}
		return out[i].Package < out[j].Package
	})
	return out
}
func assignedPackages(p plan, index int) map[string][]unit {
	out := map[string][]unit{}
	for _, u := range p.Units {
		if u.Shard == index {
			out[u.Package] = append(out[u.Package], u)
		}
	}
	for _, c := range p.Complements {
		if c.Shard == index {
			if _, ok := out[c.Package]; !ok {
				out[c.Package] = nil
			}
		}
	}
	return out
}

// Go testing splits unparenthesized top-level | into alternate complete paths.
// Each component remains anchored; grouping the whole union would prevent slash splitting.
func exactPath(name string) string {
	if name == "" {
		return "^$"
	}
	parts := strings.Split(name, "/")
	for i := range parts {
		parts[i] = "^" + regexp.QuoteMeta(parts[i]) + "$"
	}
	return strings.Join(parts, "/")
}
func selections(p plan, index int, pkg string) []selection {
	var run, skip []string
	parents := map[string]bool{}
	sdk := false
	for _, c := range p.Complements {
		if c.Shard == index && c.Package == pkg {
			parents[c.Parent] = true
			run = append(run, exactPath(c.Parent))
		}
	}
	for _, u := range p.Units {
		if u.Package != pkg {
			continue
		}
		parent, _, _ := strings.Cut(u.Test, "/")
		if parents[parent] {
			if u.Shard != index {
				skip = append(skip, exactPath(u.Test))
			}
			continue
		}
		if u.Shard == index {
			if strings.HasSuffix(pkg, "/internal/native") && u.Test == "TestWASI" {
				sdk = true
			} else {
				run = append(run, affinityRunPath(p, u))
			}
		}
	}
	sort.Strings(run)
	sort.Strings(skip)
	var out []selection
	if len(run) > 0 {
		out = append(out, selection{strings.Join(run, "|"), strings.Join(skip, "|")})
	}
	if sdk {
		out = append(out, selection{Run: "^TestWASI$"})
	}
	return out
}
func selectorKeys(p plan, index int, pkg string) []string {
	var out []string
	for _, s := range selections(p, index, pkg) {
		b, _ := json.Marshal(s)
		out = append(out, string(b))
	}
	return out
}
func selectionArgs(pkg string, s selection) []string {
	args := []string{"test", "-count=1", "-json", "-timeout", "60m", "-run", s.Run}
	if s.Skip != "" {
		args = append(args, "-skip", s.Skip)
	}
	return append(args, pkg)
}
func complementOwns(p plan, index int, r result) bool {
	for _, c := range p.Complements {
		if c.Shard == index && c.Package == r.Package && strings.HasPrefix(r.Test, c.Parent+"/") {
			return true
		}
	}
	return false
}

func schedulingIdentity(identity string, jobs int, parallel ...int) string {
	return digestJSON([]any{identity, jobs, parallel})
}
