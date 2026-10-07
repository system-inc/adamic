Built: one .a document-import candidate and independent JSX blocker proofs for the other two claimed rules; no full-parity port is claimed.
Commits: claim 8bffa76d pushed before code; document candidate dfab0924; evidence commit is named in the final response.
Commands and outputs: witnesses, ten path fixtures, 215 compiler/stage1 files, a semantic mutant, registry controls, vet and filtered uncached oracle pass under the scratch overlay; original Go families pass.
Mutants: inverting the document exemption is caught only by Go comparison on Node, emitted JavaScript and sanitized native; the compiling skip-parsing probe control is caught on all six JSX witness/backend combinations.
Not covered: full original rule-fixture parity, the two JSX rule ports and their semantic mutants, default .a registration integration, and the full repository gate.

## Ownership and implementation

All earlier work was pushed through b4656c7c before fetching. There are 310
fetched origin refs and 32 unique claim Markdown blobs. Main remains
 ef3d907ecdc4c771b016f7d9c52372def057a340. All 46 helper-ready public names occur
in claim files. The first remaining syntax-only inventory entries, in array order
with needs_type_information=false, were:

1. @next/next/no-before-interactive-script-outside-document
2. @next/next/no-css-tags
3. @next/next/no-document-import-in-page

selection.py reproduces this choice using b4656c7c for this branch's pre-claim
snapshot. The existing helper and inventory lists are not rewritten. The claim
was pushed before the candidate or new probes were written. The missing
 docs/parallel-work.md and shared registration/profile API problems from earlier
reports remain. No compiler, parser, dispatch, copied-file list, oracle source,
or cohere submodule file is edited.

The document rule only reads static imports and filenames. Its candidate lives
in rules/next-no-document-import-in-page with rule.a, messages.a, descriptor,
independent Go adapter, raw witnesses and semantic mutant. It preserves exact
module equality, all static import clause forms, whole-declaration spans, source
trivia handling, exact policy prose, and Go's final bare-word pages split and
_document prefix. Dynamic import, require, re-export and module near misses are
silent. No fix or suggestion is offered.

The other two rules depend on JSX. Their unmodified Go rules each produce one
finding on their positive witness. Node source, emitted JavaScript and native
under ASan/UBSan instead exit 70 with empty stdout and identical parser-panic
stderr bytes across backends. These are missing AST/parser capabilities rather
than rule verdicts or sanitizer errors. No incomplete descriptors are installed.
All new Adamic programs use .a; TypeScript witnesses use .ts.txt.

## Commands and observations

Every test writes directly to a log. source /workspace/adamic-tools/env.sh was
used before toolchain commands. Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc 5.
The compatibility patch previously proposed in the boolean-outcome directory
remains unapplied. Its scratch overlay enables .a discovery, emitted JavaScript
comparison and the profile API fix. validate.py in the candidate directory
reconstructs all scratch sources byte for byte; --prepare-only was verified.

```sh
# From the repository root, after sourcing env.sh:
go test -overlay=/tmp/lint-wave1-12-batch3-overlay.json ./stage1/cohere/lint \
  -run '^TestOwnedWitnesses$' -count=1 -v -timeout=10m > /tmp/batch3-witness.log 2>&1

ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 \
  go test -overlay=/tmp/lint-wave1-12-batch3-overlay.json ./stage1/cohere/lint \
  -run '^(TestDocumentPaths|TestDocumentThroughput|TestCompilerAndStage1Agree)$' \
  -count=1 -v -timeout=20m > /tmp/batch3-parity.log 2>&1

go test -overlay=/tmp/lint-wave1-12-batch3-overlay.json ./stage1/cohere/lint \
  -run '^TestMutants$/next_document_import_path_exemption_ignored$' \
  -count=1 -v -timeout=10m > /tmp/batch3-mutant.log 2>&1

go test ./stage1/cohere/lint/claims/wave1-12-batch3-evidence \
  -count=1 -v -timeout=10m > /tmp/batch3-jsx.log 2>&1

ADAMIC_NEXT_PROBE_MUTANT=1 \
  go test ./stage1/cohere/lint/claims/wave1-12-batch3-evidence \
  -count=1 -v -timeout=10m > /tmp/batch3-control.log 2>&1

go test -overlay=/tmp/lint-wave1-12-batch3-overlay.json ./stage1/cohere/lint/registry \
  -count=1 -v > /tmp/batch3-registry.log 2>&1

go vet -overlay=/tmp/lint-wave1-12-batch3-overlay.json ./... > /tmp/batch3-vet.log 2>&1

ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m > /tmp/batch3-oracle.log 2>&1

(cd cohere && go test ./internal/lint/rules/next \
  -run '^(TestNoBeforeInteractiveScriptOutsideDocument|TestNoCssTags|TestNoDocumentImportInPage)' \
  -count=1 -v -timeout=10m > /tmp/batch3-upstream.log 2>&1)
```

Observed passes: owned witnesses 11,746 identical bytes in 26.863s; corpus/path/
throughput combined run 135.055s. It contains 215 compiler/stage1 files,
12,622,549 identical bytes, in 64.94s, and ten path fixtures, 15,020 identical
bytes, in 44.40s. The TypeScript source pin is v6.0.3,
050880ce59e30b356b686bd3144efe24f875ebc8, all 77 src/compiler files. The stage1
census is the snapshot taken when the comparison started; the later JSX probe
program is validated separately across all three Adamic backends.

Semantic mutant pass: 47.587s. The mutated .a rule inverts the document exemption.
All three backends compile and run normally with empty stderr; the resulting
missing findings are caught only by comparison against unmodified Go. This is
one rule mutant, not three. No mutant is credited for an unimplemented rule.

JSX blocker check pass: 23.197s. Its independent compiling probe control skips
parser.file(), exits 0 with wrong stdout and no stderr on all six witness/backend
combinations, and is caught by the expected-refusal check, causing the intentional
control test failure in 21.380s. This proves the blocker check can fail; it is not
credited as either JSX rule's semantic mutant.

Original pinned Go rule families pass in 0.019s. Vet exits 0 with an empty log.
Registry controls pass. Filtered uncached oracle passes in 0.512s, native/node
cache hits 0 and misses 1.

Setup workaround command:

```sh
GOFLAGS=-overlay=/tmp/lint-wave1-12-batch3-overlay.json bash cloud/setup.sh \
  > /tmp/batch3-setup.log 2>&1
```

Exit 0. Timings: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s,
build cache warm 47s, done in 47s on 5 processors; cpu.max 400000 100000,
17.6 GB. The unmodified setup's shared profile compile failure is already
recorded in the previous batch's setup.log. No shared production edit is applied.

## Full-fixture failure and remaining work

The full captured comparison was attempted without removing original fixtures:

```sh
go test -overlay=/tmp/lint-wave1-12-batch3-overlay.json ./stage1/cohere/lint \
  -run '^TestRulesAgree$' -count=1 -v -timeout=10m > /tmp/batch3-upstream-parity.log 2>&1
```

FAIL, 70.018s. It captures 281 unique source/rule/options combinations. The
shared Go oracle hardcodes ScriptKindTS and rejects case-011.ts with
Unterminated regular expression literal, because this original document-rule
fixture contains JSX. Capture also discards original filenames, so path-dependent
exemptions cannot be fully certified by this shared harness. The stage1 parser
independently refuses JSX, as the positive rule probes show. No fixture is
simplified to obtain a green full-fixture result. Full original parity is blocked
even for the import-only rule until these foundations are fixed. The document
implementation is a validated candidate for supported input, not a completed
port satisfying the entire requested bar.

## Throughput

Count-only best of five interleaved runs on 77 compiler files plus one synthetic
file with 1,000 reporting imports. All three counts are 1,000; compiler files
alone report zero. Native uses the release build. Startup, read/parse/visit are
included; build and formatted findings/fixes are excluded. Other bounded checks
overlap these runs, so these are observed rates, not isolated benchmarks.

| Backend | Best seconds | Findings per second |
|---|---:|---:|
| Native | 1.465765 | 682.24 |
| Node | 1.050867 | 951.59 |
| Go | 0.257471 | 3883.93 |

No meaningful rate is available for the two unimplemented JSX rules. A parser
failure is not lint throughput. The full repository gate is not run.
