import "node:process";

declare const process: { nextTick?: unknown; browser?: string; } | undefined;

if (typeof process !== "undefined") {
    const b: string | undefined = process.browser;
    console.log(String(b));
}
