interface B { readonly n: number }
type M = (B | null) | void;
interface Holder { value: M }
function report(holder: Holder): void {
  const value = holder.value;
  let n = 'none';
  if (value !== null && value !== undefined) {
    n = String(value.n);
  }
  console.log(String(value === null) + ' ' + String(value === undefined) + ' ' + typeof value + ' ' + n);
}
function reportInPlace(holder: Holder): void {
  console.log(String(holder.value === null) + ' ' + String(holder.value === undefined) + ' ' + typeof holder.value);
}
const holder: Holder = { value: undefined };
report(holder);
reportInPlace(holder);
holder.value = null;
report(holder);
reportInPlace(holder);
holder.value = { n: 2 };
report(holder);
reportInPlace(holder);
holder.value = undefined;
report(holder);
reportInPlace(holder);
const written: Holder = { value: null };
report(written);
reportInPlace(written);
