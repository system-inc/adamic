// Fifteen syntax-only cohere rules. Node indexes keep the visitor's ancestry acyclic.
import { panic, utf8Length } from 'adamic';
import { Parser } from '../../typescript/parser/parser.ts';
import type { ParseNode } from '../../typescript/parser/nodes.ts';
import { Scanner } from '../../typescript/scanner/scanner.ts';
import { RuleContext } from './rule_context.ts';
import { visitRules } from './registry.ts';

import { Finding } from './finding.ts';

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
    context: RuleContext | undefined = undefined;
    parents: number[] = [];
    root = -1;
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
    ) {
        this.source = source;
        this.parser = parser;
        this.scanner = scanner;
        this.selected = selected;
        this.mode = mode === '' ? 'Always' : mode;
        this.nullPolicy = this.mode === 'Always' ? (nullPolicy === '' ? 'Always' : nullPolicy) : 'Ignore';
        this.allowCatch = allowCatch;
    }
    run(): void {
        this.root = this.parser.file();
        this.parents = this.parser.nodes.map(() => -1);
        this.indexParents(this.root, -1);
        this.context = new RuleContext(
            this.source,
            this.parser,
            this.scanner,
            this.parents,
            this.findings,
            this.selected,
        );
        this.walk(this.root, -1);
        this.findings.sort((left, right) => left.start - right.start);
    }
    indexParents(index: number, parent: number): void {
        this.parents[index] = parent;
        for(const child of this.node(index).children) this.indexParents(child, index);
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
        const actual = this.source.slice(this.start(operator), this.node(operator).end);
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
            const functionBody = [
                'FunctionDeclaration',
                'FunctionExpression',
                'ArrowFunction',
                'MethodDeclaration',
                'Constructor',
                'GetAccessor',
                'SetAccessor',
            ].includes(parentKind);
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
        if(
            ['VariableStatement', 'ForStatement', 'ForInStatement', 'ForOfStatement'].includes(node.kind) &&
            this.enabled('no-var')
        ) {
            this.variable(index, parent);
        }
        visitRules(this.context ?? panic('missing rule context'), index, node.kind);
        for(const child of node.children) {
            this.walk(child, index);
        }
    }
    fixed(): string {
        let current = this.source;
        let findings = this.findings;
        for(let pass = 0; pass < 10; pass++) {
            const proposals = findings.filter((finding) => finding.repair === 'fix');
            for(const finding of findings) {
                if(finding.repair !== 'fix') continue;
                for(const edit of finding.extraEdits) {
                    const proposal = new Finding(
                        finding.rule,
                        finding.id,
                        finding.message,
                        finding.start,
                        finding.end,
                        'fix',
                        edit.text,
                        '',
                    );
                    proposal.editStart = edit.start;
                    proposal.editEnd = edit.end;
                    proposals.push(proposal);
                }
            }
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
                    this.rejected.push(
                        `rejected ${finding.rule} ${utf8Length(current.slice(0, finding.editStart))} ${utf8Length(current.slice(0, finding.editEnd))}  the fix replaces text with itself`,
                    );
                    continue;
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
            const parser = new Parser(result, this.parser.path);
            const scanner = new Scanner(result);
            const next = new Linter(
                result,
                parser,
                scanner,
                this.selected,
                this.mode,
                this.nullPolicy,
                this.allowCatch,
            );
            next.run();
            current = result;
            findings = next.findings;
        }
        panic('fix pass budget exhausted');
    }
}
