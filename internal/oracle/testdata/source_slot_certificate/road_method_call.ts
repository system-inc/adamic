// A method call by the written name reads the slot as a callee, which the certificate doesn't judge.
interface Base { readonly kind: "box" | "other" }
interface Box extends Base { readonly kind: "box"; child: { ready: boolean } }
const raw = { kind: "box" as const, child: { ready: true, other: 42 } };
const held: Base = raw;
const view = held as Box;
view.child = { ready: false };
console.log(String(view.child.ready));
interface Caller { child(): string }
const caller: Caller = { child: (): string => "called" };
console.log(caller.child());
