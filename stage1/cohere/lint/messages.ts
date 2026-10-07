// Descriptions copied from the pinned cohere rules.

export const messageContinue =
    'This loop uses `continue`, which jumps to the next iteration from the middle of the body. A reader tracing what happens on one pass has to hold every `continue` above their position in mind, because any of them may have skipped the code they are looking at, and in a `for` loop the update expression still runs while the rest of the body does not. Inverting the condition and wrapping the remainder in an `if` says the same thing with the skip visible in the shape of the code rather than in a jump.';

export const messageWith =
    "A `with` block splices an object's properties into the local scope, so a bare name inside it cannot be resolved by reading the code: whether `x` means the object's property or an outer variable is decided at runtime by what the object happens to hold. That defeats the checker and every reader. Bind what you need to a name instead, as in `const { x, y } = point;`.";

export const messageNew =
    "This constructs an object and then throws it away, so the only thing the line can accomplish is whatever the constructor does on the side. A constructor that works by side effect is doing a function's job under a name that promises a value, and the next reader cannot tell the discarded result from a mistake. Call a function instead, or keep the object and use it.";

export const messageSparse =
    'This array literal has a hole: two commas with nothing between them. A hole is not undefined, it is an absent index, and the difference is visible exactly where it is least expected. `map` and `forEach` skip holes while `for...of` and spread produce undefined for them, so the same array reads as two different lengths of data depending on how it is walked. Almost every hole is a typo for one comma. If the absence is deliberate, write undefined and say so.';

export const messageYield =
    "This generator function contains no `yield`, so calling it returns an iterator that finishes immediately with the function's return value and never produces anything. A caller writing `for(const item of generate())` gets zero iterations. Either the asterisk is left over from a refactor and the function should be a plain one, or a `yield` is missing.";

export const messageAwait =
    'This waits inside a loop, so each iteration blocks on the one before it and the work runs one at a time. Where the iterations do not depend on each other, collecting the promises and awaiting them together turns a sum of latencies into a maximum. Where they do depend on each other, the sequencing is deliberate and the loop is the right shape, which is why this reports rather than repairs.';

export const messageVars =
    "All 'var' declarations must be at the top of the function scope. A `var` is hoisted to the top of its function whatever line it is written on, so a declaration buried in a loop or a branch reads as scoped to that block and is not. Writing them together at the top makes the text say what the engine already does.";
export const messageTemplateCurly =
    'This ordinary string contains a template placeholder, so the placeholder is not a placeholder: it is four literal characters that will be sent, rendered, or stored exactly as written. The failure is silent, because the string is a perfectly valid string and nothing at runtime objects to it. Backticks are almost always what was meant.';
export const messageDivRegex =
    'This regular expression begins with an equals sign, so its first two characters are the divide-assign operator spelled backwards, and a reader scanning the line has to work out which one it is. Writing the equals sign as a one-member character class matches exactly the same input and can only be read one way.';
export const messageBitwise =
    'A bitwise operator here is usually a typo for its logical twin: `&` for `&&`, `|` for `||`. When it is deliberate the intent is worth stating, because the reader has to decide which of the two you meant every time they pass it.';
export const messageUnexpectedLabel =
    'This labels a statement. A label is a second naming system that only `break` and `continue` can read, so control can leave from a line that carries no visible marker and land somewhere the reader has to search the file to find. Reading the code top to bottom no longer tells you where it goes. Extract the labeled region into a function and return from it, or restructure the loops so the exit is local.';
export const messageUnexpectedLabelInBreak =
    'This `break` names a label, so it exits an enclosing block or loop rather than the nearest one. The jump target is not on this line and not adjacent to it, which is what makes the reading cost real: the reader has to find the label before they know what was left. Return from an extracted function, or use a flag the loops already test.';
export const messageUnexpectedLabelInContinue =
    'This `continue` names a label, so it resumes an outer loop rather than the one it sits in. Which loop advances is decided by a name written elsewhere, so the iteration a reader traces is not the iteration that runs. Restructure so the inner work is a function that returns, or move the condition out to the loop it belongs to.';
export const messageUnicodeBomExpected =
    'This file is configured to start with a byte order mark and does not. The mark is what tells a reader that has no other signal which encoding and endianness the bytes are in, so a project that requires it wants every file to carry it rather than most.';
export const messageUnicodeBomUnexpected =
    'This file starts with a byte order mark. In a project that is already UTF-8 everywhere the mark carries no information and is invisible in an editor, while tools that read the first bytes literally see it: a shell script stops being executable, a concatenated bundle gains a stray character mid-file, and a diff shows a change on a line nobody edited.';
export const messageUnexpectedCommaExpression =
    'This uses the comma operator, which evaluates the expression on its left, throws the result away, and yields the expression on its right. Two statements are hiding inside one, and the discarded half is the part a reader skips: `a = b(), c()` calls both and assigns only the second, which reads as a typo for `a = b()`. Split it into separate statements, or wrap it in its own parentheses to say the sequencing was deliberate.';
export const messageNoUnneededTernaryConditionalExpression =
    'This conditional returns a boolean literal on both branches, so the whole ternary is a long way of writing the condition itself. The branches carry no information the test does not already have, and a reader has to check both to learn that. Write the condition, negated if the branches are the other way round.';
export const messageNoUnneededTernaryConditionalAssignment =
    'This conditional tests a value and then returns that same value, which is the pre-`||` way of spelling a default. `a ? a : b` evaluates `a` twice and reads as though the two branches might differ. Write `a || b`, which says the same thing once.';
