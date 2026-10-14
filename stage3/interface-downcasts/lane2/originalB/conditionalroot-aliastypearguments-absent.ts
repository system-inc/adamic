import type { ConditionalRoot } from 'original-tsc-types';
interface Base { readonly isDistributive: boolean; }
function read(base: Base): void {
 const viewed = base as ConditionalRoot;
 const items = viewed.aliasTypeArguments;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.flags}`);
}
const raw = {isDistributive: false, };
read(raw);
