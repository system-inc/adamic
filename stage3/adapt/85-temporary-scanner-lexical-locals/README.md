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

Validated: 23 declarations, AST guards pass. Complete 84-86 baseline: 106,367
passing, zero failures/pending/differences (252.427 seconds). Node tokens match.
Putting text's var declaration in a nested block is caught by the scope guard.
A standalone function-body var is refused; its let control compiles natively
and matches Node's output 1. No native scanner ownership result is claimed.
The actual full-tree syntax traversal next refuses numeric regexp lookup keys;
var lowering would be encountered later, after the earlier Debug Error-value
lowering blocker. The standalone probe establishes this known local limitation.
