interface Holder { b: Set<number>; }
function probe(a: Holder): boolean | undefined {
    return a.b?.has(a.b.size);
}
console.log(`${probe({ b: new Set<number>() })}`);
