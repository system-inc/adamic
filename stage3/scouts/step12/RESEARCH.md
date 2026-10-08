# Scanner stop research

Observations below use origin/main 45487a809f89885a3fc651cd590e7dabf31362dc,
TypeScript 6.0.3 source 050880ce59e30b356b686bd3144efe24f875ebc8,
and cohere 7945d102a6c18dd36adf9114a758ce646e8b2359. Go paths below are
relative to cohere/TypeScript/tsc. They are inspected source, not a Go scanner
execution or a promise of byte equality with Go.

The unchanged gathered scanner closure has 89 code declarations in eight files,
77 export facade records, two wrapper spans for ONE namespace, and 78 evaluation
modules. Census entries map UTF-16 offsets back to pinned upstream spans.
Coordinates below are upstream, never behind-placeholder discovery coordinates.
Reproduction: census.cjs plus the shared slicer with --no-adapt. Complete site
rows and counts are in evidence/sites.json.gz.

The requested land-area-next report is absent in the fetched main. The available
combined-records-library report's fifteen observations are behind progressive
throwing replacements; it removes whole createScanner before later observations.
It cannot establish all remaining scanner-body sites. Its scratch compiler
also merges a feature branch not present in this scout's base. This scout does
not retire that report's blockers based on main-only controls.

## Locations, runtime behavior and alternatives

| Family | Reached sites in raw closure, upstream file:line | Node runtime and observed witness | Go source alternative | ECMAScript contract |
| --- | --- | --- | --- | --- |
| Namespace | One Debug wrapper, debug.ts:113; reached members at 117,196,213,223 | TypeScript transpiles a namespace to an object/IIFE with exported members. assertEqual control prints ok; differing arguments throw | internal/debug/debug.go:7,45 uses package-level Fail and Assert, no namespace object | Namespace syntax and its transform are TypeScript, not ECMAScript. Emitted functions, lexical bindings and object property assignments follow ordinary JS evaluation |
| Mutable namespace export | One declaration debug.ts:117; scanner.ts:1115 reads it | Export is an object property, not a copied constant. Setting isDebugging true is observed as true; false mutant differs | Go debug package has no corresponding isDebugging object property. Package functions replace this debugging namespace; not proof that Adamic may omit the field | Property Set/Get must observe mutation. No ECMAScript rule says TypeScript exported let is immutable |
| Error unchecked casts | Two expressions debug.ts:200:14 and 201:14, same target any | Casts erase. V8 captureStackTrace exists as a function; String-constructor mutant observes undefined | internal/debug/debug.go:7 uses panic; optional runtime.Breakpoint is commented out. No Error-constructor reflection | Error.captureStackTrace is V8-specific, outside ECMAScript. Feature presence, stack-crawl marker and stack text cannot be inferred from the Error spec |
| String unchecked casts | Two expressions on scanner.ts:4063:67 and :112, same target any | Casts erase; fromCodePoint(0x10000) yields two UTF-16 units, first 55296; 0xffff mutant yields one | internal/scanner/scanner.go:1729,1824 uses rune/string conversion; :2771 maps UTF-16 offsets. No String constructor object cast | String.fromCodePoint checks the code point range/integrality and produces UTF-16 units; String.fromCharCode converts arguments to uint16. They differ for supplementary points |
| Transient Set assertion | One expression scanner.ts:4097:24 | undefined! as Set<string> is actually undefined until scanner.ts:4101 assigns Script's Set. Witness prints true then true; Greek mutant changes the second line | internal/scanner/unicodeproperties.go:139-161 stores Script and Script_Extensions as scriptValues directly, avoiding this asserted undefined field | Set is an object value; assertions/non-null syntax are TypeScript erasure, not JS validation. Ordinary assignment establishes the later alias |
| Generated diagnostic assertions | 2,130 expressions in retained whole Diagnostics declaration, generated file:13-2142 | Original diag has an explicit DiagnosticMessage result. Reduced original-annotation witness prints 1002 on Node and native; code mutant differs | internal/diagnostics generated messages are typed Go declarations, consumed by scanner.errorAt | ECMAScript has no DiagnosticMessage or casts. Return object fields determine runtime content |
| Unproven enum slot | Exact requested stop and its count are NOT established without land-area-next | Candidate numeric enum witness prints 99 on Node AND native; 100 input mutant differs. This is not a faithful reproduction of the claimed refusal | internal/scanner/regexp.go:36 uses map[rune]regularExpressionFlags; scanner.go:1175 checks map lookup's ok result | JS numbers have no TypeScript enum membership check. Adamic must prove/check whatever nominal enum contract it adopts; Go enum-like integer types alone do not prove membership either |
| new Array(length) holes | Two constructions core.ts:2199:20,2200:19 in levenshteinWithMax | Four slots, zero own indexed properties, first read undefined. Initialization loop fills previous. Skipping the last initialization changes final read from 3 to undefined | internal/core/core.go:638-699 uses pooled []float64 buffers, grows/reslices them and initializes relevant cells. Also uses []rune, so it is not a drop-in UTF-16 algorithm | Array(number) sets length without making element properties; invalid uint32 lengths throw RangeError. Holes differ from stored undefined for enumeration and callback methods |
| Object.entries | Three calls: scanner.ts:222:31 (binding),4073:44 (as const literal); commandLineParser.ts:575:19 (literal) | Binding-shaped witness stops with the exact unproven-shape diagnostic. Hidden extra numeric/string mutant changes value type without changing the visible type | scanner.go:44 uses a typed map literal; :197 maps.Copy joins token and keyword maps. regexp/unicode data are typed Go tables | Object.entries -> EnumerableOwnProperties(key+value), own enumerable string keys and Get; array-index keys first, then string creation order. Symbols excluded |
| Computed field | One scanner.ts:150:5, ["" + "constructor"] | Own key is constructor. constructoX mutant changes it | scanner.go:58 has literal "constructor" map key | Evaluate computed expression and ToPropertyKey, then define the own property. A constant fold is valid only if it preserves key, effects and timing |
| Uint16Array | One constructor utilities.ts:10491:22, in parsePseudoBigInt; writes at 10502,10504,10516 | Dense zero-filled two-element view; 65536 assignment truncates to 0 while residual contributes 1 to the next cell. 65535 mutant prints 65535 0 | internal/jsnum/pseudobigint.go:60-88 uses math/big.Int.SetString and String instead of a manual 16-bit little-endian division loop | Typed arrays allocate zeroed elements, use ToIndex for length and uint16 conversion on store. They are dense; plain Array(length) is not equivalent |
| Generic returns | Sixteen generic signatures: core.ts:33,921,923,925,926,1828,2169; debug.ts:223; scanner.ts:121,125,131,3930,3952,3977,3981; utilities.ts:746 | Type parameters erase. forEachEntry returns first TRUTHY callback result, otherwise undefined. 1 -> 0 mutant changes result from 1 to undefined | scanner.go:293-298 exposes Mark/Rewind with concrete ScannerState; Go callers manage speculative rollback. core.go:581,701 has generics and zero-value result conventions | Calls and Return evaluate actual callback values. ToBoolean for rollback must include false, zero, empty string, null and undefined; truthiness is not merely result !== undefined |

The counts are syntactic reachability, not counts of observed lowering failures.
There are 2,136 assertion expressions: 2,130 generated, two Error, two String,
one Set, and one as const. Five runtime-unchecked expressions form three families
(Error, String, Set). That does not corroborate a claim of exactly three assertion
expressions in an unavailable adapted snapshot. The generated original assertions
are same-type assertions; no silent admission of an unchecked structural downcast
is inferred from the green diagnostic witness.

## Owner handoff, with provenance

BLOCKERS.md supplies historical candidate branches, not confirmed current owners
for the absent land-area-next list. Namespace belongs to namespaces-tsc; generic
callback/return shapes to nested-functions and generic-maybe-undefined; Array(length)
to library-array/library; Uint16Array and static String to library; computed names
and entry shapes to records-lowering. Error casts are explicitly assigned there
to adaptation 40 after typed Node declarations are admitted. Transient undefined
initialization is associated with non-null-check/readiness, and numeric enum
contracts with flag-enums. These are handoff leads, not verified closure claims.
Object.entries's unproven-shape language ruling stays with @system_adamic,
undecided. Exact current owner identities must come from the missing report;
this scout does not fabricate them or send messages to those owners.

## Hard cases

- Structural types are open. Hidden fields, non-enumerable fields, symbols, integer
  key order, getters, prototype differences and aliases all matter to reflection.
  The hidden-string mutant gives concrete evidence that declared fields alone
  cannot justify Map<string, number>. See ENTRIES.md for undecided options.
- A holes-to-dense rewrite needs a dominance/range proof for every read of both
  Levenshtein buffers, including row swaps and early exits. Zero filling is not
  equivalent merely because the normal algorithm intends to initialize first.
- The Script_Extensions assertion deliberately claims a Set while holding
  undefined at initialization. A proof must establish read-after-assignment through
  every alias and preserve initialization order, or insert a loud check/refuse.
- A native mutable namespace export needs shared storage through qualified reads,
  writes and closures. Snapshotting its initial false value would miscompile.
- Generic callback members require per-instantiation return representation and
  ownership, including undefined and all falsy results. Speculation restores six
  state fields; scanRange restores end and directives too. Our small forEachEntry
  witness does not validate scanner rollback or closure ownership.
- V8 stack capture is host behavior, not an ECMAScript guarantee. Removing it
  changes observable failure stacks. Neither Go panic nor a successful witness
  establishes scanner failure-stack compatibility.
- The enum candidate is green. The required failing enum shape must be located
  and reproduced before anyone claims its owner closed it.

Spec algorithms actually fetched are retained in evidence/spec.json. Failed
initial anchor extraction was corrected against current unquoted HTML attributes;
only found algorithms are retained. The cited String/V8 discussion is a semantic
reading, not a native conformance sweep. Historical evidence was fully inventoried
and read as bytes (781 files, 7,218,200 decoded bytes); the current reports and
relevant stop/source evidence were reviewed. A semantic review of every historical
log is not claimed. No missing land-area-next report is invented.
