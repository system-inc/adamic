const xs: [number, number] = [1, 2];
const two = 2 as const;
xs.length = two;
console.log(String(xs.length));
