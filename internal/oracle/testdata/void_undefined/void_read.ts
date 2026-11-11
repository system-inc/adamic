const make: () => void = () => 42;
const value = make();
console.log(String(value === undefined));
