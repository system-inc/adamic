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

## 2 and 3. Illegal method bodies in an interface or type literal

The unchanged Go rule tests contain these two malformed inputs:

- [2_interface_method_body.ts](gaps/2_interface_method_body.ts)
- [3_type_method_body.ts](gaps/3_type_method_body.ts)

Go's parser recovers a parameter whose `any` produces `unexpectedAny` at byte
ranges 45..48 and 38..41 respectively. The Adamic parser running on Node and
sanitized native exits 70 with `parser slice unsupported primary CloseBraceToken`
at offsets 54 and 47. `TestBatch5RecoveryGaps` runs the proving programs through
the actual driver and compares those observations, with a two-second process
bound. `TestRulesAgree` identifies precisely the same two upstream inputs and
checks their refusal; it never counts them among agreeing bytes.

Observed: both backends refuse promptly rather than timing out. Inference: the
parser's type-literal member loop does not recover an illegal signature body in
Go's way. The parser is outside this unit's ownership and remains unchanged.

## Limits rather than language gaps

These fifteen rules require no type information. Config, suppression, binder and
type checker integration, general JSX parsing, and general malformed-source
recovery are outside this unit. The selected rules' decoded options are supported;
this is not the ESLint configuration decoder. Non-UTF-8 input is not covered.
The fix engine compares complete valid proposals and final sources. Its bounded
loop preserves the earlier batch's overlap and pass-budget machinery; these ten
rules do not supply a competing or oscillating edit control. Behavior for arbitrary
invalid third-party proposals is not claimed. Strong parent ownership and the inherited constructor proof gap
remain represented by numeric ancestry and by driver-side object construction.
