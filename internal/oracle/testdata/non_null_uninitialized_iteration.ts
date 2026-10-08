const reads: (() => number)[] = [];
for (let value: number = undefined!, index = 0; index < 2; index++) {
    reads.push(() => value);
}
console.log(`${reads[1]!()}`);
