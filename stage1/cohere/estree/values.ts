// Typed field values. Node references are arena indexes.
export class Value {
    readonly kind: string;
    text = '';
    cooked = '';
    number = 0;
    node = -1;
    list: number[] = [];
    constructor(kind: string) {
        this.kind = kind;
    }
}
export function absent(): Value {
    return new Value('null');
}
export function stringValue(text: string): Value {
    const value = new Value('string');
    value.text = text;
    return value;
}
export function boolValue(flag: boolean): Value {
    const value = new Value('bool');
    value.number = flag ? 1 : 0;
    return value;
}
export function numberValue(number: number): Value {
    const value = new Value('number');
    value.number = number;
    return value;
}
export function childValue(node: number): Value {
    if(node < 0) {
        return absent();
    }
    const value = new Value('node');
    value.node = node;
    return value;
}
export function listValue(list: number[]): Value {
    const value = new Value('list');
    value.list = list;
    return value;
}
export function templateValue(raw: string, cooked: string, valid: boolean): Value {
    const value = new Value('template');
    value.text = raw;
    value.list = [];
    value.number = valid ? 1 : 0;
    value.node = -1;
    value.cooked = cooked;
    return value;
}
