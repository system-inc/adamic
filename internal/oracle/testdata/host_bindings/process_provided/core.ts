import "node:process";

declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;

export function isNodeLikeSystem(): boolean {
    return typeof process !== "undefined" && !!process.nextTick && !process.browser;
}

console.log(String(isNodeLikeSystem()));
console.log(typeof process);
if (typeof process !== "undefined") {
    console.log(String(!!process.nextTick));
    console.log(String(!process.browser));
}
