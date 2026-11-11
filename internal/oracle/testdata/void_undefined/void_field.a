interface B { readonly n: number }
interface Holder { value: B | null | void }
function f(): void {}
const holder: Holder = { value: f() };
console.log(typeof holder.value);
