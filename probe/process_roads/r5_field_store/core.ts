import "node:process";

declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;

export function browserIsSet(): boolean {
    return typeof process !== "undefined" && !!process.browser;
}
class Holder { constructor(public held: { browser?: unknown }) { } }
if (typeof process !== "undefined") { new Holder(process).held.browser = true; }
