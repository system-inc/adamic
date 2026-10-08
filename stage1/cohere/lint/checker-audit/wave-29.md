# Wave 29 checker and shared analysis audit

Label: wave-29. Claims: the twelve rules in
`stage1/cohere/typeaware/claims/wave-29.md` on `codex/typeaware-wave-29`.

Audited area: `c4bdc23fa86d55cf7e579989201c11258f4d3a62`.
Cohere pin: `7945d102a6c18dd36adf9114a758ce646e8b2359`.
Locations below are relative to `cohere/internal/lint/` at that pin.
All twelve rules stopped before a unified-harness implementation. No new upstream
case matches or mutant kills are claimed. This file adds no port or registration.

## Missing shared surface, grouped for integration

`TypeChecker.GetSymbolAtLocation` already runs inside the bridge. What is missing
is a shared question exposing its actual symbol presence, stable identity and
complete declaration records for identifier references. The existing
`symbol-origin` returns only a ValueDeclaration filename. It loses the distinction
between an unresolved name and a symbol without a value declaration, and loses
merged declarations. The existing `declarations` question accepts only class and
interface declaration nodes. Type-symbol origins do not substitute for binding
symbols. The following rows describe required results, not a claim that every Go
method itself is missing from TypeScript-Go.

| Exact Go call or helper | Required shared result/helper | Rules blocked | Call sites |
| --- | --- | --- | --- |
| `ctx.TypeChecker.GetSymbolAtLocation` | Symbol presence, identity, all `symbol.Declarations`, declaring file and `IsDeclarationFile`; declaration node spans/kinds/ancestors for initializer/import inspection | id-denylist; id-match; nexus/concurrency-no-check-then-write; no-restricted-globals; no-setter-return; no-shadow-restricted-names; react/jsx-fragments; react/jsx-no-undef; react/jsx-no-constructed-context-values | `rules/core/id_denylist.go:377`; id-match delegates at `rules/core/id_match.go:339`; `rules/nexus/concurrency_no_check_then_write.go:358`; `ecmascript/reference/value.go:132`; `rules/core/no_setter_return.go:73`; `rules/core/no_class_assign.go:171,189`; `rules/react/jsx_fragments.go:290`; `rules/react/jsx_no_undef.go:101`; `rules/react/jsx_no_constructed_context_values.go:320` |
| `checker.SkipAlias` | Followed alias identity and its declarations, preserving nil/unresolved distinctions | nexus/concurrency-no-check-then-write | `rules/nexus/concurrency_no_check_then_write.go:362` |
| `typeChecker.GetShorthandAssignmentValueSymbol`; `ctx.TypeChecker.GetShorthandAssignmentValueSymbol` | Actual read/write binding for shorthand property references, including declaration identity | no-restricted-globals; no-shadow-restricted-names | `ecmascript/reference/value.go:128`; `rules/core/no_class_assign.go:191` |
| `typeChecker.GetExportSpecifierLocalTargetSymbol` | Local value binding read by an export specifier | no-restricted-globals | `ecmascript/reference/value.go:130` |
| `reference.IsValueReference`; `reference.ReadSymbol` | Shared value-reference discrimination and the above binding accessors | no-restricted-globals | `rules/core/no_restricted_globals.go:259-260`; definitions `ecmascript/reference/value.go:35,124` |
| `reference.WritesToBinding` | Shared write-reference discrimination | no-shadow-restricted-names; nexus/concurrency-no-check-then-write | `rules/core/no_shadow_restricted_names.go:295`; `rules/nexus/concurrency_no_check_then_write.go:857`; definition `ecmascript/reference/write.go:100` |
| `declarationAnchoredAt`; `resolvesToDeclaration`; `rule.DeclarationsIn` | Shared binding/declaration anchor resolution, with same-file filtering and shorthand correction | no-shadow-restricted-names | `rules/core/no_shadow_restricted_names.go:273,296`; helpers `rules/core/no_class_assign.go:168,188`; `rule.DeclarationsIn` calls at `rules/core/no_class_assign.go:171,194` |
| `descriptor.IsFunctionUnder` | Shared descriptor-position analysis for setter functions, with the `isGlobal` callback | no-setter-return | `rules/core/no_setter_return.go:93,109`; definition `ecmascript/descriptor/descriptor.go:46` |
| `rule.IsDeclaredOnlyInDeclarationFiles` | All-declaration global proof, rather than a single value-declaration path | no-setter-return | `rules/core/no_setter_return.go:73` |
| `high_level_intermediate_representation.ForFunctionWithoutManualMemoization` | Shared source-to-HIR/SSA/capture lowering with manual memoization erased | react-hooks/set-state-in-effect | `rules/react/set_state_in_effect.go:267` |
| `high_level_intermediate_representation.ForFunction` | Shared source-to-HIR/SSA/capture lowering | react-hooks/set-state-in-render; react-hooks/static-components | `rules/react/set_state_in_render.go:183`; `rules/react/static_components.go:120` |

The JSX constructed-context rule additionally has memo/escape/capture work, but
this audit stops first at its unavailable binding declaration/initializer answer.
It does not invent a replacement shared helper name for that later work.

Independent non-checker blocker: id-match must construct
`new RegExp(pattern, 'u')` from runtime options. `internal/lower/regexp.go:39`
refuses a nonconstant pattern. No handwritten matcher or historical regex VM is
an acceptable fallback. Its shared global-reference predicate is also absent:
`idDenylistIsReferenceToGlobalVariable` at `rules/core/id_match.go:339`.

## Dependency reproducers

These are dependency probes, not completed parity witnesses. React probes require
the upstream test project's React declarations; missing React types can suppress
findings and do not prove parity. Options below are the rule's own options, without
severity. In a config, id-denylist's array entries are its option tuple elements.

### id-denylist

Options: `["foo", "undefined"]`.

```tsx
foo = 1; undefined = 1;
```

### id-match

Options: `["^[a-z]+$"]`.

```tsx
let no_camelcased = 1; Object.keys({});
```

### nexus/concurrency-no-check-then-write

Options: `{}`.

```tsx
import { existsSync, writeFileSync } from "node:fs"; let p = "a"; while (existsSync(p)) { p += "a"; } writeFileSync(p, "x");
```

### no-restricted-globals

Options: `["event"]`.

```tsx
const x = {event};
```

### no-setter-return

Options: `{}`.

```tsx
Object.defineProperty({}, "x", {set: v => 1});
```

### no-shadow-restricted-names

Options: `{}`.

```tsx
var undefined; ({undefined} = {undefined: 1});
```

### react/jsx-fragments

Options: `"syntax"`.

```tsx
import {Fragment as F} from "react"; const x = <F><div /></F>;
```

### react/jsx-no-undef

Options: `{"allowGlobals":false}`.

```tsx
declare global { var G: any } export {}; const x = <G />;
```

### react/jsx-no-constructed-context-values

Options: `{}`.

```tsx
import React from "react"; const C = React.createContext({}); function Component() { return <C.Provider value={{}} />; }
```

### react-hooks/set-state-in-effect

Options: `{}`.

```tsx
import {useEffect,useState} from "react"; function Component() { const [s,setS] = useState(0); useEffect(() => { setS(1); }, []); return s; }
```

### react-hooks/set-state-in-render

Options: `{}`.

```tsx
import {useState} from "react"; function Component() { const [s,setS] = useState(0); setS(1); return s; }
```

### react-hooks/static-components

Options: `{}`.

```tsx
function Component() { const Inner = () => <div />; return <Inner />; }
```

## Legacy bridge surface for the landing lane

Old branch: `codex/typeaware-wave-29`, local area merge
`af15e58ac7ab35307a7a1f15c3f7f7319a6d1813` (not pushed).
The branch includes inherited typeaware extensions as well as wave29's additions.
Do not treat every extension as authored or actively used by wave29.

Wave29 actively used these questions, absent from the audited area's Inspect switch:

| Question | Legacy Go file | Legacy Adamic consumer | Raw answers to migrate |
| --- | --- | --- | --- |
| `symbol-provenance`, `symbol-provenance\nalias` | `bridge/tsgo/checker/symbol_provenance.go` (`symbolProvenance`, `finishSymbolProvenance`) | `stage1/cohere/typeaware/symbol_provenance.a` | Symbol identity, optional `checker.SkipAlias`, all declaration spans/kinds/files/ambient-module names, declaration-file flags, value declaration and resolved import metadata |
| `reference-symbol-origins` | `bridge/tsgo/checker/reference_symbol_origins.go` (`referenceSymbolOrigins`) | `stage1/cohere/typeaware/wave-29-next/reference_symbol_origins.a` | Read binding identity and declarations, correcting shorthand and export-specifier references |

`regexp-program` in `bridge/tsgo/checker/regexp_program.go` is also absent from
area, but is an unused historical ABI question. The native handwritten regex VM
was removed; do not move it as an active id-match dependency.

Other inherited questions present on the old branch and absent from this area
baseline: `node-symbol-details`, `declaration-details`, `type-symbol-details`,
`property-declarations`, `symbol-identities`, `resolved-name`, `literal-value`,
`binding-declarations`, `alias-declarations`, `type-metadata`, `type-properties`,
`function-signatures`, `property-exists`, `identical-types`, `reference-shape`,
`container-bases`, `contextual-argument`, `annotated-return-shape`, `symbol-shape`,
and `annotation-shape`. This is an inventory diff, not a wave29 demand to move
all of them. The baseline already has `symbol-origin`, `type-origin`,
`declarations` and several type-shape questions.

Observed legacy dependency API compile failures after the area merge:

* `ast.SourceFile.Path()` is gone. `declaration_facts.go:22` uses it; area uses
  `SourceFile.PathKey()` for `Compiler.IsSourceFileDefaultLibrary`.
* `ast.SourceFile.FileName()` returns `tspath.RootedFilePath`, not `string`.
  `out.text` calls fail in `declaration_facts.go:17`, `facts.go:290`,
  `reference_symbol_origins.go:33`, `symbol_provenance.go:34,55`. Area's string
  consumers call `.AsString()`.
* `ResolvedModule.ResolvedFileName` also returns `tspath.RootedFilePath`, so
  `symbol_provenance.go:68` cannot pass it directly to `out.text`.

These are observed compile errors, separate from the missing shared questions.
No bridge file or dependency pin was changed to resolve them in this audit.

## Validation and limits

On the unchanged area-based tree, setup passed in 165.789s after clearing the
rebuildable Go cache; nproc 5. Registry generation and compile-only
`go test -count=1 -run '^$' ./stage1/cohere/lint` passed (0.024s).

An earlier all-input package attempt encountered note-only directories under
`rules/` and stopped at registry initialization: 0 test passes, 1 package failure,
0 skips, outer wall 10.001s. Those unregistered directories were removed from the
registry tree. This is not a passing whole-package gate and is not evidence of an
area runtime failure. No new port, upstream match, mutant kill or whole-package
pass is claimed. This documentation-only branch changes exactly this one file.
