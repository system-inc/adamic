Built: native no-obj-calls, no-object-constructor and no-promise-executor-return, including suggestions; three object-constructor inputs remain blocked by the shared parser.
Commits: earlier batch 7d1f1fcc and 880f6f6f pushed; next claim 84c0240f pushed before implementation; implementation 7f360051 pushed.
Commands/output: core test PASS 81.533s, 242 default and 48 allowVoid findings byte-identical; 77 compiler and 287 repository roots identical; checker PASS 0.109s; Node PASS 17.118s; vet/gofmt clean.
Mutants: namespace-message judgment byte 398; missing object-suggestion semicolon byte 44075; wrong promise-braces judgment byte 74316; released registry and request-suffix mutants caught by their explicit assertions.
Not covered: three parser-blocked object-constructor rows, full repository gate, shared profile/JavaScript integration, and native JSON option decoding; no additional claims after this batch.

Selection and claim

After the previous three reservations were ported, validated and pushed, all origin
heads were fetched. The audit covered 347 refs and 33 Markdown claim documents.
The main and bridge tips remain ef3d907e and 5afbdb83, with 26 native rules.
The only explicit reservation releases were promises, spread and lost-update;
all three now have active wave-22 claims, so they were excluded. The next available
rules were full-ranking positions 146, 147 and 148, all zero-volume ties in lexical
order: no-obj-calls, no-object-constructor and no-promise-executor-return.
The complete audit and release decision are in
validation-wave-21-process/next-selection.json.xz. Claim 84c0240f was pushed
before any rule code was written.

Implementation

Each rule has its own .a file and runs in a private native driver. No shared
registration generator, test harness, parser or protected compiler file was edited.
One new raw checker question lives in reference_node.go and reference_node.a;
its only shared-file edit is a dispatch arm in facts.go (two Go lines).
The question exposes declaration-name and write-access syntax flags, the symbol
read at a name (including shorthand and local export positions), and its raw
declaration-file flags. Native code decides every global predicate, alias trace,
report range and suggestion. The namespace tracker follows assignment/default/
destructuring aliases, pass-through expressions, constant computed names and cycles,
and suppresses any global the file writes. Constructor rules intentionally use
Go's different predicate: the first declaration must be in a declaration file.
Object suggestions preserve parentheses and automatic-semicolon-insertion cases.
Promise suggestions preserve unwrapped report spans and outer-body edit anchors;
both default allowVoid=false and allowVoid=true are compared.

Validation

The independent Go oracle imports unchanged production registry rules and no bridge
implementation. It walks Go's AST and serializes complete findings, fixes and
suggestions. Native output matches all fields, not just counts or message IDs.
The private fixture extractor reads the upstream tables: 11 no-obj-calls unit rows,
163 namespace corpus rows, 59 object-constructor rows and 183 promise-executor rows,
plus 14 independent binding/Unicode/CRLF controls. Duplicate upstream rows remain
in the population. Rows carrying allowVoid=true are run under that option; the
others are run under defaults. Three parser-blocked rows are tested separately.
The supported defaults produce 242 findings and 114,046 identical bytes; allowVoid
produces 48 findings and 33,063 identical bytes, including all suggestion repairs.
ASan, UBSan and LeakSanitizer binaries produce identical outputs and empty stderr.

Both established populations pass normally and under sanitizers: 77 TypeScript
compiler roots at 050880ce59e30b356b686bd3144efe24f875ebc8 and 287 frozen repository
roots. Both have zero findings for this batch; complete outputs including file
headings are respectively 5,087 and 18,485 identical bytes.

All three rule mutants compile, exit 0 with empty stderr, and fail only comparison:
namespace direct/alias message judgment reversed, object replacement omitting its
required preceding semicolon, and promise brace-eligibility judgment reversed.
The released-handle probe exits 70 with exactly
`adamic: panic: invalid or released checker handle`; retaining its registry entry
makes it exit 0 and the assertion catches that mutant. A suffix-guard mutant
compiles and fails TestReferenceNodeFacts with `malformed suffix accepted`.
Live raw facts verify library/source declarations, writes, shorthand identity and
local-export identity. The filtered Node oracle passes nine source/native/sanitizer/
emitted-JavaScript fixtures and its independent one-byte comparison mutant.

Observed end-to-end quiet wall times, including program loading and native parsing:

| Population | Native | Go cohere |
| --- | ---: | ---: |
| Compiler 77 roots | 2.934743 s | 0.353567 s |
| Repository 287 roots | 0.336652 s | 0.109850 s |

Native is slower on both populations; these are single-run observations.

Exact shared-parser gaps

The object-constructor judgment and suggestion logic is ported, but three upstream
inputs cannot reach it through the current shared native parser. Each yields one
Go finding, while native exits 70 before linting. They are not counted as byte
agreement or silently filtered. Sources are preserved in parser-gaps.json and
both Go and native outputs are archived:

- `<foo />` followed by `Object()`: `parser slice expected GreaterThanToken, got SlashToken at 5`.
- `<foo></foo>` followed by `Object()`: `parser slice expected GreaterThanToken, got Identifier at 7`.
- A `yield:` label before a loop holding `new Object()`: `parser slice expected semicolon at 37`.

The refusals come from the shared TypeScript parser, not missing checker facts or
suggestion serialization. Its files are outside this unit's territory, so they
were left untouched. The class contains the complete rule logic for these cases
once their trees can be parsed. Object-constructor coverage remains partial on
these three rows. Shared profile registration and emitted-JavaScript checker
comparison remain integration work for codex/lint-harness-dot-a. The private driver
exposes the allowVoid boolean but does not implement a native configuration JSON
layer. No further rules were claimed.

Commands (all test output redirected directly to log files):

```
ADAMIC_WAVE21_CORE_ARTIFACTS=/workspace/wave21-core-final ADAMIC_WAVE21_COMPILER_CONFIG=/workspace/wave21-compiler/src/compiler/tsconfig.json ADAMIC_WAVE21_COMPILER_MANIFEST=/workspace/wave21-compiler.manifest ADAMIC_WAVE21_REPOSITORY_MANIFEST=/workspace/wave21-repository.manifest go test ./stage1/cohere/typeaware -run '^TestWave21CoreRules$' -count=1 -v -timeout=10m
go test ./bridge/tsgo/checker -count=1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures|method_closures|generic_functions|number_parsing)\.a$' -count=1 -timeout=10m -v
```

Original cloud setup timing: total 81 seconds, tools/submodules ready 0 seconds,
cache warmup 81 seconds; nproc 5. Tool environment:
/workspace/adamic-tools/env.sh. This continuation reused that installed toolchain.
Complete validation artifacts are in validation-wave-21-core with SHA256SUMS.
