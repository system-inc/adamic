// Additional syntax rules share the parser's index table and the owning findings list.
import { panic } from 'adamic';
import type { Parser } from '../../typescript/parser/parser.ts';
import type { ParseNode } from '../../typescript/parser/nodes.ts';
import type { Scanner } from '../../typescript/scanner/scanner.ts';
import type { Settings } from './settings.ts';
import { Finding } from './finding.ts';
import { policyMessage, cohereMessage } from './volume_messages.ts';

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
        this.visitSecond(index);
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
    readonly declaredTypes = new Set<string>();
    prepare(index: number): void {
        const kind = this.node(index).kind;
        if(
            kind === 'InterfaceDeclaration' ||
            kind === 'TypeAliasDeclaration' ||
            kind === 'ClassDeclaration' ||
            kind === 'EnumDeclaration'
        ) {
            const name = this.name(index);
            if(name >= 0 && this.node(name).kind === 'Identifier') {
                this.declaredTypes.add(this.node(name).text);
            }
        }
        for(const child of this.node(index).children) {
            this.prepare(child);
        }
    }
    hasKind(index: number, kind: string): boolean {
        return this.node(index).children.some((child) => this.node(child).kind === kind);
    }
    text(index: number): string {
        return this.source.slice(this.start(index), this.node(index).end);
    }
    containsKind(index: number, kind: string): boolean {
        return (
            this.node(index).kind === kind || this.node(index).children.some((child) => this.containsKind(child, kind))
        );
    }
    emit(
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
    returnType(index: number): number {
        const node = this.node(index);
        const last = node.children[node.children.length - 1] ?? -1;
        if(last < 0 || last === this.name(index)) {
            return -1;
        }
        const kind = this.node(last).kind;
        return kind === 'Parameter' ||
            kind === 'TypeParameter' ||
            kind === 'QuestionToken' ||
            kind === 'ReadonlyKeyword'
            ? -1
            : last;
    }
    parametersText(index: number): string {
        const parameters = this.node(index).children.filter((child) => this.node(child).kind === 'Parameter');
        const generics = this.node(index).children.filter((child) => this.node(child).kind === 'TypeParameter');
        const floor = this.start(index);
        const end = this.node(index).end;
        let result = '()';
        if(parameters.length > 0) {
            const first = parameters[0] ?? panic('first parameter');
            const last = parameters[parameters.length - 1] ?? panic('last parameter');
            const opening = this.source.slice(0, this.start(first)).lastIndexOf('(');
            const closing = this.source.indexOf(')', this.node(last).end);
            if(opening < floor || closing < 0 || closing >= end) {
                return '';
            }
            result = this.source.slice(opening, closing + 1);
        }
        if(generics.length > 0) {
            const first = generics[0] ?? panic('first generic');
            const last = generics[generics.length - 1] ?? panic('last generic');
            const opening = this.source.slice(0, this.start(first)).lastIndexOf('<');
            const closing = this.source.indexOf('>', this.node(last).end);
            if(opening < floor || closing < 0 || closing >= end) {
                return '';
            }
            result = this.source.slice(opening, closing + 1) + result;
        }
        return result;
    }
    delimiter(index: number): string {
        const last = this.source[this.node(index).end - 1] ?? '';
        return last === ';' || last === ',' ? last : '';
    }
    signatureKey(index: number): string {
        const name = this.name(index);
        return name < 0 ? '' : this.text(name) + (this.hasKind(index, 'QuestionToken') ? '?' : '');
    }
    methodSignature(index: number): void {
        const rule = '@typescript-eslint/method-signature-style';
        const node = this.node(index);
        const style = this.settings.read('style', 'property');
        const key = this.signatureKey(index);
        if(key === '') {
            return;
        }
        if(style === 'property' && node.kind === 'MethodSignature') {
            let ancestor = this.parents[index] ?? -1;
            let inModule = false;
            while(ancestor >= 0) {
                if(this.node(ancestor).kind === 'ModuleDeclaration') {
                    inModule = true;
                    break;
                }
                ancestor = this.parents[ancestor] ?? -1;
            }
            const type = this.returnType(index);
            const skip = inModule || (type >= 0 && this.containsKind(type, 'ThisType'));
            const message = cohereMessage('messageMethodSignatureStyleErrorMethod');
            if(skip) {
                this.report(index, rule, 'errorMethod', message);
                return;
            }
            const owner = this.parents[index] ?? -1;
            const group =
                owner < 0
                    ? [index]
                    : this.node(owner).children.filter(
                          (child) => this.node(child).kind === 'MethodSignature' && this.signatureKey(child) === key,
                      );
            if(group.length > 1) {
                if(group[0] !== index) {
                    this.report(index, rule, 'errorMethod', message);
                    return;
                }
                const parts: string[] = [];
                for(const member of group) {
                    const parameters = this.parametersText(member);
                    if(parameters === '') {
                        this.report(index, rule, 'errorMethod', message);
                        return;
                    }
                    const returnType = this.returnType(member);
                    parts.push(`(${parameters} => ${returnType < 0 ? 'any' : this.text(returnType)})`);
                }
                const last = group[group.length - 1] ?? panic('last overload');
                const finding = this.emit(
                    index,
                    rule,
                    'errorMethod',
                    message,
                    'fix',
                    `${key}: ${parts.join(' & ')}${this.delimiter(last)}`,
                    '',
                );
                finding.editEnd = this.node(last).end;
                return;
            }
            const parameters = this.parametersText(index);
            if(parameters === '') {
                this.report(index, rule, 'errorMethod', message);
                return;
            }
            this.emit(
                index,
                rule,
                'errorMethod',
                message,
                'fix',
                `${key}: ${parameters} => ${type < 0 ? 'any' : this.text(type)}${this.delimiter(index)}`,
                '',
            );
        }
        if(style !== 'property' && node.kind === 'PropertySignature') {
            const type = this.returnType(index);
            if(type < 0 || this.node(type).kind !== 'FunctionType') {
                return;
            }
            const message = cohereMessage('messageMethodSignatureStyleErrorProperty');
            const parameters = this.parametersText(type);
            if(parameters === '') {
                this.report(index, rule, 'errorProperty', message);
                return;
            }
            const returnType = this.returnType(type);
            const readonly = this.hasKind(index, 'ReadonlyKeyword');
            this.emit(
                index,
                rule,
                'errorProperty',
                message,
                readonly ? 'suggestion' : 'fix',
                `${key}${parameters}: ${returnType < 0 ? 'any' : this.text(returnType)}${this.delimiter(index)}`,
                readonly ? cohereMessage('messageMethodSignatureStyleConvertToMethod') : '',
            );
        }
    }
    enumSelf(declaration: number, index: number): boolean {
        const node = this.node(index);
        let name = '';
        if(node.kind === 'Identifier') {
            name = node.text;
        }
        else if(node.kind === 'PropertyAccessExpression' || node.kind === 'ElementAccessExpression') {
            const receiver = node.children[0] ?? panic('enum receiver');
            const enumName = this.name(declaration);
            if(
                enumName < 0 ||
                this.node(receiver).kind !== 'Identifier' ||
                this.node(receiver).text !== this.node(enumName).text
            ) {
                return false;
            }
            const key = node.children[node.children.length - 1] ?? panic('enum key');
            const kind = this.node(key).kind;
            if(
                node.kind === 'PropertyAccessExpression'
                    ? kind !== 'Identifier'
                    : kind !== 'StringLiteral' && kind !== 'NoSubstitutionTemplateLiteral'
            ) {
                return false;
            }
            name = this.node(key).text;
        }
        if(name === '') {
            return false;
        }
        return this.node(declaration).children.some((child) => {
            if(this.node(child).kind !== 'EnumMember') {
                return false;
            }
            const key = this.node(child).children[0] ?? panic('enum member name');
            return (
                ['Identifier', 'StringLiteral', 'NoSubstitutionTemplateLiteral'].includes(this.node(key).kind) &&
                this.node(key).text === name
            );
        });
    }
    enumAllowed(declaration: number, index: number, part: boolean, bitwise: boolean): boolean {
        const node = this.node(this.unwrap(index));
        const unwrapped = this.unwrap(index);
        if(part && this.enumSelf(declaration, unwrapped)) {
            return true;
        }
        if(
            [
                'StringLiteral',
                'NumericLiteral',
                'BigIntLiteral',
                'RegularExpressionLiteral',
                'NullKeyword',
                'TrueKeyword',
                'FalseKeyword',
                'NoSubstitutionTemplateLiteral',
            ].includes(node.kind)
        ) {
            return true;
        }
        if(node.kind === 'PrefixUnaryExpression') {
            const operand = node.children[0] ?? panic('enum operand');
            if(node.operator === 'PlusToken' || node.operator === 'MinusToken') {
                return this.enumAllowed(declaration, operand, part, bitwise);
            }
            return bitwise && node.operator === 'TildeToken' && this.enumAllowed(declaration, operand, true, bitwise);
        }
        if(node.kind === 'BinaryExpression') {
            const operator = this.node(node.children[1] ?? panic('enum operator')).kind;
            return (
                bitwise &&
                [
                    'AmpersandToken',
                    'CaretToken',
                    'LessThanLessThanToken',
                    'GreaterThanGreaterThanToken',
                    'GreaterThanGreaterThanGreaterThanToken',
                    'BarToken',
                ].includes(operator) &&
                this.enumAllowed(declaration, node.children[0] ?? panic('enum left'), true, bitwise) &&
                this.enumAllowed(declaration, node.children[2] ?? panic('enum right'), true, bitwise)
            );
        }
        return false;
    }
    negated(index: number): boolean {
        const node = this.node(this.unwrap(index));
        return node.kind === 'PrefixUnaryExpression'
            ? node.operator === 'ExclamationToken'
            : node.kind === 'BinaryExpression' &&
                  ['ExclamationEqualsToken', 'ExclamationEqualsEqualsToken'].includes(
                      this.node(node.children[1] ?? panic('negated operator')).kind,
                  );
    }
    returnAssignment(index: number): void {
        const operator = this.node(this.node(index).children[1] ?? panic('assignment operator')).kind;
        if(
            !operator.endsWith('EqualsToken') ||
            [
                'EqualsEqualsToken',
                'EqualsEqualsEqualsToken',
                'ExclamationEqualsToken',
                'ExclamationEqualsEqualsToken',
                'LessThanEqualsToken',
                'GreaterThanEqualsToken',
            ].includes(operator)
        ) {
            return;
        }
        const pos = this.node(index).pos;
        this.scanner.pos = this.node(index).end;
        this.scanner.scan();
        if(
            this.settings.read('option', 'except-parens') !== 'always' &&
            this.source[pos - 1] === '(' &&
            this.source[this.scanner.start] === ')'
        ) {
            return;
        }
        let current = index;
        for(;;) {
            const parent = this.parents[current] ?? -1;
            if(parent < 0) {
                return;
            }
            const kind = this.node(parent).kind;
            if(
                kind.endsWith('Statement') ||
                kind === 'ArrowFunction' ||
                kind === 'FunctionExpression' ||
                kind === 'ClassExpression'
            ) {
                if(
                    kind === 'ReturnStatement' ||
                    (kind === 'ArrowFunction' &&
                        this.node(parent).children[this.node(parent).children.length - 1] === current)
                ) {
                    const id = kind === 'ReturnStatement' ? 'returnAssignment' : 'arrowAssignment';
                    this.report(
                        parent,
                        'no-return-assign',
                        id,
                        cohereMessage(
                            kind === 'ReturnStatement' ? 'messageReturnAssignment' : 'messageArrowAssignment',
                        ),
                    );
                }
                return;
            }
            current = parent;
        }
    }
    visitSecond(index: number): void {
        const node = this.node(index);
        if(
            (node.kind === 'MethodSignature' || node.kind === 'PropertySignature') &&
            this.enabled('@typescript-eslint/method-signature-style')
        ) {
            this.methodSignature(index);
        }
        if(node.kind === 'TypeReference' && this.enabled('@typescript-eslint/no-wrapper-object-types')) {
            const name = node.children[0] ?? -1;
            if(name >= 0 && this.node(name).kind === 'Identifier') {
                const text = this.node(name).text;
                if(
                    ['BigInt', 'Boolean', 'Number', 'Object', 'String', 'Symbol'].includes(text) &&
                    !this.declaredTypes.has(text)
                ) {
                    let parent = this.parents[index] ?? -1;
                    let implementsHeritage = false;
                    for(let depth = 0; depth < 2 && parent >= 0; depth++) {
                        if(
                            this.node(parent).kind === 'HeritageClause' &&
                            this.node(parent).operator === 'ImplementsKeyword'
                        ) {
                            implementsHeritage = true;
                        }
                        parent = this.parents[parent] ?? -1;
                    }
                    this.emit(
                        name,
                        '@typescript-eslint/no-wrapper-object-types',
                        'bannedWrapperObjectType',
                        cohereMessage('messageBannedWrapperObjectType'),
                        implementsHeritage ? '' : 'fix',
                        implementsHeritage ? '' : text.toLowerCase(),
                        '',
                    );
                }
            }
        }
        if(
            node.kind === 'EnumMember' &&
            this.enabled('@typescript-eslint/prefer-literal-enum-member') &&
            node.children.length > 1
        ) {
            const owner = this.parents[index] ?? -1;
            const initializer = node.children[1] ?? panic('enum initializer');
            const bitwise = this.settings.read('allowbitwiseexpressions', 'false') === 'true';
            if(owner >= 0 && !this.enumAllowed(owner, initializer, false, bitwise)) {
                this.report(
                    node.children[0] ?? panic('enum name'),
                    '@typescript-eslint/prefer-literal-enum-member',
                    bitwise ? 'notLiteralOrBitwiseExpression' : 'notLiteral',
                    cohereMessage(
                        bitwise
                            ? 'preferLiteralEnumMemberNotLiteralOrBitwiseMessage'
                            : 'preferLiteralEnumMemberNotLiteralMessage',
                    ),
                );
            }
        }
        if(node.kind === 'EnumDeclaration' && this.enabled('nexus/consistency-no-enum')) {
            this.report(
                this.name(index) < 0 ? index : this.name(index),
                'nexus/consistency-no-enum',
                'noEnum',
                policyMessage('nexus/consistency-no-enum', 'noEnum', []),
            );
        }
        if(
            (node.kind === 'IfStatement' || node.kind === 'ConditionalExpression') &&
            this.enabled('no-negated-condition')
        ) {
            const test = node.children[0] ?? panic('condition');
            const alternate = node.children[2] ?? -1;
            if(
                (node.kind === 'ConditionalExpression' ||
                    (alternate >= 0 && this.node(alternate).kind !== 'IfStatement')) &&
                this.negated(test)
            ) {
                this.report(
                    index,
                    'no-negated-condition',
                    'unexpectedNegated',
                    cohereMessage('messageNoNegatedCondition'),
                );
            }
        }
        if(node.kind === 'BinaryExpression' && this.enabled('no-return-assign')) {
            this.returnAssignment(index);
        }
    }
}
