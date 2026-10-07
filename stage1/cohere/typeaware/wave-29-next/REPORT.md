Built: native no-restricted-globals, no-setter-return and no-shadow-restricted-names, with isolated reference-symbol-origins checker facts.
Commits: c5c8437d claimed and pushed before implementation; original three rules completed and pushed as 0fa8f823.
Commands: 400 configured cases plus frozen 77 compiler and 287 repository roots agree byte for byte; checker and production rule tests pass.
Mutants: three native rule faults and shorthand fact fault compile and exit 0; only comparison catches them; retained released handle fails the required panic check.
Not covered: shared JSX parser support, emitted-JavaScript checker adapter, full repository gate or older language environments.

The new rules stay in this directory. All findings, spans and option-dependent
judgments run in Adamic. The new Go question exports value-reference symbol
identity and ordered declaration path/kind/spans/declaration-file flags, using
the checker's shorthand-value and local-export accessors. It computes no lint
verdict or diagnostic. Its Go and Adamic implementations are separate new files;
only one physical dispatch line was added to `bridge/tsgo/checker/facts.go`.
No shared harness, registration generator, parser or compiler was edited.

The claim scan fetched every origin head and inspected claim Markdown blobs on
every origin branch. Rules already present on main or the bridge branch were
excluded. Other workers had claimed no-invalid-regexp, no-label-var,
no-misleading-character-class, no-obj-calls, no-object-constructor and
no-promise-executor-return. The next three available names were these three,
all with zero default findings in the ranking. Claim c5c8437d was pushed before
any new rule code was written.

`check.py` extracts 385 literal inputs from the unchanged Go production tests,
then adds 15 controls for resolution, descriptors, duplicate custom messages,
shorthand assignment, first declaration identity, UTF-8/UTF-16 spans and CRLF.
Each original environment is reproduced in a separate program using an ambient
`.d.ts` file, with `lib: ["ES2022"]`; unrelated environments never merge.
The rule-local Go oracle builds its own program and calls unmodified production
rules. It imports no bridge code and serializes complete findings, fixes and
suggestions. The native driver uses the public rule APIs and exact canonical
serialization. Configured options use Go's production decoders on the oracle
side and explicit typed rule arguments on the native side. Invalid JSON option
schemas are not independently ported here; they belong to configuration loading.

| Rule | Configured inputs | Findings |
| --- | ---: | ---: |
| no-restricted-globals | 174 | 105 |
| no-setter-return | 148 | 56 |
| no-shadow-restricted-names | 78 | 91 |

The configured streams total 120,087 equal bytes and 252 findings. All fixes and
suggestions are zero, exactly as production. The exact frozen populations from
the original wave are reused: TypeScript v6.0.3 at
050880ce59e30b356b686bd3144efe24f875ebc8, 77 compiler roots, 5,241 bytes;
repository 287 roots, 18,485 bytes. Both have zero default findings. Combined
configured/corpus stream SHA256:
`b498c6a6f1e9c0f48af63c49c4cff2e48885f3ce79f68b35245aba1d3f7b6bb2`.
The compiler and repository populations remain bounded to the branch's existing
parser-compatible manifests; this is not a claim about every source file.

Each native mutant compiles, exits 0 with empty stderr, and changes the stream:
globals inverts source-shadow resolution; setter recognizes a getter in place of
a setter; shadow exempts every undefined declaration rather than proving it safe.
Each is killed on configured group 00. The raw fact mutant uses the ordinary
symbol accessor for a shorthand instead of its value accessor, also compiles,
exits 0 with empty stderr and is killed by the independent Go stream. No mutant
is credited for a compilation failure or sanitizer failure.

All seven configured groups and both corpora agree under AddressSanitizer,
UndefinedBehaviorSanitizer and LeakSanitizer, with empty stderr. Both the native
executable and checker C archive are instrumented. The new question rejects a
released handle with exit 70 and `invalid or released checker handle`, including
under instrumentation. The retained-registry mutant makes the same request exit
0 with empty stderr, so the required panic check catches it. New raw fact tests
cover ambient/plain/shorthand/export reads, exact same-file identity, unresolved
names and invalid questions. All checker tests pass (0.120s), the selected three
production Go rule suites pass (0.101s), and vet has empty output.

One whole-process timing observation after sanitizer builds completed, with no
concurrent build work:

| Corpus | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 1.640439058s | 0.294888579s | 5.56x |
| repository | 0.240453641s | 0.118939054s | 2.02x |

These are single observations, not benchmark medians. Setup earlier in this
continuation passed in 26s: tools ready 0s, submodules ready 0s, cache warm 26s.
`nproc` is 5; `cpu.max` is `400000 100000` (four effective CPUs).

The JSX reference classifications are implemented, but cannot be exercised
through this branch's shared parser. `gaps/jsx.a` compiles successfully and exits
70 on `<Option />` with `parser slice expected GreaterThanToken, got SlashToken
at 35 in input.tsx`. The unchanged Go `JudgesOnlyValueReferences` test passes its
JSX controls. Fixing the shared parser is outside this worker's territory, so it
was left untouched. The shared emitted-JavaScript checker adapter also remains
outside this unit; native comparison is complete on the reported populations.
No additional rules were claimed while these limitations were outstanding.

Reproduction, using the already built stage-0 compiler and frozen manifests
from the original wave (see `../wave-29-configured/REGEXP_REPORT.md`):

```sh
source /workspace/adamic-tools/env.sh
go build -buildmode=c-archive -o /workspace/wave29-next-checker.a \
  ./bridge/tsgo/archive > /tmp/wave29-next-checker.log 2>&1
python3 stage1/cohere/typeaware/wave-29-next/check.py \
  /workspace/wave29-next-validation > /tmp/wave29-next-validation.log 2>&1
python3 stage1/cohere/typeaware/wave-29-next/fact_checks.py \
  /workspace/wave29-next-facts /workspace/wave29-next-validation \
  > /tmp/wave29-next-facts.log 2>&1
go test ./bridge/tsgo/checker -count=1 > /tmp/wave29-next-checker-tests.log 2>&1
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware > /tmp/wave29-next-vet.log 2>&1
(cd cohere && go test ./internal/lint/rules/core \
  -run '^Test(NoRestrictedGlobals|NoSetterReturn|NoShadowRestrictedNames)' -count=1) \
  > /tmp/wave29-next-go-rules.log 2>&1
/workspace/wave29-regex-controls/adamic build \
  stage1/cohere/typeaware/wave-29-next/gaps/jsx.a -o /workspace/wave29-next-jsx \
  > /tmp/wave29-next-jsx-build.log 2>&1
/workspace/wave29-next-jsx > /tmp/wave29-next-jsx.stdout 2> /tmp/wave29-next-jsx.stderr
# Expected final command exit: 70.
```

Complete command/timing records, fixture inputs and compressed canonical streams
are preserved in `validation/`; the local full build artifacts remain under
`/workspace/wave29-next-validation` and `/workspace/wave29-next-facts`.
