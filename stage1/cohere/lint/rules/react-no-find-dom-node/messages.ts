// Exact descriptions from cohere 715ba94.
export const messageNoFindDOMNode =
    "`findDOMNode` was deprecated in 2018 and removed in React 19, so this call throws on any current React. Even where it still runs it reaches around the component that owns the node: the parent reads a child's DOM element the child never agreed to expose, so the child cannot change what it renders without breaking a caller it does not know about. Pass a ref to the element instead, so the component decides what it hands out.";
