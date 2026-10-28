// Spec: A `void` value is unobservable instead
// Expect: refused a read of a void value
// The same read in .ts is refused too: no check can recover a value that was discarded.
const make: () => void = () => 42;
const value = make();
console.log(`${value === undefined}`);
