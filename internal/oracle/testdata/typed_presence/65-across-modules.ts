import { copied, present } from "./lib/copies.ts";
function describe(value: { readonly code?: number }): string {
    return String("code" in value) + " " + String(present(value)) + " " + String(value.code);
}
console.log(describe(copied(undefined)));
console.log(describe(copied({ code: 3 })));
