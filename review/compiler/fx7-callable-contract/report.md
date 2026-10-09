Callable contract retention for 146

Readiness keeps a producer-certified callable member contract. Both backends evaluate the receiver once, check the stored member before evaluating arguments, and retain the receiver-aware call convention. Native arguments are adapted to the selected producer's representation; primitive-union signatures carry their representation masks. Checked callable spreads remain a located NotYet; ordinary receiver and rest calls keep their existing path.

Observed: p18 prints 8 and p20 prints 3 followed by 12, agreeing with Node in both backends. Receiver-binding and evaluation-order controls agree with Node. The incompatible Boolean-parameter witness stops with exit 70 and stdout before in both backends before its argument runs. Node evaluates the argument and prints 99.

Mutants: clearing the contract in readiness compiles but makes JavaScript print before/argument/99 and native print before/argument/7, where Node prints before/argument/99. Removing producer argument boxing makes the wider producer's native stdout empty instead of Node's 8. Removing its producer mask gives the same output disagreement. Removing the spread refusal causes its located-refusal fixture to fail with no error. Each final mutant failed and sources were restored.

Commands and outcomes:
- go test ./internal/lower -run '^TestCallableContract' -v -count=1 -timeout 90s: passed, 6.576s.
- go test ./internal/lower -v -count=1 -timeout 420s: passed, 191.846s.
- go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s: the initial export-build run timed out; the warmed retry passed, 1.282s.
- go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 540s -args -update-counts: passed, 206.636s.
- Selected TestReviewProgramsAgreeWithNode p18/p20: passed, 7.571s; selected TestNativeAgreesWithNode wide producer: passed, 1.049s. These include JavaScript, sanitized native and release native observations.
- Lane checks run on the committed branch before pushing; their output is in lane-checks.log.

Every new test leaf is parallel. In the full run: p18 3.73s, p20 2.86s, wider producer 2.78s, receiver binding 3.48s, receiver once 3.84s, checked failure 1.66s, spread refusal 0.35s, plain spread 3.51s.

Counts: no existing rows moved. Three absent rows were registered with allocation/free/retain/release/peak/region values: p18 4/4/5/9/4/0; p20 7/7/6/14/5/0; wider producer 5/5/6/11/5/0.

Not covered: higher-order callable contracts and optional checked calls. No runtime C or header source changed.

Setup: GOPROXY=https://proxy.golang.org|direct; bash cloud/setup.sh; source /workspace/adamic-tools/env.sh. Go and Node ready 0.039s, submodules 0.121s, markdown dependencies 0.125s, clang 0.381s, build 83.378s, cache warm 83.582s, total 83.637s. nproc=5; CPU quota=4.
