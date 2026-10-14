Built: step 17 nullable strings, required boolean-or-undefined class slots, and checked stored scalar union tags.
Commits: 9dcadac7, 832cbc4c, bd6439bb; classification fa0a865d; current-main merge 4c355db1.
Commands: guard PASS 19.831s; ten fixture leaves and IR mutant PASS 4.299s; full lower PASS 53.447s; lane checks PASS 2.3s.
Mutants: all nine source mutants caught, plus the IR field-check removal mutant.
Not covered: remaining nullable-object and string mixed-field/local-tag slices; hidden never-array lowering is not duplicated.

The delivery branch is compiler/scout-unions-main. The classification of all eleven original commits, including the three earlier inventory/design/probe commits, is in [commits.md](commits.md). This delivery finishes the three rebuilt slices requested in the resumed turn. It does not close the original entire scout as superseded.

Validation after merging origin/main at 7d113268, with `/workspace/adamic-tools/env.sh` sourced and every command writing to its own log:

- `timeout 300 go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s`: PASS, 19.831s, resume-guard.log.
- `timeout 300 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/scout_(nullable_string|hir_optional|tsc_|class_boolean|selector_optional|values_optional|boxed_scalar)|TestNullableStringFieldCheckMutant' -count=1 -v -timeout 90s`: PASS, 4.299s, resume-fixtures.log. Node, JavaScript and native checks include sanitizers and release builds. Fixture leaves range from 0.10s to 3.36s; IR mutant leaf 2.62s.
- `timeout 600 go test ./internal/lower -count=1 -timeout 570s`: monitored background run, resume-lower.log. PASS, 53.447s. Previous aggregate timeouts and successful retries are documented in the individual slice reports; no compiling timeout occurred in this resumed validation.
- `timeout 900 python3 review/compiler/scout-unions-main/resume-mutants.py`: each source mutant invokes `timeout 180 go test -overlay <overlay> <package> -run <witness> -count=1 -v -timeout 90s`. The runner initially stopped on an unrecognized but valid stage 0 refusal, then continued from case 3 after correcting its parser. No compiler failure or timeout is counted as a caught mutant.

Mutants and their independent catchers:

| Mutant | What caught it |
| --- | --- |
| null-box | nullable strings fixture disagrees with Node after null becomes NULL |
| runtime-typeof | tsc nullish-content fixture disagrees with Node after forcing null hint |
| lower-typeof | tsc nullish-content fixture fails with stage 0 refusing typeof null plus undefined |
| nullable-admission | TestNullableStringAdmission rejects the introduced nullable-string refusal |
| equality-observation | nullable strings fixture panics on an equality observation |
| class-zero | class boolean fixture disagrees with Node after undefined becomes false |
| class-optional | TestScoutClassBooleanKeepsUnsafeViewsRefused catches optional neighbor admission |
| boxed-check | stale scalar field fixture catches wrong native output and absent checked panic |
| boxed-observation | scalar fields fixture catches panic on strict equality observation |
| IR field-check removal | TestNullableStringFieldCheckMutant catches unchecked stale nullable-string read |

Mutant sources are `.go.txt`, regenerated from current production files. Their overlays and logs carry the resume prefix. No mutant production code is committed. Ten new counts rows were recorded across the three source commits; no existing numerical row changed. Prior count-refresh commands and package results are in [nullable-strings.md](nullable-strings.md), [class-fields.md](class-fields.md), and [boxed-fields.md](boxed-fields.md).

Unfinished nullable-object work is preserved locally in stash bbca9651b3ec40524f0bbf365c2ed54320d456c6, including its three fixtures and mutant evidence. It is excluded from this delivery as requested by the resumed scope. Original commits 7b270f77 and 3230854e still require delivery; hidden never-array work remains a named dependency. No whole gate, whole oracle package, or corpus byte credit is claimed. No PR was opened.

Toolchain setup: initial bounded setup expired during cold compilation; retry completed in 70.525s (Go 0.098s, Node 0.085s, clang 0.458s, submodules 0.327s, cache 70.410s). nproc=5, CPU quota=4. stage3/api dependencies were installed from its lockfile to enable the full lowering suite; setup evidence is adjacent.

Integration lane command: `git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -`. PASS: lane checks 2.3 s, gofmt and tools on 15 Go files, t.Parallel on 4 test packages, vet 4 packages. Main remained 7d113268. Output: resume-lane.log.
