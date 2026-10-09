import type { ObjectLiteralExpressionBase, ObjectLiteralElement } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as ObjectLiteralExpressionBase<ObjectLiteralElement>;
 const items = viewed.properties;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.flags}`);
}
const raw = {kind: 0, properties: [{flags: 7}]};
function mutate(): void { raw.properties[0] = {flags: 9}; }
read(raw);
