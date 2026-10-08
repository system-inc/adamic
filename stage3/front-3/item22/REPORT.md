# Debugger integration

Source: codex/debugger-statement 7d2cbc895a1e690c18ceefa908da44618370b769.

Keep existing namespace lowering, proven record deletion and sound overload rules. Remove the debugger syntax refusal; retain ir.Debugger for JavaScript while native emission and flow analysis perform no operation. Both sides' fixtures remain.

Linux counts regeneration passed. The only stage3 change is objects/28_debugger.a: Refused debugger becomes NotYet for its existing unsupported Error throw. All bytes outside stage0, including Node, are unchanged; no Compiles downgrade.

Six source-overlay mutants are caught: native trap by emitted-C and Node exit pins, removed JavaScript emission, dropped lowering, restored refusal, and an added flow instruction. results.json and raw logs record actual intended failures. An initial runner without the sourced toolchain environment failed before verification and is not evidence.

The six-package uncached gate, vet, gofmt and whitespace results are archived here after completion. Attached-debugger interaction and optional native debug builds are not covered.
