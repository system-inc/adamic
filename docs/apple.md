# Adamic on Apple's platforms

An Adamic program calls AppKit and Foundation directly: no Objective-C or Swift is written, and none is generated. The compiler turns each call into a typed `objc_msgSend` in the C it writes, and clang links the program against Apple's frameworks.

```ts
import { Application } from 'apple/appkit/application';
import { Button } from 'apple/appkit/button';
import { Window } from 'apple/appkit/window';

const window = new Window({ contentRectangle: { x: 0, y: 0, width: 480, height: 240 }, styleMask: ['Titled', 'Closable'], backing: 'Buffered', defer: false });
window.title = 'Adamic';
const button = new Button({ title: 'Press', action: () => console.log('pressed') });
window.makeKeyAndOrderFront(undefined);
Application.shared.run();
```

`examples/apple/window.a` opens a window with a label and a button whose closure counts presses: `adamic build examples/apple/window.a -o window && ./window`.

## Why it fits

Apple's object model is reference counting with no collector, which is Adamic's. An Objective-C object Adamic holds is one strong reference, let go of on Adamic's last release, so the two counts meet at exactly one place and neither runtime has to guess about the other.

## Bindings

What an `apple/` module exports is declared in a binding file, served at `/adamic-apple/<framework>/<module>.d.ts` and loaded only when a program imports that module (`internal/load/apple.go`). Foundation's, AppKit's and CoreGraphics' are generated from the Mac's own SDK (`internal/apple/generate`, its README says how), the first time a program imports one: 1,383 modules in about six seconds, written into `~/Library/Caches/adamic/apple/macosx-<SDK build>-<generator version>/` once their check file compiles against the SDK's headers, so a binding whose tag disagrees with its header never reaches a program. `ADAMIC_APPLE_BINDINGS` names a directory to read instead, which is how a machine with no SDK compiles against generated fixtures. A declaration deprecated on the platform is left out, as Apple's "don't use", listed with its reason in its module's comments and in `internal/apple/deprecated-macos.txt` (868 on macOS 27.0's SDK). Each declaration keeps Apple's name in its doc comment and carries the Objective-C it calls in an `@objc` tag:

```ts
/**
 * -[NSWindow initWithContentRect:styleMask:backing:defer:]
 * @objc init initWithContentRect:styleMask:backing:defer: 0.contentRectangle:rectangle 0.styleMask:options(Borderless=0,Titled=1,Closable=2,...) 0.backing:enum(Retained=0,Nonretained=1,Buffered=2) 0.defer:boolean
 */
constructor(options: { readonly contentRectangle: Rectangle; readonly styleMask: readonly WindowStyleMask[]; readonly backing: BackingStoreType; readonly defer: boolean });
```

- `@objc class <Class>` on a class names the Objective-C class.
- `@objc init <selector> <arguments>` is a constructor: `alloc`, then the init.
- `@objc static <selector> <arguments> -> <result>` is a message to the class (a constructor may be one, as `+[NSButton buttonWithTitle:target:action:]` is).
- `@objc method <selector> <arguments> -> <result>` is a message to the object.
- `@objc get <selector> -> <result>` and `@objc set <selector> <type>` are a property's getter and setter.
- `@objc alloc <Class> <selector> <arguments> -> <result>` sends `alloc` to another class, then the init, its result retained: `data.utf8Text()` is `[[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding]`.
- `@objc function <symbol> <arguments> -> <result>` is a C function: on its own, as a class's static member, or as an instance member, where the object is its first argument (`CFRunLoopStop(loop)` is `loop.stop()`).
- `@objc protocol <Protocol>` on an interface names the Objective-C protocol, and `@objc implement <selector> <arguments> -> <result>` on one of its methods says what Apple hands a program's class that implements it, and what it takes back (Delegates, below).

Each argument is `<source>:<type>`, in the selector's order. The source is the Adamic argument's position (`0`), a field of an options object written at the call (`1.styleMask`, or `1.defer?:boolean=no` where the field may be left out), the object the member is called on (`this`), or a constant (`nil`, `yes`, `no`, or a number, `const(4)`). The types:

| Tag | Native | Adamic |
| --- | --- | --- |
| `double` | `double`, `CGFloat` | `number` |
| `integer`, `unsigned` | `NSInteger`, `NSUInteger` | `number`, truncated toward zero, NaN as 0, clamped to the range |
| `boolean` | `BOOL` | `boolean` |
| `string` | `NSString *` | `string`, crossing as UTF-16 units, a lone surrogate included |
| `string?` | `NSString *`, `nil` allowed | `string \| undefined` |
| `object`, `object?` | any object, `nil` allowed with `?` | a class from a binding file, `T \| undefined` where `nil` is a value |
| `rectangle` | `CGRect` | `{ x, y, width, height }` |
| `enum(Name=value,...)` | an integer | a string literal union |
| `options(Name=bit,...)` | a bit mask | a readonly array of the literals |
| `action` | a target and its selector | a closure, `() => void` |
| `block(type,...)` | a block returning nothing | a closure taking those parameters |

A result marked `-> new object` comes back retained (`alloc`, `new`, `copy`); any other object result is retained on its way into Adamic. A result Apple promises isn't `nil` (`object`, `string`), and whatever `new` makes, is checked: `nil` there panics, naming the selector, rather than reaching Adamic as a value its type says can't be. So `new Url({ string: 'not a url' })` panics, as Swift's `URL(string:)!` would.

A few binding files are written by hand and embedded in the compiler: whole modules where nothing is generated (SwiftUI's, over the Swift shim; Core Foundation's run loop, `apple/corefoundation/run-loop`), and additions to a generated module, served beside it and merged into it (`apple/foundation/data` adds `utf8Text()`, two messages to another class, to the generated `Data`).

## SwiftUI

SwiftUI is Swift only: its views are generic value types no C can name. So `internal/native/apple/swiftui.swift` makes them Objective-C classes the bridge already reaches: an `AdamicSwiftUIView` holds an `AnyView`, made by class methods (`text:`, `button:action:`, `verticalStack:children:`) and changed by modifiers (`padding:`, `font:`), and an `AdamicSwiftUIHost` shows one in an `NSHostingView` and swaps it when the program renders again. It's compiled with `swiftc` once per version of it and of `swiftc`, cached under the user cache directory, and linked only into a program whose C names one of its classes. This first cut is written by hand; the generator writes it per API an app uses (#7xv3pcs), and then it's measured against calling Swift's stable binary interface directly.

```ts
const count = new State<number>(0);
function body(): View {
	return verticalStack({ spacing: 12 }, [
		text(`Count: ${count.value}`).font('Title'),
		button('Add one', () => count.set(count.value + 1)),
	]).padding(20);
}
const host = new Host(body());
renderWhenStateChanges(() => host.render(body()));
window.contentView = host.view();
```

`State` (`apple/swiftui/state`) is written in Adamic, not declared: some `apple/` modules are Adamic code, embedded as `internal/load/apple/<path>.a` and served by the loader where the resolver looks for any package, one module however many directories import it (`internal/load/apple.go`). Setting a `State` renders again through whatever `renderWhenStateChanges` was given. Bindings for plain functions that message a class use `@objc send <Class> <selector>`, and an array of views crosses as an `NSArray` (`objects`).

A view may be a class of the program's own, as SwiftUI's are structs of the app's:

```ts
class SupplementRow implements View {
	readonly supplement: Supplement;
	constructor(supplement: Supplement) {
		this.supplement = supplement;
	}
	body(): View {
		return horizontalStack({ spacing: 8 }, [image(this.supplement.symbol), text(this.supplement.name), spacer(), text(this.supplement.dose).foregroundColor('Secondary')]);
	}
}
```

`apple/swiftui/views` and `apple/swiftui/host` are written in Adamic over the shim's bindings (`apple/swiftui/native`, `apple/swiftui/hosting`). A `View` is anything with a `body()`. What the view functions make is a `Shown`, holding the shim's view, whose body is itself and whose modifiers each make a new `Shown`. Before a view crosses (a stack's children, a host's root), `rendered` asks each body for the next until it reaches a `Shown`, in Adamic, and only that `Shown`'s Objective-C view crosses.

So ownership stays simple. SwiftUI holds only the shim's objects, by Objective-C's count, and never an object of the program's: no class of the program's is reached from native code, and none is kept alive by Swift past a render. A closure a button holds is the one thing of Adamic's SwiftUI keeps. It's a block holding the closure (as everywhere in the bridge), let go of when the shim forgets the button, at the latest as the program finishes. A view that captures its own model through a button's closure is the same cycle as an action's (below, Not yet).

`examples/apple/counter.a` is the app.

## Delegates

An object of the program's own class is handed to Apple where Apple takes a delegate, a data source, any protocol, and Apple calls its methods:

```ts
class Keeper implements WindowDelegate {
	asked = 0;
	windowShouldClose(sender: Window): boolean {
		this.asked += 1;
		return this.asked > 1;
	}
	windowWillClose(notification: Notification): void {
		console.log('closing');
	}
}
window.delegate = new Keeper();
```

A protocol is an interface whose methods take their arguments in the selector's order, each tagged with what crosses: `windowShouldClose?(sender: Window): boolean` is `@objc implement windowShouldClose: 0:object -> boolean`. An optional method is an optional member. Swift tells a protocol's methods apart by their labels (`tableView(_:objectValueFor:row:)` beside `tableView(_:viewFor:row:)`), and a class has one method per name, so where two share a name each folds its labels in: `tableViewObjectValueForRow(tableView, tableColumn, row)`.

The class says which protocols it implements (`implements WindowDelegate`), as an Objective-C class does: Apple asks a delegate what it responds to, and the class's own `implements` is what's answered. An object of a class naming none is refused where it's handed to Apple.

Lowering makes an Objective-C class for the class, the first time one of its objects crosses (`internal/lower/foreign_delegate.go`): NSObject's subclass, conforming to the protocols, with a method for each protocol method the class has, and nothing for the ones it doesn't, so `respondsToSelector:` is the truth. Each method is emitted C (`internal/native/foreign.go`, `delegateClass`) that converts what Apple hands it, calls an ordinary Adamic function lowering made, and gives back its result as Apple takes it. That function calls the class's method, virtually, as any call through the class does, so a subclass's override answers, and every analysis sees the method called as code. Nothing in the program calls the function: Apple does, through the object it was handed.

- **Ownership.** The delegate holds the object, one count, let go of in its `dealloc` (on the main thread). Apple holds a delegate weakly, so the object it's handed to holds it, as it holds an action, one per selector that sets one: setting another, or `undefined`, lets the last go.
- **Cycles.** A delegate that can reach back to what holds it is a cycle neither count sees, so the cycle finder refuses one, by the rule it holds a field to (`internal/lower/cycles.go`): a `Keeper` with a `window: Window` field can't be that window's delegate. `window: Weak<Window>` can. Strong and refuse is ruled (October 7): Adamic keeps a delegate, so a forgotten one can't vanish as Swift's do.
- **Leaves.** The finder follows Apple's declared properties, and through them nearly every Apple class reaches every other, so it doesn't walk through a leaf: a class whose instances, and every subclass's, hold strong references only to other leaves. `Url`, `String`, `Number`, `Value`, `Data` and `Date` are leaves, so a delegate may hold a URL. Collections never are. The table is derived from the headers (`internal/apple/generate/leaves.go`), never listed by hand: what an object may hold is its non-weak properties, what its initializers, factories and mutators are handed, and any escaping block, and a reference reaches only leaves when its type does. Eight facts the headers can't state (`initWithCoder:` reads its coder, a `locale:` argument is read) are cited in the table. It's generated beside the bindings, so the compiler always reads the one for its SDK, and checked in as `internal/apple/leaves-macos.txt`, which a test holds to the SDK, so a new reference path can't arrive silently. A program's class can't extend an Apple class yet, which is what keeps a leaf's own fields all it holds.
- **What a class holds.** For an Apple class that isn't a leaf, the finder follows what the same derivation says it may hold, beside its declared properties: what its methods keep (a subclass's `stickTo:` handed a record), read from `holds.txt` beside the bindings. A class the program never loads is followed by its Objective-C name; any object (`id`) is any class's holds; a protocol is the program's classes implementing it; a block is every closure the program hands Apple. So an Apple object that isn't a leaf may reach every closure handed to Apple, and a closure handed to Apple that captures one in a local variable is refused (a label an action sets, built in a function), the fix capturing it `Weak<TextField>` or keeping it in a module-level constant.
- **Threads.** Apple calls a delegate's methods on the main thread, where Adamic's counts are kept, and a method answers before Apple goes on, so it can't wait for the main thread as a block's closure does: a call anywhere else panics.
- **Reading one back.** `window.delegate` reads the Objective-C object Apple holds, the delegate, as an Apple object, not the program's own object that it holds.

`internal/apple/testdata/delegate.a` is held to `delegate.m`, the same classes written in Objective-C: a window that asks twice before closing, a subclass whose override answers through a reference typed as its base, a delegate replaced and one cleared, and a table view's data source. The checks fail when the delegate's `dealloc` keeps its object (58 allocated, 52 freed), when what was kept is never let go of (58, 52), when nothing keeps the delegate (the witness disagrees: AppKit's weak reference finds nothing), when the call isn't virtual, when a boolean result is dropped, and when the delegate doesn't retain its object (the sanitizer stops the program).

## How a call is compiled

Lowering (`internal/lower/foreign.go`) makes each call an ordinary `ir.Call` of a function whose `ir.Function.Foreign` describes the message: its kind, class, selector, and where each native argument comes from. Its parameters are the receiver, then the values the call evaluates, in the order the source evaluates them, so JavaScript's evaluation order holds. One function is made per class, selector and shape of call.

The native backend (`internal/native/foreign.go`) emits that function's body as the conversions, the message, and the result:

```c
static SEL selector;
if (selector == NULL) {
	selector = sel_registerName("setTitle:");
}
id native_0 = adamic_apple_unbox(adamic_local_14_argument0);
id native_1 = adamic_apple_string(adamic_local_15_argument1);
((void (*)(id, SEL, id))objc_msgSend)(native_0, selector, native_1);
adamic_apple_let_go(native_1);
```

The function's IR body only panics, so the JavaScript backend, which can't reach AppKit, says so out loud at the first Apple call.

## Ownership

- **Objects.** An Objective-C object in Adamic is a foreign value (`runtime/foreign.c`): an object of no fields to everything that looks at objects, so a path that reaches one by name panics rather than reading memory that isn't there, and its last release calls `objc_release`. An object has one box while Adamic holds it (`apple.c`, a table by the object's address), so the object Apple hands back is the value Adamic gave it (`window.contentView === content`), and a crossing of an object already held makes nothing. Every reference parameter of a foreign function is borrowed.
- **A constructed window** is told not to release itself when closed: Adamic's count owns it.
- **Actions.** A closure given as an action becomes an `AdamicAction`, an `NSObject` subclass made at runtime whose instance variable holds the closure (counted) and whose `adamicAct:` calls it. A control holds its target only weakly, so the action is kept by the object it was given to, as an associated object, and its `dealloc` releases the closure.
- **Blocks.** A closure given as a block is a block laid out by hand as the block ABI lays one out, no `-fblocks` needed: it starts on the stack of the call it's given to, holds the closure through an object whose count is Objective-C's (safe on any thread), and Apple copies it if it keeps it. Apple calls a block on whatever queue it likes, and Adamic's counts aren't atomic, so the block's invoke only retains what it was given and hands it to the main thread; there the arguments become Adamic values and the closure runs. A block called on the main thread runs at once, so an enumeration finishes before the call that enumerates returns. Entering Adamic from a block off the main thread panics: that's a runtime bug, never a race.
- **The pool.** `main` runs inside an autorelease pool, drained when the program finishes, after its globals are released, so whatever Apple held of Adamic's comes back and is freed before the counts are taken.
- **Waiting.** A program that isn't an app waits for its callbacks in the main thread's run loop: `RunLoop.run()` until a callback calls `RunLoop.main.stop()`.
- **Exceptions.** An Objective-C exception nothing catches ends the program as a panic, with its name and reason, before anything unwinds through Adamic's frames. An Adamic throw out of an action is uncaught, since nothing in Objective-C can catch it.
- **Output** is flushed each time the main run loop is about to wait, so an app that never exits still shows what it printed.

## How it's proven

Node can't run AppKit, so an Apple program answers to a witness instead (`internal/apple/witness_test.go`): every `internal/apple/testdata/<name>.a` has a `<name>.m` that makes the same calls in Objective-C and prints the same lines, built by Apple's toolchain with ARC, independent of everything Adamic's compiler does. The programs compile against the bindings generated from the SDK, as any program does.

1. Both are built under the address and undefined-behavior sanitizers. The program's stdout must equal the witness's byte for byte, and both exit 0.
2. A counted build must free every Adamic value it allocated.
3. The same counted build must owe no Objective-C reference: each one a conversion makes (an `NSString` for an argument, an action before its owner keeps it) is counted when made and when let go of.

macOS's `leaks` tool isn't one of the checks. Inside an AppKit process it reports nothing for an object leaked on purpose, in plain Objective-C as in Adamic, though it finds the same leak in a program that only uses Foundation. A check that can't fail proves nothing.

The test serves JSON from a server of its own and gives every program its address. `window.a` drives a window, a view, a label and a button through real target-action; `fetch.a` makes two requests through `URLSession`, one fetching the JSON and one refused, each completion a closure Apple calls as a block on a background queue; `counter.a` is a SwiftUI counter, held to `counter.swift`, the same view written as plain SwiftUI (where the API is Swift only, the witness is Swift). Both read the size SwiftUI lays the view out at: ten presses make the count a digit wider, 122 by 102 points becoming 131.5 by 102, so a render that didn't happen shows.

A SwiftUI button is pressed by running the block its `Button` holds, in both. SwiftUI's accessibility tree is empty until an assistive client connects, and its buttons aren't `NSButton`s, so nothing in process can reach the button itself: SwiftUI's own tap dispatch is the one step these tests don't exercise.

Each check has been shown to fail: dropping the closure's release in an action's `dealloc` (allocations 47, frees 46), a box that retains what it was handed already retained (47, 46), draining the pool after the counts are reported (47, 46), dropping the last UTF-16 unit of a string coming back (the witness disagrees), never letting go of the `NSString`s made for arguments (73 owed), a block's dispose that keeps its holder (26, 24), a block's deliver that keeps the values it made (26, 23), and a block's invoke that calls Adamic without the hop to the main thread (the program panics), a `State` that doesn't render when it's set (the counter disagrees with its witness), and the shim keeping its buttons' actions past exit (119 allocated, 118 freed), and for the boxes: a crossing that doesn't look for the object's box (`window.a` disagrees: the content view isn't `===` itself), a box that never leaves the table (the sanitizer finds the freed box used), and a lookup that doesn't compare the object.

## Not yet

- **Cycles through actions.** An action's closure that captures something holding the control it's attached to is a cycle neither count can see. A delegate's is refused (Delegates, above); an action's isn't yet.
- **The rest of the bridge:** blocks that return a value or take a block, struct or enumeration; an Adamic class that extends an Apple class and overrides its methods (an `NSView` whose `drawRect:` is Adamic's); a delegate called off the main thread (a session's delegate queue), and delegate methods that take a block, a rectangle or an enumeration; rectangles, enumerations and options as results; options objects passed as a value rather than written at the call; compound assignment to an Apple property.
- **Retained results.** The release of a result that comes back retained (`alloc`, `new`, `copy`) is one line of emitted C that no check counts yet: dropping it would leak without the owed count noticing. So is the release a retained result gives back when its object already has a box (`copy` of an immutable object is the object): the mutant that drops it survives every witness.
- **The analyses.** A foreign callee is taken as unknown by region planning; the other analyses read its IR body, which only panics, until every analysis asks one place what a call can do (internal/ir/call_targets.go, landing from codex/call-targets), where a foreign callee will answer unknown. That matters beyond ownership: a foreign call can run Adamic code before it returns (`performClick` runs the button's closure), so nothing may assume a variable is unchanged across one. The native side never keeps a value without retaining it, so the counts hold either way.
