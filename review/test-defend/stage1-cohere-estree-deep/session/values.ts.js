export class Value {
    kind;
    text = '';
    cooked = '';
    number = 0;
    node = -1;
    list = [];
    constructor(kind){
        this.kind = kind;
    }
}
export function absent() {
    return new Value('null');
}
export function stringValue(text) {
    const value = new Value('string');
    value.text = text;
    return value;
}
export function boolValue(flag) {
    const value = new Value('bool');
    value.number = flag ? 1 : 0;
    return value;
}
export function numberValue(number) {
    const value = new Value('number');
    value.number = number;
    return value;
}
export function childValue(node) {
    if (node < 0) {
        return absent();
    }
    const value = new Value('node');
    value.node = node;
    return value;
}
export function listValue(list) {
    const value = new Value('list');
    value.list = list;
    return value;
}
export function templateValue(raw, cooked, valid) {
    const value = new Value('template');
    value.text = raw;
    value.list = [];
    value.number = valid ? 1 : 0;
    value.node = -1;
    value.cooked = cooked;
    return value;
}
export function goNumber(raw) {
    const text = raw.split('_').join('').toLowerCase();
    const prefix = text.slice(0, 2);
    const base = prefix === '0x' ? 16 : prefix === '0o' ? 8 : prefix === '0b' ? 2 : 10;
    if (base !== 10) {
        let digits = text.slice(2);
        while(digits.startsWith('0')){
            digits = digits.slice(1);
        }
        if (base === 16 && digits.length > 16 || base === 2 && digits.length > 64 || base === 8 && (digits.length > 22 || digits.length === 22 && digits.slice(0, 1) !== '1')) {
            return NaN;
        }
        return digits === '' ? 0 : Number.parseInt(digits, base);
    }
    const number = Number.parseFloat(text);
    return number === Infinity || number === -Infinity ? NaN : number;
}
export function bigintDecimal(raw) {
    const text = raw.slice(0, -1).split('_').join('').toLowerCase();
    const prefix = text.slice(0, 2);
    const base = prefix === '0x' ? 16 : prefix === '0o' ? 8 : prefix === '0b' ? 2 : 10;
    const digits = base === 10 ? text : text.slice(2);
    let result = '0';
    for (const character of digits){
        let carry = '0123456789abcdef'.indexOf(character);
        let next = '';
        for(let index = result.length - 1; index >= 0; index--){
            const value = (result.charCodeAt(index) - 48) * base + carry;
            next = String.fromCharCode(48 + value % 10) + next;
            carry = Math.floor(value / 10);
        }
        while(carry > 0){
            next = String.fromCharCode(48 + carry % 10) + next;
            carry = Math.floor(carry / 10);
        }
        result = next;
    }
    let start = 0;
    while(start < result.length - 1 && result.slice(start, start + 1) === '0'){
        start++;
    }
    return result.slice(start);
}
