# Apple declaration naming

`Name(Declaration)` is pure and needs no SDK. `Build([]Declaration)` validates
an entire proposed export surface and returns both directions of the mapping.
The generator must call `Build` and stop on an error before emitting bindings.
No numeric suffixes, order-dependent names, or silent collision winners exist.
A bijection cannot exist for every possible input under lossy naming rules;
it exists for each accepted, qualified export surface.

## Ordered rules

| Order | Implemented rule | Example |
| --- | --- | --- |
| 1 | Literal `SwiftName` (`NS_SWIFT_NAME`) wins, then effective `APIName`, then the synchronous Swift 3 importer transformations. Keep the original spelling and the computed Swift spelling. | `dismissViewControllerAnimated:completion:` becomes `dismiss(animated:completion:)`. |
| 2 | Remove the listed namespace prefix at an uppercase boundary. Export through `apple/<lowercase framework>/<kebab type>`. Members and nested types share their owning type's module. | `NSWindow` becomes `Window` in `apple/appkit/window`. |
| 2a | A type hiding an ECMAScript global retains its framework as a prefix, including its module and references. | `NSError` becomes `FoundationError` in `apple/foundation/foundation-error`. |
| 3 | Normalize acronym words; also split adjacent recognized acronyms inside an uppercase run. Types and literal cases use PascalCase; members use camelCase. Expand a terminal `ID` word to `Identifier`. | `HTTPURLResponse` becomes `HttpUrlResponse`; `taskID` becomes `taskIdentifier`. |
| 4 | Expand only the explicit full-word table: `Rect` to `Rectangle`, and the identifier suffix above. | `contentRect` becomes `contentRectangle`; `max`, `min`, and `index` stay. |
| 5 | Enum and option cases become PascalCase string literal names. A type supplied with `OptionCases` maps to a readonly array of their literal union. | `'Titled'`; `readonly ('Titled' \| 'Closable')[]`. |
| 6 | An unlabeled first method argument stays positional. Labeled arguments go into one trailing options object. Initializers and inferred factories become constructors with the same layout. C functions keep every argument positional. | `setFrame(rectangle, { display, animate })`; `dataTask({ with: url, completionHandler })`. |
| 7 | Map the specified scalar types to `number`, `BOOL` to `boolean`, and `NSString *` to `string`. Nullable and unspecified reference types retain `\| null`. Scalar values are inherently nonnull unless explicitly marked nullable. | `NSString *` becomes `string \| null`; nonnull becomes `string`. |

The entire prefix table is `NS` (Foundation/AppKit's NeXT namespace), `UI`
(UIKit), `CG` (CoreGraphics), `CF` (CoreFoundation), `CA` (QuartzCore's Core
Animation), and `AV` (AVFoundation). No other prefixes are guessed. Prefix
normalization also applies to global exported names and labels. The two-word
expansion table is intentionally small: geometry uses Rectangle and identity
uses Identifier. Neither maximum/minimum nor ordinal index is renamed.

## Facts the generator must supply

The ordinary fields contain the declaration kind, original name/selector,
original parent, framework, written types, parameter names, and nullability.
Additional facts are required where the Swift algorithms consult clang:

- `Enumerators`: all cases of the parent enum, including custom-name,
  deprecated, and unavailable flags. The common word prefix is bounded by
  the enum name; plural prefixes, leading `k`, and underscore separators
  receive Swift's special treatment.
- `PropertyNames`: all properties, including inherited ones. These protect
  meaningful type words in a method name from being omitted.
- `DefaultArgument`, collection `Element`, and block `Function` facts:
  Swift's default-argument and type information used by word omission.
- `Object`: clang established that the written star points to an Objective-C
  object. A namespace-looking spelling alone is insufficient: `NSRect *`
  must not be misidentified as an object. `Reference` identifies pointer-like
  typedefs whose written spelling has no star. `NSString *` is a builtin.
- `APIName` and `ParentSwiftName`: effective SDK API-note, bridge, or nesting
  names. These are separate from a literal `NS_SWIFT_NAME` attribute. The
  documentation fixtures use this field where they establish the imported
  spelling but do not establish the exact header annotation responsible.

A getter/setter pair is one `Property` description with `Getter` and `Setter`.
A BOOL property takes the getter's `is...` spelling. `AccessorProperty` lets
`Name` identify an accessor's canonical property; `Build` refuses raw accessor
entries and requires the canonical property description. It never folds two
independent originals into one reverse-map entry.

`Argument.Index` is the original Objective-C parameter index. `SwiftLabel`
keeps the imported label; `Name` is its Adamic options key or positional name.
Types and original parameter order are preserved in each layout partition.
Every parameter appears exactly once. Options fields have no inferred optional
or default semantics: a default-argument naming fact does not authorize the
binding to omit a parameter.

## Swift port and deliberate boundaries

The port is pinned to Swift's `swift-6.0-RELEASE`, commit
`f79ab15f11ffea6cad859da22b17da6cd9252167`, formerly in `apple/swift` and now in
[swiftlang/swift](https://github.com/swiftlang/swift/tree/f79ab15f11ffea6cad859da22b17da6cd9252167).
`words.go` ports `lib/Basic/StringExtras.cpp`'s forward word iterator;
`swift.go` ports synchronous `omitNeedlessWords`, its type matching, self/type
omission, property safeguards, preposition splitting, and initialisms. The
complete parts-of-speech table comes from `lib/Basic/PartsOfSpeech.def`.
Selector/initializer handling comes from `lib/ClangImporter/ImportName.cpp`;
common word/plural enum prefixes come from `ImportEnumInfo.cpp`. Source and
commit comments accompany the modified Go ports. The full Apache 2.0 license
and Runtime Library Exception are in the root third-party notices.

This is a bounded importer naming layer, not a claim to reproduce the whole
Swift compiler for every SDK declaration. These are the departures and limits:

1. Go slices replace LLVM strings and forward/reverse iterators. Grammar uses
   ASCII identifiers, as this Apple C/Objective-C naming surface does. It
   rejects an imported identifier with punctuation rather than escaping it
   into a plausible Adamic identifier.
2. The clang-dependent decisions are input facts, not AST inspection. Missing
   enum siblings, malformed selectors, invalid nullability, unresolved
   `instancetype`, unsupported pointer/declarator shapes, later unlabeled
   method arguments, and duplicate normalized options keys return errors.
   `instancetype` in a method result resolves to its concrete parent. `id`
   becomes `unknown | null` rather than pretending to be a concrete object.
3. SDK API notes, overlay names, CF bridge recognition/`Ref` removal, and
   automatic nesting cannot be recovered from a single name string. Supply
   their effective name as `APIName` or `ParentSwiftName`. A small explicit
   `foundationNames` table holds the documented Foundation reference renames
   used here, plus URL and URLRequest bridges. It does not implement the String,
   Array, Date, or other value-type overlays or their methods.
4. Async alternatives, completion-handler removal, NSError-to-throws rewriting,
   subscript import, variadics, availability-version selection, conflicting
   overrides, unsafe-method renames, notifications/newtype import, reserved
   initializer-label repair, and Swift's per-API historical compatibility
   exceptions are not inferred. This package describes the synchronous,
   complete-parameter declaration. Effective SDK names can be supplied when
   the base importer spelling differs. Inferred class factories use a matching
   parent name and same-class/instancetype result; per-API factory suppression
   likewise needs an effective name. No SDK-wide equivalence claim is made.
5. Swift's refined/private marker is kept in `Output.SwiftName` and
   `RefinedForSwift`. Adamic's normalization removes the marker from the exposed
   spelling. This package does not synthesize a handwritten refined overlay.
   Swift's special dummy label for a zero-argument private initializer is not
   synthesized; it keeps the original zero-argument layout.
6. Category descriptions receive their own normalized category marker in the
   parent's module. Swift has no named category symbol; its actual members
   must be described with the owning class as parent. The category behavior is
   covered by a synthetic test, not by a purported Apple Swift category doc.
7. Enum common-prefix matching stays at word boundaries even for a deprecated
   outlier's fallback. Swift's fallback in `ImportName.cpp` can strip a raw
   character prefix. Keeping a complete word avoids an accidentally truncated
   name; a resulting collision is still an error. An empty imported case is
   rejected instead of manufacturing a spelling.
8. Swift supports overloads, including same-base methods distinguished by
   labels and separate static/instance lookup. Adamic has one member namespace.
   `Build` rejects these collisions. It also rejects duplicate originals,
   type/module collisions, and aliases that would make reverse lookup ambiguous.
9. Adamic's post-import rules deliberately differ from Swift: prefixes disappear,
   acronyms become words, Rectangle/Identifier expand, enum cases are literal
   strings, option sets are arrays, and selectors become options objects.
   A later unlabeled method argument cannot name an options field and is
   refused. Scalar/type conversion here specifies the binding's type spelling,
   not ABI conversion or numeric-range proofs.

`JSONDecoder` is a Swift-native API, not an Objective-C header declaration.
Its `JsonDecoder` spelling is held by the word tests; it is not misrepresented
as a header-derived documentation fixture. The future Mac generator and ABI
marshaller remain outside this package.

## Documentation oracle and surprises

`testdata/declarations.json` contains 264 original descriptions of declarations:
88 AppKit, 69 Foundation, 81 UIKit, and 26 CoreGraphics. Every row independently
records a documented Swift spelling, an Adamic spelling, an expected module,
and its Apple documentation URL. No Apple header text or documentation prose
is copied into the repository. The documentation names were checked against
Apple's symbol pages on October 7, 2026. Many pages exposed only their symbol
title through the browser; this was a spelling audit, not a live SDK import or
an exhaustive check of header types, availability, or annotations. Tests compare
both name layers and module paths, then build and reverse the entire surface.

Surprises that changed the fixtures or supplied facts:

- `pushViewController` and `addOperation` keep their type words because the
  parents have `viewControllers` and `operations` properties. Dropping those
  facts incorrectly shortens the names to `push` and `add`.
- AppKit and UIKit both publish `setNeedsDisplay(_:)` for the original
  `setNeedsDisplayInRect:`. The fixtures supply the effective SDK rename;
  generic synchronous omission alone leaves `setNeedsDisplayIn`.
- `NSWindow.StyleMask.hudWindow` normalizes to `'HudWindow'`, while the original
  constant spells `HUD` in uppercase.
- CoreGraphics geometry symbol pages are partly hosted under CoreFoundation.
  The framework fact still controls the exported `apple/coregraphics/...` path.
- The CoreGraphics opaque-reference typedef originals end in `Ref`; their
  published Swift type names do not. The fixtures preserve both names.
- UIKit's zero `UIViewAutoresizingNone` is not a published `.none` Swift case.
  That candidate fixture was removed rather than inventing a Swift oracle.
- Two synthetic enum families with different prefixes, both importing into
  `NSWindow.Status`, strip to `'Ready'`. `Build` rejects the collision in
  either input order. It also rejects `+reload` and `-reload` on the same type.

## Validation

Source `/workspace/adamic-tools/env.sh` (or setup's printed path), then run:

```sh
go test ./internal/apple/naming > "$TMPDIR/apple-naming.log" 2>&1
gofmt -l cmd internal > "$TMPDIR/apple-format.log"
go vet ./... > "$TMPDIR/apple-vet.log" 2>&1
python3 internal/apple/naming/testdata/mutants.py > "$TMPDIR/apple-mutants.log" 2>&1
```

The mutant runner must run alone: it edits real implementation code, runs only
the named test into a separate log, requires a test failure rather than a build
failure, and restores the original source in `finally`. It records machine-readable
results in `/tmp/apple-mutants.json`. All 16 mutants exited 1 through assertions:

| Mutant | Check that caught it |
| --- | --- |
| Drop omitNeedlessWords | `TestDocumentationFixtures` |
| Strip one extra enum-prefix character | `TestDocumentationFixtures` |
| Leave URL uppercase | `TestNamesAndWordBoundaries` |
| Permit an export collision | `TestCollisionsAreRejected` |
| Retain the NS namespace | `TestDocumentationFixtures` |
| Leave Rect unexpanded | `TestNamesAndWordBoundaries` |
| Move the positional argument into options | `TestArgumentLayout` |
| Shift original marshalling indexes | `TestArgumentLayout` |
| Drop unspecified reference nullability | `TestTypeMapping` |
| Ignore explicit Swift naming metadata | `TestDocumentationFixtures` |
| Lose the refinement marker | `TestErrorsAndMetadata` |
| Ignore a BOOL getter name | `TestErrorsAndMetadata` |
| Fail to mark a constructor | `TestArgumentLayout` |
| Accept a struct pointer as an object | `TestTypeMapping` |
| Ignore property protection during omission | `TestPortSafeguards` |
| Lose the reverse mapping | `TestFixtureBijection` |

Setup reported Go 1.27.1, clang 20.1.8, Node 24.19.0; go/clang/node/submodules
were ready at 0 seconds, cache warming at 97 seconds, and setup finished in
97 seconds. `nproc` was 5 (cgroup CPU quota 4 cores).
The filtered uncached Node/native oracle smoke command was:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 5m -run '^TestNativeAgreesWithNode$/^internal$/^load$/^testdata$/^0.1$/^compile$/^(01_hello.ts|06_stack.ts|10_unicode.ts)$' > /tmp/apple-oracle.log 2>&1
```

It passed in 9.352 seconds. The final package run,
`go test -count=1 -race -cover ./internal/apple/naming`, passed in 1.118 seconds
with 89.9% statement coverage. Repository-wide formatting and vet logs were
empty, and `git diff --check` was clean. The full compiler oracle gate was not run: this
package has no compiler callers yet. Apple documentation is this package's
external naming oracle; Node's smoke gate checks that the existing compiler
still runs its selected programs. Neither demonstrates equivalence to a Mac's
SDK importer over declarations absent from the fixture table.

## ECMAScript global names

The global audit follows all 69 library references from the pinned
`lib.es2024.d.ts` and records 194 type and value names in
`testdata/es2024-globals.json`. Names are compared after acronym normalization.
The required Apple hits are NSError/Error, NSDate/Date, NSString/String,
NSNumber/Number, NSArray/Array, NSSet/Set and NSObject/Object. NSDictionary
also becomes FoundationDictionary by explicit policy; Dictionary is not an
ECMAScript global. NSProxy/Proxy also hits. The existing documentation fixture table contains
NSObject, NSProxy, NSNumber and NSError; the synthetic bijection test holds
all eight requested names plus NSProxy. Every listed
global is reserved, including globals not present in the documentation fixtures.
Type.Framework supplies the owning framework for references; legacy standalone
MapType calls infer known namespace ownership when this fact is absent.
Members retain their ordinary camelCase names. Primitive NSString conversion
remains string; its exported class is FoundationString.
