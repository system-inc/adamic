Added isolated numeric supplied-fact verdict modules for symbol-description and structure/react-hook-no-any-type, plus require-await body filters.
Reservation was pushed in f7df1008 before implementation; previous complete six-rule landing is 6d4e8cfc on main f8013f0b.
Owned validate.py passes production Go controls, native sanitizers, clean-exit mutants and explicit missing-adapter refusal; production source lint finds 0.
Symbol/hook mutants fail byte comparison at 60/64; require-await filter mutant fails its narrower Boolean comparison at 568.
These are partial ports, not complete native rules: live supplied-node adapters, corpus comparison and require-await reporting remain unimplemented.

The three earlier React claims are PARKED under the user's explicit instruction.
The first six rules remain completed; their full landing observations are in
../wave_10_react/README.md. No additional rule is claimed beyond these three.

Each directory has rule.json with numeric kinds from the pinned typescript-go
parser's ast.Kind, measured by testdata/kinds.go. CallExpression is 214;
require-await listens to 263, 219, 220, 175, 178, 179 and 177. These declarations
are not a claim that the shared native parser now exposes numeric kinds.
The shared ParseNode still carries a string kind; its native kind mapping and
supplied-node ABI are absent. No shared harness or registration generator is
changed. The visit functions use supplied facts and numeric kind checks, with
no node refetch or file walk.

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

The concrete shared boundary is construction and dispatch of numeric supplied
nodes. Current main has neither a numeric ParseNode API nor a rule.json driver.
The owned interfaces here are isolated fact contracts, not an invented declaration
of the coming shared ABI. The live checker adapters are not implemented, and no
new bridge question is registered. Existing bridge modes do supply many relevant
raw type/origin facts; absence of a lint-verdict question is not the blocker.
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
