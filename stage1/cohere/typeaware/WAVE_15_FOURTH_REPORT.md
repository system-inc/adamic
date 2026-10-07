Built: no-throw-literal and prefer-arrow-callback, plus partial no-useless-backreference scanner/judgments, all new Adamic sources .a.
Commits: claim 90d8706a pushed before code; implementation c0c3445b; prior work pushed through 172d8eba.
Checks: differential PASS 123.244s; checker PASS 0.255s; sanitizer/released-handle checks pass; vet and diff check empty; setup 27s, nproc 5.
Mutants: throw end +1 caught byte 55; arrow replacement => to =x caught byte 330; backreference direction message caught byte 259; all compiled and exited 0.
Not covered: complete backreference corpus agreement blocked by shared regex/reference helpers; full repository gate, emitted JavaScript and complete option matrices not run.

Fetched all 348 origin refs before selection. Thirty-three claim documents
mention 126 of the 197 ranked checker rules. The first remaining names were
no-throw-literal, no-useless-backreference and prefer-arrow-callback, each with
combined volume zero. Matches on main and the bridge branch were count/inventory
records only. The claim was committed and pushed before creating code. Prior
partial regex ports remain explicitly blocked as documented in earlier reports.
No additional rules are reserved in this round.

NoThrowLiteral reproduces the production optimistic syntactic couldBeError
judgment, binary operator splits, conditional branches and whole-throw spans.
Bare undefined uses raw declaration-file facts, treating missing symbols and
zero declarations as global and source declarations as shadowing. Parentheses
are kept exactly as production Go judges its AST. Fourteen controls produce
8 findings and 3405 identical bytes, including Unicode offsets and global versus
parameter-shadowed undefined. Both 77-file TypeScript compiler and frozen
287-file repository populations agree with Go: zero findings, respectively
5318 and 18485 complete canonical bytes, normal and ASan builds.

PreferArrowCallback implements function ownership, inherited arrow captures,
resolved own-name references, implicit versus declared arguments, generator/
super/new.target exclusions, callback/bind traversal, default options and every
production fix decline. Its repairs preserve comments, omit unsafe rewrites,
remove bind(this), insert arrows and wrap operator operands as production does.
Twenty controls produce 14 findings and 4909 identical bytes, with full fix
triples and suggestion fields. Controls cover duplicate parameters, this
parameters, async newlines, bind comments/wrappers and nested captures.
Compiler and repository canonical bytes agree in normal and ASan builds, both
zero findings on the unchanged frozen populations.

The existing node-symbol-details serializer panicked while rendering a computed
property declaration name in the compiler corpus. That initial failure is
retained. Instead of changing shared serialization, callback-symbol supplies
only raw symbol identity and declaration count in new Go and .a files, with
one dispatcher registration case. Its direct test checks computed-name input,
identity stability, nil symbols, counts, wrong kinds and question suffixes.
The arrow rule's released-handle probe uses the new question and requires
panic 70 with invalid or released checker handle. No bridge answer is a lint
verdict or repair. Protected compiler files, shared registration generators
and existing harness files were not edited; new tests use existing helpers.

NoUselessBackreference ports group/alternative/lookaround path scanning,
numeric and named reference resolution, duplicate named-group aggregation,
well-formedness judgments, all five classifications and exact descriptions.
Sixteen direct regex controls produce 9 findings and 3532 identical bytes
against unchanged production Go callbacks, including every diagnostic kind,
reachable references, invalid Unicode references and duplicate group names.
Normal and sanitizer controls agree. The partial scanner explicitly refuses
character classes and non-reference escapes, pending shared RegexSyntax.ClassEnd
and RegexSyntax.SkipPatternEscape. Global RegExp constructors and constant
bindings require ReferenceTracker and ConstantStringIn, also absent on the
fetched helper branches. Only readiness/inventory records mention these helpers.
The class, escape and constructor probes each exit 70 with the named missing
helper. Both full corpus attempts exit 70 for missing ReferenceTracker and
ConstantStringIn; they are not reported as successful comparisons or zero
findings. The rule is therefore a partial port with complete independent
judgments, not a full regex frontend or complete corpus port.

Every rule mutant builds successfully and exits zero with empty stderr. Only
the independent production Go byte comparison catches the changed diagnostic
or fix bytes. Tests compare complete file headings, rule IDs, message IDs/text,
UTF-8 spans, fixes, suggestions and counts, preserving duplicate findings.
ASan/UBSan/LeakSanitizer controls and supported corpus runs have empty native
stderr. The Go heap itself is not instrumented by ASan.

Single sequential wall-clock observations, not benchmark medians:

| Rule/population | Native | Go |
| --- | ---: | ---: |
| throw, repository | 198.463ms | 122.638ms |
| throw, compiler | 1.290639s | 291.001ms |
| arrow, repository | 244.905ms | 116.682ms |
| arrow, compiler | 2.467247s | 364.521ms |
| backreference, positive controls only | 14.810ms | 20.397ms |

Compiler is TypeScript v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8;
repository roots remain the frozen 212 .a / 75 .ts population. Cohere and tsgo
pins remain unchanged. Setup output: Go 1.27.1 0s; clang 20.1.8 0s;
Node v24.19.0 0s; submodules 0s; cache warm 27s; setup done 27s;
nproc 5, cgroup quota 400000/100000. New authored Adamic files use .a.

Commands, with all test output redirected to retained logs:

```
bash cloud/setup.sh > /workspace/wave15-fourth-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE15_FOURTH_ARTIFACTS=/workspace/wave15-fourth-final-3 ADAMIC_WAVE15_ARROW_ARTIFACTS=/workspace/wave15-arrow-final-3 ADAMIC_WAVE15_BACKREFERENCE_ARTIFACTS=/workspace/wave15-backreference-final-3 ADAMIC_WAVE15_COMPILER_MANIFEST=/workspace/wave-15-compiler.manifest ADAMIC_WAVE15_REPOSITORY_MANIFEST=/workspace/wave-15-repository.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-15-typescript go test -v -count=1 -timeout=15m ./stage1/cohere/typeaware -run '^TestWave15Fourth(Throw|Arrow|Backreference)AgreementAndMutants$' > /workspace/wave15-fourth-final-3.log 2>&1
go test -v -count=1 ./bridge/tsgo/checker > /workspace/wave15-fourth-checker-final.log 2>&1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > /workspace/wave15-fourth-vet.log 2>&1
git diff --check > /workspace/wave15-fourth-diff.log 2>&1
```

The cohere stdin formatter was run over the eight new .a modules using virtual
.ts paths, before the final source gates; its stderr log is empty per module.
Evidence logs and compressed Go/native/ASan streams are in
validation-wave-15-fourth, with SHA256SUMS. The full go test ./... gate and
JavaScript backend comparison were not run. Shared integration gaps are
reported explicitly; no silent fallback or shared-file workaround was added.
