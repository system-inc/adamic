Built: corrected all six rule.json and .a listener declarations to production ast.Kind names, superseding numeric descriptors.
Commits: follows pushed 2753e4717 on wave-09, based on current main c01907a7; no new claims.
Commands/output: listeners/verify.py PASS against Go maps, native, sanitizers, source Node and emitted JavaScript.
Mutants: radix CallExpression becomes NewExpression in JSON and .a; both are caught by Go output comparison, and the native mutant builds/runs with empty stderr.
Not covered: shared visitor execution, checker-context migration, complete regex ports or a new full corpus sweep.

The user's latest instruction explicitly selects ast.Kind names, with no
numeric kinds. Both descriptor JSON and .a sidecars now use string arrays.
The independent Go oracle reads the unchanged production rule listener map,
uses kind.String with the Kind prefix removed, and sorts names. No separate
kind-number translation table is used. Exact declarations are:

| Rule | Kinds |
| --- | --- |
| @typescript-eslint/non-nullable-type-assertion-style | AsExpression, TypeAssertionExpression |
| no-invalid-regexp | CallExpression, NewExpression |
| no-label-var | LabeledStatement |
| no-misleading-character-class | RegularExpressionLiteral, SourceFile |
| prefer-const | VariableDeclarationList |
| radix | CallExpression |

The metadata does not read nodes, compare their kinds to select relevance or
refetch nodes. Relevance dispatch belongs to the shared driver. These are
listener declarations, not invented registrations or a claim that existing
legacy visitors already use the shared node:true callback. The registry and
handed-node API reviewed in HARNESS_AB70_REPORT.md now have matching kind
names. The numeric decoder mismatch is resolved and is no longer a blocker.

Reproduce with test output to a file:

```
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave09_core/listeners/verify.py > /workspace/wave-09-kind-names-test.log 2>&1
```

Both metadata sources agree with production Go. Normal native, sanitized
native, source Node and emitted JavaScript print identical names/kinds.
The JSON mutant changes only a declared listener name. The .a mutant builds
and exits 0 with empty stderr, so only the independent byte oracle catches
it. Complete streams, generated probes and measurements are retained in
validation-kind-names. This metadata-only change does not alter findings,
fixes, suggestions or checker ABI; the prior full corpus and released-handle
results in LANDING_C019_REPORT.md were not repeated. No full gate was run.
Setup is reused from the successful 88-second run, nproc 5. Metadata printing
times are retained, but do not measure full lint throughput.

Remaining scope is unchanged: dynamic new RegExp(pattern,'u') is refused by
native lowering, independently reproduced by the existing gap probe and the
shared regex branch. No hand-rolled matcher was added. The legacy checker-
backed visitors still need migration to shared RuleContext and its handed-node
API; this is unfinished work, not the now-resolved numeric schema blocker.
The named harness is not yet on current main. No React parking applies.
No additional batch is claimed while both regex rules remain incomplete.
Shared parser, harness, generator and compiler files were untouched.
