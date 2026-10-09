import type { ConditionalRoot } from 'original-tsc-types';
interface Base { readonly isDistributive: boolean; }
function read(base: Base): void {
 const viewed = base as ConditionalRoot;
 const items = viewed.aliasTypeArguments;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.flags}`);
}
const raw = {isDistributive: false, aliasTypeArguments: [{flags: 7}]};
function mutate(): void { raw.aliasTypeArguments[0] = {flags: 9}; }
read(raw);
