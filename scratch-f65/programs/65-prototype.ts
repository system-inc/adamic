const copy = { ...{ k: 1 } };
console.log(String("toString" in copy));
console.log(String("constructor" in copy));
console.log(String("missing" in copy));
