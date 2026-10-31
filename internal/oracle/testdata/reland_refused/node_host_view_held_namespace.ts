import process from "node:process";
const p: { readonly config: { readonly target_defaults: { readonly cflags: readonly number[] } } } = process;
console.log(String(p.config.target_defaults.cflags.length));
