import "node:process";

declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;

export function browserIsSet(): boolean {
    return typeof process !== "undefined" && !!process.browser;
}
if (typeof process !== "undefined") { ({ browser: process.browser } = { browser: true }); }
