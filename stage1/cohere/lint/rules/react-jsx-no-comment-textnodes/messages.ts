// Exact descriptions from cohere 715ba94.
export const messageJsxNoCommentTextnodes =
    "This looks like a comment but it is not one. Inside a tag's children, `//` and `/*` are ordinary text, so React renders them to the page and the words the author meant only for other developers become part of the interface. The failure is silent: nothing errors, nothing warns, and the text shows up in the product. Wrap it in braces as `{/* ... */}`, which is the form React reads as a comment and drops.";
