const node: { readonly enabled: true } = { enabled: true };
function store(view: { enabled: boolean }, value: boolean): void { view.enabled = value; }
store(node, true);
console.log(node.enabled ? 'true' : 'false');
