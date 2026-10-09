const node: { readonly enabled: true } = { enabled: true };
function store(view: { enabled: boolean }, value: boolean): void { view.enabled = value; }
store(node, false);
console.log(node.enabled ? 'true' : 'false');
