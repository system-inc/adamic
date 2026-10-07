# Namespaces in Adamic 0.2

Decision for Kirk, October 6, 2026: a sound qualified-name subset exists, and
this branch implements it. Admit a single module-scope declaration, nested
namespaces, interfaces/type aliases, ordinary or generic functions, exported
constants, classes, and initialized private `let`/`const` bindings. Namespace functions
can be called, detached and compared by identity. Names can coexist with an
interface or type alias of the same spelling. Type-only namespaces erase.

The namespace object's identity, reflection, escape and mutation are outside

## Stage 3: parser singleton storage

Direct private namespace `var` bindings are hoisted to namespace entry. Their
initializers run in source order; an uninitialized declaration does not overwrite
an earlier assignment. Uninitialized private `let` is also admitted. Storage
whose declared type includes undefined begins with undefined. Storage whose type
excludes undefined stays unready until assigned; reads use the existing ready
check in both backends, rather than treating native zero bits as a typed value.
This check intentionally stops a checker-accepted type lie where Node reads
undefined. Function-local var, destructuring namespace declarations, mutable
exports and calls during namespace initialization remain outside this step.

The parser-state fixture observes a var before its declaration, initializes a
built string through an exported function, and resets optional token state.
The unready fixture proves a typed number read before assignment stops with the
same ready check in native and generated JavaScript. This is singleton storage
support, not proof that parser.ts or all its function bodies compile.
