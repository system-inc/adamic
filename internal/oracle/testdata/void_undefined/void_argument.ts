interface B { readonly n: number }
function f(): void {}
function show(value: B | null | void): void { console.log(typeof value); }
show(f());
