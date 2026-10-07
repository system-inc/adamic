# Wave 11 eighth batch

Native production-default ports of `react/no-danger-with-children`,
`react/no-multi-comp` and `react/no-namespace`. Each rule lives in its own
`.a` file and declares numeric SyntaxKind listeners in rule.json and its
exported kinds array. The owned driver indexes listeners by kind and passes
the relevant decoded node. Multi-component detection uses the production
SourceFile listener and its own whole-file, source-order collection.

The new `react-syntax-details` question is implemented in a separate checker
file and native decoder. It returns raw numeric AST nodes, roles, byte ranges,
modifiers, parser whitespace flags, symbol declaration syntax and Unicode
simple-uppercase text. Native code decides component capitalization, wrapper
recognition, return traversal, spread resolution, cycle guards, properties and
all findings. Unicode text uses the pinned Go simple mapping, which preserves
sharp-s and supplementary character behavior; it is a primitive, not a Go
component verdict. No Go lint rule is invoked through the bridge.

No selected production rule uses a Go regexp. Their string operations map to
JS string operations; no regex matcher or regex translation was introduced.

TestWave11EighthAgreementAndMutants builds an independent oracle from unchanged
Go production rules through a cohere build overlay. It extracts source inputs
from upstream-derived fixture tables without copying expectations, uses default
options on both sides, and compares full diagnostics/fixes/suggestions. Supply
ADAMIC_TYPESCRIPT_SOURCE pointing at pinned TypeScript v6.0.3; optionally preserve
artifacts through ADAMIC_WAVE_11_EIGHTH_ARTIFACTS. All test output goes to logs.

The batch also fixes an earlier owned question's crash on JSX text: Node.Text
has no JsxText accessor, and the older rules need no text payload for that kind.
The earlier suite now includes plain and space text children and passes again.

See REPORT.md and validation/ for findings, complete bytes, every mutant, quiet
performance rounds, commands and explicit limitations. Shared harness integration
and emitted-JavaScript lint-rule execution are not claimed. The new shared model
ab70f38d4 is on the harness branch, not the current main base. Four analysis
reservations remain parked and are not native ports.
