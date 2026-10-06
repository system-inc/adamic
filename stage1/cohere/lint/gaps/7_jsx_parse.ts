import { Parser } from '../../../typescript/parser/parser.ts';

const parser = new Parser('export const Thing = () => <p>hi</p>;\n', 'Thing.tsx');
parser.file();
let found = false;
for(const node of parser.nodes) {
    if(node.kind === 'JsxElement') {
        found = true;
    }
}
console.log(`${found}`);
