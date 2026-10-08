Use checker-narrowed numeric reads and fit statement updates back to the declared slot.
Commits: codex/generic-function-value, 2026-10-08 UTC; host scratch integration follows.
Checks: uncached Node/native/JavaScript oracle 0.466s; full lower 24.224s; vet clean; counts alloc/free 9/9.
Mutant: use addition for optional-number decrements; both backends exit normally with 5,6,7,8 instead of Node 3,4,3,4.
Uncovered: consumed prefix/postfix expression updates and property updates keep their existing paths; complete host oracle pending.

Fixture 25's depth-- previously emitted subtraction from adamic_maybe_number itself. Lower the operand through the existing expression path, so checker narrowing yields a numeric read and preserves any inserted checks; fit the result back to the declared optional or boxed union slot. Unnarrowed operands stay NotYet. No native emitter edit or new representation.

narrowed_number_increment.a covers ++ and -- as statements, prefix and postfix, present/absent optional input, and a number narrowed from string | number. Counts retain/release 4/13, peak 3, regions 0. Logs /tmp/narrowed-increment-{oracle,operator-mutant,counts,lower,vet}.log. The pinned host fixture is unchanged and its complete oracle is rerunning with the repair.
