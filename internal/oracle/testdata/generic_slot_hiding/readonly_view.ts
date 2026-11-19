// #4ysqm2m: a Target[] seen as a readonly Source[], which nothing can write through, so no write can store a
// Source in held. It lowers and agrees with Node.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

const held: Target[] = [{ kind: 1 }, { kind: 2, label: "two" }];
const view: readonly Source[] = held;
const first = view[0];
console.log(first === undefined ? "none" : String(first.kind));
console.log(String(view.length));
