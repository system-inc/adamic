Built: the three claimed wave-03 rules as native Adamic .a files, with three raw checker questions.
Commits: claim 85651ba0; implementation 24c29dbc4f68a9dfd146fc6aef711700f3f83ea0; this report is in the following commit.
Commands and outputs: wave-03 comparison PASS, 364 corpus files and 264 findings; sanitizer and released-handle checks PASS; shared refusal harness BLOCKED.
Mutants: all three rule mutants caught only by finding bytes; three checker-fact mutants caught by direct API comparisons; metadata and released-registry mutants caught by refusal expectations.
Not covered: decorator metadata imports, nondefault rule options, emitted JavaScript comparison, exhaustive upstream fixtures, full repository gate; native is slower than Go.

## Scope and claim

Branch `codex/typeaware-wave-03` starts at `origin/codex/tsgo-c-library`,
`0d540f413625f016f20fea39761c7b184f335de6`. CLAUDE.md, language and memory
documents, and the named type-aware README and volume/coverage reports were read
before implementation. The specific requested bridge base takes precedence over
the generic instruction to start at main.

The claim was committed and pushed before code. All 276 fetched origin heads were
checked for claims and named ports. No rule was skipped. After excluding the 26
base ports, positions 7, 8 and 9 are:

| Rule | Compiler findings | Repository findings | Combined |
| --- | ---: | ---: | ---: |
| @typescript-eslint/no-unsafe-call | 91 | 4 | 95 |
| @typescript-eslint/consistent-type-imports | 83 | 2 | 85 |
| @typescript-eslint/no-unnecessary-template-expression | 4 | 80 | 84 |

Each rule has its own .a file. Helpers and a standalone wave-03 driver are also
.a. No shared registration generator, existing test harness, protected compiler
file, or submodule was edited. The only existing source change is the one-line
checker dispatcher fallback in `bridge/tsgo/checker/facts.go`, authorized by the
original unit instruction. All other source files are new and specific to this
wave. No additional rules were claimed.

## Observed comparison

Production Go cohere runs the three unmodified rules through an independent Go
loader and walker. The oracle imports no bridge code. Full canonical records
include UTF-16 finding ranges, rule names, message IDs and text, every fix range
and replacement, and every suggestion. Sorting uses the complete record.

The compiler population is the same 77 roots pinned by the base coverage report,
TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8` (6.0.3). The repository
population is the frozen 287-file base coverage manifest. It is not expanded to
include the new ports. The cohere submodule remains at
`715ba94f3608a6500086b1076ce5cb7e51b836db`; its TypeScript submodule remains at
`8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.

| Population | Files | Findings | Canonical bytes | Findings with fixes | Suggestions |
| --- | ---: | ---: | ---: | ---: | ---: |
| Compiler | 77 | 178 | 376067 | 75 | 0 |
| Repository | 287 | 86 | 34996 | 82 | 0 |
| Controls and helper | 21 | 46 | 9790 | compared | 0 |
| noImplicitThis false control | 1 | 4 | 923 | 0 | 0 |

Go and native bytes match for every row. ASan/UBSan/LSan runs match on the compiler,
repository and main controls, with empty native sanitizer stderr. The controls
exercise any/error types, Function signatures, import aliases and mixed imports,
shadowing and export types, literal/nested templates, escaping, numeric/bigint/
regex text, comments, CRLF and supplementary Unicode. Zero suggestions means the
serializer is compared but these rules did not emit a positive suggestion.

Evidence lives in `validation-wave-03`: relative manifests, per-source SHA-256s,
counts, measurements, test logs, and compressed full stdout/stderr for every
wave-03 build and run. `stream-hashes.json` records hashes of the uncompressed
streams. File headers contain the original absolute workspace paths.

## Checker boundary

All questions use the existing schema-1 framed string ABI and existing C output
ownership. They return compiler facts, not rule decisions. Type and symbol IDs
are scoped to the live program. No pointers escape into Adamic.

| Question | Node / argument | Fields after version and question |
| --- | --- | --- |
| signature-kinds | validated node and canonical nonzero type identity | construct count, call count, then return-type presence and flags for each call |
| node-binding | Identifier | own symbol ID, local export-target symbol ID, IsPartOfTypeNode boolean |
| import-runtime-options | SourceFile | emitDecoratorMetadata boolean, jsxFactory text, jsxFragmentFactory text |

Go implementations and Adamic decoders each have files named for the question.
The new wave dispatcher keeps the existing registration edit to one line.
Unknown questions still return the same error, now from the new dispatcher file.
Malformed questions, wrong node kinds and invalid type identities are refused.
A query after release panics with exit 70 and `invalid or released checker handle`.

## Mutants actually run

| Mutant | Observed catcher |
| --- | --- |
| unsafe-call: construct count threshold changed from > 0 to > 1000000 | builds, exits 0, empty stderr; byte oracle differs at byte 1524 |
| imports: binding-use discovery sets found false | builds, exits 0, empty stderr; byte oracle differs at byte 1614 |
| template: interpolation removal starts one UTF-16 unit later | builds, exits 0, empty stderr; byte oracle differs at byte 5086 |
| signature question: construct count uses call signatures | direct checker API test rejects counts 2/2 instead of 1/2 |
| binding question: IsPartOfTypeNode always false | direct checker API test rejects the type-node field |
| options question: jsxFactory text replaced with wrong | direct checker API test rejects 0/wrong/h.Fragment |
| metadata guard disabled | exits 0 instead of required panic 70 |
| registry retains released checker handle | exits 0 instead of required panic 70 |

The three rule mutants are not killed by compilation or sanitizers. Only the
independent finding comparison catches them. Checker mutants use Go overlays and
were never applied to repository source. Existing decoder tests also caught all
19 malformed-wire mutants; the existing wrong-kind guard mutant was caught.
The filtered Node oracle's own one-byte mutant was caught. The existing bridge
regression also ran seven mutants: input length +1 and output length +1 each
produced ASan heap-buffer-overflow; a retained released handle failed the stale-
handle assertion; source-file type lookup differed at byte 6; removed link opt-in
failed the refusal expectation; removed C output free and unowned region heap
allocation each produced LeakSanitizer failures.

## Timing

Single warm wall-clock samples, including process startup, checker construction,
rule traversal and canonical output. Corpus timings ran before the additional
regression jobs. Native and Go each run the same three rules. These observations
do not establish statistical performance or an optimization claim.

| Population | Native wall | Go wall | Native / Go | Native checker queries |
| --- | ---: | ---: | ---: | ---: |
| Repository | 0.763368062 s | 0.460611944 s | 1.66 | 26405 |
| Compiler | 5.777461592 s | 1.938457870 s | 2.98 | 185745 |

Repository native load/query/run: 0.155745436 / 0.242580827 / 0.597911797 s.
Compiler native load/query/run: 0.781942250 / 2.097717467 / 4.932240005 s.
Repository Go load/rule/run: 0.171049152 / 0.124057872 / 0.261983384 s.
Compiler Go load/rule/run: 0.596996928 / 1.087205492 / 1.266589682 s.
Phase timers overlap and are not added together. The native bridge still links
the Go checker; native sanitizers do not instrument Go-managed heap memory.

## Setup, commands and workarounds

`bash cloud/setup.sh` succeeded: Go ready 0s, clang ready 1s, Node ready 1s,
submodules ready 1s, cache warm 174s, total 174s. `nproc` printed 5; CPU cgroup
allows four cores. Go 1.27.1, clang 20.1.8, Node 24.19.0. Each build shell sources
`/workspace/adamic-tools/env.sh`; TMPDIR is the world-traversable gate directory.

The first broad recursive fetch and a full-history corpus clone were stopped
after obtaining origin heads because they were downloading unnecessary history.
The compiler corpus was obtained from the public codeload tarball at its exact
pinned commit. All 77 and 287 manifest paths were present and hashed.

Commands below send test output directly to files, never through a pipe:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE03_ARTIFACTS=/workspace/wave-03/final \
ADAMIC_WAVE03_REPOSITORY_MANIFEST=/workspace/wave-03/repository.manifest \
ADAMIC_WAVE03_COMPILER_MANIFEST=/workspace/wave-03/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-03/typescript-corpus \
go test ./stage1/cohere/typeaware -run '^TestWave03AgreementAndMutants$' \
  -count=1 -timeout=30m -v > /workspace/wave-03/final.log 2>&1
# PASS, 219.940s

go test ./bridge/tsgo/checker -count=1 -v > /workspace/wave-03/checker-final.log 2>&1
# PASS, 1.602s

go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-03/bridge-final.log 2>&1
# PASS: bridge 244.044s, checker 1.194s, other packages have no tests

go test ./stage1/cohere/typeaware \
  -run '^(TestFactsDecoderGuards|TestInspectRequestRefusals|TestPinnedTypeFlags|TestSixPinnedFlags)$' \
  -count=1 -timeout=10m -v > /workspace/wave-03/decoder-final.log 2>&1
# FAIL, 118.303s: nonunique mutant unknown-question

go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
  -count=1 -timeout=10m -v > /workspace/wave-03/node-final.log 2>&1
# PASS, 97.418s; eight selected fixtures, including method_closures and generic_functions

go vet ./bridge/tsgo/... ./stage1/cohere/typeaware > /workspace/wave-03/vet.log 2>&1
# exit 0, empty log

gofmt -l bridge/tsgo/checker stage1/cohere/typeaware > /workspace/wave-03/gofmt.log
# empty log

git diff --check
# exit 0
```

The pinned production CLI cannot load the new .a roots by itself. An initial
ordinary cohere check failed because those roots were absent from its program.
A scratch Go loader overlay adds AllowNonTsExtensions and read-only .a.ts module
aliases to real .a files. It changes no rule logic, repository loader or submodule.
The overlay source is preserved as `cohere-program.go.txt`, not as a Go package.
Its historical overlay JSON and tsconfig record the actual absolute scratch
paths; adjust them to local paths when reproducing. This is separate from the
independent finding oracle. With this overlay:

```sh
go build -C cohere -overlay /workspace/wave-03/cohere-overlay.json \
  -o /workspace/wave-03/cohere-a ./command/cohere
/workspace/wave-03/cohere-a --directory /workspace/adamic \
  --tsconfig /workspace/wave-03/cohere-tsconfig.json --no-cache --no-fix \
  stage1/cohere/typeaware/*.a > /workspace/wave-03/cohere-final.log 2>&1
# 276 rules, 10 checked, 100% Adamic-ready (10 of 10), zero findings
```

## Integration blocker and coverage limits

The existing `TestInspectRequestRefusals` searches `facts.go` for the literal
unsupported-question return to replace it with a permissive mutant. The one-line
wave dispatcher registration moved that return to `wave_03_questions.go`. The
observable unknown-question error remains identical, but the shared test stops
before executing its unknown-question mutant: `nonunique mutant unknown-question`.
This is a real failed regression gate, not a pass. Ahra's correction says to stop
rather than edit shared files when blocked, so no shared harness or generator
change was made. The integration worker must reconcile the registration and the
mutation location before the complete shared gate can pass.

`consistent-type-imports` implements default prefer-type-imports/separate-type-
imports behavior. Decorator metadata can cause value references invisible in
ordinary source binding uses; this path is not implemented. The suite explicitly
refuses emitDecoratorMetadata true with panic 70, rather than returning incomplete
findings. Both pinned corpora disable it. Custom rule options and exhaustive
upstream fixtures were not run. JSX runtime option facts have direct checker
controls, but JSX cases were not exhaustively covered. No emitted JavaScript
comparison or full `go test ./...` / full `go vet ./...` was run. New Adamic source
files remain .a because the native compiler and isolated validation can load them.
