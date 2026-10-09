// Cohere table.go, preserved prose layout. Cell and row widths are computed natively.
import { panic } from 'adamic';
import { stringWidth } from './width.ts';
import { printDocument, type DocumentArena } from './document.ts';
export interface TableCellInterface {
    readonly doc: number;
}
export interface TableFrameInterface {
    readonly rows: readonly (readonly TableCellInterface[])[];
    readonly align: readonly string[];
}
interface PrintedCellInterface {
    readonly text: string;
    readonly width: number;
}
function rowText(cells: readonly PrintedCellInterface[], widths: readonly number[], aligns: readonly string[]): string {
    const columns: string[] = [];
    for(let index = 0; index < cells.length; index++) {
        const cell = cells[index] ?? panic('table cell');
        const spaces = (widths[index] ?? panic('table column width')) - cell.width;
        const align = aligns[index] ?? '';
        const before = align === 'right' ? spaces : align === 'center' ? Math.floor(spaces / 2) : 0;
        columns.push(' '.repeat(before) + cell.text + ' '.repeat(spaces - before));
    }
    return `| ${columns.join(' | ')} |`;
}
export function printTable(
    arena: DocumentArena,
    frame: TableFrameInterface,
    printWidth: number,
    tabWidth: number,
): number {
    const widths: number[] = [];
    const contents: PrintedCellInterface[][] = [];
    for(const row of frame.rows) {
        const cells: PrintedCellInterface[] = [];
        for(let index = 0; index < row.length; index++) {
            const cell = row[index] ?? panic('table input cell');
            const text = printDocument(arena, cell.doc, printWidth, tabWidth);
            const width = stringWidth(text);
            while(widths.length <= index) widths.push(3);
            widths[index] = Math.max(widths[index] ?? panic('column maximum'), width);
            cells.push({ text, width });
        }
        contents.push(cells);
    }
    const header = contents[0] ?? panic('table without header');
    const delimiter: string[] = [];
    for(let index = 0; index < widths.length && index < header.length; index++) {
        const align = frame.align[index] ?? '';
        const first = align === 'center' || align === 'left' ? ':' : '-';
        const last = align === 'center' || align === 'right' ? ':' : '-';
        delimiter.push(first + '-'.repeat((widths[index] ?? panic('delimiter width')) - 2) + last);
    }
    const lines: string[] = [rowText(header, widths, frame.align), `| ${delimiter.join(' | ')} |`];
    for(let index = 1; index < contents.length; index++)
        lines.push(rowText(contents[index] ?? panic('table row'), widths, frame.align));
    const parts: number[] = [arena.add('p', '', 0, [])];
    for(let index = 0; index < lines.length; index++) {
        if(index > 0) parts.push(arena.add('h', '', 0, [], 2));
        parts.push(arena.text(lines[index] ?? panic('table output row')));
    }
    return arena.concat(parts);
}
