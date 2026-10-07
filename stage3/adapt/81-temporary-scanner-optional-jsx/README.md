Temporary: comes out when function parameter widening lands

Plan published before implementation. Slice scanner.ts only: make the nested
reScanJsxToken's allowMultilineJsxText parameter optional, matching the Scanner
interface. Its body tests this value for truthiness; omitted values already
behave as false on Node. Change only the parameter's erased type spelling.
This answers method-signature-style's Boolean-to-optional-Boolean refusal.

Validate Node equality and the complete upstream baseline. Keep the same
statements and preserve omitted, false and true argument behavior.
