# Host compiler content audit

Base: `031a1259bc7973934792dc6cb1bd4074fc2204b9`. All observations below used a binary built from this unchanged main. Each source fixture was extracted byte for byte from its source tip. Native builds use `--sanitize`, `ASAN_OPTIONS=detect_leaks=1:halt_on_error=1` and `UBSAN_OPTIONS=halt_on_error=1`; JavaScript runs through `oracle/node.mjs`.

| Topic | Source tip / containing branch | Main content | Fixture on main | Both backend observations |
| --- | --- | --- | --- | --- |
| non-null | `a02613ef` / origin/codex/host-proof-combined, origin/codex/non-null-check, origin/codex/non-null-narrowed-number | partial | `non_null_deinitialize_initialized.a` | /workspace/adamic/.host-commits-proof/non_null_deinitialize_initialized.a:2:8: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| non-null-narrowed | `3a6c8231` / origin/codex/host-proof-combined, origin/codex/non-null-narrowed-number | partial | `non_null_storage_local.a` | /workspace/adamic/.host-commits-proof/non_null_storage_local.a:1:21: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| nested-empty | `b3578751` / origin/codex/array-literal-never-element, origin/codex/host-proof-combined | complete | `array_literal_empty_first.a` | Compiles, agrees with fresh Node |
| concat | `998fb3eb` / origin/codex/host-proof-combined, origin/codex/scanner-expressions | partial | `scanner_expressions/primitive_concatenation.a` | /workspace/adamic/.host-commits-proof/scanner_expressions/primitive_concatenation.a:3:13: stage 0 can't lower a BinaryExpression with a string and a value yet |
| phantom | `d90994da` / origin/codex/host-proof-combined, origin/codex/phantom-brands | partial | `phantom_overload_results.a` | /workspace/adamic/.host-commits-proof/phantom_overload_results.a:8:14: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| memoize | `3550dbc4` / origin/codex/host-proof-combined, origin/codex/memoize-regions | none | `memoize_regions/host14.a` | /workspace/adamic/.host-commits-proof/memoize_regions/host14.a:3:21: Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable) |
| predicates | `a58ba402` / origin/codex/host-proof-combined, origin/codex/proven-predicates | partial | `proven_guards.a` | Compiles, agrees with fresh Node |
| generic-empty | `4e022aae` / origin/codex/host-proof-combined, origin/codex/parser-generic-empty-array | partial | `generic_empty_array.a` | /workspace/adamic/.host-commits-proof/generic_empty_array.a:33:22: Adamic 0.1 refuses a value of type never seen as T, a type parameter whose constraint { text: string; } can be written, so it can write what never can't hold; take it as never, or constrain T to something readonly, which can't write (adamic/invariant-mutable) |
| debugger | `7d2cbc89` / origin/codex/debugger-statement, origin/codex/host-proof-combined | complete | `debugger_fail.a` | Compiles, agrees with fresh Node |
| generic-value | `5d1b45e18` / origin/codex/generic-function-value | none | `generic_function_value.a` | /workspace/adamic/.host-commits-proof/generic_function_value.a:4:47: stage 0 can't lower a generic function as a value yet |

Main contains contextual nested-empty literal lowering and debugger IR/native no-op/JavaScript preservation. Concatenation already spells numbers and booleans but omits null, undefined and optional scalars. Generic empty-literal specialization exists, but bottom-to-readonly fallback proofs are missing. Proven guards and overload marker predicates exist, but the source topic additionally fixes computed-name and failed-generic cleanup panics. Phantom-brand overload admission must retain main's stronger overload checks, generic mapping and argument-count ABI. Generic function values are still explicitly NotYet.

## Dependencies and scope

The incoming non-null `.a` forms conflict with main's explicit refusal policy and its eager checked `.ts` semantics. Lazy placeholders/deinitialization require the absent placeholder-nonnull landing; they remain out. Narrowing/scalar checks already exist on main for checked TypeScript input. No syntax policy is weakened in this batch.

Memoize requires runtime graph-region capture-cell/environment ownership (runtime `d70bc6b1`, carried by source `3550dbc4`). Main has step-04 closure frames, not GraphCell/GraphClosure or synchronous cycle graph ownership. That runtime dependency stays out.

Views beyond main slice 1, optional-presence landings and generic nullable-array views carried by the generic-value topic remain out. Only contextually specialized generic function values and their identity guards are candidates. No old topic base is merged.

## Combined host input on main

The 25 sources and recorded Node observations are taken unchanged from `d71dfdfa`. `observe-disagreements.py --all --compiler /tmp/host-commits/main-adamic --logs /tmp/host-commits/host-main --report /tmp/host-commits/host-main.json` completed, exit 1 for the recorded stops. Fresh Node matches every recorded observation. Both backends compile and agree on exactly 05_writeFile, 11_setModifiedTime, 13_createDirectory and 24_useCaseSensitiveFileNames (4/25). The library proof's 24/25 is not a main observation. No stage-3 status record or Node observation is changed.
