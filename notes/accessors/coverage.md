# Accessor coverage at d785e87

Read CLAUDE.md, README.md, docs/0.1.md, docs/memory.md, the complete
origin/main...origin/codex/accessors diff, its commit message, and the changed files.
The only branch commit is d785e8755c3d4df174c9626e3e5edeb0e5e008ef,
"Lower instance accessors through virtual method calls". Main was 5d4c801.

Existing fixture names below are relative to internal/oracle/testdata. New runnable
names abbreviate accessors_coverage_<name>.a. Unsupported names are relative to
notes/accessors/unsupported. Differences are under notes/accessors/differences.
The oracle uses explicit registration, not directory discovery: accessors_test.go
appends each new runnable path to fixtures with lowers=true and checked=false.

## Accepted conditions, values and call sites

| Code case | Existing oracle program using accessors this way | New evidence for a gap |
|---|---|---|
| Ordinary method, getter, setter names get independent slots; signatures precede bodies | accessors.a, accessors_hierarchy.a | constructor calls a later-declared getter, which calls a later-declared method |
| Named, nonstatic, nongeneric class descriptors; getter-only, setter-only, complete pair | accessors.a, accessors_hierarchy.a | setter_only_override |
| Getter computes a number and mutates cached state | accessors.a | already covered |
| Getter returns a number field | accessors_hierarchy.a | operators uses total, including x.total += 1 |
| Getter returns an immortal string | accessors_hierarchy.a | already covered |
| Getter returns a new runtime string | accessors_ownership.a | string_snapshot exercises a stored runtime string held across a mutating operand |
| Getter returns a stored runtime string | accessors_ownership.a | string_snapshot |
| Getter returns a newly allocated structural object | accessors_ownership.a | already covered |
| Getter returns a stored structural object, and retains it elsewhere | accessors_analyses.a | optional_references, absent and present |
| Getter returns boolean | none | scalars, false and true |
| Getter returns number or undefined, both outcomes | none | scalars |
| Getter returns boolean or undefined, both outcomes without unsupported boolean field storage | none | scalars |
| Getter returns differently represented union members | none | scalars, number and string; narrow the captured result, not another getter call |
| Getter returns arrays, new and held in a field | none | collections, writes through the returned array and replaces the original field |
| Getter returns a tuple | none | collections |
| Getter returns a stored Map or Set | none | collections, replacements preserve previous aliases |
| Getter returns a closure, stored or newly allocated with a captured runtime string | none | functions; calls the accessor result directly and after receiver destruction |
| Getter returns a nominal class instance | none | objects, nominal_result |
| Getter returns an Error or RegExp | none | objects |
| Getter returns a Weak target, present and absent | none | weak, keeping a strong owner alive |
| Plain setter writes, including setter-only | accessors.a | assignment_order, setter_only_override |
| Setter receives string or object references | accessors_ownership.a for string | objects, collections, functions, weak for other reference representations |
| Setter fits scalar/undefined into optional parameter representation | none | scalars |
| Setter fits number/string into a union input independently of its string getter | none | union_setter |
| Statement numeric += and postfix ++ | accessors.a | operators also spells x.total += 1 |
| Statement string += | accessors_ownership.a | string_snapshot holds the getter's returned field while the operand overwrites it |
| -=, *=, /=, %=, **=, &=, |=, ^=, <<=, >>=, >>>= | none | operators |
| Prefix ++, prefix --, postfix -- | none | operators |
| Parenthesized update statement | none | operators |
| Updates in for-loop update position | none | operators, loop |
| Setter in for initializer; getter in condition | none | loop |
| Receiver is this inside an accessor/method | accessors_ownership.a | constructor, after every required field is initialized |
| Receiver is this in a derived constructor after super and initialization | none | constructor |
| Receiver is a variable, call result or freshly constructed object | accessors_order.a for variables/calls | operand_throw, objects, functions for fresh receivers |
| Receiver evaluated once before getter, operand and setter | accessors_order.a | string_snapshot, assignment_order |
| Getter or operand redirects a receiver variable | accessors_order.a | assignment_order also covers plain assignment, which must hold the receiver before its operand |
| Getter throw skips operand and setter, including virtual throw effects | accessors_throw.a | already covered |
| Setter throws in a complete pair | accessors.a | setter_throw adds a virtual override, temporary receiver, owned argument and finally |
| Setter-only descriptor throws | none | setter_throw |
| Right operand throws after getter, before setter | none | operand_throw, number and owned string results |
| Base-typed views and arrays dispatch to most-derived implementation | accessors_hierarchy.a | setter_only_override |
| Unchanged descriptor inherited into nongeneric subclass | accessors_hierarchy.a | already covered |
| Unchanged nongeneric descriptor inherited into generic subclass | none | generic_inherited, string and number instantiations |
| Getter-only override and complete-pair override with matching halves | accessors_hierarchy.a, accessors_ownership.a | setter_only_override covers only the setter half |
| Getter result covariance with same scalar representation; setter input contravariance | none | override_types, literal getter result and wider number setter input |
| Covariant readonly array contents | none | override_types, readonly Dog[] as readonly Animal[] |
| Nominal covariant getter result | none | nominal_result, Special as Item through a base getter |
| Mutable getter contents must remain invariant | none | unsupported/mutable_getter_result_override.a |
| Narrow class unions before selecting an accessor slot | none | union_narrowed, both classes |
| Stable narrowed optional scalar/reference reads | none | scalars, optional_references |
| Undefined after narrowing on a second getter call | none | repeated_number, repeated_boolean, repeated_string differ; both Adamic backends intentionally insert a check |
| Virtual borrow summary joins a changing getter and setter override | accessors_analyses.a | already covered |
| All virtual borrow targets are unchanged | none | borrow_safe, base and derived getter |
| Owned virtual returns can allocate and be kept, returned, or freed with receiver; heap convention and region exclusion | accessors_ownership.a | functions, collections, nominal_result, objects |
| Getter-bearing class mixed with a plain field object in inferred array | none | inferred_mixed_array differs: inferred nominal layout is not sound |

## Rejected conditions

No existing oracle fixture contains the following rejected accessor operations.
Many already have inline lower-package refusal tests; those do not constitute an
oracle program. The saved .a probes and observations.json record the actual CLI
build diagnostic and Node execution for each.

| Condition in the branch | Probe files |
|---|---|
| Object literal getter/setter | literal.a, object_setter.a |
| Descriptor outside class | interface_descriptor.a |
| Static getter/setter | static.a, static_setter.a |
| Declared generic getter/setter | generic.a, generic_setter.a |
| Computed or #private name | computed.a, private_name.a |
| Body absent, abstract or ambient | abstract.a, ambient.a |
| Super accessor | super.a |
| Optional accessor receiver | optional.a |
| Type parameter receiver | generic_function_receiver.a |
| Structural descriptor receiver | interface_getter_parameter.a |
| Writable/readonly interface conversion | writable_interface.a, readonly_interface.a |
| Interface inheriting a class descriptor, including unrelated structural implementation | interface_inherits_accessor.a, unrelated_getter_interface.a |
| Descriptor conversion in nested array, function signature, argument, field or cast | nested_writable_interface.a, function_view.a, interface_argument.a, interface_field.a, interface_cast.a |
| Descriptor erasure to interface or nominal base lacking descriptor | erased_accessor.a, erased_base.a |
| Assignment/compound/prefix/postfix result used as value | setter_value.a, compound_value.a, update_value.a; prefix uses the same value guard |
| Indexed source | indexed.a |
| Spread source | spread.a |
| Variable object binding | destructure.a |
| Parameter object binding | parameter_destructure.a |
| Missing getter, setter-only read | setter_only_read.a |
| Missing setter, getter-only write | getter_only_write.a, rejected by TypeScript before the defensive lowering guard |
| Un-narrowed class union | union.a |
| Narrowed getter changes native representation | getter_narrowing.a |
| Override member kind changes | field_replaces_accessor.a, accessor_replaces_field.a, method_replaces_accessor.a; TypeScript rejects these first |
| Override drops setter, drops getter, or adds half | partial_descriptor.a, removed_getter.a, added_half.a |
| Override mutable result covariance or narrower setter input | mutable_getter_result_override.a, setter_parameter_narrowing.a |
| Override native representation differs | override_representation.a |
| Constructor getter before required fields are initialized | constructor_escape.a |
| Base constructor getter before derived initialization | base_constructor_escape.a |
| Getter in field initializer | initializer_escape.a |
| Scalar getter result narrowed to undefined after assigning undefined | narrowed_to_undefined.a; typeOf(target) cannot represent the narrowed type |
| Other current compiler limits | discarded.a, null_result.a, undefined_result.a, never_result.a, void_result.a, optional_boolean_field.a, captured_this_result.a |

## Guards with no independently executable source

An accessor without a virtual implementation cannot be expressed by a well-typed,
concrete admitted class: instantiate installs every declared/inherited half before
lowering bodies. Unknown getter/setter representations are rejected at signature
lowering; a setter argument still mismatching representation after fit is a
defensive guard, since the checker already relates its input and fit adapts every
admitted representation. TypeScript intercepts missing setters and member-kind
changes before those lower-package guards. These branches cannot be reached by a
valid oracle program without changing the compiler or constructing IR by hand.
