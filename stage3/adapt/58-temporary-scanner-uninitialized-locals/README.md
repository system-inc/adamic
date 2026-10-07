Temporary: comes out when 01a1130a's lazy initializer check lands

Plan published before implementation. Slice scanner.ts only: the text declaration
becomes var text: string; its initial setText call still receives textInitial.
Nothing reads text before setText assigns newText || "". Replace tokenValue's
definite-assignment marker with its plain typed, uninitialized declaration.
Probe the same form for the nested operand declaration if it is the next refusal.
No eager value is invented. Preserve all scanner statements and observable Node
undefined values. Record read-before-assignment evidence, including the driver's
call contract, rather than assuming every exported getter is safe at every time.

Validate byte-identical Node tokens and the full upstream baseline. Compiler
probes decide which remaining refusal is next. This is a temporary source
spelling change, separate from the verbatim declaration gatherer.
