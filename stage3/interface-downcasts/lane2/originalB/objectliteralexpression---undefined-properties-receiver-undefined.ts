import type { ObjectLiteralExpression } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as ObjectLiteralExpression;
 const items = viewed?.properties;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.flags}`);
}
const raw = {kind: 211, properties: [{kind: 304, flags: 7}]};
read(undefined);
