// Exact descriptions from cohere 715ba94.
export const messageUnexpectedMultilineDivision =
    'A newline sits between the numerator and the `/` that follows it, and what comes after reads as a regular expression with flags rather than a division. The two parse differently and only one of them is what the line above looks like.';
export const messageUnexpectedMultilineFunction =
    'A newline sits between this expression and the `(` that follows it, so what reads as two statements is one call. Automatic semicolon insertion does not fire before an open paren, so the line above is being invoked rather than ended.';
export const messageUnexpectedMultilineProperty =
    'A newline sits between this expression and the `[` that follows it, so what reads as an array on its own line is a computed property access on the line above. Automatic semicolon insertion does not fire before an open bracket.';
export const messageUnexpectedMultilineTaggedTemplate =
    'A newline sits between this expression and the template literal that follows it, so the template is being used as a tag argument rather than standing alone. Automatic semicolon insertion does not fire before a backtick.';
