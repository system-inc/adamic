| File | Status | Reason/code |
| --- | --- | --- |
| internal/lower/testdata/generic_body_relations/refused.a | refused | generic body initializer into slot T["value"] from source 2 is not proven for every allowed type argument; checker constraint: T extends { readonly value: number; }; read a value of the same dependent type from the type parameter, or use the concrete constraint type for the slot (adamic/generic-body-relations) |
| internal/oracle/testdata/generators/basic.a | checked |  |
| internal/oracle/testdata/generators/cancellation.a | checked |  |
| internal/oracle/testdata/generators/captures.a | checked |  |
| internal/oracle/testdata/generators/checker.a | checked |  |
| internal/oracle/testdata/generators/close.a | checked |  |
| internal/oracle/testdata/generators/completions.a | checked |  |
| internal/oracle/testdata/generators/default_error.a | checked |  |
| internal/oracle/testdata/generators/defaults.a | checked |  |
| internal/oracle/testdata/generators/delegate.a | checked |  |
| internal/oracle/testdata/generators/delegate_throw.a | checked |  |
| internal/oracle/testdata/generators/first_reference.a | checked |  |
| internal/oracle/testdata/generators/map_iterator.a | checked |  |
| internal/oracle/testdata/generators/owned.a | checked |  |
| internal/oracle/testdata/generators/reentrant.a | checked |  |
| internal/oracle/testdata/generators/throw.a | checked |  |
| internal/oracle/testdata/generic_body_indexed.a | checked |  |
| internal/oracle/testdata/generic_body_read.a | checked |  |
| internal/oracle/testdata/notyet/generator_cycle.a | refused | cycle-capable generator frame slot saved without weak; keep the back-reference weak, or remove the owning path back to this generator |
| stage3/fixtures/generics/01_identity.a | checked |  |
| stage3/fixtures/generics/02_optional_return.a | checked |  |
| stage3/fixtures/generics/03_callback_return.a | checked |  |
| stage3/fixtures/generics/04_function_value.a | NotYet | adamic: /workspace/adamic/stage3/fixtures/generics/04_function_value.a:3:60: stage 0 can't lower a generic function as a value yet |
| stage3/fixtures/generics/05_nested_class.a | checked |  |
| stage3/fixtures/generics/06_polymorphic_recursion.a | refused | a generic function instantiated without end (polymorphic recursion); call it with the same type arguments it was called with, or write a function per type |
| stage3/fixtures/generics/07_generic_cast.a | refused | a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| stage3/fixtures/generics/08_readonly_view.a | refused | a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| stage3/fixtures/generics/09_recursive_optional.a | checked |  |
| stage3/fixtures/generics/10_identifier_multimap.a | checked |  |
| stage3/fixtures/generics/11_indexed_result.a | refused | generic body return into slot T["value"] from source 2 is not proven for every allowed type argument; checker constraint: T extends { readonly value: number; }; read a value of the same dependent type from the type parameter, or use the concrete constraint type for the slot (adamic/generic-body-relations) |
| stage3/fixtures/generics/12_mixed_indexed_result.a | refused | generic body return into slot U \| T["value"] \| undefined from source 2 is not proven for every allowed type argument; checker constraint: T extends { readonly value: number; }; U (unconstrained); read a value of the same dependent type from the type parameter, or use the concrete constraint type for the slot (adamic/generic-body-relations) |
| stage3/fixtures/generics/13_constrained_local.a | refused | generic body initializer into slot T["value"] from source 2 is not proven for every allowed type argument; checker constraint: T extends { readonly value: number; }; U (unconstrained); read a value of the same dependent type from the type parameter, or use the concrete constraint type for the slot (adamic/generic-body-relations) |
| stage3/fixtures/generics/14_optional_literal_union.a | checked |  |
| stage3/fixtures/generics/15_array_callback.a | checked |  |
| stage3/fixtures/iteration-dispatch/body_throw.a | checked |  |
| stage3/fixtures/iteration-dispatch/break.a | checked |  |
| stage3/fixtures/iteration-dispatch/cached_next.a | checked |  |
| stage3/fixtures/iteration-dispatch/class.a | checked |  |
| stage3/fixtures/iteration-dispatch/close_throw_body.a | checked |  |
| stage3/fixtures/iteration-dispatch/close_throw_break.a | checked |  |
| stage3/fixtures/iteration-dispatch/completion.a | checked |  |
| stage3/fixtures/iteration-dispatch/done_throw.a | checked |  |
| stage3/fixtures/iteration-dispatch/labeled_continue.a | checked |  |
| stage3/fixtures/iteration-dispatch/method_getters.a | checked |  |
| stage3/fixtures/iteration-dispatch/next_throw.a | checked |  |
| stage3/fixtures/iteration-dispatch/non_callable_next.a | refused | a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| stage3/fixtures/iteration-dispatch/non_object_close.a | refused | a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| stage3/fixtures/iteration-dispatch/non_object_iterator.a | refused | a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| stage3/fixtures/iteration-dispatch/non_object_step.a | refused | a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| stage3/fixtures/iteration-dispatch/object.a | checked |  |
| stage3/fixtures/iteration-dispatch/return.a | checked |  |
| stage3/fixtures/iteration-dispatch/value_throw.a | checked |  |
| stage3/fixtures/iteration/accessor_binding.a | NotYet | adamic: /workspace/adamic/stage3/fixtures/iteration/accessor_binding.a:24:10: stage 0 can't lower a for...of destructuring an object yet |
| stage3/fixtures/iteration/array_view_stress.a | checked |  |
| stage3/fixtures/iteration/array_view_variance.a | refused | a value of type Slot<Large>[] seen as NodeArray<Slot<Small>>, which can write Small where Large is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| stage3/fixtures/iteration/array_view_weak.a | checked |  |
| stage3/fixtures/iteration/collections.a | checked |  |
| stage3/fixtures/iteration/delegated_generator.a | checked |  |
| stage3/fixtures/iteration/generator.a | checked |  |
| stage3/fixtures/iteration/intrinsic_iterator_alias.a | refused | writing Symbol.iterator on a built-in or its prototype (adamic/intrinsic-iterator); use a separate object with its own Symbol.iterator method; built-in iteration cannot be overridden |
| stage3/fixtures/iteration/intrinsic_iterator_prototype.a | refused | writing Symbol.iterator on a built-in or its prototype (adamic/intrinsic-iterator); use a separate object with its own Symbol.iterator method; built-in iteration cannot be overridden |
| stage3/fixtures/iteration/intrinsic_iterator_write.a | refused | writing Symbol.iterator on a built-in or its prototype (adamic/intrinsic-iterator); use a separate object with its own Symbol.iterator method; built-in iteration cannot be overridden |
| stage3/fixtures/iteration/iterable_view.a | NotYet | adamic: /workspace/adamic/stage3/fixtures/iteration/iterable_view.a:4:25: stage 0 can't lower an iterable object view of built-in storage needing a protocol adapter yet |
| stage3/fixtures/iteration/object_binding.a | NotYet | adamic: /workspace/adamic/stage3/fixtures/iteration/object_binding.a:3:6: stage 0 can't lower a for...of destructuring an object yet |
| stage3/fixtures/iteration/object_binding_stress.a | NotYet | adamic: /workspace/adamic/stage3/fixtures/iteration/object_binding_stress.a:42:10: stage 0 can't lower a for...of destructuring an object yet |
| stage3/fixtures/iteration/object_iteration.a | checked |  |
| stage3/fixtures/iteration/strings.a | checked |  |
| stage3/fixtures/iteration/test262_array_views.a | checked |  |
| stage3/fixtures/iteration/user_forwarding.a | checked |  |
| stage3/project-loader/fixture/app/main.a | checked |  |
| stage3/project-loader/fixture/dependency/value.a | checked |  |
| stage3/project-references-source/fixture/chain/app/main-print.a | checked |  |
| stage3/project-references-source/fixture/chain/app/main.a | checked |  |
| stage3/project-references-source/fixture/chain/leaf/base.a | checked |  |
| stage3/project-references-source/fixture/chain/leaf/unused.a | checked |  |
| stage3/project-references-source/fixture/chain/middle/value.a | checked |  |
| stage3/project-references-source/fixture/two/app/main.a | checked |  |
| stage3/project-references-source/fixture/two/dependency/value.a | checked |  |
| stage3/project-references-source/fixture/types/app/main.a | checked |  |
| stage3/project-references-source/fixture/types/dependency/value.a | checked |  |
| stage3/project-references-source/fixture/types/packages/dependency.a | checked |  |
| stage3/project-references-source/fixture/types/packages/entry.a | checked |  |
