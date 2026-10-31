// A .ts program is held to the same rule: typeof reads the value, and Node names a number.
const make: () => void = () => 42;
console.log(typeof make());
