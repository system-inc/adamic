import "node:process";

declare const process: { nextTick?: unknown; browser?: number; } | undefined;

if (typeof process !== "undefined") {
    const b = process.browser ?? 0;
    console.log(String(b));
}
