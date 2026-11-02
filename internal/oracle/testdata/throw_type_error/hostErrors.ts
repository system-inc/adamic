/** @internal */
export function errorCode(error: unknown): string | undefined {
    return typeof error === "object" && error !== null && "code" in error && typeof error.code === "string" ? error.code : undefined;
}

/** @internal */
export function errorMessage(error: unknown): string | undefined {
    return typeof error === "object" && error !== null && "message" in error && typeof error.message === "string" ? error.message : undefined;
}

/** @internal */
export function propertyOf(value: unknown, key: string): unknown {
    if (value === null || value === undefined) {
        throw new TypeError(`Cannot read properties of ${String(value)} (reading '${key}')`);
    }
    return Reflect.get(Object(value), key, value);
}
