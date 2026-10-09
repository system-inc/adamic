import { printWhitespace } from '../whitespace.ts';
import { whitespaceFrame } from '../whitespaceCodec.ts';
import { printLeaf } from '../leaves.ts';
import { printRoot, type RootChildInterface } from '../root.ts';
import { printHeading, printParagraph, printSentence, type SentenceChildInterface } from '../structure.ts';
// Component driver: Go AST/doc fixtures supply unported blocks, not production Markdown input.
import { panic, programArguments, readTextFile } from 'adamic';
import { decode, encode } from '../codec.ts';
import { DocumentArena, printDocument } from '../document.ts';
import { printList, type ListBlockInterface, type ListItemInterface } from '../lists.ts';
import { printHTMLBlock } from '../htmlblocks.ts';
import { printCodeBlock } from '../codeblocks.ts';
import { printTable, type TableCellInterface } from '../tables.ts';
import { printQuote } from '../quotes.ts';
import { printWord } from '../../markdowninline/inline.ts';
function integer(text: string): number {
    if(text === '' || text === '-') panic('missing fixture integer');
    let value = 0;
    const negative = text.startsWith('-');
    for(let index = negative ? 1 : 0; index < text.length; index++) {
        const code = text.charCodeAt(index);
        if(code < 48 || code > 57) panic('fixture integer');
        value = value * 10 + code - 48;
    }
    return negative ? -value : value;
}
function numbers(text: string): number[] {
    const values: number[] = [];
    if(text === '') return values;
    for(const field of text.split(',')) values.push(integer(field));
    return values;
}
function documentIDs(text: string, documents: readonly number[]): number[] {
    const values: number[] = [];
    for(const id of numbers(text)) values.push(documents[id] ?? panic('fixture document reference'));
    return values;
}
function blocks(serialized: string, documents: readonly number[]): ListBlockInterface[] {
    const children: ListBlockInterface[] = [];
    if(serialized !== '') {
        for(const block of serialized.split(';')) {
            const parts = block.split(',');
            children.push({
                kind: parts[0] ?? '',
                doc: documents[integer(parts[1] ?? '')] ?? panic('fixture block document'),
                start: integer(parts[2] ?? ''),
                end: integer(parts[3] ?? ''),
                column: integer(parts[4] ?? ''),
                indented: parts[5] === '1',
                ignoreNext: parts[6] === '1',
            });
        }
    }
    return children;
}
const args = programArguments();
const input = readTextFile(args[0] ?? panic('usage: list_probe.ts <component fixtures>'));
if(input.kind === 'Error') panic(input.message);
let arena = new DocumentArena();
let items: ListItemInterface[] = [];
let documents: number[] = [];
let rootChildren: RootChildInterface[] = [];
for(const line of input.text.split('\n')) {
    if(line === '') continue;
    const fields = line.split('\t');
    const kind = fields[0] ?? panic('fixture kind');
    if(kind === 'D')
        documents.push(
            arena.add(
                fields[1] ?? '',
                decode(fields[2] ?? ''),
                integer(fields[3] ?? ''),
                documentIDs(fields[6] ?? '', documents),
                integer(fields[4] ?? ''),
                integer(fields[5] ?? ''),
            ),
        );
    else if(kind === 'S') documents.push(printWhitespace(arena, whitespaceFrame(fields)));
    else if(kind === 'Y')
        documents.push(
            printLeaf(arena, {
                kind: fields[1] ?? '',
                value: decode(fields[2] ?? ''),
                source: decode(fields[3] ?? ''),
                url: decode(fields[4] ?? ''),
                title: decode(fields[5] ?? ''),
                hasTitle: fields[6] === '1',
                label: decode(fields[7] ?? ''),
                referenceType: fields[8] ?? '',
                alt: decode(fields[9] ?? ''),
                originalAlt: decode(fields[10] ?? ''),
                metadata: decode(fields[11] ?? ''),
                flags: integer(fields[12] ?? ''),
                sibling: integer(fields[13] ?? ''),
                children: blocks(fields[14] ?? '', documents),
            }),
        );
    else if(kind === 'J')
        rootChildren.push({
            doc: documents[integer(fields[1] ?? '')] ?? panic('root document'),
            kind: fields[2] ?? '',
            start: integer(fields[3] ?? ''),
            end: integer(fields[4] ?? ''),
            startOffset: integer(fields[5] ?? ''),
            endOffset: integer(fields[6] ?? ''),
            value: decode(fields[7] ?? ''),
        });
    else if(kind === 'O') {
        documents.push(printRoot(arena, decode(fields[1] ?? ''), rootChildren));
        rootChildren = [];
    }
    else if(kind === 'N') {
        const children: SentenceChildInterface[] = [];
        for(const value of (fields[5] ?? '').split(';')) {
            if(value === '') continue;
            const pair = value.split(',');
            children.push({
                doc: documents[integer(pair[0] ?? '')] ?? panic('structural child'),
                whitespace: pair[1] === '1',
            });
        }
        const ids = children.map((child) => child.doc);
        const name = fields[1] ?? '';
        documents.push(
            name === 'sentence'
                ? printSentence(arena, children)
                : name === 'paragraph'
                  ? printParagraph(arena, ids)
                  : printHeading(arena, ids, integer(fields[2] ?? ''), fields[3] === '1', decode(fields[4] ?? '')),
        );
    }
    else if(kind === 'W') {
        const flags = integer(fields[2] ?? '');
        const text = printWord(
            decode(fields[1] ?? ''),
            (flags & 1) !== 0,
            (flags & 2) !== 0,
            (flags & 4) !== 0,
            decode(fields[4] ?? ''),
            decode(fields[5] ?? ''),
        );
        documents.push(arena.text(text));
    }
    else if(kind === 'H')
        documents.push(
            printHTMLBlock(arena, {
                rootLast: fields[1] === '1',
                value: decode(fields[3] ?? ''),
            }),
        );
    else if(kind === 'C')
        documents.push(
            printCodeBlock(arena, {
                indented: fields[1] === '1',
                language: decode(fields[2] ?? ''),
                metadata: decode(fields[3] ?? ''),
                value: decode(fields[7] ?? ''),
            }),
        );
    else if(kind === 'T') {
        const rows: TableCellInterface[][] = [];
        for(const row of (fields[3] ?? '').split(':')) {
            const cells: TableCellInterface[] = [];
            if(row !== '') {
                for(const cell of row.split(';')) {
                    const parts = cell.split(',');
                    cells.push({
                        doc: documents[integer(parts[0] ?? '')] ?? panic('table cell document'),
                    });
                }
            }
            rows.push(cells);
        }
        documents.push(printTable(arena, { rows, align: (fields[1] ?? '').split(',') }, 120, 4));
    }
    else if(kind === 'Q') documents.push(printQuote(arena, blocks(fields[1] ?? '', documents)));
    else if(kind === 'I') {
        const children = blocks(fields[6] ?? '', documents);
        items.push({
            marker: decode(fields[1] ?? ''),
            start: integer(fields[2] ?? ''),
            end: integer(fields[3] ?? ''),
            spread: fields[4] === '1',
            checked: integer(fields[5] ?? ''),
            children,
        });
    }
    else if(kind === 'L') {
        const children: ListItemInterface[] = [];
        for(const id of numbers(fields[7] ?? '')) children.push(items[id] ?? panic('fixture list item'));
        documents.push(
            printList(
                arena,
                {
                    ordered: fields[1] === '1',
                    start: integer(fields[2] ?? ''),
                    sibling: integer(fields[3] ?? ''),
                    ancestorAligned: fields[4] === '1',
                    nextIndented: fields[5] === '1',
                    nextCode: decode(fields[6] ?? ''),
                    items: children,
                },
                4,
            ),
        );
    }
    else if(kind === 'R') {
        const text = printDocument(
            arena,
            documents[integer(fields[1] ?? '')] ?? panic('fixture root document'),
            120,
            4,
        );
        console.log(encode((fields[2] === '1' ? '\ufeff' : '') + text));
        arena = new DocumentArena();
        items = [];
        documents = [];
    }
    else panic('fixture opcode');
}
