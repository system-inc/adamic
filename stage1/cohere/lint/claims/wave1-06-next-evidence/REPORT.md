Built: next-batch claim, origin ownership audit, .a proposal and three-backend JSX boundary probe; no completed rule ports.
Commits: prior work pushed through d2164d7; next claim 4d972f6f was pushed before code; evidence commit follows.
Commands and outputs: setup PASS 36s, nproc 5; upstream Go tests PASS; JSX probe PASS 16.057s; proposed registry tests PASS 0.013s.
Mutants: no per-rule semantic mutants run; identical-source extension and parser controls demonstrate the two boundaries, not rule parity.
Not covered: three implementations, corpus findings/fixes parity, per-rule mutants, native/Node/Go throughput, full gate.

# Selection

The branch remains codex/lint-wave1-06. Existing work was pushed first (Git:
Everything up-to-date), then all origin heads were fetched without recursive
submodule fetching. Main is ef3d907e. The audit examined 297 origin refs.

The ordered helper handoff is in HELPERS.md and readiness.json, linked from
helpers/REPORT.md. The first 45 entries are named in claim documents on origin
branches, including documents that record skipped existing branch ports. The
remaining entry is structure/tailwind-no-physical-direction, position 46.
The first two syntax inventory entries without a main port or claim-document
reference are @eslint-community/eslint-comments/require-description and
@next/next/google-font-display. Syntax means neither needs_type_information nor
binding_only in inventory.json from origin/codex/lint-inventory.

The selected names appear in main's frequency log, but no selected rule name
appears in its lint source modules. Logs are not implementations. selection.json
pins the inspected refs, examined entries, main source references and claim paths.
audit.py reproduces the selection using this worker's pre-claim tip for its own
origin ref; other refs use their currently fetched values. Later claims can change
its outcome, so the saved snapshot is the historical evidence.

Claim 4d972f6f was pushed before writing the proving program or patch. The original
three-rule claim remains in place. This batch did not release or overwrite another
worker's reservation.

# Concrete blockers and proposed correction

The checked-in directory registry opens rule.ts, emits rule.ts imports and rejects
.a mutant filenames. The existing five-rule set registers successfully in scratch.
Renaming only no-debugger/rule.ts to rule.a, without changing its contents, makes
registration exit 1: rule.ts: no such file or directory. This blocks every new
registered .a rule under the current ownership contract.

CLAUDE.md says: "Add a stage 1 lint rule only under
stage1/cohere/lint/rules/<slug>/" and "Never edit a dispatch, oracle, corpus or
copied-file list." A scope exception for shared .a registration and discovery was
requested asynchronously and has not been received. No shared production file was
changed. The earlier helper-merge conflict remains recorded in the first report;
this evidence uses the retained five-rule registration foundation.

The concrete a-support.patch changes registry/registry.go and lint_test.go:
select rule.a or rule.ts, reject ambiguous modules, emit the selected import,
accept owned .a mutant files, select their default module and discover/copy .a
sources for witnesses and corpora. It is an unapplied proposal, not a landed fix.
propose-a-support.py regenerates it and writes a temporary Go overlay.
Under that overlay the identical .a source registers successfully and its generated
import ends in rule.a. Existing registry tests pass. The test-harness part of the
proposal was not run against an actual new .a rule and is not claimed fully tested.
No copied .ts implementation substitutes for an authored .a rule.

Google Fonts additionally requires JsxOpeningElement and JsxSelfClosingElement.
The selected upstream positive witness is:

    export const Test = () => <link href="https://fonts.googleapis.com/css2?family=Krona+One" />;

The existing stage1 parser refuses it at byte 32: expected GreaterThanToken, got
Identifier. parser-probe.a and parser_probe_test.go run this source through the
same existing parser on source Node, emitted JavaScript and sanitized native.
All three exit 70 with identical stderr and no stdout. The ordinary TypeScript
control exits 0, prints SourceFile and has no stderr on all three backends.
The refusal is explicit; it is not zero findings and not a successful lint run.
The .a proposal does not implement JSX, entity decoding or the missing rule ports.

# Commands, logs and observations

All execution/test output was written directly to files; no test run was piped.
The .a proving program uses the existing oracle/node.mjs for source execution.
Sanitized native is built through native.Build with Sanitize: true. Compiler
source files were not edited.

Toolchain setup:

    bash cloud/setup.sh > /tmp/wave1-06-next-setup.log 2>&1
    source /workspace/adamic-tools/env.sh
    nproc

Exit 0, Go 1.27.1, clang 20.1.8, Node v24.19.0. Timing lines:

    setup: go ready (0s)
    setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
    setup: node ready (0s)
    setup: submodules ready (0s)
    setup: build cache warm (36s)
    setup: done in 36s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB

From cohere, with the environment sourced:

    go test ./internal/lint/rules/core -run '^TestRequireDescription' -count=1 -v -timeout 10m
    go test ./internal/lint/rules/tailwind ./internal/lint/rules/next -run '^(TestNoPhysicalDirection|TestGoogleFontDisplay)' -count=1 -v -timeout 10m

Output went to comments-go.txt and tailwind-next-go.txt (copied unchanged from
/tmp logs). Core PASS 0.022s, tailwind PASS 0.035s, next PASS 0.013s. These verify
only Go's own rule tests. The core command initially downloaded its regexp
prerequisites, then completed successfully without a workaround.

From the Adamic root:

    python3 stage1/cohere/lint/claims/wave1-06-next-evidence/audit.py > /tmp/wave1-06-next-selection.json
    python3 stage1/cohere/lint/claims/wave1-06-next-evidence/probe-registration.py > /tmp/wave1-06-next-registration.log 2>&1
    go test ./stage1/cohere/lint/claims/wave1-06-next-evidence -run '^TestJSXBoundary$' -count=1 -v -timeout 10m > /tmp/wave1-06-next-parser.log 2>&1
    go vet ./stage1/cohere/lint/claims/wave1-06-next-evidence > /tmp/wave1-06-next-vet.log 2>&1

All exit 0. Parser probe PASS 16.057s; vet produces an empty log. Early proving
harness attempts used the wrong lower.Lower signature, then the wrong expected
refusal text. Both were corrected; neither early attempt counts as a pass.

Proposed patch checks:

    python3 stage1/cohere/lint/claims/wave1-06-next-evidence/propose-a-support.py > /tmp/wave1-06-next-overlay-path.txt
    wave06_overlay=$(cat /tmp/wave1-06-next-overlay-path.txt)
    GOFLAGS="-overlay=$wave06_overlay" WAVE06_EXPECT_A=accepted python3 stage1/cohere/lint/claims/wave1-06-next-evidence/probe-registration.py
    go test -overlay="$wave06_overlay" ./stage1/cohere/lint/registry -count=1 -v

Output was redirected to proposed-a.txt and overlay-registry-tests.txt. Both
exit 0; final registry run PASS 0.013s. This is scratch proposal validation,
not a claim that default integration or any selected rule passes the required bar.

# Remaining work

All three ports, real findings/fixes comparisons across the required corpora and
backends, a comparison-only semantic mutant per rule and findings-per-second
measurements remain undone. No throughput of zero is reported for absent ports.
The full repository gate and filtered external compiler oracle were not run;
the new parser boundary test does independently execute source Node.
No pull request was opened. No forbidden compiler or parser file was modified.
