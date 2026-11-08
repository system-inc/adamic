import "node:process";

declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;

export function browserIsSet(): boolean {
    return typeof process !== "undefined" && !!process.browser;
}
function set(o: { browser?: unknown }): void { o.browser = true; }
if (typeof process !== "undefined") { set(process); }
