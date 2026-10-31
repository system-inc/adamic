import { Buffer } from "node:buffer";
const b = Buffer.from("ab");
const v: { toJSON(): { readonly data: readonly number[] } } = b;
console.log(String(v.toJSON().data.length));
