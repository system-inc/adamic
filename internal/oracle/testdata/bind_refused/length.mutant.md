Mutant evidence for TestFunctionBindLengthFunctionSlot's rows (#0ty4971).

Each length_*.ts program reads a bound function's .length through a Function-typed value: passed as
a parameter (length_function_slot.ts), returned (length_returned.ts), stored in a field
(length_field.ts) and in an array (length_array.ts), a bound method (length_method.ts), and a bound
arrow typed wider than it is (length_typed_arrow.ts). A bound closure has no length of its own yet,
so native would read 0 where Node gives the target's length less the fixed arguments. Each stops
at graph_types.go's graph adoption boundary, and the test pins that stop at the read's line with
Node's answer beside it.

The mutant: in internal/lower/graph_types.go, under `case ir.DynamicProperty, ir.DynamicKeyedProperty:`
in graphAdoptionBoundary, the two lines that set reason and site commented out. All six rows go red,
each with "want existing graph/dynamic stop at Function.length, got <nil>": every program lowers, so
the boundary is the only stop between these roads and the missing length. Codex session
01a1c4e2-29c6-73e5-a685-9e15a562f02d (go 1.27.1, clang 20.1.8, node 24.19) ran it at 97d2cc69, its
log at emit/bind-length-pins-log 3e0c2a20 (proof/pins.log). Unmutated, all six rows and
TestFunctionBindKeyedLength's four passed in the same session.
