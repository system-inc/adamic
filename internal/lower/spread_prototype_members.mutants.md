# spread_prototype_members.go's mutants

Each is run on Loom as its own red job, and each goes red on exactly the rows named.

- **The refusal answers nil** (`checkSpreadPrototypeMembers` returns nil before its loop): a .a spread whose
  type declares as a field a member a class reaching it holds on its prototype lowers on, so
  `TestSpreadPrototypeField` (internal/oracle/spread_prototype_test.go) goes red on c13b_a.a, setter_a.a,
  method_field_a.a, base_getter_a.a, and every other row it refuses, generic, union and signature; own_field_a.a and
  undeclared_getter_a.a, which the rule admits, stay green. `TestSpreadRecordsTheGettersItMayRun`
  (internal/lower/spread_getters_test.go) goes red too, on its class's and inherited rows, each wanting its .a
  twin refused.
- **The type parameter's arm answers false** (`spreadViewParts` judges a T as any other part, by the fields
  what it stands for declares): a class's methods and accessors aren't fields, so a spread of a T that
  stands for one claims none, and `TestSpreadPrototypeField` goes red on generic_method_call_a.a,
  generic_method_typeof_a.a and generic_setter_a.a, which lower and differ from Node, and on
  generic_getter_a.a, which stops at the read instead.
- **The union walk dropped** (`spreadViewParts`' walk takes a union whole, as its properties): a union's
  properties are only the names every member has, so `TestSpreadPrototypeField` goes red on union_setter_a.a,
  which lowers and stops on the compiler bug natively, and on union_getter_a.a, which stops at the read.
- **prototypeFree answers false** (library_spread.go): a class whose chain holds none of the copy's names on
  its prototype is no longer proof, so `TestSpreadPrototypeField` goes red on over_refusal_a.a, refused at the
  spread where it wants the read's stop.
- **The accessor literal's clause dropped** (`plainSpreadSources`' `graph.program.Classes[value.Class-1].Literal`):
  an object literal's synthetic class holds its getter as its own, so `TestSpreadPrototypeField` goes red on
  literal_with_class_a.a, refused at the spread where it wants the read's stop, and on literal_accessor_a.a,
  refused where it lowers and agrees.
- **The signature forms dropped** (`copyClaim` answers no claim for anything but a field): a view that
  declares a member as a method signature or an accessor signature claims nothing, so `TestSpreadPrototypeField`
  goes red on method_signature_a.a, which lowers and panics where Node throws, method_signature_label_a.a,
  which lowers and agrees where it wants the refusal, and interface_getter_a.a, which stops at the read.
- **The contextual type judged even when generic** (`spreadHeld` without its type parameter fallback): a
  return typed Partial<T> holds the copy as Partial of what T stands for, whose properties are the class's
  own methods, which claim nothing, so `TestSpreadPrototypeField` goes red on generic_partial_a.a, which
  lowers where it wants the refusal. A return typed T stays refused either way, since T's arm claims every
  member of what it stands for, and so does s05_a.a, whose callback's contextual return the checker has
  already resolved to the callback's own concrete return.
- **The satisfies target judged** (`spreadContextHeld` takes a `satisfies` as a slot): s01_call_a.a's copy is
  judged as Plain, though it keeps the literal's own type, which claims peek, so `TestSpreadPrototypeField`
  goes red on s01_call_a.a, which lowers and differs from Node where it wants the refusal. A function under
  `satisfies` is then held too, so it goes red on satisfies_return_setter_a.a as well, which lowers and ends on
  the compiler bug natively, and `TestSpreadContextHeldAnswersByThePlace` (internal/lower) goes red on its two
  satisfies rows, each answered held.
- **spreadReturnHeld answers held for any function** (the bare answer for a return or an arrow body, before
  Diagnostics' must on the re-read): satisfies_return_setter_a.a's literal is judged as Light, the context
  `satisfies` gives it, though make() is typed as its own type, Source, which claims weight, so
  `TestSpreadPrototypeField` goes red on satisfies_return_setter_a.a, which lowers and ends on the compiler bug
  natively, and `TestSpreadContextHeldAnswersByThePlace` goes red on its two satisfies rows.
- **alsoOwn always false** (an assignment's target taken as the one holder even where its value is read):
  `TestSpreadContextHeldAnswersByThePlace` goes red on its read assignment's row, answered (true, false). No
  oracle row reaches it today: Diagnostics' assign_value probes stop earlier, at stage-0 stops that aren't this
  rule's.
