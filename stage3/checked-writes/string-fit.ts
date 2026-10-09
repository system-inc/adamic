const node: { readonly text: 'left' | 'right' } = { text: 'left' };
function store(view: { text: string }, value: string): void { view.text = value; }
store(node, 'right');
console.log(node.text);
