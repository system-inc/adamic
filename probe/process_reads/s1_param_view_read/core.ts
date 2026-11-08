import "node:process";

declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;

export function browserIsSet(): boolean {
    return typeof process !== "undefined" && !!process.browser;
}
function get(o: { browser?: unknown }): string { return typeof o.browser; }
export function readAll(): void { if (typeof process !== "undefined") { console.log(get(process)); } }
