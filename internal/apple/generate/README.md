# Objective-C binding generation

`Generate(io.Reader, Configuration)` consumes clang's JSON translation unit and
returns a sorted `Output` containing module files and `bindings-check.m`.
`RunClang` supplies a stdout pipe and refuses failed clang invocations.
`Output.Write` writes the validated result. The loader can call the library
without spawning the command or changing its own embedded-binding mechanism.

On a Mac, using the current SDK:

```sh
go run ./cmd/adamic-apple-bindings AppKit Foundation CoreGraphics -o internal/load/apple
```

The command accepts framework operands before or after flags. `-sdk` selects
macosx, iphoneos/iphonesimulator, appletvos/appletvsimulator,
watchos/watchsimulator, or xros/xrsimulator. `-platform` can explicitly set the
availability platform. The command obtains clang, the sysroot and SDK version
from xcrun, then uses an arm64 target at that SDK version. It generates an
umbrella importing the selected frameworks. Framework dependencies needed as
object types must also be selected; unbound types are omitted with a reason.
Only public headers under the configured framework header directories export
declarations. The caller supplies header roots when using the library.

For the original Linux fixture SDK:

```sh
go run ./cmd/adamic-apple-bindings AppKit Foundation -o /tmp/apple-fixture-bindings \
  -headers internal/apple/generate/testdata/SDK \
  -umbrella internal/apple/generate/testdata/umbrella.h
clang -x objective-c -fobjc-runtime=macosx-10.13 -fblocks -fsyntax-only \
  -Werror -Wno-nullability-completeness \
  -F internal/apple/generate/testdata/SDK \
  -I internal/apple/generate/testdata \
  /tmp/apple-fixture-bindings/bindings-check.m
```

Linux clang 20.1.8 crashes in its legacy Objective-C runtime's JSON method-name
dumper. Selecting the macOS Objective-C runtime makes syntax-only fixture
inspection work; it neither links an Apple framework nor emulates one.

## Patterns and rules

| Pattern | Generated behavior |
| --- | --- |
| Translation-unit `inner` array | Decode and process one declaration at a time. Never decode the whole JSON tree or use ReadAll on the AST. |
| File deltas | Consume locations and ranges in document order, including discarded children and macro spelling/expansion locations. `includedFrom` does not change the location's own file. A declaration keeps the file resolved at its own `loc`. |
| Class definitions | Naming chooses the class and module. Emit `@objc class`, original-name documentation and callable members. Forward declarations are not exports. A class with no bound constructor has a private constructor, or a protected one when bound subclasses need to extend it. |
| Class inheritance | Import and extend the bound superclass. Inherited property names are supplied to naming's word-omission protection. |
| Categories | Merge members into their owning class; do not manufacture a separate runtime class. |
| Protocols | Export an interface tagged `@objc protocol <Name>`, whose methods a program's class implements as a delegate (docs/apple.md). Each method takes its arguments in the selector's order, positional, and carries `@objc implement <selector> <arguments> -> <result>` when its types can cross back into Adamic; a required one also keeps its `method` tag, for an object of Apple's that conforms. Optional markers are recovered from source: an optional method is an optional member (`windowShouldClose?(sender: Window): boolean`), left out with its reason when it can't be implemented, never promised present. Swift tells a protocol's methods apart by their labels; a class has one method per name, so where two share a name each folds its labels in (`tableViewObjectValueForRow`), and a folded name that lands on another's becomes its selector's words (`applicationDidUpdateUserActivity`). The check file holds every implement tag to its header with a class conforming to the protocol that implements the method in the bridge's C types, so a conflicting type fails the compile. Class messages and constructors need a concrete class and are omitted. |
| Explicit Swift names | Read the attribute's source expansion range and balance parentheses to recover `NS_SWIFT_NAME`, including initializer labels. Direct `swift_name("...")` attributes are also read. |
| Refined declarations | Feed `SwiftPrivateAttr` to naming; keep the original selector in the tag. No handwritten Swift overlay is synthesized. |
| Availability | Read `API_UNAVAILABLE` or direct `availability(...,unavailable)` attributes. Omit declarations unavailable on the selected platform. Availability for another platform does not suppress a declaration. Unknown attribute spellings fail explicitly. |
| Deprecation | Omit declarations deprecated on the selected platform, Apple's "don't use", each listed with `deprecated on macos` in its module's comments, enumeration cases included (deprecated.go). The JSON tree has the attribute but not its arguments, so they're read from the macro as the header writes it: `API_DEPRECATED`, `NS_DEPRECATED_MAC`, `availability(..., deprecated=)`, a region's macro, or a framework's own, expanded. `API_TO_BE_DEPRECATED` and another platform's deprecation keep a declaration; a deprecation macro that can't be read omits it. |
| Leaves | Derive the cycle finder's leaf table (leaves.go), written beside the bindings as `leaves.txt`: classes whose instances, and every subclass's, hold strong references only to other leaves, each listed with the references it may hold. docs/apple.md, "Delegates", says what counts. |
| Nullability | `_Nonnull` references are present; `_Nullable`, `_Null_unspecified` and unannotated references include `undefined` and use nullable native tags. Clang supplies assume-nonnull-region annotations. |
| Properties | Preserve explicit custom getter/setter selectors and readonly/class facts. Do not emit implicit accessors as duplicate methods. Explicit getter/setter method pairs become one property. BOOL properties preserve an `is...` getter. |
| Instance/class methods | Keep Apple's selector in `method`/`static` tags. Naming supplies the base and argument labels. An unlabeled first argument is positional; labeled arguments use a single trailing object. |
| Initializers/factories | Naming identifies constructors; instance initializers use `init`, class factories use `static`. All declared parameters remain required; nullability does not invent default arguments. |
| Ownership | Object results in alloc/new/copy/mutableCopy families, or with `NSReturnsRetainedAttr`, use `new object`; explicit `NSReturnsNotRetainedAttr` overrides the family inference. Retained C object/string results use the same ownership tag. |
| Enum constants | Recover evaluated nested ConstantExpr values, including shifted expressions; implicit values increment the previous value. Naming strips the common word prefix and rejects collisions. Enumeration values must fit a signed 64-bit long, over `NSInteger` or `NSUInteger` (both 64 bits in the bridge's long, the check file accepting either); option bits, over `NSUInteger` or `unsigned long long`, are unsigned and may reach bit 63 (`NSAlignRectFlipped`), written unsigned in the tag and with a `UL` suffix in the witness, and carried in the bridge's long with their bits intact. |
| `FlagEnumAttr` | Emit PascalCase literal domains, readonly-array parameters and `options(Name=bit,...)` tags. Ordinary enums use unions and `enum(Name=value,...)`. Fixed underlying types must match the bridge's long/unsigned long ABI when passed. |
| Scalars | Double/CGFloat, long/NSInteger, unsigned long/NSUInteger and BOOL use their documented tags and Adamic number/boolean types. NSString pointers become strings. Narrow integers and float are omitted because the bridge has no matching native call type. |
| CGRect/NSRect | Resolve NSRect's desugared CGRect identity. Emit the four-number Rectangle interface and rectangle argument tag. No rectangle result is emitted. |
| C functions | Emit standalone tagged functions; preserve positional C parameters and their order. |
| Module output | One ambient declaration file per module, sorted paths, imports, declarations, comments and witness calls. Original spellings are comments and tags; source offsets, addresses and absolute SDK paths are absent. |
| Name collisions | Validate the emitted surface with naming.Build before returning output, which reports every collision at once. A member's slot is its owner and name, static and instance apart; callables may share a slot as overloads when their argument shapes differ (lowering takes the overload the checker resolved, each with its own tag), and a value member may not share one. Imports that would hide a local export or another imported type also fail. A collision aborts generation. |
| Class and protocol of one name | Swift's rule: the protocol imports as `...Protocol` (`NSAccessibilityElement` and `AccessibilityElementProtocol`), whichever is declared first; `id<Name>` means the protocol and `Name *` the class, and the protocol's members carry its imported name in their identity. |
| Factory beside an initializer | Swift's rule: a class factory imported as an initializer gives way to an instance initializer of the same shape (`+[NSAffineTransform transform]` beside `-init`). |
| Method beside a property | Swift tells `abbreviation(for:)` from the property `abbreviation` by labels; a class can't, so the method folds its first label into its name and takes that argument positionally, Objective-C's own reading: `abbreviationFor(date)`, `isValidDateIn(calendar)`. Only properties of the same staticness count. |
| Inherited members | A class is written after its superclass and carries its whole member surface, its own and inherited, by Adamic name. A method that would take an inherited name without overriding that selector folds its first label in (`cellAtRow(row, { column })` beside `NSControl`'s `cell`). A method name shared with an ancestor's is one overload set: the class re-declares the inherited overloads it doesn't override, which send the ancestor's selector (`NSStackView`'s `remove(view)` beside `NSView`'s four). A property or class method meeting an ancestor's declaration of that name that disagrees (a method, another type, a property) is dropped, and the ancestor's stands, sending the same getter: `NSMatrix`'s `selectedCell`, `NSSavePanel`'s nullable `title`, `+[NSCalendarDate distantFuture]`. |
| Informal protocols | Categories on `NSObject` itself are delegate methods declared on the root for the compiler (`NSURLClient`), not methods `NSObject` has, and are left out. |
| Swift's conflict rule | Methods whose shortened names would collide keep their omitted words, each one whose name omission changed: `addObject(object)` beside the action `add(sender)`. |
| Any object | `id` binds as `NSObject`, `FoundationObject` in Adamic, which `objc/NSObject.h` declares and Foundation's module holds. |
| Moved headers | `CGRect`, `CGPoint` and `CGSize` are defined in CoreFoundation's `CFCGTypes.h` on current SDKs and generated with CoreGraphics, where every caller finds them. |
| Framework availability macros | A framework's own function-like macro over `availability(...)` (CoreGraphics' `SCREEN_CAPTURE_OBSOLETE(10.5,14.0,15.0)`) is expanded from its `#define`; `unavailable` or `obsoleted` on this platform leaves the declaration out. Macro names read with their digits. |
| Blocks | A block parameter that returns nothing and takes what the bridge carries becomes a closure: `completionHandler:(void (NS_SWIFT_SENDABLE ^)(NSData *data, NSURLResponse *response, NSError *error))` is `block(object?,object?,object?)` and `(data: Data \| undefined, response: UrlResponse \| undefined, error: FoundationError \| undefined) => void`, its parameter names read from the header where it wrote them (after any attribute before the caret), `argument1` onward where it didn't. A block that returns a value, or takes a rectangle, enumeration, options or block, is left out. The check file declares the block in the bridge's C types, which clang refuses where the header's block differs. |\n| Targets and actions | An `id` target followed by a `SEL` action is one closure, the `action` tag, which the bridge hands Apple as both: `buttonWithTitle:target:action:` is `constructor({ title, action })`. |\n| Setters for uncrossable getters | A readwrite property whose getter's result can't cross but whose setter can (`NSView`'s `frame`, a rectangle) is offered as its setter, `setFrame(frame)`, Objective-C's own spelling, through the ordinary method rules. |\n| 64-bit integers | `long long` and typedefs over it (`int64_t`, `NSProgress`'s counts) cross as the bridge's 64-bit integers, and the check file accepts either spelling of 64 bits. |\n| Keeping words across inheritance | A method under an inherited name with another selector and no label to fold keeps the words omission dropped, when that frees the name: `NSControl`'s `drawCell(cell)` beside `NSView`'s `draw(dirtyRect)`, `NSStackView`'s `removeView(view)`. Re-declaring the inherited overloads is the last resort, since Adamic refuses to relate overloads whose parameters differ. |\n| Overload order | A name's overloads are written most fields first: tsc types a closure's parameters from the first overload it tries, so `dataTask({ with, completionHandler })` precedes `dataTask({ with })`. |\n| Runtime-only messages | `+initialize` and `+load` are the runtime's to send, and are left out. |
| Redeclared properties | A category that redeclares a class's property (`NSSlider`'s `vertical`, readonly in a category beside the class's readwrite) is dropped, unless it widens readonly to readwrite. |
| Unnameable members | A member the naming layer can't name yet (a zero-argument named initializer without an explicit Swift name, such as `initListDescriptor`) is left out with its reason rather than stopping the framework. |

The naming amendment reserves every audited ES2024 global, plus Dictionary by
explicit policy. NSError, NSDate, NSString, NSNumber, NSArray, NSSet,
NSDictionary and NSObject retain Foundation in their exported type names;
NSProxy does too. Primitive NSString parameters still use string.
See [the naming rules](../naming/README.md) and its complete global audit.
The pure naming package retains its previously accepted null-based type
spelling; the binding generator uses the bridge's undefined-based spelling.

## Header witness

Every emitted method, property accessor, constructor and C function has a typed
call in `bindings-check.m`. The witness derives native C types from the emitted
tags: double, long, unsigned long, BOOL, id and adamic_apple_rectangle. Calls use
Objective-C receivers whose declared types select the actual header declaration;
they do not cast objc_msgSend and thereby bypass header checking.

Scalar parameter/result compatibility, enum values and the 64-bit ABI are held
by static assertions. Rectangle assertions compare size, alignment, coordinate
offsets and coordinate types before copying the native rectangle's bits into
CGRect/NSRect for the typed call. The copying bridges C's nominally different
struct types; the check does not cast an incompatible struct into a call.
A deliberately wrong enumeration parameter type fails the header ABI assertion.
The check is compiled without running it. On the Mac, compile it against the
same SDK/sysroot and target used to generate it, with `-fsyntax-only -Werror`.
Deprecated declarations remain represented; a SDK may require choosing warning
settings for deprecations when compiling the full check file.

## Explicit omissions and limits

One-line comments identify omitted declarations. Blocks beyond the shapes above,
C arrays, variadics, unions, function pointers, other structs,
unknown/object-pointer shapes, generic collections, unbound object types,
opaque CF references and unsupported scalar widths are omitted. Manual retain/release/autorelease/dealloc messages, consumed parameters and
nonconstructor consumed receivers would violate the bridge's borrowed-reference
contract and are omitted. Global constants
have no native-load tag, so they are omitted too. Noncanonical typedef aliases
have no independent bridge representation. Anonymous enum constants likewise
have no global native-load tag. Rectangle, enumeration and option results are
omitted because lowering cannot yet carry them.

This is a synchronous generator over JSON/header facts, not Swift's complete
SDK importer. It does not read SDK API-note YAML, bridge overlays, async/throws
transformations, deployment-version availability, category implementations,
protocol inheritance/conformance synthesis, or inherited methods from an
unbound superclass. Two overloads of one shape, or a value member and a method
that can't fold a label, remain a fatal naming error. A method colliding with an
inherited property isn't detected here; the checker sees it. Overloads whose
argument types are a class and its subclass resolve to the one declared first.

The existing bridge cannot lower a literal undefined as a nullable string
argument: it assigns that literal the object representation. The compiler
fixture instead passes a nullable string returned by a generated property.
Its string | undefined declaration remains independently held by the golden
and pattern tests. No compiler change is made for this limitation.

The Linux run cannot generate the real SDK's bindings, compile the check against
Apple headers, link the generated C to Apple frameworks, or run Apple's
framework methods. The Mac generates them, compiles the check file, and only
then serves them (below).

## In the loader

The loader calls `FromSDK` for Foundation, AppKit and CoreGraphics the first time
a program imports one of their modules, writes the output into
`~/Library/Caches/adamic/apple/macosx-<SDK build>-<Version>/` (a partial
directory renamed into place only once complete), and refuses to serve it unless
`CheckSDK` compiles its check file against the SDK. `Version` hashes this
package's sources and the naming layer's, so any change to a rule makes a new
cache. `ADAMIC_APPLE_BINDINGS` names a directory to read instead, which is how
the lowering test compiles against the fixture's generated modules on Linux.
Hand-written binding files stay embedded in the loader: whole modules where
nothing is generated, and additions merged into a generated module.

## The real SDK (macOS 27.0 SDK, October 6, 2026)

`go run ./cmd/adamic-apple-bindings Foundation AppKit CoreGraphics -o <directory>`
generates all three whole: 1,383 modules, 9.1 MB, about 6 seconds. Its check
file holds 8,664 calls and compiles clean against the real headers with
`xcrun clang -x objective-c -fsyntax-only -fno-objc-arc -Werror
-Wno-deprecated-declarations -fmodules`, the bridge's own mode. Under ARC, the
checks of `NSAutoreleasePool` and `NSGarbageCollector` are refused, as they
should be; without `-Wno-deprecated-declarations`, deprecated declarations warn.

Every module type-checks under Adamic's own compiler with no errors: a program
importing each of the 1,383, compiled with `ADAMIC_APPLE_BINDINGS` naming the
output and `go run ./cmd/adamic c`. The first such run, through an overlay that
kept the hand-written seeds beside them, held 197 errors: duplicates of the
seeds, protocol overloads written as properties, and, most of them, subclass
members meeting inherited ones. internal/apple's witnesses (a window with a
button's action, two URLSession fetches through blocks, a SwiftUI counter)
pass on these bindings, and `examples/apple/window.a` opens its window.

Each rule above marked as Swift's, and the unsigned option bits, the protocol
results, and the redeclared properties, is a wall the real headers showed, held
on Linux by `Collision.h`, `Shapes.h`, `Inheritance.h`, `Closures.h` and the
golden fixture. The largest
omissions, by count: global constants (2,835); unbound types (1,748), mostly
`NSRange`, `NSPoint`, `NSSize`, `CGContextRef`, `SEL` and `Class`;
enumeration and option results (413); pointer-to-pointer parameters (331); and
rectangle results (82).

Output.Write does not delete previously generated files; the loader writes each
version into a directory of its own.

## Evidence

The original fixture headers define a fictional root and APIs. Vendor header
text was not copied. They produce 18 declaration files plus the header witness,
held byte for byte by testdata/golden. They exercise both ordinary and macro
attribute ranges, assume-nonnull and unannotated pointers, enum expressions,
custom properties, getter/setter pairs, categories, protocols and an excluded
framework followed by another selected declaration.

`TestCompilerLowersGeneratedModules` writes the fixture's generated modules into
a temporary directory and runs `go run ./cmd/adamic c
internal/apple/generate/testdata/bindings.a` from the repository root with
`ADAMIC_APPLE_BINDINGS` naming it. The program imports the generated modules,
constructs objects and calls rectangle, enum, options, nullable-object,
nullable-string, property, protocol, overloaded and C-function bindings, and
hands an object of its own class to Apple as a protocol it implements. It
checks the emitted C for the selectors and conversion routines. This proves
checking/lowering and C generation, not linking on Linux.

The standalone mutant runner edits implementation files one at a time, logs the
selected test, requires an assertion failure rather than a Go build failure and
restores the file in finally. Run it alone:

```sh
python3 internal/apple/generate/testdata/mutants.py > /tmp/apple-generator-mutants.log 2>&1
```

| Mutant | Holding test |
| --- | --- |
| Drop nullability from takeText's parameter | TestPatterns |
| Emit an option set with an enum tag | TestPatterns |
| Keep a platform-unavailable method | TestPatterns |
| Resolve a declaration's file after its children | TestDocumentOrderFileDelta |
| Leave module paths in map iteration order | TestCanonicalModuleOrder |
| Drop the global type's framework prefix | naming.TestGlobalNameBijection |
| Read the entire AST before visiting declarations | TestStreamVisitsBeforeEOF |
| Promise an optional protocol method is present | TestPatterns |
| Implement a protocol method in a type its header contradicts | TestHeaderWitness |
| Leave two protocol methods one name | TestPatterns |
| Permit an import to hide a local export | TestImportNameCollision |
| Feed a raw struct declarator to the naming layer | TestWrittenTagTypeSpellings |
| Borrow a consumed parameter | TestPatterns |
| Bind manual release | TestPatterns |
| Borrow a consumed receiver | TestPatterns |
| Lose retained C string ownership | TestCompilerLowersGeneratedModules |
| Hide a class/protocol declaration kind | TestDeclarationKindCollision |
| Pass a double where an enum tag promises long | TestHeaderWitness's independent clang assertion |

Naming surprises visible in the fixture: takeRecord becomes take, copyRecord
becomes copy, currentPanel becomes current, while takeObject retains Object.
A Swift-labeled init(count:) uses an options object even with one parameter;
the explicit init(_:) leaves its first argument positional.

Final Linux validation on October 7, 2026:

```sh
go test -count=1 -race -cover ./internal/apple/... ./cmd/adamic-apple-bindings > /tmp/apple-generator-race.log 2>&1
gofmt -l cmd internal > /tmp/apple-generator-format.log
go vet ./... > /tmp/apple-generator-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 5m -run '^TestNativeAgreesWithNode$/^internal$/^load$/^testdata$/^0.1$/^compile$/^(01_hello.ts|06_stack.ts|10_unicode.ts)$' > /tmp/apple-generator-oracle.log 2>&1
```

All passed. Generate: 2.602 seconds, 79.8% coverage; naming: 6.898 seconds,
90.1%; command: 1.384 seconds, 51.9%. Formatting and vet logs were empty,
and git diff --check was clean. The uncached external oracle passed in
9.602 seconds. The fifteen implementation mutants each exited 1 through an
assertion failure; the additional ABI mutant failed its clang static assertion.
The full repository test suite was not run; the scoped package checks include
the compiler command probe, and the external oracle is explicitly filtered.

Setup completed successfully: Go 1.27.1, clang 20.1.8 and Node 24.19.0,
all ready at 0 seconds, submodules ready at 0 seconds, build-cache warming
92 seconds, total 92 seconds. nproc was 5, with cgroup CPU quota 4 cores.
The printed tool environment is /workspace/adamic-tools/env.sh.
