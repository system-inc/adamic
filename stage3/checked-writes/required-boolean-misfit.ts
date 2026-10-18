const original: { readonly multiLine: true } = { multiLine: true };
function store(view: { multiLine: boolean | undefined }, value: boolean | undefined): void { view.multiLine = value; }
store(original, undefined);
console.log('stored');
