import "node:process";

declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;

export function browserIsSet(): boolean {
    return typeof process !== "undefined" && !!process.browser;
}
export function readAll(): void { if (typeof process !== "undefined") { const view: { browser?: unknown } = process; console.log(typeof view.browser); } }
