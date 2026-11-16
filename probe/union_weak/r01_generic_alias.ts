import type { Weak } from "adamic";
interface Box { readonly label: string }
function box(): Box { return { label: "label" }; }
type W<T extends object> = Weak<T>;
type Strong = { readonly kind: "v"; readonly v: Box } | { readonly kind: "w"; readonly w: number };
type Viewed = { readonly kind: "v"; readonly v: W<Box> } | { readonly kind: "w"; readonly w: number };
const strong: Strong = { kind: "v", v: box() };
const viewed: Viewed = strong;
console.log(viewed.kind === "v" ? viewed.v?.label ?? "none" : "w");
