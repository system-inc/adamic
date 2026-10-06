# Lint slice gaps

## 1. Constructor proof carried into another object's method

Observed at the current branch baseline: `gaps/1_nested_constructor.ts` prints
`1` on Node. Stage 0 returns `lower.Refused` at line 7, column 16:

```
Adamic 0.1 refuses this escaping a constructor before every field is set
```

That location is `Inner.touch` calling `this.read()`. `Inner.value` is already
initialized. The unset field belongs to Outer (`label`), and the inner object
has no reference to Outer. The refusal prevents compiling correct TypeScript;
this is an observed compiler defect, not a reference-counting cycle or an
intended language restriction. `TestNestedConstructorGap` requires Node's exact
answer and the constructor-refusal category. Removing the method delegation
made the reduction compile, which isolates the proof-state interaction.

Inference from the reduction and the inspected `unsetUntil` code: lowering a
callee while a caller's constructor is active carries the caller's initialization
boundary into the other object's method. No compiler or runtime code is changed
in this unit.

The initial linter constructed its parser inside its constructor and encountered
this refusal in `Parser.token`, at its delegated `this.make` call. Workaround:
construct the parser and scanner in the driver, pass the initialized objects to
Linter, and parse/walk from `run()` after Linter's construction. The ordinary
three-way tests compile and execute that workaround.

## Existing parser representation gaps used by this slice

Go AST nodes own parent pointers. Adamic cannot retain those alongside owning
children: reference counting cannot free the cycle. The linter uses child
indexes from the parser and a parallel numeric parent table. The inherited
proving program and test are in
[the parser gap directory](../../typescript/parser/gaps/1_strong_ast_parent.ts)
and `TestStrongAstParentGap`; see
[parser GAPS.md](../../typescript/parser/GAPS.md). Concrete Parser, Scanner and
Finding classes avoid the parser's observed structural-interface method bug;
this slice adds no new interface-method workaround.

## Limits rather than language gaps

These five rules require no type information. Other lint rules, suppression,
config decoding, JSX, JSDoc lint traversal, malformed-source recovery, and a
fix engine resolving competing edits are outside this unit. The current parser
accepts TS source files used by the corpus; the Go oracle rejects diagnostics
before comparing. Non-UTF-8 input is not covered. No finding or repair from an
unsupported additional rule is claimed.
