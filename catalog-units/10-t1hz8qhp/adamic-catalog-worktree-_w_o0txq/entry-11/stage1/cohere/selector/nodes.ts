// postcss-selector-parser 2.2.3's own properties, with children and parents as
// indexes into a table, as in the GraphQL slice. No strong ownership cycle.
import { panic, utf8Length } from 'adamic';

import type { Source } from './source.ts';

export class SelectorNode {
    readonly type: string;
    value: string | undefined;
    source: Source | undefined;
    sourceIndex: number | undefined;
    before = '';
    after = '';
    // gap 2: the namespace's true value is a separate field.
    namespace: string | undefined = undefined;
    namespaceTrue = false;
    attribute: string | undefined = undefined;
    operator: string | undefined = undefined;
    insensitive = false;
    // gap 3: represent an absent boolean with a presence bit.
    quoted = false;
    quotedPresent = false;
    rawInsensitive: string | undefined = undefined;
    unquoted: string | undefined = undefined;
    trailingComma = false;
    readonly children: number[] = [];
    parent = -1;
    constructor(type: string, value: string | undefined, source: Source | undefined, sourceIndex: number | undefined) {
        this.type = type;
        this.value = value;
        this.source = source;
        this.sourceIndex = sourceIndex;
    }
}

export function quote(text: string): string {
    const parts: string[] = ['"'];
    let start = 0;
    for(let index = 0; index < text.length; index++) {
        const code = text.charCodeAt(index);
        let escape = '';
        if(code === 34) {
            escape = '\\"';
        }
        else if(code === 92) {
            escape = '\\\\';
        }
        else if(code === 8) {
            escape = '\\b';
        }
        else if(code === 9) {
            escape = '\\t';
        }
        else if(code === 10) {
            escape = '\\n';
        }
        else if(code === 12) {
            escape = '\\f';
        }
        else if(code === 13) {
            escape = '\\r';
        }
        else if(code < 32 || code === 0x2028 || code === 0x2029) {
            escape = `\\u${code.toString(16).padStart(4, '0')}`;
        }
        else if(
            code >= 0xd800 &&
            code <= 0xdbff &&
            text.charCodeAt(index + 1) >= 0xdc00 &&
            text.charCodeAt(index + 1) <= 0xdfff
        ) {
            index++;
        }
        else if(code >= 0xd800 && code <= 0xdfff) {
            escape = '�';
        }
        if(escape !== '') {
            parts.push(text.slice(start, index));
            parts.push(escape);
            start = index + 1;
        }
    }
    parts.push(text.slice(start));
    parts.push('"');
    return parts.join('');
}

function numeric(value: number | undefined): string {
    return value === undefined ? '"<undefined>"' : Number.isNaN(value) ? 'null' : `${value}`;
}

// Cohere converts UTF-16 positions to bytes, including a position in the middle
// of a surrogate pair and a splitWord position beyond the end of the input.
export function convertSourceIndexes(nodes: readonly SelectorNode[], text: string): void {
    const offsets: number[] = [];
    let bytes = 0;
    for(const character of text) {
        offsets.push(bytes);
        if(character.length === 2) {
            offsets.push(bytes);
        }
        bytes += utf8Length(character);
    }
    offsets.push(bytes);
    for(const node of nodes) {
        const index = node.sourceIndex;
        if(index !== undefined && !Number.isNaN(index) && index >= 0) {
            node.sourceIndex =
                index < offsets.length
                    ? (offsets[index] ?? panic('missing offset'))
                    : bytes + index - offsets.length + 1;
        }
    }
}

// Canonical JSON of every own property, including own undefined properties.
export function render(nodes: readonly SelectorNode[], index: number, text: string): string {
    const node = nodes[index] ?? panic('missing selector node');
    const fields = new Map<string, string>();
    fields.set('type', quote(node.type));
    fields.set('spaces', `{"after":${quote(node.after)},"before":${quote(node.before)}}`);
    if(node.type === 'root' || node.type === 'selector' || node.type === 'pseudo') {
        const children: string[] = [];
        for(const child of node.children) {
            children.push(render(nodes, child, text));
        }
        fields.set('nodes', `[${children.join(',')}]`);
    }
    if(node.type !== 'root' && node.type !== 'selector') {
        fields.set('value', node.value === undefined ? '"<undefined>"' : quote(node.value));
        const source = node.source ?? panic('missing source');
        fields.set(
            'source',
            `{"end":{"column":${numeric(source.endColumn)},"line":${numeric(source.endLine)}},"start":{"column":${numeric(source.startColumn)},"line":${numeric(source.startLine)}}}`,
        );
        const offset = node.sourceIndex;
        fields.set('sourceIndex', numeric(offset));
    }
    if(node.namespace !== undefined) {
        fields.set('namespace', quote(node.namespace));
    }
    if(node.namespaceTrue) {
        fields.set('namespace', 'true');
    }
    if(node.type === 'attribute') {
        fields.set('attribute', quote(node.attribute ?? ''));
        fields.set('operator', node.operator === undefined ? '"<undefined>"' : quote(node.operator));
        const raws: string[] = [];
        if(node.rawInsensitive !== undefined) {
            raws.push(`"insensitive":${quote(node.rawInsensitive)}`);
        }
        if(node.unquoted !== undefined) {
            raws.push(`"unquoted":${quote(node.unquoted)}`);
        }
        fields.set('raws', `{${raws.join(',')}}`);
        if(node.insensitive) {
            fields.set('insensitive', 'true');
        }
        if(node.quotedPresent) {
            fields.set('quoted', node.quoted ? 'true' : 'false');
        }
    }
    if(node.trailingComma) {
        fields.set('trailingComma', 'true');
    }
    // gap 4: collect map keys without Array.from(iterator).
    const keys: string[] = [];
    for(const key of fields.keys()) {
        keys.push(key);
    }
    keys.sort((left, right) => (left < right ? -1 : left > right ? 1 : 0));
    const parts: string[] = [];
    for(const key of keys) {
        parts.push(`${quote(key)}:${fields.get(key) ?? panic('missing field')}`);
    }
    return `{${parts.join(',')}}`;
}
