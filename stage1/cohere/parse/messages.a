// Exact descriptions from cohere 715ba94.
export const messageOctalEscape =
    'This is a legacy octal escape, a numbering the language kept only for old web pages and forbids outright in strict mode and in modules, where it is a syntax error rather than a character. It also reads as decimal to almost everybody: the escape here is not the number it looks like. A unicode escape says the same character in a numbering that is unambiguous and legal everywhere.';
export const messageUnexpectedMultilineFunction =
    'A newline sits between this expression and the `(` that follows it, so what reads as two statements is one call. Automatic semicolon insertion does not fire before an open paren, so the line above is being invoked rather than ended.';
export const messageUnexpectedMultilineProperty =
    'A newline sits between this expression and the `[` that follows it, so what reads as an array on its own line is a computed property access on the line above. Automatic semicolon insertion does not fire before an open bracket.';
export const messageUnexpectedMultilineTaggedTemplate =
    'A newline sits between this expression and the template literal that follows it, so the template is being used as a tag argument rather than standing alone. Automatic semicolon insertion does not fire before a backtick.';
export const messageUnexpectedMultilineDivision =
    'A newline sits between the numerator and the `/` that follows it, and what comes after reads as a regular expression with flags rather than a division. The two parse differently and only one of them is what the line above looks like.';
export const messageNoUselessConstructor =
    'This constructor does nothing the language would not do without it. An empty constructor on a base class, or one whose only statement forwards its own parameters to `super`, is exactly the behaviour a class gets when it declares no constructor at all, so the code reads as though something happens at construction time when nothing does.';
export const messageNoUselessConstructorRemove = 'Remove the constructor, leaving the behaviour the class already had.';
export const messagePreferTemplateUnexpectedStringConcatenation =
    "This builds a string by concatenating a literal with something that is not one. The `+` operator does double duty in JavaScript, so a reader has to work out from the operands whether a line adds or joins, and a single non-string operand silently changes the answer for the whole expression. A template literal says which one is meant in its first character, keeps the literal text contiguous instead of split across quotes and plus signs, and does not change meaning when an operand's type does.";
export const messageForwardRefUsesRef =
    'A function wrapped in `forwardRef` takes only one parameter, so the ref React passes as the second one is dropped on the floor. Every consumer writing `<Component ref={r} />` gets a ref that never attaches, and it fails silently rather than erroring. Accept the second parameter and forward it to whatever should receive it, or drop the `forwardRef` wrapper, which does nothing for a component that ignores the ref.';
export const messageForwardRefAddRefParameter = 'Add a `ref` parameter.';
export const messageForwardRefRemoveWrapper = 'Remove the `forwardRef` wrapper.';
export const messageJsxNoCommentTextnodes =
    "This looks like a comment but it is not one. Inside a tag's children, `//` and `/*` are ordinary text, so React renders them to the page and the words the author meant only for other developers become part of the interface. The failure is silent: nothing errors, nothing warns, and the text shows up in the product. Wrap it in braces as `{/* ... */}`, which is the form React reads as a comment and drops.";
export const messageNoFindDOMNode =
    "`findDOMNode` was deprecated in 2018 and removed in React 19, so this call throws on any current React. Even where it still runs it reaches around the component that owns the node: the parent reads a child's DOM element the child never agreed to expose, so the child cannot change what it renders without breaking a caller it does not know about. Pass a ref to the element instead, so the component decides what it hands out.";
export const messageNoIsMounted =
    '`isMounted` is called on `this`. It was removed from React and never existed on a function component, so the call either throws or silently reads `undefined`. Where it does still run, it is an anti-pattern: it hides a leak rather than fixing one, because the only reason to ask whether a component is still mounted is that something asynchronous outlived it. Cancel that work when the component unmounts instead, so nothing is left holding a reference to ask about.';
