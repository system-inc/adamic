// The process row's declaration, as tsc's core.ts writes it, and its typeof.
declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;

console.log(typeof process);
