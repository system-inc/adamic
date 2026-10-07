export const messageMethodSignatureStyleErrorMethod =
    'This is written as a shorthand method signature, which TypeScript checks bivariantly: a parameter can be narrowed by an implementer and the compiler will not object, so a call that type-checks can still be wrong at runtime. A function property is checked the strict way and catches that.';

export const messageMethodSignatureStyleErrorProperty =
    'This project writes interface members as shorthand methods, and this one is a property holding a function type. The two differ in how the compiler checks parameters, so mixing them in one codebase means a reader cannot tell which rule applies without looking at the punctuation.';

export const messageMethodSignatureStyleConvertToMethod =
    'Convert to a method signature. There is no syntax for a readonly method, so this drops the `readonly` modifier and the member becomes reassignable.';
