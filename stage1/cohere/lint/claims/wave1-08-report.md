Built: pushed rule claim and merged foundations; rule implementation is blocked by shared contracts.
Commits: registration merge 9c8510f; helpers merge 615fe31; claim 346d122.
Commands and outputs: registry tests PASS 0.049s; helper gap test PASS 12.221s; lint baseline build FAIL.
Mutants: no rule semantic mutants run; .a filename probe rejected with missing rule.ts, exit 1.
Not covered: the three ports, cross-backend findings/fixes parity, corpus runs, throughput and full gate.

# Slot 08 blocked handoff

The claimed rules are no-useless-computed-key, react-hooks/gating and
react/forbid-foreign-prop-types. Positions come from HELPERS.md's measured handoff;
helpers/REPORT.md links there rather than containing an ordered list itself.
All origin branches were fetched and searched. No implementation was found for
these rules. This is not a skip for an existing port.

Base: origin/main d090af531216ddd3c25a0dede6b82d7c0a6edf76.
Registration: 48ecd9302bf3954a4ddbbd14c28ba09148c1c1a8.
Helpers: 5d13f5baaecaf11d4ea62de693426f69a1f41bba.
Claim was pushed before implementation; no new Adamic source was written.

## Observed blockers

1. The foundations do not merge cleanly. Six conflicts occurred in lint README,
   lint.ts, lint_test.go, main.ts, settings.ts and testdata/oracle.go. Registration
   versions were retained and helper additions retained. This preserves the
   registration branch's five rules, not the helpers branch's historical twenty
   inline rules. Legacy added files remain; this merge is not a validated
   integration of those additional rules.
2. Registry discovery requires rule.ts (registry/registry.go:95), rendering
   imports rule.ts (:187), mutant files must end in .ts (:151), and the test copy
   walks select .ts. A scratch copy renaming the existing debugger implementation
   to rule.a produces:
   open /tmp/lint-wave1-08-extension/rules/no-debugger/rule.ts: no such file or directory
   exit status 1
   This is an infrastructure rejection, not a comparison-killed semantic mutant.
3. The helpers branch adds profile_test.go which ranges over portFiles as a slice;
   registration changes it to func(t *testing.T) []string. The resolved-tree
   TestRulesAgree command cannot compile:
   profile_test.go:32:23: cannot range over portFiles
   No finding comparison ran.
4. testdata/oracle.go:61 rejects a fix range unequal to the finding range.
   no-useless-computed-key reports the member and fixes only the brackets.
   This incompatibility is observed in the two source contracts; a runtime
   reproduction of the fix-shape panic was not run.
5. docs/parallel-work.md is absent from main and both foundation branches.

CLAUDE.md says "Add a stage 1 lint rule only under
stage1/cohere/lint/rules/<slug>/" and "Never edit a dispatch, oracle, corpus or
copied-file list." The user requires .a source. Implementing registered .a rules
and byte-exact bracket fixes therefore needs an expanded shared-foundation scope.
A scope question was sent; no shared registry or harness fixes have been made.

## Reproducible bounded checks

Commands were run from the repository root after sourcing
/workspace/adamic-tools/env.sh. All test output went directly to log files.

- go test ./stage1/cohere/lint/registry -count=1 -v
- go test ./stage1/cohere/lint/helpers -run '^TestKnownGapsAreExplicit$' -count=1 -v
- go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -v -timeout 10m
- go run ./cmd/lint-registry -root /tmp/lint-wave1-08-extension

Evidence logs are in wave1-08-evidence. Registry tests reject their existing
malformed-descriptor mutations, including duplicate adapter/name, unknown field,
missing exports/hooks and invalid kind. Those are not semantic mutants for any
assigned rule. No native, Node or Go findings-per-second figures are claimed.

## Setup

nproc: 5. Go 1.27.1, clang 20.1.8, Node 24.19.0.
The first bash cloud/setup.sh run was started during the unresolved merge and
failed in its warm-cache step on a conflict marker in lint_test.go:28. Tool
installation succeeded. A resolved-tree rerun was started to isolate the actual
integration failure; its final output is recorded alongside this report.

Resolved-tree setup rerun exited 1 in the build-cache warm step on
profile_test.go:32:23 (portFiles function ranged as a slice). Its timing lines:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
```

There is no successful build-cache warm or done timing line. Tool installation
needed no workaround: source /workspace/adamic-tools/env.sh makes the installed
tools available. The unaffected registry and helper checks ran with that setup.

Upstream Go checks also passed. From cohere/:

```
go test ./internal/lint/rules/core ./internal/lint/rules/react -run '^(TestNoUselessComputedKey|TestGating|TestForbidForeignPropTypes)' -count=1 -timeout 5m
ok github.com/system-inc/cohere/internal/lint/rules/core 0.017s
ok github.com/system-inc/cohere/internal/lint/rules/react 0.015s
```

This checks the existing Go rules, not the absent Adamic ports. The blocked
handoff and evidence were committed as 959d1ae and pushed after claim 346d122.
