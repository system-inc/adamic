import { panic } from 'adamic';
import type { Parser } from '../../typescript/parser/parser.ts';
import type { ParseNode } from '../../typescript/parser/nodes.ts';
import type { Scanner } from '../../typescript/scanner/scanner.ts';

import { Finding } from './finding.ts';
import type { Settings } from './settings.ts';
import { SuggestionEdit } from './suggestions.a';
import type { Suggestion } from './suggestions.a';

export const noEdits: readonly SuggestionEdit[] = [];
export const noSuggestions: readonly Suggestion[] = [];

// lineStartsOf is where each line begins, breaking on every line terminator ECMAScript names.
function lineStartsOf(source: string): number[] {
    const starts: number[] = [0];
    for(let position = 0; position < source.length; position++) {
        const code = source.charCodeAt(position);
        if(code === 13) {
            if(source.charCodeAt(position + 1) === 10) {
                position++;
            }
            starts.push(position + 1);
        }
        else if(code === 10 || code === 8232 || code === 8233) {
            starts.push(position + 1);
        }
    }
    return starts;
}

// isFunctionKind is a hot kind test, so it switches rather than allocating a lookup array.
function isFunctionKind(kind: string): boolean {
    switch(kind) {
        case 'FunctionDeclaration':
        case 'FunctionExpression':
        case 'ArrowFunction':
        case 'MethodDeclaration':
        case 'Constructor':
        case 'GetAccessor':
        case 'SetAccessor':
            return true;
        default:
            return false;
    }
}

// isLiteralPartKind is a token whose text is literal content rather than code, so a scan for punctuation
// or comments has to step over it.
function isLiteralPartKind(kind: string): boolean {
    switch(kind) {
        case 'StringLiteral':
        case 'RegularExpressionLiteral':
        case 'NoSubstitutionTemplateLiteral':
        case 'TemplateHead':
        case 'TemplateMiddle':
        case 'TemplateTail':
            return true;
        default:
            return false;
    }
}

export function space(character: string): boolean {
    return character === ' ' || character === '\t' || character === '\r' || character === '\n';
}

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
    readonly root: number;
    // lineStarts is built on the first line() a rule asks for, since most files need none.
    lineStarts: number[] | undefined = undefined;
    // literalEndsByStart is built on the first literalEnds() a rule asks for, once per file however many
    // rules share it.
    literalEndsByStart: number[] | undefined = undefined;
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
        root: number,
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
        this.root = root;
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
    // reportNode and reportRange report a finding with its automatic edit and its suggestions, the shape
    // batch 8's rules were written in (#zmh9v36). A finding carries at most one automatic edit, as the Go
    // oracle's report does; its suggestions may carry several, in order.
    reportNode(
        index: number,
        rule: string,
        id: string,
        message: string,
        edits: readonly SuggestionEdit[] = noEdits,
        suggestions: readonly Suggestion[] = noSuggestions,
    ): Finding {
        return this.reportRange(this.start(index), this.node(index).end, rule, id, message, edits, suggestions);
    }
    reportRange(
        start: number,
        end: number,
        rule: string,
        id: string,
        message: string,
        edits: readonly SuggestionEdit[] = noEdits,
        suggestions: readonly Suggestion[] = noSuggestions,
    ): Finding {
        if(edits.length > 1) {
            panic('a finding carries at most one automatic edit');
        }
        const edit = edits[0];
        const finding = new Finding(
            rule,
            id,
            message,
            start,
            end,
            edit === undefined ? '' : 'fix',
            edit === undefined ? '' : edit.text,
            '',
        );
        if(edit !== undefined) {
            finding.editStart = edit.start;
            finding.editEnd = edit.end;
        }
        for(const suggestion of suggestions) {
            finding.suggestions.push(suggestion);
        }
        this.findings.push(finding);
        return finding;
    }
    // edit replaces a node's tokens, from its first token's start to its end.
    edit(index: number, text: string): SuggestionEdit {
        return new SuggestionEdit(this.start(index), this.node(index).end, text);
    }
    line(position: number): number {
        if(this.lineStarts === undefined) {
            this.lineStarts = lineStartsOf(this.source);
        }
        const starts = this.lineStarts;
        let low = 0;
        let high = starts.length;
        while(low + 1 < high) {
            const middle = Math.floor((low + high) / 2);
            if((starts[middle] ?? 0) <= position) {
                low = middle;
            }
            else {
                high = middle;
            }
        }
        return low;
    }
    // literalEnds maps the start of every string, regular expression and template part in the file to its
    // end, and every other position to -1. Positions form a bounded integer domain, so lookup needs no
    // hash table.
    literalEnds(): readonly number[] {
        if(this.literalEndsByStart === undefined) {
            const ends = new Array<number>(this.source.length + 1).fill(-1);
            this.literalSpans(this.root, ends);
            this.literalEndsByStart = ends;
        }
        return this.literalEndsByStart;
    }
    literalSpans(index: number, ends: number[]): void {
        const node = this.node(index);
        if(isLiteralPartKind(node.kind)) {
            ends[this.start(index)] = node.end;
        }
        for(const child of node.children) {
            this.literalSpans(child, ends);
        }
    }
    child(index: number, position: number): number {
        return this.node(index).children[position] ?? panic('missing child');
    }
    parent(index: number): number {
        return this.parents[index] ?? -1;
    }
    // kind is a node's kind, and '' for no node (an index below zero), so a chain of lookups can end in a
    // comparison rather than a guard.
    kind(index: number): string {
        return index < 0 ? '' : this.node(index).kind;
    }
    functionLike(index: number): boolean {
        return isFunctionKind(this.node(index).kind);
    }
    // has is whether a direct child is of this kind, without collecting the children that are.
    has(index: number, kind: string): boolean {
        for(const child of this.node(index).children) {
            if(this.node(child).kind === kind) {
                return true;
            }
        }
        return false;
    }
    children(index: number, kind: string): number[] {
        return this.node(index).children.filter((child) => this.kind(child) === kind);
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
    scanKind(): string {
        return this.scanner.kind;
    }
    raw(index: number): string {
        return this.source.slice(this.start(index), this.node(index).end);
    }
    // name is a declaration's name child: the first identifier, private identifier, literal or computed key
    // past its modifiers, and -1 when something else comes first.
    name(index: number): number {
        for(const child of this.node(index).children) {
            const kind = this.kind(child);
            if(
                ['Identifier', 'PrivateIdentifier', 'StringLiteral', 'NumericLiteral', 'ComputedPropertyName'].includes(
                    kind,
                )
            ) {
                return child;
            }
            if(
                !kind.endsWith('Keyword') &&
                kind !== 'Decorator' &&
                kind !== 'AsteriskToken' &&
                kind !== 'DotDotDotToken'
            ) {
                return -1;
            }
        }
        return -1;
    }
    memberName(index: number): number {
        return ['PropertyDeclaration', 'MethodDeclaration', 'GetAccessor', 'SetAccessor'].includes(this.kind(index))
            ? this.name(index)
            : -1;
    }
    members(index: number): number[] {
        return this.node(index).children.filter((child) =>
            [
                'PropertyDeclaration',
                'MethodDeclaration',
                'GetAccessor',
                'SetAccessor',
                'Constructor',
                'ClassStaticBlockDeclaration',
                'SemicolonClassElement',
                'IndexSignature',
            ].includes(this.kind(child)),
        );
    }
    arguments(index: number): number[] {
        const node = this.node(index);
        return node.list === 0 ? [] : node.children.slice(node.children.length - node.list);
    }
    property(index: number): number {
        const children = this.node(index).children;
        return children[children.length - 1] ?? panic('access without name');
    }
    questionDot(index: number): boolean {
        return this.has(index, 'QuestionDotToken');
    }
    initializer(index: number): number {
        const children = this.node(index).children;
        const last = children[children.length - 1] ?? -1;
        return last >= 0 && this.source.slice(this.node(last).pos - 1, this.node(last).pos) === '=' ? last : -1;
    }
    extendsClass(index: number): boolean {
        return this.children(index, 'HeritageClause').some((child) => this.node(child).operator === 'ExtendsKeyword');
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
