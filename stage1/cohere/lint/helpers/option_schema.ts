// optionschema.Validate's supported schema vocabulary. No unrecognized constraint passes silently.
import { panic } from 'adamic';
import { OptionsJson } from './options_json.ts';
export class OptionSchema {
    readonly schema: OptionsJson;
    readonly root: number;
    unsupported = '';
    readonly initialUnsupported: string = '';
    constructor(text: string) {
        this.schema = new OptionsJson(text); this.root = this.schema.parse();
        if(this.root < 0) { this.unsupported = 'invalid schema JSON'; }
        else { this.audit(this.root); }
        this.initialUnsupported = this.unsupported;
    }
    audit(index: number): void {
        const s = this.schema.node(index);
        if(s.kind === 'boolean') { return; }
        if(s.kind !== 'object') { this.unsupported = 'schema must be an object or boolean'; return; }
        const known = ['type', 'enum', 'const', 'minimum', 'maximum', 'minLength', 'maxLength', 'minItems', 'maxItems', 'uniqueItems', 'required', 'properties', 'additionalProperties', 'items', 'additionalItems', 'allOf', 'anyOf', 'oneOf', 'not', '$ref', '$defs', 'definitions', 'description', 'default', 'title', 'deprecated'];
        for(let i = 0; i < s.keys.length; i++) {
            const key = s.keys[i] ?? panic('schema key'); const child = s.children[i] ?? panic('schema child');
            if(key === '$ref' && this.reference(this.schema.node(child).text) < 0) { return; }
            if(!known.includes(key)) { this.unsupported = `unsupported schema keyword ${key}`; return; }
            if(['properties', '$defs', 'definitions'].includes(key)) {
                for(const member of this.schema.node(child).children) { this.audit(member); }
            } else if(['allOf', 'anyOf', 'oneOf'].includes(key) || (key === 'items' && this.schema.node(child).kind === 'array')) {
                for(const member of this.schema.node(child).children) { this.audit(member); }
            } else if(['items', 'additionalItems', 'additionalProperties', 'not'].includes(key)) { this.audit(child); }
            if(this.unsupported !== '') { return; }
        }
    }
    field(index: number, name: string): number { return this.schema.field(index, name); }
    limit(index: number, name: string, fallback: number): number {
        const field = this.field(index, name); return field < 0 ? fallback : this.schema.node(field).number;
    }
    reference(text: string): number {
        if(!text.startsWith('#/')) { this.unsupported = 'external schema reference'; return -1; }
        let target = this.root;
        for(const raw of text.slice(2).split('/')) {
            const key = raw.split('~1').join('/').split('~0').join('~');
            const node = this.schema.node(target);
            target = node.kind === 'array' ? node.children[Number.parseInt(key, 10)] ?? -1 : this.field(target, key);
            if(target < 0) { this.unsupported = 'missing schema reference'; return -1; }
        }
        return target;
    }
    validate(schemaIndex: number, values: OptionsJson, valueIndex: number, depth: number): boolean {
        if(depth > 512) { this.unsupported = 'schema recursion exceeds 512'; return false; }
        const s = this.schema.node(schemaIndex); const v = values.node(valueIndex);
        if(s.kind === 'boolean') { return s.text === 'true'; }
        const ref = this.field(schemaIndex, '$ref');
        if(ref >= 0) { const target = this.reference(this.schema.node(ref).text); return target >= 0 && this.validate(target, values, valueIndex, depth + 1); }
        const type = this.field(schemaIndex, 'type');
        if(type >= 0) {
            const t = this.schema.node(type); const names = t.kind === 'array' ? t.children.map(i => this.schema.node(i).text) : [t.text];
            if(!names.some(name => name === v.kind || (name === 'integer' && v.kind === 'number' && Number.isFinite(v.number) && Number.isInteger(v.number)))) { return false; }
        }
        const constant = this.field(schemaIndex, 'const');
        if(constant >= 0 && !values.equal(valueIndex, this.schema, constant)) { return false; }
        const enumeration = this.field(schemaIndex, 'enum');
        if(enumeration >= 0 && !this.schema.node(enumeration).children.some(i => values.equal(valueIndex, this.schema, i))) { return false; }
        if(v.kind === 'number' && (v.number < this.limit(schemaIndex, 'minimum', -Infinity) || v.number > this.limit(schemaIndex, 'maximum', Infinity))) { return false; }
        if(v.kind === 'string') {
            const length = [...v.text].length;
            if(length < this.limit(schemaIndex, 'minLength', 0) || length > this.limit(schemaIndex, 'maxLength', Infinity)) { return false; }
        }
        if(v.kind === 'array') {
            if(v.children.length < this.limit(schemaIndex, 'minItems', 0) || v.children.length > this.limit(schemaIndex, 'maxItems', Infinity)) { return false; }
            const unique = this.field(schemaIndex, 'uniqueItems');
            if(unique >= 0 && this.schema.node(unique).text === 'true') {
                for(let i = 0; i < v.children.length; i++) { for(let j = 0; j < i; j++) {
                    if(values.equal(v.children[i] ?? panic('item'), values, v.children[j] ?? panic('item'))) { return false; }
                } }
            }
            const items = this.field(schemaIndex, 'items');
            if(items >= 0) {
                const tuple = this.schema.node(items);
                for(let i = 0; i < v.children.length; i++) {
                    const item = tuple.kind === 'array' ? tuple.children[i] ?? this.field(schemaIndex, 'additionalItems') : items;
                    if(item >= 0 && !this.validate(item, values, v.children[i] ?? panic('item'), depth + 1)) { return false; }
                }
            }
        }
        if(v.kind === 'object') {
            const required = this.field(schemaIndex, 'required');
            if(required >= 0 && this.schema.node(required).children.some(i => !v.keys.includes(this.schema.node(i).text))) { return false; }
            const properties = this.field(schemaIndex, 'properties'); const additional = this.field(schemaIndex, 'additionalProperties');
            for(let i = 0; i < v.keys.length; i++) {
                const name = v.keys[i] ?? panic('key');
                const declared = properties < 0 ? -1 : this.field(properties, name);
                const constraint = declared >= 0 ? declared : additional;
                if(constraint >= 0 && !this.validate(constraint, values, v.children[i] ?? panic('member'), depth + 1)) { return false; }
            }
        }
        for(const group of ['allOf', 'anyOf', 'oneOf']) {
            const branches = this.field(schemaIndex, group);
            if(branches < 0) { continue; }
            let matched = 0;
            const children = this.schema.node(branches).children;
            for(const branch of children) { if(this.validate(branch, values, valueIndex, depth + 1)) { matched++; } }
            if((group === 'allOf' && matched !== children.length) || (group === 'anyOf' && matched === 0) || (group === 'oneOf' && matched !== 1)) { return false; }
        }
        const not = this.field(schemaIndex, 'not');
        return not < 0 || !this.validate(not, values, valueIndex, depth + 1);
    }
    check(options: string): string {
        this.unsupported = this.initialUnsupported;
        if(this.unsupported.length > 0) { return `NotYet: ${this.unsupported}`; }
        const values = new OptionsJson(options); const root = values.parse();
        if(root < 0 || values.node(root).kind !== 'array') { return 'invalid'; }
        const valid = this.validate(this.root, values, root, 0);
        return this.unsupported !== '' ? `NotYet: ${this.unsupported}` : valid ? 'valid' : 'invalid';
    }
}
