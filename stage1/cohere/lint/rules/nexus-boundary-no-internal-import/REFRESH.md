Checked: codex/lint-wave1-12 still contains current origin/main e8ba3d5d81de4d3773c723914fccd4c76248b965; no rebase or new claim was needed.
Commits: the rule branch was already pushed at ae1934595604a404d0f42570d6eadb0e316a8816; this commit adds fresh evidence only.
Checks: JSX decisions PASS 20.470s; Tailwind decisions PASS 51.784s; import paths/parser-gap package PASS 72.436s; default harness and registry fail.
Mutants: five JSX decisions, three Tailwind decisions and one import-path semantic change compile and finish, caught only by actual-Go comparison on Node source, emitted JavaScript and sanitized native.
Uncovered: default integration is not green; profile-test compilation and eight incomplete rule directories prevent landing readiness. No new helper is claimed.

After fetching every origin head, current main and both owned remote heads were unchanged. Both owned branches already contain current main. The published .a harness branch is still f4d98cab50048692781da3599131317dc569d466. This branch has no compiler delta against main. Prior landing evidence remains historical and tied to the same base, with the explicit scratch accommodations documented in LANDING.md.

Fresh commands, with /workspace/adamic-tools/env.sh sourced:

```sh
go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout=10m > /tmp/wave12-refresh-rules-default.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/rules/next-google-font-preconnect ./stage1/cohere/lint/rules/better-tailwindcss-enforce-shorthand-classes ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import ./stage1/cohere/lint/registry -count=1 -v -timeout=15m > /tmp/wave12-refresh-rules-owned.log 2>&1
```

Both commands exit 1. The default package fails before running tests at profile_test.go:32:23: cannot range over portFiles (value of type func(t *testing.T) []string): func must be func(yield func(...) bool): unexpected results. The three independent rule packages pass; the registry fails in 0.243s because better-tailwindcss-enforce-shorthand-classes is missing a rule descriptor. Eight partial JSX/Tailwind decision directories are listed in LANDING.md. They still lack complete extraction or registration and are not represented as final ports. The registry failure also masks its descriptor-rejection mutant checks, so those checks are not credited as successful.

Shared-harness/compiler/registration changes are outside the authorized rule directories. No placeholder descriptors or new shared edits conceal these failures. Existing owned kernels and path comparisons cover the available behavior; fresh independent Go parser-gap probes confirm the remaining extraction boundaries. Work stops at the exact blockers under the landing cap. No new claims, main pushes or area/ pushes were made.

Setup succeeds in 149s: Go, clang, Node and submodules ready at 0s, cache warm at 149s, done at 149s; nproc 5, four-core cgroup quota, 17.6 GB. refresh-*.log retains complete outputs. No complete repository gate, new TypeScript/stage1 corpus replay or fresh throughput samples were run. Earlier bounded scratch comparisons and historical findings/second tables retain their original scope; they do not establish green default integration.
