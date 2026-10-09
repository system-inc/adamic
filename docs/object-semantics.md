# Object semantics

This is the bounded library contract for roadmap step 22 (#nzbprw9). Node 24.19.0
is the authority on Linux, JavaScript and WASI. Library lowering uses the existing
counted IR and runtime; it introduces no POSIX dependencies.

Complete plain literal shapes and their unannotated bindings support Object.keys,
getOwnPropertyNames, values and entries. Numeric literal field names use the
checker's canonical spelling. Canonical array indices from 0 through 4294967294
sort numerically; other strings retain insertion order, including 4294967295,
fractional keys, leading-zero strings and negative-zero strings. The existing
for...in implementation shares this ordering on proven plain-object origins.
It continues to refuse arrays, unknown origins and observable prototype chains.
Object.hasOwn accepts dynamic strings and numbers on its existing proven shapes;
it tests actual own-property presence and does not require the key to be declared.
Numbers undergo the existing JavaScript number-to-string conversion.

Immediate Object.prototype.toString.call observations support exact null,
undefined, number, boolean and string primitives, arrays, functions, and complete
plain shapes. The receiver evaluates once. Own toString overrides do not change
this intrinsic's tag. Boxed primitives, exotic objects, widened object views,
unknown tags and Symbol.toStringTag remain NotYet.

Immediate hasOwnProperty.call and propertyIsEnumerable.call support the existing
represented primitive, dense-array and function descriptors, plus complete plain
shapes, with string keys. Boxed strings expose UTF-16 indices and non-enumerable
length. Plain data fields are enumerable; inherited fields are absent. Regex
result arrays with additional own descriptors are refused. Nullish receivers
and keys requiring unrepresented ToPropertyKey conversion remain NotYet.

The Object constructor, Object.prototype and six standard Object prototype
methods allow static own-property observations. Their intrinsic data properties
are non-enumerable. The six methods and standard static Object methods allow typeof, truthiness, name, length and
observation of their absent prototype property. These observations do not allow
prototype objects or detached callable values to escape. Mutation of library
prototypes stays refused. Metadata is pinned to Node 24.19.0.

Number on complete plain shapes follows V8's OrdinaryToPrimitive number hint:
read valueOf before toString, skip noncallable members, continue after object
results, and stop at the first primitive result. The default valueOf returns the
receiver; the default toString returns the Object tag. Own callbacks must be
proven zero-argument arrows with number, string, boolean or object results.
Thrown callback errors propagate without running the fallback. Mixed results,
getters, Symbol.toPrimitive, implicit this, parameterized callbacks and the final
TypeError path stay refused. String retains its existing string-hint conversion,
which tries toString before valueOf.

Descriptor mutation, prototype creation or replacement, arbitrary prototype
traversal, boxed-value construction and dynamic records remain outside this
slice. Declaration and expression features rejected by the checker or compiler
are not rewritten into different programs to bypass those refusals.
