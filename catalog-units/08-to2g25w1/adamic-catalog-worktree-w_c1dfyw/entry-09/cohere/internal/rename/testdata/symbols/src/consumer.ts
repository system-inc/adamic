import { err } from './origin';
import { err as failure } from './origin';

export function useBoth(): string {
    const shadowed = (() => {
        const err = 'inner';
        return err;
    })();
    return err('a') + failure('b') + shadowed;
}

export function punning() {
    const punned = 'value';
    return { punned };
}

export function stringKeyed(table: Record<string, unknown>) {
    return table['err'];
}

export function outside(value: string): string {
    return value.toUpperCase();
}
