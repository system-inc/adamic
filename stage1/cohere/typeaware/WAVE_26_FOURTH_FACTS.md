# Constructor expression metadata

`constructor-expression` is an isolated raw syntax question implemented in
`bridge/tsgo/checker/constructor_expression.go` and decoded by
`stage1/cohere/typeaware/constructor_expression.a`. A single switch registration
in facts.go routes it. It does not change the existing binding-structure schema.

The request must name an exact NewExpression by source, UTF-8 byte start/end
(including leading trivia), and kind. Additional question fields and other node
kinds are refused. The existing inspection frame begins with version 1 and the
question name. Payload fields are:

1. A boolean saying whether Expression is present.
2. When present, the callee's raw kind name, byte start, and byte end.

The native decoder validates framing and matches this identity against the
NewExpression's immediate raw syntax children. Absence yields identity zero;
a present callee missing from those children panics. CallExpression callees use
the existing expression role. Parenthesis handling, static method selection,
binding-origin checks, report spans, messages, findings and repairs are decided
by the three native rule files, not this question.

The question uses the unchanged tsgo_inspect C ABI, owned UTF-8 output buffers,
and program-scoped handle registry. Results are copied and freed by the existing
native adapter. Program release invalidates subsequent questions. Go heap memory
is not instrumented by the C sanitizers.

The direct contract test compares metadata to the compiler's actual Expression
pointer for direct, parenthesized and argumentless constructors, including Unicode
trivia. It rejects malformed requests and nonconstructor nodes. A wrong-callee
mutant substitutes the first argument, producing valid frames and a normal native
exit; the independent production Go finding comparison catches changed bytes.
The released-handle probe asks this question before and after release, and the
retained-registry mutant must fail the required panic-70 expectation.
