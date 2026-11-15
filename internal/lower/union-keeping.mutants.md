# Union keeping stop: library and recursion repairs

This head is stacked on 37d96306. It changes a path revisit to false, retains
64 active types as a fail-closed limit, and caches nothing across queries.
Concrete identity is checked before the conservative union stop.

## Library members walked again

Remove the all-library declaration shortcut in `mentionsWeak`, retaining the
argument walks. Run:

```
go test ./internal/lower -run '^TestUnionKeepingRegExpLibrary$' -v
```

The pin uses `const p: string | RegExp = "[[/]/"`, constructs `new RegExp(p,
'u')`, and prints source, flags and the match result, like
`RegExpUnicodeClassSource`. Observed exit 1 in both extensions: the mutant refuses the initializer at
line 1, column 28 with the weak sentence. Named red tests:
`TestUnionKeepingRegExpLibrary/.a` and `.ts`.

## A path revisit answers true

Change only the visited-path arm from false to true. Run:

```
go test ./internal/lower -run '^TestUnionKeepingRecursiveNoWeak$' -v
```

The pin views a recursive readonly union both as itself and as a structural
twin. Identity admits the first view; the twin must complete its Weak search
without treating a cycle as a Weak. Observed exit 1 in both extensions: the mutant refuses the twin view at
line 6, column 20 with the weak sentence. Named red tests:
`TestUnionKeepingRecursiveNoWeak/.a` and `.ts`.

Both mutants are measured with build overlays. Production files are never
replaced by their mutant versions, and the fixed source is compared afterward.

## Library arguments and declaration ownership

`weakTarget` is asked before any library shortcut. Alias and reference arguments
are walked before skipping library properties, index values and signatures.
`TestUnionKeepingLibraryWeakArgument` views a union holding readonly Weak<Box>[]
as its strong twin. The twin retains undefined, so TypeScript accepts the view;
the assignment stops at line 7, column 24 in both extensions.

The shortcut requires every symbol declaration to satisfy `load.IsLibrary`,
`load.IsNodeLibrary` or `load.IsPrelude`, and requires at least one declaration. The prelude's own Weak
alias is the sole permitted type reference to Weak/WeakBrand there;
`TestUnionKeepingPreludeDoesNotHideWeak` walks its declarations to guard this
assumption. A directly written Weak is still found before the shortcut.

The loader treats source files as modules: a top-level interface Array<T> is
local, not a merge with the library Array. `TestUnionKeepingLocalArrayDoesNotMerge`
pins the local held member and the absence of held from the library array, in
both extensions. No accepted program merge pin or any-declaration mutant is
claimed. The conservative every-declaration condition is retained unchanged.

## Preserved stops and identity

The p5–p8 assignment pins, two-deep Weak pin and generic view<Box> return pin
remain NotYet. P9 keeps its original stop. P7 uses a function declaration to
reach its cross-view assignment instead of stopping at a typed arrow initializer.
Known leaves, checked downcasts, optional reads and no-Weak union controls retain
their behavior. An identical mapped Weak union now lowers and agrees with Node.
The depth-limit pin uses distinct twins and stops at line 76, column 20.

## Instantiated library objects must not use the named-type shortcut

Only ObjectFlagsReference, ObjectFlagsInterface or ObjectFlagsClass can take
the library shortcut. Mapped, reverse-mapped and evolving-array objects fail
closed above it; anonymous objects continue through their members and signatures.
Their instantiation mapper can carry program types not exposed as arguments.

`TestUnionKeepingInstantiatedLibraryObjects` checks both extensions against the
real checker. Its input has v: Weak<Box>. Object.getOwnPropertyDescriptors(x)
produces a mapped half, which must answer true. It also takes a Weak<Box>[]'s
push method: that type is anonymous, has a library symbol and no alias/reference
arguments, yet its instantiated parameter carries Weak. It too must answer true.
No executable descriptor row is claimed: production already refuses
Object.getOwnPropertyDescriptors through its property-descriptor rule.

Mutant: remove only the Reference/Interface/Class restriction, allowing any
all-library symbol to skip. Run:

```
go test ./internal/lower -run '^TestUnionKeepingInstantiatedLibraryObjects$' -v
```

Observed exit 1: the anonymous query returns false, redding both extension rows
with `instantiated library anonymous type hid Weak`. The mapped query remains
fail-closed above the shortcut. The mutant is built with an overlay, leaving
production unchanged. RegExp is still an interface and retains its shortcut.

Node declaration files use load.IsNodeLibrary, separately from TypeScript's
load.IsLibrary. Both are accepted for named declaration types, alongside the
prelude; the every-declaration requirement still excludes program declarations.
This avoids descending through Buffer/URL's library graphs into unrelated
mapped members such as Array[Symbol.unscopables], while anonymous or mapped
library instantiations still cannot take the shortcut.

## Library aliases and tuple references

Library aliases are checked through every alias argument before their mapped
body can take the fail-closed arm. Union/intersection traversal remains ahead of
this shortcut so the prelude's Weak union still reaches weakTarget. Tuple
references are checked through their element arguments before returning false;
they do not descend into Array's anonymous mapped Symbol.unscopables member.
Alias-less mapped and anonymous library instantiations retain the previous rules.

`TestUnionKeepingLibraryAliasAndTupleControls` pins the Record object-tag NotYet
at 4:5 and the tuple-destructuring NotYet at 4:7, in .a and .ts.
`TestUnionKeepingLibraryAliasAndTupleWeakArguments` uses a function parameter to
reach the view assignment without first constructing a Weak union. Both Record
and tuple Weak arguments still stop with the weak sentence at 6:25, in both
extensions. The strong twin retains undefined so TypeScript accepts the view.

Removal mutants, run through Go overlays without changing production:

```
go test -overlay /workspace/shared/union-keeping-alias-mutant.json ./internal/lower -run '^TestUnionKeepingLibraryAliasAndTupleControls/record' -v
go test -overlay /workspace/shared/union-keeping-tuple-mutant.json ./internal/lower -run '^TestUnionKeepingLibraryAliasAndTupleControls/tuple' -v
```

Both exit 1. Removing the alias rule reds record.a and record.ts: the weak
sentence at 3:79 replaces the object-tag NotYet. Removing the tuple rule reds
tuple.a and tuple.ts: the weak sentence at 2:9 replaces the destructuring NotYet.
The fixed production source is unchanged by either mutant.
