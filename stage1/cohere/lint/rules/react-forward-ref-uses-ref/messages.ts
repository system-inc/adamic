// Exact descriptions from cohere 715ba94.
export const messageForwardRefAddRefParameter = 'Add a `ref` parameter.';
export const messageForwardRefRemoveWrapper = 'Remove the `forwardRef` wrapper.';
export const messageForwardRefUsesRef =
    'A function wrapped in `forwardRef` takes only one parameter, so the ref React passes as the second one is dropped on the floor. Every consumer writing `<Component ref={r} />` gets a ref that never attaches, and it fails silently rather than erroring. Accept the second parameter and forward it to whatever should receive it, or drop the `forwardRef` wrapper, which does nothing for a component that ignores the ref.';
