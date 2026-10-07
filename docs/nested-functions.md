# Named nested function declarations

Design written before implementation, against origin/main ef3d907.

## Existing implementation

Arrows lower to an IR function with Closure set and a MakeClosure expression.
Ordinary function expressions use the same convention, reject dynamic this, and
bind a named expression to ClosureSelf. Each captured local moves into one counted
cell. Every intervening closure carries that cell; writes update it and reads
snapshot its value before another call can change it. A closure owns its cells,
its caller owns a returned closure, and a containing array owns its elements.
The cycle finder follows function types into every compatible closure's captures.
Captured cells are never relaxed by the fresh-write proof.

## Representation and scope

Keep the existing counted cell representation. A frame's nested declarations
share one logical environment: an ordered vector of the same captured cells.
Each declaration's closure record contains its code and references to those
cells. This duplicates the vector of pointers, rather than allocating another
record to own the vector. It retains the established arrow calling convention
and destruction paths, at the cost of one retain per cell per declaration.
Measure that cost in the oracle counts table. Do not claim it is free.

Every sibling has exactly the same environment layout. A direct sibling call
calls the sibling's code with the current environment carrier. Neither function
captures a variable holding the other function. Mutual recursion adds no strong
edge between function values. Returning a declaration from its enclosing frame,
passing it, or storing it in an array keeps its counted closure and cells alive.
Every invocation of the enclosing function creates new cells. Three lexical
levels forward the same ancestor cells through the intermediate closure.

Initially accept declarations directly in a function body. Block-scoped
function declarations, generic nested declarations, dynamic this, rebinding a
nested declaration, first-class sibling references inside another sibling, and
calls to a declaration in a different ancestor group remain loud NotYet cases.
These are implementation gaps, not permanent language refusals. A sibling call
is supported; returning a sibling value from inside another sibling is not yet
supported, because preserving declaration identity needs a separate binding
strategy. Never substitute a freshly allocated closure and change ===.

## Cycles

A captured variable holding a nested function that reaches that variable closes
a cycle. Register every nested function in the existing cycle finder, including
the complete shared environment, and refuse this with adamic/cycle-capable.
No fresh-write relaxation for cells. Shared sibling environments can retain more
cells than a function reads, so the finder must use the complete retained vector,
not merely syntactic reads. This deliberately makes refusal conservative.

For example, refuse:

```a
function make(): () => number {
    let saved: (() => number) | undefined = undefined;
    function read(): number { return saved === undefined ? 0 : saved(); }
    saved = read;
    return read;
}
```

Restructure so saved is not captured, or declare saved Weak<() => number> and
keep read in a separate strong owner. The same rule applies when the captured
slot is an array or an object containing the closure. Never silently leak.
A mutant that captures sibling declaration bindings as ordinary closure cells
must pass output comparison and fail the leak check, proving that output alone
cannot certify this representation.

## Hoisting and the temporal dead zone

Prebind all direct let/const names and nested function signatures before lowering
any sibling body. Allocate captured cells at block entry, empty and not ready.
Then initialize nested declaration values before source statements. Original
let/const initializers still execute at their original lines. A cell's ready bit
becomes true only after initialization completes. Captured parameters start ready.

Extend the existing Checked read/write and ready-check machinery to captured
cells. Reads check before fetching the value. Assignments evaluate their right
side before checking readiness, as JavaScript does. A call before a nested
function's declaration is valid. A call before a captured let/const declaration
panics with ReferenceError: Cannot access 'name' before initialization in both
backends. A not-ready reference cell contains NULL so unwinding never releases
uninitialized storage. Initialization is distinct from assignment.

The JavaScript backend must use explicit cells and readiness as native does;
Node executing the original source is the independent TDZ oracle. Preserve the
existing limitation that inserted panics do not become catchable exceptions.

## Analyses and ownership

Keep captured variables excluded from SSA and reuse sources. A nested call can
write captured state, so direct code dispatch must retain the conservative
CallClosure effects for freshness, mutable ranges, array-element borrowing and
lent reads. Do not promote it to a pure named call just because its target is
known. Nested parameters keep the closure convention's owned counts. Captured
values and closure environments stay on the heap; regions must not allocate
anything that a closure can keep. No global disabling of regions or Perceus.

Preallocation adds empty declarations before initialization. Analyses must see
the real initialization and every subsequent write. The ready bit is not a
proof of freshness. Captures still escape even when the body only calls another
sibling. Throw and return cleanup release frame-owned cells and closure values
exactly once; escaped closures keep their own cell references.

## Evidence required before completion

Run minimal/r01, read/write captures, hoisting, even/odd, returned closure,
array-stored closure and three-level capture fixtures against original-source
Node, both backends, ASan/UBSan and LeakSanitizer. Add refusal probes for every
listed gap and a captured function cycle. Record counts. Run mutants for sibling
capture cycles, lost write-through, lost hoisting and omitted readiness checks.
Report actual observations separately from predicted results.

The pinned census is 429c117 and its filename is stage3/census/REPORT.md.
Its 5,574 sites in 57 files are a source inventory, not 57 successful checker
runs reaching NotYet. The corpus stops at checking. Rerun the inventory and
actual minimal probes separately; never describe a removed diagnostic branch
as proof that tsc compiles. Count files clearing this specific source-form gap
only after classifying retained NotYet cases, rather than claiming all 57 clear.
