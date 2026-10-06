import type { Value } from './values.ts';
export class Property {
    readonly key: string;
    value: Value;
    constructor(key: string, value: Value) {
        this.key = key;
        this.value = value;
    }
}
