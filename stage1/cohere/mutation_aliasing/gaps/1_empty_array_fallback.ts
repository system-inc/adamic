// An empty array literal as the fallback of ??, which tsc types as never[].
const lists = new Map<number, readonly number[]>([[1, [4, 5]]]);
let total = 0;
for(const each of lists.get(2) ?? []) {
    total += each;
}
for(const each of lists.get(1) ?? []) {
    total += each;
}
console.log(`${total}`);
