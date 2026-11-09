package lower

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
)

// This form-scoped row describes the binding's members, not a declaration-wide
// exemption for any. The typeof lines are forms but introduce no extra member.
func TestHostAnyModelledPerformanceForms(t *testing.T) {
	t.Parallel()
	var performance map[string]string
	rows := 0
	for _, row := range hostAnyModelledRows(t) {
		if strings.HasPrefix(row["Declaration"], "`performance`") {
			rows++
			performance = row
		}
	}
	if rows != 1 {
		t.Fatalf("want exactly one Modelled performance row, got %d", rows)
	}
	forms, scoped := performance["Forms"]
	if !scoped {
		t.Fatal("performance's Modelled row must have a Forms column")
	}
	memberForm := regexp.MustCompile(`\bperformance\.([A-Za-z_$][A-Za-z0-9_$]*)`)
	documented := map[string]bool{}
	typeofBinding, typeofMembers := false, false
	for _, form := range strings.Split(forms, "<br>") {
		form = strings.TrimSpace(form)
		switch form {
		case "`typeof performance`":
			typeofBinding = true
		case "`typeof` of each of the six":
			typeofMembers = true
		}
		for _, match := range memberForm.FindAllStringSubmatch(form, -1) {
			documented[match[1]] = true
		}
	}
	if !typeofBinding || !typeofMembers {
		t.Fatal("performance's Forms must retain typeof of the binding and of each of the six members")
	}
	registered := []string{}
	const prefix = "node:perf_hooks.Performance."
	for name, implemented := range implementedNodeMembers {
		if implemented && strings.HasPrefix(name, prefix) {
			registered = append(registered, strings.TrimPrefix(name, prefix))
		}
	}
	named := []string{}
	for name := range documented {
		named = append(named, name)
	}
	slices.Sort(named)
	slices.Sort(registered)
	if len(named) == 0 || !slices.Equal(named, registered) {
		t.Fatalf("performance Forms do not match registered members: documented=%v registered=%v", named, registered)
	}
	if !implementedNodeMembers["node:perf_hooks.performance"] {
		t.Fatal("node:perf_hooks.performance must be registered")
	}
	for location := range hostAnyModelled(t) {
		if strings.Contains(location, "perf_hooks.d.ts:") {
			t.Fatalf("form-scoped performance must not allow-list %s", location)
		}
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	paths := regexp.MustCompile(`(?:[A-Za-z0-9_.-]+/)+[A-Za-z0-9_.-]+`).FindAllString(performance["Fixtures and divergences"], -1)
	if len(paths) == 0 {
		t.Fatal("performance's Modelled row names no repository evidence paths")
	}
	for _, path := range paths {
		if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil || info.IsDir() {
			t.Fatalf("performance's Modelled evidence %s does not name a repository file: %v", path, err)
		}
	}
}

func hostAnyPerformanceToJSONMutant(t *testing.T) string {
	t.Helper()
	return buildcache.GoTestMutated(t, "", "host-any-performance-tojson.test", "./internal/lower", nil, []buildcache.Mutation{{
		File:   "internal/lower/library_node_process.go",
		Before: `[]string{"timeOrigin", "now", "mark", "measure", "clearMarks", "clearMeasures"}`,
		After:  `[]string{"timeOrigin", "now", "mark", "measure", "clearMarks", "clearMeasures", "toJSON"}`,
	}})
}

func TestProduct_HostAnyPerformanceToJSONMutant(t *testing.T) {
	t.Parallel()
	hostAnyPerformanceToJSONMutant(t)
}

func TestHostAnyPerformanceToJSONMutantIsCaught(t *testing.T) {
	t.Parallel()
	binary := hostAnyPerformanceToJSONMutant(t)
	for _, pin := range []struct{ name, line, location string }{
		{"TestHostAnyModelledPerformanceForms", "performance Forms do not match registered members: documented=[clearMarks clearMeasures mark measure now timeOrigin] registered=[clearMarks clearMeasures mark measure now timeOrigin toJSON]", ""},
		{"TestHostAnyAudit", "node:perf_hooks.Performance.toJSON: value-carrying any at node:perf_hooks.Performance.toJSON.signature[0:0].return", "perf_hooks.d.ts:116:"},
	} {
		t.Run(pin.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, binary, "-test.run", "^"+pin.name+"$", "-test.count=1", "-test.v")
			stdout, stderr, err := childguard.Capture(ctx, command)
			var exited *exec.ExitError
			if !errors.As(err, &exited) || exited.ExitCode() != 1 || !strings.Contains(string(stdout), pin.line) || !strings.Contains(string(stdout), pin.location) {
				t.Fatalf("performance toJSON mutant: err=%v stdout=%s stderr=%s", err, stdout, stderr)
			}
			for _, line := range strings.Split(string(stdout), "\n") {
				if strings.Contains(line, pin.line) {
					t.Logf("%s goes red: %s", pin.name, strings.TrimSpace(line))
					break
				}
			}
		})
	}
}
