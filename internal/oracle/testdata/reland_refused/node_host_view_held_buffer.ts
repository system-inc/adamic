import { Buffer } from "node:buffer";
const b = Buffer.from("ab");
const v: { readonly byteLength: number } = b;
console.log(String(v.byteLength));
