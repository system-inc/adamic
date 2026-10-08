import json, subprocess
from pathlib import Path
root=Path(__file__).resolve().parent
r=json.loads((root/'results.json').read_text())
m=json.loads((root/'mutants.json').read_text()) if (root/'mutants.json').exists() else []
base=subprocess.check_output(['git','rev-parse','origin/main'],text=True).strip()
text=f'''Built nine .a probes: seven agree in four modes; two are explicit compiler refusals.
Probe commit 9197ad5b4f0716bb2f0a6fd924ae95d594fe646c; tested base {base}; history examined: 73352e87..5a2681b1.
Commands: run.py compares source Node, ASan/UBSan native, -O2 native and backend Node; seven leak runs finish cleanly.
Mutants: negative-wrap and same-kind guards caught; regexp guard survives selected native tests but oracle catches it; readiness call invalidation survives lower; extra view retain caught by leak check.
Not covered: every readiness path, unsupported element types, allocation exhaustion, resizable/shared buffers, or the full repository gate.

Observations

The requested branch was created with `git fetch origin && git checkout -b codex/coverage-oct8-typedarrays origin/main`. The named five source files and their requested `git log -p` history were read. CLAUDE.md, README.md and the language/memory documentation were consulted. No compiler changes are committed. These probes are notes, not new registered oracle fixtures; internal/oracle/counts.md is unchanged.

`source /workspace/adamic-tools/env.sh` was used for builds and tests. Setup succeeded; its complete output is in setup.log. It reported Go ready 0.131s, Node ready 0.148s, clang ready 0.517s, markdown dependencies ready 1.763s, submodules ready 25.052s, Go build ready 366.299s, cache warm 366.642s and done 366.813s. `nproc` is 5; cgroup CPU quota is 4. Go 1.27.1, clang 20.1.8, Node 24.19.0. The first package run lacked the pinned @types/node 25.3.3 dependency; `npm ci --prefix stage3/api` installed the lockfile dependencies successfully, then lowering was rerun. The first attempt to run probes preceded compiler build completion and failed with FileNotFoundError; the completed final run is recorded here.

Reproduction: build `go build -o /tmp/coverage-adamic ./cmd/adamic`, then `python3 notes/coverage-oct8/typedarrays/run.py > /tmp/coverage-runs.log 2>&1`. The runner saves each process stream in /tmp/coverage-oct8-typedarrays and records commands, streams and exit codes in results.json. Builds use the compiler's own --sanitize option (oracle flags: -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all) and default -O2. Sanitizer comparison uses detect_leaks=0, as the oracle does; successful binaries are run separately with detect_leaks=1. Each subprocess has a 60-second deadline. Source runs use oracle/node.mjs to strip types from the source, independently of Adamic lowering. Compilation failures are recorded as compilation outcomes, not invented native executions.

Coverage

| Program | Cases | Four modes | Leak check |
|---|---|---|---|
'''
cases={
'uint8array':'Uint8 modulo/truncation, NaN/infinities, huge values, subnormals, constructor vs assignment',
'int32array':'Int32 signed wrap, NaN/infinities, huge values, subnormals, constructor vs assignment',
'float64array':'Float64 preservation, signed zero reciprocal, NaN/infinities and subnormals',
'replacement':'Named/unmatched captures; $$, $&, $1, $2, $<name>, prefix/suffix tokens; string and regexp patterns; function replacers; Unicode/UTF-16 empty matches; callback lastIndex writes; no match',
'readiness':'Branches, loop assignment, caught throw/finally, field alias, direct-call field write, supported object spread',
'readiness-fields':'Constructor branch initialization; narrowed interface reads; captured reads; false overload result to boolean-or-undefined; rest with default and explicit undefined',
'rest-closure':'Rest function value and class method, comment witness',
'buffer-gap':'Offset/length view over ArrayBuffer, deliberately unsupported',
'readiness-callback':'Historical uninitialized-slot syntax and callback rebinding, deliberately refused on current main'}
for name,v in r.items():
 leak=v.get('leaks',{});clean=leak.get('exit')==0 and leak.get('stderr')==''
 text+=f"| {name}.a | {cases[name]} | {'agree, exit 0' if v['agree'] else 'compiler refusal; source exit 0'} | {'clean' if clean else 'not applicable'} |\n"
text+='''
All three typed-array programs additionally cover nested nonzero offsets, lengths after subarray, left and right overlapping set between sibling/nested views, fractional offsets (including -0.9), fill over a view, empty views and empty set, and a returned nested view after its original bindings have gone out of scope. Existing typed_arrays_views.a already exercises basic shared views and overlap; these programs add the same nested offset and lifetime scenario for each width and much larger conversion magnitudes.

Current-main readiness differs from the historical landing: non-null assertions in .a are refused, and .ts assertions are eager. The historical assertion-based readiness witness therefore never reaches lowering. It cannot establish a readiness miscompile on this base. Accepted readiness controls were written without assertions. No accepted-program output disagreement, C compilation failure, runtime abort, or leak was observed in these nine probes.

The two compiler disagreements below are supported-gap observations, not silent miscompiles. `internal/lower/typed_arrays.go:52` returns NotYet for ArrayBuffer. `internal/lower/readiness.go:21` excludes checked TypeScript syntax from lazy initialization; the .a witness is refused by the non-null refusal in internal/lower/refusals.go:104 before readiness can create its historical slots.

Verbatim four-mode observations

Empty stderr is written as an empty fenced block. Native/compiler phases are explicitly distinguished. The programs themselves are adjacent .a files. For agreeing programs all four streams are identical, but each is included so the evidence is self-contained.

'''
for name,v in r.items():
 text+=f'### {name}.a\n\n'
 for mode in ['source','sanitized','release','javascript']:
  out=v[mode];text+=f"{mode}: exit {out['exit']}, phase {out.get('phase','execution')}.\n\nstdout:\n\n```text\n{out['stdout']}```\n\nstderr:\n\n```text\n{out['stderr']}```\n\n"
text+='''Comment and documentation audit

Every comment in the five named files was compared to its adjacent implementation. This is a source audit, not a claim that every specification sentence was experimentally proven.

| Location | Observation |
|---|---|
| typed_arrays.go:30,216,290-291 | Unsupported library facilities are rejected while identity is available; optional offsets are fitted to MaybeNumber; set accepts only same-kind storage, not an object shape. Comments match these paths. |
| typed_array.c:1,34,44,67,136 | Counted views retain an owner; empty allocation has a non-NULL byte base; -0.5 truncates to -0; integer stores use finite/trunc/fmod and explicit signed wrapping; same-kind set uses memmove. Comments match. The probes exercise numeric conversion and views; allocation failure is not tested. |
| regexp_replace.c:1-2,9-10 | Intrinsic callback replacement first collects matches, constructs actual-kind arguments and validates/converts them to closure representations. The ECMA section reference describes this intrinsic path, not arbitrary overridden RegExp execution. No mismatch found in these comments. |
| readiness.go:18-19,380,385,399 | Initializer syntax recognition, eager .ts checks, and lazy operand storage match the functions. Historical lazy .a syntax is now refused before this machinery; this limits its reachable coverage. |
| readiness.go:52-54 | The monotonic statement concerns binding readiness between declarations. Field facts are separately invalidated by calls at lines 192 and 214. Read in that scope it matches; it would be misleading if taken to claim fields survive calls. |
| readiness.go:56-57,124-125,152,249,344 | Record field names keep representation checks; local metadata is cloned for CFG construction; must analysis starts at top and intersects predecessor states; reflection transforms this instruction without recursively transforming nested statements with graph locations; field facts key binding and name. Comments match the code. |
| census_small.go:10-11 | **Comment mismatch:** function values and methods do not remain NotYet on this base. rest-closure.a prints 3:2 in all four modes. The census rest helpers have no callers found by rg; functions.go:138-149 now handles rest parameters. |
| census_small.go:34,53-54,97 | The helper implementations recover a rest declaration, build a fresh packed array and pad fixed arguments. No callers were found, so the comments describe the helper bodies rather than the currently active path. |
| census_small.go:109 | **Comment mismatch:** 'other slotless representations remain refused' excludes Union in the actual condition. slotless (expression.go:1071) recognizes only MaybeBoolean and Union; the helper exempts both. |
| census_small.go:114,131,158-159,169-170,184-186 | Only body-bearing implementation declarations are recovered; the relation applies classAssignable and widened; generic binders are mapped, extra binders inferred from the signature and no-evidence candidates still checked. Comments match the code. |
| census_small.go:215,218,240,270-272 | Extra parameters ignored, absent rest gives an empty array, nullable overload results are proved or checked and calls fit the resolved result representation. Comments match; readiness-fields.a covers the present false result. |
| census_small.go:332 | Callable slots share the tagged boolean exemption and additionally exempt Union. The comment does not say boolean is the only exemption; no contradiction established. |
| census_small.go:337,348-349,360,378-379,400-401,409-410 | Defaults add undefined to the incoming contract; boolean-or-undefined truth is exactly true; checker-never branches with two booleans retain Boolean; general conditions use one toBoolean argument; never-rest marker checks rest/never and storage can compare results. Comments match their code. |

Mutant observations

Run `python3 notes/coverage-oct8/typedarrays/mutants.py > /tmp/coverage-mutants.log 2>&1` with the setup environment sourced. Each mutation starts from the same restored source and the finally block restores it. All edits stay in named territory. Exact replacements, commands and exit codes are in mutants.json; complete catcher logs are beside this report. No killed mutant is claimed on the basis of clang warnings alone. The negative-wrap mutant fails with a release output mismatch (-2147483648 instead of 2147483647) and UBSan float-to-unsigned-char overflow. The same-kind guard mutant fails because a mismatched-kind set returns normally instead of panic 70. The callback argument guard survives the selected native tests, but its oracle fixture fails because backend exits 70 while native exits 0 and prints absent. Disabling both readiness call invalidation branches at readiness.go:190 and 212 survives the complete lower package (exit 0): category branch no package test caught. No inference that the branch is removable follows from this survival.

'''
for row in m:
 text+=f"- `{row['mutant']}` in `{row['file']}:{ {'typed-negative-wrap':70,'typed-kind-guard':127,'regexp-argument-guard':17,'readiness-call-invalidation':190}[row['mutant']]}`: `{row['before']}` becomes `{row['after']}`. Package command `{' '.join(row['command'])}` exits {row['exit']}."
 if 'oracle' in row:text+=f" Additional uncached oracle command `{' '.join(row['oracle']['command'])}` exits {row['oracle']['exit']}."
 text+='\n'
probe_file=root/'probe-mutants.json'
if probe_file.exists():
 for row in json.loads(probe_file.read_text()):
  text+=f"\nNew-probe experiment `{row['mutant']}` at typed_array.c:{70 if row['program']=='uint8array.a' else 149}: compiler build exit {row['build_exit']}, run.py on {row['program']} exits {row['exit']}. Exact edit: `{row['before']}` to `{row['after']}`. Full four-mode and leak streams are in {row['mutant']}.json.\n\n"
  observed=json.loads((root/(row['mutant']+'.json')).read_text())
  for program,result in observed.items():
   for mode in ['source','sanitized','release','javascript','leaks']:
    if mode not in result:continue
    out=result[mode]
    text+=f"{mode}: exit {out['exit']}.\n\nstdout:\n\n```text\n{out['stdout']}```\n\nstderr:\n\n```text\n{out['stderr']}```\n\n"
text+='''
Failure categories for the new-probe mutants: negative wrap produces different output in release and a runtime abort under UBSan; excess owner retain is memory not freed, while all four comparison modes still agree. These are deliberate mutants, not defects observed on main.

Inferences and limits

Agreement supports these concrete programs, not arbitrary correctness of the runtime or lowering. A surviving mutant establishes only that its listed tests did not notice the change. Historical readiness guard survival may be explained by its former syntax being unreachable on current main; that is an inference, not evidence that the guard is safe to remove. Native callback unit-test survival does not establish that the oracle lacks coverage; its separately run argument-guard fixture is the relevant external behavior check.

The safe multiargument buffer constructors, unsupported integer/float element kinds, cross-kind set, memory exhaustion/overflow allocation guards, callback group/rest guard combinations, throwing callback cleanup beyond existing oracle fixtures, arbitrary RegExp overrides and every readiness CFG path were not exhaustively covered. No production fix, PR, merge or full repository gate is part of this unit.

Validation logs

Exact unmutated commands (all output redirected to logs):

```sh
go test ./internal/lower ./internal/native -count=1 -timeout 30m > /tmp/coverage-packages.log 2>&1
go test ./internal/lower -count=1 -timeout 30m > /tmp/coverage-lower-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(typed_arrays_|regexp_replace|non_null_uninitialized)' -count=1 -timeout 30m > /tmp/coverage-oracle.log 2>&1
go test ./internal/native -run 'TestTypedArrayRuntime|RegexPrograms|ClosureConvention' -count=1 -timeout 10m > /tmp/coverage-native-focused-final.log 2>&1
go vet ./internal/lower ./internal/native > /tmp/coverage-vet.log 2>&1
git diff --check > /tmp/coverage-diffcheck.log 2>&1
```

The initial combined command exits 1 because lower lacks @types/node; its entire native package passes (437.854s). After installing the dependency, the entire lower package passes (49.131s). The uncached filtered oracle passes (42.486s). The final focused native command passes (1.495s). Vet and diff-check exit 0 with empty output. Final run.py exits 0 and checks seven agreeing/leak-clean programs plus two expected refusals. The full repository gate was not run. The later staged `git diff --cached --check` reported one trailing space in typed-negative-wrap.log:7 (exit 2). That space is the verbatim UBSan output and is intentionally preserved; the earlier unstaged diff check had no output.

'''
for name in ['setup','npm','build','runs','packages-initial','lower-final','oracle','mutants','probe-mutants','native-focused-final','vet','diffcheck']:
 p=root/(name+'.log')
 if p.exists():text+=f'`{name}.log`:\n\n```text\n{p.read_text()}```\n\n'
(root/'REPORT.md').write_text(text)
