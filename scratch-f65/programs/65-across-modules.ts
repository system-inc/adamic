import { copied } from "./lib/copies.ts";
function describe(value: { readonly code?: number; readonly label: string }): string {
    return String("code" in value) + " " + String("label" in value) + " " + value.label;
}
console.log(describe(copied(undefined)));
console.log(describe(copied({ code: 3 })));
