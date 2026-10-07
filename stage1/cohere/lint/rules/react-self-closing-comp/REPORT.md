Built `react/self-closing-comp` in Adamic with exact tag-option selection, protected whitespace and independent diagnostic/edit spans.
Validation: 79 captured upstream cases passed using actual Go AST snapshots on Node, emitted JavaScript and ASan/UBSan native: 6,747 identical bytes.
Mutant: selfclosing_fix_suppressed compiled and finished, then failed only output comparison on all three runtimes; test passed in 103.357s.
Shared block: the unmodified Adamic parser rejects `export const c = <div></div>;` before the listener runs, at column/offset 24 expecting GreaterThanToken.
Not covered: end-to-end JSX parsing and its findings-per-second measurement; a parser-free engine comparison is explicitly not an end-to-end pass.

The implementation uses the JSX AST layout published by origin/codex/stage1-jsx-lint: an opening element, zero or more child nodes, and a closing element. It subscribes only to JsxElement. The Go rule is unchanged, and snapshot.go.txt independently parses each captured source and runs that actual Go rule. It writes snapshot.a with the real Go tree geometry converted from UTF-8 positions to UTF-16 units, plus the expected diagnostic text, byte ranges, message IDs, repairs, edit ranges and final fixed source. snapshot.a calls the real ported rule; it does not substitute a Go finding or precomputed repair.

The snapshot checks cover all 79 unique source/options cases collected by the actual upstream tests, including namespaced/member/this tags, separate component/html switches, nil/empty options, non-ASCII whitespace and NBSP, and comment/expression children. Fixes are applied back to front on both sides. No overlapping fixes are proposed in this corpus. The rule excludes NBSP and includes the actual Go whitespace set's U+0085 and U+FEFF. The edited span begins at opening.End()-1, while the finding spans the complete opening element.

The compiler and stage1 corpus has no JSX nodes to visit in these files; this rule's clean corpus results alone would prove no positive decision. The Go AST snapshots provide the bounded positive comparison while the independent parser remains blocked. The owned mutant suppresses the fix metadata. It runs to completion with no stderr and changes its output, rather than failing compilation, parsing or sanitizers.

Reproduction after the owned upstream capture has produced /tmp/lint-wave1-10-next/captured.json:

```sh
source /workspace/adamic-tools/env.sh
# Map /workspace/adamic/cohere/adamic_selfclosing_snapshot.go to the owned snapshot.go.txt using a Go overlay.
cd cohere
go run -overlay /tmp/wave10-selfclosing-overlay.json /workspace/adamic/cohere/adamic_selfclosing_snapshot.go /tmp/lint-wave1-10-next/captured.json /workspace/adamic/stage1/cohere/lint/rules/react-self-closing-comp/snapshot.a /tmp/wave10-selfclosing-want.txt > /tmp/wave10-selfclosing-generate.log 2>&1
cd ..
python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/validate.py > /tmp/wave10-original-prepare.log 2>&1
go test -overlay /tmp/lint-wave1-10-next/overlay.json ./stage1/cohere/lint -run '^TestOriginalSelfSnapshots$' -count=1 -v -timeout 15m > /tmp/wave10-selfclosing-snapshots.log 2>&1
```

The virtual test resides in sort-vars/validation_test.go.txt. Three preliminary failures were owned driver issues: unsupported string-plus-number output construction, a Node experimental warning on stderr, and bypassing Adamic's runtime loader for emitted JavaScript. Corrected these without changing the rule or shared repository files. Final source, emitted and native comparison and mutant checks passed. Raw public fixture evidence is included adjacent to this report.

The shared harness foundation appeared on the second successful fetch at 2650ad595b82220c368631ea13139fad4b306ed6. It is recorded as an integration dependency, not hand-edited here. This branch still uses its existing isolated scratch compatibility overlay; source registration and profile compilation are shared work. No dispatch, generator, shared oracle, shared corpus or compiler file was modified by this rule unit.
