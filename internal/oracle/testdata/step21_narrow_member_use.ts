// A TypeError subclass supplies a member that an ordinary Error does not have.
class DetailedTypeError extends TypeError { readonly detail: number = 21; }
let value: unknown = new DetailedTypeError(`initial${1}`);
function change(): void { value = new Error(`replacement${2}`); }
if (value instanceof DetailedTypeError) {
    change();
    console.log(`${value.detail}`);
}
console.log('continued');
