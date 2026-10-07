Built: migrated no-label-var to visit(node,index), removed its node scan/relevance comparison and eliminated scope-query target refetch.
Commits: follows pushed 6075e36ee on wave-09, based on current main b8fb957a; no new claims.
Commands/output: label_var/verify.py, label_var/verify_handed.py and listeners/verify.py PASS; full Go controls/corpora, sanitizers and released handles match.
Mutants: scope membership, scope value-mask, handle retention and wrong kind reran; new clean-running node-refetch mutant is caught at byte 43.
Not covered: shared registry installation, source/emitted Node checker bridge, one parser-refused control, complete regex rules or a full repository gate.

NoLabelVar now acts only on the supplied ParseNode. It has no run loop and
does not read node.kind or compare kinds for relevance. Reading the label's
child identifier is a semantic child lookup, not a refetch of the handed node.
The isolated driver switches on kind and invokes visit only for labeled
statements, fetching each candidate once. The owned rule.json uses the verified
LabeledStatement kind name and node:true. The descriptor remains metadata
for integration; this commit does not invent a shared checker context.

ScopeValueSymbols now receives that same node and calls the existing raw
tsgoInspect ABI with its span, the declared LabeledStatement kind and the raw
scope-value-symbols question. It increments the existing query counter and
retains the exact framed fact decoder. No Rules.ask call or target node
refetch occurs on this path. No new checker judgment, ABI, registration or
shared file changed. Existing raw question registration is still validated
via the isolated Go overlay, pending shared integrator installation.

Reproduce after sourcing /workspace/adamic-tools/env.sh, redirecting each test
to its own log file:

```
python3 stage1/cohere/typeaware/wave09_core/label_var/verify.py > /workspace/wave-09-handed-label-test.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/label_var/verify_handed.py > /workspace/wave-09-handed-contract-test.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/listeners/verify.py > /workspace/wave-09-handed-listeners-test.log 2>&1
```

The migrated rule matches production Go on 34 supported controls, 18 findings,
7700 bytes; 77 compiler sources, zero findings, 7859 bytes; and 287 repository
sources, zero findings, 18485 bytes. These include normal native and
ASan/UBSan/LeakSanitizer. The same undefined-label fixture remains parser-
refused and is recorded, not silently removed from coverage. Membership and
value-mask mutants compile/run and fail only the independent Go comparison.
Released handles still panic 70; the retaining mutant removes that required
panic and is caught. Named descriptor checks pass Go/native/sanitizers and
both Node modes; the two wrong-kind mutations are caught at byte 276.

The new handed-node check supplies callback index zero for every visit while
preserving each actual handed node. Normal and sanitized executions still
produce the full 7700 Go bytes. A mutant replaces the handed node with
parser.node(_index); it compiles and exits zero with empty stderr, then
differs from Go at byte 43. This proves the contract check detects a target
refetch rather than merely certifying output on matching node/index pairs.
Go timing stderr is retained separately; native stderr must be empty.

Three-round native/Go medians: compiler 2.039702/0.743248 seconds,
repository 0.298251/0.230361 seconds. These are whole-process comparisons
under concurrent work, not isolated throughput or a claimed measured speedup.
Setup is reused from its successful 88-second run; nproc 5.
Full streams, generated mutant probes and measurements are retained in
validation-handed-label, with executable probes archived outside module
discovery. No full gate or unrelated rule suite was rerun.

The updated named harness 41eb6eab2 was inspected: it adds the batch runner
migration and dedup ledger on the previously reviewed node:true API. It is
not on current main. No shared source was modified. Both regex claims remain
incomplete, independently blocked on dynamic RegExp lowering and remaining
visitor/tracker work. No hand-rolled matcher or new regex parser was added.
No React parking applies and no new batch was claimed. Publication is only
to codex/typeaware-wave-09.
