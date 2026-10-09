import { panic } from 'adamic';
const kind = 'text';
console.log(kind === 'text' ? 'T' : kind === 'line' ? 'L' : panic('unknown'));
