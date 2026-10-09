enum Kind { A = 4, B = 9 }
type A = { readonly kind: Kind.A; readonly n: number };
type B = { readonly kind: Kind.B; readonly n: number };
function cast(value: A | B): A { return value as A; }
try {
	console.log('casting');
	console.log(`${cast({ kind: Kind.B, n: 2 }).n}`);
} catch { console.log('caught'); } finally { console.log('finally'); }
