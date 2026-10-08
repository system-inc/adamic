Claimed the largest unheld independent format package: non-JSON ESTree, 3,674 Go lines.
Built its indexed model, converter, postprocessing and source-to-tree driver; the port is incomplete.
14,699 repository files and 86 generated cases match Go on source Node, sanitized native and emitted JS.
Full-output throughput medians: native 5,709, Node 4,601, Go 11,345 texts/s; three port mutants caught.
Whole-checkout parity remains blocked: 92 wrong outputs, 984 acceptance disagreements and proved input/parser gaps.

Branch `codex/stage1-estree`, from main `5d4c8012a0877094134e6c6bac367ff68f9313e8`.
Cohere pin `715ba94f3608a6500086b1076ce5cb7e51b836db`. Dashboard read whole at
`origin/codex/stage1-progress` commit `7ed6d30c3f2d407a183c971ec2ed2577c0e5f615`.
The origin survey, package sizes and boundaries are in this directory's `CLAIM.md`.
All 142 visible origin refs were fetched. Existing formatter packages with active
branches were treated as held, including unfinished composition work.

Commits pushed at green checkpoints:

| Commit | Result |
| --- | --- |
| `ffa9ea6` | Claim pushed before implementation. |
| `c77fb8d` | Model, visitor keys, locations, comments and postprocess; 47 cases, 46,768 identical bytes; three mutants caught. |
| `fab8c3e` | Functions, declarations, members, generics, imports/exports and modules; 75 cases, 112,528 bytes; pinned original libraries and exact gaps. |
| `2619a35` | Import-type options, tagged escapes, JavaScript boundary, scalar helpers and cached offsets; 78 cases, 120,355 bytes; full source corpus audit and proving gaps. |

The final audit/throughput commit contains this report and retained logs. Its SHA
is reported in the final response. No pull request was opened.

The driver prints canonical ESTree, not formatted JavaScript. It compares ordered
own fields, ranges, parenthesis flags, ContentEnd, location helper results,
ignored-node semicolon behavior, comments and stripped source. Go calls the
unmodified public cohere Parse API through a build overlay. Conversion is built
on the existing stage 1 TypeScript parser and scanner. No parser, compiler or
runtime files were changed, including the four files reserved for other workers.

Corpus inventory: 27,369 JS/TS/Adamic files, 63,031,352 source bytes, nine files with
malformed UTF-8. The inventory starts with every matching extension in the
checkout and its submodules; the final unit files were added separately. Inputs
were frozen in scratch with SHA-256 recorded before Go and Node comparisons.
The complete inventory and every disposition are retained as compressed JSONL.

| Final source Node versus Go boundary | Files |
| --- | ---: |
| Identical | 14,699 |
| Port refuses a Go answer | 861 |
| Port answers when the Go boundary refuses | 123 |
| Both refuse | 11,594 |
| Different output | 92 |
| Total | 27,369 |

Go refusal includes Parse API errors and caught panics at canonical serialization
or Go location helpers; it does not mean every such source is language-invalid.
This table records failures rather than redefining them as supported coverage.
The source audit uses a catchable panic shim and a Node-only repeated-scan guard
in the imported parser: after more than 32 identical scanner positions it throws.
That guard classified 13 stalls. It does not alter the actual driver or native
build. The initial unguarded audit exhausted Node's heap after 13,900 files.

Every one of the 14,699 matching files was then checked using the uninstrumented
source driver, ASan/UBSan native and emitted JavaScript, against stored Go output.
504,506,334 output bytes matched in each implementation. The main 14,697-file
check passed in 452.694s; the final two matching files passed separately in
17.480s. The unsupported and mismatching files were not tested as native successes.

Generated coverage: 78 general files produce 120,355 bytes; eight scalar edge files
produce 5,969 bytes. All 86 match Go in all four implementations. Pinned independent
libraries: typescript-estree 8.65.0, TypeScript 6.0.3 and Prettier 3.9.6, installed
in `/workspace/scratch/estree-library`. The raw original library matches 78 general
cases and five finite scalar/bigint cases. The postprocessed original wrapper
matches 75 general cases plus those five, with exactly three held discrepancies:
NBSP/U+2028 ContentEnd and CRLF normalization. Two numeric range discrepancies are
proved independently: Go produces NaN/NaN for `1e999; 0x10000000000000000;`, while
the original library produces Infinity/18446744073709552000.

Three qualifying port mutants, each compiled and finished normally on source
Node and sanitized native; byte comparison caught each on both:

| Mutant | Caught difference in the final 78-case corpus |
| --- | --- |
| Reverse computed member access | Line 3810: `.computed bool 0` becomes `1`. |
| Remove logical rebalancing | Line 3543: nested LogicalExpression becomes Identifier. |
| Change merged JSDoc separator | Line 4661: `*//*` becomes `*/ /*` in the comment value. |

Compiler refusals, sanitizer failures and nonzero exits were not counted as
qualifying port mutants. The separate interface-default gap is a compiler bug:
Node prints 5; main now refuses during lowering at interfaceDefault.ts:10:12
with typed NotYet for a class method through a view that erases its prototype origin.
The interfaceTypeMethod.ts proof likewise prints 1 on Node and now has the same
typed refusal at 10:12; its workaround now uses a renamed default-free concrete
method and an explicit callback property, and remains green on all builds.
Both native failure paths are blocked, but origin-preserving dispatch remains
open for @system_adamic. The exact assertions and unchanged proving programs
are retained in this package. Method replacement is separately
refused at lowering as unbound-method. Postfix increment used as a value is a
held stage 0 NotYet. Each has a successful independent Node answer and an exact
recorded refusal expectation. The input API and recovery gaps are described
with minimal programs in `GAPS.md`.

Throughput uses the same 78-case manifest repeated 20 times: 1,560 texts,
2,407,100 byte-identical output bytes per run. Release native, source Node and
Go each run once to warm up, then five runs in rotated order. Every output is
compared. Timing includes process startup, input reads, complete canonical
serialization, file-backed stdout and the verification read. This is a tree-driver
measurement, not parser-only or whole-JavaScript-formatter throughput. Benchmarking
ran after corpus validation finished, without concurrent test loads.

| Driver | Five samples, texts/s | Median texts/s |
| --- | --- | ---: |
| Go | 11,345; 11,222; 11,088; 11,566; 11,618 | 11,345 |
| Source Node | 4,578; 4,601; 4,638; 4,521; 4,872 | 4,601 |
| Native | 5,004; 5,650; 5,736; 5,709; 5,873 | 5,709 |

Commands and observed results are retained in `validation/`:

```sh
bash cloud/setup.sh > /tmp/stage1-estree-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
# 5
# setup: go ready (1s)
# setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
# setup: node ready (1s)
# setup: submodules ready (1s)
# setup: build cache warm (72s)
# setup: done in 72s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB

ADAMIC_ESTREE_LIBRARY=/workspace/scratch/estree-library \
  go test -count=1 -v ./stage1/cohere/estree > /tmp/stage1-estree-step3-final.log 2>&1
# PASS, 101.899s. Generated 78-case gate, three successful mutants, library and gap checks.

go test -count=1 -v ./stage1/cohere/estree -run '^TestInterfaceDefaultGap$' \
  > /tmp/stage1-estree-interface-default-test.log 2>&1
# PASS, 0.357s. Expected sanitizer failure is a proved gap, not a port mutant.

go test -count=1 -v ./stage1/cohere/estree -run '^TestScalarEdges$' \
  > /tmp/stage1-estree-scalar-edges.log 2>&1
# PASS, 16.904s. Eight files and 5,969 byte-identical output bytes.

ADAMIC_ESTREE_LIBRARY=/workspace/scratch/estree-library \
  go test -count=1 -v ./stage1/cohere/estree \
  -run '^(TestScalarOriginalLibraries|TestPinnedNumericGaps|TestOriginalLibraries)$' \
  > /tmp/stage1-estree-final-library.log 2>&1
# PASS, 3.097s. Five scalar library agreements and the two exact range gaps.

ADAMIC_ESTREE_CORPUS=/workspace/scratch/estree-corpus-final \
  go test -count=1 -v ./stage1/cohere/estree -run '^TestRepositoryAgreement$' \
  > /tmp/stage1-estree-repository-agreement.log 2>&1
# PASS, 452.694s. 14,697 files; 504,498,084 identical bytes per implementation.

ADAMIC_ESTREE_CORPUS=/workspace/scratch/estree-corpus-final/extra \
  go test -count=1 -v ./stage1/cohere/estree -run '^TestRepositoryAgreement$' \
  > /tmp/stage1-estree-repository-extra-agreement.log 2>&1
# PASS, 17.480s. Two files; 8,250 identical bytes per implementation.

ADAMIC_ESTREE_BENCHMARK=1 go test -count=1 -v ./stage1/cohere/estree \
  -run '^TestThroughput$' > /tmp/stage1-estree-throughput-final.log 2>&1
# PASS, 10.015s. Samples and methodology above, with every output compared.

go test -count=1 -v ./internal/oracle -run '^TestTheOracleCatchesOneByte$' \
  > /tmp/stage1-estree-oracle-filter.log 2>&1
# PASS, 2.987s. Cache native/node misses; the wrong byte is caught.

go vet ./stage1/cohere/estree > /tmp/stage1-estree-vet.log 2>&1
# Exit 0, empty output.

/workspace/scratch/cohere --no-cache stage1/cohere/estree/*.ts \
  > /tmp/stage1-estree-step3-cohere.log 2>&1
# Exit 0; 16 checked, 100 percent Adamic-ready, 0.6s.
```

Corpus reproduction: build `testdata/oracle.go` through the overlay used by
`goOracle` in `estree_test.go`; run its `--audit <manifest> <answer-directory>`
mode, redirecting stdout to Go JSONL and stderr to a log. Run
`node --disable-warning=ExperimentalWarning testdata/corpus.mjs <go-jsonl>
<port-jsonl> <identical-manifest>` with output redirected to a log. The saved
inventory supplies relative paths, sizes and hashes. Scratch snapshots and stored
Go answer paths in the audit records are local artifacts; reconstruct them from
that inventory to reproduce elsewhere. `TestRepositoryAgreement` batches the
identical records into groups of 100 and checks every byte, without instrumentation.

Not covered or completed: complete checkout parity, diagnostic equivalence, JSX,
raw invalid-UTF-8 inputs, unpaired cooked surrogate representation, the remaining
92 converter/parser disagreements, whole JavaScript formatting or composition
with existing printer branches. The unchanged parser can stall on invalid input;
the driver is experimental and needs external time bounds for arbitrary input.
Literal U+FFFD and all TSX/JSX extensions are conservatively refused, including
some valid sources. Not every remaining cost or gap belongs to the runtime.
No full repository integration gate was run; the touched package, complete
matching-subset corpus, focused uncached oracle and vet were run as listed.

The final combined package gate was rerun after adding the scalar and compiler-gap
tests and factoring the independent library check:

```sh
ADAMIC_ESTREE_LIBRARY=/workspace/scratch/estree-library \
  go test -count=1 -v ./stage1/cohere/estree > /tmp/stage1-estree-final-gate.log 2>&1
# PASS, 124.508s. Generated agreement, all three qualifying mutants, original
# libraries, raw-input/recovery/refusal/default-argument gaps and scalar edges.
# The opt-in corpus and throughput tests skip here; both passed separately above.
go vet ./stage1/cohere/estree > /tmp/stage1-estree-final-vet.log 2>&1
# Exit 0, empty output.
```

Full combined output: `validation/final-gate.log`. All test output was redirected
to files, never piped. Pushes at each listed green checkpoint returned exit 0.

## Merge-seat repair: stage1-format/estree-2

Seat base cbcc1575, merged with main 71d7e491. The method-origin proof follows
ts-printer commit b07a0643: both interface programs now assert errors.As to
*lower.NotYet, exact file:10:12 and What
"a class method through a view that erases its prototype origin".
Node still prints 5 and 1 respectively. The former default-free type-method
control is also refused on main; its replacement retains both arguments, removes
defaults, renames the concrete method typeValue and exposes an explicit callback
property. Node, sanitized native and emitted JS all print 1. The original gap
programs remain unchanged. No other recorded gap moved or closed in this run.
Origin-preserving interface dispatch remains open for @system_adamic.

Pinned full-package command (Linux, output /tmp/estree-2-full.log):

    env -u ADAMIC_ESTREE_CORPUS -u ADAMIC_ESTREE_BENCHMARK \
      ADAMIC_ESTREE_LIBRARY=/tmp/estree-2-library \
      go test -v -count=1 -timeout 60m ./stage1/cohere/estree

The npm prefix contains @typescript-eslint/typescript-estree@8.65.0,
typescript@6.0.3 and prettier@3.9.6. No package test requires
ADAMIC_TYPESCRIPT_SOURCE. Result: exit 0, PASS, 777.287s. All Go byte comparisons,
pinned-library checks and mutant checks passed without relaxed checks. The only
skips are TestRepositoryAgreement and TestCorpusNativeRefusals (corpus opt-in),
and TestThroughput (benchmark opt-in), left unset as requested.
Focused interface proofs also passed. go vet ./stage1/cohere/estree, gofmt and
git diff --check are clean. Leak detection is unchanged; only this package is
edited beyond the requested merge from main. No full-repository gate is claimed.
