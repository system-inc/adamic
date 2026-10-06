// Syntax indexes and source spans shared by the independent rule visitors.
import { panic } from 'adamic';
import type { Parser } from '../../typescript/parser/parser.ts';
import type { ParseNode } from '../../typescript/parser/nodes.ts';
import type { Scanner } from '../../typescript/scanner/scanner.ts';
import type { Settings } from './settings.ts';
import { Finding } from './finding.ts';

export class RuleContext {
    readonly source: string;
    readonly parser: Parser;
    readonly scanner: Scanner;
    readonly parents: readonly number[];
    readonly findings: Finding[];
    readonly selected: string;
    readonly settings: Settings;
    constructor(
        source: string,
        parser: Parser,
        scanner: Scanner,
        parents: readonly number[],
        findings: Finding[],
        selected: string,
        settings: Settings,
    ) {
        this.source = source;
        this.parser = parser;
        this.scanner = scanner;
        this.parents = parents;
        this.findings = findings;
        this.selected = selected;
        this.settings = settings;
    }
    node(index: number): ParseNode {
        return this.parser.node(index);
    }
    parent(index: number): number {
        return this.parents[index] ?? -1;
    }
    enabled(rule: string): boolean {
        return this.selected === 'all' || this.selected === rule;
    }
    start(index: number): number {
        this.scanner.pos = this.node(index).pos;
        this.scanner.scan();
        return this.scanner.start;
    }
    scanAt(position: number): string {
        this.scanner.pos = position;
        return this.scanner.scan();
    }
    scanStart(): number {
        return this.scanner.start;
    }
    scanEnd(): number {
        return this.scanner.pos;
    }
    unwrap(index: number): number {
        let current = index;
        while(current >= 0 && this.node(current).kind === 'ParenthesizedExpression')
            current = this.node(current).children[0] ?? -1;
        return current;
    }
    hasKind(index: number, kind: string): boolean {
        return this.node(index).children.some((child) => this.node(child).kind === kind);
    }
    functionLike(index: number): boolean {
        return [
            'FunctionDeclaration',
            'FunctionExpression',
            'ArrowFunction',
            'MethodDeclaration',
            'Constructor',
            'GetAccessor',
            'SetAccessor',
        ].includes(this.node(index).kind);
    }
    loop(index: number): boolean {
        return ['WhileStatement', 'DoStatement', 'ForStatement', 'ForInStatement', 'ForOfStatement'].includes(
            this.node(index).kind,
        );
    }
    comment(index: number): boolean {
        const opening = this.source.indexOf('{', this.start(index));
        this.scanner.pos = opening + 1;
        this.scanner.scan();
        const trivia = this.source.slice(opening + 1, this.scanner.start);
        return trivia.includes('//') || trivia.includes('/*');
    }
    accessName(index: number): string {
        const node = this.node(index);
        if(node.kind !== 'PropertyAccessExpression' && node.kind !== 'ElementAccessExpression') return '';
        const key = node.children[node.children.length - 1] ?? panic('missing property');
        const kind = this.node(key).kind;
        return (
            node.kind === 'PropertyAccessExpression'
                ? kind === 'Identifier' || kind === 'PrivateIdentifier'
                : kind === 'StringLiteral' || kind === 'NoSubstitutionTemplateLiteral' || kind === 'NumericLiteral'
        )
            ? this.node(key).text
            : '';
    }
    comparison(index: number): string {
        if(index < 0 || this.node(index).kind !== 'BinaryExpression') return '';
        switch(this.node(this.node(index).children[1] ?? panic('operator')).kind) {
            case 'EqualsEqualsToken':
                return '==';
            case 'EqualsEqualsEqualsToken':
                return '===';
            case 'ExclamationEqualsToken':
                return '!=';
            case 'ExclamationEqualsEqualsToken':
                return '!==';
            case 'LessThanToken':
                return '<';
            case 'LessThanEqualsToken':
                return '<=';
            case 'GreaterThanToken':
                return '>';
            case 'GreaterThanEqualsToken':
                return '>=';
            default:
                return '';
        }
    }
    assignment(index: number): boolean {
        if(index < 0 || this.node(index).kind !== 'BinaryExpression') return false;
        return [
            'EqualsToken',
            'PlusEqualsToken',
            'MinusEqualsToken',
            'AsteriskEqualsToken',
            'AsteriskAsteriskEqualsToken',
            'SlashEqualsToken',
            'PercentEqualsToken',
            'LessThanLessThanEqualsToken',
            'GreaterThanGreaterThanEqualsToken',
            'GreaterThanGreaterThanGreaterThanEqualsToken',
            'AmpersandEqualsToken',
            'BarEqualsToken',
            'CaretEqualsToken',
            'BarBarEqualsToken',
            'AmpersandAmpersandEqualsToken',
            'QuestionQuestionEqualsToken',
        ].includes(this.node(this.node(index).children[1] ?? panic('operator')).kind);
    }
    initializer(index: number): number {
        const children = this.node(index).children;
        const last = children[children.length - 1] ?? -1;
        return last >= 0 && this.source[this.node(last).pos - 1] === '=' ? last : -1;
    }
    tokensBetween(start: number, end: number): string {
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
            result += this.tokensBetween(cursor, this.start(child));
            result += `${this.signature(child)},`;
            cursor = this.node(child).end;
        }
        result +=
            node.children.length === 0 ? this.source.slice(start, node.end) : this.tokensBetween(cursor, node.end);
        return `${result})`;
    }
    tokens(index: number): string {
        const node = this.node(index);
        if(
            (node.kind === 'ArrayLiteralExpression' || node.kind === 'ObjectLiteralExpression') &&
            node.children.length === 0
        )
            return node.kind;
        return this.signature(index);
    }
    report(
        index: number,
        rule: string,
        id: string,
        message: string,
        repair = '',
        replacement = '',
        suggestion = '',
    ): Finding {
        const finding = new Finding(
            rule,
            id,
            message,
            this.start(index),
            this.node(index).end,
            repair,
            replacement,
            suggestion,
        );
        this.findings.push(finding);
        return finding;
    }
    replaceRange(finding: Finding, start: number, end: number): void {
        const replacement = new Finding(
            finding.rule,
            finding.id,
            finding.message,
            start,
            end,
            finding.repair,
            finding.replacement,
            finding.suggestion,
        );
        replacement.editStart = finding.editStart;
        replacement.editEnd = finding.editEnd;
        this.findings[this.findings.length - 1] = replacement;
    }
}
