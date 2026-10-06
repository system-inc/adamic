// Additional syntax rules share the parser's index table and the owning findings list.
import { panic } from 'adamic';
import type { Parser } from '../../typescript/parser/parser.ts';
import type { ParseNode } from '../../typescript/parser/nodes.ts';
import type { Scanner } from '../../typescript/scanner/scanner.ts';
import type { Settings } from './settings.ts';
import { Finding } from './finding.ts';
import { policyMessage } from './volume_messages.ts';

export class VolumeRules {
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
    enabled(rule: string): boolean {
        return this.selected === 'all' || this.selected === rule;
    }
    start(index: number): number {
        this.scanner.pos = this.node(index).pos;
        this.scanner.scan();
        return this.scanner.start;
    }
    report(index: number, rule: string, id: string, message: string): void {
        this.findings.push(new Finding(rule, id, message, this.start(index), this.node(index).end, '', '', ''));
    }
    name(index: number): number {
        for(const child of this.node(index).children) {
            const kind = this.node(child).kind;
            if(
                kind === 'Identifier' ||
                kind === 'StringLiteral' ||
                kind === 'ComputedPropertyName' ||
                kind === 'PrivateIdentifier'
            ) {
                return child;
            }
        }
        return -1;
    }
    unwrap(index: number): number {
        let current = index;
        while(this.node(current).kind === 'ParenthesizedExpression') {
            current = this.node(current).children[0] ?? panic('empty parentheses');
        }
        return current;
    }
    initializer(index: number): number {
        const children = this.node(index).children;
        // The parser records initializer presence separately from optional type children.
        const last = children[children.length - 1] ?? -1;
        return last >= 0 && this.source[this.node(last).pos - 1] === '=' ? last : -1;
    }
    constEnum(index: number): boolean {
        if(index < 0 || this.node(index).kind !== 'AsExpression') {
            return false;
        }
        const children = this.node(index).children;
        const expression = children[0] ?? -1;
        const type = children[1] ?? -1;
        if(expression < 0 || type < 0 || this.node(type).kind !== 'TypeReference') {
            return false;
        }
        const typeName = this.node(type).children[0] ?? -1;
        if(
            typeName < 0 ||
            this.node(typeName).text !== 'const' ||
            this.node(expression).kind !== 'ObjectLiteralExpression' ||
            this.node(expression).children.length === 0
        ) {
            return false;
        }
        for(const property of this.node(expression).children) {
            if(this.node(property).kind !== 'PropertyAssignment') {
                return false;
            }
            const key = this.node(property).children[0] ?? -1;
            const value = this.node(property).children[1] ?? -1;
            if(
                key < 0 ||
                value < 0 ||
                this.node(key).kind !== 'Identifier' ||
                this.node(value).kind !== 'StringLiteral' ||
                this.node(key).text !== this.node(value).text
            ) {
                return false;
            }
        }
        return true;
    }
    afterthought(index: number): boolean {
        let current = index;
        for(;;) {
            const parent = this.parents[current] ?? -1;
            if(parent < 0) {
                return false;
            }
            const node = this.node(parent);
            if(
                node.kind === 'ParenthesizedExpression' ||
                (node.kind === 'BinaryExpression' &&
                    this.node(node.children[1] ?? panic('operator')).kind === 'CommaToken')
            ) {
                current = parent;
                continue;
            }
            if(node.kind !== 'ForStatement') {
                return false;
            }
            // Header-level semicolon count distinguishes incrementor from initializer and test.
            const limit = this.start(current);
            this.scanner.pos = this.start(parent);
            let depth = 0;
            let count = 0;
            while(this.scanner.scan() !== 'EndOfFile' && this.scanner.start < limit) {
                if(this.scanner.kind === 'OpenParenToken') {
                    depth++;
                }
                if(this.scanner.kind === 'CloseParenToken') {
                    depth--;
                }
                if(this.scanner.kind === 'SemicolonToken' && depth === 1) {
                    count++;
                }
            }
            return count === 2 && node.children[node.children.length - 1] !== current;
        }
    }
    visit(index: number): void {
        const node = this.node(index);
        if(node.kind === 'TypePredicate' && this.enabled('adamic/no-type-predicate')) {
            this.report(
                index,
                'adamic/no-type-predicate',
                'typePredicate',
                policyMessage('adamic/no-type-predicate', 'typePredicate', []),
            );
        }
        if(
            (node.kind === 'PropertyAccessExpression' || node.kind === 'ElementAccessExpression') &&
            this.enabled('base/consistency-no-console')
        ) {
            const receiver = this.unwrap(node.children[0] ?? panic('receiver'));
            const key = node.children[node.children.length - 1] ?? panic('property');
            if(this.node(receiver).kind === 'Identifier' && this.node(receiver).text === 'console') {
                this.report(
                    index,
                    'base/consistency-no-console',
                    'consistencyNoConsole',
                    policyMessage('base/consistency-no-console', 'consistencyNoConsole', [
                        'methodName',
                        this.node(key).kind === 'Identifier' ? this.node(key).text : 'log',
                    ]),
                );
            }
        }
        if(
            (node.kind === 'PrefixUnaryExpression' || node.kind === 'PostfixUnaryExpression') &&
            (node.operator === 'PlusPlusToken' || node.operator === 'MinusMinusToken') &&
            this.enabled('no-plusplus')
        ) {
            if(!(this.settings.read('allowforloopafterthoughts', 'false') === 'true' && this.afterthought(index))) {
                const operator = node.operator === 'MinusMinusToken' ? '--' : '++';
                this.report(
                    index,
                    'no-plusplus',
                    'unexpectedUnaryOp',
                    `Unary operator \`${operator}\` used. It mutates its operand and evaluates to a different value depending on which side it sits, so \`a[i++]\` and \`a[++i]\` read almost identically and index different elements. Automatic semicolon insertion compounds it: a line ending in a value followed by a line starting with \`++\` joins into one statement. Write the assignment out as \`x += 1\`, which does one thing and reads the same wherever it appears.`,
                );
            }
        }
        if(
            (node.kind === 'TypeAliasDeclaration' ||
                node.kind === 'InterfaceDeclaration' ||
                node.kind === 'VariableDeclaration') &&
            this.enabled('nexus/consistency-require-type-suffix')
        ) {
            const name = this.name(index);
            if(name >= 0) {
                const text = this.node(name).text;
                const suffixes =
                    node.kind === 'InterfaceDeclaration'
                        ? ['Interface', 'Properties', 'Options']
                        : ['Type', 'Properties', 'Interface', 'Options'];
                const id =
                    node.kind === 'VariableDeclaration'
                        ? 'noConstEnumSuffix'
                        : node.kind === 'InterfaceDeclaration'
                          ? 'noInterfaceSuffix'
                          : 'noTypeAliasSuffix';
                const needs =
                    node.kind === 'VariableDeclaration'
                        ? this.node(name).kind === 'Identifier' &&
                          this.constEnum(this.initializer(index)) &&
                          !text.endsWith('Kind')
                        : !suffixes.some((suffix) => text.endsWith(suffix));
                if(needs) {
                    this.report(
                        name,
                        'nexus/consistency-require-type-suffix',
                        id,
                        policyMessage('nexus/consistency-require-type-suffix', id, ['name', text]),
                    );
                }
            }
        }
    }
}
