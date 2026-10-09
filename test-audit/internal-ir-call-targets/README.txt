Base origin/main: 0f57c462ba9e5489c48373a9e27ed75e2ca75f44
nproc: 5
Code under test: Go IR target analysis and lowering-produced target facts. Oracle: hand-written expectations in call_targets_test.go, self.
M1-M8.diff are independent unified diffs against that base, with no switch. Apply one at a time. switch.diff records the single compiled switch used in this session.
results.json records first-catch and whole-package commands. M2/M3/M4 panic; completion.json records every other top-level IR test run separately to complete their matrices.
Baseline coverage reached IR functions are in functions.txt; nonzero entries are the reached-function inventory.
Package-unique kills: M1/M2 descendant test; M3/M4/M5 closure-proof test; M8 TestArgumentLayouts. Survivors: M6 direct-closure target off-by-one; M7 CallMayThrow always false.
No repository-wide uniqueness claim. No external authority checked. No test, harness, oracle, or fixture was mutated. Production files restored.
