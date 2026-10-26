export function claim(value: number | string): value is string { return true; }
export function use(value: number | string): string { if (claim(value)) return value; return "other"; }
