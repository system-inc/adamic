import "node:process";

// tracing.ts's process row's declaration, as tsc's tracing.ts ends after adaptation 65, and its typeof.
declare const process: { readonly pid: number };

console.log(typeof process);
