// yaml 2.9.0 CST shapes. Numeric links keep the arena's ownership graph acyclic.
import type { CollectionItem } from './collectionItem.ts';
export class Token {
    type: string;
    offset: number;
    indent: number;
    source: string;
    indentPresent: boolean;
    message = '';
    start: number[] = [];
    flowStart = -1;
    value = -1;
    end: number[] = [];
    endPresent = false;
    props: number[] = [];
    items: CollectionItem[] = [];
    constructor(type: string, offset: number, indent: number, source: string, indentPresent: boolean) {
        this.type = type;
        this.offset = offset;
        this.indent = indent;
        this.source = source;
        this.indentPresent = indentPresent;
    }
}
export function tokenType(source: string): string {
    switch(source) {
        case '':
        case '\n':
        case '\r\n':
            return 'newline';
        case '\ufeff':
            return 'byte-order-mark';
        case '\x02':
            return 'doc-mode';
        case '\x18':
            return 'flow-error-end';
        case '\x1f':
            return 'scalar';
        case '-':
            return 'seq-item-ind';
        case '?':
            return 'explicit-key-ind';
        case ':':
            return 'map-value-ind';
        case '{':
            return 'flow-map-start';
        case '}':
            return 'flow-map-end';
        case '[':
            return 'flow-seq-start';
        case ']':
            return 'flow-seq-end';
        case ',':
            return 'comma';
        case '---':
            return 'doc-start';
        case '...':
            return 'doc-end';
    }
    switch(source.slice(0, 1)) {
        case ' ':
        case '\t':
            return 'space';
        case '#':
            return 'comment';
        case '%':
            return 'directive-line';
        case '*':
            return 'alias';
        case '&':
            return 'anchor';
        case '!':
            return 'tag';
        case "'":
            return 'single-quoted-scalar';
        case '"':
            return 'double-quoted-scalar';
        case '|':
        case '>':
            return 'block-scalar-header';
    }
    return '';
}
