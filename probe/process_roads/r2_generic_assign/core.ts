import "node:process";

declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;

export function browserIsSet(): boolean {
    return typeof process !== "undefined" && !!process.browser;
}
function set<T extends object>(o: T): T { return Object.assign(o, { browser: true }); }
if (typeof process !== "undefined") { set(process); }
