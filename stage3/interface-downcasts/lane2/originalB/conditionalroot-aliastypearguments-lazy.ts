import type { ConditionalRoot } from 'original-tsc-types';
interface Base { readonly isDistributive: boolean; }
function read(base: Base): void {
 const viewed = base as ConditionalRoot;
 console.log('kept');
}
const raw = {isDistributive: false, aliasTypeArguments: 7};
read(raw);
