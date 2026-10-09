class Counter {
    value = 0;
}
const counter = new Counter();
const before = counter.value++;
console.log(`${before}`);
console.log(`${counter.value}`);
