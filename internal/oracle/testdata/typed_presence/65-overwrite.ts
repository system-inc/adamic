const copy = { k: 1, ...{ k: undefined } };
console.log(String("k" in copy));
console.log(String(copy.k));
