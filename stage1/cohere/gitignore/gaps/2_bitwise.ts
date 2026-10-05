// Gap 2: the bitwise operators, in 0.1 with JavaScript's ToInt32 semantics, are not lowered.
const value = 200;
console.log(`${value >> 3} ${value << 2} ${-value >>> 28} ${value & 63} ${value | 1} ${value ^ 3} ${~value}`);
