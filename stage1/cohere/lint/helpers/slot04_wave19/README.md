# Utility evaluation traversal helpers

Three Adamic files port Go cohere's utilityEvaluation.walk, resolveDeclaration and resolveValueFunctions. Each removes a prerequisite for all four rules listed in readiness.json. This directory changes no rules, harness, registration or compiler files.

walk takes stable CSS node identities, kind/children accessors and a declaration callback. It visits declarations depth first under rule, at-rule, context and at-root; other kinds are ignored. The tags are CSS AST kinds, not TypeScript lint listener kinds.

resolveDeclaration takes a stable identity, value presence and text, and callbacks for ParseValue, resolveValueFunctions, ValueToCss, rewriting and marking removal. Missing or empty values invoke no callback. Resolution failure marks the original identity and leaves its value intact. Success prints and rewrites the parsed value.

resolveValueFunctions takes value-node identities and accessors, one mutable EvaluationState shared across declarations, and resolution/replacement callbacks. The resolver receives false for a --value call and true for a --modifier call: callers pass the candidate value or modifierAsValue accordingly. A Resolution carries replacement text, ratio and success separately. Replacement must match replaceValueNode: turn the function into a word containing ValueToCss(resolved), clear children and do not revisit it. Other functions recurse. First failure returns true immediately, retaining side effects from earlier resolutions. Only a non-ratio value resolution adds the declaration to nonRatio; modifier ratio flags are ignored as in Go.

EvaluationState's two maps are allocated Adamic maps. Their presence booleans preserve Go's distinction between nil and allocated empty maps. Initialize presence false with empty maps for a fresh evaluation, or preserve both maps and presence from prior calls; existing false entries and declaration identities must survive. droppedPresent and dropped belong to the declaration caller, not resolveValueFunctions.

The private Go overlay calls the actual claimed methods. Go supplies parser trees and per-function resolveValueFunction results; these dependencies are not reimplemented. The driver compares a composed walk, separate declaration calls and separate value-function calls on source Node, emitted JavaScript and ASan/UBSan native. Controls exercise distinct declaration identities; input CSS trees have no shared pointers or cycles. Parser/replacement/printer arena projections cover the text and fields these helpers read and mutate, not unrelated CSS fields.

From the repository root, source the setup environment and run:

```
python3 stage1/cohere/lint/helpers/slot04_wave19/testdata/regenerate.py > /tmp/slot04-wave19-regenerate.log 2>&1
python3 stage1/cohere/lint/helpers/slot04_wave19/testdata/capture.py > /tmp/slot04-wave19-capture.log 2>&1
go test -count=1 -v -timeout=15m ./stage1/cohere/lint/helpers/slot04_wave19 > /tmp/slot04-wave19-tests.log 2>&1
```

Capture instruments Go through temporary overlays, including the real asserted consumer fixture harness, and installs pinned Tailwind 4.3.3 only in a temporary fixture directory. Consumer source texts are replayed as declaration-value parser controls; they do not establish rule finding parity. Captured live helper inputs are replayed separately with fresh state. Synthetic controls test initial flags and map state. No hand-written regex matcher or new Go-regex translation is involved.
