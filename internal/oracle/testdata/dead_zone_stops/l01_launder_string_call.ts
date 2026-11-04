import { view } from './l01_launder_string_call/back.ts';
function labelOf<T extends { other: number; label?: string }>(value: T): string { return value.label ?? 'none'; }
function make(): string { return new Late().x; }
let first = '';
try { first = String(labelOf(view)); } catch (error) { first = error instanceof Error ? error.name : 'other'; }
class Late { x = 'late'; }
console.log(first);
console.log(String(labelOf(view)));
export { make as other };
