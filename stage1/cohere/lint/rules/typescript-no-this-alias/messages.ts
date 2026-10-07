export const assignment =
    'This binding holds `this` under another name. The alias exists to carry the receiver into a nested `function`, which rebinds `this` to something else, and an arrow function closes over `this` directly and needs no carrier. Every later reader now has to establish which of the two names is the real receiver and whether they are still the same object. Use an arrow function and write `this`.';
export const destructure =
    'This pattern pulls members off `this` into local bindings. Each one is read once, at the moment of destructuring, so a later write to the property is invisible here, and a method taken this way has lost its receiver and throws when called. Read through `this` at the point of use instead.';
