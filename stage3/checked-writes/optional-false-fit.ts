const original: { readonly multiLine: false | undefined } = { multiLine: false };
function store(view: { multiLine: boolean | undefined }, value: boolean | undefined): void { view.multiLine = value; }
store(original, false);
console.log(original.multiLine === undefined ? 'undefined' : original.multiLine.toString());
