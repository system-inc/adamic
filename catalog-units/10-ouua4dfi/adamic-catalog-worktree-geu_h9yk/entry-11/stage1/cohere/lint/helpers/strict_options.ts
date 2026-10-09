// rule.UnmarshalOptions without reflection: the generated descriptor supplies the Go target's fields.
import { panic } from 'adamic';
import { OptionsJson } from './options_json.ts';
export class StrictOptions {
    readonly shapes: OptionsJson;
    readonly root: number;
    unsupported = '';
    constructor(descriptor: string) { this.shapes = new OptionsJson(descriptor); this.root = this.shapes.parse(); }
    matches(shape: number, values: OptionsJson, index: number): boolean {
        const kindIndex = this.shapes.field(shape, 'kind'); const kind = this.shapes.node(kindIndex).text;
        if(kind === 'unsupported') { this.unsupported = 'custom or unsupported Go option type'; return false; }
        const value = values.node(index);
        // encoding/json null leaves scalars/structs untouched and sets pointers/slices/maps to nil.
        if(value.kind === 'null') { return true; }
        if(kind !== value.kind) { return false; }
        if(kind === 'array') {
            const item = this.shapes.field(shape, 'item');
            return value.children.every(child => this.matches(item, values, child));
        }
        if(kind === 'object') {
            const map = this.shapes.field(shape, 'item');
            if(map >= 0) { return value.children.every(child => this.matches(map, values, child)); }
            const fields = this.shapes.field(shape, 'fields');
            for(let i = 0; i < value.keys.length; i++) {
                const key = value.keys[i] ?? panic('option key'); let field = this.shapes.field(fields, key);
                if(field < 0) {
                    if([...key].some(c => c.charCodeAt(0) > 127)) { this.unsupported = 'Unicode fold for untagged option keys'; return false; }
                    for(const candidate of this.shapes.node(fields).keys) {
                        const entry = this.shapes.field(fields, candidate);
                        const tagged = this.shapes.field(entry, 'tagged');
                        if(this.shapes.node(tagged).text === 'false' && candidate.toLowerCase() === key.toLowerCase()) {
                            if([...candidate + key].some(c => c.charCodeAt(0) > 127)) { this.unsupported = 'Unicode fold for untagged option keys'; return false; }
                            field = entry; break;
                        }
                    }
                }
                if(field < 0) { return false; }
                const target = this.shapes.field(field, 'shape');
                if(!this.matches(target, values, value.children[i] ?? panic('option value'))) { return false; }
            }
        }
        if(kind === 'number') {
            if(!Number.isFinite(value.number)) { return false; }
            const integer = this.shapes.field(shape, 'integer');
            if(integer >= 0 && this.shapes.node(integer).text === 'true') {
                // Go integer JSON forbids decimals/exponents, even those whose numeric value is integral.
                if(value.text.includes('.') || value.text.includes('e') || value.text.includes('E')) { return false; }
                const negative = value.text.startsWith('-');
                const digits = negative ? value.text.slice(1) : value.text;
                const bound = this.shapes.node(this.shapes.field(shape, negative ? 'rawMinimum' : 'rawMaximum')).text;
                const limit = negative ? bound.slice(1) : bound;
                if(digits.length > limit.length || (digits.length === limit.length && digits > limit)) { return false; }

            }
        }
        return true;
    }
    check(raw: string): string {
        this.unsupported = '';
        if(this.root < 0) { return 'NotYet: invalid descriptor'; }
        const values = new OptionsJson(raw); const root = values.parse();
        if(root < 0) { return 'invalid'; }
        const valid = this.matches(this.root, values, root);
        return this.unsupported.length > 0 ? `NotYet: ${this.unsupported}` : valid ? 'valid' : 'invalid';
    }
}
