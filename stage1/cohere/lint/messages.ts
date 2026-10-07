// Descriptions copied from the pinned cohere rules.

export const messageVars =
    "All 'var' declarations must be at the top of the function scope. A `var` is hoisted to the top of its function whatever line it is written on, so a declaration buried in a loop or a branch reads as scoped to that block and is not. Writing them together at the top makes the text say what the engine already does.";
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
