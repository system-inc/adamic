import "node:process";

declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;

export function browserIsSet(): boolean {
    return typeof process !== "undefined" && !!process.browser;
}
if (typeof process !== "undefined") { const all = [process]; for (const p of all) { p.browser = true; } }
