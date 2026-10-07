Built native .a ports of no-throw-literal, no-useless-backreference and prefer-arrow-callback, with one isolated compiler question.
Previous batch pushed in 0df6cc4c and 0a14418e; claim abdad124 pushed before code; this accompanying commit contains implementation and evidence.
Rule gate PASS 65.980s: 364 valid controls, 209 findings, 364 corpus files; byte equality, sanitizers and released handles pass; timings below.
Three native rule mutants are caught only by Go byte comparison; the write-fact mutant fails the new compiler assertion; the Node byte mutant fails its oracle.
Not covered: shared profile integration, emitted-JavaScript comparison of these rules, nondefault callback options, full repository test gate, CLI fix application and parse-invalid inputs.

## Selection and scope

All previously claimed rules were completed, tested and pushed before this
batch. Fetch covered 348 origin refs and 77 distinct textual claim records.
The combined-volume ranking excludes ports on origin/codex/tsgo-c-library
at 5afbdb83da2ed7ad9815657cd3f6ececd5294bf6 and origin/main at
ef3d907ecdc4c771b016f7d9c52372def057a340, and every rule named in a claim
on any origin branch. Baseline TypeScript diagnostic names are normalized to
their registry prefix. Descending combined volume and lexical ties select
no-throw-literal, no-useless-backreference and prefer-arrow-callback, each with
recorded volume zero. The full origin-tip, ranking and exclusion snapshot,
selection script and successful claim push are in validation-wave-25-fourth.
Claim abdad124 was pushed before implementation. No further rules are claimed.

Every new Adamic source is .a. Shared registration generators and shared test
harnesses are unchanged, as are protected compiler files and submodule pins.
The only existing implementation edit is the dispatch case in this worker's
own process_questions.go. Tests reuse the existing harness without editing it;
the runner and outside Go oracle are specific to these three rules.

## Native behavior and compiler boundary

No-throw-literal reproduces cohere's expression-kind and operator decisions,
including its treatment of parentheses, and distinguishes global undefined
from a source binding. Findings cover the whole throw statement with the exact
object or undef message. It supplies no fixes or suggestions.

No-useless-backreference scans capture groups, alternatives and lookarounds
in Adamic. It preserves nested, disjunctive, forward, backward and negative
lookaround priority, duplicate named groups, legacy octals, Unicode/set flags,
character classes and the same malformed-pattern guards as cohere. Native
reference tracking follows global RegExp through aliases, globals, assignments,
destructuring and default values; scoped constant folding reads initializer
bindings and checks later writes. A literal and a constructor containing that
literal can both report. It supplies no fixes or suggestions.

Prefer-arrow-callback maintains native function frames, resolves self-reference
and arguments bindings, checks callback position and bind(this), and preserves
the generator, super and new.target exclusions. Native byte-indexed edits
remove function/name/bind syntax, insert the arrow and any needed parentheses,
preserve comments and decline all production repair guards. The class accepts
both callback settings; the comparison runner uses production defaults.

The new syntax-reference-facts question returns compiler declaration-name and
write-access booleans, meta-property keyword, binding/export property identity,
and the symbol read with its declaration provenance. It does not return a lint
verdict, RegExp classification, regex structure, callback decision or edits.
Its .a decoder caches facts and makes value-reference/write policy decisions
natively. Existing generic syntax and symbol questions are reused.

The source byte helper initially met a native compiler refusal on nested method
calls. Its implementation now calls a free offset lookup function. No compiler
file was changed; the final native build and comparisons pass.

## Independent comparison

The oracle invokes the unmodified production cohere registry at
715ba94f3608a6500086b1076ce5cb7e51b836db with its own compiler, checker and walk.
It imports no native bridge implementation and serializes complete findings,
ordered fix ranges/text and suggestions. Complete sorted bytes are compared.

| Population | Root files | Findings | Identical bytes | ASan/UBSan/LSan |
| --- | ---: | ---: | ---: | --- |
| Production and additional edge controls | 364 | 209 | 82212 | Pass |
| Frozen repository corpus | 287 | 0 | 18485 | Pass |
| TypeScript v6.0.3 src/compiler | 77 | 0 | 5010 | Pass |

Controls produce 19 throw, 133 backreference and 57 callback findings. The
callback findings include 48 with repairs, 117 ordered edits in total, and nine
that decline repair. All suggestions are zero, as these production rules
provide none. All five backreference message kinds and both throw kinds occur.
Inputs include aliases, local/global writes, initializer cycles, template
substitutions, duplicate named groups, async/comment repair guards, Unicode
names and CRLF. Only source inputs are extracted from production tests;
expected findings and edits come from the independent Go run.

367 controls were generated. The Go parser excluded three parse-invalid inputs:
`function f() { throw; }`, a string with a legacy octal escape, and a RegExp
call with a legacy octal string escape. Their exact text and disposition are
recorded in inputs.json. Nondefault callback test inputs are compared under
default options, not claimed as an options-matrix verification. External source
fixtures are .a files; scratch-only .ts symlinks preserve TypeScript identities.
No new .ts source file is written or committed.

The corpus uses the unchanged validation-coverage manifests. TypeScript is
pinned at 050880ce59e30b356b686bd3144efe24f875ebc8. Normal and sanitized native
comparison stderr is empty. Full output, hashes, fixture inputs, manifests and
logs are preserved in validation-wave-25-fourth.

## Mutants and lifetime

| Mutation | Observed result | Catcher |
| --- | --- | --- |
| Conditional throws never count as possibly Error | Native builds, exits 0, empty stderr; mismatch byte 1159 | Independent Go finding bytes |
| Forward backreference classified backward | Native builds, exits 0, empty stderr; mismatch byte 9603 | Independent Go finding bytes |
| Arrow repair inserts ` ->` | Native builds, exits 0, empty stderr; mismatch byte 56465 | Independent Go fix bytes |
| Compiler write access forced false, scratch Go overlay | Test fails with lost compiler write classification | TestSyntaxReferenceFacts |
| Native result byte changed, existing oracle mutant | TestTheOracleCatchesOneByte passes by detecting disagreement | External Node oracle |

The new compiler question queried after release exits 70 with exactly
`adamic: panic: invalid or released checker handle`. The full bridge gate passes
its ABI, sanitizer, lifetime and existing fault-injection checks. The new raw
question assertion also verifies declarations, shorthand value identity, local
export identity, new.target, binding property identity and malformed requests.

## Commands and measurements

Setup from this session ran bash cloud/setup.sh and sourced
/workspace/adamic-tools/env.sh. Its reported lines were:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (89s)
setup: done in 89s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

nproc is 5, with a four-core quota. Go 1.27.1, clang 20.1.8, Node v24.19.0.
All command output was written to logs without piping a test run.

```
ADAMIC_WAVE25_FOURTH_ARTIFACTS=/workspace/wave-25-fourth-final \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-25-corpus \
go test ./stage1/cohere/typeaware \
  -run '^TestWave25FourthAgreementAndMutants$' -count=1 -v
go test ./bridge/tsgo/... -count=1
go test ./bridge/tsgo/checker -run '^TestSyntaxReferenceFacts$' -count=1 -v
go test -overlay /workspace/wave-25-validation/fourth-reference-mutant.json \
  ./bridge/tsgo/checker -run '^TestSyntaxReferenceFacts$' -count=1 -v
go test -v -count=1 -timeout 10m ./internal/oracle \
  -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|classes|strings|array_from)\.a$'
go vet ./...
python3 /workspace/wave-25-validation/fourth-benchmark.py
```

The rule gate passes in 65.980s. The full bridge package passes in 69.635s,
checker in 0.209s. The new question assertion passes in 0.042s; its mutant
fails as expected. The filtered Node oracle passes in 14.386s. Go vet passes.
An initial mistaken TestNode filter selected no tests; the actual filtered
oracle command above was then run and its results are preserved.

Three interleaved Go/native full-output runs were measured after tests finished;
every timed pair also matches byte for byte. These are process wall times,
including load, checker questions, native decoding/walk and serialization.

| Population | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 4.020699s | 0.383909s | 10.473 |
| Repository | 0.535719s | 0.129155s | 4.148 |

Native is slower on these populations. One compiler sample reports 527 bridge
queries taking 1.238s, load 0.267s and total run 3.728s; its Go sample reports
load 0.276s, rule callbacks 0.115s and total run 0.156s. Those counters and all
samples are preserved. Raw syntax transfer and native decoding/traversal are
plausible contributors; no causal optimization measurement is claimed.

Shared normal-profile integration remains with codex/lint-harness-dot-a. These
ports run through their own native comparison entry. This batch does not claim
an emitted-JavaScript comparison, CLI suppression/fix application, nondefault
callback options verification, exhaustive regex/TypeScript coverage, all prior
rule suites, or the full repository test gate.
