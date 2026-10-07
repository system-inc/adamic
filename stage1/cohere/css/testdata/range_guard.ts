// Mutate only the public Range, leaving every source property intact.
import { panic } from 'adamic';
import { parse } from '../parser.ts';
import { render } from '../nodes.ts';
const result = parse('é{}');
if(result.kind === 'Refused') {
    panic(result.message);
}
const root = result.parser.nodes[0] ?? panic('missing root');
root.range[1] = (root.range[1] ?? 0) + 1;
console.log(render(result.parser.nodes, 0, result.parser.input).includes('Range [') ? 'caught' : 'missed');
