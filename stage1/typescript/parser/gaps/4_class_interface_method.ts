interface CounterInterface {
    readonly count: () => number;
}
class Counter {
    count(): number { return 1; }
}
function read(counter: CounterInterface): number { return counter.count(); }
console.log(`${read(new Counter())}`);
