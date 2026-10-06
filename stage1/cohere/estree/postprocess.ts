// Go blanks comment bytes, rather than UTF-16 units, to preserve byte positions.
import { panic, utf8Length } from 'adamic';
import type { Arena } from './arena.ts';
import { absent, childValue, stringValue } from './values.ts';
import { locStart, locEnd, shouldAddContentEnd } from './location.ts';
import { StrippedText, byteUnit, unitOffsets, trimEndGo } from './strippedText.ts';
import { visitorKeys } from './visitorKeys.ts';
export function indentableLines(arena: Arena, id: number): string[] {
    const comment = arena.node(id);
    const empty: string[] = [];
    if(comment.type !== 'Block' || !comment.string('value').includes('\n')) {
        return empty;
    }
    const result: string[] = [];
    for(const line of `*${comment.string('value')}*`.split('\n')) {
        const trimmed = line.trimStart();
        if(!trimmed.startsWith('*')) {
            return empty;
        }
        result.push(trimmed);
    }
    return result;
}
export function mergeComments(arena: Arena, comments: number[]): number[] {
    let following = -1;
    for(let index = comments.length - 1; index >= 0; index--) {
        const id = comments[index] ?? panic('missing comment');
        const node = arena.node(id);
        if(
            following >= 0 &&
            locEnd(arena, id) === locStart(arena, following) &&
            indentableLines(arena, id).length > 0 &&
            indentableLines(arena, following).length > 0
        ) {
            const next = arena.node(following);
            // eslint-disable-next-line nexus/correctness-no-caller-data-mutation -- Go merges the shared comment list in place before attachment.
            comments.splice(index + 1, 1);
            node.set('value', stringValue(`${node.string('value')}*//*${next.string('value')}`));
            node.end = next.end;
        }
        following = id;
    }
    return comments;
}
export class Postprocessor {
    readonly arena: Arena;
    readonly text: string;
    readonly offsets: readonly number[];
    readonly stripped: StrippedText;
    constructor(arena: Arena, text: string, comments: readonly number[], offsets: readonly number[]) {
        this.arena = arena;
        this.text = text;
        this.offsets = offsets;
        this.stripped = new StrippedText(text, arena, comments, offsets);
    }
    contentEnd(id: number): void {
        const node = this.arena.node(id);
        if(!shouldAddContentEnd(node.type) || node.end === 0) {
            return;
        }
        const endUnit = byteUnit(this.offsets, node.end);
        if(this.text.slice(endUnit - 1, endUnit) !== ';') {
            return;
        }
        // Strip first, then map the byte slice in that text's own UTF-16 coordinates.
        const stripped = this.stripped.text();
        const offsets = unitOffsets(stripped);
        const before = stripped.slice(byteUnit(offsets, locStart(this.arena, id)), byteUnit(offsets, node.end - 1));
        node.contentEnd = node.end - 1 - (utf8Length(before) - utf8Length(trimEndGo(before)));
        node.hasContentEnd = true;
    }
    unbalanced(id: number): boolean {
        const node = this.arena.node(id);
        const right = node.child('right');
        return (
            node.type === 'LogicalExpression' &&
            this.arena.type(right) === 'LogicalExpression' &&
            node.string('operator') === this.arena.node(right).string('operator')
        );
    }
    logical(start: number, end: number, operator: string, left: number, right: number): number {
        const id = this.arena.newNode('LogicalExpression', start, end);
        const node = this.arena.node(id);
        node.set('operator', stringValue(operator));
        node.set('left', childValue(left));
        node.set('right', childValue(right));
        return id;
    }
    rebalance(id: number): number {
        if(!this.unbalanced(id)) {
            return id;
        }
        const node = this.arena.node(id);
        const right = this.arena.node(node.child('right'));
        const left = this.rebalance(
            this.logical(
                locStart(this.arena, node.child('left')),
                locEnd(this.arena, right.child('left')),
                node.string('operator'),
                node.child('left'),
                right.child('left'),
            ),
        );
        return this.rebalance(
            this.logical(
                locStart(this.arena, id),
                locEnd(this.arena, id),
                node.string('operator'),
                left,
                right.child('right'),
            ),
        );
    }
    visit(id: number): number {
        if(id < 0) {
            return id;
        }
        const node = this.arena.node(id);
        this.contentEnd(id);
        if(node.type === 'TemplateElement') {
            node.start = locStart(this.arena, id) + 1;
            node.end = locEnd(this.arena, id) - (node.bool('tail') ? 1 : 2);
        }
        else if(node.type === 'TSParenthesizedType') {
            return this.visit(node.child('typeAnnotation'));
        }
        else if(
            (node.type === 'TSUnionType' || node.type === 'TSIntersectionType') &&
            node.list('types').length === 1
        ) {
            return this.visit(node.list('types')[0] ?? -1);
        }
        const keys: readonly string[] = visitorKeys.get(node.type) ?? [];
        for(const key of keys) {
            const value = node.get(key);
            if(value.kind === 'node') {
                node.set(key, childValue(this.visit(value.node)));
            }
            else if(value.kind === 'list') {
                for(let index = 0; index < value.list.length; index++) {
                    value.list[index] = this.visit(value.list[index] ?? -1);
                }
            }
            else if(!node.has(key)) {
                node.set(key, absent());
            }
        }
        return this.rebalance(id);
    }
    program(id: number): number {
        const node = this.arena.node(id);
        node.start = 0;
        node.end = utf8Length(this.text);
        return this.visit(id);
    }
}
