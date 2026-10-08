export function memoize(callback) {
    let value;
    let pending = callback;
    return () => {
        if (pending) {
            value = pending();
            pending = undefined;
        }
        return value;
    };
}
let calls = 0;
const cached = memoize(() => { calls += 1; return 42; });
const first = cached();
const second = cached();
console.log(`${first} ${second} ${calls}`);
