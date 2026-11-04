class First { readonly seed: number; constructor(seed: number) { this.seed = seed; } }
class Second extends First { readonly doubled = this.seed * 2; }
class Third extends Second { readonly tripled = this.doubled + this.seed; }
function show(value: First): string { return `${value instanceof Third}/${value.seed}`; }
console.log(show(new Third(3)));
console.log(show(new First(1)));
