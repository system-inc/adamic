Temporary: comes out when function scoped var lowering lands

Plan published before implementation. Slice scanner.ts only: replace its 23
function-body-level var declaration keywords with let. All belong directly to
createScanner or scanRegularExpressionWorker; none is in a nested block, loop
or repeated declaration. Preserve initializers and declaration positions.

createScanner's first setText happens after every state declaration and before
the scanner object; setText references only those already-declared state locals.
The scanner object's methods are invoked after construction. Regex-worker
state declarations all precede its first parsing call; nested function declarations
introduce no eager reads. No lexical binding is accessed within its TDZ.
An AST guard verifies function-body positions and distinct declared names.
This answers the existing local var refusal. Validate Node and full upstream
baselines; do not claim native ownership passes until it is actually checked.
