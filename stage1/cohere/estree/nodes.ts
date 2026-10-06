// Ordered own fields preserve the converter and postprocessor's observable writes.
import { absent } from './values.ts';
import type { Value } from './values.ts';
import { Property } from './property.ts';
export class Node {
    type: string;
    start: number;
    end: number;
    readonly properties: Property[] = [];
    parenthesized = false;
    contentEnd = 0;
    hasContentEnd = false;
    constructor(type: string, start: number, end: number) {
        this.type = type;
        this.start = start;
        this.end = end;
    }
    get(key: string): Value {
        for(const property of this.properties) {
            if(property.key === key) {
                return property.value;
            }
        }
        return absent();
    }
    has(key: string): boolean {
        for(const property of this.properties) {
            if(property.key === key) {
                return true;
            }
        }
        return false;
    }
    set(key: string, value: Value): void {
        for(const property of this.properties) {
            if(property.key === key) {
                property.value = value;
                return;
            }
        }
        this.properties.push(new Property(key, value));
    }
    delete(key: string): void {
        for(let index = 0; index < this.properties.length; index++) {
            if(this.properties[index]?.key === key) {
                this.properties.splice(index, 1);
                return;
            }
        }
    }
    child(key: string): number {
        const value = this.get(key);
        return value.kind === 'node' ? value.node : -1;
    }
    list(key: string): number[] {
        const value = this.get(key);
        const empty: number[] = [];
        return value.kind === 'list' ? value.list : empty;
    }
    string(key: string): string {
        const value = this.get(key);
        return value.kind === 'string' ? value.text : '';
    }
    bool(key: string): boolean {
        const value = this.get(key);
        return value.kind === 'bool' && value.number === 1;
    }
    truthy(key: string): boolean {
        const value = this.get(key);
        return value.kind === 'null'
            ? false
            : value.kind === 'bool'
              ? value.number !== 0
              : value.kind === 'number'
                ? value.number !== 0 && !Number.isNaN(value.number)
                : value.kind === 'string'
                  ? value.text !== ''
                  : true;
    }
    keys(): string[] {
        const keys: string[] = [];
        for(const property of this.properties) {
            keys.push(property.key);
        }
        return keys;
    }
}
