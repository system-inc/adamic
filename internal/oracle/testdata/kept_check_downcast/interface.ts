// kept_check_downcast_refused/interface.a as .ts, where the cast stays a checked cast: its test passes, so the
// program prints what Node prints.
interface Node { readonly kind: 'identifier' | 'number'; readonly pos: number; }
interface Identifier extends Node { readonly kind: 'identifier'; readonly name: string; }
const raw = { kind: 'identifier' as const, pos: 0, name: 'ok' };
const node: Node = raw;
console.log('viewing');
const identifier = node as Identifier;
console.log(`name ${identifier.name}`);
