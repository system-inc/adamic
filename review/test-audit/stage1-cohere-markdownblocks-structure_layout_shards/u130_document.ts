import { DocumentArena, printDocument } from '../document.ts';
const a = new DocumentArena();
console.log(printDocument(a, a.text('x'), 120, 4));
