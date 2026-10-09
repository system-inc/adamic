import { absent } from './values.ts';
import { Property } from './property.ts';
export class Node {
    type;
    start;
    end;
    properties = [];
    parenthesized = false;
    contentEnd = 0;
    hasContentEnd = false;
    constructor(type, start, end){
        this.type = type;
        this.start = start;
        this.end = end;
    }
    get(key) {
        for (const property of this.properties){
            if (property.key === key) {
                return property.value;
            }
        }
        return absent();
    }
    has(key) {
        for (const property of this.properties){
            if (property.key === key) {
                return true;
            }
        }
        return false;
    }
    set(key, value) {
        for (const property of this.properties){
            if (property.key === key) {
                property.value = value;
                return;
            }
        }
        this.properties.push(new Property(key, value));
    }
    delete(key) {
        for(let index = 0; index < this.properties.length; index++){
            if (this.properties[index]?.key === key) {
                this.properties.splice(index, 1);
                return;
            }
        }
    }
    child(key) {
        const value = this.get(key);
        return value.kind === 'node' ? value.node : -1;
    }
    list(key) {
        const value = this.get(key);
        const empty = [];
        return value.kind === 'list' ? value.list : empty;
    }
    string(key) {
        const value = this.get(key);
        return value.kind === 'string' ? value.text : '';
    }
    bool(key) {
        const value = this.get(key);
        return value.kind === 'bool' && value.number === 1;
    }
    truthy(key) {
        const value = this.get(key);
        return value.kind === 'null' ? false : value.kind === 'bool' ? value.number !== 0 : value.kind === 'number' ? value.number !== 0 && !Number.isNaN(value.number) : value.kind === 'string' ? value.text !== '' : true;
    }
    keys() {
        const keys = [];
        for (const property of this.properties){
            keys.push(property.key);
        }
        return keys;
    }
}
