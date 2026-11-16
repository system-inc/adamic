import type { Weak } from "adamic";
interface Box { readonly label: string }
function box(): Box { return { label: "label" }; }
type Strong = { readonly kind: "v"; readonly v: Box } | { readonly kind: "w"; readonly w: number };
type Twin = { readonly kind: "w"; readonly w: number } | { readonly kind: "v"; readonly v: Box };
const strong: Strong = { kind: "v", v: box() };
const twin: Twin = strong;
console.log(twin.kind === "v" ? twin.v.label : "w");
