// Narrow Node host declarations. Only the compiler census surface is exposed.
declare module 'node:process' {
    export function cwd(): string;
    // The census only observes this member; scheduling is not available natively.
    export function nextTick(callback: (...args: any[]) => void, ...args: any[]): void;
    export const argv: string[];
    export const execArgv: string[];
    export const platform: string;
    export const pid: number;
    export let exitCode: number | undefined;
    export function exit(code?: number): never;
    export const env: { readonly [key: string]: string | undefined };
    export const stdout: {
        readonly _handle?: { readonly setBlocking: (value: boolean) => void };
        readonly columns: number | undefined;
        readonly isTTY: true | undefined;
        write(buffer: string): boolean;
    };
    export const stderr: { readonly isTTY: true | undefined };
    export interface MemoryUsage { heapUsed: number; }
    export function memoryUsage(): MemoryUsage;
}
declare module 'node:os' {
    export const EOL: string;
    export function platform(): string;
}
declare module 'node:perf_hooks' {
    export interface PerformanceEntry {
        readonly name: string;
        readonly entryType: string;
        readonly startTime: number;
        readonly duration: number;
    }
    export interface PerformanceMark extends PerformanceEntry {}
    export interface PerformanceMeasure extends PerformanceEntry {}
    export interface Performance {
        readonly timeOrigin: number;
        now(): number;
        mark(name: string): PerformanceMark;
        measure(name: string, startMark?: string, endMark?: string): PerformanceMeasure;
        clearMarks(name?: string): void;
        clearMeasures(name?: string): void;
    }
    export const performance: Performance;
}
declare const performance: import('node:perf_hooks').Performance;
