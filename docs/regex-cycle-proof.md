# Regex cycle proof

## Cause and scope

Main base: `d799ede40af1d947efe8c4a897efbb90c2ef2264`.
CSS dependency: `e8e03a8df4179c90b44e1bf93218588759769ae0`.

The failure is not a missing edge rule in `internal/lower/cycles.go`.
`internal/lower/fresh.go` asks `fresh.ProveWrites` to judge writes into recursive
slots. An unrecognized IR expression produces `WriteUnknown`, which refuses
**every** cycle-capable slot, regardless of the unknown expression's location.
`ir.RegExpCall` in selector parsing consequently poisons the unrelated
`ValueTree.nodes` and `MediaNode.nodes` constructor writes.

The change teaches `internal/fresh` the regex IR operations. No lowering cycle
rule is weakened. Unknown operations still refuse everything they could affect.

Current main contains the Go regex matcher, but not native regex lowering and
emission. `internal/ir/regexp.go` is copied verbatim from the CSS branch so this
small proof change builds on main, without importing its native regex runtime,
emitter or lowerer. The same file merges without a conflict. The source fixture
is registered with the ordinary oracle when the dependency's `Program.Regexps`
field exists. On main its registration test explicitly skips; the IR proof tests
run. On the CSS scratch merge the source fixture runs against Node, native,
the JavaScript backend, ASan, UBSan and LeakSanitizer.

## Proof boundary

A compiled regex holds static bytecode, immutable source/flags strings and
scalar state. It cannot retain a user object that points back to its holder.
Construction evaluates arguments, then makes a fresh object with no tracked
children. Supported native methods can update lastIndex or iterator scalar state;
they cannot store user references. Lowering already excludes replacement
callbacks and overriding regex/iterator properties.

Every receiver and argument is interpreted in evaluation order, including writes,
throws and escapes in those expressions. A regex call is not an opaque pure
expression. Primitive/string results contain no cycle-capable references.

Result arrays and iterator results are **fresh containers**, not reference-free
objects. Their children are conservatively modeled as outside: nested groups,
indices and match arrays must not vanish from alias analysis. Mutable metadata
and named-group reads are outside values too. This can refuse some safe programs,
but cannot prove a write by forgetting a returned alias. Ordinary object fields,
array elements, callbacks and future regex methods retain their conservative
rules. Null has no references; testing it still interprets its operand.

## Fixtures and mutants

`internal/fresh/regexp_test.go` checks every supported operation alongside a tree
write, cycle-closing writes inside regex operands, and refusal of a future method.
`internal/fresh/testdata/regexp_tree.ts` composes recursive tree construction with
regex test, replacement, exec, named groups, indices, match, split, search,
replaceAll and matchAll iteration.

The unsafe mutant changes ordinary `ir.Property` loads to evaluate the receiver
and return no tracked references, applying the regex scalar rule to a
cycle-capable user field. It compiles and accepts `fresh_refused/spread_old.a`.
Node and native both print `1` and exit 0; LeakSanitizer alone exposes the false
proof: **205 bytes leaked in 4 allocations**. The refusal test fails with exit 1.
The production proof is restored afterward.

## Scratch integration

The CSS branch predates main's added `adamic_shape.methods` field. Its five
runtime shape initializers and generated named-group shape initializer need
named fields rather than positional initializers to compile under `-Werror`.
This compatibility adjustment is confined to the scratch merge. It is not part
of the compiler proof change.

The scratch CSS test enables native, JS backend and leak comparisons on the
existing complete Go composition corpus. The formerly refused two-slice program
is checked against `Parsed\nOk\n` on all three backends. The old refusal assertions
must be replaced when the branches are integrated. The retained scratch patch
contains these integration-only changes.

## Commands and observations

Environment: Go 1.27.1, clang 20.1.8, Node 24.19.0, `nproc` = 5.
`bash cloud/setup.sh` reported Go, clang, Node and submodules ready in 0s each,
build-cache warming 70s, total 70s, CPU quota 4 and 17.6 GB memory. Every shell
sourced `/workspace/adamic-tools/env.sh`. Test output went to log files.

Main commands:

```sh
go test ./internal/fresh ./internal/lower -count=1 > proof-initial.log 2>&1
go test ./internal/fresh -run 'TestRegex|TestFutureRegex' -v -count=1 > proof.log 2>&1
go vet ./... > vet.log 2>&1
go test -count=1 -timeout 30m ./... > gate.log 2>&1
go vet ./... > vet-final.log 2>&1
go test ./internal/fresh ./internal/lower ./internal/oracle \
  -run 'TestRegex|TestFutureRegex|TestFreshWriteProbesStayRefused|TestCountsAreRecorded' \
  -count=1 -timeout 30m > final-proof-oracle.log 2>&1
```

Initial fresh/lower tests passed in 14.712s/7.135s. All three regex proof tests
passed in 0.009s. Both vet runs exited 0 with no output. The final focused run
passed: fresh 0.014s, lower 0.005s (no matching tests), oracle 17.022s. Its oracle
run includes all existing cycle-closing refusal probes and the complete main
counts table. Main's counts table is unchanged.

Additional independently caught proof mutants:

| Mutant | Catch |
|---|---|
| Do not interpret regex call arguments | `TestRegexOperandsStillJudgeCycleWrites`: lost cycle-closing operand write |
| Accept a future method without `WriteUnknown` | `TestFutureRegexMethodRemainsUnknown`: became silently safe |

Both mutant runs exited 1. All mutants were restored. The unsafe ordinary-field
mutant above is the memory-safety demonstration; stdout alone would not catch it.

Scratch commands, after merging the proof and CSS branches:

```sh
ADAMIC_CSS_FIXTURES=/tmp/adamic-regex-cycle/prettier \
  go test ./stage1/cohere/css -run '^TestCompositionOnNodeMatchesGo$' \
  -v -count=1 -timeout 30m > css-composition.log 2>&1
go test ./stage1/cohere/css -run '^TestRegexAndValueTreeComposeNatively$' \
  -v -count=1 > closed-gap.log 2>&1
go test ./internal/oracle \
  -run '^TestNativeAgreesWithNode$/internal/fresh/testdata/regexp_tree' \
  -v -count=1 > fixture-final.log 2>&1
go test ./internal/oracle \
  -run '^TestCountsAreRecorded$/fixtures/internal/fresh/testdata/regexp_tree.ts$' \
  -v -count=1 -args -update-counts > fixture-counts-test.log 2>&1
```

Composition passed in 246.656s: **24,076/24,076** native, Node and JS-backend
answers agree with Go, including error positions; ASan, UBSan and leaks pass.
The minimal closed gap passed in 2.826s. The final source-fixture oracle passed
in 0.532s, with fresh Node/native observations; its initial memory checks passed
in 51.006s. The standard counts harness measured this fixture's row in 0.177s:

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| internal/fresh/testdata/regexp_tree.ts | 97 | 97 | 65 | 81 | 35 | 0 |

The main table cannot include this row until the native regex dependency merges:
the ordinary main fixture set does not contain it. The scratch table was restored
after the filtered measurement. A scratch-wide counts update **failed**, separately,
because the older `sweeps/regexp_methods.a:7` passes `RegExpExecArray` to mutable
`RegExpMatchArray | null`; main's newer invariant-mutable rule refuses its wider
`index` field. This is not a cycle-proof failure. The sweep was not changed or
counted as passing; its failure log is retained.

Public Prettier fixture checkout: `cb4b33fba24a8428d00e54be85fc886288a374ea`,
sparse paths CSS, SCSS, Less and the two CSS-in-JS directories named by the port.
The walker found 158 CSS, 90 SCSS and 43 Less files, producing 12,038 texts in
both dialects. Go reports 5,006 successful composed parses. No new Prettier
library comparison or throughput claim is made in this compiler unit.

## Reproduction and limits

From this branch, merge `origin/codex/stage1-css` into a **scratch** branch.
Resolve the counts-table merge by keeping main's table, then apply
[the integration patch](verification/regex-cycle-proof/css-integration.patch).
Initialize the submodules in that worktree. Run the commands above with the
public fixture checkout. The patch contains only CSS test/documentation changes
and shape-initializer compatibility changes; the proof lives in this branch.

Proof commit: `32f8106613fb7708ce7192e04f7f1799b35c8434`.
Verified scratch merge: `3614121da2e75cb9f295653a8dfaf0853d344fa5`, followed by
scratch-only gap documentation commit `b154ab5`. Scratch branches are not pushed.

Dynamic regex compilation, replacement callbacks and mutable regex extensions
are not newly supported. Nested regex result references are deliberately
conservative, so this closes the demonstrated composition blocker without
claiming a complete alias proof for every possible regex/result program.

## Final gates

The complete main `go test -count=1 -timeout 30m ./...` exited 0. Notable package
results: fresh 32.529s, lower 10.347s, native 221.256s, oracle 258.769s,
unicodeproperties 687.458s, scanner 61.735s and parser 64.895s. All 24 reported
package lines were successful or had no tests. This run started before the final
oracle registration file was added; the final focused oracle run afterward
checks its main dependency gate, all refusal probes and the complete counts table.

After formally merging the committed proof, the full scratch CSS suite ran:

```sh
ADAMIC_CSS_FIXTURES=/tmp/adamic-regex-cycle/prettier \
  go test ./stage1/cohere/css -v -count=1 -timeout 30m > css-final.log 2>&1
```

It exited 0 in **141.826s**. Raw and composed 24,076-case Go agreements, both
backends, sanitizers/leaks, the three existing CSS semantic mutants, canonical
Range guards, the remaining four gap refusals and the closed regex/tree gap all
pass. PostCSS library comparison and throughput were explicitly skipped because
their optional environment variables were not set. No CSS production source or
submodule source changed. `gofmt -l` and `git diff --check` produced no output.

Evidence is retained under [verification/regex-cycle-proof](verification/regex-cycle-proof).
