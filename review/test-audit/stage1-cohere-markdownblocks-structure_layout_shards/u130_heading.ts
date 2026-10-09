import { DocumentArena, printDocument } from '../document.ts';
import { printHeading } from '../structure.ts';
const a = new DocumentArena();
console.log(printDocument(a, printHeading(a, [a.text('x')], 1, false, '# x'), 120, 4));
