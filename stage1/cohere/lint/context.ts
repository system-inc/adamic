import { panic } from 'adamic';
import type { Parser } from '../../typescript/parser/parser.ts';
import type { ParseNode } from '../../typescript/parser/nodes.ts';
import type { Scanner } from '../../typescript/scanner/scanner.ts';

import { Finding } from './finding.ts';
import type { Settings } from './settings.ts';

export class RuleContext {
    readonly source: string;
    readonly settings: Settings;
    readonly parser: Parser;
    readonly scanner: Scanner;
    readonly findings: Finding[] = [];
    readonly parents: readonly number[];
    readonly selected: string;
    readonly mode: string;
    readonly nullPolicy: string;
    readonly allowCatch: boolean;
    constructor(
        source: string,
        parser: Parser,
        scanner: Scanner,
        selected: string,
        mode: string,
        nullPolicy: string,
        allowCatch: boolean,
        parents: readonly number[],
        settings: Settings,
    ) {
        this.source = source;
        this.parser = parser;
        this.scanner = scanner;
        this.settings = settings;
        this.selected = selected === '' ? 'all' : selected;
        this.mode = mode === '' ? 'Always' : mode;
        this.nullPolicy = this.mode === 'Always' ? (nullPolicy === '' ? 'Always' : nullPolicy) : 'Ignore';
        this.allowCatch = allowCatch;
        this.parents = parents;
    }
    node(index: number): ParseNode {
        return this.parser.node(index);
    }
    enabled(name: string): boolean {
        return this.selected === 'all' || this.selected === name;
    }
    start(index: number): number {
        const node = this.node(index);
        this.scanner.pos = node.pos;
        this.scanner.scan();
        return this.scanner.start;
    }
    report(
        index: number,
        rule: string,
        id: string,
        message: string,
        repair: string,
        replacement: string,
        suggestion: string,
    ): Finding {
        const finding = new Finding(rule, id, message, this.start(index), this.node(index).end, repair, replacement, suggestion);
        this.findings.push(finding);
        return finding;
    }
    comment(index: number): boolean {
        const end = this.node(index).end;
        const opening = this.source.indexOf('{', this.start(index));
        if(opening < 0 || opening >= end) {
            return false;
        }
        // The next token's trivia is the interior of an empty braced node.
        this.scanner.pos = opening + 1;
        this.scanner.scan();
        const trivia = this.source.slice(opening + 1, this.scanner.start);
        return trivia.includes('//') || trivia.includes('/*');
    }
    unwrap(index: number): number {
        let cursor = index;
        while(this.node(cursor).kind === 'ParenthesizedExpression') {
            cursor = this.node(cursor).children[0] ?? panic('parenthesis without operand');
        }
        return cursor;
    }
    literalType(index: number): string {
        switch(this.node(this.unwrap(index)).kind) {
            case 'StringLiteral':
            case 'NoSubstitutionTemplateLiteral':
                return 'string';
            case 'NumericLiteral':
                return 'number';
            case 'BigIntLiteral':
                return 'bigint';
            case 'TrueKeyword':
            case 'FalseKeyword':
                return 'boolean';
            case 'NullKeyword':
            case 'RegularExpressionLiteral':
                return 'object';
            default:
                return '';
        }
    }
    tokens(start: number, end: number): string {
        this.scanner.pos = start;
        let result = '';
        while(this.scanner.scan() !== 'EndOfFile' && this.scanner.start < end) {
            result += `${this.source.slice(this.scanner.start, this.scanner.pos)} `;
        }
        return result;
    }
    signature(index: number): string {
        const node = this.node(index);
        const start = this.start(index);
        let result = `${node.kind}(`;
        let cursor = start;
        for(const child of node.children) {
            result += this.tokens(cursor, this.start(child));
            result += `${this.signature(child)},`;
            cursor = this.node(child).end;
        }
        result += node.children.length === 0 ? this.source.slice(start, node.end) : this.tokens(cursor, node.end);
        return `${result})`;
    }
}
