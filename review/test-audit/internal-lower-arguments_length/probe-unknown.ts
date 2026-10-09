function test(value: unknown): unknown { if (Array.isArray(value)) return value[0]; return undefined; }
