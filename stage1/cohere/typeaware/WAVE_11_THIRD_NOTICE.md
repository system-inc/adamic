# Third batch provenance

no_eval.a, no_extend_native.a and no_func_assign.a port the corresponding
production Go rules in cohere/internal/lint/rules/core at
715ba94f3608a6500086b1076ce5cb7e51b836db. Rule messages and builtin names preserve
that source byte for byte. Cohere is distributed under its existing MIT license;
see the pinned submodule license. No production Go oracle source was modified.
The new Go oracle and validation file invoke those production rules independently.
