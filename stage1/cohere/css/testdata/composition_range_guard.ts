// The same corruption on the complete composed tree, while native is blocked.
import { panic } from 'adamic';
import { compose } from '../compose.ts';
import { renderTree } from '../tree.ts';
const result = compose('é{}', false);
if(result.kind === 'Refused') {
    panic(result.message);
}
const root = result.tree.at(result.tree.root);
root.range[1] = (root.range[1] ?? 0) + 1;
console.log(renderTree(result.tree, result.tree.root).includes('<range>') ? 'caught' : 'missed');
