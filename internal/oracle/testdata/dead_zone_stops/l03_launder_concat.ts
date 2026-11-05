import { view } from './l03_launder_concat/back.ts';
function labelOf<T extends { other: number; label?: string }>(value: T): string { return value.label ?? 'none'; }
function make(): string { return new Late().x; }
let first = '';
try { first = 'x'.concat(labelOf(view)); } catch (error) { first = error instanceof Error ? error.name : 'other'; }
class Late { x = 'late'; }
console.log(first);
console.log('x'.concat(labelOf(view)));
export { make as other };
