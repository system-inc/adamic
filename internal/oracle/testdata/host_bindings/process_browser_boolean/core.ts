import "node:process";

declare const process: { nextTick?: unknown; browser?: boolean; } | undefined;

if (typeof process !== "undefined") {
    const b: boolean | undefined = process.browser;
    console.log(String(b));
}
