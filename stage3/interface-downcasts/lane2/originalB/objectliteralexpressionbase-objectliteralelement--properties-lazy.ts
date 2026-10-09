import type { ObjectLiteralExpressionBase, ObjectLiteralElement } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as ObjectLiteralExpressionBase<ObjectLiteralElement>;
 console.log('kept');
}
const raw = {kind: 0, properties: 7};
read(raw);
