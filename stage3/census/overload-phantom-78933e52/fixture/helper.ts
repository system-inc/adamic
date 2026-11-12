export function identity<T>(x: T): T { return x; }
export function unused<T>(x: T): T { return x; }
export function honest(x: string): string;
export function honest(x: string | undefined): string | undefined;
export function honest(x: string | undefined): string | undefined { return x; }
