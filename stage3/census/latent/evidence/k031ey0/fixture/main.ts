interface Grow<T> { readonly value?: T; readonly next?: Grow<Grow<T>> }
interface N { readonly kind: number; readonly grow?: Grow<number> }
interface Id extends N { readonly kind: 1; readonly text: string }
declare function isId<T>(n: N): n is Id;
const answer: number = 1;
