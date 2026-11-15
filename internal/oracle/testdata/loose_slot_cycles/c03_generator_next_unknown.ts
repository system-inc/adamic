function* pairs(count: number): Generator<[string, number], void, unknown> {
    for (let index = 0; index < count; index++) {
        const pair: [string, number] = [`k${String(index % 2)}`, index];
        yield pair;
    }
}
const map = new Map(pairs(3));
console.log(String(map.size));
console.log([...map.keys()].join(","));
