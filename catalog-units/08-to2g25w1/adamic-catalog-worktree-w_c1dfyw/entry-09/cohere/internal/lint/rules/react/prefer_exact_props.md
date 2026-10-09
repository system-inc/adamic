# `react/prefer-exact-props`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | none |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Prefer exact proptype definitions

## Why this recommendation

Stylistic, but the tree already satisfies it, so enabling costs nothing and prevents drift.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.


## Ported? No, and the reason is measured

**Declined 2026-09-06.** Not ported, and the recommendation above should be read as superseded
rather than outstanding. The audit's "violations: none" is the off-by-default zero rather than a
clean tree: this rule is structurally unable to fire here.

### Every reporting case needs one of two things this tree does not have

Upstream's corpus is 31 cases, 12 of them reporting. Partitioned mechanically rather than by eye:

| | |
|---|---|
| Reporting cases needing `settings.propWrapperFunctions` | **9** |
| Reporting cases needing Flow's `ObjectTypeAnnotation` | **3** |
| Overlap between the two | 0 |
| **Remainder that would be reachable here** | **0** |

`propWrapperFunctions` is a *shared ESLint setting*, not a rule option: `meta.schema` is `[]`, and
`exactWrappers` comes from `getExactPropWrapperFunctions(context)`, whose only source is
`context.settings`. `rule.Context` carries a source file, a program, a checker, a report function
and a file cache, and no path to settings. The `propTypes` message additionally requires that set to
be **non-empty**, so with no settings the arm cannot fire even in principle.

The other three cases need Flow type annotations. `ObjectTypeAnnotation` does not exist in the
vendored TypeScript AST at all; a grep of the shim finds zero occurrences against a control kind
that is present.

The control on the partition: 8 of the 19 *passing* cases also carry the setting, so the predicate
that selected the 9 is discriminating rather than matching everything.

### This was already established here, independently, before this measurement

`internal/lint/rules/react/forbid_prop_types.go` loses one arm to the same missing setting and
records the finding for this rule in its doc comment, including the numbers. That measurement and
this one were taken independently and agree exactly: 31 cases, 17 carrying a settings block, all 12
reporting cases going silent with no settings.

### What would reopen this

A settings surface on `rule.Context`. Four rules in this plugin read `propWrapperFunctions`, so it
is a shared substrate question rather than this rule's, and `forbid_prop_types.go` is the file that
already explains the gap.
