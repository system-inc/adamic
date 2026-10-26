export function claim(value: number | string): value is number { return true; }
export function use(value: number | string): string { if (claim(value)) return value.toString(); return "other"; }
