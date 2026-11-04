import { ready } from './l04_patched_prototype/back.ts';
function make(): string { return new Late().x; }
let first = '';
try { first = String('a,b'.split(',').length); } catch (error) { first = error instanceof Error ? error.name : 'other'; }
class Late { x = 'late'; }
console.log(first);
console.log(String(ready));
console.log(String('a,b'.split(',').length));
export { make as other };
