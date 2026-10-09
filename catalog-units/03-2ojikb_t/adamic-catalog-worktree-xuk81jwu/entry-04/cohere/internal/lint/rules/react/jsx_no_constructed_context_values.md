# `react/jsx-no-constructed-context-values`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **0** (measured 2026-10-01, after the react wave fixed the 18 it had) |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallows JSX context provider values from taking values that will cause needless rerenders

## Why this recommendation

The 18 hits are genuine perf defects in shared primitives such as DialogRoot.tsx and CalendarContext.tsx, where a fresh context object each render re-renders every consumer in the tree.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Maybe** to **Yes**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

18 in the tree. Showing the first few.

**`libraries/structure/source/api/web-sockets/providers/WebSocketViaSharedWorkerProvider.tsx:479`**

```
const contextValue: WebSocketViaSharedWorkerContextInterface = {
```

> The 'contextValue' object (at line 479) passed as the value prop to the Context provider (at line 506) changes every render. To fix this consider wrapping it in a useMemo hook

**`libraries/structure/source/components/calendars/CalendarContext.tsx:136`**

```
const value: CalendarContextValueInterface = {
```

> The 'value' object (at line 136) passed as the value prop to the Context provider (at line 155) changes every render. To fix this consider wrapping it in a useMemo hook

**`libraries/structure/source/components/dialogs/DialogRoot.tsx:139`**

```
const contextValue = {
```

> The 'contextValue' object (at line 139) passed as the value prop to the Context provider (at line 219) changes every render. To fix this consider wrapping it in a useMemo hook

**`libraries/structure/source/components/drawers/DrawerRoot.tsx:195`**

```
const contextValue = {
```

> The 'contextValue' object (at line 195) passed as the value prop to the Context provider (at line 208) changes every render. To fix this consider wrapping it in a useMemo hook

**`libraries/structure/source/components/drawers/DrawerRoot.tsx:220`**

```
<DrawerNestedContext.Provider value={{ value: isNestedDrawer ? 2 : 1 }}>
```

> The object passed as the value prop to the Context provider (at line 220) changes every render. To fix this consider wrapping it in a useMemo hook

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.


## Deliberate divergence: a memoized value is only stable if every input is

Kirk's ruling on the 2026-10-01 rule review: "go for C right out of the gate". cohere reports
everything upstream reports, and also two findings upstream cannot make, because upstream stops the
moment it sees `useMemo`:

- **`unstableDependencyMsg`**, anchored on the dependency entry: the provider's value is a
  `useMemo`/`useCallback` (inline, or a render-scoped `const` initialized with one) whose dependency
  list names something proven to be a new identity on a render. The message names the dependency and
  says why.
- **`memoWithoutDependenciesMsg`**, anchored on the hook: the memo has no dependency list, so it
  recomputes every render.

The review found the shape at four sites (`DialogRoot.tsx:139`, `DrawerRoot.tsx:194`,
`TableRoot.tsx:526`, `Field.tsx:186`): an upstream-style fix that wraps the value in `useMemo` would
have changed nothing, because `mergeTheme(...)`, `buildColumnPinMap(...)` and
`buildFieldContextValue(...)` each return a new object on every render. Probed against those real
helpers through the program (a scratch file importing them, the four memoized forms): all five
inputs reported, and a memo over hooks and props stayed silent.

The analysis is one-sided, and that is how it keeps zero false positives. A dependency is unstable
only when the code proves it:

| reported | why |
|---|---|
| an object, array, function, class, `new`, `JSX` or regular expression built during render | a new identity each evaluation |
| a function declared in the render body | the same |
| a `const` in the render body initialized with any of these, including through `?:`, `&&`, `\|\|`, `??`, a comma or `=` | either branch is reachable |
| a call whose callee's body returns such a construction on some path, directly or through a `const` local or another such call | on that path every call is a new identity (subject to the escape rule below) |
| a call to an async function or a generator | a new promise or iterator every call |
| another memo whose own dependencies are unstable, and whose factory returns a new value (a `useCallback` always does) | transitive |

**The escape rule.** When a function (or a render initializer) can return a new value on one path
and an existing object on another, the new value proves nothing if it can have been stored on the
way out, because the next call may read it back through the other path. That is a caching helper
(`if(cached) return cached; const created = {...}; cache.set(key, created); return created;`). So in
that case every `const` (or assigned local) holding the new value must stay home in the function
that declared it: returned by that function, read or written through a field, compared, negated,
`typeof`-tested, used as an `if` test, or reassigned. Any other reference is an escape and silences
the finding: a call argument (`cache.set(key, created)`), a method receiver (`created.attachTo(...)`),
the right side of a store (`cacheObject[key] = created`), an alias, or a capture by a closure. An
assignment into anything but a local (`return cacheObject[key] ??= ...`, `(cacheObject[key] = {...})`)
is itself a store and proves nothing. In a render body, the memo's dependency list and the object a
memo factory returns are home too. A function whose every non-new path returns a primitive or
nothing needs no escape check: the new value cannot have been read back.

A memo factory and a helper declared inside the component see the component's declarations as
per-render values; nothing above the component does (a component declared inside another function,
or a helper a factory returned, keeps the outer value's identity).

Silent, because nothing proves instability: a memo whose factory returns an existing object however
often it recomputes, memoize wrappers (from source or from a library), parameters and props, imports, module-scope constants,
destructured hook results, `let` bindings (a later line may replace the initializer), property paths
(`a.b` of a fresh `a` need not be fresh), logical assignments (`shared ||= {}` keeps what the target
held), calls into anything without a body (declaration files, interface members, `declare function`),
a dependency list passed by name, and any dependency whose type is primitive (it compares by value
however it was computed, which also covers a body that lies through a cast).

**Overlap with `react-hooks/exhaustive-deps`.** That rule's `exhaustiveDepsConstruction` already
reports a dependency declared as a literal construction in the component, at the declaration. This
rule reports it again at the provider memo's dependency, which is the point where the context
identity is lost. The calls, transitive memos and async results are this rule's alone.

**Divergence from ESLint on ahra:** none today. Measured 2026-10-01: 18 findings before and after,
identical lines, because none of the four sites is memoized yet; the six memoized provider values in
ahra (`AppearanceProvider`, `SurveyForm`, `OpsNavigationBar`, `ProductProvider`, `AccountProvider`,
`WebViewServiceProvider`) were evaluated and every input is stable, which a hand read agrees with. So
`internal/differential/acknowledged.go` carries no entry for this rule. The day one of the four sites
is wrapped in `useMemo` without stabilizing its theme or helper result, cohere reports it and ESLint
goes silent; that line belongs there as a `SideCohere` entry naming this section.

Fixtures: `jsx_no_constructed_context_values_stability_test.go`, both directions, modeled on the four
real sites, plus the caching-helper and memoize shapes as must-stay-silent cases. Mutation sweep:
every verdict-bearing condition caught, the escape rule's every arm included; the only survivors are
the termination guards (the in-progress set and the depth bound), which bound the walk and cannot
change a verdict.

## Deliberate divergence: which declarations count, by function rather than by scope

Found 2026-10-01 by the react wave: `Providers.tsx` hoisted its icon context value to a module-scope
`const iconContextValue = {...}`, and cohere reported it (`withIdentifierMsg`, 38:37) while ESLint did
not. That was a cohere false positive: a module-scope object is evaluated once and keeps its identity.

Upstream resolves a name only in the scope the `JSX` sits in (`callScope.set`, no walk up). Measured on
the installed 7.37.5 through the Linter `API`:

| shape | upstream | cohere now |
|---|---|---|
| module-scope object, function or arrow as the value | silent | silent |
| component `const v = m` where `m` is module-scope | silent | silent |
| object declared in a function enclosing the component | silent | silent |
| object declared in the component | reports | reports |
| component-level object used by a provider inside an `if` block | **silent** | **reports** |

cohere follows an identifier only to a declaration whose nearest function is the identifier's own,
which is "evaluated each time this function runs". The last row is a true positive upstream misses,
because the `if` block is its own scope and the name lives one scope up; cohere compares functions, so
it reports. No ahra site has that shape today, so there is no `acknowledged.go` entry; when one
appears it is a `SideCohere` entry naming this section. Fixtures:
`TestJsxNoConstructedContextValuesOnlyCountsPerRenderDeclarations`, both directions; removing the
guard fails all five silent rows.

Measured after the react wave fixed the 18 sites: the session-start binary still reports
`Providers.tsx:38:37`, and the fixed binary reports 0 for this rule. The wave memoized the four
unstable sites properly (each theme merge in its own `useMemo` over `theme?.Dialog`-style property
paths, `columnPinMap`, `gridTemplate` and `fieldContextValue` memoized), and the stricter memo check
stays silent on all of them, which is the right answer.
