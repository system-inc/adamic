# External runtime call-site handoff

Source locations are from library snapshot 132d9919. All four sites are now carried on codex/closure-convention. node_process.c
locations move to 442/447 after prior typed declarations were inserted; parallel
locations remain 197/300. Both complete translation units compile under -Werror
in plain, counted and counted plus receiver/canonical configurations. Broader
host/shared-heap activation remains outside this compiler's admitted operations.

| File:line | Owner area | Exact replacement line |
| --- | --- | --- |
| node_process.c:436 | Library | `adamic_value result = adamic_closure_call(closure, padded, count);` |
| node_process.c:441 | Library | `return adamic_closure_call(closure, args, count);` |
| parallel.c:197 | Runtime | `adamic_value result = adamic_closure_call(scope->work, arguments, 2);` |
| parallel.c:300 | Runtime | `adamic_value result = adamic_closure_call(work, arguments, 2);` |

The performance builtin buffer has three slots; its observable count remains
`count`, not 3. Both parallel callbacks receive two actual arguments. Their
frontend adapter must pack optional/rest slots using the shared closure layout;
a two-word runtime buffer by itself does not implement arbitrary rest packing.
Builtin definitions and closure initializers must use their matching typed
shape, with no function-pointer cast.

The helper invokes exactly `adamic_code(self, arguments)` or
`adamic_counted_code(self, arguments, count)`. Count is last. Each shape has one
pointer typedef derived from one function typedef, and the counted tag selects
the active union member. No function-pointer cast is involved.

Already integrated:
- Runtime array.c:198: `double result = adamic_closure_call(compare, (adamic_value[]){left, right}, 2).number;`
- Library regexp_replace.c:114: `adamic_value returned = adamic_closure_call(callback, packed, argument_count);`

Whole-runtime searches are preserved in final-runtime-call-audit.log. Actual
indirect invocations are only adamic.h:139/143 (counted/plain closure) and
adamic.h:293/294 (counted/plain method). Object method lookup returns a typed
entry; the nominal virtual table stores numeric identities, not function pointers.
