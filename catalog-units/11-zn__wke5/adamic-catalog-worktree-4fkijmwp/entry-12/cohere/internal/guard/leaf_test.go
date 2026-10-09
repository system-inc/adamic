package guard

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// allowedRuleImports are the packages of ours a rule package may depend on.
//
// `internal/rule` is the interface a rule codes against, so depending on it is the entire point.
var allowedRuleImports = map[string]bool{
	"github.com/system-inc/cohere/internal/lint/rule": true,

	// The generated Tailwind collapse table and its gate. Exact rather than a prefix, because a
	// prefix without a trailing slash would also admit `internal/tailwindanything`, and one with a
	// trailing slash matches no package at all when the package is the directory itself.
	"github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse": true,
	// The house policy both engines read as data: rule names, message text. It imports only the
	// standard library and moves when a shared decision changes, which is the event a rebuild should
	// follow. Measured on 2026-10-03 when Base's bare-throw rule began rendering its message from it,
	// three warm rounds each under the same load: editing a rule that imports it rebuilds in 1.13 to
	// 1.25s, editing one that does not in 1.21 to 1.23s, and editing policy itself in about 0.2s.
	"github.com/system-inc/cohere/policy": true,
	// How a source file's name reads to a TypeScript decision: an Adamic `.a` name as `.ts` (#6mhafvb).
	// Every rule that decides on a name's `.ts` ending asks it (#kwt1htp), and a guard of its own holds
	// them to it. It imports only `strings`, so it is a leaf itself and costs a rule edit nothing.
	"github.com/system-inc/cohere/internal/types/sourcename": true,
	// The ids, phis and single assignment the high-level IR and Adamic's flow graph share (#ejcnkja).
	// It imports only the standard library, and every rule package that names it reached it through
	// the high-level IR already, which is under the shelf's prefix, so naming it directly adds no
	// package to any rule edit's rebuild.
	"github.com/system-inc/cohere/static_single_assignment": true,
}

// allowedRuleImportPrefixes are subtrees a rule package may depend on wholesale.
//
// What makes an import expensive is not depth but movement: a rebuild is triggered by a dependency
// changing, and neither of these changes when someone edits a rule. So they cost a rule edit
// nothing, and refusing them would push legitimate dependencies into a worse shape for no measured
// gain.
//
// `internal/lint/ecmascript/` and `internal/lint/checking/` are the shared shelf, which is the
// thing rules are supposed to reach for. They also hold what vendoring leaves behind: a ported
// ECMAScript regex parser the regex rules read patterns with, and tsgolint's type-checker helper
// shelf. Neither is re-synced against anything, so both move when we change a shared decision
// rather than when an upstream releases.
//
// Measured twice, independently, when `no-invalid-regexp` first reached the parser: editing a
// rule that imports the parser and editing one that does not both rebuild in 0.30s warm, and a rule edit stays in the
// ~1.9s band it was in before the import existed. Editing the parser itself costs 3.24s, which is
// the real depth cost and is paid by whoever edits the parser rather than by rule authors.
//
// `internal/lint/rules/tailwind/collapse/` is the generated collapse table and its gate. It is regenerated when Tailwind
// releases and at no other time, which is the movement property this list is about rather than an
// exception to it. Measured on 2026-08-23, three warm rounds each: editing a tailwind rule that
// imports it rebuilds in 0.06s against 0.07s for a rule that imports nothing, which is noise. Editing
// the table itself costs 1.34s for a full binary rebuild, and that is paid by whoever regenerates it
// rather than by rule authors, which is the same shape as the regex parser above.
//
// Widened from `ecmascript/` to `utils/` on 2026-08-23, when the first lift out of a rule package
// landed and this guard refused it. `jsx/` is our own code rather than vendored, which looked like
// the distinction that mattered and is not: the rationale above is movement rather than provenance,
// and a shared helper moves when someone changes a shared decision, which is exactly the event a
// rebuild should follow. Measured on the same tree, three warm rounds:
//
//	editing a rule that imports jsx/     1.24s
//	editing a rule that imports nothing  1.20s
//	editing jsx/ itself                  0.42s
//
// A 0.04s difference is noise, and there is no cliff to protect against. Refusing the import would
// have pushed every lifted helper back into one rule package, which is the drift the retrofit guard
// in this same file exists to prevent. Two guards of ours would have been pulling against each
// other, and the measurement is what says which one was wrong.
var allowedRuleImportPrefixes = []string{
	"github.com/system-inc/cohere/internal/lint/ecmascript/",
	"github.com/system-inc/cohere/internal/lint/checking",
}

// A rule edit must not trigger a deep rebuild, and this is a build-time check rather than a review
// convention because the cost of losing it is measured and large.
//
// Measured on this tree: editing a rule package rebuilds in about 1.8s, while editing something
// deep in the import graph costs 8.52s — a 4.2x cliff on the loop a rule author sits in all day.
// The depth that causes it is invisible in review, because the import that does it looks completely
// reasonable in the file that adds it. So it is asserted here, where it fails loudly the moment it
// stops being true.
//
// The constraint is deliberately "do not depend on things that move" rather than the simpler "do
// not depend on anything." Those came apart the first time a rule package legitimately needed
// vendored upstream: the simpler rule would have rejected an import that costs nothing, which is
// how a guard stops being trusted and starts being worked around.
func TestRulePackagesStayLeaves(t *testing.T) {
	t.Parallel()

	/*
	 * A rule package is one that REGISTERS rules, not merely one that lives under `rules/`.
	 *
	 * The two were the same thing while every directory under `rules/` held rules and nothing else.
	 * They came apart when each namespace's own machinery moved in beside its rules: the Tailwind
	 * collapse engine, React's vendored conformance corpus and the score computed against it. Those
	 * are not on the loop a rule author sits in, and the guard reported all three as violations for
	 * importing exactly what they exist to import.
	 *
	 * Registration is the honest test because it is what makes a package cost a rule edit anything:
	 * `registry` imports every registering package, so a deep import inside one is on the rebuild
	 * path for all of them. A package that registers nothing is not.
	 */
	rulePackages := registeringPackages(t, listPackages(t, "./internal/lint/rules/..."))
	if len(rulePackages) == 0 {
		// A check that found nothing to check passes for the wrong reason, which is the same shape
		// as the defect it guards against.
		t.Fatal("found no rule packages, so this test proved nothing")
	}

	for _, rulePackage := range rulePackages {
		for _, imported := range listImports(t, rulePackage) {
			if !strings.HasPrefix(imported, "github.com/system-inc/cohere/") {
				// Third-party and standard-library imports are not what this guards: the compiler
				// shim is most of what a rule touches, and it is not ours to be deep in.
				continue
			}
			if allowedRuleImports[imported] || hasAllowedPrefix(imported) {
				continue
			}

			t.Errorf(
				"rule package %s imports %s, which pushes rules off the leaf — a leaf edit rebuilds in about 1.8s, a deep one in about 8.5s",
				rulePackage, imported,
			)
		}
	}
}

// hasAllowedPrefix reports whether an import sits in a subtree rules may depend on wholesale.
func hasAllowedPrefix(imported string) bool {
	for _, prefix := range allowedRuleImportPrefixes {
		if strings.HasPrefix(imported, prefix) {
			return true
		}
	}
	return false
}

// listPackages returns the import paths matching a pattern.
func listPackages(t *testing.T, pattern string) []string {
	t.Helper()
	return runGoList(t, "-f", "{{.ImportPath}}", pattern)
}

// listImports returns everything a package imports directly.
func listImports(t *testing.T, packagePath string) []string {
	t.Helper()
	return runGoList(t, "-f", `{{range .Imports}}{{.}}{{"\n"}}{{end}}`, packagePath)
}

// runGoList runs `go list` from the module root and returns its non-empty lines.
func runGoList(t *testing.T, arguments ...string) []string {
	t.Helper()

	command := exec.Command("go", append([]string{"list"}, arguments...)...)
	// Tests run in their own package directory, and the patterns here are written relative to the
	// module root.
	command.Dir = filepath.Join("..", "..")

	output, err := command.Output()
	if err != nil {
		t.Fatalf("go list %s: %v", strings.Join(arguments, " "), err)
	}

	lines := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines
}

// registeringPackages narrows a package list to those that actually register rules.
//
// Measured by the presence of a `*_register.go` file, which is the convention every rule in this
// tree follows and which `TestEveryRuleIsRegistered` independently holds to.
func registeringPackages(t *testing.T, packages []string) []string {
	t.Helper()

	const modulePrefix = "github.com/system-inc/cohere/"
	registering := make([]string, 0, len(packages))
	for _, importPath := range packages {
		directory := strings.TrimPrefix(importPath, modulePrefix)
		matches, err := filepath.Glob(filepath.Join("..", "..", directory, "*_register.go"))
		if err != nil {
			t.Fatalf("looking for registrations in %s: %v", importPath, err)
		}
		if len(matches) > 0 {
			registering = append(registering, importPath)
		}
	}
	return registering
}
