import { Parser } from '../../../typescript/parser/parser.ts';
const parser = new Parser('type X = {');
console.log(`${parser.file()}`);
