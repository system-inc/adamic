interface Plain { readonly a: number; }
const counts: Record<string, number> = { a: 1 };
const real = { a: 1, k: 41 };
function read(o: Plain): string {
  if ('k' in o) { return typeof o.k; }
  return 'absent';
}
const plain: Plain = { a: 2 };
console.log(`${read(real)} ${read(plain)} ${counts.a}`);
