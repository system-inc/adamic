import { performance as hooks } from "node:perf_hooks";

/** @internal */
export interface PerformanceTime {
    now(): number;
    readonly timeOrigin: number;
}

/** @internal */
export interface Performance extends PerformanceTime {
    mark(name: string): void;
    measure(name: string, startMark?: string, endMark?: string): void;
    clearMeasures(name?: string): void;
    clearMarks(name?: string): void;
}

// Browser globals for the Web Performance User Timings API
declare const performance: Performance | undefined;

console.log(typeof performance);
if (typeof performance === "object") {
    console.log(String(performance === hooks));
    console.log(typeof performance.now);
    console.log(typeof performance.timeOrigin);
    console.log(typeof performance.mark);
    console.log(typeof performance.measure);
    console.log(typeof performance.clearMeasures);
    console.log(typeof performance.clearMarks);
    console.log(String(performance.now() >= 0));
    console.log(String(performance.now() < 60000));
    const first = performance.now();
    const second = performance.now();
    console.log(String(second >= first));
    console.log(String(typeof performance.timeOrigin === "number"));
}
