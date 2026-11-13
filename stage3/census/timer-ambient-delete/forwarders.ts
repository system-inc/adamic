import type {} from "node:fs";
export const forwarders = {
    setTimeout: (callback: (label?: unknown) => void, ms: number, label?: unknown): unknown => setTimeout(callback, ms, label),
    clearTimeout: (timeoutId: unknown): void => clearTimeout(timeoutId),
};
export const globalSetTimeout = setTimeout;
export const globalClearTimeout = clearTimeout;
