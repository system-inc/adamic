const view: { next(): unknown } = new Set<number>([1]).values();
console.log(`${view}`);
