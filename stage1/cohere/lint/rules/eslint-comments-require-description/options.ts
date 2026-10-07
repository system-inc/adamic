import { panic } from 'adamic';
import { OptionsJson } from '../../helpers/options_json.ts';
import { StrictOptions } from '../../helpers/strict_options.ts';
// Generated from the pinned Go target descriptor, not inferred from settings keys.
const shape = new StrictOptions("{\"fields\":{\"additionalDirectives\":{\"shape\":{\"item\":{\"kind\":\"string\"},\"kind\":\"array\"},\"tagged\":true},\"ignore\":{\"shape\":{\"item\":{\"kind\":\"string\"},\"kind\":\"array\"},\"tagged\":true}},\"kind\":\"object\"}");
const allowed = ['eslint', 'eslint-disable', 'eslint-disable-line', 'eslint-disable-next-line', 'eslint-enable', 'eslint-env', 'exported', 'global', 'globals'];
export class Options {
    valid = true;
    ignored: string[] = [];
    additional: string[] = [];
}
export function decode(raw: string): Options {
    const result = new Options();
    if(raw === '') { return result; }
    if(shape.check(raw) !== 'valid') { result.valid = false; return result; }
    const values = new OptionsJson(raw);
    const root = values.parse();
    if(root < 0) { result.valid = false; return result; }
    if(values.node(root).kind === 'null') { return result; }
    for(const key of ['ignore', 'additionalDirectives']) {
        const index = values.field(root, key);
        if(index < 0 || values.node(index).kind === 'null') { continue; }
        const target = key === 'ignore' ? result.ignored : result.additional;
        for(const child of values.node(index).children) {
            // Go decoding a null string element leaves its zero value, the empty string.
            const value = values.node(child).kind === 'null' ? '' : values.node(child).text;
            if(key === 'ignore' && (!allowed.includes(value) || target.includes(value))) { result.valid = false; return result; }
            target.push(value);
        }
    }
    return result;
}
export function checked(raw: string): Options {
    const options = decode(raw);
    if(!options.valid) { panic('@eslint-community/eslint-comments/require-description: invalid options'); }
    return options;
}
