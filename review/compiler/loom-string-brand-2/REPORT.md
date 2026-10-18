Re-cut the structural phantom-string proof beyond __String without further language changes, toward Outcome 11.
Base 87ad5a4fe0b2ddf75adcde9d326e9d67edc1f2e4; parked source 8fef5fbadf7ba0fbf129226ac8771888da5f3370; one squashed commit on compiler/loom-string-brand-2.
Node fixtures, five mutants, +6/-0 admission delta, six new count rows only, reader/SortedMap guards and vet passed.
Mutants: real-field and read-brand refusals, undefined Node mismatch, literal refinement rejection, missing Defined IR check all caught.
Not covered: full historical 44-site closure rerun, full gate, traced Workshop runner-cache shipping; no Loom submission.

## Re-cut and conflict resolution

Started with git fetch origin and git checkout -b compiler/loom-string-brand-2 origin/main, then git merge --squash origin/compiler/loom-string-brand. counts.md was the only conflict. Its main version was restored with git checkout --ours and checked byte-for-byte against main-counts.md before counting. The automatic expression.go merge preserved main's changes; the final production diff is the parked proof, not a replacement of main's files. No restricted emitter/orchestration file was edited. The original review/compiler/string-brand-proof evidence remains historical and untouched; fresh recut evidence is in this directory.

Current main's counts writer preserves unmeasured rows. The selected six count rows were regenerated successfully once after runtime product setup, followed by the table-integrity and six-row comparison tests. counts-delta.json proves precisely six additions, no deleted row and no changed existing row. The first cold counts attempt timed out before TestMain could write a table; it did not regenerate anything.

The six new programs are escapeLeadingUnderscores, unescapeLeadingUnderscores, Map<__String, Symbol> hits/misses, __String | undefined function/closure returns, a separate plain phantom brand, and nullable __String values. Every new admission was held to source Node, native ASan/UBSan, release native, backend Node and leak checks. All six have balanced allocation/free counts and zero remaining region values.

## Fresh admission delta

admission.json records a fresh current-main CLI and delivery CLI, both built from the resolved current tree. The baseline overlay restores current main's expression.go, refusals.go, cast.go and cast_proof.go, not their older snapshots. No prior baseline observations were reused. Across 1,363 oracle testdata .a/.ts files (excluding .d.ts), there are exactly six new admissions, all named string_brand_ fixtures, and no losses. Both requested first-error reasons are 1 on this baseline and 0 on delivery. This is a local first-error scan; it does not establish that all historical 28 value and 16 return sites close.

## Main's newer rules

All test shells set GOMAXPROCS=4; cpu.max is 400000 100000 and nproc reports 5. Ordinary test runs use -timeout 60s. leaf-seconds.json records every passing test from the final runs, maximum 38.31s. Every new or touched test leaf is below 60 seconds. Existing build-ahead product leaves were 32.46s sanitized, 28.78s counted and 15.40s release; those were prepared separately before ordinary fixture runs.

The proof introduces no new IR map fields or product declarations. Main's SortedMap fields are retained. TestEveryMapAProgramHoldsIsSorted, TestAProgramIsTheSameBytesEveryTimeGobWritesIt and TestASortedMapIsNilWhereAPlainMapIs all pass. The new IR is the existing string representation and Defined check; it adds no build-directory name to an artifact. The six emitted programs require only the already-declared base runtime feature set (runtime-features.json).

The optional ADAMIC_BUILD_CACHE=read audit failed before running native programs because this local cache has no traced Workshop read-set manifests: product native-runtime was not built under the required trace. The ordinary uncached source-Node checks and existing product builders passed. This does not establish runner-cache shipping parity; the limitation is in fixtures-products-read.log. No cache policy or main product machinery was altered to bypass it.

## Commands and observed outputs

All output went directly to logs. Commands ran after sourcing /workspace/adamic-tools/env.sh with GOMAXPROCS=4.

- timeout 300 go test ./internal/lower -run '^(TestStringBrand|TestClockGenericReturnsT01DeclinesBrandedPrimitives|TestMixedUnionContractPhantomBrandUsesPrimitiveBase|TestMixedUnionContractPhantomVoidIsUndefined)' -v -count=1 -timeout 60s: PASS, 1.965s (lower.log).
- timeout 300 go test ./internal/oracle -run '^TestProduct_NativeRuntimes(Sanitized|Release|Counted)0$' -parallel=1 -v -count=1 -timeout 240s: PASS, 76.689s total over three existing product leaves (runtime-products.log).
- ADAMIC_GATE_UNCACHED=1 timeout 180 go test ./internal/oracle -run '^TestStringBrand' -v -count=1 -timeout 60s: PASS, 11.229s (fixtures-green.log).
- timeout 180 go test ./internal/oracle -run '^TestCountsAreRecorded_[0-9]+$/internal/oracle/testdata/string_brand_' -v -count=1 -timeout 60s -args -update-counts: PASS, 8.260s (counts-update-green.log).
- timeout 120 go test ./internal/oracle -run '^TestCountsAreRecordedTable$|^TestCountsAreRecorded_[0-9]+$/internal/oracle/testdata/string_brand_' -v -count=1 -timeout 60s: PASS, 18.902s (counts-check.log).
- timeout 300 go test ./internal/ir -run '^(TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted|TestAProgramIsTheSameBytesEveryTimeGobWritesIt|TestASortedMapIsNilWhereAPlainMapIs)$' -v -count=1 -timeout 60s: PASS, 38.351s, reader leaf 38.31s (readers-and-maps-retry.log).
- timeout 180 go vet ./internal/lower ./internal/oracle: exit 0, no output (vet.log).
- timeout 780 python3 review/compiler/loom-string-brand-2/run-mutants.py: all five caught, exit 0; every nested test exits 1 for its specific assertion and retains a 60s test limit plus bounded compilation (mutants-green.log, mutants.json).
- timeout 780 python3 review/compiler/loom-string-brand-2/admission.py: exit 0, 1,363 files, +6/-0; local __String reasons each 1 -> 0 (admission.log).
- gofmt -l on the eight changed Go files and git diff --check: no output (gofmt.log, diff-check.log).

## Mutants

| Mutant | Catcher | Observed failure |
| --- | --- | --- |
| real-field-phantom | TestStringBrandRealRefused | named actual refusal required; got nil |
| drop-undefined | TestStringBrandOptional | source Node exit 0 versus backend exit 70 |
| read-brand | TestStringBrandReadRefused | expected named __brand refusal; different lowering stop |
| literal-refinement | TestStringBrandCastKeepsLiterals | unproven string literal refinement admitted |
| unchecked-missing-cast | TestStringBrandCastChecksMissing | missing checked undefined removal |

Mutated source snapshots are .go.txt, overlaid onto current source without editing production files. None was counted as caught for a compiler build error, timeout or warning.

## Setup and retries

GOPROXY was set to https://proxy.golang.org|direct before bash cloud/setup.sh. The first setup build overlapped the squash and enumerated a partial file set, stopping on missing new proof methods. The completed-tree retry passed. Its timing lines: Node ready 0.077s, Go ready 0.077s, markdown ready 0.209s, submodules ready 0.247s, clang ready 0.604s, shared cache ready 3.455s, Go build ready 128.141s, tests deferred 128.357s, build cache warm 128.360s, done 128.467s. Tools are Go 1.27.1, clang 20.1.8 and Node 24.19.0. The setup printed /workspace/adamic-tools/env.sh, which was sourced. npm ci --prefix stage3/api installed the pinned declarations.

Initial fixture/count/mutant attempts hit the 60s limit while cold runtime archives were being compiled inside ordinary tests. Main now declares these as build-ahead products. Preparing the required existing product group before retrying restored the intended leaf grain, without raising ordinary test limits or changing a test's meaning. The reader exports were also warmed separately under a 300s bound. The final reader leaf passed below 60s.

The required integration lane command runs on the sole committed delivery branch before push; its post-commit output is retained locally in lane-checks.log and reported in the final response. No whole package tests, full gate, PR or Loom submission were run.
