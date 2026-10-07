import { panic } from 'adamic';
import { OptionsJson } from '../../helpers/options_json.ts';
import { StrictOptions } from '../../helpers/strict_options.ts';
const shape = new StrictOptions("{\"fields\":{\"requirements\":{\"shape\":{\"item\":{\"fields\":{\"contextKey\":{\"shape\":{\"kind\":\"string\"},\"tagged\":true},\"requiresAny\":{\"shape\":{\"item\":{\"kind\":\"string\"},\"kind\":\"array\"},\"tagged\":true}},\"kind\":\"object\"},\"kind\":\"array\"},\"tagged\":true}},\"kind\":\"object\"}");
export class Options {
    valid = true;
    readonly requirements = new Map<string, string[]>();
}
export function decode(raw: string): Options {
    const result = new Options();
    if(raw === '') { return result; }
    if(shape.check(raw) !== 'valid') { result.valid = false; return result; }
    const data = new OptionsJson(raw);
    const root = data.parse();
    if(root < 0) { result.valid = false; return result; }
    const array = data.field(root, 'requirements');
    if(array < 0 || data.node(array).kind === 'null') { return result; }
    for(const row of data.node(array).children) {
        const keyIndex = data.field(row, 'contextKey');
        const key = keyIndex < 0 || data.node(keyIndex).kind === 'null' ? '' : data.node(keyIndex).text;
        const requiresIndex = data.field(row, 'requiresAny');
        const requires: string[] = [];
        if(requiresIndex >= 0) { for(const value of data.node(requiresIndex).children) { requires.push(data.node(value).kind === 'null' ? '' : data.node(value).text); } }
        if(key === '' || requires.length === 0) { result.valid = false; return result; }
        // Go's map keeps the last requirement for a repeated key.
        result.requirements.set(key, requires);
    }
    return result;
}
export function checked(raw: string): Options {
    const result = decode(raw);
    if(!result.valid) { panic('base/security-require-context-access: invalid options'); }
    return result;
}
