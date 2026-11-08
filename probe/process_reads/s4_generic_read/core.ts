import "node:process";

declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;

export function browserIsSet(): boolean {
    return typeof process !== "undefined" && !!process.browser;
}
function get<T extends { browser?: unknown }>(o: T): string { return typeof o.browser; }
export function readAll(): void { if (typeof process !== "undefined") { console.log(get(process)); } }
