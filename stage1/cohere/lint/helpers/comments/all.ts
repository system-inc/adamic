import type { Parser } from '../../../../typescript/parser/parser.ts';
import { utf8Length } from 'adamic';
import { isLineBreak, isSpace } from '../../../../typescript/scanner/characters.ts';
import { canBeginAt } from './can_begin_at.ts';
import { collectListInteriors } from './collect_list_interiors.ts';
import { sortByPosition } from './sort_by_position.ts';

import { Comment } from './comment.ts';
// Parser and root must describe one immutable source file; positions are UTF-16
// inside the parser and converted to UTF-8 bytes at this public boundary.
export function all(parser: Parser | undefined, root: number): Comment[] {
    if(parser === undefined) {
        return [];
    }
    const anchors = new Array<boolean>(parser.scanner.text.length + 1).fill(false);
    const pending = [root];
    anchors[0] = true;
    while(pending.length > 0) {
        const index = pending.pop() ?? -1;
        const node = parser.node(index);
        anchors[node.pos] = true;
        anchors[node.end] = true;
        for(const anchor of collectListInteriors(parser, index)) {
            anchors[anchor] = true;
        }
        for(const child of node.children) {
            pending.push(child);
        }
    }
    const ends = new Map<number, number>();
    for(let anchor = 0; anchor < anchors.length; anchor++) {
        if(anchors[anchor] !== true || !canBeginAt(parser.scanner.text, anchor)) {
            continue;
        }
        let position = anchor;
        if(anchor === 0 && parser.scanner.text.startsWith('#!')) {
            while(position < parser.scanner.text.length && !isLineBreak(parser.scanner.text.charCodeAt(position))) {
                position++;
            }
        }
        for(;;) {
            while(
                position < parser.scanner.text.length &&
                (isSpace(parser.scanner.text.charCodeAt(position)) ||
                    isLineBreak(parser.scanner.text.charCodeAt(position)))
            ) {
                position++;
            }
            const opening = parser.scanner.text.slice(position, position + 2);
            if(opening !== '//' && opening !== '/*') {
                break;
            }
            const start = position;
            position += 2;
            if(opening === '//') {
                while(position < parser.scanner.text.length && !isLineBreak(parser.scanner.text.charCodeAt(position))) {
                    position++;
                }
            }
            else {
                const close = parser.scanner.text.indexOf('*/', position);
                position = close < 0 ? parser.scanner.text.length : close + 2;
            }
            ends.set(start, position);
        }
    }
    const offsets = new Array<number>(parser.scanner.text.length + 1).fill(0);
    const lines = new Array<number>(parser.scanner.text.length + 1).fill(0);
    const columns = new Array<number>(parser.scanner.text.length + 1).fill(0);
    let byte = 0;
    let line = 0;
    let column = 0;
    for(let index = 0; index < parser.scanner.text.length;) {
        offsets[index] = byte;
        lines[index] = line;
        columns[index] = column;
        const point = parser.scanner.text.codePointAt(index) ?? 0;
        const width = point > 65535 ? 2 : 1;
        const size = utf8Length(parser.scanner.text.slice(index, index + width));
        byte += size;
        column += size;
        if(isLineBreak(point) && !(point === 13 && parser.scanner.text.charCodeAt(index + 1) === 10)) {
            line++;
            column = 0;
        }
        index += width;
    }
    offsets[parser.scanner.text.length] = byte;
    lines[parser.scanner.text.length] = line;
    columns[parser.scanner.text.length] = column;
    const comments: Comment[] = [];
    for(const [start, end] of ends) {
        comments.push(
            new Comment(
                offsets[start] ?? 0,
                offsets[end] ?? 0,
                parser.scanner.text.slice(start, start + 2) === '/*',
                parser.scanner.text.slice(start, end),
                lines[start] ?? 0,
                columns[start] ?? 0,
                lines[end] ?? 0,
            ),
        );
    }
    sortByPosition(comments);
    return comments;
}
