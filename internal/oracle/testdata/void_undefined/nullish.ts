interface B { readonly n: number }
interface Holder { value: B | null | void }
const fallback: B = { n: 9 };
function pick(value: B | null | void): void {
  const chosen = value ?? fallback;
  console.log(String(chosen.n) + ' ' + String(value?.n) + ' ' + String(value?.n ?? -1));
}
function pickField(holder: Holder): void {
  const chosen = holder.value ?? fallback;
  console.log(String(chosen.n) + ' ' + String(holder.value?.n));
}
pick(undefined);
pick(null);
pick({ n: 1 });
pickField({ value: undefined });
pickField({ value: null });
pickField({ value: { n: 2 } });
