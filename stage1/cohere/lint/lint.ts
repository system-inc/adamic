// Twenty syntax-only cohere rules. Node indexes keep the visitor's ancestry acyclic.
import { RuleContext, isFunctionKind } from './context.ts';
import { createRuleSet, type RuleSet } from './.generated/registry.ts';
import { panic, utf8Length } from 'adamic';
import { Parser } from '../../typescript/parser/parser.ts';
import type { ParseNode } from '../../typescript/parser/nodes.ts';
import type { Scanner } from '../../typescript/scanner/scanner.ts';

import { matches, quote, selfDirective } from './comments.ts';
import { isLineBreak, isSpace } from '../../typescript/scanner/characters.ts';
import { Finding } from './finding.ts';
import { VolumeRules } from './volume.ts';
import { Scanner as SourceScanner } from '../../typescript/scanner/scanner.ts';
import type { Settings } from './settings.ts';

import {
    messageUnexpectedLabel,
    messageUnexpectedLabelInBreak,
    messageUnexpectedLabelInContinue,
    messageUnicodeBomExpected,
    messageUnicodeBomUnexpected,
    messageUnexpectedCommaExpression,
    messageVars,
} from './messages.ts';

// These hot kind tests use switches instead of allocating lookup arrays.
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
function isParameterListKind(kind: string): boolean {
    switch(kind) {
        case 'FunctionDeclaration':
        case 'FunctionExpression':
        case 'ArrowFunction':
        case 'MethodDeclaration':
        case 'Constructor':
        case 'GetAccessor':
        case 'SetAccessor':
        case 'FunctionType':
        case 'ConstructorType':
        case 'CallSignature':
        case 'ConstructSignature':
        case 'MethodSignature':
        case 'IndexSignature':
            return true;
        default:
            return false;
    }
}
function isBracedListKind(kind: string): boolean {
    switch(kind) {
        case 'ClassDeclaration':
        case 'ClassExpression':
        case 'Block':
        case 'CaseBlock':
        case 'ModuleBlock':
        case 'InterfaceDeclaration':
        case 'EnumDeclaration':
        case 'TypeLiteral':
        case 'ObjectLiteralExpression':
        case 'ObjectBindingPattern':
        case 'NamedImports':
        case 'NamedExports':
            return true;
        default:
            return false;
    }
}

function compareFindings(left: Finding, right: Finding): number {
    const position = left.start - right.start;
    if(position !== 0) {
        return position;
    }
    const labelOrder = (left.rule === 'no-labels' ? -1 : 0) - (right.rule === 'no-labels' ? -1 : 0);
    if(labelOrder !== 0) {
        return labelOrder;
    }
    // Registration collects no-var separately; retain the Go order for tied declarations.
    if(left.rule === 'no-var' && right.rule === 'vars-on-top') {
        return -1;
    }
    if(left.rule === 'vars-on-top' && right.rule === 'no-var') {
        return 1;
    }
    return 0;
}
function compareEdits(left: Finding, right: Finding): number {
    const position = left.editStart - right.editStart;
    if(position !== 0) {
        return position;
    }
    const end = left.editEnd - right.editEnd;
    if(end !== 0) {
        return end;
    }
    return left.rule < right.rule ? -1 : left.rule > right.rule ? 1 : 0;
}

export class Linter {
    readonly source: string;
    readonly parser: Parser;
    readonly scanner: Scanner;
    readonly findings: Finding[] = [];
    readonly rejected: string[] = [];
    literalEnds: number[] = [];
    anchors: boolean[] = [];
    parents: number[] = [];
    root = -1;
    rules: RuleSet | undefined = undefined;
    volume: VolumeRules | undefined = undefined;
    readonly selected: string;
    readonly mode: string;
    readonly nullPolicy: string;
    readonly settings: Settings;
    readonly allowCatch: boolean;
    constructor(
        source: string,
        parser: Parser,
        scanner: Scanner,
        selected: string,
        mode: string,
        nullPolicy: string,
        allowCatch: boolean,
        settings: Settings,
    ) {
        this.settings = settings;
        this.source = source;
        this.parser = parser;
        this.scanner = scanner;
        this.selected = selected === '' ? 'all' : selected;
        this.mode = mode === '' ? 'Always' : mode;
        this.nullPolicy = this.mode === 'Always' ? (nullPolicy === '' ? 'Always' : nullPolicy) : 'Ignore';
        this.allowCatch = allowCatch;
    }
    run(): void {
        this.root = this.parser.file();
        this.parents = this.parser.nodes.map(() => -1);
        this.volume = new VolumeRules(
            this.source,
            this.parser,
            this.scanner,
            this.parents,
            this.findings,
            this.selected,
            this.settings,
        );
        this.ancestry(this.root, -1);
        const context = new RuleContext(
            this.source,
            this.parser,
            this.scanner,
            this.selected,
            this.mode,
            this.nullPolicy,
            this.allowCatch,
            this.parents,
            this.settings,
        );
        this.rules = createRuleSet(context);
        this.rules.prepare(this.root);
        this.volume.prepare(this.root);
        this.walk(this.root, -1);
        this.rules.finish(this.root);
        for(const finding of this.rules.context.findings) {
            this.findings.push(finding);
        }
        this.findings.sort(compareFindings);
    }
    ancestry(index: number, parent: number): void {
        this.parents[index] = parent;
        for(const child of this.node(index).children) {
            this.ancestry(child, index);
        }
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
    ): void {
        this.findings.push(
            new Finding(rule, id, message, this.start(index), this.node(index).end, repair, replacement, suggestion),
        );
    }
    walk(index: number, parent: number): void {
        if(this.rules !== undefined) {
            this.rules.visit(index, parent);
        }
        this.parents[index] = parent;
        const node = this.node(index);
        this.additional(index, parent);
        (this.volume ?? panic('missing additional rules')).visit(index);
        for(const child of node.children) {
            this.walk(child, index);
        }
    }
    varsOnTop(index: number, parent: number): void {
        const subject = parent >= 0 && this.node(parent).kind === 'VariableStatement' ? parent : index;
        const container = this.parents[subject] ?? -1;
        if(container < 0) {
            return;
        }
        const owner = this.parents[container] ?? -1;
        const staticBlock = owner >= 0 && this.node(owner).kind === 'ClassStaticBlockDeclaration';
        if(
            this.node(container).kind === 'SourceFile' ||
            (this.node(container).kind === 'Block' && owner >= 0 && (isFunctionKind(this.node(owner).kind) || staticBlock))
        ) {
            let prefix = !staticBlock;
            for(const statement of this.node(container).children) {
                const node = this.node(statement);
                const expression = node.children[0] ?? -1;
                const directive =
                    node.kind === 'ExpressionStatement' &&
                    expression >= 0 &&
                    this.node(expression).kind === 'StringLiteral';
                if(
                    prefix &&
                    (directive || node.kind === 'ImportDeclaration' || node.kind === 'ImportEqualsDeclaration')
                ) {
                    continue;
                }
                prefix = false;
                if(node.kind !== 'VariableStatement') {
                    break;
                }
                if(statement === subject) {
                    return;
                }
            }
        }
        this.report(subject, 'vars-on-top', 'top', messageVars, '', '', '');
    }
    additional(index: number, parent: number): void {
        const node = this.node(index);
        if(node.kind === 'SourceFile' && this.enabled('no-warning-comments')) {
            this.warnings();
        }
        if(node.kind === 'SourceFile' && this.enabled('unicode-bom')) {
            const hasMark = this.source.startsWith('﻿');
            const require = this.settings.read('require', 'never');
            if(require === 'always' && !hasMark) {
                this.findings.push(
                    new Finding('unicode-bom', 'expected', messageUnicodeBomExpected, 0, 0, 'fix', '﻿', ''),
                );
            }
            if(require === 'never' && hasMark) {
                const finding = new Finding(
                    'unicode-bom',
                    'unexpected',
                    messageUnicodeBomUnexpected,
                    0,
                    0,
                    'fix',
                    '',
                    '',
                );
                finding.editEnd = 1;
                this.findings.push(finding);
            }
        }
        if(
            (node.kind === 'LabeledStatement' || node.kind === 'BreakStatement' || node.kind === 'ContinueStatement') &&
            this.enabled('no-labels')
        ) {
            if(node.kind === 'LabeledStatement' && !this.allowedLabel(index)) {
                this.report(index, 'no-labels', 'unexpectedLabel', messageUnexpectedLabel, '', '', '');
            }
            if((node.kind === 'BreakStatement' || node.kind === 'ContinueStatement') && node.children.length > 0) {
                const name = this.node(node.children[0] ?? panic('label')).text;
                let allowed = false;
                let cursor = parent;
                while(cursor >= 0) {
                    if(
                        this.node(cursor).kind === 'LabeledStatement' &&
                        this.node(this.node(cursor).children[0] ?? panic('label name')).text === name
                    ) {
                        allowed = this.allowedLabel(cursor);
                        break;
                    }
                    cursor = this.parents[cursor] ?? -1;
                }
                if(!allowed) {
                    this.report(
                        index,
                        'no-labels',
                        node.kind === 'BreakStatement' ? 'unexpectedLabelInBreak' : 'unexpectedLabelInContinue',
                        node.kind === 'BreakStatement'
                            ? messageUnexpectedLabelInBreak
                            : messageUnexpectedLabelInContinue,
                        '',
                        '',
                        '',
                    );
                }
            }
        }
        if(
            node.kind === 'BinaryExpression' &&
            this.enabled('no-sequences') &&
            this.comma(index) &&
            !this.comma(parent) &&
            !this.forSlot(index)
        ) {
            const grandparent = this.parents[parent] ?? -1;
            const parens =
                parent >= 0 &&
                this.node(parent).kind === 'ParenthesizedExpression' &&
                !(grandparent >= 0 && this.node(grandparent).kind === 'ArrowFunction');
            if(!(this.settings.read('allowinparentheses', 'true') === 'true' && parens)) {
                let cursor = index;
                while(this.comma(this.node(cursor).children[0] ?? -1)) {
                    cursor = this.node(cursor).children[0] ?? panic('left');
                }
                this.report(
                    this.node(cursor).children[1] ?? panic('comma'),
                    'no-sequences',
                    'unexpectedCommaExpression',
                    messageUnexpectedCommaExpression,
                    '',
                    '',
                    '',
                );
            }
        }
        if(node.kind === 'VariableDeclarationList' && this.enabled('vars-on-top') && node.semantic === '0') {
            this.varsOnTop(index, parent);
        }
    }
    allowedLabel(index: number): boolean {
        const body = this.node(this.node(index).children[1] ?? panic('labeled body')).kind;
        return ['ForStatement', 'ForInStatement', 'ForOfStatement', 'WhileStatement', 'DoStatement'].includes(body)
            ? this.settings.read('allowloop', 'false') === 'true'
            : body === 'SwitchStatement' && this.settings.read('allowswitch', 'false') === 'true';
    }
    comma(index: number): boolean {
        return (
            index >= 0 &&
            this.node(index).kind === 'BinaryExpression' &&
            this.node(this.node(index).children[1] ?? panic('operator')).kind === 'CommaToken'
        );
    }
    forSlot(index: number): boolean {
        let cursor = index;
        let parent = this.parents[cursor] ?? -1;
        while(parent >= 0 && this.node(parent).kind === 'ParenthesizedExpression') {
            cursor = parent;
            parent = this.parents[cursor] ?? -1;
        }
        if(parent < 0 || this.node(parent).kind !== 'ForStatement') {
            return false;
        }
        const children = this.node(parent).children;
        if(children[children.length - 1] === cursor) {
            return false;
        }
        // Count only header-level semicolons; nested function bodies and regex
        // literal spans cannot masquerade as separators.
        this.literalSpans(parent);
        const limit = this.start(cursor);
        this.scanner.pos = this.start(parent);
        let parens = 0;
        let brackets = 0;
        let braces = 0;
        let separators = 0;
        while(this.scanner.scan() !== 'EndOfFile' && this.scanner.start < limit) {
            const literalEnd = this.literalEnds[this.scanner.start] ?? -1;
            if(literalEnd >= 0) {
                this.scanner.pos = literalEnd;
                continue;
            }
            if(this.scanner.kind === 'OpenParenToken') {
                parens++;
            }
            if(this.scanner.kind === 'CloseParenToken') {
                parens--;
            }
            if(this.scanner.kind === 'OpenBracketToken') {
                brackets++;
            }
            if(this.scanner.kind === 'CloseBracketToken') {
                brackets--;
            }
            if(this.scanner.kind === 'OpenBraceToken') {
                braces++;
            }
            if(this.scanner.kind === 'CloseBraceToken') {
                braces--;
            }
            if(this.scanner.kind === 'SemicolonToken' && parens === 1 && brackets === 0 && braces === 0) {
                separators++;
            }
        }
        return separators === 0 || separators === 2;
    }
    literalSpans(index: number): void {
        if(this.literalEnds.length === 0) {
            // Positions form a bounded integer domain, so lookup needs no hash table.
            this.literalEnds = new Array<number>(this.source.length + 1).fill(-1);
        }
        const node = this.node(index);
        if(isLiteralPartKind(node.kind)) {
            this.literalEnds[this.start(index)] = node.end;
        }
        for(const child of node.children) {
            this.literalSpans(child);
        }
    }
    commentAnchors(index: number): void {
        const node = this.node(index);
        this.anchors[node.pos] = true;
        this.anchors[node.end] = true;
        const parameterList = isParameterListKind(node.kind);
        const argumentsList = node.kind === 'CallExpression' || node.kind === 'NewExpression';
        const bracedList = isBracedListKind(node.kind);
        const arrayList = node.kind === 'ArrayLiteralExpression' || node.kind === 'ArrayBindingPattern';
        const caseList = node.kind === 'CaseClause' || node.kind === 'DefaultClause';
        if(parameterList || argumentsList || bracedList || arrayList || caseList) {
            const children = new Map<number, number>();
            const start = this.start(index);
            for(const child of node.children) {
                const childStart = Math.max(start, this.node(child).pos);
                children.set(childStart, Math.max(children.get(childStart) ?? 0, this.node(child).end));
            }
            let depth = 0;
            for(let position = start; position < node.end; position++) {
                const childEnd = children.get(position);
                if(childEnd !== undefined && childEnd > position) {
                    position = childEnd - 1;
                    continue;
                }
                const code = this.source.charCodeAt(position);
                if(code === 47) {
                    const next = this.source.charCodeAt(position + 1);
                    if(next === 47) {
                        while(position < node.end && !isLineBreak(this.source.charCodeAt(position))) {
                            position++;
                        }
                        continue;
                    }
                    if(next === 42) {
                        const close = this.source.indexOf('*/', position + 2);
                        position = close < 0 ? node.end : close + 1;
                        continue;
                    }
                }
                const opening =
                    parameterList || argumentsList ? (node.kind === 'IndexSignature' ? 91 : 40) : arrayList ? 91 : 123;
                const closing = opening === 40 ? 41 : opening === 91 ? 93 : 125;
                if(code === opening) {
                    depth++;
                    if(depth === 1) {
                        this.anchors[position + 1] = true;
                    }
                }
                if(code === closing) {
                    depth--;
                }
                if(code === 44 && depth === 1) {
                    this.anchors[position + 1] = true;
                }
                if(caseList && code === 58) {
                    this.anchors[position + 1] = true;
                }
            }
        }
        for(const child of node.children) {
            this.commentAnchors(child);
        }
    }
    warnings(): void {
        const terms = this.settings.list('terms', ['todo', 'fixme', 'xxx']);
        if(terms.length === 0) {
            return;
        }
        this.literalSpans(this.root);
        this.anchors = new Array<boolean>(this.source.length + 1).fill(false);
        this.commentAnchors(this.root);
        this.anchors[0] = true;
        // Only parser owners named by cohere's collectListInteriors contribute
        // empty or trailing-comma list anchors; arbitrary punctuation is not an anchor.
        const reachable = this.reachableComments(this.anchors);
        this.scanWarnings(this.literalEnds, reachable, terms);
    }
    reachableComments(anchors: readonly boolean[]): boolean[] {
        const reachable = new Array<boolean>(this.source.length + 1).fill(false);
        for(let anchor = 0; anchor < anchors.length; anchor++) {
            if(anchors[anchor] !== true) {
                continue;
            }
            let position = anchor;
            while(
                position < this.source.length &&
                (this.source.charCodeAt(position) === 32 ||
                    (this.source.charCodeAt(position) >= 9 && this.source.charCodeAt(position) <= 13))
            ) {
                position++;
            }
            if(anchor === 0 && this.source.startsWith('#!')) {
                while(position < this.source.length && !isLineBreak(this.source.charCodeAt(position))) {
                    position++;
                }
            }
            if(this.source.charCodeAt(position) !== 47) {
                continue;
            }
            for(;;) {
                while(
                    position < this.source.length &&
                    (isSpace(this.source.charCodeAt(position)) || isLineBreak(this.source.charCodeAt(position)))
                ) {
                    position++;
                }
                const opening =
                    this.source.charCodeAt(position) === 47 ? this.source.slice(position, position + 2) : '';
                if(opening !== '//' && opening !== '/*') {
                    break;
                }
                reachable[position] = true;
                if(opening === '//') {
                    while(position < this.source.length && !isLineBreak(this.source.charCodeAt(position))) {
                        position++;
                    }
                }
                else {
                    const close = this.source.indexOf('*/', position + 2);
                    position = close < 0 ? this.source.length : close + 2;
                }
            }
        }
        return reachable;
    }
    scanWarnings(literalEnds: readonly number[], reachable: readonly boolean[], terms: readonly string[]): void {
        for(let cursor = 0; cursor < this.source.length;) {
            const end = literalEnds[cursor] ?? -1;
            if(end >= 0) {
                cursor = end;
                continue;
            }
            if(cursor === 0 && this.source.startsWith('#!')) {
                while(cursor < this.source.length && !isLineBreak(this.source.charCodeAt(cursor))) {
                    cursor++;
                }
                continue;
            }
            const opening = this.source.charCodeAt(cursor) === 47 ? this.source.slice(cursor, cursor + 2) : '';
            if(opening !== '//' && opening !== '/*') {
                cursor++;
                continue;
            }
            const start = cursor;
            cursor += 2;
            const bodyStart = cursor;
            let bodyEnd: number;
            if(opening === '//') {
                while(cursor < this.source.length && !isLineBreak(this.source.charCodeAt(cursor))) {
                    cursor++;
                }
                bodyEnd = cursor;
            }
            else {
                const closing = this.source.indexOf('*/', cursor);
                bodyEnd = closing < 0 ? this.source.length : closing;
                cursor = closing < 0 ? this.source.length : closing + 2;
            }
            if(reachable[start] !== true) {
                continue;
            }
            const value = this.source.slice(bodyStart, bodyEnd);
            if(selfDirective(value)) {
                continue;
            }
            for(const term of terms) {
                if(
                    matches(value, term, this.settings.read('location', 'start'), this.settings.list('decoration', []))
                ) {
                    const message = `This comment opens with \`${term}\`, which marks the code as knowingly unfinished: ${quote(value)}. A warning comment is a note to a future reader that something was deferred, so it should be tracked somewhere that gets read rather than left where only the next editor of this file will find it.`;
                    this.findings.push(
                        new Finding('no-warning-comments', 'unexpectedComment', message, start, cursor, '', '', ''),
                    );
                }
            }
        }
    }
    fixed(): string {
        let current = this.source;
        let findings = this.findings;
        for(let pass = 0; pass < 10; pass++) {
            const proposals = findings.filter((finding) => finding.repair === 'fix');
            proposals.sort(compareEdits);
            const applied: Finding[] = [];
            let previous = -1;
            let winner = '';
            for(const finding of proposals) {
                if(
                    finding.editStart < previous ||
                    (finding.editStart === previous && finding.editStart === finding.editEnd)
                ) {
                    this.rejected.push(
                        `rejected ${finding.rule} ${utf8Length(current.slice(0, finding.editStart))} ${utf8Length(current.slice(0, finding.editEnd))} ${winner} overlaps another fix`,
                    );
                    continue;
                }
                if(finding.editStart < 0 || finding.editEnd < finding.editStart || finding.editEnd > current.length) {
                    panic('invalid fix range');
                }
                if(current.slice(finding.editStart, finding.editEnd) === finding.replacement) {
                    panic('nonprogressing fix');
                }
                applied.push(finding);
                previous = finding.editEnd;
                winner = finding.rule;
            }
            if(applied.length === 0) {
                return current;
            }
            let result = current;
            for(let index = applied.length - 1; index >= 0; index--) {
                const finding = applied[index] ?? panic('missing fix');
                result = result.slice(0, finding.editStart) + finding.replacement + result.slice(finding.editEnd);
            }
            const parser = new Parser(result, 'fixed source');
            const scanner = new SourceScanner(result);
            const next = new Linter(
                result,
                parser,
                scanner,
                this.selected,
                this.mode,
                this.nullPolicy,
                this.allowCatch,
                this.settings,
            );
            next.run();
            current = result;
            findings = next.findings;
        }
        panic('fix pass budget exhausted');
    }
}
