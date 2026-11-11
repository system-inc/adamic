// The runner's control that must agree on every backend: ordinary code, its output held to Node.
const values: number[] = [3, 1, 2];
values.sort((left: number, right: number): number => left - right);
console.log(values.join(','));
