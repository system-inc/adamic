Built: no-undef-init, prefer-for-of and consistent-indexed-object-style, plus raw type-reference-graph facts.
Commits: claim 54694064; implementation 82eec308b2a313467f1bbc06c5c35fce6809799b, based on 0d540f4.
Commands and outputs: wave gate PASS; 173 control, 16 repository, 24 compiler findings; bridge, source lint, vet and filtered Node gates PASS.
Mutants: three rule mutants, two graph mutants, released-registry mutant and the bridge/decoder/refusal/Node mutants below were caught.
Not covered: two malformed recovery probes, alternate rule options, full repository gate and rerunning all 26 previous rule suites.

# Wave 07 report

The claim was committed and pushed before implementation. The branch is
`codex/typeaware-wave-07`, from `origin/codex/tsgo-c-library` at
`0d540f413625f016f20fea39761c7b184f335de6`. The specific base overrides the
unit's generic main-branch instruction. All origin heads and their stage1 claims
were fetched and searched before selection. No rule was skipped.

## Selection and implementation

The combined by-volume ranking counts compiler and repository findings, excluding
the 26 ports on the base. Its unported positions are:

| Position | Rule | Compiler | Repository | Combined |
| --- | --- | ---: | ---: | ---: |
| 19 | no-undef-init | 1 | 14 | 15 |
| 20 | @typescript-eslint/prefer-for-of | 11 | 2 | 13 |
| 21 | @typescript-eslint/consistent-indexed-object-style | 12 | 0 | 12 |

Each judgment is in its own `.a` file, with a dedicated wave runner and independent
production Go oracle. No new Adamic `.ts` file was added. Existing dependencies
remain unchanged. `no-undef-init` preserves shadowing and the difference between
let fixes and var/destructuring reports. `prefer-for-of` preserves exact symbol
lookup, shorthand handling, loop spellings and assignment exceptions.
`consistent-indexed-object-style` preserves readonly/optional wrappers, interface
modifiers, comment-dependent suggestions, unsafe fix refusal and circular types.

The only shared file edit is one switch registration in
`bridge/tsgo/checker/facts.go` (two physical lines after gofmt). The new question
has separate Go and Adamic files. Adamic asks it by string, so there is no shared
Adamic registry edit. It exposes AST kind/edges, direct checker symbols and their
declarations, including other files. The circularity decision stays in Adamic;
the bridge returns no lint verdict. See
[the wire contract](../../../bridge/tsgo/checker/type_reference_graph.md).
No protected compiler files or submodule pins changed.

Production oracle pins:

- cohere: `715ba94f3608a6500086b1076ce5cb7e51b836db`.
- cohere/TypeScript: `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.
- Compiler corpus: TypeScript v6.0.3,
  `050880ce59e30b356b686bd3144efe24f875ebc8`.

## Complete finding agreement

The oracle calls the unchanged production Go rules, with their default options.
Both runners sort and serialize all findings, fixes and suggestions. Comparisons
include file headers and the summary, preserve duplicate diagnostics, and use
UTF-8 byte offsets. The frozen base manifests select exactly 77 compiler roots
and 287 repository roots; these are not expanded to include wave implementation
files. The roots and their source hashes are preserved in `validation-wave-07`.

| Corpus | Roots | Findings | Fix edits | Suggestions | Complete output bytes |
| --- | ---: | ---: | ---: | ---: | ---: |
| Controls | 283 | 173 | 69 | 7 | 42,541 |
| Repository | 287 | 16 | 14 | 0 | 22,753 |
| Compiler | 77 | 24 | 13 | 0 | 9,282 |

Normal and ASan/UBSan/LeakSanitizer native builds match independent Go bytes on
all three. Sanitizer processes exit successfully without sanitizer errors. The
controls contain 26 no-undef-init, 78 prefer-for-of and 69 indexed-style findings.
Production test literals are extracted through Go AST, deduplicated, and augmented
with Unicode/CRLF, shadowing, numeric, shorthand, assignee, comment and circularity
controls. Bare yield probes are wrapped in a generator for both implementations.
Two malformed empty index signatures (`interface Foo { []; }` and
`type Foo = { [] };`) are explicitly excluded because native parsing refuses
these recovery forms. All 283 remaining sources pass the independent parse gate.

Raw compressed output, control sources and SHA-256 values are committed in
[validation-wave-07](validation-wave-07/streams.json). The three implementations
have identical hashes for each corpus. Absolute path headers explain differences
in control byte lengths between scratch directory names.

## Mutants and refusal checks

The first five mutants compile and exit 0 with empty stderr. Only the independent
Go diagnostic byte comparison catches them, rather than compilation or a crash.
Offsets below are zero-based first differing bytes in the final control run.

| Mutant | Detector and observed result |
| --- | --- |
| no-undef-init removal start plus one | Go bytes differ at 15,180 |
| prefer-for-of body verdict inverted | Go bytes differ at 541 |
| indexed-style diagnostic punctuation changed | Go bytes differ at 164 |
| Adamic graph symbol equality inverted | Go bytes differ at 219 |
| Go graph symbol identity omitted | Go bytes differ at 8,055 |
| Released-program registry deletion omitted | Mutant exits 0; required panic 70 assertion fails |

An ordinary new-question call after release produces exactly
`adamic: panic: invalid or released checker handle\n` with exit 70.
The direct checker graph test verifies identities against direct symbol lookups,
a two-file declaration cycle, local record IDs, an opaque array constructor and
suffix rejection.

The unchanged bridge gate was also rerun and caught all its mutants:

| Mutant | Detector |
| --- | --- |
| Input length plus one | ASan heap-buffer-overflow |
| Output string length plus one | ASan heap-buffer-overflow |
| Released handle kept live | Stale-handle assertion |
| C output free removed | LeakSanitizer leaked-buffer report |
| Wrong-kind guard bypassed | Expected panic 70, mutant exits 0 |
| Unknown-question guard bypassed | Expected panic 70, mutant exits 0 |
| Native Node oracle output changed by one byte | Independent Node byte comparison |

`TestFactsDecoderGuards` exercises all 19 malformed frame mutations: empty,
length-negative, length-short, trailing, version, question, not-integer,
rounded-integer, noncanonical-integer, negative-natural, bad-boolean, zero-id,
negative-tuple, missing-root, missing-constraint, missing-element, duplicate-id,
missing-link-17 and missing-link-18. Each is caught by its expected panic 70.
Exact output and the emitted mutant streams are preserved with the evidence.

## Commands, gates and setup

All test output was redirected to log files, never piped. The following commands
ran from the repository, after `source /workspace/adamic-tools/env.sh`:

```sh
bash cloud/setup.sh > /tmp/wave-07-setup.log 2>&1
ADAMIC_WAVE07_ARTIFACTS=/workspace/wave-07-final-test \
ADAMIC_WAVE07_COMPILER_MANIFEST=/workspace/wave-07-artifacts/compiler.manifest \
ADAMIC_WAVE07_REPOSITORY_MANIFEST=/workspace/wave-07-artifacts/repository.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-07-typescript \
go test ./stage1/cohere/typeaware -run '^TestWave07AgreementAndMutants$' -count=1 -v -timeout 30m > /workspace/wave-07-artifacts/full-test-final.log 2>&1
ADAMIC_TSGO_CORPUS=/workspace/wave-07-typescript \
go test ./bridge/tsgo/... -count=1 -v -timeout 15m > /workspace/wave-07-artifacts/bridge-test.log 2>&1
go test ./stage1/cohere/typeaware -run '^(TestFactsDecoderGuards|TestInspectRequestRefusals|TestPinnedTypeFlags)$' -count=1 -v -timeout 30m > /workspace/wave-07-artifacts/facts-test.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -v -timeout 30m > /workspace/wave-07-artifacts/node-oracle-test.log 2>&1
go vet ./... > /workspace/wave-07-artifacts/vet-final.log 2>&1
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware > /workspace/wave-07-artifacts/gofmt-final.log 2>&1
```

Results: wave gate PASS in 74.085 s; bridge TestBridge PASS in 148.32 s and
checker package PASS in 0.376 s; fact/refusal gate PASS in 33.236 s; filtered Node
oracle PASS in 12.361 s. The Go subtest regexp also selected `method_closures.a`
and `generic_functions.a`; eight native cases passed, plus the one-byte mutant.
Vet and gofmt both exit 0 with empty output.

Setup succeeded: Go 1.27.1 ready 0 s, clang 20.1.8 ready 0 s, Node 24.19.0 ready
1 s, submodules ready 1 s, build-cache warm 81 s, total 81 s. `nproc` is 5;
`cpu.max` is `400000 100000`, reported memory 17.6 GB. Setup chose
`/workspace/adamic-tools/env.sh` instead of the suggested `/opt` path.

The pinned cohere CLI does not recognize `.a` inputs and reports nothing to
check. For the source gate only, isolated worktrees built cohere
`e7cfe4d1aceb524bc6f5936564284d54ca63d771` with its TypeScript
`d92d9bfee114c80be2c375d72edae966176e3a4f`. The following complete no-fix lint
check passes, with 276 rules on all eight new `.a` files:

```sh
/workspace/wave-07-artifacts/cohere-adamic --no-cache --no-fix \
stage1/cohere/typeaware/{no_undef_init,prefer_for_of,consistent_indexed_object_style,type_reference_graph,reference_node,wave_symbol,wave_node,wave_07_suite}.a \
> /workspace/wave-07-artifacts/source-lint-final.log 2>&1
```

This source-gate workaround did not change the pinned comparison oracle. The
lint-suggested logical assignment was replaced with an explicit if because
stage 0 refuses `||=`; the final comparison gate was rerun after that change.

## Native time against Go

Three alternating full-output rounds ran after all builds and gates finished,
with `ADAMIC_TSGO_TIMING=1`. Each round compares complete diagnostic bytes.
The small [timing script](validation-wave-07/full_bench.py) is the base profiling
script changed to omit `--count` and recognize the final summary line. Saved
stdout, stderr and [raw records](validation-wave-07/bench/results.json) reproduce
these median seconds:

| Corpus | Go process | Native process | Native / Go | Load Go / native | Run Go / native | Native query | Queries |
| --- | ---: | ---: | ---: | --- | --- | ---: | ---: |
| repository | 0.117998 | 0.269499 | 2.28x | 0.056770 / 0.062742 | 0.047554 / 0.201192 | 0.014272 | 409 |
| compiler | 0.314149 | 2.030888 | 6.46x | 0.234846 / 0.235306 | 0.060197 / 1.771218 | 0.200948 | 1278 |

Native is slower on these corpora. Process time includes startup, loading,
walking, rendering, output and teardown. Load/query/run phases are not separate
additive costs: native query time is inside native run time. Go rule medians are
0.003761 s (repository) and 0.020684 s (compiler); native run time also includes
the native parser and byte-offset work. No inference about CPU architecture or
future speed is made from these measurements.

The timing command used the final binaries and the same manifests:

```sh
python3 /workspace/wave-07-artifacts/full_bench.py \
/workspace/wave-07-final-test/wave07 /workspace/wave-07-final-test/wave07-oracle \
/workspace/wave-07-artifacts/bench \
--corpus repository /workspace/adamic/tsconfig.json /workspace/wave-07-artifacts/repository.manifest \
--corpus compiler /workspace/wave-07-typescript/src/compiler/tsconfig.json /workspace/wave-07-artifacts/compiler.manifest \
--rounds 3 > /workspace/wave-07-artifacts/bench.log 2>&1
```

## Limits

Agreement is for the pinned default-option rules, controls and exact frozen
corpora. Alternate indexed-style `index-signature` mode and configurable rule
options are not implemented in this wave's runner. The two named malformed
recovery probes remain uncovered. The comparison serializes fixes and
suggestions; it does not replace the CLI's suppression or edit-application tests.
The whole repository `go test ./...`, every previous type-aware rule suite and a
repository-wide lint run were not rerun. Targeted affected packages, foundation
bridge guards, selected Node oracles and the new-file source lint gate were run.
No PR was opened.
