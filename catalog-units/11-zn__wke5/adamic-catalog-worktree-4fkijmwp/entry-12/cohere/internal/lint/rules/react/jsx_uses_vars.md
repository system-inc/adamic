# `react/jsx-uses-vars`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | none |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow variables used in JSX to be incorrectly marked as unused

## Why this recommendation

Stylistic, but the tree already satisfies it, so enabling costs nothing and prevents drift.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.


## Ported? No, and the reason is measured

**Declined 2026-09-06.** Not ported, and the recommendation above should be read as superseded
rather than outstanding.

### It has no reporting surface at all

`lib/rules/jsx-uses-vars.js` contains **zero** `context.report` calls. Its whole body calls
`markVariableAsUsed`, which exists to stop ESLint's `no-unused-vars` from flagging a binding that is
only referenced from a JSX tag name. Upstream disables its own message-id lint on the file for that
reason. There is nothing to report, so there is no fires-and-silent pair and
`TestEveryRuleShipsAFixturePair` could not be satisfied by a faithful port.

That argument alone is weaker than it looks, because it stops applying the moment somebody adds a
marking surface. The stronger reason is below.

### A faithful port would be redundant, because our checker already resolves the reference

Upstream needs this rule because `eslint-scope` does not connect a JSX element name to its
declaration. Ours does.

Measured with a probe at `internal/jsx_uses_vars_probe/`, which is **still in the tree and still
runnable**, driving this repository's own `no-unused-vars` over **upstream's own corpus for this
rule**:

| | |
|---|---|
| All 11 of upstream's `valid` cases, which are bindings referenced only from JSX | **clean** |
| 7 controls taken from upstream's `invalid` list, which are genuinely unread bindings | **all 7 report** |

The controls firing are what make the clean column a measurement rather than a broken instrument.
The probe was then broken on purpose -- one clean case edited so its binding really is unused -- and
it went red on exactly that case and green again when restored. Baseline-green, mutant-red,
restored-green.

So the outcome this rule exists to produce is already true here: no JSX-only component is falsely
flagged, and porting it would add a rule that reports nothing and changes nothing.

### What would reopen this

A change to how `no-unused-vars` resolves references, such that a binding used only as a JSX tag
name starts being reported.

The probe is kept rather than deleted precisely so that this does not depend on anyone re-reading
this document. It is an ordinary test: `go test ./internal/jsx_uses_vars_probe/` goes red the day
the decline stops being correct. A paragraph cannot do that.
