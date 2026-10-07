Built prefer-promise-reject-errors, prefer-regex-literals and prefer-rest-params as native .a rules.
Claim e1597607 was pushed before code; implementation commit is e062ed93 on codex/typeaware-wave-18.
Targeted agreement passed: 419 controls, 77 compiler files and 287 repository files; sanitizers, checker, Node and vet passed.
Promise/rest span mutants, regex replacement and grammar mutants were caught by byte comparison; released-registry mutant was caught by the handle guard.
Full gate and JavaScript driver comparison were not run; native is slower than Go on both measured corpora.

The previous batch was already pushed through c6d9d07f before this selection. The audit fetched
356 origin refs, checked 33 distinct claim Markdown blobs and 132 claimed names, and excluded
25 ranked baseline ports on main ef3d907e and tsgo-c-library 5afbdb83. These three were the
first remaining by-volume names, each with zero recorded corpus findings. The complete selection
record is validation-wave-18-preference/selection.json. No further rules were claimed this turn.

Each rule owns a separate .a file. Promise resolution requires every declaration to be ambient;
its executor follows checker symbol identities through nested functions, duplicate parameters and
parameter defaults. Bare global undefined remains a separate predicate, and allowEmptyReject
exempts only empty calls. The rest rule requires a resolved symbol with zero declarations, while
an immediate dotted receiver remains exempt. Neither rule proposes fixes or suggestions.

The regex rule runs its reference tracker and every lint decision natively: aliases, global-object
paths, destructuring, conditional/logical propagation, writes, static argument shapes, comment
and preceding-token gates, flag validation, balance, printable policy, literal rewrites, padding,
messages, spans and ordered suggestions. The bridge question regex-pattern-facts supplies raw
regex grammar structure: scanner success, character spans, escape ends and class ends. Its isolated
Go file copies the pinned cohere regexsyntax/regexpattern shelves, with prefixed names, and imports
no lint rules. The question contains no finding, fix, suggestion or lint-policy decision. It takes
canonical decimal UTF-16 units so NUL and newline patterns cross the existing C string boundary;
returned offsets are UTF-16 units, converted to source UTF-8 only at diagnostic/fix boundaries.
The only shared source edit is one case-registration line in bridge/tsgo/checker/facts.go.
The native decoder is regex_pattern_facts.a; shared registration generation and test harnesses
were not edited.

A first character-span representation stalled inside internal/fresh.ProveWrites during native
compilation. Replacing object spans with scalar start/end arrays resolved it entirely within the
rule files. The .a formatter CLI currently skips these files, so a private scratch adapter called
cohere's existing TypeScript printer on their text and wrote the same .a paths. No .ts source files
were created and no compiler, formatter or shared harness source was changed. All targeted
agreement and mutants ran again after formatting, passing in 75.901s.

The oracle loads an independent Go program and calls the unmodified production rules at cohere
715ba94f3608a6500086b1076ce5cb7e51b836db. Canonical output compares complete diagnostics,
UTF-8 spans, messages, fix counts/text and each ordered suggestion and edit. Inputs include the
verbatim Go rule tables plus alias, shadow, global-write, option, Unicode, NUL and trivia controls.
Source hashes and both corpus manifests are saved with the evidence. TypeScript is v6.0.3 at
050880ce59e30b356b686bd3144efe24f875ebc8.

| Input/configuration | Findings | Identical bytes | Suggestions |
| --- | ---: | ---: | ---: |
| 419 controls, defaults | 288 | 163057 | 184 |
| 419 controls, allowEmptyReject | 285 | 162227 | 184 |
| 419 controls, disallowRedundantWrapping | 312 | 176024 | 211 |
| 419 controls, both options, ASan/UBSan | 309 | 175194 | 211 |
| Controls, no library | 205 | 134246 | compared in full |
| Controls, custom ambient globals | 288 | 163057 | 184 |
| Compiler corpus, 77 files | 0 | 5318 | 0 |
| Repository corpus, 287 files | 0 | 18485 | 0 |

Defaults contain 61 Promise findings, 15 rest findings and 212 regex findings. All automatic fix
counts are zero. Normal and sanitized runs match for default controls, no-library and custom-ambient
controls, both corpora, and the combined nondefault options. Sanitized runs exit zero with empty
stderr. Checker tests pin astral UTF-16 offsets, multiline framing and malformed numeric payloads.

Every rule mutant compiles with the normal native warnings-as-errors gate, runs with exit zero and
empty stderr, and preserves 288 findings. Only complete-byte comparison kills it:

| Mutant | Detection |
| --- | --- |
| Promise reports its callee instead of the complete rejection call | first different byte 1983 |
| Rest reports the parent instead of the arguments identifier | first different byte 23425 |
| Regex appends x to its suggestion replacement | first different byte 34968 |
| Raw regex scanner-success bit inverted | first different byte 34880; 288 findings but 3 rather than 184 suggestions |
| Released handle retained in the Go registry | guard expects exit 70 and exact invalid-or-released-handle panic; mutant exits zero and returns facts |

The live handle probe constructs and releases a real checker program before asking the new
encoded grammar question. The original registry rejects the stale handle with exit 70 and
`adamic: panic: invalid or released checker handle`. The deliberately retained registry is the
only mutation and incorrectly succeeds; the test catches it.

Quiet whole-process timings use three interleaved rounds per corpus after validation, with full
stdout/stderr retained. Medians are:

| Corpus | Native seconds | Go seconds | Native/Go |
| --- | ---: | ---: | ---: |
| Repository, 287 files | 0.215997 | 0.122187 | 1.77 |
| Compiler, 77 files | 1.767094 | 0.368887 | 4.79 |

These measure load, parse and lint count mode; both corpora have zero findings for these rules.
They establish neither a positive-corpus speed result nor native speed parity. Measurements and
query/load breakdowns are in validation-wave-18-preference/measurements.json and quiet logs.
Setup reports go 0s, clang 0s, node 0s, submodules 0s, build-cache warm 20s, total 20s;
nproc is 5, cpu.max is 400000 100000, and RAM is 17.6 GB.

Commands and observed outputs:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE18_PREFERENCE_ARTIFACTS=/workspace/wave-18-preference-validation-final \
ADAMIC_WAVE18_COMPILER_MANIFEST=/tmp/wave-18-compiler.manifest \
ADAMIC_WAVE18_REPOSITORY_MANIFEST=/tmp/wave-18-repository.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-18-typescript TMPDIR=/tmp/adamic-gate \
go test ./stage1/cohere/typeaware -run '^TestWave18PreferenceAgreementAndMutants$' \
  -count=1 -timeout=30m -v > /tmp/wave-18-preference-final-test.log 2>&1
# PASS, 75.901s

go test ./bridge/tsgo/checker -count=1 -timeout=10m -v \
  > /tmp/wave-18-preference-checker.log 2>&1
# PASS, 0.114s

# From the pinned cohere checkout:
go test ./internal/lint/rules/core -run '^TestPrefer(PromiseRejectErrors|RegexLiterals|RestParams)' \
  -count=1 -timeout=20m -v > /tmp/wave-18-preference-production.log 2>&1
# PASS, 0.340s

# From Adamic:
TMPDIR=/tmp/adamic-gate go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw)\.a$' \
  -count=1 -timeout=15m -v > /tmp/wave-18-preference-node.log 2>&1
# PASS, 11.727s; one-byte oracle mutant and five native/JavaScript/Node fixtures

go vet ./... > /tmp/wave-18-preference-vet.log 2>&1
# Exit 0, empty log
```

The full repository test gate and this C-bridge driver's JavaScript backend were not run. This
remains canonical diagnostic comparison rather than a complete CLI suppression/configuration
or suggestion-application engine. Regex grammar remains in the isolated foreign Go scanner,
while rule policy and output are native; the existing Go rule's structural-parser quirks are held,
not corrected. The earlier title rule's shared native JSX-parser integration gap remains as
recorded in WAVE_18_TITLE_REPORT.md. No current-rule shared harness gap blocks these three ports.
No pull request was opened.
