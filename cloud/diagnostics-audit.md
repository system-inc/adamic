# Lower diagnostic audit

Baseline: `ef3d907ecdc4c771b016f7d9c52372def057a340` (`origin/main`). This is a first-pass wording audit, not a change to the language or diagnostic transport.

## Scope and counting

AST inventory of every production `Refused{}`, `NotYet{}`, and `l.notYet(...)` in `internal/lower/*.go`; `_test.go` fixtures are excluded (the same constructor/call search finds no sites in them). There are **543 sites: 61 Refused constructors, 8 NotYet constructors and 474 notYet calls**. The helper constructor and its callers are counted separately as requested, so these are source-site counts, not unique runtime diagnostics.

A location means a source path is emitted, even when no line/column exists. All sites name a path; the seven direct constructors in `class_accessors.go` give only `l.result.Source`. A construct means the message identifies the failing syntax, operation or type, including expressions that interpolate its name. A fix means an actionable edit or design alternative, rather than a reason, a support roadmap or a documentation link alone. A rule means a stable semantic identifier; existing lint identifiers count alongside `adamic/...`. Calls through `flow.l.notYet` are included because they invoke the same helper. Shared producers receive yes only when every route qualifies. The generic NotYet helper therefore has no guaranteed fix.

## Counts

| Kind | Sites | Location yes/no | Construct yes/no | Fix before yes/no | Fix after yes/no | Rule before yes/no | Rule after yes/no |
|---|---:|---|---|---|---|---|
| Refused | 61 | 61/0 | 61/0 | 57/4 | 57/4 | 19/42 | 61/0 |
| NotYet | 8 | 8/0 | 8/0 | 0/8 | 0/8 | 0/8 | 0/8 |
| notYet | 474 | 474/0 | 474/0 | 17/457 | 90/384 | 0/474 | 0/474 |
| Total | 543 | 543/0 | 543/0 | 74/469 | 147/396 | 19/524 | 61/482 |

## Every site and its baseline message

Source coordinates below refer to the baseline. `notYet` locations are the node expressions passed to `l.program.Where`; constructor locations are their `Where` expressions. Message cells preserve Go expressions and interpolation rather than inventing concrete source values. Render templates: `Where + ": stage 0 can't lower " + What + " yet"`; `Where + ": Adamic 0.1 refuses " + What + "; " + Fix`. Dynamic producers are expanded in the next section. Y/N columns describe baseline wording. Delta records this branch’s change.

| Site | Kind | Location expression | What; Fix (Refused only) | Location | Construct | Fix | Rule | Delta |
|---|---|---|---|---|---|---|---|---|
| internal/lower/assignments.go:16 | notYet | node | describe(node) + " as a statement" | Y | Y | N | N |  |
| internal/lower/assignments.go:38 | notYet | target | "assigning to " + describe(target) | Y | Y | N | N |  |
| internal/lower/assignments.go:41 | notYet | target | "assigning to what a catch caught" | Y | Y | N | N |  |
| internal/lower/assignments.go:45 | notYet | target | "assigning to a parameter that only ever receives undefined" | Y | Y | N | N |  |
| internal/lower/assignments.go:68 | notYet | where | "a name past its tuple's elements" | Y | Y | N | N |  |
| internal/lower/assignments.go:72 | notYet | where | "a tuple element of type " + l.checker.TypeToString(elements[index]) | Y | Y | N | N |  |
| internal/lower/assignments.go:76 | notYet | where | "a tuple element of type " + l.checker.TypeToString(elements[index]) + " given to a " + typeName(to) | Y | Y | N | N |  |
| internal/lower/assignments.go:86 | notYet | pattern | "destructuring other than a tuple" | Y | Y | N | N |  |
| internal/lower/assignments.go:93 | notYet | pattern | "destructuring a " + typeName(value.Type()) | Y | Y | N | N |  |
| internal/lower/assignments.go:105 | notYet | element | "assigning to " + describe(element) + " in a destructuring assignment" | Y | Y | N | N |  |
| internal/lower/assignments.go:144 | notYet | node | describe(node) + " as a statement" | Y | Y | N | N |  |
| internal/lower/assignments.go:156 | notYet | operand | "incrementing " + describe(operand) | Y | Y | N | N |  |
| internal/lower/cast.go:32 | Refused | l.program.Where(node) | "a cast the runtime can't check" ; "narrow it instead (===, typeof, a discriminant), or cast a discriminated union to its members (adamic/no-unchecked-cast)" | Y | Y | Y | Y |  |
| internal/lower/class.go:72 | notYet | where | "a method receiver without a nominal class view" | Y | Y | N | N |  |
| internal/lower/class.go:82 | notYet | where | "a class whose type arguments aren't known here" | Y | Y | N | N |  |
| internal/lower/class.go:87 | notYet | where | "a class instantiated with " + l.checker.TypeToString(typeArguments[index]) | Y | Y | N | N |  |
| internal/lower/class.go:123 | Refused | l.program.Where(where) | "a generic class instantiated without end (polymorphic recursion)" ; "keep recursive type arguments unchanged, or write a class per type" | Y | Y | Y | N | rule: adamic/polymorphic-recursion |
| internal/lower/class.go:167 | notYet | member | "a method with a computed name" | Y | Y | N | N | alternative: give the method a fixed identifier name |
| internal/lower/class.go:191 | notYet | member | describe(member) + " in a class" | Y | Y | N | N |  |
| internal/lower/class.go:295 | Refused | l.program.Where(node) | "new of an abstract class" ; "construct a concrete subclass that implements its abstract methods" | Y | Y | Y | N | rule: adamic/concrete-construction |
| internal/lower/class.go:335 | notYet | node | "a method call through a structural signature in a program with statics; use typeof the declaring class" | Y | Y | Y | N |  |
| internal/lower/class.go:343 | notYet | node | "a method of a class stage 0 doesn't have" | Y | Y | N | N |  |
| internal/lower/class.go:346 | notYet | receiver | "a method call through a union of class types; narrow with instanceof first" | Y | Y | Y | N |  |
| internal/lower/class.go:376 | notYet | node | "super outside a derived class" | Y | Y | N | N |  |
| internal/lower/class.go:419 | notYet | target | "replacing a represented method at runtime" | Y | Y | N | N | alternative: store a replaceable arrow function in a declared callback field instead |
| internal/lower/class.go:426 | notYet | target | "assigning a field of a " + typeName(object.Type()) | Y | Y | N | N |  |
| internal/lower/class.go:438 | notYet | target | "storing " + l.checker.TypeToString(l.checker.GetTypeAtLocation(target)) + " in a field" | Y | Y | N | N |  |
| internal/lower/class.go:459 | notYet | target | "assigning a field of a " + typeName(object.Type()) | Y | Y | N | N |  |
| internal/lower/class.go:573 | Refused | l.program.Where(node) | "this before super returns" ; "call super(...) before using this" | Y | Y | Y | N | rule: adamic/this-after-super |
| internal/lower/class.go:581 | Refused | l.program.Where(node) | "this escaping a base constructor before derived fields are initialized" ; "use this only to read or write initialized base fields; call methods and publish the object after construction" | Y | Y | Y | N | rule: adamic/initialized-this |
| internal/lower/class.go:591 | Refused | l.program.Where(node) | "this escaping a constructor before every field is set (stored, passed, or a method called on it, which could read a field that holds undefined while its type says otherwise)" ; "assign every field first, then use this" | Y | Y | Y | N | rule: adamic/initialized-this |
| internal/lower/class_accessors.go:59 | notYet | target | "super outside a derived accessor or method" | Y | Y | N | N |  |
| internal/lower/class_accessors.go:92 | notYet | target | "super of a member without the requested accessor" | Y | Y | N | N |  |
| internal/lower/class_accessors.go:147 | notYet | property | "an accessor literal with a spread" | Y | Y | N | N |  |
| internal/lower/class_accessors.go:151 | notYet | property | "a computed accessor literal name" | Y | Y | N | N |  |
| internal/lower/class_accessors.go:154 | notYet | name | "an accessor literal property name reserved for closure storage; rename the property" | Y | Y | Y | N |  |
| internal/lower/class_accessors.go:170 | notYet | property | "an accessor without a native property representation" | Y | Y | N | N |  |
| internal/lower/class_accessors.go:183 | notYet | property | "a non-accessor method in an accessor literal" | Y | Y | N | N |  |
| internal/lower/class_accessors.go:192 | notYet | property | "an accessor literal's field without a native slot" | Y | Y | N | N |  |
| internal/lower/class_accessors.go:242 | NotYet | l.result.Source | "a property name shared with a descriptor missing the requested getter or setter" | Y | Y | N | N |  |
| internal/lower/class_accessors.go:249 | NotYet | l.result.Source | "reading a setter-only property or writing a getter-only property" | Y | Y | N | N |  |
| internal/lower/class_accessors.go:258 | NotYet | l.result.Source | "accessors sharing a name with different native representations" | Y | Y | N | N |  |
| internal/lower/class_accessors.go:305 | NotYet | l.result.Source | "an optional accessor read" | Y | Y | N | N |  |
| internal/lower/class_accessors.go:397 | NotYet | l.result.Source | "spreading in a program with static constructor objects" | Y | Y | N | N |  |
| internal/lower/class_accessors.go:400 | NotYet | l.result.Source | "spreading a setter-only property, whose read value is undefined" | Y | Y | N | N |  |
| internal/lower/class_accessors.go:403 | NotYet | l.result.Source | "spreading an accessor literal whose getter may throw" | Y | Y | N | N |  |
| internal/lower/class_accessors.go:425 | notYet | member | "a getter without a native adamic_value representation" | Y | Y | N | N |  |
| internal/lower/class_accessors.go:428 | notYet | member | "a setter whose input needs two native words" | Y | Y | N | N |  |
| internal/lower/class_features.go:37 | notYet | node | "Object.keys without exactly one argument" | Y | Y | N | N |  |
| internal/lower/class_features.go:43 | notYet | node | "Object.keys on a literal with symbol-key storage" | Y | Y | N | N |  |
| internal/lower/class_features.go:60 | notYet | node | "Object.keys on a non-object" | Y | Y | N | N |  |
| internal/lower/class_generic_calls.go:33 | notYet | node | "a generic call whose concrete signature isn't known" | Y | Y | N | N |  |
| internal/lower/class_generic_calls.go:39 | notYet | node | "a generic call whose checker type mapper isn't exposed" | Y | Y | N | N |  |
| internal/lower/class_generic_calls.go:114 | Refused | l.program.Where(where) | "a generic argument without nominal ancestry seen as " + l.checker.TypeToString(mismatch) ; "use that class or a subclass as the argument; use an interface for structural constraints (adamic/nominal-class)" | Y | Y | Y | Y |  |
| internal/lower/class_inheritance.go:21 | notYet | clause | "a computed class base; name the base class directly" | Y | Y | Y | N |  |
| internal/lower/class_inheritance.go:25 | notYet | clause | "a class base whose type isn't known" | Y | Y | N | N | alternative: name a class declared in this program as the base |
| internal/lower/class_inheritance.go:30 | notYet | clause | "a base that isn't a declared class" | Y | Y | N | N | alternative: extend a declared class; use implements for an interface |
| internal/lower/class_inheritance.go:63 | Refused | l.program.Where(member) | what ; fix | Y | Y | Y | N | all shared routes now carry IDs |
| internal/lower/class_inheritance.go:120 | notYet | member | "an inherited field override with a different native representation" | Y | Y | N | N | alternative: keep the inherited field type unchanged and narrow a local after reading it |
| internal/lower/class_inheritance.go:129 | notYet | member | "an overloaded override" | Y | Y | N | N | alternative: use one override signature with union parameters and narrow them inside the method |
| internal/lower/class_inheritance.go:133 | notYet | member | "an override with a different parameter count" | Y | Y | N | N | alternative: keep the base method parameter list and count |
| internal/lower/class_inheritance.go:144 | notYet | member | "an override with a different native parameter representation" | Y | Y | N | N | alternative: keep the base parameter types and narrow them inside the method |
| internal/lower/class_inheritance.go:155 | notYet | member | "an override with a different native result representation" | Y | Y | N | N | alternative: keep the base result type in the override signature |
| internal/lower/class_inheritance.go:203 | notYet | member | "a declare or abstract class field" | Y | Y | N | N |  |
| internal/lower/class_inheritance.go:209 | notYet | member | "a field with a computed name" | Y | Y | N | N | alternative: declare the field with a fixed identifier name |
| internal/lower/class_inheritance.go:216 | notYet | member | "a field of type " + l.checker.TypeToString(l.checker.GetTypeAtLocation(member.Name())) | Y | Y | N | N |  |
| internal/lower/class_inheritance.go:260 | notYet | returnValue | "a constructor returning a replacement value" | Y | Y | N | N | alternative: use a module-level factory function when construction must return a different object |
| internal/lower/class_inheritance.go:371 | notYet | node | "super outside a derived constructor" | Y | Y | N | N |  |
| internal/lower/class_inheritance.go:411 | notYet | node | "instanceof against a value that isn't a declared class" | Y | Y | N | N | alternative: test against a declared class name, or use an explicit discriminant |
| internal/lower/class_inheritance.go:502 | Refused | l.program.Where(node) | "a value without nominal ancestry seen as " + l.checker.TypeToString(target) ; "construct that class or a subclass; use an interface for structural values (adamic/nominal-class)" | Y | Y | Y | Y |  |
| internal/lower/class_inheritance.go:549 | Refused | l.program.Where(node) | "this in a field initializer before the fields it reads are initialized" ; "declare the field it reads earlier, or initialize it in the constructor after super and all required fields are set" | Y | Y | Y | N | rule: adamic/initialized-this |
| internal/lower/class_inheritance.go:729 | Refused | l.program.Where(binding) | "a destructured value without nominal ancestry seen as " + l.checker.TypeToString(mismatch) ; "construct that class or a subclass; use an interface for structural values (adamic/nominal-class)" | Y | Y | Y | Y |  |
| internal/lower/class_static.go:109 | notYet | member | "a static member without an identifier name" | Y | Y | N | N | alternative: give the static member a fixed identifier name |
| internal/lower/class_static.go:113 | notYet | member | "a declare or abstract static field" | Y | Y | N | N |  |
| internal/lower/class_static.go:120 | notYet | member | "a static field without a native slot" | Y | Y | N | N |  |
| internal/lower/class_static.go:123 | notYet | member | "an uninitialized nonnullable static field; initialize it at its declaration" | Y | Y | Y | N |  |
| internal/lower/class_static.go:135 | notYet | member | "a static field override with a different native representation" | Y | Y | N | N | alternative: keep the inherited static field type unchanged |
| internal/lower/class_static.go:345 | Refused | l.program.Where(child) | "a static field read before its initializer has run" ; "declare the field earlier, or move the read after that field's initialization" | Y | Y | Y | N | rule: adamic/static-initialization |
| internal/lower/class_static.go:366 | Refused | l.program.Where(child) | "an outside helper reading a class during its declaration (the temporal dead zone)" ; "run the helper after the class declaration, or read this inside a static method" | Y | Y | Y | N | rule: adamic/class-temporal-dead-zone |
| internal/lower/class_static.go:371 | Refused | l.program.Where(child) | "a constructor object escaping before static fields are initialized" ; "publish the class value after the last static field initializer" | Y | Y | Y | N | rule: adamic/static-initialization |
| internal/lower/class_static.go:378 | Refused | l.program.Where(child) | "a constructor object escaping before static fields are initialized" ; "publish this after the last static field initializer" | Y | Y | Y | N | rule: adamic/static-initialization |
| internal/lower/class_static.go:418 | notYet | argument | "an indirect callback before static fields are initialized; move it after the field initializers" | Y | Y | Y | N |  |
| internal/lower/class_static.go:426 | notYet | child | "an indirect call before static fields are initialized; move it after the field initializers" | Y | Y | Y | N |  |
| internal/lower/class_static.go:452 | notYet | node | "an implicit constructor property such as prototype, name or length" | Y | Y | N | N |  |
| internal/lower/class_static.go:457 | notYet | declaration | "a declare class without a runtime definition" | Y | Y | N | N |  |
| internal/lower/class_static.go:461 | notYet | clause | "a computed class base; name the base class directly" | Y | Y | Y | N |  |
| internal/lower/class_super.go:33 | notYet | child | "super used as a value or in a deferred function; call it as a statement in each branch" | Y | Y | Y | N |  |
| internal/lower/class_super.go:109 | notYet | node | "an explicit return from a derived constructor" | Y | Y | N | N |  |
| internal/lower/collections.go:81 | notYet | node | "iterating a " + typeName(value.Type()) | Y | Y | N | N |  |
| internal/lower/collections.go:128 | notYet | source | "new Map from something that isn't [key, value] pairs" | Y | Y | N | N |  |
| internal/lower/collections.go:131 | notYet | source | "new Map from pairs held otherwise than the Map's keys and values" | Y | Y | N | N |  |
| internal/lower/collections.go:138 | notYet | source | "new Map from something that isn't [key, value] pairs" | Y | Y | N | N |  |
| internal/lower/collections.go:206 | notYet | node | "clear with arguments" | Y | Y | N | N |  |
| internal/lower/collections.go:214 | notYet | node | "forEach with other than one callback" | Y | Y | N | N |  |
| internal/lower/collections.go:221 | notYet | arguments[0] | "forEach with a callback that isn't a function" | Y | Y | N | N |  |
| internal/lower/collections.go:225 | notYet | arguments[0] | "forEach with an overloaded callback" | Y | Y | N | N |  |
| internal/lower/collections.go:231 | notYet | arguments[0] | "forEach with a callback returning " + l.checker.TypeToString(result) | Y | Y | N | N |  |
| internal/lower/collections.go:244 | notYet | node | "the length of a tuple that isn't a plain local" | Y | Y | N | N | alternative: bind the tuple to a local const before reading its length |
| internal/lower/collections.go:247 | notYet | node | "the length of a tuple with optional or rest elements" | Y | Y | N | N |  |
| internal/lower/collections.go:257 | notYet | pattern | "a destructuring declaration without a value" | Y | Y | N | N | alternative: provide an initializer: const [first, second] = pair |
| internal/lower/collections.go:284 | notYet | pattern | "destructuring anything but a tuple into [names]" | Y | Y | N | N | alternative: bind the array once, then read its indexes explicitly and handle undefined |
| internal/lower/collections.go:287 | notYet | pattern | "destructuring a " + typeName(heldAs) | Y | Y | N | N |  |
| internal/lower/collections.go:301 | notYet | binding | "a destructured name that isn't plain" | Y | Y | N | N | alternative: bind a plain name first, then destructure or apply a default in a separate statement |
| internal/lower/collections.go:308 | notYet | binding | "destructuring past a tuple's end" | Y | Y | N | N |  |
| internal/lower/collections.go:316 | notYet | declared.PropertyName | "a computed field name" | Y | Y | N | N | alternative: spell the declared field name as an identifier or string literal |
| internal/lower/collections.go:322 | notYet | binding | "destructuring a field the type doesn't name" | Y | Y | N | N |  |
| internal/lower/collections.go:328 | Refused | l.program.Where(binding) | "a method in object destructuring" ; "call it on its receiver or wrap that call in an arrow; destructuring would lose this" | Y | Y | Y | N | rule: unbound-method |
| internal/lower/collections.go:333 | notYet | binding | "an inherited library member in object destructuring, which is not an own field" | Y | Y | N | N | alternative: read the inherited member explicitly from its receiver instead of destructuring it |
| internal/lower/collections.go:349 | notYet | binding | "a destructured name held otherwise than its field" | Y | Y | N | N |  |
| internal/lower/collections.go:364 | notYet | node | "?.[] on anything but a tuple, at a position written out" | Y | Y | N | N |  |
| internal/lower/collections.go:369 | notYet | node | "?.[] past a tuple's end" | Y | Y | N | N |  |
| internal/lower/collections.go:373 | notYet | node | "?.[] to a tuple element of type " + l.checker.TypeToString(elements[position]) | Y | Y | N | N |  |
| internal/lower/control.go:103 | Refused | l.program.Where(node) | "a " + typeName(condition.Type()) + " as a condition" ; "compare it explicitly, like name.length &gt; 0 or count !== 0" | Y | Y | Y | N | rule: strict-boolean-expressions |
| internal/lower/cycles.go:113 | Refused | l.program.Where(node) | "'" + declared.Name + "', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free" ; "write the function as a function declaration (function " + declared.Name + "() {}), which captures nothing, or declare the variable Weak&lt;...&gt; and keep the function somewhere strong (adamic/cycle-capable)" | Y | Y | Y | Y |  |
| internal/lower/cycles.go:277 | Refused | l.program.Where(f.where[holder]) | name + ", an array whose elements can reach back to an array like it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")" ; "declare the elements weak, Weak&lt;" + target + "&gt;[] (import type { Weak } from 'adamic'), which don't count and read undefined once what they point to is freed; or make it readonly " + target + "[]; or write into such an array only values this function made, or only into one it made (adamic/cycle-capable)" | Y | Y | Y | Y |  |
| internal/lower/cycles.go:288 | Refused | l.program.Where(f.where[holder]) | name + ", a set whose elements can reach back to a set like it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")" ; "make it a ReadonlySet, built whole when it's made; or add to such a set only values this function made, or only to one it made (adamic/cycle-capable)" | Y | Y | Y | Y |  |
| internal/lower/cycles.go:299 | Refused | l.program.Where(f.where[holder]) | name + ", a map whose keys can reach back to a map like it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")" ; "make it a ReadonlyMap, built whole when it's made; or set in such a map only keys this function made, or only in one it made (adamic/cycle-capable)" | Y | Y | Y | Y |  |
| internal/lower/cycles.go:308 | Refused | l.program.Where(f.where[holder]) | name + ", a map whose values can reach back to a map like it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")" ; "declare the values weak, Map&lt;" + l.checker.TypeToString(arguments[0]) + ", Weak&lt;" + f.present(arguments[1]) + "&gt;&gt; (import type { Weak } from 'adamic'), or make it a ReadonlyMap; or set in such a map only values this function made, or only in one it made (adamic/cycle-capable)" | Y | Y | Y | Y |  |
| internal/lower/cycles.go:339 | Refused | l.program.Where(where) | name + "." + l.cycleFieldName(field) + ", " + kind + " of type " + target + ", which can reach back to the " + name + " holding it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")" ; "declare it " + l.cycleFieldName(field) + ": Weak&lt;" + f.present(proven) + "&gt; (import type { Weak } from 'adamic'), which doesn't count and reads undefined once what it points to is freed" + orReadonly + "; or write into it only values this function made, or only into what it made (adamic/cycle-capable)" | Y | Y | Y | Y |  |
| internal/lower/diagnostics.go:33 | NotYet | l.program.Where(node) | what | Y | Y | N | N |  |
| internal/lower/exceptions.go:29 | notYet | thrown | "throwing an Error that isn't made where it's thrown or caught by the catch around it" | Y | Y | N | N |  |
| internal/lower/exceptions.go:31 | Refused | l.program.Where(thrown) | "throwing a " + l.checker.TypeToString(l.checker.GetTypeAtLocation(thrown)) ; "throw an Error: throw new Error(String(value)); what a catch takes is unknown, and an Error is what it can be sure of" | Y | Y | Y | N | rule: adamic/throw-error |
| internal/lower/exceptions.go:46 | notYet | node | "new Error with options" | Y | Y | N | N |  |
| internal/lower/exceptions.go:53 | notYet | node | "new Error with a message that isn't a string" | Y | Y | N | N | alternative: convert the message explicitly, for example new Error(String(value)) for a scalar value |
| internal/lower/exceptions.go:73 | notYet | name | "a catch that destructures what it caught" | Y | Y | N | N | alternative: catch one name, narrow it with instanceof Error, then read its fields |
| internal/lower/exceptions.go:141 | notYet | record.node | "a try around " + failing + ", whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md)" | Y | Y | N | N |  |
| internal/lower/expression.go:19 | notYet | node | "a value of type " + l.checker.TypeToString(l.checker.GetTypeAtLocation(node)) | Y | Y | N | N |  |
| internal/lower/expression.go:161 | notYet | node | "a " + l.checker.TypeToString(own) + " seen as a " + l.checker.TypeToString(contextual) + " (one keeps something weakly that the other keeps strongly)" | Y | Y | N | N |  |
| internal/lower/expression.go:163 | notYet | node | "a " + l.checker.TypeToString(tuple) + " seen as a " + l.checker.TypeToString(array) + " (a tuple is held as an object, not an array, so far; write it as an array where it's made, or copy it into one: [pair[0], pair[1]])" | Y | Y | Y | N |  |
| internal/lower/expression.go:171 | notYet | skipped | "a tuple where an array goes (as " + l.checker.TypeToString(contextual) + ")" | Y | Y | N | N |  |
| internal/lower/expression.go:186 | notYet | node | "a Weak of " + l.checker.TypeToString(target) | Y | Y | N | N |  |
| internal/lower/expression.go:192 | notYet | node | "a Weak of " + l.checker.TypeToString(read) | Y | Y | N | N |  |
| internal/lower/expression.go:399 | notYet | node | "a generic function as a value" | Y | Y | N | N |  |
| internal/lower/expression.go:402 | notYet | node | "reading String as a first-class constructor (its any-typed call signature, construction and static members need an intrinsic value representation)" | Y | Y | N | N |  |
| internal/lower/expression.go:405 | notYet | node | "reading " + node.Text() | Y | Y | N | N |  |
| internal/lower/expression.go:476 | notYet | node | "this outside a method" | Y | Y | N | N |  |
| internal/lower/expression.go:496 | notYet | node | "a void call used as a value" | Y | Y | N | N |  |
| internal/lower/expression.go:507 | notYet | node | "a void call used as a value" | Y | Y | N | N |  |
| internal/lower/expression.go:511 | notYet | node | describe(node) | Y | Y | N | N |  |
| internal/lower/expression.go:580 | notYet | node | describe(node) + " on a " + typeName(operand.Type()) | Y | Y | N | N |  |
| internal/lower/expression.go:634 | notYet | node | "null comparison with a scalar" | Y | Y | N | N |  |
| internal/lower/expression.go:660 | notYet | node | "comparing a " + typeName(value.Type()) + " with undefined" | Y | Y | N | N |  |
| internal/lower/expression.go:701 | notYet | node | describe(node) + " with a " + typeName(left.Type()) + " and a " + typeName(right.Type()) | Y | Y | N | N |  |
| internal/lower/expression.go:734 | notYet | span | "a template interpolating a union with an object, an array, a map or a function in it" | Y | Y | N | N |  |
| internal/lower/expression.go:742 | notYet | span | "a template interpolating an object, an array, a map, a function or undefined" | Y | Y | N | N |  |
| internal/lower/expression.go:785 | notYet | node | "a conditional whose branches have different types" | Y | Y | N | N |  |
| internal/lower/expression.go:852 | notYet | node | "a call to " + describe(callee) | Y | Y | N | N |  |
| internal/lower/expression.go:902 | notYet | node | "?? whose sides have different types" | Y | Y | N | N |  |
| internal/lower/expression.go:937 | notYet | node | "a function with an optional or rest parameter, as a value" | Y | Y | N | N |  |
| internal/lower/expression.go:942 | notYet | node | "a function value returning " + typeName(callee.Returns) | Y | Y | N | N |  |
| internal/lower/expression.go:950 | notYet | node | "a function value taking " + typeName(declared.Type) | Y | Y | N | N |  |
| internal/lower/expression.go:992 | notYet | call | "a call through ?. (an optional call)" | Y | Y | N | N |  |
| internal/lower/expression.go:1019 | notYet | node | "a call returning " + l.checker.TypeToString(result) | Y | Y | N | N |  |
| internal/lower/expression.go:1035 | notYet | node | "passing " + typeName(argument.Type()) + " to a function value" | Y | Y | N | N |  |
| internal/lower/expression.go:1039 | notYet | node | "a function value returning " + typeName(returns) | Y | Y | N | N |  |
| internal/lower/from.go:22 | notYet | node | "Array.from with thisArg" | Y | Y | N | N | alternative: capture the receiver in an arrow callback instead of passing thisArg |
| internal/lower/from.go:37 | notYet | node | "Array.from with other than { length } and a callback" | Y | Y | N | N | alternative: write Array.from({ length: count }, (_, index) =&gt; value) |
| internal/lower/from.go:41 | notYet | arguments[0] | "Array.from of anything but { length }" | Y | Y | N | N | alternative: copy an array with [...items], or generate values with Array.from({ length: count }, (_, index) =&gt; value) |
| internal/lower/from.go:45 | notYet | arguments[0] | "Array.from of anything but { length }" | Y | Y | N | N | alternative: write a source literal with only a numeric length field: { length: count } |
| internal/lower/from.go:61 | notYet | property | "Array.from with a length that isn't a number" | Y | Y | N | N | alternative: give length a number value |
| internal/lower/from.go:67 | notYet | arguments[1] | "Array.from with a callback that isn't an arrow function written in place" | Y | Y | N | N | alternative: write an inline arrow, (_, index) =&gt; callback(undefined, index) |
| internal/lower/from.go:80 | Refused | l.program.Where(parameters[0]) | "a first Array.from parameter typed " + l.checker.TypeToString(received) + ", which is undefined every time" ; "name it _ and leave it untyped, (_, index) =&gt; ..., or type it undefined (the type would be a lie the checker can't see)" | Y | Y | Y | N | rule: adamic/array-from-undefined |
| internal/lower/functions.go:78 | notYet | where | "a function returning " + l.checker.TypeToString(returns) | Y | Y | N | N |  |
| internal/lower/functions.go:102 | notYet | parameter | "a parameter that isn't a plain name" | Y | Y | N | N |  |
| internal/lower/functions.go:110 | notYet | parameter | "a function value with an optional parameter" | Y | Y | N | N |  |
| internal/lower/functions.go:114 | notYet | parameter | "a function value taking " + l.checker.TypeToString(l.checker.GetTypeAtLocation(parameter.Name())) | Y | Y | N | N |  |
| internal/lower/functions.go:122 | notYet | parameter | "a default for a " + typeName(missing) + " parameter" | Y | Y | N | N |  |
| internal/lower/functions.go:132 | notYet | declaration | "a function value returning " + typeName(function.Returns) | Y | Y | N | N |  |
| internal/lower/functions.go:135 | notYet | declaration | "a function without a body" | Y | Y | N | N |  |
| internal/lower/functions.go:169 | notYet | declaration | "a destructured parameter beside a parameter with a default" | Y | Y | N | N |  |
| internal/lower/functions.go:191 | notYet | parameter.initializer | "a default of another type than its parameter" | Y | Y | N | N |  |
| internal/lower/generic.go:28 | notYet | call | "a call to a generic function whose signature the checker didn't resolve" | Y | Y | N | N |  |
| internal/lower/generic.go:61 | Refused | l.program.Where(argument) | "a type argument makes a value of type " + l.checker.TypeToString(from) + " seen as " + l.checker.TypeToString(to) + ", which can write " + l.checker.TypeToString(found.target) + " where " + l.checker.TypeToString(found.source) + " is read" ; "use the value's invariant type argument, or make the parameter readonly (adamic/invariant-mutable)" | Y | Y | Y | Y |  |
| internal/lower/generic.go:93 | Refused | l.program.Where(call) | "a generic function instantiated without end (polymorphic recursion)" ; "call it with the same type arguments it was called with, or write a function per type" | Y | Y | Y | N | rule: adamic/polymorphic-recursion |
| internal/lower/generic.go:198 | Refused | l.program.Where(argument) | "instantiating a generic function makes a value of type " + l.checker.TypeToString(from) + " written where " + l.checker.TypeToString(to) + " is read" ; "make the collection readonly, or use a type parameter for the value being written (adamic/invariant-mutable)" | Y | Y | Y | Y |  |
| internal/lower/invariance.go:462 | Refused | l.program.Where(node) | what ; fix | Y | Y | Y | Y |  |
| internal/lower/invariance.go:600 | Refused | l.program.Where(node) | "a value of type " + l.checker.TypeToString(own) + " seen as " + l.checker.TypeToString(contextual) + ", which can write " + l.checker.TypeToString(found.target) + " where " + l.checker.TypeToString(found.source) + " is read" ; "make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)" | Y | Y | Y | Y |  |
| internal/lower/invariance.go:611 | Refused | l.program.Where(target) | "a value of type " + l.checker.TypeToString(elements[index]) + " seen as " + l.checker.TypeToString(to) + ", which can write " + l.checker.TypeToString(found.target) + " where " + l.checker.TypeToString(found.source) + " is read" ; "make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)" | Y | Y | Y | Y |  |
| internal/lower/iteration.go:72 | notYet | where | "an optional iterator method (runtime method presence is not represented)" | Y | Y | N | N |  |
| internal/lower/iteration.go:83 | notYet | where | "a declared iterator method with no runtime field" | Y | Y | N | N |  |
| internal/lower/iteration.go:86 | notYet | where | "an iterator method whose concrete origin is erased by a structural signature" | Y | Y | N | N |  |
| internal/lower/iteration.go:89 | notYet | where | "iterator methods with mixed receiver conventions" | Y | Y | N | N |  |
| internal/lower/iteration.go:95 | notYet | where | "an iterator method without a concrete declaration" | Y | Y | N | N |  |
| internal/lower/iteration.go:109 | notYet | where | "a class iterator method through a structural object view" | Y | Y | N | N |  |
| internal/lower/iteration.go:113 | notYet | where | "an iterator class not declared in this program" | Y | Y | N | N |  |
| internal/lower/iteration.go:147 | notYet | where | "an iterator protocol method with arguments or overloads" | Y | Y | N | N |  |
| internal/lower/iteration.go:151 | notYet | where | "an iterator protocol method that does not return a represented object" | Y | Y | N | N |  |
| internal/lower/iteration.go:169 | notYet | where | "a custom iterable without a present object representation" | Y | Y | N | N |  |
| internal/lower/iteration.go:188 | notYet | where | "an iterator result without a required boolean done field" | Y | Y | N | N |  |
| internal/lower/iteration.go:191 | notYet | where | "an iterator result whose done is not a boolean" | Y | Y | N | N |  |
| internal/lower/iteration.go:195 | notYet | where | "an iterator result without a represented value field" | Y | Y | N | N |  |
| internal/lower/iteration.go:199 | notYet | where | "an iterator result whose value has no single-slot representation" | Y | Y | N | N |  |
| internal/lower/iteration.go:219 | notYet | where | "a generic iterable view without proven native type arguments" | Y | Y | N | N |  |
| internal/lower/iteration.go:222 | notYet | where | "a generic iterator view without proven native type arguments" | Y | Y | N | N |  |
| internal/lower/iteration.go:238 | notYet | where | "an iterator result with declared fields absent at runtime" | Y | Y | N | N |  |
| internal/lower/iteration.go:248 | notYet | where | "an iterable view that erases its method receiver convention" | Y | Y | N | N |  |
| internal/lower/iteration.go:255 | notYet | where | "an iterator view that erases its method receiver convention" | Y | Y | N | N |  |
| internal/lower/iteration.go:260 | notYet | where | "an iterator view that erases its return receiver convention" | Y | Y | N | N |  |
| internal/lower/iteration.go:267 | notYet | where | "an iterator built by object spread (own method presence is not proved)" | Y | Y | N | N |  |
| internal/lower/iteration.go:273 | notYet | where | "an iterator view that can hide a return method" | Y | Y | N | N |  |
| internal/lower/iteration.go:284 | notYet | where | "replacing an iterator protocol method at runtime" | Y | Y | N | N |  |
| internal/lower/iteration.go:382 | notYet | name | "an iteration variable with a different value representation" | Y | Y | N | N |  |
| internal/lower/iteration.go:423 | notYet | node | "an explicit Symbol.iterator call on a built-in value" | Y | Y | N | N |  |
| internal/lower/iteration.go:448 | notYet | node | "a literal method call with an unrepresented result" | Y | Y | N | N |  |
| internal/lower/iteration_consume.go:27 | notYet | callbackNode | "an Array.from mapper with overloads or more than two parameters" | Y | Y | N | N |  |
| internal/lower/iteration_consume.go:32 | notYet | callbackNode | "an Array.from mapper with a different result representation" | Y | Y | N | N |  |
| internal/lower/iteration_consume.go:37 | notYet | callbackNode | "an Array.from mapper with a different input representation" | Y | Y | N | N |  |
| internal/lower/iteration_consume.go:43 | notYet | callbackNode | "an Array.from mapper with a different index representation" | Y | Y | N | N |  |
| internal/lower/iteration_consume.go:48 | notYet | where | "a collected iterator value with a different element representation" | Y | Y | N | N |  |
| internal/lower/iteration_consume.go:126 | notYet | binding | "a nested custom iterator destructuring pattern" | Y | Y | N | N |  |
| internal/lower/iteration_consume.go:135 | notYet | binding | "an iterator rest binding with a different element representation" | Y | Y | N | N |  |
| internal/lower/iteration_consume.go:168 | Refused | l.program.Where(binding) | "an iterator binding whose type omits undefined on exhaustion" ; "give the binding a default, [value = fallback], or make the yielded type include undefined; TypeScript does not account for early iterator exhaustion" | Y | Y | Y | N | rule: adamic/iterator-exhaustion |
| internal/lower/iteration_consume.go:185 | notYet | binding | "an iterator default with a different binding representation" | Y | Y | N | N |  |
| internal/lower/iteration_origin.go:80 | notYet | where | "a literal method through a view that erases its receiver" | Y | Y | N | N |  |
| internal/lower/iteration_origin.go:92 | notYet | where | "a class method through a view that erases its prototype origin" | Y | Y | N | N |  |
| internal/lower/library_array.go:46 | notYet | node | name + " without one callback" | Y | Y | N | N |  |
| internal/lower/library_array.go:50 | notYet | node | name + " with more than three callback parameters" | Y | Y | N | N |  |
| internal/lower/library_array.go:55 | notYet | node | "toReversed with arguments" | Y | Y | N | N |  |
| internal/lower/library_array.go:61 | notYet | node | "toSorted on optional elements" | Y | Y | N | N |  |
| internal/lower/library_array.go:66 | notYet | node | "toSorted without a comparator on these elements" | Y | Y | N | N |  |
| internal/lower/library_array.go:69 | notYet | node | "toSorted without a comparator on optional elements" | Y | Y | N | N |  |
| internal/lower/library_array.go:77 | notYet | node | "toSorted with an effectful comparator expression" | Y | Y | N | N |  |
| internal/lower/library_array.go:82 | notYet | node | name + " with these arguments" | Y | Y | N | N |  |
| internal/lower/library_array.go:90 | notYet | node | name + " with a value of another type than the elements" | Y | Y | N | N |  |
| internal/lower/library_array.go:103 | notYet | node | name + " with a starting index that isn't a number" | Y | Y | N | N |  |
| internal/lower/library_array.go:180 | notYet | node | "toSpliced with nonnumeric bounds" | Y | Y | N | N |  |
| internal/lower/library_array.go:185 | notYet | node | "toSpliced with different element types" | Y | Y | N | N |  |
| internal/lower/library_array.go:209 | notYet | node | "flat with multiple depths" | Y | Y | N | N |  |
| internal/lower/library_array.go:214 | notYet | node | "flat with a nonconstant depth" | Y | Y | N | N |  |
| internal/lower/library_array.go:218 | notYet | node | "flat with this depth" | Y | Y | N | N |  |
| internal/lower/library_array.go:229 | notYet | node | "flat on heterogeneous layers" | Y | Y | N | N |  |
| internal/lower/library_array.go:233 | notYet | node | "flat on optional nested arrays" | Y | Y | N | N |  |
| internal/lower/library_array.go:245 | notYet | node | "flat whose result element type differs from its depth" | Y | Y | N | N |  |
| internal/lower/library_array.go:265 | notYet | node | "copyWithin with these arguments" | Y | Y | N | N |  |
| internal/lower/library_array.go:274 | notYet | node | "copyWithin with a nonnumeric bound" | Y | Y | N | N |  |
| internal/lower/library_array.go:316 | notYet | node | "with without an index and value" | Y | Y | N | N |  |
| internal/lower/library_array.go:328 | notYet | node | "with with incompatible arguments" | Y | Y | N | N |  |
| internal/lower/library_array.go:351 | notYet | node | "flatMap with other than one callback" | Y | Y | N | N |  |
| internal/lower/library_array.go:359 | notYet | node | "flatMap with this callback" | Y | Y | N | N |  |
| internal/lower/library_array.go:368 | notYet | node | "flatMap with a mixed scalar and array callback result" | Y | Y | N | N |  |
| internal/lower/library_array.go:372 | notYet | node | "flatMap with optional array results" | Y | Y | N | N |  |
| internal/lower/library_array.go:376 | notYet | node | "flatMap with this callback element type" | Y | Y | N | N |  |
| internal/lower/library_array.go:393 | notYet | node | "flatMap with this callback parameter type" | Y | Y | N | N |  |
| internal/lower/library_array.go:503 | notYet | node | "for await" | Y | Y | N | N |  |
| internal/lower/library_array.go:507 | notYet | node | "an array iterator loop without const or let" | Y | Y | N | N |  |
| internal/lower/library_array.go:511 | notYet | node | "an array iterator loop with multiple declarations" | Y | Y | N | N |  |
| internal/lower/library_array.go:550 | notYet | node | "destructuring more than two array entry fields" | Y | Y | N | N |  |
| internal/lower/library_array.go:558 | notYet | node | "an array entry binding with a default or rest" | Y | Y | N | N |  |
| internal/lower/library_array.go:571 | notYet | node | "destructuring this array iterator" | Y | Y | N | N |  |
| internal/lower/library_array.go:599 | notYet | node | "join on heterogeneous nested arrays" | Y | Y | N | N |  |
| internal/lower/library_array.go:607 | notYet | node | "join on recursively nested arrays" | Y | Y | N | N |  |
| internal/lower/library_array.go:612 | notYet | node | "join on nested arrays of objects or functions" | Y | Y | N | N |  |
| internal/lower/library_array.go:617 | notYet | node | "join with more than one separator" | Y | Y | N | N |  |
| internal/lower/library_array.go:626 | notYet | node | "join with a nonstring separator" | Y | Y | N | N |  |
| internal/lower/library_for_in.go:17 | notYet | statement.Expression | "for...in over an array (holes and own enumerable properties are not represented; use for...of for elements)" | Y | Y | Y | N |  |
| internal/lower/library_for_in.go:23 | notYet | statement.Expression | "for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly)" | Y | Y | N | N |  |
| internal/lower/library_for_in.go:30 | notYet | statement.Expression | "for...in over a value that is not a plain object" | Y | Y | N | N |  |
| internal/lower/library_for_in.go:37 | Refused | l.program.Where(initializer) | "var" ; "use const or let" | Y | Y | Y | N | rule: adamic/no-var |
| internal/lower/library_for_in.go:41 | notYet | initializer | "a for...in binding that is not one plain name" | Y | Y | N | N | alternative: write for (const key in object) |
| internal/lower/library_for_in.go:48 | notYet | initializer | "a for...in assignment that is not a string variable" | Y | Y | N | N | alternative: declare a fresh string binding with for (const key in object), then assign it in the body |
| internal/lower/library_function_expressions.go:12 | notYet | node | "a generator function expression" | Y | Y | N | N |  |
| internal/lower/library_function_expressions.go:15 | notYet | node | "a generic function expression" | Y | Y | N | N | alternative: declare a generic function at module scope and call it with concrete type arguments |
| internal/lower/library_function_expressions.go:19 | notYet | parameter | "a function expression with a this parameter (dynamic receivers are not implemented)" | Y | Y | N | N | alternative: take the receiver as an explicit named parameter instead of this |
| internal/lower/library_function_expressions.go:33 | Refused | l.program.Where(inner) | "this in a function expression" ; "dynamic receivers are not implemented; capture a named object, or use an arrow in a method" | Y | Y | Y | N | rule: adamic/no-dynamic-this |
| internal/lower/library_globals.go:29 | notYet | node | name + " as a value outside equality or typeof (overloaded calls and static properties need their own representation)" | Y | Y | N | N |  |
| internal/lower/library_json_stringify.go:14 | Refused | l.program.Where(node) | "JSON.parse: its result's type can't be proven from the text" ; "a checked parse against a declared type is a later design; construct typed values explicitly for now" | Y | Y | Y | N | rule: adamic/checked-json-parse |
| internal/lower/library_json_stringify.go:17 | notYet | node | "JSON." + name | Y | Y | N | N |  |
| internal/lower/library_json_stringify.go:21 | notYet | node | "JSON.stringify with spread or extra arguments" | Y | Y | N | N |  |
| internal/lower/library_json_stringify.go:33 | notYet | args[1] | "JSON.stringify replacer functions (the callback must have a proven type for every visited value and its holder)" | Y | Y | N | N | alternative: construct the serializable values explicitly before stringify, or use a key-array replacer to select fields |
| internal/lower/library_json_stringify.go:41 | notYet | args[1] | "JSON.stringify replacer other than a key array, null or undefined" | Y | Y | N | N | alternative: pass an array of string keys, null or undefined as the replacer |
| internal/lower/library_json_stringify.go:47 | notYet | args[1] | "JSON.stringify replacer keys with object coercion" | Y | Y | N | N | alternative: convert replacer keys to strings explicitly before the call |
| internal/lower/library_json_stringify.go:52 | notYet | args[1] | "JSON.stringify replacer keys with object coercion" | Y | Y | N | N | alternative: convert replacer keys to strings explicitly before the call |
| internal/lower/library_json_stringify.go:61 | notYet | args[2] | "JSON.stringify space requiring object coercion" | Y | Y | N | N | alternative: pass a number or string as the space argument |
| internal/lower/library_json_stringify.go:85 | notYet | v | "JSON.stringify a literal with " + name + " semantics" | Y | Y | N | N |  |
| internal/lower/library_json_stringify.go:89 | notYet | v | "JSON.stringify duplicate literal keys" | Y | Y | N | N |  |
| internal/lower/library_json_stringify.go:97 | notYet | v | "JSON.stringify a literal field with a two-word representation" | Y | Y | N | N |  |
| internal/lower/library_json_stringify.go:106 | notYet | v | "JSON.stringify a spread or hole in a literal" | Y | Y | N | N |  |
| internal/lower/library_json_stringify.go:115 | notYet | f | "JSON.stringify a literal with spread, shorthand or methods" | Y | Y | N | N | alternative: write explicit field: value entries in a plain literal |
| internal/lower/library_json_stringify.go:119 | notYet | key | "JSON.stringify computed keys" | Y | Y | N | N | alternative: spell each key as an identifier or string literal |
| internal/lower/library_json_stringify.go:125 | notYet | key | "JSON.stringify this numeric key" | Y | Y | N | N |  |
| internal/lower/library_json_stringify.go:129 | notYet | key | "JSON.stringify a numeric key outside the array-index range (spell it as a string)" | Y | Y | Y | N |  |
| internal/lower/library_json_stringify.go:161 | notYet | node | "JSON.stringify recursive array types" | Y | Y | N | N |  |
| internal/lower/library_json_stringify.go:171 | notYet | node | "JSON.stringify a value of type " + l.checker.TypeToString(t) | Y | Y | N | N |  |
| internal/lower/library_json_stringify.go:177 | notYet | node | "JSON.stringify an array without a proven element type" | Y | Y | N | N |  |
| internal/lower/library_json_stringify.go:183 | notYet | node | "JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata)" | Y | Y | N | N |  |
| internal/lower/library_json_stringify.go:192 | notYet | node | "JSON.stringify a union containing containers without runtime element metadata" | Y | Y | N | N |  |
| internal/lower/library_json_stringify.go:198 | notYet | node | "JSON.stringify this representation" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:21 | notYet | node | name + " with anything but a concrete library Set" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:32 | notYet | node | name + " between Sets whose elements have different representations" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:111 | notYet | node | part + " with arguments" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:144 | notYet | item | "a Set literal element held otherwise than the Set's" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:165 | notYet | node | "forEach thisArg with a callback other than an arrow" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:193 | notYet | arguments[1] | "forEach thisArg without a value representation" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:206 | notYet | node | "forEach with a callback result held in more than one word" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:272 | notYet | name | "destructuring a scalar iterator element" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:280 | notYet | binding | "an iterator binding that isn't plain" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:288 | notYet | binding | "an iterator binding held in more than one word" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:305 | notYet | node | "Map.groupBy with other than an iterable and callback" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:343 | notYet | sourceNode | "Map.groupBy over an unsupported iterable" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:347 | notYet | sourceNode | "Map.groupBy over an unsupported iterable" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:353 | notYet | sourceNode | "Map.groupBy elements held in more than one word" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:367 | notYet | node | "Map.groupBy without known grouped elements" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:371 | notYet | node | "Map.groupBy grouped elements held otherwise than the input's" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:375 | notYet | node | "Map.groupBy widening its Map entries" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:380 | notYet | node | "Map.groupBy widening its Map entry components" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:384 | notYet | node | "Map.groupBy widening its mutable elements" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:391 | notYet | node | "Map.groupBy without a function callback" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:395 | notYet | node | "Map.groupBy with an overloaded callback" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:400 | notYet | node | "Map.groupBy callback elements held otherwise than the input's" | Y | Y | N | N |  |
| internal/lower/library_map_set.go:405 | notYet | node | "Map.groupBy callback keys held otherwise than the result's" | Y | Y | N | N |  |
| internal/lower/library_math_number.go:59 | notYet | node | "Number conversion of a union containing objects" | Y | Y | N | N |  |
| internal/lower/library_math_number.go:62 | notYet | node | "Number conversion of an object" | Y | Y | N | N |  |
| internal/lower/library_math_number.go:75 | notYet | node | "Number with spread or extra arguments" | Y | Y | N | N |  |
| internal/lower/library_math_number.go:90 | notYet | node | "Math." + name + " with spread or extra arguments" | Y | Y | N | N |  |
| internal/lower/library_math_number.go:107 | notYet | node | "Number.hasOwnProperty with other than one key" | Y | Y | N | N |  |
| internal/lower/library_math_number.go:114 | notYet | node | "Number.hasOwnProperty with a non-string key" | Y | Y | N | N |  |
| internal/lower/library_math_number.go:134 | notYet | node | "Number prototype method without a numeric receiver" | Y | Y | N | N |  |
| internal/lower/library_math_number.go:145 | notYet | node | "Number prototype method on a non-number receiver" | Y | Y | N | N |  |
| internal/lower/library_math_number.go:156 | notYet | node | "Number.prototype." + name | Y | Y | N | N |  |
| internal/lower/library_math_number.go:159 | notYet | node | "Number prototype format with extra arguments" | Y | Y | N | N |  |
| internal/lower/library_math_number.go:171 | notYet | node | "Number prototype format with a non-number argument" | Y | Y | N | N |  |
| internal/lower/library_object.go:13 | Refused | l.program.Where(node) | "Object." + name ; reason | Y | Y | N | N | rule: adamic/object-shape |
| internal/lower/library_object.go:23 | notYet | node | "Object.groupBy's partial record with dynamically present keys (use Map and an explicitly typed grouping loop)" | Y | Y | Y | N |  |
| internal/lower/library_object.go:32 | notYet | node | "Object.assign without a target" | Y | Y | N | N | alternative: pass a present plain-object target as the first argument |
| internal/lower/library_object.go:35 | notYet | node | "Object." + name + " with these arguments" | Y | Y | N | N |  |
| internal/lower/library_object.go:64 | notYet | written[0] | "Object." + name + " on a shape not proven by a plain literal or its const binding" | Y | Y | N | N |  |
| internal/lower/library_object.go:71 | notYet | written[0] | "Object." + name + " on other than a present plain object" | Y | Y | N | N |  |
| internal/lower/library_object.go:127 | notYet | argument | "Object.assign adding a field to its target's fixed shape" | Y | Y | N | N | alternative: construct a new literal with every destination field declared explicitly |
| internal/lower/library_object.go:144 | notYet | node | "Object." + name | Y | Y | N | N |  |
| internal/lower/library_string.go:91 | notYet | node | "String conversion of an object, array, map or function (ToPrimitive is not lowered)" | Y | Y | N | N | alternative: format the needed scalar fields explicitly instead of relying on object coercion |
| internal/lower/library_string.go:102 | notYet | node | "String with spread or extra arguments" | Y | Y | N | N |  |
| internal/lower/library_string.go:118 | notYet | node | "String.prototype." + method + ".call without a present receiver" | Y | Y | N | N | alternative: pass a present string receiver as the first argument to .call |
| internal/lower/library_string.go:122 | notYet | node | "String prototype call on null or undefined (its TypeError is not catchable natively yet)" | Y | Y | N | N | alternative: check the receiver for null and undefined before calling the method |
| internal/lower/library_string.go:126 | notYet | node | "String.prototype." + method + " on a non-string receiver (requires a String internal slot)" | Y | Y | N | N | alternative: use String(value) on a scalar first, then call the method on that string |
| internal/lower/library_string.go:145 | notYet | node | "localeCompare (Node uses locale collation, not ordinal UTF-16 order)" | Y | Y | N | N |  |
| internal/lower/library_string.go:154 | notYet | node | name + " with arguments" | Y | Y | N | N |  |
| internal/lower/library_string.go:186 | notYet | node | "String.prototype." + name + " (no sound lowering for this method yet)" | Y | Y | N | N |  |
| internal/lower/library_string.go:189 | notYet | node | name + " with these arguments" | Y | Y | N | N |  |
| internal/lower/library_string.go:201 | notYet | arg | "a " + typeName(lowered.Type()) + " argument to " + name | Y | Y | N | N |  |
| internal/lower/library_string.go:282 | notYet | node | "String.raw without a template" | Y | Y | N | N | alternative: pass a template object such as { raw: ["text"] } |
| internal/lower/library_string.go:286 | notYet | node | "String.raw without a present array of strings in raw" | Y | Y | N | N | alternative: pass an object with a present raw string array: { raw: ["text"] } |
| internal/lower/library_string.go:290 | notYet | node | "String.raw with raw elements that are not strings" | Y | Y | N | N | alternative: convert the raw elements to strings explicitly before the call |
| internal/lower/library_string.go:297 | notYet | node | "String.raw with a template that is not an object" | Y | Y | N | N | alternative: pass a plain template object with a raw string array |
| internal/lower/library_string.go:364 | notYet | node | "a tagged template other than the intrinsic String.raw" | Y | Y | N | N | alternative: call the tag as an ordinary function with explicit arguments and handle raw text explicitly |
| internal/lower/locals.go:15 | Refused | l.program.Where(list) | "var" ; "use const or let" | Y | Y | Y | N | rule: adamic/no-var |
| internal/lower/locals.go:30 | notYet | name | "a destructuring declaration" | Y | Y | N | N |  |
| internal/lower/locals.go:131 | notYet | identifier | "a " + typeName(declared.Type) + " variable a function value captures" | Y | Y | N | N |  |
| internal/lower/locals.go:149 | notYet | name | "a function value that captures the variable its own initializer declares" | Y | Y | N | N | alternative: use a module-level function declaration for noncapturing recursion, passing needed state as parameters |
| internal/lower/modules.go:20 | Refused | l.program.Where(from) | "an import cycle" ; "move what both modules need into a third that neither imports" | Y | Y | Y | N | rule: adamic/import-cycle |
| internal/lower/object.go:25 | Refused | l.program.Where(property) | "a spread after the first field" ; "spread once, first: { ...source, field: value } (adamic/single-spread)" | Y | Y | Y | Y |  |
| internal/lower/object.go:32 | notYet | property | "spreading a " + typeName(spread.Type()) | Y | Y | N | N |  |
| internal/lower/object.go:40 | notYet | property | "an unsupported computed method" | Y | Y | N | N |  |
| internal/lower/object.go:50 | notYet | property | "a string field with the reserved iterator slot name" | Y | Y | N | N | alternative: rename the string field; use [Symbol.iterator] only for an iterator method |
| internal/lower/object.go:54 | notYet | name | "a computed field name" | Y | Y | N | N | alternative: spell the field name as an identifier or string literal |
| internal/lower/object.go:57 | Refused | l.program.Where(property) | "__proto__ in an object literal" ; "JavaScript changes the prototype instead of making an own field; Adamic objects have fixed shapes and no prototype mutation" | Y | Y | N | N | rule: adamic/no-prototype-mutation |
| internal/lower/object.go:70 | notYet | property | "a spread that adds a field the source doesn't have" | Y | Y | N | N | alternative: construct a literal with all fields written explicitly instead of spreading |
| internal/lower/object.go:77 | notYet | property | "a field holding " + typeName(value.Type()) | Y | Y | N | N |  |
| internal/lower/object.go:81 | notYet | property | describe(property) + " in an object literal" | Y | Y | N | N |  |
| internal/lower/object.go:117 | notYet | spread | "spreading a value that may be undefined, whose field " + property.Name + " can't be left undefined yet" | Y | Y | N | N |  |
| internal/lower/object.go:129 | notYet | node | "a value of type " + l.checker.TypeToString(declared) | Y | Y | N | N |  |
| internal/lower/object.go:197 | notYet | node | "a custom spread whose element representation differs from its destination" | Y | Y | N | N |  |
| internal/lower/object.go:209 | notYet | node | "a spread whose element representation differs from its destination" | Y | Y | N | N |  |
| internal/lower/object.go:216 | notYet | item | describe(item) + " in an array literal" | Y | Y | N | N |  |
| internal/lower/object.go:233 | notYet | item | "spreading a " + typeName(value.Type()) + " among other elements" | Y | Y | N | N |  |
| internal/lower/object.go:236 | notYet | item | "spreading an array of other elements" | Y | Y | N | N |  |
| internal/lower/object.go:276 | notYet | node | "a value of type " + l.checker.TypeToString(arrayType) + " where an array goes" | Y | Y | N | N |  |
| internal/lower/object.go:282 | notYet | node | "an array of " + l.checker.TypeToString(element) | Y | Y | N | N |  |
| internal/lower/object.go:298 | notYet | node | "a collection iterator property other than next" | Y | Y | N | N |  |
| internal/lower/object.go:302 | notYet | node | "an optional chain longer than one step" | Y | Y | N | N |  |
| internal/lower/object.go:316 | notYet | node | "Math." + name | Y | Y | N | N |  |
| internal/lower/object.go:322 | notYet | node | "Number." + name | Y | Y | N | N |  |
| internal/lower/object.go:342 | Refused | l.program.Where(node) | "a method read off its object, which loses its this when called (unbound-method)" ; fmt.Sprintf("wrap the call in an arrow function, which keeps its object: (value) =&gt; %s.%s(value)", object, name) | Y | Y | Y | Y |  |
| internal/lower/object.go:352 | notYet | node | "." + name + " on a tuple" | Y | Y | N | N | alternative: copy the tuple elements into a typed array before using array members |
| internal/lower/object.go:362 | notYet | node | "a prototype property on a RegExp" | Y | Y | N | N |  |
| internal/lower/object.go:371 | notYet | node | "a named-group key with an Object-prototype type" | Y | Y | N | N |  |
| internal/lower/object.go:406 | notYet | node | "a prototype property on a RegExp result" | Y | Y | N | N |  |
| internal/lower/object.go:423 | notYet | node | "optional chaining to ." + name + " on a " + typeName(object.Type()) | Y | Y | N | N |  |
| internal/lower/object.go:447 | Refused | l.program.Where(node) | "a narrowed accessor reread, which can return a different value" ; "read the getter into a local once, then narrow and use that local" | Y | Y | Y | N | rule: adamic/accessor-reread |
| internal/lower/object.go:466 | notYet | node | "a narrowed boolean &#124; undefined field; copy the field into a local and narrow that local instead" | Y | Y | Y | N |  |
| internal/lower/object.go:484 | notYet | node | "a field of type " + l.checker.TypeToString(l.checker.GetTypeAtLocation(node)) | Y | Y | N | N |  |
| internal/lower/object.go:492 | notYet | node | "?. to a " + typeName(of) + ", which would be " + typeName(of) + " &#124; undefined" | Y | Y | N | N |  |
| internal/lower/object.go:496 | notYet | node | "." + name + " on a " + typeName(object.Type()) | Y | Y | N | N |  |
| internal/lower/object.go:527 | notYet | node | "hasOwnProperty with other than one argument" | Y | Y | N | N |  |
| internal/lower/object.go:530 | notYet | node | "hasOwnProperty with a key that isn't a string" | Y | Y | N | N |  |
| internal/lower/object.go:569 | Refused | l.program.Where(node) | "Math.random" ; "0.1 programs are deterministic, so the oracle can hold them to Node; compute the values you need, with a generator of your own seeded by a constant" | Y | Y | Y | N | rule: adamic/deterministic |
| internal/lower/object.go:672 | notYet | node | name + " on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made)" | Y | Y | Y | N |  |
| internal/lower/object.go:695 | notYet | argument | "a spread argument to " + name | Y | Y | N | N |  |
| internal/lower/object.go:702 | notYet | argument | "a " + typeName(lowered.Type()) + " argument to " + name | Y | Y | N | N |  |
| internal/lower/object.go:712 | notYet | node | "Math." + name | Y | Y | N | N |  |
| internal/lower/object.go:715 | notYet | node | "Math." + name + " with other than its arguments" | Y | Y | N | N |  |
| internal/lower/object.go:720 | notYet | node | "toFixed with more than one argument" | Y | Y | N | N |  |
| internal/lower/object.go:739 | notYet | node | name + " with more than one argument" | Y | Y | N | N |  |
| internal/lower/object.go:748 | notYet | arguments[0] | "a " + typeName(argument.Type()) + " argument to " + name | Y | Y | N | N |  |
| internal/lower/object.go:766 | notYet | node | "Number." + name | Y | Y | N | N |  |
| internal/lower/object.go:774 | notYet | node | name + " with these arguments" | Y | Y | N | N |  |
| internal/lower/object.go:784 | notYet | argument | "a " + typeName(lowered.Type()) + " argument to " + name | Y | Y | N | N |  |
| internal/lower/object.go:805 | notYet | node | "for await" | Y | Y | N | N |  |
| internal/lower/object.go:809 | notYet | initializer | "a for...of that doesn't declare its variable with const or let" | Y | Y | N | N | alternative: declare one loop binding with for (const value of source) |
| internal/lower/object.go:813 | notYet | initializer | "a for...of declaring more than one variable" | Y | Y | N | N | alternative: declare one loop binding and declare additional locals in the body |
| internal/lower/object.go:817 | notYet | initializer | "a for...of destructuring an object" | Y | Y | N | N | alternative: bind the element to a name, then read its fields in the body |
| internal/lower/object.go:876 | notYet | iterated | "for...of over an object" | Y | Y | N | N |  |
| internal/lower/object.go:880 | notYet | statement.Expression | "for...of over a " + typeName(iterable.Type()) | Y | Y | N | N |  |
| internal/lower/object.go:895 | notYet | name | "destructuring a " + typeName(element) | Y | Y | N | N |  |
| internal/lower/object.go:903 | notYet | binding | "a destructured name that isn't plain" | Y | Y | N | N | alternative: bind the element to a name and destructure it in separate statements |
| internal/lower/object.go:910 | notYet | binding | "a tuple element of type " + typeName(of) | Y | Y | N | N |  |
| internal/lower/object.go:940 | notYet | name | "destructuring more than a key and a value" | Y | Y | N | N | alternative: destructure only [key, value] |
| internal/lower/object.go:948 | notYet | binding | "a destructured name that isn't plain" | Y | Y | N | N | alternative: bind [key, value] as plain names and handle defaults in the body |
| internal/lower/object.go:961 | notYet | name | "for...of over a map's " + part + " into this name (write [key, value], or iterate keys() or values())" | Y | Y | Y | N |  |
| internal/lower/object.go:983 | notYet | inner | "a declaration directly in a case (wrap the case in a block)" | Y | Y | Y | N |  |
| internal/lower/object.go:995 | notYet | clause | "a case that isn't a constant" | Y | Y | N | N |  |
| internal/lower/object.go:998 | notYet | clause | "a case whose type differs from the switch's" | Y | Y | N | N |  |
| internal/lower/object.go:1040 | notYet | node | "map with other than one callback" | Y | Y | N | N | alternative: pass one callback: items.map((value) =&gt; mappedValue) |
| internal/lower/object.go:1047 | notYet | arguments[0] | "map with a callback that isn't a function" | Y | Y | N | N | alternative: pass a function or arrow callback |
| internal/lower/object.go:1064 | notYet | node | "concat on an array of arrays" | Y | Y | N | N |  |
| internal/lower/object.go:1077 | notYet | node | "slice with an index that isn't a number" | Y | Y | N | N |  |
| internal/lower/object.go:1081 | notYet | node | "slice with more than two arguments" | Y | Y | N | N |  |
| internal/lower/object.go:1087 | notYet | node | "push with other than one value" | Y | Y | N | N |  |
| internal/lower/object.go:1094 | notYet | node | name + " with a starting index" | Y | Y | N | N |  |
| internal/lower/object.go:1097 | notYet | node | name + " with a value of another type than the elements" | Y | Y | N | N |  |
| internal/lower/object.go:1102 | notYet | node | "at with other than one number" | Y | Y | N | N |  |
| internal/lower/object.go:1112 | notYet | node | "fill with other than a value of the elements' type" | Y | Y | N | N |  |
| internal/lower/object.go:1117 | notYet | node | "fill with a bound that isn't a number" | Y | Y | N | N |  |
| internal/lower/object.go:1128 | notYet | node | "splice without a start and a count that are numbers" | Y | Y | N | N |  |
| internal/lower/object.go:1135 | notYet | node | "splice inserting a value of another type than the elements" | Y | Y | N | N |  |
| internal/lower/object.go:1144 | notYet | node | "concat with a value that isn't an array (JavaScript appends it)" | Y | Y | N | N |  |
| internal/lower/object.go:1147 | notYet | node | "concat of arrays of different elements" | Y | Y | N | N |  |
| internal/lower/object.go:1155 | notYet | node | "join on an array of objects, arrays, maps or functions" | Y | Y | N | N |  |
| internal/lower/object.go:1161 | notYet | node | "join with a separator that isn't a string" | Y | Y | N | N |  |
| internal/lower/object.go:1165 | notYet | node | "join with more than one argument" | Y | Y | N | N |  |
| internal/lower/object.go:1187 | notYet | node | "new Array filled with other than one length and one value (a part left unfilled is a hole)" | Y | Y | N | N |  |
| internal/lower/object.go:1198 | notYet | node | "new Array(length).fill(value) with a length that isn't a number or a value of another type" | Y | Y | N | N |  |
| internal/lower/object.go:1212 | notYet | node | name + " with other than one callback" | Y | Y | N | N | alternative: pass one function or arrow callback |
| internal/lower/object.go:1219 | notYet | arguments[0] | name + " with a callback that isn't a function" | Y | Y | N | N | alternative: pass a function or arrow callback |
| internal/lower/object.go:1223 | notYet | arguments[0] | name + " with an overloaded callback" | Y | Y | N | N | alternative: wrap the callback in an arrow with one concrete signature |
| internal/lower/object.go:1229 | notYet | arguments[0] | name + " with a callback returning " + l.checker.TypeToString(result) | Y | Y | N | N |  |
| internal/lower/object.go:1233 | Refused | l.program.Where(arguments[0]) | "a " + name + " callback that doesn't return a boolean" ; "return a comparison, like word.length &gt; 0: 0.1 has no truthiness" | Y | Y | Y | N | rule: strict-boolean-expressions |
| internal/lower/object.go:1243 | Refused | l.program.Where(node) | "reduce without an initial value" ; "pass one, like reduce((sum, value) =&gt; sum + value, 0): without it, an empty array throws" | Y | Y | Y | N | rule: adamic/reduce-initial |
| internal/lower/object.go:1246 | notYet | node | "reduce with other than a callback and an initial value" | Y | Y | N | N | alternative: pass a callback and an initial value: items.reduce((sum, value) =&gt; sum + value, 0) |
| internal/lower/object.go:1253 | notYet | arguments[0] | "reduce with a callback that isn't a function" | Y | Y | N | N | alternative: pass a function or arrow as the first argument |
| internal/lower/object.go:1265 | notYet | node | "reduce to a " + l.checker.TypeToString(l.checker.GetTypeAtLocation(node)) | Y | Y | N | N |  |
| internal/lower/object.go:1274 | notYet | node | "a Map whose key and value types aren't known" | Y | Y | N | N |  |
| internal/lower/object.go:1279 | notYet | node | "a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions" | Y | Y | N | N |  |
| internal/lower/object.go:1283 | notYet | node | "a Map of " + l.checker.TypeToString(arguments[1]) | Y | Y | N | N |  |
| internal/lower/object.go:1305 | notYet | node | "new " + describe(created.Expression) | Y | Y | N | N |  |
| internal/lower/object.go:1316 | notYet | node | "new Map with more than one argument" | Y | Y | N | N |  |
| internal/lower/object.go:1329 | notYet | pair | "a Map entry that isn't [key, value] written out" | Y | Y | N | N |  |
| internal/lower/object.go:1344 | notYet | pair | "a Map entry whose key or value is of another type than the Map's" | Y | Y | N | N |  |
| internal/lower/object.go:1400 | notYet | node | "map." + name + " with arguments of other types" | Y | Y | N | N |  |
| internal/lower/object.go:1422 | notYet | node | "trim with arguments" | Y | Y | N | N |  |
| internal/lower/object.go:1428 | notYet | node | "charCodeAt with more than one argument" | Y | Y | N | N |  |
| internal/lower/object.go:1435 | notYet | node | "charCodeAt with an index that isn't a number" | Y | Y | N | N |  |
| internal/lower/object.go:1453 | notYet | property | "reading " + property.Name().Text() | Y | Y | N | N |  |
| internal/lower/object.go:1458 | notYet | property | "a field from a " + typeName(of) + " variable" | Y | Y | N | N |  |
| internal/lower/object.go:1500 | notYet | node | name + " with these arguments" | Y | Y | N | N |  |
| internal/lower/object.go:1512 | notYet | argument | "a " + typeName(lowered.Type()) + " argument to " + name | Y | Y | N | N |  |
| internal/lower/object.go:1534 | Refused | l.program.Where(node) | "sort without a comparator" ; "pass one: the default compares numbers as strings, so [10, 9, 1].sort() is [1, 10, 9]" | Y | Y | Y | N | rule: adamic/sort-comparator |
| internal/lower/object.go:1546 | notYet | comparator | "a comparator that isn't a function" | Y | Y | N | N | alternative: pass a comparator function with one concrete signature |
| internal/lower/object.go:1549 | notYet | comparator | "a comparator that doesn't return a number" | Y | Y | N | N | alternative: return a numeric ordering from the comparator, for example (a, b) =&gt; a - b for numbers |
| internal/lower/object.go:1555 | notYet | comparator | "a comparator that doesn't take two elements and return a number" | Y | Y | N | N | alternative: use a comparator taking two elements and returning a number |
| internal/lower/object.go:1571 | notYet | node | "an optional chain longer than one step" | Y | Y | N | N |  |
| internal/lower/object.go:1579 | notYet | node | "a computed named-group key" | Y | Y | N | N |  |
| internal/lower/object.go:1586 | notYet | node | "a named-group key with an Object-prototype type" | Y | Y | N | N |  |
| internal/lower/object.go:1592 | notYet | node | "?.[] on a " + typeName(object.Type()) | Y | Y | N | N |  |
| internal/lower/object.go:1600 | notYet | node | "a string index that isn't a number" | Y | Y | N | N |  |
| internal/lower/object.go:1614 | notYet | node | "an array index that isn't a number" | Y | Y | N | N |  |
| internal/lower/object.go:1623 | notYet | node | describe(node) | Y | Y | N | N |  |
| internal/lower/object.go:1638 | notYet | node | "a tuple element of type " + typeName(of) | Y | Y | N | N |  |
| internal/lower/object.go:1651 | notYet | target | "assigning an element of a " + typeName(array.Type()) | Y | Y | N | N |  |
| internal/lower/object.go:1662 | notYet | target | "an array index that isn't a number" | Y | Y | N | N |  |
| internal/lower/object.go:1685 | notYet | argument | "a spread argument to String." + node.AsCallExpression().Expression.Name().Text() | Y | Y | N | N |  |
| internal/lower/object.go:1692 | notYet | argument | "a " + typeName(value.Type()) + " argument to String." + node.AsCallExpression().Expression.Name().Text() | Y | Y | N | N |  |
| internal/lower/object.go:1723 | notYet | node | "a tuple literal with more values than its tuple has elements" | Y | Y | N | N |  |
| internal/lower/object.go:1731 | notYet | node | "a tuple literal leaving out an element of type " + l.checker.TypeToString(elements[index]) | Y | Y | N | N |  |
| internal/lower/object.go:1737 | notYet | item | describe(item) + " in a tuple literal" | Y | Y | N | N |  |
| internal/lower/object.go:1741 | notYet | item | "a tuple element of type " + l.checker.TypeToString(elements[index]) | Y | Y | N | N |  |
| internal/lower/object.go:1749 | notYet | item | "a tuple element of another type than its tuple's" | Y | Y | N | N |  |
| internal/lower/object.go:1782 | notYet | element | describe(element) + " in an array spread into a call" | Y | Y | N | N |  |
| internal/lower/object.go:1789 | notYet | element | "a " + typeName(value.Type()) + " argument where numbers go" | Y | Y | N | N |  |
| internal/lower/object.go:1802 | notYet | argument | "spreading other than an array of numbers into a call" | Y | Y | N | N |  |
| internal/lower/object.go:1805 | notYet | argument | "a " + typeName(value.Type()) + " argument where numbers go" | Y | Y | N | N |  |
| internal/lower/object.go:1896 | notYet | target | "assigning an element of a " + typeName(array.Type()) | Y | Y | N | N |  |
| internal/lower/object.go:1903 | notYet | target | "updating an array element of type " + typeName(element) | Y | Y | N | N |  |
| internal/lower/object.go:1910 | notYet | target | "an array index that isn't a number" | Y | Y | N | N |  |
| internal/lower/prelude.go:59 | notYet | callee | "a call to " + describe(callee) | Y | Y | N | N |  |
| internal/lower/prelude.go:64 | notYet | callee | "a method call" | Y | Y | N | N |  |
| internal/lower/prelude.go:72 | notYet | callee | "console." + callee.Name().Text() | Y | Y | N | N |  |
| internal/lower/prototype.go:72 | Refused | l.program.Where(node) | "isPrototypeOf" ; "Adamic has no observable prototype chain; use instanceof for class identity or an explicit discriminant" | Y | Y | Y | N | rule: adamic/no-prototype-reflection |
| internal/lower/prototype.go:74 | Refused | l.program.Where(node) | "inherited library member " + name + " read as an own field" ; "prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method)" | Y | Y | Y | Y |  |
| internal/lower/prototype.go:88 | Refused | l.program.Where(node) | "isPrototypeOf" ; "Adamic has no observable prototype chain; use instanceof for class identity or an explicit discriminant" | Y | Y | Y | N | rule: adamic/no-prototype-reflection |
| internal/lower/prototype.go:91 | notYet | node | name + " on Error (its prototype and non-enumerable own descriptors differ from plain objects)" | Y | Y | N | N |  |
| internal/lower/prototype.go:96 | notYet | node | name + " through an object view (" + reason + ")" | Y | Y | N | N |  |
| internal/lower/prototype.go:103 | Refused | l.program.Where(node) | name + " on a number" ; "JavaScript uses locale-sensitive number formatting; use toString for deterministic formatting" | Y | Y | Y | N | rule: adamic/deterministic |
| internal/lower/prototype.go:111 | Refused | l.program.Where(node) | name + " on this array" ; "JavaScript formats each element with its own toLocaleString, which may be locale-sensitive or user-defined; use toString for deterministic formatting" | Y | Y | Y | N | rule: adamic/deterministic |
| internal/lower/prototype.go:119 | notYet | node | name + " on a tuple (its representation includes absent optional slots and no length descriptor)" | Y | Y | N | N |  |
| internal/lower/prototype.go:126 | notYet | node | name + " with arguments" | Y | Y | N | N |  |
| internal/lower/prototype.go:129 | notYet | node | name + " on a tuple (its native representation is not an array)" | Y | Y | N | N |  |
| internal/lower/prototype.go:134 | notYet | node | "toLocaleString delegating to a user-defined toString" | Y | Y | N | N |  |
| internal/lower/prototype.go:143 | notYet | node | "valueOf whose library result type erases the " + typeName(of) + " representation to Object (keeping or returning that result needs a tagged object view)" | Y | Y | N | N |  |
| internal/lower/prototype.go:170 | notYet | node | name + " on a " + typeName(of) + " (function source text and mixed value dispatch are not represented)" | Y | Y | N | N |  |
| internal/lower/prototype_own.go:20 | notYet | node | "hasOwnProperty on a function (its prototype property depends on whether it was declared or made as an arrow)" | Y | Y | N | N |  |
| internal/lower/prototype_own.go:23 | notYet | node | name + " on a " + typeName(of) + " (mixed value descriptors are not represented)" | Y | Y | N | N |  |
| internal/lower/prototype_own.go:27 | notYet | node | name + " with other than one key" | Y | Y | N | N |  |
| internal/lower/prototype_own.go:30 | notYet | node | name + " with a key that isn't a string" | Y | Y | N | N |  |
| internal/lower/refusals.go:59 | Refused | fmt.Sprintf("%s:%d:%d", l.program.FileName(module), line+1, column+1) | name + " suppression directive" ; "remove it and fix the type error" | Y | Y | Y | N | rule: ban-ts-comment |
| internal/lower/refusals.go:68 | Refused | l.program.Where(node) | refused.what ; refused.fix | Y | Y | N | N | all shared routes now carry IDs |
| internal/lower/refusals.go:81 | Refused | l.program.Where(assertion) | "a definite assignment assertion !" ; "remove ! and initialize it where it is declared or in the constructor, or type it T &#124; undefined" | Y | Y | Y | N | rule: adamic/no-definite-assignment |
| internal/lower/refusals.go:86 | Refused | l.program.Where(node.AsBinaryExpression().OperatorToken) | refused.what ; refused.fix | Y | Y | Y | N | all shared routes now carry IDs |
| internal/lower/refusals.go:100 | Refused | l.program.Where(node) | "a generator function" ; "use an explicit iterator object; suspended frames need ownership and cancellation rules before generators can be compiled without a collector (docs/user-iterators.md)" | Y | Y | Y | N | rule: adamic/no-generators |
| internal/lower/refusals.go:104 | Refused | l.program.Where(node) | "an async function" ; "0.1 has no async; it arrives with the concurrency model" | Y | Y | N | N | rule: adamic/no-async |
| internal/lower/refusals.go:110 | Refused | l.program.Where(node) | "arguments" ; "name the parameters, or take a rest parameter" | Y | Y | Y | N | rule: adamic/no-arguments |
| internal/lower/refusals.go:136 | Refused | l.program.Where(node) | "a method read as a value (" + symbol.Name + " would lose its object, and this with it)" ; "call it in an arrow that keeps the object: (" + parameters + ") =&gt; " + object + "." + symbol.Name + "(" + parameters + ") (unbound-method)" | Y | Y | Y | Y |  |
| internal/lower/regexp.go:20 | notYet | node | "a malformed regular expression literal" | Y | Y | N | N |  |
| internal/lower/regexp.go:33 | notYet | node | "RegExp with more than two arguments" | Y | Y | N | N |  |
| internal/lower/regexp.go:39 | notYet | args[0] | "RegExp with a nonconstant pattern" | Y | Y | N | N |  |
| internal/lower/regexp.go:46 | notYet | args[1] | "RegExp with nonconstant flags" | Y | Y | N | N |  |
| internal/lower/regexp.go:64 | notYet | node | err.Error() | Y | Y | N | N |  |
| internal/lower/regexp.go:145 | notYet | node | "RegExp." + name | Y | Y | N | N |  |
| internal/lower/regexp.go:148 | notYet | node | "RegExp." + name + " without exactly one string" | Y | Y | N | N |  |
| internal/lower/regexp.go:152 | notYet | node | "RegExp iterator." + name | Y | Y | N | N |  |
| internal/lower/regexp.go:171 | notYet | node | "String." + name + " with a RegExp" | Y | Y | N | N |  |
| internal/lower/regexp.go:187 | notYet | node | "RegExp input other than a string" | Y | Y | N | N |  |
| internal/lower/regexp.go:191 | notYet | node | "regex replacement other than a string" | Y | Y | N | N |  |
| internal/lower/regexp.go:207 | notYet | node | "regex split limit other than a number" | Y | Y | N | N |  |
| internal/lower/regexp.go:211 | notYet | node | "an optional RegExp call" | Y | Y | N | N |  |
| internal/lower/regexp.go:297 | notYet | parent | "spreading a RegExp or its iterator" | Y | Y | N | N |  |
| internal/lower/regexp.go:302 | notYet | parent | "overriding a RegExp or iterator property" | Y | Y | N | N |  |
| internal/lower/set.go:23 | notYet | node | "a Set whose element type isn't known" | Y | Y | N | N |  |
| internal/lower/set.go:27 | notYet | node | "a Set of " + l.checker.TypeToString(arguments[0]) + " (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)" | Y | Y | N | N |  |
| internal/lower/set.go:44 | notYet | node | "new Set with more than one argument" | Y | Y | N | N |  |
| internal/lower/set.go:66 | notYet | arguments.Nodes[0] | "new Set from elements held otherwise than the Set's" | Y | Y | N | N |  |
| internal/lower/set.go:94 | notYet | node | "set." + name + " with other than one value" | Y | Y | N | N |  |
| internal/lower/set.go:105 | notYet | arguments[0] | "set." + name + " with a value of another type than the elements" | Y | Y | N | N |  |
| internal/lower/set.go:133 | notYet | name | "destructuring more than a Set entry's two elements" | Y | Y | N | N |  |
| internal/lower/set.go:142 | notYet | binding | "a destructured name that isn't plain" | Y | Y | N | N |  |
| internal/lower/set.go:155 | notYet | name | "a Set entry destructured into no names" | Y | Y | N | N |  |
| internal/lower/set.go:162 | notYet | name | "for...of over a Set's " + part + " into this name" | Y | Y | N | N |  |
| internal/lower/statements.go:35 | Refused | l.program.Where(node) | describe(node) ; "export where you declare: export function, export const (one name for one thing)" | Y | Y | Y | N | rule: adamic/named-declaration-export |
| internal/lower/statements.go:38 | notYet | node | "a function inside a function (a closure)" | Y | Y | N | N | alternative: use an arrow function for a closure, or move a noncapturing function to module scope |
| internal/lower/statements.go:44 | notYet | node | "a class inside a function" | Y | Y | N | N | alternative: declare the class at module scope and pass captured values to its constructor |
| internal/lower/statements.go:73 | notYet | node | "a labeled " + strings.ToLower(strings.TrimSuffix(strings.TrimPrefix(node.Kind.String(), "Kind"), "Statement")) | Y | Y | N | N |  |
| internal/lower/statements.go:84 | notYet | node | describe(node) | Y | Y | N | N |  |
| internal/lower/statements.go:138 | notYet | expression | describe(expression) + " as a statement" | Y | Y | N | N |  |
| internal/lower/tsgo.go:28 | Refused | l.program.Where(node) | "an unlinked typescript-go library call" ; "build with --tsgo &lt;checker archive&gt;" | Y | Y | Y | N | rule: adamic/tsgo-link |

## Shared and computed message producers

These expansions explain dynamic expressions in the main table; they do not add construction sites. The generic helper forwards each caller’s What unchanged. Regexp `err.Error()` forwards the native regexp emitter’s failure text (`native regexp quantifier bounds above uint64 are not yet supported`); cycle and invariant messages interpolate type names and write locations.

### Syntax and operator tables

```go
var refusals = map[ast.Kind]refusal{
	ast.KindAwaitExpression:   {"await", "0.1 has no async; it arrives with the concurrency model"},
	ast.KindYieldExpression:   {"yield (generators)", "build an array, or call a function per item"},
	ast.KindDecorator:         {"a decorator", "write the behavior where it applies; 0.1 doesn't rewrite classes at runtime"},
	ast.KindLabeledStatement:  {"a label", "move the loop into a function and return from it"},
	ast.KindWithStatement:     {"with", "name the object you mean"},
	ast.KindDeleteExpression:  {"delete", "an object's shape is fixed; use a Map for keys that come and go"},
	ast.KindDebuggerStatement: {"debugger", "remove it"},
	ast.KindEnumDeclaration:   {"enum", "use a union of string literals, like 'Circle' | 'Square'"},
	ast.KindModuleDeclaration: {"a namespace", "use a module: a file of its own, with named exports"},
	ast.KindVoidExpression:    {"the void operator", "evaluate the expression as a statement"},
	ast.KindIndexSignature:    {"an index signature", "use a Map, which keeps keys in the order they were added"},
	ast.KindExportAssignment:  {"export default", "export by name: one name for one thing"},
	ast.KindTypePredicate:     {"a type predicate", "narrow where you use it, with ===, typeof or instanceof (adamic/no-type-predicate)"},
	ast.KindNonNullExpression: {"the non-null assertion !", "write ?? panic('why it can't be missing'), or narrow and handle the missing case"},
}

// refusedOperators are binary operators 0.1 refuses.
var refusedOperators = map[ast.Kind]refusal{
	ast.KindEqualsEqualsToken:             {"==", "use ===, which doesn't coerce"},
	ast.KindExclamationEqualsToken:        {"!=", "use !==, which doesn't coerce"},
	ast.KindInKeyword:                     {"in", "an object's shape is known; use a discriminant, or a Map"},
	ast.KindCommaToken:                    {"the comma operator", "write each expression as its own statement"},
	ast.KindAmpersandAmpersandEqualsToken: {"&&=", "write the if"},
	ast.KindBarBarEqualsToken:             {"||=", "write the if"},
}
```

### Override helper calls

```go
return refuse("an inherited method replaced by a field, or a field replaced by a method", "keep the inherited member kind; use a different name for the new member")
return refuse("an inherited data property replaced by an accessor, or an accessor replaced by data", "keep the inherited member kind; use another name for the new property")
return refuse("an accessor override that hides the inherited getter or setter", "override both halves of the inherited descriptor; delegate an unchanged half to super")
return refuse("an accessor override that narrows a setter parameter (adamic/contravariant-override)", "accept the base setter's parameter type or a wider type; narrow it inside the setter")
return refuse("an accessor override with an unsafe read or write type (adamic/invariant-mutable)", "keep the inherited accessor type; narrow values inside the accessor")
return refuse("a mutable inherited field redeclared with a different type (adamic/invariant-mutable)", "keep the base field's type; narrow a local after reading it, or make the field readonly in the base")
return refuse("an inherited readonly field whose mutable contents are narrowed (adamic/invariant-mutable)", "keep mutable contents at the base type, or make the contents readonly in the base too")
return refuse("a readonly inherited field redeclared mutable", "keep the inherited field readonly")
return refuse("an override that narrows a method parameter (adamic/contravariant-override)", "accept the base method's parameter type or a wider type; narrow it inside the method")
return refuse("an override that widens its return type", "return the base method's result type or a subtype")
```

### Object helper calls

```go
return refused("property descriptors can change the presence, type or access behavior of fields; Adamic fields have a fixed shape and are plain loads and stores")
return refused("prototypes expose or replace fields outside the declared shape; use a declared object or class with composition")
return refused("tsc returns an index-signature object with unproven keys; Adamic fixes object shapes and refuses index signatures; use Map")
return refused("hasOwn requires a string literal naming a declared public field or method; use Map for arbitrary keys")
return refused("tsc's result must be an array with a proven element type")
return refused("tsc's entries must have a proven value type")
return refused("tsc's result must have one homogeneous number, string or boolean value type; any and widened field views are unsound")
return refused("every present field must have tsc's result element representation; optional or heterogeneous fields cannot be read soundly")
return refused("tsc's result is any or unknown, not a proven object type; use one to three typed sources")
return refused("each source must have its complete shape proven by a literal or const binding; a widened source can hide overwriting fields")
return refused("source and target field types must agree in both directions with tsc's intersection result; widening, conflicting fields and reference cycles are refused")
```

### Mutable invariance helper

The shared constructor at `invariance.go:462` uses the following message builder. Its branches already have stable IDs.

```go
	what := "a value of type " + l.checker.TypeToString(own) + " seen as " + l.checker.TypeToString(contextual) + ", which can write " + l.checker.TypeToString(found.target) + " where " + l.checker.TypeToString(found.source) + " is read"
	fix := "make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"
	if found.parameter {
		what = "a function taking " + l.checker.TypeToString(found.source) + " seen as one taking " + l.checker.TypeToString(found.target) + " (tsc relates a method's parameters both ways), so it can be handed what it can't take"
		fix = "write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style)"
	}
	if found.readonlyField != "" {
		what = "a value of type " + l.checker.TypeToString(own) + " seen as " + l.checker.TypeToString(contextual) + ", whose readonly field " + found.readonlyField + " becomes writable: a readonly field may hold something narrower than " + l.checker.TypeToString(found.source) + ", which a write of " + l.checker.TypeToString(found.target) + " would replace"
		fix = "keep " + found.readonlyField + " readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"
	}
	if found.target.Flags()&checker.TypeFlagsTypeParameter != 0 {
		what = "a value of type " + l.checker.TypeToString(found.source) + " seen as " + l.checker.TypeToString(found.target) + ", a type parameter whose constraint " + l.checker.TypeToString(l.checker.GetBaseConstraintOfType(found.target)) + " can be written, so it can write what " + l.checker.TypeToString(found.source) + " can't hold"
		fix = "take it as " + l.checker.TypeToString(found.source) + ", or constrain " + l.checker.TypeToString(found.target) + " to something readonly, which can't write (adamic/invariant-mutable)"
	}
	return &Refused{Where: l.program.Where(node), What: what, Fix: fix}
```

## Changes and remaining gaps

Added **73 NotYet alternatives**, **39 direct Refused rule suffixes**, **19 syntax/operator table suffixes**, and **5 override-helper suffixes**. All 61 Refused constructors now guarantee an ID, including shared routes. Alternatives suggest explicit typed operations, named bindings, module-scope declarations, preserved override types, or supported callbacks. They require the writer to preserve intended behavior; they are not automatic rewrites. Exact-text expectations in `lower_test.go` follow the revised words.

**392 NotYet sites still have no guaranteed actionable alternative**, including the forwarding constructor; **4 Refused constructors have an explanation-only route** (async, shared syntax table’s await, __proto__, shared Object helper). Seven accessor constructors still lack line/column. This first pass deliberately leaves alternatives that need deeper semantic investigation for later work. NotYet still has no stable rule IDs.

## Observations and verification

- Setup: `bash cloud/setup.sh`, then `source /workspace/adamic-tools/env.sh`. Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 176s; total 176s; `nproc` = 5. Go 1.27.1, clang 20.1.8, Node v24.19.0.
- Before and after: `go test -json -count=1 -timeout 30m ./internal/lower` and `go test -json -count=1 -timeout 30m ./internal/oracle -run 'Refus|NotYet'`. Each exited 0; 627 lower and 51 oracle passing test/subtest records on each side. Sorted package/test pass, fail and skip records were identical (628 lower and 52 oracle records including package completion). Output files: `/tmp/diagnostics-{before,after}-{lower,oracle}.jsonl`.
- Probes loaded and lowered all `.a`/legacy `.ts` files matching `internal/load/testdata/0.1/refuse/*`, `internal/oracle/testdata/fresh_refused/*.a`, and `internal/oracle/testdata/*.a`, sorted by path. Both runs: 341 programs, 294 accepted, 45 refused, 2 not-yet. `diff -u /tmp/diagnostics-before-probes.jsonl /tmp/diagnostics-after-probes.jsonl > /tmp/diagnostics-outcomes.diff` exited 0 with an empty diff.
- AST comparison of all 21 edited production files against baseline replaced only diagnostic What/Fix expressions, notYet text arguments, shared refusal-helper text arguments, and syntax-table message values. The resulting Go ASTs were identical (`/tmp/diagnostics-structure.log`). This is structural evidence that acceptance branches, predicates and lowerer operations did not change, beyond the sampled probes.
- `gofmt -l cmd internal` and `go vet ./...` exited 0 with empty logs (`/tmp/diagnostics-gofmt.log`, `/tmp/diagnostics-vet.log`). `git diff --check` passed. Three worked proposal records parsed as JSON.
- Optional broader gate: `go test -count=1 -timeout 30m ./... > /tmp/diagnostics-gate.log 2>&1` was stopped after 579.9 seconds while native/oracle work continued. Its process tree was terminated; shell reported `gate_exit=143`. It did not complete and is not claimed green. This was the ordinary worker gate with oracle caching enabled, not the uncached integration gate. The final required lower and filtered oracle suites above completed separately.
- Mutants: not applicable to this wording-only unit, as requested. No new acceptance checks were introduced.

- `before-probes.jsonl` SHA-256: `d6fe8a1751873b8210a6e8b14305f1c392e15bc5bb848afc2241515b75b72d65`.
- `after-probes.jsonl` SHA-256: `d6fe8a1751873b8210a6e8b14305f1c392e15bc5bb848afc2241515b75b72d65`.

### Probe reproduction

Place this temporary Go harness beneath the repository root so it can import internal packages; run with `go run ./.diagnostics-probes > outcomes.jsonl 2> probes.log` at baseline and at this branch, then diff outcomes. The original temporary harness is removed before commit.

```go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	paths := []string{}
	for _, pattern := range []string{"internal/load/testdata/0.1/refuse/*", "internal/oracle/testdata/fresh_refused/*.a", "internal/oracle/testdata/*.a"} {
		matches, e := filepath.Glob(pattern)
		if e != nil {
			panic(e)
		}
		for _, p := range matches {
			if filepath.Ext(p) == ".a" || filepath.Ext(p) == ".ts" {
				paths = append(paths, p)
			}
		}
	}
	sort.Strings(paths)
	out := json.NewEncoder(os.Stdout)
	for _, p := range paths {
		absolute, e := filepath.Abs(p)
		if e != nil {
			panic(e)
		}
		program, e := load.Load([]string{absolute})
		kind := "accepted"
		if e != nil {
			kind = "checker-refused"
		} else {
			_, e = lower.Lower(context.Background(), program)
			if e != nil {
				var r *lower.Refused
				var n *lower.NotYet
				switch {
				case errors.As(e, &r):
					kind = "refused"
				case errors.As(e, &n):
					kind = "not-yet"
				default:
					fmt.Fprintln(os.Stderr, p, e)
					os.Exit(1)
				}
			}
		}
		out.Encode(struct{ Path, Outcome string }{p, kind})
	}
}
```

### Inventory reproduction

The inventory uses Go AST nodes rather than regex matching comments or test expectations. Run this temporary extractor from the repository root with `go run extractor.go > sites.json`, at baseline and after edits. Classifications above are a manual semantic review of those message expressions and shared producers.

```go
package main

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type Site struct {
	File             string
	Line             int
	Kind             string
	Start, End       int
	What, Fix, Where string
	FixStart, FixEnd int
}

func main() {
	fs := token.NewFileSet()
	out := []Site{}
	files, _ := filepath.Glob("internal/lower/*.go")
	for _, p := range files {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		f, e := parser.ParseFile(fs, p, nil, 0)
		if e != nil {
			panic(e)
		}
		show := func(n ast.Node) string { var b bytes.Buffer; format.Node(&b, fs, n); return b.String() }
		ast.Inspect(f, func(n ast.Node) bool {
			if n == nil {
				return true
			}
			s := Site{File: p, Line: fs.Position(n.Pos()).Line, Start: fs.Position(n.Pos()).Offset, End: fs.Position(n.End()).Offset}
			switch v := n.(type) {
			case *ast.CompositeLit:
				t, ok := v.Type.(*ast.Ident)
				if !ok || (t.Name != "Refused" && t.Name != "NotYet") {
					return true
				}
				s.Kind = t.Name
				for _, el := range v.Elts {
					k := el.(*ast.KeyValueExpr)
					switch show(k.Key) {
					case "What":
						s.What = show(k.Value)
					case "Where":
						s.Where = show(k.Value)
					case "Fix":
						s.Fix = show(k.Value)
						s.FixStart = fs.Position(k.Value.Pos()).Offset
						s.FixEnd = fs.Position(k.Value.End()).Offset
					}
				}
			case *ast.CallExpr:
				t, ok := v.Fun.(*ast.SelectorExpr)
				if !ok || (show(t.X) != "l" && !strings.HasSuffix(show(t.X), ".l")) || t.Sel.Name != "notYet" {
					return true
				}
				s.Kind = "notYet"
				s.Where = show(v.Args[0])
				s.What = show(v.Args[1])
			default:
				return true
			}
			out = append(out, s)
			return true
		})
	}
	json.NewEncoder(os.Stdout).Encode(out)
}
```
