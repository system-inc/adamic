Built require-await, symbol-description and valid-typeof as native Adamic .a modules with AST-name listener declarations and handed-node execution.
Claim e21595a11 was pushed before implementation; current-main landing refresh follows this implementation commit.
217 valid controls produce 125 findings with identical full diagnostic/fix/suggestion bytes; 287 repository and 77 compiler roots also agree, including ASan/UBSan.
Three rule mutants, three raw-question mutants and a listener metadata mutant are caught; new released questions refuse with panic 70.
Default options only; React claims remain parked, native is slower than Go, and the full repository gate is not covered.

## Implementation and validation

Rules receive cached parsed nodes through an owned numeric kind-indexed driver. Parser node lookup happens once in the driver; rules neither refetch parser nodes nor compare kind strings. rule.json and compiled exported listener kinds are compared against production Go listener maps. Shared harness and generator sources are untouched. The existing owned question dispatcher adds only registration and routing lines.

require-await implements contextual promise contracts, generic substitutions, callback signatures, heritage contracts and suggestion ranges. New raw checker questions live in generic_call_signature.go, type_index.go and heritage_types.go with matching Adamic decoders. They expose checker facts, with lint decisions in Adamic. Symbol provenance uses existing raw bridge support. valid-typeof implements the default option, including global undefined suggestions; requireStringLiterals:true is not implemented.

The independent Go oracle imports pinned production cohere rules. Of 227 candidate roots (67 hand controls and 160 extracted production fixture literals), Go rejects 10 syntactically invalid roots before comparison. All 217 remaining roots are compared; native refusal does not filter them. Corpus manifests are the frozen validation-coverage manifests, not a new sample. All diagnostics, ranges, fixes and ordered suggestions are serialized and compared byte for byte. Corpus populations yield zero findings, so controls provide the positive findings and suggestion checks. This evidence establishes tested parity, not exhaustive correctness on arbitrary programs.

Mutants each exit successfully with empty stderr and are caught by the byte comparison: require-await suggestion wording; Symbol declaration provenance inversion; valid-typeof undefined wording; resolved generic signature target removal; numeric index replacement with string index; and heritage root omission. A rule.json CallExpression-to-BinaryExpression mutation fails the production listener comparison. Three released raw-question calls require exact panic 70. Normal and sanitized control and corpus runs pass. Bridge tests and vet pass; the complete repository gate was not run.

Reproduction (test output goes directly to files):

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave-24-fifth/validate.py /workspace/wave24-fifth-check --stage0 /tmp/wave24-landing3-adamic --compiler /workspace/wave-24-corpus > /tmp/wave24-fifth-check.log 2>&1
go test ./bridge/tsgo/... -count=1 > /tmp/wave24-fifth-bridge.log 2>&1
go vet ./bridge/tsgo/... > /tmp/wave24-fifth-vet.log 2>&1
```

Compiler can be rebuilt with go build -o /tmp/adamic ./cmd/adamic. Setup tools succeeded: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s. Cache warming exited 1 with no space left on device in /tmp; there are no successful cache-warm or done timing lines. nproc=5. Owned targeted gates above completed despite that failure. Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0.

## Timing

Three alternating whole-process count runs, no build time: repository native median 273.655 ms versus Go 93.345 ms (2.932x slower); compiler native 1434.151 ms versus Go 277.132 ms (5.175x slower). Samples are in evidence/timings.json. Driver numeric kind lookup uses a map built once per file. These zero-finding corpora include load/parse/walk costs and demonstrate no native speed advantage. Earlier concurrent measurements are not used for the final timing claim.

Evidence contains complete run logs, canonical outputs compressed with gzip, setup failure, bridge/vet logs, listener comparisons, mutants and source-selection audit. No binaries or generated C archives are committed. React native HIR/SSA/capture work remains parked pending #dnv6f2c and JSX integration; require-atomic-updates also needs capture-sensitive control-flow graph support. No further eligible non-React rule remains in the last audit after these three claims.

## Landing refresh

Rebased onto origin/main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. No bridge, native compiler or type-aware shared harness source changed in that main advance; the compiler binary was reused. The new full validator passes after rebase (217 controls, 125 findings, both corpora and sanitizer modes, all six compiled byte mutants and three released-question refusals). Bridge tests pass (98.500s; checker 0.409s), vet and gofmt are clean. All nine earlier rule suites and listener metadata are rechecked; detailed results follow in this evidence directory.

Pins: cohere 715ba94f3608a6500086b1076ce5cb7e51b836db; typescript-go 8d550c837c90bd1805b047b7eeccc2baac2d5e7a; TypeScript v6.0.3 corpus 050880ce59e30b356b686bd3144efe24f875ebc8. source-hashes.json hashes new native modules, metadata, raw questions, production Go sources/tests and the frozen manifests. Previous batches' coverage and mutants are documented in ../WAVE_24_REPORT.md, ../wave-24-next/README.md and ../wave-24-third/README.md.

The owned timing runner reproduces three alternating counts, checks successful execution, native empty stderr, Go numeric phase stderr and identical counts, and writes samples and process output to artifact files:

```sh
python3 stage1/cohere/typeaware/wave-24-fifth/timing.py /workspace/wave24-fifth-landing --repository /workspace/adamic --compiler /workspace/wave-24-corpus > /tmp/wave24-fifth-landing-timing.log 2>&1
```

All nine earlier ports pass after this rebase: original controls 43 findings and six byte mutants plus released-registry mutation; next controls 65 findings and six byte mutants; third controls 295 findings and five byte mutants. Every suite compares both complete corpora in normal and sanitized modes and checks released handles. Nine existing listener declarations match production and their native/JSON mutants fail. No new eligible non-React/non-capture-analysis rule was selected. Main was verified unchanged at c01907a70 after these tests.

The first timing-runner attempt wrongly required empty Go stderr and failed on the expected phase-timing line; the corrected runner permits only that numeric line. The failed attempt is retained as timing-initial-failure.log. It did not affect diagnostic validation.

## Post-push selection audit

The tested implementation and landing report were pushed as 8f07c6af94ed4b41d2bcb97f3a988f2cb099fb43. A normal force-with-lease attempt refused because the default origin refspec fetches only main and the worker tracking ref was stale. An explicit lease on the independently verified, previously pushed claim tip e21595a114b30168da537ff0fbd46772ce674377 succeeded; no concurrent worker changes were overwritten. Both attempts are retained.

After that push, explicitly refreshed all origin heads and scanned 569 origin refs and 33 distinct claim blobs. The conservative mention-based audit counts 164 claimed rules. Eight unclaimed candidates remain, all React rules under the user's parking instruction. No eligible non-React rule remains, so no new claim is made. The complete audit is evidence/wave24-fifth-post-push-audit.json. Main remains c01907a70 and the working tree is clean.

The current named-kind contract and latest full rechecks supersede numeric declaration descriptions above; see [KIND_NAMES_REPORT.md](KIND_NAMES_REPORT.md).

Latest requested lint-area rebase and validation: [AREA_LANDING_REPORT.md](AREA_LANDING_REPORT.md).
