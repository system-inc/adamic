import { Parser } from '../sourceParser.ts';
const parser = new Parser('type X = {');
console.log(`${parser.file()}`);
