# Method values as readonly function-typed properties

TypeScript v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8. Task #rv75m9k, road A. Type-only: the emitted JavaScript
doesn't change, and every member it touches is `@internal`, so typescript.d.ts doesn't move.

## Why

tsc reads 32 of its own `@internal` methods as values: it hands `factory.liftToBlock` to visitNode, the parenthesizer's
rules to the printer's list emitters, `host.getCanonicalFileName` to path helpers, and so on. Each read is "a method read
as a value", refused by methodValueRefusal (internal/lower/refusals.go), because a method signature can hold a method
that reads `this`, and the read would lose its object. The rule stays (docs/rulings.md:114): admitting the read needs the
body, and a method signature doesn't give one.

What tsc means by these members is a function, held in a field. Every implementation that reaches them is a closure, an
arrow or a function reference in an object literal (createParenthesizerRules and nullParenthesizerRules,
createNodeFactory, createNodeConverters, the program's and the emit host's object literals), and none reads `this`. So
each becomes what it is used as: a readonly property of the same function type. The read is then a field read.

## The edit

Each member in owners.json, a method signature in an `@internal` interface or an `@internal` member of a public one,
becomes `readonly name: (parameters) => ReturnType;`, built from the signature's own text: type parameters, parameter
list (comments kept) and return type, verbatim. An optional member keeps its `?`. The one overload group,
`ParenthesizerRules.parenthesizeConciseBodyOfArrowFunction`, becomes one property whose type literal holds both
signatures in their order. `Program` merges, so each member is found in the one declaration of it that holds the member.

32 members in two files, 31 in types.ts and ResolutionCacheHost's in resolutionCache.ts: 23 in ParenthesizerRules, 2 in NodeFactory (liftToBlock and cloneNode), 2 in NodeConverters,
EmitHost.getCanonicalFileName, EmitResolver.hasGlobalName, ResolutionCacheHost.globalCacheResolutionModuleName, and
Program's getCommonSourceDirectory and getFileIncludeReasons.

**Left out on purpose:** a member the tree assigns stays a method. Five of the road-A candidates are written after their
object is built, by the save-then-override hosts (CompilerHostLikeForCache's readFile and fileExists, CompilerHost's
hasInvalidatedLibResolutions, onReleaseOldSourceFile and onReleaseParsedCommandLine). Readonly would refuse the write
(stock tsc says TS2540), and a writable field would meet "a method becomes a writable field" wherever a public
CompilerHost is seen as the internal type. They take road B, a view alias at the read, never an arrow: the member is
overwritten on the next line, so an arrow that reads it through its receiver would call the override and recurse.

owners.json pins each member whole, before and after, with SHA-256. The adapter refuses before writing anything on drift,
a missing or duplicate member, a changed overload group, a half-adapted tree, a bare CR or mixed line endings. A second
run writes nothing. `mutant.cjs TREE` restores every member byte for byte.

## Premises a reader can re-check

- **No implementation reads `this`.** By reading: parenthesizerRules.ts holds no `this` outside comments; the factory
  and converters are closures; the host and program members are arrows or function references in object literals.
- **The compiler holds it, not the trace.** A method that reads `this`, converted into one of these function-typed
  properties, stops at the read through the view (erasedMethodField, internal/lower/iteration_origin.go), as Diagnostics'
  probe on #trxxkge showed. So a later implementation written with method syntax is refused there, not trusted.
- **No member is written after its object is built.** The scan that picked the members reads every write in the tree,
  and a written member is left out (above). Readonly makes stock tsc enforce it from here on.

See [PROOF.md](PROOF.md) for what was measured.
