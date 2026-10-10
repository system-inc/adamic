let value: unknown = new TypeError(`initial${1}`);
function change(): void { value = new Error(`replacement${2}`); }
function requireError(error: Error): void { console.log(error.name); }
function observe(error: unknown): void { console.log(typeof error); }
if (value instanceof TypeError) {
    change();
    requireError(value);
    observe(value);
}
console.log('continued');
