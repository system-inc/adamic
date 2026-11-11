interface B { readonly n: number }
type MaybeB = B | null;
function written(value: (B | null) | void): void { console.log(String(value === null) + ' ' + String(value === undefined) + ' ' + typeof value); }
function named(value: MaybeB | void): void { console.log(String(value === null) + ' ' + String(value === undefined) + ' ' + typeof value); }
written(null); written(undefined); written({ n: 1 });
named(null); named(undefined); named({ n: 2 });
