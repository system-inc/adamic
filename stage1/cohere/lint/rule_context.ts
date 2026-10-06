// Syntax indexes and source spans shared by the independent rule visitors.
import { panic } from 'adamic';
import type { Parser } from '../../typescript/parser/parser.ts';
import type { ParseNode } from '../../typescript/parser/nodes.ts';
import type { Scanner } from '../../typescript/scanner/scanner.ts';
import { Finding } from './finding.ts';
import { ExtraEdit } from './extra_edit.ts';

export class RuleContext {
    readonly source: string;
    readonly parser: Parser;
    readonly scanner: Scanner;
    readonly parents: readonly number[];
    readonly findings: Finding[];
    readonly selected: string;
    constructor(
        source: string,
        parser: Parser,
        scanner: Scanner,
        parents: readonly number[],
        findings: Finding[],
        selected: string,
    ) {
        this.source = source;
        this.parser = parser;
        this.scanner = scanner;
        this.parents = parents;
        this.findings = findings;
        this.selected = selected;
    }
    text(index: number): string {
        return index < 0 ? '' : this.source.slice(this.start(index), this.node(index).end);
    }
    name(index: number): number {
        for(const child of this.node(index).children) {
            if(
                !this.node(child).kind.endsWith('Keyword') &&
                this.node(child).kind !== 'Decorator' &&
                this.node(child).kind !== 'AsteriskToken' &&
                this.node(child).kind !== 'DotDotDotToken'
            )
                return child;
        }
        return -1;
    }
    child(index: number, kind: string): number {
        return this.node(index).children.find((child) => this.node(child).kind === kind) ?? -1;
    }
    body(index: number): number {
        const children = this.node(index).children;
        const last = children[children.length - 1] ?? -1;
        if(last < 0) return -1;
        return this.node(last).kind === 'Block' || this.node(index).kind === 'ArrowFunction' ? last : -1;
    }
    parts(index: number): number[] {
        const name = this.name(index);
        let annotation = -1;
        let initializer = -1;
        let previous = name < 0 ? this.start(index) : this.node(name).end;
        for(const child of this.node(index).children) {
            if(child === name || this.node(child).end <= previous) continue;
            const kind = this.node(child).kind;
            if(kind === 'QuestionToken' || kind === 'ExclamationToken') {
                previous = this.node(child).end;
                continue;
            }
            const token = this.scanAt(previous);
            if(token === 'ColonToken') annotation = child;
            if(token === 'EqualsToken') initializer = child;
            previous = this.node(child).end;
        }
        return [name, annotation, initializer];
    }
    returnType(index: number): number {
        const children = this.node(index).children;
        for(let cursor = children.length - 1; cursor >= 0; cursor--) {
            const child = children[cursor] ?? -1;
            if(child < 0) continue;
            const kind = this.node(child).kind;
            if(kind === 'Block') continue;
            if(
                kind === 'Parameter' ||
                kind === 'TypeParameter' ||
                (child === this.name(index) && !['ConstructSignature', 'CallSignature'].includes(this.node(index).kind))
            )
                return -1;
            return child;
        }
        return -1;
    }
    outer(index: number): number {
        let current = index;
        while(
            current >= 0 &&
            [
                'ParenthesizedExpression',
                'AsExpression',
                'SatisfiesExpression',
                'TypeAssertionExpression',
                'NonNullExpression',
            ].includes(this.node(current).kind)
        ) {
            current = this.node(current).children[this.node(current).kind === 'TypeAssertionExpression' ? 1 : 0] ?? -1;
        }
        return current;
    }
    edit(start: number, end: number, text: string): void {
        const finding = this.findings[this.findings.length - 1] ?? panic('missing finding');
        finding.extraEdits.push(new ExtraEdit(start, end, text));
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
        replacement.editStart = finding.repair === '' ? start : finding.editStart;
        replacement.editEnd = finding.repair === '' ? end : finding.editEnd;
        this.findings[this.findings.length - 1] = replacement;
    }
}
