import type { EmitNode } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as EmitNode;
 const items = viewed?.identifierTypeArguments;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.flags}`);
}
const raw = {flags: 8, identifierTypeArguments: [{kind: 116, flags: 7}]};
function mutate(): void { raw.identifierTypeArguments[0] = {kind: 116, flags: 9}; }
read(raw);
