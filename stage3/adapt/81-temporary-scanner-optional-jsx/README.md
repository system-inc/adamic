Temporary: comes out when function parameter widening lands

Plan published before implementation. Slice scanner.ts only: annotate the nested
reScanJsxToken default parameter as allowMultilineJsxText: boolean | undefined = true.
The default remains true: omitted and explicit undefined values take that default.
Only the erased annotation changes. This answers method-signature-style's
Boolean-to-optional-Boolean refusal while preserving the original initializer.

Validate Node equality and the complete upstream baseline. Keep the same
statements and preserve omitted, false and true argument behavior.

Validated with 59/80-83: 106,367 upstream tests pass with zero failures/pending/
differences, 230.719 seconds. All Node tokens match. The default initializer is
still true; only the explicit union annotation changes. Next compiler refusal:
scanner:558:9, generic callback () => T seen as () => T (method-signature-style).
