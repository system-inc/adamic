let first: number | string = 1;
let second: number | string = 'two';
let tuple: [number, string] | [string, number] = ['a', 1];
[first, second] = tuple;
console.log(`${first} ${second}`);
