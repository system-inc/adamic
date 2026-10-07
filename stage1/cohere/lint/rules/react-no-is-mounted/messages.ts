// Exact descriptions from cohere 715ba94.
export const messageNoIsMounted =
    '`isMounted` is called on `this`. It was removed from React and never existed on a function component, so the call either throws or silently reads `undefined`. Where it does still run, it is an anti-pattern: it hides a leak rather than fixing one, because the only reason to ask whether a component is still mounted is that something asynchronous outlived it. Cancel that work when the component unmounts instead, so nothing is left holding a reference to ask about.';
