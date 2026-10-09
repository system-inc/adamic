# `no-unused-vars`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **12** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow unused variables

## Deliberate divergence: a decorated class or parameter is used

A class declaration or a parameter that carries a decorator is never reported, whatever reads its
name. The decorator runs when the class is defined, with the class (or the parameter's index) as its
argument, so deleting the declaration changes what runs. By @system_cohere's ruling of 2026-10-03
(#ynneze5), from Base's tests, which declare a class only so its decorators register (11 sites) or a
parameter only so a parameter decorator records it (7 sites). typescript-eslint 8.67 has no decorator
handling here and reports both. The exemption is the decorator itself, never a naming convention, so
`argsIgnorePattern` stays unset. A decorator on a member does not make its class used. The fixture is
`TestNoUnusedVarsTreatsADecoratedClassOrParameterAsUsed`.

## Why this recommendation

The stated reason is wrong (it is a correctness rule, not formatting), but the verdict stands for a different reason: verify already ships no-unused-vars and all 12 sites are already suppressed with `eslint-disable-next-line @typescript-eslint/no-unused-vars`, so the base rule would double-report and its disable comments would not match.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Strong No** to **No**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

12 in the tree. Showing the first few.

**`libraries/structure/source/services/network/NetworkService.ts:196`**

```
infer TVariables // Infer but don't use
```

> 'TVariables' is defined but never used

**`modules/apple/contacts/ContactsApi.ts:170`**

```
return allContacts.map(({ _pk: _pk, _dbPath: _dbPath, ...contact }) => contact);
```

> '_pk' is defined but never used

**`modules/apple/contacts/ContactsApi.ts:170`**

```
return allContacts.map(({ _pk: _pk, _dbPath: _dbPath, ...contact }) => contact);
```

> '_dbPath' is defined but never used

**`modules/asana/AsanaApi.ts:941`**

```
static async myTasks(_workspaceGid: string = defaultWorkspace): Promise<void> {
```

> '_workspaceGid' is assigned a value but never used

**`modules/facets/FacetsMonthlies.ts:127`**

```
_options: Record<string, string | number | boolean> = {},
```

> '_options' is assigned a value but never used

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

