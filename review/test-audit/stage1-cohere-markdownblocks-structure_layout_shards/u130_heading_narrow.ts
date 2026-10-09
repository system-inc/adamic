import { DocumentArena } from '../document.ts';
import { printHeading } from '../structure.ts';
const a = new DocumentArena();
console.log(a.node(printHeading(a, [a.text('x')], 1, false, '# x')).kind);
