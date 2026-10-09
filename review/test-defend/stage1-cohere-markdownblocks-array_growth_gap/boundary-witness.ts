import { identifier as original } from './original/identifier.ts';
import { identifier as mutant } from './mutant/identifier.ts';
const text = String.fromCodePoint(125217);
console.log('U+1E921: original=' + original(text).codePointAt(0) + ', D1=' + mutant(text).codePointAt(0));
