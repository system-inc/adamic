Added isolated numeric supplied-fact verdict modules for symbol-description and structure/react-hook-no-any-type, plus require-await body filters.
Reservation was pushed in f7df1008 before implementation; previous complete six-rule landing is 6d4e8cfc on main f8013f0b.
Owned validate.py passes production Go controls, native sanitizers, clean-exit mutants and explicit missing-adapter refusal; production source lint finds 0.
Symbol/hook mutants fail byte comparison at 60/64; require-await filter mutant fails its narrower Boolean comparison at 568.
These are partial ports, not complete native rules: live supplied-node adapters, corpus comparison and require-await reporting remain unimplemented.

The three earlier React claims are PARKED under the user's explicit instruction.
The first six rules remain completed; their full landing observations are in
../wave_10_react/README.md. No additional rule is claimed beyond these three.

Each directory has rule.json listener metadata using canonical typescript-go
ast.Kind names and "node": true, as specified by the registry at harness commit
ab70f38d4. Symbol and hook listeners declare CallExpression; require-await declares
FunctionDeclaration, FunctionExpression, ArrowFunction, MethodDeclaration,
GetAccessor, SetAccessor and Constructor. The generated driver hands each listener
(node, index), with optional parent when requested. Numeric kind minting is not
planned and is not a blocker. The isolated supplied-fact modules still use numeric
checker AST observations internally; they are not ParseNode adapters.

These are partial listener declarations, not complete Discover registrations.
The old unrecognized entry field is removed. The registry additionally requires
named factories/classes, hooks, provenance, owned oracle adapters, mutants and
witnesses. Those integration pieces remain unwritten. No shared registry or
harness is changed. reportNode and reportRange are present in the referenced
RuleContext; its source does not expose the live checker facts these rules need.

Symbol-description implements all verdict logic on its supplied call facts,
including parentheses-normalized callee syntax and the FIRST declaration's
source-file flag, matching the production helper rather than its broader comment.
Hook-result checking implements the ASCII uppercase/digit hook-name predicate,
raw callee versus normalized React receiver, result-type Any flag and exact policy
wording. Its blinded-rule names are supplied raw registry metadata; the test oracle
--catalog mode reads the live registry's ResolvesReactValueTypes flags without
making a lint decision. The linked catalog currently names three rules. No list
is frozen into the native rule.

require-await implements its modifier/generator/body guards and numeric preorder
body walk, including nested function/class barriers, concise bodies, for-await,
and the correct await-using composite mask. Reporting candidates explicitly panic
with NotYet and exit 70. Contextual promise demand, declared generic signature
replay, heritage member types, function-head ranges and removeAsync suggestions
are NOT ported. The current owned interface models only the body facts, not a
finished function adapter. This is incomplete work, not a complete rule hidden
behind a harness exception.

The remaining work is adapting the handed ParseNode and checker facts to these
owned fact contracts and emitting results through reportNode/reportRange. The
handed-node contract is now specified on origin/lint-rules/harness, though it is
not on this landing base. Live checker adapters are not implemented and no new
bridge question is registered. Existing bridge modes supply many relevant raw
type/origin facts; absence of a lint-verdict question is not a blocker.
No rule verdict is delegated to Go by the native modules.

Validation (source the setup environment, redirect output to a log):

```sh
python3 stage1/cohere/typeaware/wave_10_leaf/validate.py \
  --stage0 /workspace/wave-10-f801-original/adamic \
  --artifacts /workspace/wave-10-leaf-final \
  > /tmp/wave-10-leaf-final-validation.log 2>&1
```

The reusable harness materializes TypeScript oracle inputs from controls.json
and builds the unchanged production Go registry rules as overlays. New Adamic
sources are all .a; generated .ts files are TypeScript oracle inputs, not Adamic
modules. Native controls receive hand-authored raw fact projections corresponding
to those inputs. They do not parse the inputs or query a live bridge. Therefore
agreement here proves isolated verdict behavior, not adapter correctness.

Symbol: 13 controls, 4 findings, 2036 identical diagnostic bytes. Hook: 16 controls,
7 findings, 5806 identical bytes, including all message/span/fix/suggestion fields
(the latter counts are zero). Require-await: 14 context-free controls, 798 bytes
of Boolean candidate results derived from production Go's actual findings. Its
ranges/messages/suggestions are not compared. Each module passes ASan/UBSan with
empty stderr. Three source mutants compile and exit 0 with empty stderr; only
comparison catches them. Candidate refusal is separately checked at exit 70.

An initial isolated timing observation over the original artifact paths was:
symbol core native 1.784ms versus full Go 33.759ms; hook core 1.915ms versus full
Go 29.239ms. These are NOT native-lint speed comparisons: native receives facts,
whereas Go loads/parses/checks source. No full native pipeline timing is claimed.
Setup/environment is unchanged from the last landing: total 100s, nproc 5.

No native compiler77/repository287 comparison for these three, live handle release
check, full upstream matrix, emitted-JavaScript comparison or full repository gate
is claimed. They remain in progress and do not authorize another reservation.
Evidence contains raw stdout/stderr, parser numbers, source lint and mutant logs.

## Current landing on c01907a7

Main advances again to c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06.
The only changed path under compiler/bridge/stage1/config scope is an added
internal/oracle/stage3_hook_test.go; production compiler, bridge, stage1,
submodule and configuration sources are unchanged. Rebase succeeds cleanly,
yielding source tip 39310513. Main is fetched again after the gates and remains
c01907a7. The six completed ports are green on this base; the new three remain
partial for the reasons above.

Rerun the exact earlier six-rule gates, with the same frozen manifests and
artifact settings and -count=1: original three PASS 92.098s; timeout PASS
50.990s; final process/blocking PASS 163.061s. All controls and compiler77/
repository287 findings/fixes/suggestions agree with Go, normal and sanitized.
All six clean-exit native mutants and both retaining-registry mutants are caught;
released queries exit 70. The final gate reuses only the process compiler/bridge
archives whose production sources are unchanged. No new process bootstrap or
full repository gate is claimed. Logs are in evidence/landing-c019.

The new isolated gate reruns with --artifacts /workspace/wave-10-leaf-c019:
symbol 2023 bytes, hook 5790 bytes, require-await filter 784 bytes; normal and
sanitized results agree with their respective Go authority. Their clean-exit
mutants are caught at bytes 59, 63 and 558. Different artifact-path lengths
account for changed serialized byte positions. Candidate reporting still refuses
at exit 70; full live-node integration and reporting are not validated.

New complete-process timing observations for the previously completed process
and blocking rules below use three alternating runs and compare each stdout.
They include load/parse/checker/serialization; worker load is uncontrolled.
They are separate from the incomparable isolated-core timing above.

| Rule and corpus | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| process-controls | 3.622435 | 0.079991 | 45.286x |
| process-compiler | 1.689670 | 0.292615 | 5.774x |
| process-repository | 0.279804 | 0.172372 | 1.623x |
| blocking-controls | 5.667838 | 0.077326 | 73.298x |
| blocking-compiler | 2.148552 | 0.307177 | 6.995x |
| blocking-repository | 0.366139 | 0.131993 | 2.774x |

## Canonical listener contract correction

The user confirmed canonical kind names and the ab70f38d4 handed-node contract.
The three listener declarations now use that contract. Earlier numeric-API
blocker descriptions are superseded; adapter implementation and require-await
reporting remain incomplete. Main remains c01907a7 after fetching origin, so
this metadata/documentation correction does not change the previously green
runtime sources or require another rebase. No additional rule is claimed.

Validation uses the exact registry.go from ab70f38d4 with a scratch probe:

```sh
go run /tmp/wave-10-kind-contract/registry.go /tmp/wave-10-kind-contract/probe.go \
  stage1/cohere/typeaware/wave_10_leaf/symbol_description/rule.json \
  stage1/cohere/typeaware/wave_10_leaf/react_hook_no_any_type/rule.json \
  stage1/cohere/typeaware/wave_10_leaf/require_await/rule.json \
  > /tmp/wave-10-kind-contract-validation.log 2>&1
```

Exit 0: all three listener subsets pass canonical ast.Kind validation. The actual
registry Render function produces their named kind buckets and visit(node, index)
calls using explicitly supplied scratch factory/class placeholders. Negative
controls reject numeric kinds, an unknown name, duplicate kinds and node: false.
This checks listener metadata and generated dispatch, not full Discover or module
compilation. Evidence preserves the probe and its output. The existing native
sources are unchanged; no new runtime, sanitizer or corpus run is claimed.
