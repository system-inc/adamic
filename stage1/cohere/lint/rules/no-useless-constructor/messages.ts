// Exact descriptions from cohere 715ba94.
export const messageNoUselessConstructor =
    'This constructor does nothing the language would not do without it. An empty constructor on a base class, or one whose only statement forwards its own parameters to `super`, is exactly the behaviour a class gets when it declares no constructor at all, so the code reads as though something happens at construction time when nothing does.';
export const messageNoUselessConstructorRemove = 'Remove the constructor, leaving the behaviour the class already had.';
