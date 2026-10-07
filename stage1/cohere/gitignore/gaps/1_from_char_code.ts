// Gap 1: String.fromCharCode and String.fromCodePoint, in 0.1's library, are not lowered.
console.log(String.fromCharCode(104) + String.fromCodePoint(233));
