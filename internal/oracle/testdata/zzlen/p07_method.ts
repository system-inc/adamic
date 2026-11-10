class Counter {
    count = 0;
    add(step: number, times: number): number { this.count += step * times; return this.count; }
}
function arity(f: Function): number { return f.length; }
const counter = new Counter();
console.log(String(arity(counter.add)));
