# V4 checked bind core

Bind captures the checking adapter, receiver and prefix values in owning cells. All operands evaluate before the prefix is checked against the underlying producer's leading parameter domains. Bind does not execute the producer. The returned counted closure forwards its captured prefix and later arguments through the retained adapter, preserving the actual argument count and the argument/result checks. Bound reflection reads the original length and name. A fresh native bound name is owned.

Bound call/apply and rebinding refuse with path and the exact fixes `call the bound function directly` and `bind once`. Typed aliases and helper arguments are covered. Bound wrappers get no synthetic producer signature: `.ts` and `.a` checked re-views remain refused with path and fix.

Setup completed in 37.529s (build 37.317s, Node 0.021s, Go 0.025s, submodules 0.073s, markdown dependencies 0.082s, clang 0.184s). nproc=5; cpu.max=400000 100000.

Every command sources /workspace/adamic-tools/env.sh and has outer timeout 90. Go tests use -count=1 -timeout 90s. Raw logs are adjacent.

| Command scope | Seconds | Outcome |
| --- | ---: | --- |
| internal/oracle -run 'TestV4EscapeAdapter(Bind|Bound)' -v | 2.535 | 16 leaves passed, Node and both backends, native ASan/UBSan/LeakSanitizer |
| internal/oracle -run '^TestV4' -v | 17.855 | V4 regressions passed |
| internal/ir -run '^TestCallTargetReaders$' | 22.512 | passed |
| internal/lower -run 'View|Closure|Arguments' | 5.487 | passed |
| internal/native -run 'View|Closure|Arguments' | 10.795 | passed; focused replacement for the previous broader native timeout |
| internal/javascript -run 'View|Closure|Arguments' | 1.049 | passed |

Bind leaves: Compatible 0.77s, Prefix 1.46s, RemainingArgument 1.30s, Result 1.46s, ReceiverAndCapture 0.93s, CountAndSurface 0.95s, Boxed 0.76s, AllArguments 0.84s, NameOwnership 0.81s. Refusal leaves: Call 0.34s, Apply 0.30s, Bind 0.25s, AliasCall 0.27s, HelperApply 0.26s, HelperBind 0.25s, View 0.52s.

Mutants: removing the factory prefix check allows the invalid bound prefix to reach Node's next line without ever running the producer. Removing the remaining-argument domain check allows producer output that the checked program must not reach. Removing the result domain check prints Node's invalid result. All three mutants were caught in native, ASan/UBSan/LeakSanitizer native and JavaScript, and each mutated program exits zero with Node's output. The source mutant removing the native bound-name borrow exclusion fails TestV4EscapeAdapterBindNameOwnership with ASan heap-use-after-free (0.438s); the runner restores the exact compiler bytes.

No persistent fixture or counts.md row was added or changed. Temporary positive controls have allocations/frees/retains/releases/peak/regions: Compatible 11/11/26/29/9/0; ReceiverAndCapture 17/17/34/41/13/0; CountAndSurface 16/16/48/56/10/0; Boxed 10/10/19/21/10/0; AllArguments 10/10/23/23/10/0; NameOwnership 7/7/12/13/7/0.

This lands bind toward V4 precision 3. New through an adapter, actual later call-site blame, generic relations and lane 5's remaining negatives are not part of this commit. Runtime pool/statics and region pending conditions are unchanged. No area/runtime dependency was imported.
