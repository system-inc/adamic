Built three CFG composition helpers in separate .a files: conditionalExpression, bindWithDefault and memberHeader.
Commits: rebased implementation a30ffd8e0f441fc56b98ee2b000b996ce99d9dbe on lint area b84a9d93 containing main c7991b90; replacement claim originally published at 397caa0b; final publication SHA is named in the final response.
Commands: all 51 owned helpers PASS 832.766s and 3,045,173 comparisons; shared helpers PASS 92.769s; selected harness PASS 245.331s; vet/format clean; six uncached probes PASS 19.319s; setup 53s, nproc 5.
Mutants: all 182 owned variants (twelve new), four inherited variants, missing-consumer check, emitted-JavaScript mismatch and suggestion-edit variants caught; compiler failures are never counted.
Not covered: complete rule findings, parser/backend integration, arbitrary graph dependencies and the full repository gate; no required correctness check was skipped, weakened or deleted.

# Scope and ownership

The original JSX/hook-search claim dd8d7cfa was ten seconds later than slot 04 df5d5097. All three were withdrawn. Their duplicate sources are absent. The abandoned run's raw evidence is compressed in evidence/withdrawn-search.log.gz; seven initial variants passed, while the redundant-search variant was stopped after its recursive work expanded. None is delivered or counted. Replacement claim 397caa0b at 07:39:38 UTC was pushed before source. Final wildcard fetch checked twenty origin helper branches; no other claims mention any replacement. The comment collection bundle remains reserved in shared HELPERS.md; regex-engine internals are excluded by the JS RegExp instruction.

The replacement helpers tie the highest eligible unclaimed fan-out, four rules each. No shared generator, test harness, rule directory, compiler or runtime file was changed. The helpers take parser views and opaque node/block arena handles, rather than dispatching rules or inventing numeric AST kinds. No regex is introduced. Prior forty-eight helpers were green and published on the unchanged fetched bases before these claims.

# Rules and readiness

Each of the three helpers removes a dependency for all four rules:

- array-callback-return
- consistent-return
- no-unreachable-loop
- react-hooks/rules-of-hooks

This is twelve removed dependency edges, four unique consumers and zero final blockers removed by this batch alone in the frozen inventory. readiness.json subtracts only these three symbols and lists every remaining dependency. This is helper readiness, not completed rules or native findings parity.

# What the oracle observes

The pinned upstream cohere revision is 715ba94f3608a6500086b1076ce5cb7e51b836db. The capture overlay records actual fixture strings from the upstream core and React packages without changing their worktrees. Both packages pass, capturing 2,119 sources across every consumer. The comparator parses those sources with the actual Go parser and builds their real CFG roots.

Temporary overlays rename the helper bodies and dependency methods, then invoke the original bodies through owned observation wrappers. Nested dependency work executes the actual Go implementation; only the outer helper's direct calls are recorded. Records hold operation names, node/block identities, callback results and before/after cursor state. The Adamic driver replays callback results/cursor changes and independently generates the requested calls and arguments. It compares the complete trace and final cursor against Go on source Node, emitted JavaScript and sanitized native. Expression evaluation, pattern binding, decorators, graph allocation, edge storage and entry remain explicit dependencies; their semantics are not implemented or proved by replay.

Observed live calls by consumer:

| Consumer | conditionalExpression | bindWithDefault | memberHeader |
|---|---:|---:|---:|
| array-callback-return | 5 | 75 | 262 |
| consistent-return | 0 | 2 | 38 |
| no-unreachable-loop | 0 | 114 | 171 |
| react-hooks/rules-of-hooks | 51 | 72 | 139 |

There are 929 live helper calls plus 19 synthetic-control calls, 948 comparisons total. Zero conditional-expression calls in two captured corpora is an observation, not a claim those rules cannot use the helper. Controls cover nested conditionals, defaults/destructuring, computed member names, decorators, property annotations and an index signature's parameter/return types. All helpers have real consumer observations and controls. Every parsed consumer input is represented in the capture, but only helper calls actually reached by Go produce comparison lines.

# Validation

Toolchain commands source /workspace/adamic-tools/env.sh. Test stdout and stderr go directly to log files, never pipes. Native builds use ASan/UBSan; successful driver runs require exit zero with empty stderr. The owned command wrapper enforces successful compilation/execution before stdout mismatch counts as a mutant witness.

```
bash cloud/setup.sh > stage1/cohere/lint/helpers/slot03/batch17/evidence/setup.log 2>&1
python3 stage1/cohere/lint/helpers/slot03/batch17/testdata/regenerate.py > stage1/cohere/lint/helpers/slot03/batch17/evidence/regeneration.log 2>&1
go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch17' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch17/evidence/helpers.log 2>&1
go vet ./... > stage1/cohere/lint/helpers/slot03/batch17/evidence/vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot03 > stage1/cohere/lint/helpers/slot03/batch17/evidence/format.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch17/evidence/input-oracle.log 2>&1
```

Before landing: helper gate PASS 41.568s, source/emitted/native comparison PASS 9.96s, twelve mutants PASS 31.60s. Vet and format logs are empty. Six uncached input probes PASS 1.986s, zero cache hits/six misses. Capture Go package results: core PASS 8.882s, React PASS 5.085s.

Setup timing: Go 0s, clang 0s, Node 0s, submodules 1s, cache warm 53s, done 53s; nproc 5, cpu.max 400000 100000, 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0.

The replay driver initially used an unsupported local function, then a closure refused as cycle-capable; a readonly-list assignment also exposed a typing mismatch. Moving the driver callback to module scope and preserving the reader's readonly list resolved these locally. The retained closure-refusal log is failed exploratory evidence, not a mutant kill. No compiler or harness change was made.

# Every new mutant

| Helper | Mutation | Caught at line |
|---|---|---:|
| conditionalExpression | omit condition evaluation | 335 |
| conditionalExpression | link false branch from current true-branch end instead of original test end | 335 |
| conditionalExpression | evaluate false node in true branch | 335 |
| conditionalExpression | omit final join entry | 335 |
| bindWithDefault | omit fallback fork | 462 |
| bindWithDefault | bind fallback instead of target | 15 |
| bindWithDefault | omit join entry before binding | 462 |
| memberHeader | omit member decorators | 1 |
| memberHeader | decorate member instead of parameter | 14 |
| memberHeader | omit computed key expression | 5 |
| memberHeader | omit property annotation, including nil annotation call | 11 |
| memberHeader | evaluate return type instead of index parameter type | 938 |

All twelve variants compile and run successfully with empty stderr; each is caught by stdout disagreement with actual Go. The following raw first differences identify exact node/block handles and cursor transitions:

```
batch17_test.go:133: compiled semantic mutant caught at line 335: got "3/new:-1:-1:0>0:-1;new:-1:-1:0>0:1;link:0:1:0>0:2;enter:1:-1:0>0:-1;expr:7256:-1:0>2:-1;link:2:-1:2>2:-1;new:-1:-1:2>2:-1;link:0:-1:2>2:3;enter:-1:-1:2>2:-1;expr:7257:-1:2>3:-1;link:3:-1:3>3:-1;enter:-1:-1:3>3:-1;" Go "1/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7256:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:0:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;enter:1:-1:3>1:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 335: got "1/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7256:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:2:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;enter:1:-1:3>1:-1;" Go "1/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7256:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:0:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;enter:1:-1:3>1:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 335: got "1/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7257:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:0:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;enter:1:-1:3>1:-1;" Go "1/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7256:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:0:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;enter:1:-1:3>1:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 335: got "3/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7256:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:0:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;" Go "1/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7256:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:0:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;enter:1:-1:3>1:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 462: got "0/bind:7852:-1:0>0:1;" Go "1/new:-1:-1:0>0:1;link:0:1:0>0:-1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7853:-1:2>2:-1;link:2:1:2>2:-1;enter:1:-1:2>1:-1;bind:7852:-1:1>1:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 15: got "0/bind:-1:-1:0>0:-1;" Go "0/bind:76:-1:0>0:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 462: got "1/new:-1:-1:0>0:1;link:0:1:0>0:-1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7853:-1:2>2:-1;link:2:1:2>2:-1;bind:7852:-1:2>1:-1;" Go "1/new:-1:-1:0>0:1;link:0:1:0>0:-1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7853:-1:2>2:-1;link:2:1:2>2:-1;enter:1:-1:2>1:-1;bind:7852:-1:1>1:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 1: got "0/" Go "0/decorators:0:-1:0>0:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 14: got "0/decorators:74:-1:0>0:-1;decorators:74:-1:0>0:-1;" Go "0/decorators:74:-1:0>0:-1;decorators:75:-1:0>0:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 5: got "0/decorators:39:-1:0>0:-1;" Go "0/decorators:39:-1:0>0:-1;expr:40:-1:0>0:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 11: got "0/decorators:65:-1:0>0:-1;" Go "0/decorators:65:-1:0>0:-1;expr:-1:-1:0>0:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 938: got "0/decorators:10880:-1:0>0:-1;expr:10882:-1:0>0:-1;expr:10882:-1:0>0:-1;" Go "0/decorators:10880:-1:0>0:-1;expr:10881:-1:0>0:-1;expr:10882:-1:0>0:-1;"
```

# Limits and landing

These are bounded helper-composition observations, not full findings/fix/suggestion corpora, arbitrary AST/graph fuzzing, or a port of dependency implementations. The full repository gate and its seventeen external stage-1 correctness comparators were not run in this bounded helper gate. No required input check was skipped, weakened, deleted or reported green; the executed helper and input-oracle gates contain no skips.

Final pre-publication fetch observed new main c7991b90 and lint area b84a9d93. The batch is committed before rebasing, and will be oracle-checked again on that base before publication. Later landing evidence below supersedes prelanding timing and bases.

# Landing verification on current main

All sixty branch commits rebased cleanly onto origin/area/stage1-lint b84a9d9314b65d3d0261ee017e233287b4f071da, containing current origin/main c7991b900362796aefd111474e65eb5398e91953. Rebased implementation a30ffd8e0f441fc56b98ee2b000b996ce99d9dbe. Comparing all owned slot paths against original implementation 5ccc9ac2 immediately after rebase produced no differences. Integration changes were accepted; none were reverted. Final fetch after the full helper gate confirms both tips unchanged and both ancestors of HEAD.

```
go test ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch17/evidence/landing-helpers.log 2>&1
go vet ./... > stage1/cohere/lint/helpers/slot03/batch17/evidence/landing-vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot03 > stage1/cohere/lint/helpers/slot03/batch17/evidence/landing-format.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch17/evidence/landing-input-oracle.log 2>&1
go test ./stage1/cohere/lint -run '^(TestEmittedJavaScriptMismatch|TestDotARename|TestCompleteSuggestionSerialization|TestSuggestionAlongsideAutomaticFix|TestWitnessScriptKind)$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch17/evidence/landing-harness.log 2>&1
```

Observed results:

- Owned fifty-one-helper package PASS 832.766s, 3,045,173 comparison lines; 181 successful stdout-mutant witnesses plus the compiled empty-stack guard mutant, 182 owned variants total. The missing-consumer coverage mutant is separate. The first historical batch compares Go/source Node/native; later batches also compare emitted JavaScript. The three new helpers compare all four implementations over 948 lines, PASS 5.54s. No claim that every historical line was compared four ways.
- Inherited shared helpers PASS 92.769s, all four compiling stdout mutants and ten message-refusal cases.
- Selected shared lint harness PASS 245.331s. The emitted-JavaScript mismatch runs a deliberately failing child test; its parent passes because the clean-running mutant is rejected. Nested FAIL text in that raw log is expected proof, not a failed gate. The second suggestion edit changes range 9..10 to 9..11 and is caught on source Node, emitted JavaScript and native. `.ts`/`.a` rename, automatic-fix/suggestion separation and script-kind checks pass.
- Repository-wide vet and formatting exit zero with empty logs.
- All six uncached input probes PASS 19.319s, zero hits and six misses.

The bounded gate does not execute the full repository's seventeen external stage-1 correctness comparisons. No executed helper/input/harness gate skipped a required-input correctness check. No shared checks were changed. Every new helper has real-Go consumer and control observations; whole-rule findings and dependency implementations remain outside this unit.

Every compiling mutant witness from the final landing run is named below, including inherited helpers, earlier batches and the new batch. The earlier batches' source mutations are described in their existing reports and test tables. Every successful stdout variant reaches the ordinary comparison only after clean compilation and execution; the empty-stack variant is specifically caught by expected panic/exit behavior rather than stdout.

```
TestHelperMutants/options_json.ts
helpers_test.go:145: compiled semantic mutant caught at output line 15157: got "valid", Go "invalid"
TestHelperMutants/option_schema.ts
helpers_test.go:145: compiled semantic mutant caught at output line 15166: got "valid", Go "invalid"
TestHelperMutants/policy_message.ts
helpers_test.go:145: compiled semantic mutant caught at output line 22166: got "This throws a bare `{{constructor}}`, which names no declared failure. Raise it through the tier that declares it, `AccountModule.error(identifier, data, cause)`, `ApiWorker.error(...)` or `Base.error(...)`. A bare throw carries no identifier, so the board groups it by its message and one interpolated value mints one identity per value, and it normalizes to 500, so a refusal reads as our fault.", Go "This throws a bare `sentinel é😀`, which names no declared failure. Raise it through the tier that declares it, `AccountModule.error(identifier, data, cause)`, `ApiWorker.error(...)` or `Base.error(...)`. A bare throw carries no identifier, so the board groups it by its message and one interpolated value mints one identity per value, and it normalizes to 500, so a refusal reads as our fault."
TestHelperMutants/strict_options.ts
helpers_test.go:145: compiled semantic mutant caught at output line 15178: got "valid", Go "invalid"
TestBatch10Mutants/is_url.a
batch10_test.go:91: compiled semantic mutant caught at line 79: got "true" Go "false"
TestBatch10Mutants/is_url.a#01
batch10_test.go:91: compiled semantic mutant caught at line 916: got "true" Go "false"
TestBatch10Mutants/is_absolute_size.a
batch10_test.go:91: compiled semantic mutant caught at line 1046: got "false" Go "true"
TestBatch10Mutants/is_absolute_size.a#01
batch10_test.go:91: compiled semantic mutant caught at line 65: got "true" Go "false"
TestBatch10Mutants/is_relative_size.a
batch10_test.go:91: compiled semantic mutant caught at line 783: got "false" Go "true"
TestBatch11Mutants/is_generic_name.a
batch11_test.go:93: compiled semantic mutant caught at line 549: got "false" Go "true"
TestBatch11Mutants/is_generic_name.a#01
batch11_test.go:93: compiled semantic mutant caught at line 101: got "true" Go "false"
TestBatch11Mutants/has_math_function.a
batch11_test.go:93: compiled semantic mutant caught at line 38: got "false" Go "true"
TestBatch11Mutants/has_math_function.a#01
batch11_test.go:93: compiled semantic mutant caught at line 60: got "false" Go "true"
TestBatch11Mutants/has_math_function.a#02
batch11_test.go:93: compiled semantic mutant caught at line 88: got "true" Go "false"
TestBatch11Mutants/loaded_utilities.a
batch11_test.go:93: compiled semantic mutant caught at line 1115: got "-1/false/true" Go "0/true/true"
TestBatch11Mutants/loaded_utilities.a#01
batch11_test.go:93: compiled semantic mutant caught at line 1115: got "0/true/false" Go "0/true/true"
TestBatch12Mutants/is_angle.a
batch12_test.go:120: compiled semantic mutant caught at line 1: got "false/suffix:deg,rad,grad;" Go "false/suffix:deg,rad,grad,turn;"
TestBatch12Mutants/is_angle.a#01
batch12_test.go:120: compiled semantic mutant caught at line 1: got "true/suffix:deg,rad,grad,turn;" Go "false/suffix:deg,rad,grad,turn;"
TestBatch12Mutants/is_number.a
batch12_test.go:120: compiled semantic mutant caught at line 2: got "true/scan;" Go "false/scan;math;"
TestBatch12Mutants/is_number.a#01
batch12_test.go:120: compiled semantic mutant caught at line 92: got "true/scan;" Go "false/scan;math;"
TestBatch12Mutants/is_number.a#02
batch12_test.go:120: compiled semantic mutant caught at line 2: got "false/math;scan;math;" Go "false/scan;math;"
TestBatch12Mutants/is_percentage.a
batch12_test.go:120: compiled semantic mutant caught at line 3: got "false/suffix:percent;math;" Go "false/suffix:%;math;"
TestBatch12Mutants/is_percentage.a#01
batch12_test.go:120: compiled semantic mutant caught at line 3: got "false/suffix:%;" Go "false/suffix:%;math;"
TestBatch12Mutants/is_percentage.a#02
batch12_test.go:120: compiled semantic mutant caught at line 3: got "false/math;suffix:%;math;" Go "false/suffix:%;math;"
TestBatch12ArgumentMutants/is_angle.a
batch12_test.go:166: compiled semantic mutant caught at line 1: got "false/suffix:deg,rad,grad,turn;wrong-value;" Go "false/suffix:deg,rad,grad,turn;"
TestBatch12ArgumentMutants/is_number.a
batch12_test.go:166: compiled semantic mutant caught at line 2: got "false/scan;wrong-value;math;" Go "false/scan;math;"
TestBatch12ArgumentMutants/is_number.a#01
batch12_test.go:166: compiled semantic mutant caught at line 2: got "false/scan;math;wrong-value;" Go "false/scan;math;"
TestBatch13Mutants/matches_data_type.a
batch13_test.go:228: compiled semantic mutant caught at line 1: got "false/check:1;" Go "false/check:0;"
TestBatch13Mutants/matches_data_type.a#01
batch13_test.go:228: compiled semantic mutant caught at line 2: got "false/check:2;" Go "false/check:1;"
TestBatch13Mutants/matches_data_type.a#02
batch13_test.go:228: compiled semantic mutant caught at line 3: got "false/check:3;" Go "false/check:2;"
TestBatch13Mutants/matches_data_type.a#03
batch13_test.go:228: compiled semantic mutant caught at line 4: got "false/check:4;" Go "false/check:3;"
TestBatch13Mutants/matches_data_type.a#04
batch13_test.go:228: compiled semantic mutant caught at line 5: got "false/check:5;" Go "false/check:4;"
TestBatch13Mutants/matches_data_type.a#05
batch13_test.go:228: compiled semantic mutant caught at line 6: got "false/check:6;" Go "false/check:5;"
TestBatch13Mutants/matches_data_type.a#06
batch13_test.go:228: compiled semantic mutant caught at line 7: got "false/check:7;" Go "false/check:6;"
TestBatch13Mutants/matches_data_type.a#07
batch13_test.go:228: compiled semantic mutant caught at line 8: got "false/check:8;" Go "false/check:7;"
TestBatch13Mutants/matches_data_type.a#08
batch13_test.go:228: compiled semantic mutant caught at line 9: got "false/check:9;" Go "false/check:8;"
TestBatch13Mutants/matches_data_type.a#09
batch13_test.go:228: compiled semantic mutant caught at line 10: got "false/check:10;" Go "false/check:9;"
TestBatch13Mutants/matches_data_type.a#10
batch13_test.go:228: compiled semantic mutant caught at line 11: got "true/check:11;" Go "false/check:10;"
TestBatch13Mutants/matches_data_type.a#11
batch13_test.go:228: compiled semantic mutant caught at line 12: got "false/check:12;" Go "true/check:11;"
TestBatch13Mutants/matches_data_type.a#12
batch13_test.go:228: compiled semantic mutant caught at line 13: got "false/check:13;" Go "false/check:12;"
TestBatch13Mutants/matches_data_type.a#13
batch13_test.go:228: compiled semantic mutant caught at line 14: got "false/check:14;" Go "false/check:13;"
TestBatch13Mutants/matches_data_type.a#14
batch13_test.go:228: compiled semantic mutant caught at line 15: got "false/check:15;" Go "false/check:14;"
TestBatch13Mutants/matches_data_type.a#15
batch13_test.go:228: compiled semantic mutant caught at line 16: got "false/check:16;" Go "false/check:15;"
TestBatch13Mutants/matches_data_type.a#16
batch13_test.go:228: compiled semantic mutant caught at line 17: got "false/check:0;" Go "false/check:16;"
TestBatch13Mutants/matches_data_type.a#17
batch13_test.go:228: compiled semantic mutant caught at line 18: got "false/check:0;" Go "false/"
TestBatch13Mutants/matches_data_type.a#18
batch13_test.go:228: compiled semantic mutant caught at line 1: got "false/check:0;wrong-value;" Go "false/check:0;"
TestBatch13Mutants/infer_data_type.a
batch13_test.go:228: compiled semantic mutant caught at line 29879: got "/type:color;check:0;type:length;check:1;type:percentage;check:2;type:ratio;check:3;type:number;check:4;type:integer;check:5;type:url;check:6;type:position;check:7;type:bg-size;check:8;type:line-width;check:9;type:image;check:10;type:family-name;check:11;type:generic-name;check:12;type:absolute-size;check:13;type:relative-size;check:14;type:angle;check:15;type:vector;check:16;" Go "/"
TestBatch13Mutants/infer_data_type.a#01
batch13_test.go:228: compiled semantic mutant caught at line 14279: got "/" Go "length/type:color;check:0;type:length;check:1;"
TestBatch13Mutants/infer_data_type.a#02
batch13_test.go:228: compiled semantic mutant caught at line 23: got "family-name/type:vector;check:16;type:angle;check:15;type:relative-size;check:14;type:absolute-size;check:13;type:generic-name;check:12;type:family-name;check:11;" Go "family-name/type:color;check:0;type:length;check:1;type:percentage;check:2;type:ratio;check:3;type:number;check:4;type:integer;check:5;type:url;check:6;type:position;check:7;type:bg-size;check:8;type:line-width;check:9;type:image;check:10;type:family-name;check:11;"
TestBatch13Mutants/infer_data_type.a#03
batch13_test.go:228: compiled semantic mutant caught at line 23: got "family-name/type:color;wrong-value;check:0;wrong-value;type:length;wrong-value;check:1;wrong-value;type:percentage;wrong-value;check:2;wrong-value;type:ratio;wrong-value;check:3;wrong-value;type:number;wrong-value;check:4;wrong-value;type:integer;wrong-value;check:5;wrong-value;type:url;wrong-value;check:6;wrong-value;type:position;wrong-value;check:7;wrong-value;type:bg-size;wrong-value;check:8;wrong-value;type:line-width;wrong-value;check:9;wrong-value;type:image;wrong-value;check:10;wrong-value;type:family-name;wrong-value;check:11;wrong-value;" Go "family-name/type:color;check:0;type:length;check:1;type:percentage;check:2;type:ratio;check:3;type:number;check:4;type:integer;check:5;type:url;check:6;type:position;check:7;type:bg-size;check:8;type:line-width;check:9;type:image;check:10;type:family-name;check:11;"
TestBatch13Mutants/is_family_name.a
batch13_test.go:228: compiled semantic mutant caught at line 5328: got "true/segment:,;" Go "false/segment:,;"
TestBatch13Mutants/is_family_name.a#01
batch13_test.go:228: compiled semantic mutant caught at line 14304: got "false/segment:,;" Go "true/segment:,;"
TestBatch13Mutants/is_family_name.a#02
batch13_test.go:228: compiled semantic mutant caught at line 48: got "false/segment:,;" Go "true/segment:,;"
TestBatch13Mutants/is_family_name.a#03
batch13_test.go:228: compiled semantic mutant caught at line 29904: got "true/segment:,;" Go "false/segment:,;"
TestBatch13Mutants/is_family_name.a#04
batch13_test.go:228: compiled semantic mutant caught at line 48: got "true/segment:;;" Go "true/segment:,;"
TestBatch13Mutants/is_family_name.a#05
batch13_test.go:228: compiled semantic mutant caught at line 48: got "true/segment:,;wrong-value;" Go "true/segment:,;"
TestBatch14Mutants/is_line_width.a
batch14_test.go:135: compiled semantic mutant caught at line 955: got "false/segment: ;length:0;number:0;length:1;number:1;" Go "true/segment: ;length:0;number:0;length:1;number:1;"
TestBatch14Mutants/is_line_width.a#01
batch14_test.go:135: compiled semantic mutant caught at line 1: got "false/segment: ;number:0;length:0;number:0;" Go "false/segment: ;length:0;number:0;"
TestBatch14Mutants/is_line_width.a#02
batch14_test.go:135: compiled semantic mutant caught at line 1: got "false/segment: ;length:-1;number:0;" Go "false/segment: ;length:0;number:0;"
TestBatch14Mutants/is_line_width.a#03
batch14_test.go:135: compiled semantic mutant caught at line 1: got "false/segment: ;wrong-value;length:0;number:0;" Go "false/segment: ;length:0;number:0;"
TestBatch14Mutants/is_line_width.a#04
batch14_test.go:135: compiled semantic mutant caught at line 1: got "false/segment:,;length:0;number:0;" Go "false/segment: ;length:0;number:0;"
TestBatch14Mutants/is_line_width.a#05
batch14_test.go:135: compiled semantic mutant caught at line 1: got "true/segment: ;length:0;number:0;" Go "false/segment: ;length:0;number:0;"
TestBatch14Mutants/is_image.a
batch14_test.go:135: compiled semantic mutant caught at line 2426: got "false/segment:,;" Go "false/segment:,;url:0;"
TestBatch14Mutants/is_image.a#01
batch14_test.go:135: compiled semantic mutant caught at line 7484: got "true/segment:,;" Go "false/segment:,;"
TestBatch14Mutants/is_image.a#02
batch14_test.go:135: compiled semantic mutant caught at line 5321: got "false/segment:,;url:0;" Go "true/segment:,;url:0;"
TestBatch14Mutants/is_image.a#03
batch14_test.go:135: compiled semantic mutant caught at line 5288: got "false/segment:,;url:0;" Go "true/segment:,;url:0;"
TestBatch14Mutants/is_image.a#04
batch14_test.go:135: compiled semantic mutant caught at line 3674: got "false/segment:,;url:0;" Go "true/segment:,;url:0;"
TestBatch14Mutants/is_image.a#05
batch14_test.go:135: compiled semantic mutant caught at line 2099: got "true/segment:,;url:0;" Go "false/segment:,;url:0;"
TestBatch14Mutants/is_image.a#06
batch14_test.go:135: compiled semantic mutant caught at line 2: got "false/segment:,;url:-1;" Go "false/segment:,;url:0;"
TestBatch14Mutants/is_image.a#07
batch14_test.go:135: compiled semantic mutant caught at line 2: got "false/segment:,;wrong-value;url:0;" Go "false/segment:,;url:0;"
TestBatch14Mutants/is_image.a#08
batch14_test.go:135: compiled semantic mutant caught at line 113: got "false/segment:,;url:0;url:1;" Go "false/segment:,;url:0;"
TestBatch14Mutants/is_image.a#09
batch14_test.go:135: compiled semantic mutant caught at line 6473: got "false/segment:,;url:0;" Go "true/segment:,;url:0;"
TestBatch14Mutants/is_background_position.a
batch14_test.go:135: compiled semantic mutant caught at line 1308: got "false/segment: ;length:0;percentage:0;length:1;percentage:1;" Go "true/segment: ;length:0;percentage:0;"
TestBatch14Mutants/is_background_position.a#01
batch14_test.go:135: compiled semantic mutant caught at line 1365: got "false/segment: ;length:0;percentage:0;length:1;percentage:1;" Go "true/segment: ;length:0;percentage:0;"
TestBatch14Mutants/is_background_position.a#02
batch14_test.go:135: compiled semantic mutant caught at line 7485: got "true/segment: ;" Go "false/segment: ;"
TestBatch14Mutants/is_background_position.a#03
batch14_test.go:135: compiled semantic mutant caught at line 3: got "false/segment: ;length:0;" Go "false/segment: ;length:0;percentage:0;"
TestBatch14Mutants/is_background_position.a#04
batch14_test.go:135: compiled semantic mutant caught at line 3: got "false/segment: ;percentage:0;length:0;percentage:0;" Go "false/segment: ;length:0;percentage:0;"
TestBatch14Mutants/is_background_position.a#05
batch14_test.go:135: compiled semantic mutant caught at line 3: got "false/segment: ;length:-1;percentage:0;" Go "false/segment: ;length:0;percentage:0;"
TestBatch14Mutants/is_background_position.a#06
batch14_test.go:135: compiled semantic mutant caught at line 3: got "false/segment:,;length:0;percentage:0;" Go "false/segment: ;length:0;percentage:0;"
TestBatch15Mutants/push_jump.a
batch15_test.go:126: compiled semantic mutant caught at line 90: got "1/2:1:0:false:false:/labels:0;" Go "1/2:1:0:true:false:/labels:0;"
TestBatch15Mutants/push_jump.a#01
batch15_test.go:126: compiled semantic mutant caught at line 90: got "1/2:1:0:true:true:/labels:0;" Go "1/2:1:0:true:false:/labels:0;"
TestBatch15Mutants/push_jump.a#02
batch15_test.go:126: compiled semantic mutant caught at line 90: got "1/2:1:0:true:false:/labels:1;" Go "1/2:1:0:true:false:/labels:0;"
TestBatch15Mutants/push_jump.a#03
batch15_test.go:126: compiled semantic mutant caught at line 90: got "1/1:2:0:true:false:/labels:0;" Go "1/2:1:0:true:false:/labels:0;"
TestBatch15Mutants/push_jump.a#04
batch15_test.go:126: compiled semantic mutant caught at line 90: got "1/2:1:-1:true:false:/labels:0;" Go "1/2:1:0:true:false:/labels:0;"
TestBatch15Mutants/pop_jump.a
batch15_test.go:126: compiled semantic mutant caught at line 125: got "0/" Go "1/2:1:6:true:false:/"
TestBatch15Mutants/pop_jump.a#01
batch15_test.go:126: compiled semantic mutant caught at line 92: got "1/2:1:0:true:false:/" Go "0/"
TestBatch15Mutants/make_unreachable.a
batch15_test.go:126: compiled semantic mutant caught at line 6770: got "0:false:false/1,/true,|1:false:false//nil" Go "0:false:false/1,/false,|1:false:false//nil"
TestBatch15Mutants/make_unreachable.a#01
batch15_test.go:126: compiled semantic mutant caught at line 1: got "2:true:true/2,/nil|3:false:false//nil" Go "2:true:true/3,/nil|3:false:false//nil"
TestBatch15Mutants/make_unreachable.a#02
batch15_test.go:126: compiled semantic mutant caught at line 1: got "2:true:true/3,/nil|2:true:true/3,/nil" Go "2:true:true/3,/nil|3:false:false//nil"
TestBatch15Mutants/make_unreachable.a#03
batch15_test.go:126: compiled semantic mutant caught at line 1: got "2:true:true/3,/nil|3:true:false//nil" Go "2:true:true/3,/nil|3:false:false//nil"
TestBatch15EmptyStackRefusal
batch15_test.go:189: compiled empty-stack guard mutant caught: baseline exit 70, mutant exit 0 and survived
TestBatch16Mutants/make_return.a
batch16_test.go:137: compiled semantic mutant caught at line 931: got "0:false//return:-1/;final:0/;unreachable:0/;" Go "0:false//"
TestBatch16Mutants/make_return.a#01
batch16_test.go:137: compiled semantic mutant caught at line 1: got "2:true//return:-1/;final:2/;" Go "3:false//return:-1/;final:2/;unreachable:2/;"
TestBatch16Mutants/make_return.a#02
batch16_test.go:137: compiled semantic mutant caught at line 2: got "4:false/false:true,/return:0/false:true,;link:3:2/false:true,;unreachable:3/false:true,;" Go "4:false/true:true,/return:0/false:true,;link:3:2/true:true,;unreachable:3/true:true,;"
TestBatch16Mutants/make_return.a#03
batch16_test.go:137: compiled semantic mutant caught at line 2: got "4:false/true:true,/return:0/false:true,;link:3:3/true:true,;unreachable:3/true:true,;" Go "4:false/true:true,/return:0/false:true,;link:3:2/true:true,;unreachable:3/true:true,;"
TestBatch16Mutants/make_return.a#04
batch16_test.go:137: compiled semantic mutant caught at line 1: got "3:false//return:-1/;final:3/;unreachable:2/;" Go "3:false//return:-1/;final:2/;unreachable:2/;"
TestBatch16Mutants/make_throw.a
batch16_test.go:137: compiled semantic mutant caught at line 7: got "1:false/true:true,/throw:0/true:false,;target:0/true:true,;link:1:2/true:true,;unreachable:1/true:true,;" Go "1:false/true:false,/"
TestBatch16Mutants/make_throw.a#01
batch16_test.go:137: compiled semantic mutant caught at line 3: got "5:true/true:true,/throw:0/true:true,;target:0/true:true,;link:5:2/true:true,;" Go "6:false/true:true,/throw:0/true:true,;target:0/true:true,;link:5:2/true:true,;unreachable:5/true:true,;"
TestBatch16Mutants/make_throw.a#02
batch16_test.go:137: compiled semantic mutant caught at line 3: got "6:false/true:false,/throw:0/true:true,;target:0/true:false,;link:5:2/true:false,;unreachable:5/true:false,;" Go "6:false/true:true,/throw:0/true:true,;target:0/true:true,;link:5:2/true:true,;unreachable:5/true:true,;"
TestBatch16Mutants/make_throw.a#03
batch16_test.go:137: compiled semantic mutant caught at line 3: got "6:false/true:true,/throw:0/true:true,;link:5:2/true:true,;unreachable:5/true:true,;" Go "6:false/true:true,/throw:0/true:true,;target:0/true:true,;link:5:2/true:true,;unreachable:5/true:true,;"
TestBatch16Mutants/make_throw.a#04
batch16_test.go:137: compiled semantic mutant caught at line 5: got "4:false/false:true,/throw:-1/false:true,;thrown:2/false:true,;unreachable:1/false:true,;" Go "4:false/false:true,/throw:-1/false:true,;thrown:1/false:true,;unreachable:1/false:true,;"
TestBatch16Mutants/make_yield.a
batch16_test.go:137: compiled semantic mutant caught at line 935: got "1:false//return:-1/;final:0/;throw:-1/;thrown:0/;new/;link:0:1/;enter:1/;" Go "0:false//"
TestBatch16Mutants/make_yield.a#01
batch16_test.go:137: compiled semantic mutant caught at line 1368: got "1:true/false:true,/return:0/false:false,;link:0:-1/false:false,;throw:0/false:false,;target:0/false:true,;link:0:-1/false:true,;new/false:true,;link:0:1/false:true,;enter:1/false:true,;" Go "1:true/true:true,/return:0/false:false,;link:0:-1/true:false,;throw:0/true:false,;target:0/true:true,;link:0:-1/true:true,;new/true:true,;link:0:1/true:true,;enter:1/true:true,;"
TestBatch16Mutants/make_yield.a#02
batch16_test.go:137: compiled semantic mutant caught at line 1368: got "1:true/true:true,/return:0/false:false,;link:0:0/true:false,;throw:0/true:false,;target:0/true:true,;link:0:-1/true:true,;new/true:true,;link:0:1/true:true,;enter:1/true:true,;" Go "1:true/true:true,/return:0/false:false,;link:0:-1/true:false,;throw:0/true:false,;target:0/true:true,;link:0:-1/true:true,;new/true:true,;link:0:1/true:true,;enter:1/true:true,;"
TestBatch16Mutants/make_yield.a#03
batch16_test.go:137: compiled semantic mutant caught at line 928: got "1:true//return:-1/;final:1/;throw:-1/;thrown:0/;new/;link:0:1/;enter:1/;" Go "1:true//return:-1/;final:0/;throw:-1/;thrown:0/;new/;link:0:1/;enter:1/;"
TestBatch16Mutants/make_yield.a#04
batch16_test.go:137: compiled semantic mutant caught at line 1320: got "1:true/false:false,/return:-1/false:false,;final:0/false:false,;throw:0/false:false,;target:0/false:false,;link:0:-1/false:false,;new/false:false,;link:0:1/false:false,;enter:1/false:false,;" Go "1:true/false:true,/return:-1/false:false,;final:0/false:false,;throw:0/false:false,;target:0/false:true,;link:0:-1/false:true,;new/false:true,;link:0:1/false:true,;enter:1/false:true,;"
TestBatch16Mutants/make_yield.a#05
batch16_test.go:137: compiled semantic mutant caught at line 1320: got "1:true/false:true,/return:-1/false:false,;final:0/false:false,;throw:0/false:false,;link:0:-1/false:true,;new/false:true,;link:0:1/false:true,;enter:1/false:true,;" Go "1:true/false:true,/return:-1/false:false,;final:0/false:false,;throw:0/false:false,;target:0/false:true,;link:0:-1/false:true,;new/false:true,;link:0:1/false:true,;enter:1/false:true,;"
TestBatch16Mutants/make_yield.a#06
batch16_test.go:137: compiled semantic mutant caught at line 928: got "1:true//return:-1/;final:0/;throw:-1/;thrown:1/;new/;link:0:1/;enter:1/;" Go "1:true//return:-1/;final:0/;throw:-1/;thrown:0/;new/;link:0:1/;enter:1/;"
TestBatch16Mutants/make_yield.a#07
batch16_test.go:137: compiled semantic mutant caught at line 928: got "1:true//return:-1/;final:0/;throw:-1/;thrown:0/;new/;enter:1/;" Go "1:true//return:-1/;final:0/;throw:-1/;thrown:0/;new/;link:0:1/;enter:1/;"
TestBatch16Mutants/make_yield.a#08
batch16_test.go:137: compiled semantic mutant caught at line 928: got "0:true//return:-1/;final:0/;throw:-1/;thrown:0/;new/;link:0:1/;" Go "1:true//return:-1/;final:0/;throw:-1/;thrown:0/;new/;link:0:1/;enter:1/;"
TestBatch16Mutants/make_yield.a#09
batch16_test.go:137: compiled semantic mutant caught at line 928: got "0:true//return:-1/;final:0/;throw:-1/;thrown:0/;new/;link:0:1/;enter:0/;" Go "1:true//return:-1/;final:0/;throw:-1/;thrown:0/;new/;link:0:1/;enter:1/;"
TestBatch16Mutants/make_return.a#05
batch16_test.go:137: compiled semantic mutant caught at line 2: got "4:false/false:true,/return:0/false:true,;final:3/false:true,;unreachable:3/false:true,;" Go "4:false/true:true,/return:0/false:true,;link:3:2/true:true,;unreachable:3/true:true,;"
TestBatch16Mutants/make_throw.a#05
batch16_test.go:137: compiled semantic mutant caught at line 3: got "6:false/true:true,/throw:0/true:true,;thrown:5/true:true,;unreachable:5/true:true,;" Go "6:false/true:true,/throw:0/true:true,;target:0/true:true,;link:5:2/true:true,;unreachable:5/true:true,;"
TestBatch16Mutants/make_yield.a#10
batch16_test.go:137: compiled semantic mutant caught at line 1368: got "1:true/false:true,/return:0/false:false,;final:0/false:false,;throw:0/false:false,;target:0/false:true,;link:0:-1/false:true,;new/false:true,;link:0:1/false:true,;enter:1/false:true,;" Go "1:true/true:true,/return:0/false:false,;link:0:-1/true:false,;throw:0/true:false,;target:0/true:true,;link:0:-1/true:true,;new/true:true,;link:0:1/true:true,;enter:1/true:true,;"
TestBatch16Mutants/make_yield.a#11
batch16_test.go:137: compiled semantic mutant caught at line 1320: got "1:true/false:false,/return:-1/false:false,;final:0/false:false,;throw:0/false:false,;thrown:0/false:false,;new/false:false,;link:0:1/false:false,;enter:1/false:false,;" Go "1:true/false:true,/return:-1/false:false,;final:0/false:false,;throw:0/false:false,;target:0/false:true,;link:0:-1/false:true,;new/false:true,;link:0:1/false:true,;enter:1/false:true,;"
TestBatch17Mutants/conditional_expression.a
batch17_test.go:133: compiled semantic mutant caught at line 335: got "3/new:-1:-1:0>0:-1;new:-1:-1:0>0:1;link:0:1:0>0:2;enter:1:-1:0>0:-1;expr:7256:-1:0>2:-1;link:2:-1:2>2:-1;new:-1:-1:2>2:-1;link:0:-1:2>2:3;enter:-1:-1:2>2:-1;expr:7257:-1:2>3:-1;link:3:-1:3>3:-1;enter:-1:-1:3>3:-1;" Go "1/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7256:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:0:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;enter:1:-1:3>1:-1;"
TestBatch17Mutants/conditional_expression.a#01
batch17_test.go:133: compiled semantic mutant caught at line 335: got "1/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7256:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:2:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;enter:1:-1:3>1:-1;" Go "1/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7256:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:0:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;enter:1:-1:3>1:-1;"
TestBatch17Mutants/conditional_expression.a#02
batch17_test.go:133: compiled semantic mutant caught at line 335: got "1/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7257:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:0:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;enter:1:-1:3>1:-1;" Go "1/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7256:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:0:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;enter:1:-1:3>1:-1;"
TestBatch17Mutants/conditional_expression.a#03
batch17_test.go:133: compiled semantic mutant caught at line 335: got "3/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7256:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:0:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;" Go "1/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7256:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:0:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;enter:1:-1:3>1:-1;"
TestBatch17Mutants/bind_with_default.a
batch17_test.go:133: compiled semantic mutant caught at line 462: got "0/bind:7852:-1:0>0:1;" Go "1/new:-1:-1:0>0:1;link:0:1:0>0:-1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7853:-1:2>2:-1;link:2:1:2>2:-1;enter:1:-1:2>1:-1;bind:7852:-1:1>1:-1;"
TestBatch17Mutants/bind_with_default.a#01
batch17_test.go:133: compiled semantic mutant caught at line 15: got "0/bind:-1:-1:0>0:-1;" Go "0/bind:76:-1:0>0:-1;"
TestBatch17Mutants/bind_with_default.a#02
batch17_test.go:133: compiled semantic mutant caught at line 462: got "1/new:-1:-1:0>0:1;link:0:1:0>0:-1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7853:-1:2>2:-1;link:2:1:2>2:-1;bind:7852:-1:2>1:-1;" Go "1/new:-1:-1:0>0:1;link:0:1:0>0:-1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7853:-1:2>2:-1;link:2:1:2>2:-1;enter:1:-1:2>1:-1;bind:7852:-1:1>1:-1;"
TestBatch17Mutants/member_header.a
batch17_test.go:133: compiled semantic mutant caught at line 1: got "0/" Go "0/decorators:0:-1:0>0:-1;"
TestBatch17Mutants/member_header.a#01
batch17_test.go:133: compiled semantic mutant caught at line 14: got "0/decorators:74:-1:0>0:-1;decorators:74:-1:0>0:-1;" Go "0/decorators:74:-1:0>0:-1;decorators:75:-1:0>0:-1;"
TestBatch17Mutants/member_header.a#02
batch17_test.go:133: compiled semantic mutant caught at line 5: got "0/decorators:39:-1:0>0:-1;" Go "0/decorators:39:-1:0>0:-1;expr:40:-1:0>0:-1;"
TestBatch17Mutants/member_header.a#03
batch17_test.go:133: compiled semantic mutant caught at line 11: got "0/decorators:65:-1:0>0:-1;" Go "0/decorators:65:-1:0>0:-1;expr:-1:-1:0>0:-1;"
TestBatch17Mutants/member_header.a#04
batch17_test.go:133: compiled semantic mutant caught at line 938: got "0/decorators:10880:-1:0>0:-1;expr:10882:-1:0>0:-1;expr:10882:-1:0>0:-1;" Go "0/decorators:10880:-1:0>0:-1;expr:10881:-1:0>0:-1;expr:10882:-1:0>0:-1;"
TestBatch2Mutants/intrinsic_element_named.a
batch2_test.go:101: compiled semantic mutant caught at line 24: got "true" Go "false"
TestBatch2Mutants/hole_edges.a
batch2_test.go:101: compiled semantic mutant caught at line 179230: got "false,false" Go "true,false"
TestBatch2Mutants/read_class_values.a
batch2_test.go:101: compiled semantic mutant caught at line 185959: got "Attribute::" Go "Callee::"
TestBatch3Mutants/hex_value.a
batch3_test.go:99: compiled semantic mutant caught at line 43: got "-1" Go "15"
TestBatch3Mutants/unescape_string_literal_text.a
batch3_test.go:99: compiled semantic mutant caught at line 1114212: got "0,99,111,112,121,10,:99,111,112,121," Go "0,169,10,:99,111,112,121,"
TestBatch3Mutants/parameter_nodes.a
batch3_test.go:99: compiled semantic mutant caught at line 1115385: got "2,-1,3,2" Go "-1,-1,3,2"
TestBatch3Mutants/parameter_nodes.a#01
batch3_test.go:99: compiled semantic mutant caught at line 1115375: got "0,1" Go ""
TestBatch4Mutants/escape_terminator.a
batch4_test.go:85: compiled semantic mutant caught at line 13: got "false" Go "true"
TestBatch4Mutants/followed_by_whitespace.a
batch4_test.go:85: compiled semantic mutant caught at line 3585: got "true:10" Go "false:10,11"
TestBatch4Mutants/ignored_theme_key.a
batch4_test.go:85: compiled semantic mutant caught at line 75243: got "true" Go "false"
TestBatch5Mutants/split_theme_key.a
batch5_test.go:92: compiled semantic mutant caught at line 283: got "1:" Go "0:"
TestBatch5Mutants/split_theme_key.a#01
batch5_test.go:92: compiled semantic mutant caught at line 286: got "0:" Go "1:"
TestBatch5Mutants/split_theme_key.a#02
batch5_test.go:92: compiled semantic mutant caught at line 288: got "1:99,104,97,110,103,101,100," Go "1:"
TestBatch5Mutants/join_segments.a
batch5_test.go:92: compiled semantic mutant caught at line 290: got "0," Go "0,45,"
TestBatch5Mutants/breakpoint_group_order.a
batch5_test.go:92: compiled semantic mutant caught at line 9639: got "2:true" Go "0:false"
TestBatch5Mutants/breakpoint_group_order.a#01
batch5_test.go:92: compiled semantic mutant caught at line 9641: got "23:true" Go "-17:true"
TestBatch6Mutants/at_rule.a
batch6_test.go:85: compiled semantic mutant caught at line 1: got "rule|||112,97,114,97,109,115,58,|||false|false|false|false|0|0|" Go "at-rule|||112,97,114,97,109,115,58,|||false|false|false|false|0|0|"
TestBatch6Mutants/at_rule.a#01
batch6_test.go:85: compiled semantic mutant caught at line 14: got "1:1:true:0" Go "1:1:true:-1"
TestBatch6Mutants/at_rule.a#02
batch6_test.go:85: compiled semantic mutant caught at line 14: got "0:0:true:" Go "1:1:true:-1"
TestBatch6Mutants/at_rule.a#03
batch6_test.go:85: compiled semantic mutant caught at line 3: got "at-rule|99,104,97,110,103,101,100,|99,104,97,110,103,101,100,|112,97,114,97,109,115,58,|||false|false|false|false|0|0|" Go "at-rule|||112,97,114,97,109,115,58,|||false|false|false|false|0|0|"
TestBatch6Mutants/style_rule.a
batch6_test.go:85: compiled semantic mutant caught at line 4: got "at-rule||||||false|false|false|false|0|0|" Go "rule||||||false|false|false|false|0|0|"
TestBatch6Mutants/style_rule.a#01
batch6_test.go:85: compiled semantic mutant caught at line 17: got "1:1:true:0" Go "1:1:true:-1"
TestBatch6Mutants/style_rule.a#02
batch6_test.go:85: compiled semantic mutant caught at line 17: got "0:0:true:" Go "1:1:true:-1"
TestBatch6Mutants/style_rule.a#03
batch6_test.go:85: compiled semantic mutant caught at line 6: got "rule|99,104,97,110,103,101,100,|99,104,97,110,103,101,100,||||false|false|false|false|0|0|" Go "rule||||||false|false|false|false|0|0|"
TestBatch6Mutants/variant_next_order.a
batch6_test.go:85: compiled semantic mutant caught at line 29488: got "-9007199254740989" Go "-9007199254740991"
TestBatch6Mutants/variant_next_order.a#01
batch6_test.go:85: compiled semantic mutant caught at line 29485: got "-9007199254740988" Go "-9007199254740989"
TestBatch6Mutants/variant_next_order.a#02
batch6_test.go:85: compiled semantic mutant caught at line 29486: got "false:0:-9007199254740989" Go "false:0:-9007199254740990"
TestBatch7Mutants/new_variant_registry.a
batch7_test.go:85: compiled semantic mutant caught at line 1: got "1:false:0:0:0" Go "0:false:0:0:0"
TestBatch7Mutants/new_variant_registry.a#01
batch7_test.go:85: compiled semantic mutant caught at line 25: got "84:false:0:2:2" Go "0:false:0:0:0"
TestBatch7Mutants/register.a
batch7_test.go:85: compiled semantic mutant caught at line 8: got ":84:replacement" Go ":83:replacement"
TestBatch7Mutants/register.a#01
batch7_test.go:85: compiled semantic mutant caught at line 680: got ":83:static" Go ":-17:static"
TestBatch7Mutants/register.a#02
batch7_test.go:85: compiled semantic mutant caught at line 679: got "-17:true:-17:1:0" Go "82:true:-17:1:0"
TestBatch7Mutants/attach_comparison.a
batch7_test.go:85: compiled semantic mutant caught at line 16: got "84:false:0:2:0" Go "84:false:0:2:1"
TestBatch7Mutants/attach_comparison.a#01
batch7_test.go:85: compiled semantic mutant caught at line 21: got "-8" Go "15"
TestBatch7Mutants/attach_comparison.a#02
batch7_test.go:85: compiled semantic mutant caught at line 15: got "missing" Go "-8"
TestBatch8Mutants/recursively_decode_arbitrary_values.a
batch8_test.go:85: compiled semantic mutant caught at line 413: got "function:102,110,|function:118,97,114,|word:45,45,97,95,98,|separator:44,|word:32,99,32,100,|separator:44,|function:117,114,108,|word:97,32,98,|" Go "function:102,110,|function:118,97,114,|word:45,45,97,95,98,|separator:44,|word:32,99,32,100,|separator:44,|function:117,114,108,|word:97,95,98,|"
TestBatch8Mutants/recursively_decode_arbitrary_values.a#01
batch8_test.go:85: compiled semantic mutant caught at line 71: got "function:99,97,108,99,|function:118,97,114,|word:45,45,97,32,98,|word:43,49,112,120,|" Go "function:99,97,108,99,|function:118,97,114,|word:45,45,97,95,98,|word:43,49,112,120,|"
TestBatch8Mutants/recursively_decode_arbitrary_values.a#02
batch8_test.go:85: compiled semantic mutant caught at line 714: got "unknown:97,32,117,114,108,|word:45,45,97,95,98,|word:99,92,95,100,95,101,|word:95,102,95,103,|function:102,111,111,95,98,97,114,|word:104,95,105,|" Go "unknown:97,95,117,114,108,|word:45,45,97,95,98,|word:99,92,95,100,95,101,|word:95,102,95,103,|function:102,111,111,95,98,97,114,|word:104,95,105,|"
TestBatch8Mutants/decode_arbitrary_value.a
batch8_test.go:85: compiled semantic mutant caught at line 70: got "99,97,108,99,40,49,112,120,43,50,112,120,41," Go "99,97,108,99,40,49,112,120,32,43,32,50,112,120,41,"
TestBatch8Mutants/decode_arbitrary_value.a#01
batch8_test.go:85: compiled semantic mutant caught at line 70: got "99,97,108,99,40,49,112,120,43,50,112,120,41," Go "99,97,108,99,40,49,112,120,32,43,32,50,112,120,41,"
TestBatch8Mutants/decode_arbitrary_value.a#02
batch8_test.go:85: compiled semantic mutant caught at line 30: got "117,110,101,120,112,101,99,116,101,100,32,109,97,116,104,32,105,110,112,117,116,58,117,110,101,120,112,101,99,116,101,100,32,112,97,114,115,101,32,105,110,112,117,116,58,85,82,76,40,97,32,98,41,32," Go "85,82,76,40,97,32,98,41,"
TestBatch8Mutants/register_theme_breakpoint_variants.a
batch8_test.go:85: compiled semantic mutant caught at line 851: got "49,48,48,:-17:static|101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|" Go "101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|"
TestBatch8Mutants/register_theme_breakpoint_variants.a#01
batch8_test.go:85: compiled semantic mutant caught at line 855: got "49,48,48,:-17:static|101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|" Go "101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|"
TestBatch8Mutants/register_theme_breakpoint_variants.a#02
batch8_test.go:85: compiled semantic mutant caught at line 747: got "101,120,105,115,116,105,110,103,:1:static|115,109,:2:compound|" Go "101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|"
TestBatch8Mutants/register_theme_breakpoint_variants.a#03
batch8_test.go:85: compiled semantic mutant caught at line 859: got "49,48,48,:-16:static|101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|" Go "49,48,48,:-17:static|101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|"
TestBatch9Mutants/parse_css.a
batch9_test.go:115: compiled semantic mutant caught at line 31: got "ok:[rule:46,97,:::::false:false:true:[declaration::::99,111,108,111,114,:114,101,100,:false:true:false:[]|]|]" Go "ok:[comment:::::33,32,108,105,99,101,110,115,101,32,:false:false:false:[]|rule:46,97,:::::false:false:true:[declaration::::99,111,108,111,114,:114,101,100,:false:true:false:[]|]|]"
TestBatch9Mutants/parse_css.a#01
batch9_test.go:115: compiled semantic mutant caught at line 418: got "ok:[rule:117,110,101,120,112,101,99,116,101,100,32,116,114,105,109,58,65279,46,97,32,:::::false:false:true:[declaration::::99,111,108,111,114,:114,101,100,:false:true:false:[]|]|]" Go "ok:[rule:46,97,:::::false:false:true:[declaration::::99,111,108,111,114,:114,101,100,:false:true:false:[]|]|]"
TestBatch9Mutants/parse_css.a#02
batch9_test.go:115: compiled semantic mutant caught at line 7: got "ok:[]" Go "ok:[declaration::::45,45,120,::false:true:false:[]|]"
TestBatch9Mutants/parse_css.a#03
batch9_test.go:115: compiled semantic mutant caught at line 21: got "ok:[rule:46,97,:::::false:false:true:[unexpected declaration:::::98,97,99,107,103,114,111,117,110,100,58,32,117,114,108,40,97,:false:false:false:[]|unexpected declaration:::::98,46,112,110,103,41,:false:false:false:[]|]|]" Go "ok:[rule:46,97,:::::false:false:true:[declaration::::98,97,99,107,103,114,111,117,110,100,:117,114,108,40,97,59,98,46,112,110,103,41,:false:true:false:[]|]|]"
TestBatch9Mutants/parse_css.a#04
batch9_test.go:115: compiled semantic mutant caught at line 27: got "ok:[]" Go "error:5:77,105,115,115,105,110,103,32,99,108,111,115,105,110,103,32,125,32,97,116,32,46,97,"
TestBatch9Mutants/parse_css.a#05
batch9_test.go:115: compiled semantic mutant caught at line 5: got "error:1:77,105,115,115,105,110,103,32,111,112,101,110,105,110,103,32,40," Go "error:0:77,105,115,115,105,110,103,32,111,112,101,110,105,110,103,32,40,"
TestBatch9Mutants/design_system_decline_message.a
batch9_test.go:115: compiled semantic mutant caught at line 433: got "32,99,111,117,108,100,32,110,111,116,32,98,117,105,108,100,32,116,104,105,115,32,112,114,111,106,101,99,116,39,115,32,84,97,105,108,119,105,110,100,32,100,101,115,105,103,110,32,115,121,115,116,101,109,44,32,115,111,32,105,116,32,105,115,32,114,101,112,111,114,116,105,110,103,32,110,111,116,104,105,110,103,32,114,97,116,104,101,114,32,116,104,97,110,32,114,101,112,111,114,116,105,110,103,32,97,32,99,108,101,97,110,32,116,114,101,101,58,32,60,110,105,108,62," Go "32,99,111,117,108,100,32,110,111,116,32,98,117,105,108,100,32,116,104,105,115,32,112,114,111,106,101,99,116,39,115,32,84,97,105,108,119,105,110,100,32,100,101,115,105,103,110,32,115,121,115,116,101,109,32,102,114,111,109,32,116,104,101,109,101,46,99,115,115,44,32,115,111,32,105,116,32,105,115,32,114,101,112,111,114,116,105,110,103,32,110,111,116,104,105,110,103,32,114,97,116,104,101,114,32,116,104,97,110,32,114,101,112,111,114,116,105,110,103,32,97,32,99,108,101,97,110,32,116,114,101,101,58,32,60,110,105,108,62,"
TestBatch9Mutants/design_system_decline_message.a#01
batch9_test.go:115: compiled semantic mutant caught at line 421: got "32,99,111,117,108,100,32,110,111,116,32,98,117,105,108,100,32,116,104,105,115,32,112,114,111,106,101,99,116,39,115,32,84,97,105,108,119,105,110,100,32,100,101,115,105,103,110,32,115,121,115,116,101,109,44,32,115,111,32,105,116,32,105,115,32,114,101,112,111,114,116,105,110,103,32,110,111,116,104,105,110,103,32,114,97,116,104,101,114,32,116,104,97,110,32,114,101,112,111,114,116,105,110,103,32,97,32,99,108,101,97,110,32,116,114,101,101,58,32," Go "32,99,111,117,108,100,32,110,111,116,32,98,117,105,108,100,32,116,104,105,115,32,112,114,111,106,101,99,116,39,115,32,84,97,105,108,119,105,110,100,32,100,101,115,105,103,110,32,115,121,115,116,101,109,44,32,115,111,32,105,116,32,105,115,32,114,101,112,111,114,116,105,110,103,32,110,111,116,104,105,110,103,32,114,97,116,104,101,114,32,116,104,97,110,32,114,101,112,111,114,116,105,110,103,32,97,32,99,108,101,97,110,32,116,114,101,101,58,32,60,110,105,108,62,"
TestBatch9Mutants/loaded_theme.a
batch9_test.go:115: compiled semantic mutant caught at line 734: got "-1:false:true" Go "0:true:true"
TestBatch9Mutants/loaded_theme.a#01
batch9_test.go:115: compiled semantic mutant caught at line 734: got "0:true:false" Go "0:true:true"
TestSlot03HelperMutants/component_base_name.a/value
slot03_test.go:160: compiled semantic mutant caught at line 159: got "false" Go "true"
TestSlot03HelperMutants/tailwind_space.a/value
slot03_test.go:160: compiled semantic mutant caught at line 847: got "false" Go "true"
TestSlot03HelperMutants/listener_kinds.a/value
slot03_test.go:160: compiled semantic mutant caught at line 1114948: got "JsxAttribute,CallExpression,StringLiteral" Go "JsxAttribute,CallExpression,VariableDeclaration"
TestSlot03HelperMutants/listener_kinds.a/shared-list
slot03_test.go:160: compiled semantic mutant caught at line 1114949: got "StringLiteral,CallExpression,VariableDeclaration" Go "JsxAttribute,CallExpression,VariableDeclaration"
TestConsumerCoverageRejectsMutant
slot03_test.go:277: missing-consumer mutant caught: consumer capture mismatch: missing [better-tailwindcss/no-unknown-classes] extra []
```
