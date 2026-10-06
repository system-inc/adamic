import type { Entry } from './3_import_cycle.ts';
export function value(entry: Entry): number {
    return entry.number;
}
