Prepared guarded replay; syntax-kinds compiler work is blocked by incompatible base prerequisites.
Base b410340dc8f889b5799c3bc519117c63def3aa24; replay merge a8bf2982873b9a4abc2330cba6f91d92f6d05d9c; evidence c5bc4a97.
Stable setup passed in 38.592s with nproc=5; completed-input entry replay exited 1, signature not reproduced.
No mutants ran: no semantic rule or fixture changed.
Zero of 127 root sites covered; all 22 kinds remain skipped under the prerequisite block.

Created codex/notyet-syntax-kinds from newest origin/area/compiler, following the unit-specific base instruction, and merged census replay 9a1f14c5. The table was measured on 44583d3283fdd8674085a7ddcce040cf2a73a94e. That compiler differs in 322 lowering files and has entire nested-function, checked-view, predicate-proof and Node fs implementations absent from the requested base. Importing it would be substantial integration beyond small kind-specific diffs. Production compiler files remain unchanged.

Read CLAUDE.md and the language/memory documents before implementation. Initial setup overlapped the census merge and failed with undefined typedArrayWrite, censusFieldSlotless and parameterProperty helpers. All exist on the completed checkout; stable setup then passed. Timings: Go 0.022s, Node 0.018s, markdown 0.068s, submodules 0.070s, clang 0.150s, build 38.458s, cache warm 38.566s, done 38.592s. Go 1.27.1, clang 20.1.8, Node v24.19.0; nproc=5, CPU quota 400000/100000. Sourced /workspace/adamic-tools/env.sh.

stage3/apply.sh /tmp/adamic-syntax-adapted exited 0. Independently hashed all 81 files from the table branch source-manifest.json: zero differences. An initial directory replay overlapped adaptation; only the completed-input entry replay is evidence:

```sh
source /workspace/adamic-tools/env.sh
go run ./stage3/census/latent/replay -project /tmp/adamic-syntax-adapted/src/tsc/tsc.ts -where /tmp/adamic-syntax-adapted/src/compiler/binder.ts:2451:21 -kind NotYet -reason 'a checked write without a reifiable source-slot type certificate' > /tmp/adamic-syntax-replay-entry.json 2> /tmp/adamic-syntax-replay-entry.log
```

Exit 1, selected bindBlockScopedDeclaration at binder.ts:2437:5. Instead observed debug.ts:189:13 NotYet assigning an element of a value, debug.ts:190:14 Refused a cast the runtime can't check, and debug.ts:213:28 NotYet a value of type unknown. Load/register/lower 2.393549739s, total 2.73620463s. This establishes failure to reproduce on identical source bytes, not that sites were lowered or echoes.

Fetched all origin/codex/notyet-* refs and checked each branch history under internal/lower relative to area/compiler. Only binary and statics had unique lowering commits at inspection: binary changes combine (e7320353, 845e49f9); statics changes callOrMethod, methodList and staticDeclaration (b6aa4f00). These do not establish ownership of every expression or class function. No kind was skipped on an invented whole-file ownership claim.

All kinds below are skipped because the initial exact replay and base prerequisites block implementation. Later kinds were not replayed or reduced. No design refusal ruling is inferred.

| Sites | Kind | Specific prerequisite or limit |
| ---: | --- | --- |
| 46 | checked write without source-slot certificate | measured checked-view write path absent |
| 25 | generic or unnamed nested function declaration | nested_functions.go absent |
| 10 | SpreadElement | not attempted after initial block |
| 9 | ModuleDeclaration | namespace implementation differs; not attempted |
| 7 | overload argument with different implementation representation | nested and phantom overload paths absent |
| 5 | overloaded function read as a value | phantom_overload_results.go absent |
| 5 | substr with a length argument | measured substr admission absent |
| 4 | first-class nested reference from another nested function | nested_functions.go absent |
| 3 | dictionary checked view | view_objects.go absent |
| 1 | ClassExpression | not attempted |
| 1 | PostfixUnaryExpression | not attempted |
| 1 | field of boolean or undefined | measured assignmentReference absent |
| 1 | optional comparator requiring recursive default string conversion | view_array_consumers.go absent |
| 1 | uninitialized object field without supported declared slot type | measured objectLiteral path differs; not attempted |
| 1 | untyped length Array escaping without proven element representation | library_array_holes.go absent |
| 1 | checked predicate overload target intersection | predicates_proof.go absent |
| 1 | generic overload result proofs | phantom_overload_results.go absent |
| 1 | node:fs.unwatchFile | Node fs lowering absent |
| 1 | node:fs.watch | Node fs lowering absent |
| 1 | node:fs.watchFile | Node fs lowering absent |
| 1 | statSync options other than fixed literal or plain const binding | library_node_fs_file.go absent |
| 1 | toLocaleTimeString in fs file host | library_node_fs_file.go absent |

No fixtures, counts rows, semantic tests or mutants added. No whole package or full gate ran. Required setup go build ./... passed, and git diff --check passed. Evidence contains stable setup and completed-input replay; replay-findings.json omits the declaration catalog but retains selected units and every finding.

Resume needs an area/compiler tip containing the measured prerequisites or authorization to incorporate the table compiler lineage. A clarification was sent while preparing this blocked report.
