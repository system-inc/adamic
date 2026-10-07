Built: no-label-var and partial no-invalid-regexp/no-misleading-character-class judgments in .a.
Commits: claim 58abbd19; implementation ea2a3879; prior continuation pushed through b13ded14.
Checks: differential PASS 80.127s; scope/checker PASS 0.286s; vet and diff check clean; sanitizer and released-handle checks pass.
Mutants: label span +1 caught byte 52; duplicate-to-unknown flag message caught byte 124; suggestion u-to-v caught byte 2626, all compiled and exited 0.
Not covered: full regex parsing/reference/constant/cooked-offset helpers, Unicode flag quoting, undefined keyword labels, emitted-JavaScript comparison.

The third claim was pushed before implementation after fetching all 335 origin
heads and excluding existing implementations and claims. No further rules are
claimed. The prior HTTPS push failures were resolved; b13ded14 is on origin.

NoLabelVar uses raw checker GetSymbolsInScope(Value) names and flags, with an
exact LabeledStatement question in its own Go and .a files. Shared facts.go has
only its dispatcher case. The direct question test compares names and flags
without assuming checker iteration order. Findings match production Go byte
for byte on 15 controls (10 findings), the frozen TypeScript compiler corpus
(77 files, zero findings), and repository corpus (287 files, zero findings),
including full canonical fixes and suggestions. Native and ASan runs agree.
Native wall time versus Go: repository 226.581ms / 157.239ms; compiler
1.535142s / 337.161ms. These are single wall-clock measurements, not benchmarks.
The shared parser refuses `undefined:` with `parser slice expected semicolon
at 38`; the failed initial probe is not counted as a passing comparison.

NoInvalidRegexp implements global constructor resolution, shadowing exclusion,
argument selection, duplicate/unknown/u-v flag priority and allowed extra
flags. String patterns requiring ECMAScript validation explicitly panic rather
than silently pass. The refusal probe exits 70 with
`no-invalid-regexp requires the shared ECMAScript pattern validator`.
Non-ASCII flag quoting also explicitly refuses pending a shared Unicode rune
quoting helper. These are partial ports, not full corpus agreement claims.

NoMisleadingCharacterClass implements all six sequence judgment messages and
Unicode-flag suggestions. Its control-only frontend supports bare, unescaped,
non-range character-class literals. The production run entry explicitly
refuses because shared RegexSyntax, ReferenceTracker, ConstantStringIn and
CookedToRaw helpers are absent. General escapes, ranges, constructors, constant
references and source mapping therefore remain unintegrated. Shared parser,
registration generator and existing test harness files were not edited.

Ten regex controls produce 10 findings and 3390 identical serialized bytes
against independent production Go callbacks, including suggestions, in normal
and sanitizer runs. This proves the tested subset, not a general regex parser.
Each mutant compiled successfully, exited zero with empty stderr, and only
the independent byte comparison rejected its result. Released checker handles
are rejected with panic 70. New tests use existing harness helpers.

Commands (outputs retained in validation-wave-15-core):

```
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE15_CORE_ARTIFACTS=/workspace/wave15-core-final ADAMIC_WAVE15_COMPILER_MANIFEST=/workspace/wave-15-compiler.manifest ADAMIC_WAVE15_REPOSITORY_MANIFEST=/workspace/wave-15-repository.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-15-typescript ADAMIC_WAVE15_REGEX_ARTIFACTS=/workspace/wave15-regex-final go test -v -count=1 -timeout=15m ./stage1/cohere/typeaware -run '^TestWave15(CoreAgreementAndMutants|RegexPartialControlsAndMutants)$' > /workspace/wave15-third-final.log 2>&1
go test -v -count=1 ./bridge/tsgo/checker > /workspace/wave15-scope-final.log 2>&1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > /workspace/wave15-third-vet.log 2>&1
git diff --check > /workspace/wave15-third-diff.log 2>&1
```

Toolchain setup previously passed in 30s (Go/clang/Node/submodules 0s, warm
cache 30s); nproc is 5, CPU quota 4. The same cached toolchain was used here.
Full repository gate and emitted-JavaScript comparison were not run. Regex
rules cannot yet satisfy the full requested corpus bar, so work stops at the
explicit shared integration gaps with partial implementations committed.
