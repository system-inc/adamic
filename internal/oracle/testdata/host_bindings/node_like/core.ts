// require as tsc's core.ts declares it (adaptation 30) and reads it in isNodeLikeSystem. It is Absent on
// every target, as Node answers under the oracle's module mode, so typeof is "undefined" on Node and on
// every backend and the test is false. tsc's isNodeLikeSystem also tests process, whose own reads
// TestHostBindingsAgreeWithNode's process_provided holds (#zqz01vh), so this program keeps require's half alone.
declare const require: unknown;

export function isNodeLikeSystem(): boolean {
    return typeof require !== "undefined";
}

export function loader(): string {
    if (typeof require !== "undefined") {
        return String(require);
    }
    return "no loader";
}

export function requireType(): string {
    return typeof require;
}
