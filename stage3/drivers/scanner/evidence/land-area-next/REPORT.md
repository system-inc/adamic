The combined merge conflicts in 44 paths; scanner measurement uses cloud/land-area-next alone.
Base 28285421 contains compiler area-next 337aa466; pinned combined branch is 2648ad33.
Compiler built in 185.816s; both real scanner modes stop at corePublic.ts:9:5, MapLike's index signature.
Both full-tree Node diffs pass and both exactly-one-byte Node mutants are caught.
Fifteen distinct stops have Node-success/build-failure witnesses; typed Error capture also remains unsupported.

# Landing compiler alone after the combined merge conflicts

The user overrode the area/compiler instruction with origin/cloud/land-area-next.
The superseded compiler-area scratch merge was aborted and excluded entirely.
Created fresh never-pushed branch scratch/scanner-native3-next at
/workspace/scanner-native3-next, from fetched landing tip
28285421bf59ea46269545144c19123156c00c91. Ancestry confirms compiler/area-next
337aa466 is included; the landing commit states main 547b37d6 as its main base.
Merged pinned combined branch 2648ad3362812b0f184a8cb922bc322d18729c66, which
conflicted in exactly 44 paths. The complete set below was captured before abort.
No hand resolution was attempted. Compiler and scanner measurements then used
28285421 alone, exactly as the overriding instruction requests.

## Exact conflicting paths

```text
CohereSettings.json
internal/flow/flow_test.go
internal/fresh/fresh.go
internal/ir/ir.go
internal/javascript/javascript.go
internal/load/load.go
internal/load/node_library.go
internal/load/source_fs.go
internal/lower/cast.go
internal/lower/cast_proof.go
internal/lower/control.go
internal/lower/expression.go
internal/lower/functions.go
internal/lower/generic.go
internal/lower/invariance.go
internal/lower/library_json_stringify.go
internal/lower/library_math_number.go
internal/lower/library_method_values.go
internal/lower/library_node_buffer.go
internal/lower/library_node_fs_file.go
internal/lower/library_node_fs_file_test.go
internal/lower/library_node_test.go
internal/lower/library_object.go
internal/lower/library_string.go
internal/lower/locals.go
internal/lower/modules.go
internal/lower/object.go
internal/lower/refusals.go
internal/lower/statements.go
internal/native/emit_functions.go
internal/native/emit_objects.go
internal/native/fields.go
internal/native/library.go
internal/native/runtime/adamic.h
internal/native/runtime/node_fs_file.c
internal/native/runtime/node_process.c
internal/native/runtime/object.c
internal/native/runtime/regexp.h
internal/oracle/counts.md
internal/oracle/node_fs_file_test.go
internal/oracle/oracle_test.go
internal/oracle/testdata/node_fs_file_close.a
internal/oracle/testdata/optional_widening_refused/node_fs_file_buffer.a
stage3/fixtures/host/status.json
```

Raw merge output is merge.log. The compiler checkout has no source edits.
The only worktree filesystem replacement is the empty cohere directory with a
symlink to /workspace/adamic/cohere at the pinned commit. The old hand-resolved
checkouts are excluded. npm ci later installs ignored API dependencies from
this base's own lockfile; it changes no tracked package or compiler files.
Git source status is clean with submodule inspection excluded. No global blocker
retirement is claimed: MapLike closed on the previous combined compiler, but
that compiler could not merge here and is not the compiler measured below.

## Real scanner result and reference comparison

Both split=0 and split=1, jobs=5, stop at corePublic.ts:9:5:
Adamic refuses an index signature; use a Map, which keeps keys in insertion order.
The unchanged scanner never reaches native emission or clang. No native scanner
binary, full native byte comparison, scanner timing, binary size or native-output
mutant is available.

The first attempts hit a loader prerequisite instead: this base requires
@types/node 25.3.3 in stage3/api/node_modules/@types/node. Installed the committed
lockfile with npm ci --prefix stage3/api --ignore-scripts, then reran both modes
in new output directories. Three packages were installed in the logged 444ms.
Initial loader reports are retained under each mode's prerequisite directory.
This setup failure is excluded from the feature stop count.

Both final slice Node dumps equal the fixed full-tree reference byte for byte,
with empty diffs and exit 0, over every file in the same 81-file src/compiler
corpus. Counts: 1,369,432 tokens and 466 errors, 108,019,935 bytes;
509,014 skipped-trivia tokens and 860,418 retained-trivia tokens. SHA256:
5cce1570354cc48b5d9db246daf2abe8db3c21edba3c35e613da812a26920182.
The dump includes cooked values, error rows and fixed absolute path headers.
Copied Node controls pass diff, exit 0. Both token-end mutants change exactly
one byte and fail the same diff, exit 1. The audit counts the changed bytes.
These are Node-output mutants, not native scanner evidence.

## Error capture and ordered stops

The real Debug body first reaches its Error-as-any cast at fresh order 3 below.
Replacing that body with a throwing placeholder skips its later capture call;
no claim is made that the actual unchanged body reached a static-method gate.
An independent typed capture-stack-marker-control.a, unchanged from the prior
control, was rerun as witnesses/error-capture-control.a. Node exits 0 and prints
ok; this compiler exits 1 at line 3, column 1 with:

```text
stage 0 can't lower node:globals.ErrorConstructor.captureStackTrace yet
```

That is supplementary ErrorConstructor evidence, not an extra ordered scanner
stop. It establishes the expected library omission on this base. Source, both
outputs and timed exit records are in witnesses/error-capture-control.*.

The fifteen distinct stops below each have a small program that finishes on
Node (exit 0) and fails this compiler (exit 1) with the same diagnostic. They are
fourteen lowering gates and one native clang failure. Coordinates refer to the
successively modified scratch discovery copy, except order 15's generated C.
The driver location is part of the same scanner unit; it is labelled separately.
Full messages and locations are in stops.json, and witness results in
witnesses/results.json.

| Order | Scratch location | Message | Node witness |
| --- | --- | --- | --- |
| 1 | corePublic.ts:9:5 | Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added | [01-index.a](witnesses/01-index.a) |
| 2 | debug.ts:8:5 | stage 0 can't lower a mutable namespace export; use a module or export functions around private state yet | [02-mutable-namespace.a](witnesses/02-mutable-namespace.a) |
| 3 | debug.ts:15:14 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) | [03-error-cast.a](witnesses/03-error-cast.a) |
| 4 | diagnosticInformationMap.generated.ts:13:34 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) | [04-diagnostic-cast.a](witnesses/04-diagnostic-cast.a) |
| 5 | utilities.ts:65:22 | stage 0 can't lower typed array element type Uint16Array yet | [05-uint16.a](witnesses/05-uint16.a) |
| 6 | scanner.ts:1269:28 | Adamic 0.1 refuses an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Identifier; compare with this literal and return that constant, or widen the slot to the whole numeric enum or number (adamic/enum-literal) | [06-enum-slot.a](witnesses/06-enum-slot.a) |
| 7 | scanner.ts:3497:67 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) | [07-string-cast.a](witnesses/07-string-cast.a) |
| 8 | core.ts:107:20 | stage 0 can't lower new an Identifier yet | [08-array.a](witnesses/08-array.a) |
| 9 | utilities.ts:13:17 | stage 0 can't lower a function returning U &#124; undefined yet | [09-generic-return.a](witnesses/09-generic-return.a) |
| 10 | scanner.ts:608:87 | stage 0 can't lower a value of type any yet | [10-any.a](witnesses/10-any.a) |
| 11 | scanner.ts:113:5 | stage 0 can't lower a computed field name yet | [11-computed-field.a](witnesses/11-computed-field.a) |
| 12 | scanner.ts:100:23 | stage 0 can't lower a Map of any yet | [12-map-any.a](witnesses/12-map-any.a) |
| 13 | scanner.ts:432:59 | stage 0 can't lower Object.entries on a shape not proven by a plain literal or its const binding yet | [13-entries.a](witnesses/13-entries.a) |
| 14 | driver/main.a:13:42 | stage 0 can't lower a value of type any yet | [14-any-callback.a](witnesses/14-any-callback.a) |
| 15 | generated/main.c:7923:9 | incompatible pointer types returning 'adamic_object *' (aka 'struct adamic_object *') from a function with result type 'adamic_string *' (aka 'struct adamic_string *') [-Werror,-Wincompatible-pointer-types] | [15-never-string.a](witnesses/15-never-string.a) |

Order 15 occurs in getNameOfScriptTarget's emitted return after the generic
forEachEntry placeholder was changed to a throwing never-return signature.
The returned temporary is adamic_object*, while the function returns adamic_string*;
clang's -Werror catches it. The discovery path is dependent on that placeholder.
The independent three-line 15-never-string.a has no scanner stubs: an unused
string|undefined function returns a never-returning call, and the program prints
ok on Node. Adamic emits the same incompatible pointer return at generated
main.c:30:9. This is a native build failure, not a runtime silent miscompile.
The full discovery C is losslessly compressed in discovery.c.gz, with its relevant
return in clang-source-excerpt.txt. The emitted-C source location is main.c:7923:9.
No mutant is claimed from a clang type error.

## Discovery placeholders and limits

The bounded walk successively uses these explicit uncommitted placeholders:

1. Remove the MapLike index signature and append a dormant throwing marker.
2. Replace Debug's mutable exported let with const false plus a dormant throwing
   marker function; retain the namespace and its actual function bodies initially.
3. Replace Debug.fail's body with a throw, preserving its signature.
4. Replace diag with a throwing body and explicit DiagnosticMessage return.
5. Replace parsePseudoBigInt with a throwing body.
6. Replace getIdentifierToken with a throwing body, preserving its union return.
7. Replace the UTF16 worker with a typed throwing arrow.
8. Replace levenshteinWithMax with a throwing body and inferred return annotation.
9. Replace forEachEntry with a throwing body and never return; this changes the
   signature and is not a validated generic implementation.
10. Replace the error helper's body, then the whole createScanner factory body
    because the any parameter signature remains otherwise.
11. Replace textToKeywordObj with a typed throwing initializer.
12. Replace its keyword Map initializer with a typed throwing Map initializer.
13. Replace the Unicode property alias Map with a typed throwing initializer.
14. Replace driver error registration with a named throwing void function call,
    because the callback's any parameter remains after replacing its body.

The fifteenth failure is not bypassed. Type-only and mutable-binding placeholders
use explicit dormant throw markers rather than claiming an implemented replacement.
All modifications are shown in discovery-only.json, including the copied driver.
No compiler or adaptation implementation edit is delivered.
Whole createScanner replacement omits its nested bodies from later traversal;
this is a bounded discovery walk, not an exhaustive census. Order 12's Map<any>
requires the erased index declaration and keyword-object placeholder, and is
explicitly qualified as dependent on that changed representation. The independent
witness demonstrates the Map<any> diagnostic, not correctness of that placeholder.
No successful native scanner behavior behind these placeholders is claimed.

Six repeated diagnostics are excluded from the fifteen: five repeats of the
unchanged scanner error-helper any signature and one of the unchanged driver
callback any signature. Their rows/logs remain in the two repeated-attempt folders.
A throwing registration IIFE made subsequent driver code unreachable and induced
two TS2339 narrowing errors; that checker artifact is retained and excluded.
A named void-return throwing helper avoided that artifact. Intermediate logs
retain all attempts, while the final table contains fifteen distinct locations/
messages. The first enum witness attempt compiled because it used a whole enum
slot; the final witness assigns an unproven number to SyntaxKind.Identifier,
which reproduces the exact enum-literal diagnostic. The initial attempt remains
in witnesses.log and is not counted as a final failure witness.

## Build timings, commands and validation

Setup ran with GOPROXY='https://proxy.golang.org|direct' and passed. Cumulative
lines: Node 0.021s, Go 0.028s, submodules 0.068s, markdown dependencies 0.082s,
clang 0.163s, Go build 61.913s, deferred test binaries 63.104s,
warm build cache 63.125s, total 63.387s. nproc=5, cgroup quota=4 CPUs.
Raw timing lines are in setup.log.

The landing-branch Go compiler build exited 0: wall 185.816s, user 464.636s,
system 35.133s. SHA256:
e916052982e889f32efc9a1b492323fc19c6883a882496e257d0cfa5a854c0f0.
Setup and scratch build overlapped; these are observed timings, not isolated
performance comparisons. The real scanner has no native execution timing.

Compiler merge/build/API installation ran in /workspace/scanner-native3-next;
setup and driver/discovery commands ran in /workspace/adamic. Each execution
sources /workspace/adamic-tools/env.sh. All output went to log files.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/scanner-next-setup.log 2>&1
TIMEFORMAT='wall=%3R user=%3U sys=%3S'
{ time go build -buildvcs=false -o /workspace/scratch/scanner-next-adamic ./cmd/adamic; } > /tmp/scanner-next-compiler.log 2>&1
npm ci --prefix stage3/api --ignore-scripts > /tmp/scanner-next-api-install.log 2>&1
STAGE3_CACHE=/workspace/scratch/native3-cache ADAMIC_NATIVE_SPLIT=0 bash stage3/drivers/scanner/run.sh /workspace/scratch/native3-next-unsplit-ready --tree /workspace/scratch/native3-slice --inputs /workspace/scratch/native3-full/adapted --oracle /workspace/scratch/native3-reference-final/node.stdout --compiler /workspace/scratch/scanner-next-adamic --compiler-cwd /workspace/scanner-native3-next > /tmp/scanner-next-unsplit-ready.log 2>&1
STAGE3_CACHE=/workspace/scratch/native3-cache ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=$(nproc) bash stage3/drivers/scanner/run.sh /workspace/scratch/native3-next-split-ready --tree /workspace/scratch/native3-slice --inputs /workspace/scratch/native3-full/adapted --oracle /workspace/scratch/native3-reference-final/node.stdout --compiler /workspace/scratch/scanner-next-adamic --compiler-cwd /workspace/scanner-native3-next > /tmp/scanner-next-split-ready.log 2>&1
python3 /tmp/scanner-next-walk.py > /tmp/scanner-next-walk-ready.log 2>&1
# Resume logs 1 through 8 retain manual placeholder steps described above.
python3 /tmp/scanner-next-witnesses.py > /tmp/scanner-next-witnesses-final.log 2>&1
python3 /tmp/scanner-next-finalize.py > /tmp/scanner-next-finalize.log 2>&1
python3 /tmp/scanner-next-evidence.py > /tmp/scanner-next-audit.log 2>&1
```

Reproduction scripts preserve the actual paths. Evidence audit passes all
fifteen Node-success/build-failure diagnostic pairs, exact 44 conflict count,
both reference/control diffs and exactly-one-byte mutant catches, plus the
supplementary typed ErrorConstructor probe. Readable logs strip trailing
whitespace and extra EOF blank lines; raw-whitespace-logs.json preserves the exact affected bytes.
No whole package or full gate was run, no registered fixture was added and
counts.md is unchanged. Native scanner execution, ownership/sanitizer proof,
parser-directed rescans and a merged combined-compiler measurement remain
uncovered. Delivery changes are scanner evidence and notes only; scratch is
never pushed.
