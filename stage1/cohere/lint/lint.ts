// Twenty syntax-only cohere rules. Node indexes keep the visitor's ancestry acyclic.
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
    messageNoUnneededTernaryConditionalExpression,
    messageNoUnneededTernaryConditionalAssignment,
    messageTemplateCurly,
    messageDivRegex,
    messageBitwise,
    messageContinue,
    messageWith,
    messageNew,
    messageSparse,
    messageYield,
    messageAwait,
    messageVars,
} from './messages.ts';

const debuggerMessage =
    'A debugger statement stops execution when devtools are open and does nothing otherwise, so it is a breakpoint written into the source. Shipped, it halts the page for anyone with devtools open and is invisible to everyone else, which is the worst pairing of symptoms for reproducing a report. Set the breakpoint in the debugger instead.';

const blockMessage =
    'This block is empty, which leaves a reader unable to tell a deliberate no-op from a body that was deleted or never written. The distinction matters most in a `catch`, where an empty block is the difference between swallowing an error on purpose and swallowing it by accident. Write a comment saying why nothing happens here, or remove the block.';

const switchMessage =
    'This switch has no clauses at all, so it evaluates its subject and does nothing with the result. Either the arms were removed and the statement outlived them, or they were never written. Add the clauses, or remove the switch and keep whatever the subject expression was doing.';

const varMessage =
    'This is a `var` declaration, which is function-scoped and hoisted rather than block-scoped. The binding exists from the top of the enclosing function, so it is readable above the line that declares it and it leaks out of the block it looks like it belongs to: a `var` inside an `if` is still in scope after the `if` ends. A `var` in a loop is one binding shared by every iteration, which is why a closure created in the loop sees the final value rather than its own. Use `let`, or `const` when nothing reassigns it.';

const duplicateMessage =
    'Duplicate case label: an earlier arm in this switch tests the same expression, so this one can never run. Almost always a clause was copied and its test never updated, which means the body here is dead and the case it was meant to handle falls through to `default`. Change the test to the value this arm was written for, or delete the arm.';

// These hot kind tests use switches instead of allocating lookup arrays.
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
function isVariableContainerKind(kind: string): boolean {
    switch(kind) {
        case 'VariableStatement':
        case 'ForStatement':
        case 'ForInStatement':
        case 'ForOfStatement':
            return true;
        default:
            return false;
    }
}
function isGeneratorKind(kind: string): boolean {
    switch(kind) {
        case 'FunctionDeclaration':
        case 'FunctionExpression':
        case 'MethodDeclaration':
            return true;
        default:
            return false;
    }
}
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

function equalityOperator(kind: string): string {
    switch(kind) {
        case 'EqualsEqualsToken':
            return '==';
        case 'ExclamationEqualsToken':
            return '!=';
        case 'EqualsEqualsEqualsToken':
            return '===';
        case 'ExclamationEqualsEqualsToken':
            return '!==';
        default:
            return '';
    }
}
function bitwiseOperator(kind: string): string {
    switch(kind) {
        case 'CaretToken':
            return '^';
        case 'BarToken':
            return '|';
        case 'AmpersandToken':
            return '&';
        case 'LessThanLessThanToken':
            return '<<';
        case 'GreaterThanGreaterThanToken':
            return '>>';
        case 'GreaterThanGreaterThanGreaterThanToken':
            return '>>>';
        case 'CaretEqualsToken':
            return '^=';
        case 'BarEqualsToken':
            return '|=';
        case 'AmpersandEqualsToken':
            return '&=';
        case 'LessThanLessThanEqualsToken':
            return '<<=';
        case 'GreaterThanGreaterThanEqualsToken':
            return '>>=';
        case 'GreaterThanGreaterThanGreaterThanEqualsToken':
            return '>>>=';
        case 'TildeToken':
            return '~';
        default:
            return '';
    }
}

function compareFindings(left: Finding, right: Finding): number {
    const position = left.start - right.start;
    if(position !== 0) {
        return position;
    }
    return (left.rule === 'no-labels' ? -1 : 0) - (right.rule === 'no-labels' ? -1 : 0);
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
        this.volume.prepare(this.root);
        this.walk(this.root, -1);
        this.findings.sort(compareFindings);
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
    equality(index: number): void {
        const children = this.node(index).children;
        if(children.length !== 3) {
            return;
        }
        const left = children[0] ?? panic('missing left');
        const operator = children[1] ?? panic('missing operator');
        const right = children[2] ?? panic('missing right');
        const actual = equalityOperator(this.node(operator).kind);
        if(actual === '') {
            return;
        }
        const isNull =
            this.node(this.unwrap(left)).kind === 'NullKeyword' || this.node(this.unwrap(right)).kind === 'NullKeyword';
        const hasTypeOf =
            this.node(this.unwrap(left)).kind === 'TypeOfExpression' ||
            this.node(this.unwrap(right)).kind === 'TypeOfExpression';
        const type = this.literalType(left);
        const sameType = type !== '' && type === this.literalType(right);
        let expected: string;
        if(actual === '===' || actual === '!==') {
            if(this.nullPolicy !== 'Never' || !isNull) {
                return;
            }
            expected = actual === '===' ? '==' : '!=';
        }
        else if(actual === '==' || actual === '!=') {
            if(this.mode === 'Smart' && (hasTypeOf || sameType || isNull)) {
                return;
            }
            if(this.nullPolicy !== 'Always' && isNull) {
                return;
            }
            expected = actual === '==' ? '===' : '!==';
        }
        else {
            return;
        }
        const message = `This compares with \`${actual}\`, which coerces its operands to a common type before comparing. That makes the comparison non-transitive, which is the one property every reader assumes a comparison has: \`0 == '0'\` and \`'0' == []\` disagree while \`0 == []\` is true. Write \`${expected}\`, which compares without coercing.`;
        const suggestion = `Use \`${expected}\` instead of \`${actual}\`. This changes what the comparison answers wherever the operands can differ in type, so it is offered rather than applied.`;
        this.report(
            operator,
            'eqeqeq',
            'unexpected',
            message,
            hasTypeOf || sameType ? 'fix' : 'suggestion',
            expected,
            hasTypeOf || sameType ? '' : suggestion,
        );
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
    duplicateCases(index: number): void {
        const seen = new Set<string>();
        for(const clause of this.node(index).children) {
            if(this.node(clause).kind !== 'CaseClause') {
                continue;
            }
            const expression = this.node(clause).children[0] ?? panic('case without expression');
            const signature = this.signature(expression);
            if(seen.has(signature)) {
                this.report(expression, 'no-duplicate-case', 'unexpected', duplicateMessage, '', '', '');
            }
            seen.add(signature);
        }
    }
    variable(index: number, parent: number): void {
        const node = this.node(index);
        if(node.kind === 'VariableStatement') {
            for(const child of node.children) {
                if(this.node(child).kind === 'DeclareKeyword') {
                    return;
                }
            }
            if(parent >= 0 && this.node(parent).kind === 'ModuleBlock') {
                const grandparent = this.parents[parent] ?? -1;
                if(grandparent >= 0 && this.node(grandparent).operator === 'GlobalKeyword') {
                    return;
                }
            }
        }
        for(const child of node.children) {
            const list = this.node(child);
            if(list.kind === 'VariableDeclarationList' && list.semantic === '0') {
                this.report(child, 'no-var', 'unexpectedVar', varMessage, '', '', '');
            }
        }
    }
    walk(index: number, parent: number): void {
        this.parents[index] = parent;
        const node = this.node(index);
        const parentKind = parent < 0 ? '' : this.node(parent).kind;
        if(node.kind === 'DebuggerStatement' && this.enabled('no-debugger')) {
            const removable = ['Block', 'SourceFile', 'ModuleBlock', 'CaseClause', 'DefaultClause'].includes(
                parentKind,
            );
            this.report(index, 'no-debugger', 'unexpectedDebugger', debuggerMessage, removable ? 'fix' : '', '', '');
        }
        if(node.kind === 'Block' && this.enabled('no-empty') && node.children.length === 0) {
            const functionBody = isFunctionKind(parentKind);
            if(!functionBody && !(this.allowCatch && parentKind === 'CatchClause') && !this.comment(index)) {
                this.report(index, 'no-empty', 'unexpectedBlock', blockMessage, '', '', '');
            }
        }
        if(node.kind === 'SwitchStatement') {
            for(const child of node.children) {
                if(this.node(child).kind !== 'CaseBlock') {
                    continue;
                }
                if(this.enabled('no-empty') && this.node(child).children.length === 0 && !this.comment(child)) {
                    this.report(index, 'no-empty', 'unexpectedSwitch', switchMessage, '', '', '');
                }
                if(this.enabled('no-duplicate-case')) {
                    this.duplicateCases(child);
                }
            }
        }
        if(node.kind === 'BinaryExpression' && this.enabled('eqeqeq')) {
            this.equality(index);
        }
        if(isVariableContainerKind(node.kind) && this.enabled('no-var')) {
            this.variable(index, parent);
        }
        this.additional(index, parent);
        (this.volume ?? panic('missing additional rules')).visit(index);
        for(const child of node.children) {
            this.walk(child, index);
        }
    }
    functionLike(index: number): boolean {
        return isFunctionKind(this.node(index).kind);
    }
    hasKind(index: number, kind: string): boolean {
        for(const child of this.node(index).children) {
            if(this.node(child).kind === kind) {
                return true;
            }
        }
        return false;
    }
    assignmentTarget(index: number): boolean {
        let cursor = index;
        for(;;) {
            const parent = this.parents[cursor] ?? -1;
            if(parent < 0) {
                return false;
            }
            const node = this.node(parent);
            const first = node.children[0] ?? -1;
            switch(node.kind) {
                case 'BinaryExpression': {
                    const operator = this.node(node.children[1] ?? panic('missing operator')).kind;
                    return (
                        first === cursor &&
                        [
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
                        ].includes(operator)
                    );
                }
                case 'PrefixUnaryExpression':
                case 'PostfixUnaryExpression':
                    return node.operator === 'PlusPlusToken' || node.operator === 'MinusMinusToken';
                case 'ForInStatement':
                case 'ForOfStatement':
                    return (
                        (first >= 0 && this.node(first).kind === 'AwaitKeyword' ? node.children[1] : first) === cursor
                    );
                case 'ParenthesizedExpression':
                case 'ArrayLiteralExpression':
                case 'SpreadElement':
                case 'NonNullExpression':
                    cursor = parent;
                    break;
                case 'SpreadAssignment':
                    cursor = this.parents[parent] ?? -1;
                    break;
                case 'ShorthandPropertyAssignment':
                    if(first !== cursor) {
                        return false;
                    }
                    cursor = this.parents[parent] ?? -1;
                    break;
                case 'PropertyAssignment':
                    if(first === cursor) {
                        return false;
                    }
                    cursor = this.parents[parent] ?? -1;
                    break;
                default:
                    return false;
            }
            if(cursor < 0) {
                return false;
            }
        }
    }
    containsYield(index: number): boolean {
        if(this.node(index).kind === 'YieldExpression') {
            return true;
        }
        for(const child of this.node(index).children) {
            if(
                [
                    'FunctionDeclaration',
                    'FunctionExpression',
                    'ArrowFunction',
                    'MethodDeclaration',
                    'ClassDeclaration',
                    'ClassExpression',
                ].includes(this.node(child).kind)
            ) {
                continue;
            }
            if(this.containsYield(child)) {
                return true;
            }
        }
        return false;
    }
    awaitedLoop(index: number): boolean {
        let cursor = index;
        for(;;) {
            const parent = this.parents[cursor] ?? -1;
            if(parent < 0) {
                return false;
            }
            const node = this.node(parent);
            if(
                this.functionLike(parent) ||
                node.kind === 'ClassStaticBlockDeclaration' ||
                (node.kind === 'ForOfStatement' && this.hasKind(parent, 'AwaitKeyword'))
            ) {
                return false;
            }
            if(node.kind === 'WhileStatement' || node.kind === 'DoStatement') {
                return true;
            }
            if(node.kind === 'ForOfStatement' || node.kind === 'ForInStatement') {
                if(cursor === node.children[node.children.length - 1] || this.node(cursor).semantic === '6') {
                    return true;
                }
            }
            if(node.kind === 'ForStatement') {
                // Before the initializer there is no semicolon token. Every later
                // immediate child follows one, including omitted-header forms.
                this.scanner.pos = this.start(parent);
                while(this.scanner.scan() !== 'EndOfFile' && this.scanner.start < this.node(cursor).pos) {
                    if(this.scanner.kind === 'SemicolonToken') {
                        return true;
                    }
                }
            }
            cursor = parent;
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
            (this.node(container).kind === 'Block' && owner >= 0 && (this.functionLike(owner) || staticBlock))
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
        if(node.kind === 'ConditionalExpression' && this.enabled('no-unneeded-ternary')) {
            this.ternary(index);
        }
        if(node.kind === 'StringLiteral' && this.enabled('no-template-curly-in-string')) {
            let from = 0;
            for(;;) {
                const opening = node.text.indexOf('${', from);
                if(opening < 0) {
                    break;
                }
                const body = opening + 2;
                if(node.text.indexOf('}', body) > body) {
                    this.report(
                        index,
                        'no-template-curly-in-string',
                        'unexpectedTemplateExpression',
                        messageTemplateCurly,
                        '',
                        '',
                        '',
                    );
                    break;
                }
                from = body;
            }
        }
        if(node.kind === 'RegularExpressionLiteral' && this.enabled('no-div-regex')) {
            const start = this.start(index);
            if(node.end - start >= 2 && this.source[start + 1] === '=') {
                const finding = new Finding(
                    'no-div-regex',
                    'unexpected',
                    messageDivRegex,
                    start,
                    node.end,
                    'fix',
                    '[=]',
                    '',
                );
                finding.editStart = start + 1;
                finding.editEnd = start + 2;
                this.findings.push(finding);
            }
        }
        if((node.kind === 'BinaryExpression' || node.kind === 'PrefixUnaryExpression') && this.enabled('no-bitwise')) {
            const operator = bitwiseOperator(
                node.kind === 'BinaryExpression'
                    ? this.node(node.children[1] ?? panic('operator')).kind
                    : node.operator,
            );
            if(operator !== '') {
                const right = node.children[2] ?? -1;
                const hint =
                    this.settings.read('int32hint', 'false') === 'true' &&
                    operator === '|' &&
                    right >= 0 &&
                    this.node(right).kind === 'NumericLiteral' &&
                    this.node(right).text === '0';
                if(!this.settings.list('allow', []).includes(operator) && !hint) {
                    this.report(
                        index,
                        'no-bitwise',
                        'unexpected',
                        `Unexpected use of '${operator}'. ${messageBitwise}`,
                        '',
                        '',
                        '',
                    );
                }
            }
        }
        if(node.kind === 'ContinueStatement' && this.enabled('no-continue')) {
            this.report(index, 'no-continue', 'unexpected', messageContinue, '', '', '');
        }
        if(node.kind === 'WithStatement' && this.enabled('no-with')) {
            const start = this.start(index);
            this.findings.push(new Finding('no-with', 'noWith', messageWith, start, start + 4, '', '', ''));
        }
        if(node.kind === 'ExpressionStatement' && this.enabled('no-new')) {
            const expression = node.children[0] ?? -1;
            if(expression >= 0 && this.node(this.unwrap(expression)).kind === 'NewExpression') {
                this.report(index, 'no-new', 'noNewStatement', messageNew, '', '', '');
            }
        }
        if(
            node.kind === 'ArrayLiteralExpression' &&
            this.enabled('no-sparse-arrays') &&
            !this.assignmentTarget(index) &&
            this.hasKind(index, 'OmittedExpression')
        ) {
            this.report(index, 'no-sparse-arrays', 'unexpectedSparseArray', messageSparse, '', '', '');
        }
        if(isGeneratorKind(node.kind) && this.enabled('require-yield') && this.hasKind(index, 'AsteriskToken')) {
            for(const child of node.children) {
                if(
                    this.node(child).kind === 'Block' &&
                    this.node(child).children.length > 0 &&
                    !this.containsYield(child)
                ) {
                    this.report(index, 'require-yield', 'missingYield', messageYield, '', '', '');
                }
            }
        }
        if(
            (node.kind === 'AwaitExpression' ||
                (node.kind === 'ForOfStatement' && this.hasKind(index, 'AwaitKeyword')) ||
                (node.kind === 'VariableDeclarationList' && node.semantic === '6')) &&
            this.enabled('no-await-in-loop') &&
            this.awaitedLoop(index)
        ) {
            const subject =
                node.kind === 'VariableDeclarationList' && parent >= 0 && this.node(parent).kind === 'VariableStatement'
                    ? parent
                    : index;
            this.report(subject, 'no-await-in-loop', 'unexpectedAwait', messageAwait, '', '', '');
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
    text(index: number): string {
        return this.source.slice(this.start(index), this.node(index).end);
    }
    precedence(index: number): number {
        const node = this.node(index);
        if(node.kind === 'ParenthesizedExpression') {
            return this.precedence(node.children[0] ?? panic('operand'));
        }
        if(node.kind === 'BinaryExpression') {
            const operator = this.text(node.children[1] ?? panic('operator'));
            if(operator === ',') {
                return 0;
            }
            if(
                [
                    '=',
                    '+=',
                    '-=',
                    '*=',
                    '**=',
                    '/=',
                    '%=',
                    '<<=',
                    '>>=',
                    '>>>=',
                    '&=',
                    '|=',
                    '^=',
                    '&&=',
                    '||=',
                    '??=',
                ].includes(operator)
            ) {
                return 1;
            }
            if(['??', '||'].includes(operator)) {
                return 4;
            }
            if(operator === '&&') {
                return 5;
            }
            if(operator === '|') {
                return 6;
            }
            if(operator === '^') {
                return 7;
            }
            if(operator === '&') {
                return 8;
            }
            if(['==', '!=', '===', '!=='].includes(operator)) {
                return 9;
            }
            if(['<', '<=', '>', '>=', 'in', 'instanceof'].includes(operator)) {
                return 10;
            }
            if(['<<', '>>', '>>>'].includes(operator)) {
                return 11;
            }
            if(['+', '-'].includes(operator)) {
                return 12;
            }
            if(['*', '/', '%'].includes(operator)) {
                return 13;
            }
            if(operator === '**') {
                return 15;
            }
            return -1;
        }
        if(node.kind === 'ConditionalExpression') {
            return 3;
        }
        if(['ArrowFunction', 'YieldExpression'].includes(node.kind)) {
            return 1;
        }
        if(
            [
                'PrefixUnaryExpression',
                'TypeOfExpression',
                'VoidExpression',
                'DeleteExpression',
                'AwaitExpression',
            ].includes(node.kind)
        ) {
            return 16;
        }
        if(node.kind === 'PostfixUnaryExpression') {
            return 17;
        }
        if(node.kind === 'CallExpression') {
            return 18;
        }
        if(node.kind === 'NewExpression') {
            return 19;
        }
        if(['AsExpression', 'SatisfiesExpression'].includes(node.kind)) {
            return 2;
        }
        if(
            [
                'Identifier',
                'ThisKeyword',
                'SuperKeyword',
                'StringLiteral',
                'NumericLiteral',
                'BigIntLiteral',
                'TrueKeyword',
                'FalseKeyword',
                'NullKeyword',
                'RegularExpressionLiteral',
                'NoSubstitutionTemplateLiteral',
                'TemplateExpression',
                'ArrayLiteralExpression',
                'ObjectLiteralExpression',
                'PropertyAccessExpression',
                'ElementAccessExpression',
                'TaggedTemplateExpression',
                'FunctionExpression',
                'ClassExpression',
            ].includes(node.kind)
        ) {
            return 20;
        }
        return -1;
    }
    invert(index: number): string {
        if(this.node(index).kind === 'BinaryExpression') {
            const operator = this.node(index).children[1] ?? panic('operator');
            const spelling = this.text(operator);
            const inverse =
                spelling === '=='
                    ? '!='
                    : spelling === '!='
                      ? '=='
                      : spelling === '==='
                        ? '!=='
                        : spelling === '!=='
                          ? '==='
                          : '';
            if(inverse !== '') {
                return (
                    this.source.slice(this.start(index), this.start(operator)) +
                    inverse +
                    this.source.slice(this.node(operator).end, this.node(index).end)
                );
            }
        }
        return this.precedence(index) < 16 ? `!(${this.text(index)})` : `!${this.text(index)}`;
    }
    alwaysBoolean(index: number): boolean {
        const node = this.node(this.unwrap(index));
        return node.kind === 'PrefixUnaryExpression'
            ? node.operator === 'ExclamationToken'
            : node.kind === 'BinaryExpression' &&
                  ['==', '!=', '===', '!==', '<', '<=', '>', '>=', 'in', 'instanceof'].includes(
                      this.text(node.children[1] ?? panic('operator')),
                  );
    }
    ternary(index: number): void {
        const children = this.node(index).children;
        const test = children[0] ?? panic('test');
        const yes = children[2] ?? panic('consequent');
        const no = children[4] ?? panic('alternate');
        const yesKind = this.node(this.unwrap(yes)).kind;
        const noKind = this.node(this.unwrap(no)).kind;
        if(['TrueKeyword', 'FalseKeyword'].includes(yesKind) && ['TrueKeyword', 'FalseKeyword'].includes(noKind)) {
            let repair = 'fix';
            let replacement: string;
            if(yesKind === noKind) {
                replacement = yesKind === 'TrueKeyword' ? 'true' : 'false';
                if(this.node(this.unwrap(test)).kind !== 'Identifier') {
                    repair = '';
                    replacement = '';
                }
            }
            else if(noKind === 'TrueKeyword') {
                replacement = this.invert(test);
            }
            else {
                replacement = this.alwaysBoolean(test) ? this.text(test) : `!${this.invert(test)}`;
            }
            this.report(
                index,
                'no-unneeded-ternary',
                'unnecessaryConditionalExpression',
                messageNoUnneededTernaryConditionalExpression,
                repair,
                replacement,
                '',
            );
            return;
        }
        const testNode = this.node(this.unwrap(test));
        const yesNode = this.node(this.unwrap(yes));
        if(
            this.settings.read('defaultassignment', 'true') === 'false' &&
            testNode.kind === 'Identifier' &&
            yesNode.kind === 'Identifier' &&
            testNode.text === yesNode.text
        ) {
            let alternate = this.text(no);
            const noNode = this.node(no);
            const coalesce =
                noNode.kind === 'BinaryExpression' && this.text(noNode.children[1] ?? panic('operator')) === '??';
            if(noNode.kind !== 'ParenthesizedExpression' && (this.precedence(no) < 4 || coalesce)) {
                alternate = `(${alternate})`;
            }
            this.report(
                index,
                'no-unneeded-ternary',
                'unnecessaryConditionalAssignment',
                messageNoUnneededTernaryConditionalAssignment,
                'fix',
                `${this.text(test)} || ${alternate}`,
                '',
            );
        }
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
