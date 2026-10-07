Built: pushed third-batch claim and reproducible selection/blocker evidence; no completed ports.
Commits: prior work b63449cf pushed first; claim 6732d3c2 pushed before evidence; evidence commit follows.
Commands and outputs: setup 30s, nproc 5; Go reference tests PASS 0.016s; three-backend parser probe PASS 16.432s.
Mutants: no rule mutants run; extension rename is a registration boundary control, not a semantic mutant.
Not covered: implementations, corpus findings/fixes parity, per-rule mutants, throughput and full gate.

# Selection

Fetched every origin head after pushing all prior work. Audited 310 refs.
All helper-ready entries appear in origin claim documents. The next available
syntax inventory entries, in rules-array order, are:

1. @next/next/no-before-interactive-script-outside-document
2. @next/next/no-css-tags
3. @next/next/no-document-import-in-page

No selected name appears in main lint source modules or in another origin claim.
The audit excludes nested evidence and considers syntax entries without
needs_type_information or binding_only. audit.py substitutes the pre-claim
b63449cf tip for this worker's own branch to reproduce the selection.
selection.json records inspected ref tips and matching ownership references.
Claim 6732d3c2 was pushed before writing new evidence. Claims remain reserved;
this report does not assert they are implementations.

# Observed blockers

All three upstream rule sources were read in full. NoCssTags and
NoBeforeInteractiveScriptOutsideDocument listen for JSX opening/self-closing
nodes. The existing parser probe reconfirms identical explicit JSX refusal
(exit 70, no stdout) on source Node, emitted JavaScript and sanitized native;
its ordinary TypeScript control succeeds on all three. This probe uses the
previous batch's font witness, not either newly selected rule's own corpus.
It establishes a parser boundary and does not establish rule equivalence.
NoDocumentImportInPage is import-only and does not have this JSX requirement.

The registry control succeeds for the five existing .ts rules. Changing only
no-debugger/rule.ts's extension to .a fails registration with
rule.ts: no such file or directory. This blocks registering all three new .a
ports on the retained registration foundation.

The concrete unapplied correction is ../wave1-06-next-evidence/a-support.patch.
CLAUDE.md says "Never edit a dispatch, oracle, corpus or copied-file list."
The requested shared-file scope exception has not been received, so no shared
production file or compiler core was changed. The original helper merge
conflicts remain documented in ../wave1-06-evidence/REPORT.md.

# Commands and logs

Every test wrote directly to a log. Tool environment was sourced from
/workspace/adamic-tools/env.sh. Setup succeeded with these timing lines:

    setup: go ready (1s)
    setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
    setup: node ready (1s)
    setup: submodules ready (1s)
    setup: build cache warm (30s)
    setup: done in 30s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB

From cohere:

    go test ./internal/lint/rules/next -run '^(TestNoBeforeInteractiveScriptOutsideDocument|TestNoCssTags|TestNoDocumentImportInPage)' -count=1 -v -timeout 10m > /tmp/wave1-06-third-go.log 2>&1

Exit 0, PASS 0.016s; go.txt contains all reference case results.
This validates Go only, not an Adamic port.

From Adamic root:

    bash cloud/setup.sh > /tmp/wave1-06-third-setup.log 2>&1
    nproc > /tmp/wave1-06-third-nproc.log
    python3 stage1/cohere/lint/claims/wave1-06-third-evidence/audit.py > /tmp/wave1-06-third-selection.json
    python3 stage1/cohere/lint/claims/wave1-06-next-evidence/probe-registration.py > /tmp/wave1-06-third-registration.log 2>&1
    go test ./stage1/cohere/lint/claims/wave1-06-next-evidence -run '^TestJSXBoundary$' -count=1 -v -timeout 10m > /tmp/wave1-06-third-parser.log 2>&1

All exit 0. Parser probe PASS 16.432s. These reused boundary checks require no
new Adamic source; no .ts implementation was authored. Native uses ASan/UBSan.
No throughput is reported for absent implementations. No full gate or PR.
