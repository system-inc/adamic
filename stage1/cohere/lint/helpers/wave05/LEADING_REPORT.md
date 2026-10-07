Built collapse.leadingInteger in one .a helper, preserving the pinned Go host's 64-bit wrapping arithmetic.
Claim 652db0c7 was pushed before code; implementation commit contains this report.
Go consumer tests and TestWave05LeadingInteger pass; 2,589 cases yield 46,136 identical bytes on source Node, emitted JavaScript and sanitized native.
Mutant ignores a minus sign; it compiles, executes with exit zero and empty stderr, and all three comparisons catch it.
Uncovered: 32-bit Go hosts, arbitrary invalid UTF-8 Go strings, full consumer findings parity and shared integration.

Six dependency entries removed, zero final blockers alone. Consumers are enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes and no-unknown-classes, all under better-tailwindcss/. The frozen readiness list is unchanged. leading_consumers.json records these consumers, 145 captured fixtures, four actual calls with two distinct arguments, and the supplement including each consumer source/token plus whitespace, signs, Unicode and randomized long decimal/overflow controls. Go remains the only source of expected results.

The Go rule bodies and helper arithmetic are unchanged. Scratch overlays rename the original private helper and wrap it only to observe inputs; a bridge calls that original. Tests' hard-coded macOS package-search paths are redirected to /tmp/w05-helper-tailwind, containing installed tailwindcss 4.3.3. Upstream fixtures genuinely run; skips are not counted as coverage.

Integer exposes exact high/low unsigned words, valid and decimal(), rather than rounding a Go int through JavaScript's number representation. Callers must retain this representation when values exceed safe integer bounds; there is no numeric approximation. All intermediates remain below 2^53. Go's actual parser accepts only ASCII space, tab, LF and CR before an optional sign, scans ASCII digits and wraps int multiplication/addition/negation.

Reproduce from the repository root: source /workspace/adamic-tools/env.sh; npm install --prefix /tmp/w05-helper-tailwind --no-audit --no-fund --ignore-scripts tailwindcss@4.3.3 > /tmp/w05-helper-npm.log 2>&1; go test ./stage1/cohere/lint/helpers -run '^TestWave05LeadingInteger$' -count=1 -v > /tmp/w05-helper-owned-test-final.log 2>&1. The Python validator regenerates deterministic corpus and hashes. Logs are retained under wave05/evidence; leading_evidence.json records sizes, hashes and empty stderr checks.

Setup Go/clang/Node/submodule readiness: 0s each. Build-cache warm 80s, done in 80s; nproc 5, cgroup permits four CPUs. Go 1.27.1, clang 20.1.8, Node 24.19.0. No compiler or shared harness files edited; no full gate claimed.
