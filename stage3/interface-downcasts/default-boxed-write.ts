interface Node { readonly kind: 'number' | 'identifier'; }
interface NumberNode extends Node { readonly kind: 'number'; value: number; }
const raw: { kind: 'number'; value: string | number } = {kind:'number', value:undefined!};
const held: Node = raw; const view = held as NumberNode; view.value = 42;
console.log(`${view.value}`); console.log(`${view === held}`);
