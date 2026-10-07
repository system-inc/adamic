export function description(subject: string): string { return `${subject} only has a call signature, you should use a function type instead.`; }
export function thisDescription(name: string): string { return "`this` refers to the function type '" + name + "', did you intend to use a generic `this` parameter like `<Self>(this: Self, ...) => Self` instead?"; }
