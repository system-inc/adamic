import "node:process";

declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;

export function browserIsSet(): boolean {
    return typeof process !== "undefined" && !!process.browser;
}
function set(o: unknown): void { if (typeof o === "object" && o !== null) { Object.defineProperty(o, "browser", { value: true }); } }
if (typeof process !== "undefined") { set(process); }
