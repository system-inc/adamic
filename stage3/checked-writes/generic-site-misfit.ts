interface Node { flags: number }
interface NarrowNode { readonly flags: 16 }
function clone<T extends Node>(node: T): T { return node; }
function store(view: Node): void { view.flags = 0; }
const narrow: NarrowNode = { flags: 16 };
store(clone(narrow));
console.log(narrow.flags.toString());
