import type { ObjectLiteralExpression } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as ObjectLiteralExpression;
 console.log('kept');
}
const raw = {kind: 211, properties: 7};
read(raw);
