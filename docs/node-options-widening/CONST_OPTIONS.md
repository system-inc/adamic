Extended the const-literal proof to require every reference to be a consuming host options argument.
Base 23ea5d92 on codex/host-stats-dirent-union; implementation and this report are committed together.
Lower/IR/flow, vet, Linux counts and the whole oracle with WASI pass; six combined host fixtures remain green.
The extra console.log(options) reference inside a capturing closure is caught by the predicate mutant.
The compiler hook is unchanged; forwarded readdir options and dynamic encoding fields remain outside the exemption.

The clarified rule supersedes the single-capture and prior-use assumptions in
checkpoint 8563e6aa below. Both local and module const bindings may have multiple
consuming references, including references in multiple closures. Every reference
must occupy a scalar-consuming options position in the host lowerer. Any other
reference refuses every consuming use, even when textually after a call. Scalar
field reads, writes, exports, aliases, returns, casts, storage, spreads and
arguments to other functions fail this rule.

Consumption-position recognition is separate from exact-shape verification,
so checking another reference does not recursively prove the same binding.
Each reference also retains the scalar-field restriction. The walk covers every non-declaration source file in the checker program,
including imported files and all closure bodies. An exported binding or export
alias is still refused conservatively. A cross-file regression checks both an
imported consuming closure and an object exported to another module. The loader
forces module scope, so the attempted script-global witness stopped at TS2304
(Cannot find name options); globals are not admitted by this base. Shorthand
object properties resolve their value symbol rather than the property symbol,
preventing {options} from hiding storage.

The updated node_options_const_capture.a covers two module consuming functions,
a local arrow closure and a direct consuming call using that same local binding.
All agree with Node on native, JavaScript and WASI. Nested function declarations
remain a language limitation on this base; the predicate test covers their shape,
while the executable fixture uses the supported arrow closure form. No language
rule was changed. node_options_widening.a now reuses its const for both calls;
Node stdout and stderr remain byte-identical, with both runs exiting zero.
It no longer needs an inline literal at its second call.

The new production mutant treats every call-argument reference as consuming,
including console.log(options) inside the capturing closure. Its direct predicate
test must fail with consumes argument = true, want false, rather than being
masked by a separate compiler check. Complete compiler tests additionally verify
adamic/no-optional-widening for this extra reference and for an escape that makes
both a direct call and a captured call unsafe. Existing let, prior-escape,
returned-object, planted-force and hidden-initializer mutants are rerun.

Validation logs for this clarification are /tmp/const-captures-*.log and
/tmp/const-consuming-capture-mutants/*.log. All six compile-valid mutants exit 1
with library_node_options_test.go:77: consumes argument = true, want false:
captured_extra_reference_mutant, let_mutant, prior_escape_mutant,
returned_object_mutant, cast_planted_field_mutant and cast_hidden_field_mutant.
Compile failures and a mutation that did not reach the parenthesized initializer
are not counted as kills. The corrected mutations exercise the intended checks.

Observed completed gates:

    go test ./internal/lower -run '^TestNodeHost' -count=1 -timeout=10m
    ok internal/lower 18.405s

    ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/node_options_(const_capture|widening)\.a$' -count=1 -parallel=2 -timeout=10m
    ok internal/oracle 1.517s

    go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -parallel=2 -timeout=30m -args -update-counts
    ok internal/oracle 467.777s

Linux counts change only node_options_const_capture.a, to 17/17/3/16/11/0.
The first attempted executable local test used a nested function declaration:
TestNodeHostOptionsConsumptionLowers at library_node_options_test.go:126 refused
main.a:1:112: stage 0 can't lower a function inside a function (a closure) yet.
Changing the test to an arrow closure resolves that existing language limitation.
The production proof is not weakened to accommodate it.

Final implementation package gate:

    go test ./internal/lower ./internal/ir ./internal/flow -count=1 -timeout=30m
    lower 210.674s; ir 42.257s; flow 246.707s; all ok

    go vet ./internal/lower ./internal/ir ./internal/flow ./internal/oracle
    exit 0

The six unchanged combined host fixtures (06, 07, 08, 10, 13 and 24) remain green
on native and JavaScript, comparing stdout, stderr and exit to status.json and
stock Node. The final program-wide proof is applied to the combined scratch;
its newer boxed-constant option handling is preserved. No fixture adaptation,
status recording, declaration, runtime or compiler hook was edited.
The existing compiler call is still optional_widening.go:135.

The whole oracle runs in isolation with WASI_SYSROOT set and
GOPROXY='https://proxy.golang.org|direct':

    ADAMIC_ORACLE_WASI=1 go test -v ./internal/oracle -count=1 -parallel=2 -timeout=30m

The complete final oracle passes in 601.498s with zero failing tests. Both
options fixtures pass TestNativeAgreesWithNode, TestWASIAgreesWithNode,
TestWASIEmission and TestCountsAreRecorded. The final counts verification takes
93.55s and agrees with the Linux table. Worker oracle observation caches are
permitted by CLAUDE.md; WASI comparisons execute in this configured leg. No
recorded broad-gate failure remains.

Only codex/host-stats-dirent-union is committed and pushed. Main is never pushed;
no force push, rebase or PR is used. The report update after these gates changes
no implementation, fixture or count, so it needs no further runtime test. No compiler-hook edit, main push, force push or rebase
is part of this change.

Historical checkpoint follows; its old single-capture and inline-second-call
restrictions are superseded by the clarification above.

Built a shared exact-shape proof for fresh const Node option objects, including parenthesized as const initializers.
Implementation 8563e6aa on codex/host-stats-dirent-union, from 70a2a970; the final report follows in a documentation commit.
Lower/IR/flow and Linux counts pass; six unchanged combined host fixtures agree with Node on native and JavaScript.
All four requested proof mutants and an extra hidden-initializer cast mutant are caught by the direct predicate assertion.
Whole oracle with WASI passes in 613.371s; remaining inline cases and the separate compiler-hook boundary are listed below.

The compiler hook at internal/lower/optional_widening.go:135 is unchanged.
Both the exemption and the fs-file field reader call nodeHostExactOptions.
Source-declared scalar fields are read from their declared type; a missing
optional field defaults only after the exact absence proof. No wider object is
stored, returned or passed to the runtime. Existing named NotYet overload and
encoding restrictions remain.

A binding must be an unannotated const with a fresh object-literal initializer.
Only parentheses and as const are unwrapped; structural casts, aliases, function
returns, parameters, let, spreads, methods and object-valued options fail.
Symbol identity distinguishes shadowed names. The source file is checked for
writes (including destructuring and loop targets), deletion, updates, receiver
method calls, aliases, storage, prior calls and captures. Declared scalar reads
are allowed. Exported objects fail the proof.

Assumption, following the request that the six real sys.ts fixtures turn green:
a private module const captured only as this proven consuming host argument is
consumption, not an earlier escaping capture. Every other capture is refused,
including closures returning the object and local captured bindings. For a module
consumer, every other object use is rejected regardless of text order: an escape
written after the function declaration can still precede its invocation.
This exception never admits general closure capture or a wider stored view.

The scratch combined proof is 3de95428, the existing merge of d18d2a4 with the
Stats/Dirent runtime dispatch fix. Only this library proof and the variable-options
condition were applied; its newer boxed-constant handling was preserved.
No host fixture, status.json, compiler refusal hook or language rule was changed.
The exact Node stdout, stderr and exit status were checked for each:

| Fixture | Native | JavaScript |
| --- | --- | --- |
| 06_fileExists | green | green |
| 07_directoryExists | green | green |
| 08_getDirectories | green | green |
| 10_getModifiedTime | green | green |
| 13_createDirectory | green | green |
| 24_useCaseSensitiveFileNames | green | green |

The new node_options_const_capture.a fixture holds the module consumer to Node
on native, JavaScript and WASI. The existing node_options_widening.a uses a
parenthesized as const initializer. Its second host call now uses a fresh inline
literal: the direct binding has already been passed to a call. Node output
before/after is byte-identical: directory followed by absent=true.
node_fs_file_stat.a also uses a parenthesized as const initializer and agrees
on native/JavaScript. No other existing fs fixture was rewritten in this unit.

Inline literals still needed in existing sources:

- node_options_widening.a: the second direct call, because the binding was used
  by an earlier call.
- node_fs_directory_entries.a and node_fs_directory_union.a: readdirSync forwards
  its options object, so it has no consumption exemption. The same inline call
  remains in stage3/host/fs_directory.a (used by node_fs_directory_system.a)
  and host fixtures 08_getDirectories.a and 25_readDirectory.a.
- node_fs_file_read.a: its encoding option remains inline because the host reader
  supports constant UTF-8 encodings, not encoding fields read dynamically from a
  binding. This is the existing named NotYet limitation, not an absence failure.

The six host fixtures keep their original module const and need no inline rewrite.
The remaining previously green 17 fixtures need no new adaptation for this proof.

Production-source mutants use isolated Go build overlays. Each intended test
exits 1 specifically with consumes argument = true, want false; compile failures
are not counted as kills:

| Mutant | Assertion |
| --- | --- |
| Remove the const-only requirement | let_mutant |
| Ignore prior escapes | prior_escape_mutant |
| Accept a function-return initializer | returned_object_mutant |
| Skip binding-use analysis after proving a fresh literal | cast_planted_field_mutant |
| Additionally accept a structural cast initializer | cast_hidden_field_mutant |

The fourth witness starts with {recursive:false}, then plants force='yes' through
a cast before rmSync. Node reads that runtime field and prints:

    The "options.force" property must be of type boolean. Received type string ('yes')

This differs from defaulting force to false and throwing ENOENT. The direct
predicate probe prevents an independent cast refusal from masking a bad host
classification. Separate complete compiler tests also verify the let, escaped,
returned and hidden-initializer programs receive adamic/no-optional-widening.

Completed commands and output at checkpoint:

    go test ./internal/lower ./internal/ir ./internal/flow -count=1 -timeout=30m
    lower 234.982s; ir 5.852s; flow 324.431s; all ok

    go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -parallel=2 -timeout=30m -args -update-counts
    ok internal/oracle 223.492s

Linux counts add the new row 5/5/1/7/5/0. Directory-system enumeration also grows
by the new source file: 2932/2932/5206/4374/1722/0. No macOS counts were used.
Focused native/JavaScript/WASI comparisons pass in 3.274s. The four required
mutants plus the extra assertion-cast mutant are all observed failures.
The complete configured oracle command is:

    ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -count=1 -parallel=2 -timeout=30m

WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot and
GOPROXY='https://proxy.golang.org|direct' are set. Test output is written to logs.

An environment restart interrupted earlier final gates and the first combined
compiler build; those incomplete runs are not passes. The first six native
builds then failed only because the root filesystem was full (29 GB Go build
cache). Clearing that disposable cache recovered 27 GB; the complete six-fixture
retry passes. Sources and pinned dependencies were preserved.

Only the owned branch is pushed. No main push, force push, rebase or PR is used.

Final gate on the unchanged implementation 8563e6aa:

    ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -count=1 -parallel=2 -timeout=30m
    ok github.com/system-inc/adamic/internal/oracle 613.371s

This complete retry ran without concurrent build work. All host comparisons,
existing mutants, counts and the WebAssembly leg pass. Vet also passes:

    go vet ./internal/lower ./internal/ir ./internal/flow ./internal/oracle

The initial complete oracle failed in 1281.469s with these exact failing cases
and first error lines (cache-miss notices are not failure reasons):

| Test | First error |
| --- | --- |
| TestMapSmallMutants/promotion_order | library_map_small_test.go:83: native: mkdir /home/agent/.cache/adamic/runtime/.build-3829941808: no space left on device |
| TestMapSmallMutants/nan | library_map_small_test.go:83: native: mkdir /home/agent/.cache/adamic/runtime/.build-2461712602: no space left on device |
| TestMapSmallMutants/zero | library_map_small_test.go:83: native: mkdir /home/agent/.cache/adamic/runtime/.build-3092398120: no space left on device |
| TestMapSmallMutants/promotion_iterator | library_map_small_test.go:83: native: mkdir /home/agent/.cache/adamic/runtime/.build-3770310513: no space left on device |
| TestRegExpLongBacktrackNode | regexp_long_backtrack_test.go:47: long regex exceeded 3m: context deadline exceeded |
| TestRegExpNativeTiming/quadratic_exec | regexp_native_fixes_test.go:57: native execution exceeded 3s or failed: signal: killed (3.001221985s) |
| TestRegExpNativeTiming/quadratic_matchall | regexp_native_fixes_test.go:57: native execution exceeded 3s or failed: signal: killed (3.007826054s) |
| TestRegExpNativeTiming/quadratic_test | regexp_native_fixes_test.go:57: native execution exceeded 3s or failed: signal: killed (3.003938065s) |

Every initially failed group passes without concurrent build work in 222.461s:

    go test ./internal/oracle -run '^(TestMapSmallMutants|TestRegExpLongBacktrackNode|TestRegExpNativeTiming)$' -count=1 -parallel=2 -timeout=10m

The first broad run overlapped other build jobs, so its timing results were not
an isolated algorithmic-complexity gate. The isolated broad retry establishes
that no recorded failure remains. No map, regexp, deadline or harness change
was made in this unit.

Separate compiler-hook boundary, verified on the implementation's compiler:

```a
import * as fs from 'node:fs'; const options={throwIfNoEntry:false} as const; fs.statSync('x',(options));
```

This still refuses at 1:96 for missing optional bigint. The library predicate
proves the parenthesized argument's shape, but the unchanged compiler hook
checks the inner identifier again outside its direct-argument exemption.
Parentheses in the initializer work; the six actual host sites pass an identifier
directly and are green. Automatic approval review rejected a proposed change to
optional_widening.go because the existing compiler-hook boundary was fixed.
That edit was not applied, and no alternate hook or AST workaround was used.
The hook boundary remains compiler work, separate from this library proof.

Module initialization remains protected by the core dead-zone check before a
captured global's fields are read. A helper invoked before its const initializer
emits the ReferenceError guard before extracting throwIfNoEntry; the proof does
not discard or replace that check.

The final documentation commit changes no implementation or fixture, so it does
not require another gate. Only the owned branch is pushed, without a main push,
force push, rebase or PR.
