import type { Weak } from "adamic";
interface Box { readonly label: string }
function box(): Box { return { label: "label" }; }
type Strong = { readonly kind: "v"; readonly v: Box } | { readonly kind: "w"; readonly w: number };
type Viewed = { readonly kind: "v"; readonly v: Weak<Box> } | { readonly kind: "w"; readonly w: number };
const make = (): Strong => ({ kind: "v", v: box() });
const viewed: () => Viewed = make;
const got = viewed();
console.log(got.kind === "v" ? got.v?.label ?? "none" : "w");
