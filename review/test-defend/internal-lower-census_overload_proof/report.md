Both requested rows are defended in the full default internal/lower package matrix.
Base: 7b9d4272c28f59530ab13daa5c49067e47933b06. Scope comes from go test -list . ./internal/lower/.
276 top-level tests, including 37 added since the audit. Each unique mutant fails one row, passes 273, and skips two.

Code under test and oracle

TestCensusRestMutableElements calls Lower through lowerSource. The code under test is lowering.refuseWidening and its contextual mutable-property relation, reached for spread arguments. Its oracle is self: a handwritten expectation of the Refused diagnostic class. The clean run actually refused the spread at main.a:8:9, DogHouse seen as House. Its named subsumer exercises arrays, properties, maps, ordinary calls, copies, generics and object spreads, but no rest-call SpreadElement input.

TestCensusBooleanDeadBranch calls Lower through lowersAndAgreesWithNode. The code under test is lowering.censusBooleanLogical and combine, lowering flow-narrowed never Boolean expressions. Its current oracle is external-run: Node runs both the source and emitted JavaScript, compares stdout and exit, and checks stderr. The source prints done. This row has strengthened since the audit, which checked only successful lowering. Dead arms must still lower even though their values cannot affect stdout.

Coverage evidence

Commands: timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run '^NAME$' -coverpkg=./internal/lower -coverprofile=/tmp/defend-census/NAME.cover > NAME.log 2>&1.
Names: TestCensusRestMutableElements, TestAMutableLocationSeenWiderIsRefused, TestCensusBooleanDeadBranch.
The Boolean comparator used the rest of the package: -run '^Test' -skip '^TestCensusBooleanDeadBranch$' with identical coverage options, saved as rest.cover and rest.log.gz.
coverage-differences.json lists all exclusive blocks. Rest has six relative to its named subsumer; R1 instead targets a semantic distinction on shared code: spread arguments versus ordinary arguments. Boolean has two blocks exclusive against the rest of the package, census_small.go:380 and :383. B1 and B2 target :383.

Mutants and validation

R1: invariance.go:457, replace ParenthesizedExpression with SpreadElement in the early-skip condition. Only TestCensusRestMutableElements fails: census_small_test.go:72: a rest copy still shares mutable elements; expected refusal, got <nil>. The original named subsumer passes.
B1: census_small.go:383, change ir.Boolean to ir.String. No row fails. Equivalent candidate: this value is only compared with MaybeBoolean subsequently; String and Boolean both choose the same operand path. Binary.Type derives the result type from its operator and operand types, not this local. No differing output was demonstrated; do not label unguarded behavior.
B2: census_small.go:383, return ir.Binary{Operator: ir.Add, Left: left, Right: right}, true early. This is a return-early mutant replacing the fallback assignment, not an empty-answer probe. It prematurely returns a numeric expression instead of selecting the logical operator. Only TestCensusBooleanDeadBranch fails: census_small_test.go:84: Lower refused acceptance row: main.a:5:6: stage 0 can't lower a value of type never yet. All other runnable rows pass.

Each standalone diff was validated while applied with go vet ./internal/lower/ and, after restoration, git apply --check. Full matrix commands, complete passed-row lists, failure output, isolated cache paths and wall seconds are in matrix.json. Logs are gzip-compressed without filtering. Production code and tests are restored. restored.log.gz confirms both requested rows pass after restoration.

Limits and brief issues

The initial disk check found /tmp with 8.7 GB free and /workspace with 18 GB free. Only the earlier named /tmp/defend-nonnull scratch directory was removed. /tmp's entire filesystem is 8.8 GB, so the requested 15 GB free threshold cannot be attained there. It was nearly empty and no run failed for disk space.
The supplied audit evidence is historical. Its baseline used 7b18d0576930caca4e22ce2eef92fcf563af52d0; current main changes both scope and the Boolean row's assertion. We fetched the audit with the full refspec and read report.md, plan.md, mutants.md and rows.json before selecting mutants.
TestOriginalCycleLedger and TestOptionalWideningCensus skipped by default. One existing TestMixedUnionContractGraph subcase also skipped. Their mutant results are unknown. Uniqueness here means among the default runnable package rows, including all 37 newly added tests. No package narrowing, panic recovery, timeout, twin exception or cost exception was needed.
Both rows were defended, so no unsupported name-versus-assertion finding is asserted for an undefended row. The rest row asserts diagnostic class rather than exact reason; its observed clean reason matches its named rest-element check. The Boolean row now checks executable agreement, but its dead arms still cannot distinguish wrong Boolean answers if lowering and the live output remain unchanged.

Cost and coverage

Toolchain setup skipped, warm env.sh worked; nproc 5. npm ci installed three stage3/api packages in 405 ms. Baseline package binary: 31.193 s. Coverage package binaries: rest target 0.053 s, named subsumer 0.531 s, Boolean target 0.169 s, all other rows 31.155 s. Mutant wall durations including vet, Go compilation and package execution: R1 38.623 s, B1 38.563 s, B2 38.399 s. Separate native build caches were used for every mutant. Build time is not separately isolated. No other packages, opt-in corpus matrices or exhaustive mutants were covered. The first unique witness suffices, so three attempts were not required for either defended row.
