// a-check: refused a void-returning function at the initializer site seen as () => unknown
// A void-typed function seen as returning unknown, then read through the wider type.
const make: () => void = () => 42;
const wider: () => unknown = make;
const result = wider();
console.log(typeof result);
