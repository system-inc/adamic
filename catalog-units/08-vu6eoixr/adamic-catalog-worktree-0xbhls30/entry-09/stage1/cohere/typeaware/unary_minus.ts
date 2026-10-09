// The checker supplies constrained union parts; the rule decision stays in Adamic.
import { panic, tsgoTypeParts } from 'adamic';
import type { Parser } from '../../typescript/parser/parser.ts';
import type { Scanner } from '../../typescript/scanner/scanner.ts';
import { Finding } from '../lint/finding.ts';

// Pinned TypeFlags: Any | Never | Number | NumberLiteral | Enum | BigInt | BigIntLiteral.
const numericOrExempt = 1 | 262144 | 64 | 2048 | 65536 | 128 | 4096;

export function byteOffsets(source: string): number[] {
    const offsets: number[] = [0];
    let bytes = 0;
    for(let index = 0; index < source.length; index++) {
        const code = source.codePointAt(index) ?? panic('missing source code point');
        if(code > 65535) {
            offsets.push(bytes);
            index++;
        }
        bytes += code < 128 ? 1 : code < 2048 ? 2 : code < 65536 ? 3 : 4;
        offsets.push(bytes);
    }
    return offsets;
}

export class UnaryMinus {
    readonly program: number;
    readonly path: string;
    readonly parser: Parser;
    readonly scanner: Scanner;
    readonly offsets: readonly number[];
    readonly findings: Finding[] = [];
    queries = 0;
    constructor(program: number, path: string, parser: Parser, scanner: Scanner, offsets: readonly number[]) {
        this.program = program;
        this.path = path;
        this.parser = parser;
        this.scanner = scanner;
        this.offsets = offsets;
    }
    visit(index: number): void {
        const node = this.parser.node(index);
        if(node.kind === 'PrefixUnaryExpression' && node.operator === 'MinusToken') {
            const operand = this.parser.node(node.children[0] ?? panic('minus without operand'));
            const parts = tsgoTypeParts(
                this.program,
                this.path,
                this.offsets[operand.pos] ?? panic('operand start outside source'),
                this.offsets[operand.end] ?? panic('operand end outside source'),
                operand.kind,
            );
            this.queries++;
            let cursor = 0;
            let offending = '';
            let unsafe = false;
            while(cursor < parts.length) {
                const first = parts.indexOf('\n', cursor);
                const second = parts.indexOf('\n', first + 1);
                if(first < cursor || second < first + 1) {
                    panic('invalid checker type frame');
                }
                const flags = Number.parseInt(parts.slice(cursor, first), 10);
                const length = Number.parseInt(parts.slice(first + 1, second), 10);
                cursor = second + 1;
                if(
                    !Number.isInteger(flags) ||
                    flags < 0 ||
                    !Number.isInteger(length) ||
                    length < 0 ||
                    cursor + length > parts.length
                ) {
                    panic('invalid checker type length or flags');
                }
                if(!unsafe && (flags & numericOrExempt) === 0) {
                    unsafe = true;
                    offending = parts.slice(cursor, cursor + length);
                }
                cursor += length;
            }
            if(parts.length === 0) {
                panic('checker returned no type parts');
            }
            if(unsafe) {
                this.scanner.pos = node.pos;
                this.scanner.scan();
                this.findings.push(
                    new Finding(
                        '@typescript-eslint/no-unsafe-unary-minus',
                        'unaryMinus',
                        `Argument of unary negation should be assignable to number | bigint but is ${offending} instead.`,
                        this.offsets[this.scanner.start] ?? panic('finding start outside source'),
                        this.offsets[node.end] ?? panic('finding end outside source'),
                        '',
                        '',
                        '',
                    ),
                );
            }
        }
    }
    walk(index: number): void {
        this.visit(index);
        for(const child of this.parser.node(index).children) {
            this.walk(child);
        }
    }
    run(): void {
        this.walk(this.parser.file());
        this.findings.sort((left, right) => left.start - right.start);
    }
}
