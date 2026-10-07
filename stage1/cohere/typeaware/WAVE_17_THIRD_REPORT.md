Built: native `.a` ports of no-global-assign, no-implicit-globals and no-implied-eval.<br>
Commits: claim bc581497; implementation a7306edbcde675f7038002cd281e669c5ecae3a2; prior Head decisions 8f3d8d05.<br>
Commands/output: final differential PASS 147.198s; checker PASS 0.110s; filtered Node oracle PASS 11.375s; go vet ./... exits 0.<br>
Mutants: all three rule mutants and both raw-fact mutants caught only by Go bytes; retained released handles caught by required panic 70.<br>
Not covered: full repository gate; original end-to-end JSX still refused; existing shared unknown-request mutation anchor remains broken.

# Three global-rule ports

The previous claimed work and explicit JSX boundary were pushed before the next
claim. The fresh origin scan inspected 335 refs and 33 distinct Markdown claim
blobs, plus native stage1 sources on main and the library branch. The first
eligible checker-dependent entries in the combined VOLUME_REPORT.md ranking
were ranks 137, 138 and 139, all zero volume. selection.json records the refs,
claim blobs, existing port matches and full ranking. bc581497 was pushed before
any implementation. All work continues on codex/typeaware-wave-17. No PR opened.

Each rule lives in its own Adamic file. The own-directory wave_17_third_suite.a
loads one checker program and uses the existing native parser, parent links and
binding/write helpers. Its three registrations do not change the shared generator
or harness. Findings include the production rule names, ids, messages and byte
ranges, plus every serialized fix and suggestion field. These production rules
have no fixes or suggestions, so those fields are empty on both sides.

no-global-assign distinguishes binding writes from member writes and reads,
handles nested destructuring/rest/defaults, and asks the correct shorthand value
symbol before applying the first declaration-file test. It supports exceptions.
no-implicit-globals uses the checker program's actual module indicator for global
declarations and combines compiler options, JavaScript status, class ancestry and
strict directive prologues for leaks. It supports lexicalBindings, recursive
binding patterns and the production leak-target forms. no-implied-eval recognizes
the three production functions, global receiver chains, static property names,
parentheses and syntactic string concatenation/templates; it preserves unresolved
name and merged local declaration behavior. It follows Go's existing boundaries,
including no constant-expression evaluator.

The two raw questions have dedicated Go and `.a` files. global-source-facts returns
module/JavaScript booleans and raw strict/alwaysStrict tristates. global-binding-facts
returns ordinary, local merged and shorthand value symbols through the existing
raw declaration serializer. Every lint judgment stays native. facts.go has one
changed dispatch line; dispatch proceeds through the two new files to the existing
ancestry question. Protected compiler files and all shared harness files are untouched.

## Complete differential evidence

The independent oracle is an overlay inside pinned cohere, uses its own loader
and traversal, and invokes the unchanged production registry rules. It imports no
bridge code. Sources are extracted from upstream tests without expected judgments,
with module controls and separate targeted TypeScript/JavaScript scripts. All 315
candidate controls parse under Go; none was filtered out. JavaScript fixture paths
are symlinks to `.a` source, preserving the checker's JavaScript status.

| Population | Roots | Findings | Identical bytes, normal and sanitized |
| --- | ---: | ---: | ---: |
| Upstream and targeted controls | 315 | 159 | 59,641 |
| Frozen repository | 287 | 0 | 18,485 |
| TypeScript src/compiler | 77 | 0 | 5,318 |

Control findings by rule are recorded in control-counts.json. Unicode/CRLF,
shorthand writes and defaults, shadowed globals, nested receiver chains, templates,
type arguments, strict directives, classes, module versus script scope and JS versus
TS strictness are positive or negative controls. Options also match under sanitizers:
lexicalBindings: 62,043 bytes; Object/String exceptions: 52,903 bytes; lexicalBindings
plus Array exception: 61,337 bytes. Program-wide declarations are resolved by both
independent programs over the same manifest, rather than borrowing upstream expected
verdicts from separately configured single-file tests.

TypeScript remains v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8. Cohere remains
715ba94f3608a6500086b1076ce5cb7e51b836db. Configured declaration roots are retained.
The manifests are the frozen validation-coverage populations; additions in this unit
are not substituted into that corpus. Sanitizers cover native C and ownership, not
Go's heap. All normal/sanitized native comparison streams have empty stderr.

| Mutant | Clean runtime observation | What catches it |
| --- | --- | --- |
| Global assignment ignores declaration-file guard | exit 0, empty sanitizer stderr | Go bytes, 12,775 |
| Implicit globals ignores module declaration gate | exit 0, empty sanitizer stderr | Go bytes, 44 |
| Implied eval accepts source shadows | exit 0, empty sanitizer stderr | Go bytes, 44 |
| Source question marks JavaScript as TypeScript | exit 0, empty sanitizer stderr | Go bytes, 52,741 |
| Binding question uses property symbol for shorthand value | exit 0, empty sanitizer stderr | Go bytes, 4,468 |
| Registry retains released program, binding question | exit 0, empty stderr | required panic 70 |
| Registry retains released program, source question | exit 0, empty stderr | required panic 70 |

Both unmutated new questions reject released handles with exactly
`adamic: panic: invalid or released checker handle` and exit 70. Mutants are scratch
native source variants or Go overlays and do not alter shared repository sources.
The dedicated Go metadata tests cover module/script/JS facts, symbol presence and
malformed requests. The filtered external Node gate proves its one-byte mutant.

## Commands and observations

Setup: Go/clang/Node/submodules ready 0s; build cache warm 25s; done 25s.
nproc: 5; cgroup quota remains four CPUs. Source /workspace/adamic-tools/env.sh.
All commands write output directly to logs:

```sh
ADAMIC_WAVE17_THIRD_ARTIFACTS=/workspace/wave-17-third \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-17-typescript \
TMPDIR=/workspace/wave-17-artifacts \
go test ./stage1/cohere/typeaware \
  -run '^TestWave17ThirdAgreementAndMutants$' -count=1 -v -timeout=20m \
  > /workspace/wave-17-third-final.log 2>&1
go test ./bridge/tsgo/checker -count=1 -v \
  > /workspace/wave-17-third-checker.log 2>&1
go vet ./... > /workspace/wave-17-third-vet.log 2>&1
TMPDIR=/workspace/wave-17-artifacts go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(functions|closures|generic_functions)\.a$' \
  -count=1 -v -timeout=10m > /workspace/wave-17-third-node.log 2>&1
python3 stage1/cohere/typeaware/validation-wave-17-third/measure.py \
  /workspace/wave-17-third stage1/cohere/typeaware/validation-wave-17-third \
  /workspace/wave-17-typescript > /workspace/wave-17-third-measure.log 2>&1
```

The initial differential run passed before option/raw-fact extensions. The final
run passed all those extensions after formatting through virtual `.ts` stdin paths;
no Adamic `.ts` source was created. The standalone build probe initially used an
incorrect CLI argument order and printed usage; the corrected build succeeded.

Three quiet alternating rounds include program loading, native parsing, traversal,
queries, output and teardown. Every timed pair has identical full-stream SHA-256.

| Corpus | Native median | Go median | Native / Go | Native queries |
| --- | ---: | ---: | ---: | ---: |
| Compiler | 2.097368s | 0.428167s | 4.90x | 6,198 |
| Repository | 0.291933s | 0.126645s | 2.31x | 1,424 |

Native is slower in these observations. Measurements.json preserves all rounds and
phase counters. These are zero-finding corpus timings, not a positive-control speed
claim. Compressed raw outputs, logs, manifests, input hashes and diagnostic hashes
are in validation-wave-17-third.

## Outstanding integration limits and prior claimed work

The original Head rules now have complete native decision implementations and
40 projected JSX controls with 21 findings and 10,044 identical bytes. Final Head
comparison PASS 46.212s; updated production regression PASS 70.166s. Duplicate count
and Script spelling mutants are caught by Go bytes at 46 and 456. That regression
also catches the async range mutant at byte 52 and the retained-handle mutant by the
required panic. Positive production JSX still refuses with parser panic 70; raw AST
projection is validation only. WAVE_17_HEADS_REPORT.md explains this boundary.

The already documented TestInspectRequestRefusals mutation anchor still hardcodes
the base bridge's former unsupported-question return line. It cannot construct its
mutant after dedicated question dispatch is added. Actual malformed and unknown
question refusals pass. This known shared-harness gap was not edited or hidden, and
the full repository gate was not claimed. Full pinned CLI lint of `.a` and shared
emitted-JavaScript lint comparison were not run. No new rule was blocked by those
gaps. The prior Nexus ports and their evidence remain in WAVE_17_NEXT_REPORT.md.
