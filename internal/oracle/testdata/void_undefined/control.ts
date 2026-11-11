interface B { readonly n: number }
function show(value: B | undefined | null): void { console.log(String(value === null) + ' ' + String(value === undefined) + ' ' + String(value === null || value === undefined) + ' ' + typeof value); }
show(null); show(undefined); show({ n: 1 });
