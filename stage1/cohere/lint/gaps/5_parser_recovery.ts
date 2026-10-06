// A valid Adamic program whose input is a malformed TypeScript interface.
// Node and native currently loop at EOF; Go's parser recovers the method.
import { Parser } from '../../../typescript/parser/parser.ts';
const parser = new Parser('interface I { m(a: string): void;', 'missing-brace.ts');
parser.file();
console.log('recovered');
