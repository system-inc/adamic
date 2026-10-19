let value: Error = new TypeError(`initial${1}`);
function change(): void { value = new Error(`replacement${2}`); }
function requireTypeError(error: TypeError): void { console.log(error.name); }
if (value instanceof TypeError) {
    change();
    requireTypeError(value);
}
console.log('continued');
