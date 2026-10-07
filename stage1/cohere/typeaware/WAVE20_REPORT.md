Built: native no-floating-promises, no-implied-eval, and no-meaningless-void-operator, plus two isolated checker questions.
Commits: base 0d540f413625f016f20fea39761c7b184f335de6; claim 2d839012 pushed before implementation; implementation is the commit containing this report.
Commands and outputs: wave20 parity gate PASS 119.669s; checker PASS 0.239s; bridge PASS 132.988s; filtered Node oracle PASS 23.665s; go vet and formatting checks clean.
Mutants: three rule decisions, two checker answers, and released-handle retention caught; earlier message/fix probes and existing bridge/Node failure probes also caught as detailed below.
Not covered: optional rule configurations, the entire upstream fixture matrix, the previous 26-rule full suite, or the full repository test gate; native is slower than Go on both measured corpora.

## Scope and claims

The explicit unit base superseded the generic main-base instruction. The branch is `codex/typeaware-wave-20`, based on `origin/codex/tsgo-c-library` at the commit above. Read CLAUDE.md and the three named typeaware reports before changes. Fetched all 268 origin heads and checked claim files and distinct typeaware source trees. No existing claim or port of these three rules was found, so none was skipped.

The ranking sums compiler-all and repository-all volume, sorts descending with lexical ties, and excludes the already ported rules. Of 197 checker-dependent entries, 25 appear among the existing 26 ports (method-signature-style is outside this table), leaving 172 entries. Remaining positions 58, 59, and 60 are respectively the three rules above. All have zero findings in the existing volume corpora. The claim file and commit preceded implementation and were pushed independently.

Each rule has its own `.a` file. Shared new helpers and the suite are also `.a`; no new Adamic source file is `.ts`. New Go and Adamic files implement `accessed-property` and `callback-parameters`. The former exposes checker-resolved property names, including enum element accesses; the latter exposes apparent first callback parameters with rest-array element indexing. The only existing source edit is two dispatch registrations in `bridge/tsgo/checker/facts.go`. No protected compiler files, pins, existing rules, or existing shared Adamic files changed.

Default Go cohere options are implemented: floating promises ignores void and does not enable custom thenable checks or allowlists; meaningless void does not enable checkNever. Findings include every message, position, automatic fix, and suggestion, compared as canonical bytes. The independent Go oracle invokes the pinned production rules and does not import the bridge or native rule code. Controls exercise promise arrays, handlers, enum keys, rest callbacks, logical and conditional expressions, declarations, global and local eval constructors, void type unions, numeric literals, assignments, casts, comments, and suggestions.

## Validation

Environment setup command was `bash cloud/setup.sh`, followed by sourcing `/workspace/adamic-tools/env.sh` for builds. Setup timings: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 163s, done 163s. `nproc` reported 5, with a four-CPU cgroup quota. Go 1.27.1, clang 20.1.8, Node 24.19.0. A recursive fetch of historical submodules was stopped after all remote heads were available; setup confirmed the pinned submodules ready. No pin was changed.

Tests wrote logs rather than piping output. Logs and compressed process streams are in [validation-wave20](validation-wave20/). The decisive commands, with the toolchain environment sourced, were:

```sh
ADAMIC_WAVE20_ARTIFACTS=/workspace/wave20-validation/complete \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave20-typescript \
go test ./stage1/cohere/typeaware -run '^TestWave20AgreementAndMutants$' -count=1 -v -timeout 30m

go test ./bridge/tsgo/checker -count=1 -v
go test ./bridge/tsgo -count=1 -v -timeout 15m
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|maps_and_text)\.a$' -count=1 -v -timeout 10m
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware
go vet ./...
```

| Corpus | Roots | Findings | Identical output bytes | Normal and ASan/UBSan/LSan |
| --- | ---: | ---: | ---: | --- |
| Generated positive controls | 17 | 48 | 12851 | Pass |
| Frozen repository manifest | 287 | 0 | 18485 | Pass |
| Frozen compiler manifest | 77 | 0 | 5241 | Pass |

Compiler sources are TypeScript v6.0.3, commit `050880ce59e30b356b686bd3144efe24f875ebc8`, using the branch's frozen `validation-coverage/compiler.manifest`. Repository roots use its frozen repository manifest. Output bytes include per-file headers even for zero findings. Positive controls are necessary because the two production corpora alone cannot distinguish silent omission of these rules.

The released accessed-property query exits with panic code 70 and `invalid or released checker handle`. Bridge tests additionally exercise 100 retained C query buffers, zero/stale handles, distinct handles, and 162 independent positions with 3261 matching bytes under sanitizers. The filtered Node oracle covers closures, method closures, generics, regions, maps and text under the existing native/Node comparison and sanitizer checks.

### Every mutant run

All five final rule/question mutants compile, exit 0, and have empty stderr. Only the independent byte comparison rejects them:

| Mutation | Detector |
| --- | --- |
| Reverse floating-promise unhandled guard | First differing byte 57 |
| Reverse implied-eval local declaration guard | First differing byte 6399 |
| Change void/undefined type mask from 16\|4 to 64\|4 | First differing byte 7638 |
| Append `-mutant` to accessed-property answer | First differing byte 457 |
| Replace rest callback indexed type with number type | First differing byte 9737 |
| Retain released program in registry | Mutant exits 0; required panic 70 assertion rejects it |

Earlier probes also ran successfully and were rejected: floating message identifier swap at byte 109, implied-eval message changing function to string at byte 6661, and void fix replacing empty text with a space at byte 7597. See `gate.log`; the final gate uses the stronger decision mutations above.

Existing bridge failure probes were run, not merely inspected: input length plus one and output length plus one trigger ASan heap-buffer-overflow; stale registry retention violates the released-handle assertion; wrong source-file type differs at oracle byte 6; removed link opt-in triggers the expected build refusal; omitted C output frees trigger LSan; region heap allocation triggers LSan. The filtered Node oracle's one-byte output mutant is caught by its comparison. Details are retained in `bridge.log` and `node-oracle.log`.

### Source tooling

The production cohere CLI rejects `.a` inputs as not TypeScript or JavaScript. Its format-only invocation is not accepted as coverage. Instead, the pinned native formatter and configured lint-rule runner were invoked directly through scratch Go overlays. The lint runner presents virtual `.ts` names in memory, backed by the real `.a` sources, and rewrites only import extensions in memory so the pinned checker resolves dependencies. No physical `.ts` version of the new Adamic sources was required by these checks. Adapter sources are saved as `.go.txt` evidence.

The valid lint run first found 12 issues, then one shadowed parameter after fixes, and finally zero findings across all eight `.a` files. Adamic refuses the suggested logical assignment, so explicit if guards satisfy both compiler and lint. Formatting checked idempotence on eight files. The first direct `.a` lint attempt yielded unresolved-import findings and was discarded as an invalid measurement. CLI rejection, final source checks, formatter output, and adapters are preserved. Go import grouping was normalized after the final gate without changing behavior.

## Timing

Three alternating native/Go executions per corpus ran after all concurrent builds stopped. Every timed output hash agrees; raw measurements are in `bench.json`, and `bench.py` records the procedure. These are whole-process wall-time medians, including loading; independent phase medians need not sum to the total.

| Corpus | Native seconds | Go seconds | Native / Go | Native run phase | Go run phase |
| --- | ---: | ---: | ---: | ---: | ---: |
| Repository | 0.449263 | 0.185041 | 2.43 | 0.357483 | 0.089477 |
| Compiler | 4.547673 | 0.856809 | 5.31 | 4.138846 | 0.425994 |

Native performs 17518 checker queries for repository roots and 271021 for compiler roots. These observations establish parity, not speed improvement. No findings-per-second metric is meaningful for these zero-finding corpora.

## Limits and reproduction

The gate builds all three rules with stage0, links the C checker, compares complete findings/fixes/suggestions, rebuilds mutants, and repeats controls and both corpora with ASan, UBSan, and LSan. It does not apply suggested edits and recompile them, test every upstream configuration/JSX variant, run the old 26-rule comprehensive suite, or run `go test ./...`. Run the recorded targeted commands with the frozen manifests and pinned external compiler checkout. Artifacts contain absolute workspace headers; regenerate them in a different workspace rather than treating path-dependent hashes as portable expected output. `source-sha256.json` records the validated source and manifest contents.
