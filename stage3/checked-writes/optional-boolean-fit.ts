const original: { readonly multiLine: true | undefined } = { multiLine: true };
function store(view: { multiLine: boolean | undefined }, value: boolean | undefined): void { view.multiLine = value; }
store(original, undefined);
console.log(original.multiLine === undefined ? 'undefined' : original.multiLine.toString());
store(original, true);
console.log(original.multiLine === undefined ? 'undefined' : original.multiLine.toString());
