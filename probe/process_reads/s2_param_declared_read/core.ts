import "node:process";

declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;

export function browserIsSet(): boolean {
    return typeof process !== "undefined" && !!process.browser;
}
function get(o: { nextTick?: unknown; browser?: unknown; }): string { return o.browser === undefined ? "absent" : "present"; }
export function readAll(): void { if (typeof process !== "undefined") { console.log(get(process)); } }
