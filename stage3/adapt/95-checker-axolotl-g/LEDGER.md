## 4. Missing index declaration crashes getSymbolAtLocation

Status: JavaScript-changing fork fix in 95g, scoped to the previously throwing
parent fallback. Draft issue, not filed. No public API line is changed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Location: src/compiler/checker.ts:50153, getApplicableIndexSymbol reads
copy.declarations[0].parent after mapDefined may remove every declaration.

An augmented CommonJS export of a class with a broad computed static index
reaches one applicable index with no declaration and neither aliasSymbol nor
symbol. getSymbolAtLocation(v.xyz) throws reading parent with zero diagnostics,
at both noImplicitAny settings. The fix keeps the ordered aliasSymbol / symbol /
fallback selection. Only inside the third arm, a missing first declaration
returns undefined; every successful arm retains its original result.

This follows the behavior at the throwing arm in
[typescript-go 89d5d5b2849a0db0957065889ca58536fa6d2e4a, checker.go:32209](https://github.com/microsoft/typescript-go/blob/89d5d5b2849a0db0957065889ca58536fa6d2e4a/internal/checker/checker.go#L32209)
and [TypeScript main 6161772c1370bdd4a82f808e6823bd755eb1870d, checker.go:32895](https://github.com/microsoft/TypeScript/blob/6161772c1370bdd4a82f808e6823bd755eb1870d/tsc/internal/checker/checker.go#L32895).
Both Go implementations guard the empty filtered list more broadly. They also
return no symbol where 6.0.3 successfully returns a zero-declaration __index
symbol with a known type symbol. 95g deliberately preserves that successful
6.0.3 path, including Computed.xyz and the unchanged super.xyz implementation
query (undefined, zero diagnostics).

On an imported derived-class super.xyz witness, getImplementationAtPosition
already threw in the checker. It now throws reading valueDeclaration in
services/findAllReferences.ts:580. This is a moved existing crash, not a new
crash on a successful input; no service edit is included. The FourSlash
verifySymbolAtLocation assertion likewise now reports its missing-symbol error
instead of the checker exception. Actual built-compiler service probes preserve
all 29 successful rows and introduce zero new crashes; 27 formerly throwing
rows succeed and one moves. Full evidence is on typescript/axolotl-checker-g-evidence
under stage3/evidence/95-checker-axolotl-g/scoped-final/.

### Draft GitHub issue

Title: getSymbolAtLocation crashes on a declaration-free computed static index in TypeScript 6.0.3

Version: TypeScript 6.0.3 on Node 24.19.0. Current Go implementations already
return no symbol on the throwing input (references above).

Save exported.js:

```js
/** @type {string} */
let key = 'xyz';
class Computed { static [key] = 1; }
module.exports = /** @type {typeof Computed} */ (class {});
module.exports.extra = 2;
```

Save consumer.ts:

```ts
import v = require('./exported');
v.xyz;
```

Create a program containing consumer.ts using allowJs, checkJs, noEmit and
CommonJS module resolution. Obtain its checker and call getSymbolAtLocation on
the second statement's PropertyAccessExpression. Both noImplicitAny settings
produce zero diagnostics. Actual: TypeError: Cannot read properties of undefined
(reading 'parent') in getApplicableIndexSymbol. Expected: undefined when neither
type symbol supplies a parent and the fallback has no declaration, following
the Go behavior on this input. Preserve successful symbol-supplied parent paths.

Suggested fix: guard the first declaration only in the third parent-selection
arm. Follow-up: the service super-property implementation reader assumes a
symbol, so the derived-class variant moves the existing exception there rather
than completing. Audit that reader independently; it is not changed by 95g.

This entry and F's entry independently append to 95a's ledger end. Resolve the
append conflict when merging second; preserve both issue drafts and leave
renumbering to the wave cutter.
