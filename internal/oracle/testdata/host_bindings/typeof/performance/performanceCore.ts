// The performance row's declaration, as tsc's performanceCore.ts writes it, and its typeof. Performance
// is tsc's own interface there, cut to one member.
interface Performance {
    now(): number;
}

declare const performance: Performance | undefined;

console.log(typeof performance);
