// Statement and declaration descent follows typescript-go's parser.
import { panic } from 'adamic';
import { Scanner } from '../scanner/scanner.ts';
import type { ParseNode } from './nodes.ts';
import { isReservedKind } from './grammar.ts';
import { keywordSuggestion } from './spelling.ts';
import { accessorAhead, declarationAhead, declarationModifierAhead } from './lookahead.ts';

export interface StatementContextInterface {
    readonly scanner: Scanner;
    readonly path: string;
    readonly roots: number[];
    readonly kind: () => string;
    readonly next: () => void;
    readonly expect: (kind: string) => boolean;
    readonly beginList: (context: string) => void;
    readonly endList: (context: string) => void;
    readonly listElement: (context: string) => boolean;
    readonly listTerminator: (context: string) => boolean;
    readonly recoverList: (context: string) => boolean;
    readonly errorAt: (code: number, start: number, end: number, message: string) => void;
    readonly node: (index: number) => ParseNode;
    readonly make: (kind: string, pos: number, children?: number[]) => number;
    readonly entityName: () => number;
    readonly identifier: () => number;
    readonly bindingIdentifier: () => number;
    readonly token: () => number;
    readonly type: (minimum?: number, conditional?: boolean) => number;
    readonly typeArguments: () => number[];
    readonly typeParameters: () => number[];
    readonly modifiers: (decorators: boolean, permitConst: boolean) => number[];
    readonly parameters: () => number[];
    readonly returnType: () => number;
    readonly bindingName: () => number;
    readonly rootAssignment: () => number;
    readonly rootExpression: () => number;
    readonly peek: () => string;
    readonly nextIdentifierSameLine: () => boolean;
    readonly propertyName: () => number;
    readonly methodBody: (
        pos: number,
        prefix: readonly number[],
        kind: string,
        async: boolean,
        generator: boolean,
    ) => number;
    readonly primary: () => number;
    readonly suffix: (expression: number, call: boolean) => number;
    readonly tupleNameAhead: () => boolean;
    readonly typeLiteral: () => number;
    readonly decorator: () => number;
    readonly getAwait: () => boolean;
    readonly setAwait: (value: boolean) => void;
    readonly getYield: () => boolean;
    readonly setYield: (value: boolean) => void;
    readonly getIn: () => boolean;
    readonly setIn: (value: boolean) => void;
    readonly depth: () => number;
}

export class Statements {
    readonly parser: StatementContextInterface;
    constructor(parser: StatementContextInterface) {
        this.parser = parser;
    }
    statementList(context: string): number[] {
        this.parser.beginList(context);
        const children: number[] = [];
        while(!this.parser.listTerminator(context)) {
            if(this.parser.listElement(context)) {
                children.push(this.statement());
            }
            else if(this.parser.recoverList(context)) {
                break;
            }
        }
        this.parser.endList(context);
        return children;
    }
    block(ignoreMissing = false): number {
        const pos = this.parser.scanner.fullStart;
        if(!this.parser.expect('OpenBraceToken') && !ignoreMissing) {
            return this.parser.make('Block', pos);
        }
        const children = this.statementList('block');
        this.parser.expect('CloseBraceToken');
        const block = this.parser.make('Block', pos, children);
        if(this.parser.kind() === 'EqualsToken') {
            this.parser.errorAt(2809, this.parser.scanner.start, this.parser.scanner.pos, "Declaration or statement expected. This '=' follows a block of statements, so if you intended to write a destructuring assignment, you might need to wrap the whole assignment in parentheses.");
            this.parser.next();
        }
        return block;
    }
    semicolon(): void {
        if(this.parser.kind() === 'SemicolonToken') {
            this.parser.next();
        }
        else if(
            this.parser.kind() !== 'CloseBraceToken' &&
            this.parser.kind() !== 'EndOfFile' &&
            (this.parser.scanner.flags & 1) === 0
        ) {
            this.parser.expect('SemicolonToken');
        }
    }
    expressionSemicolon(expression: number, tokenStart: number): void {
        const node = this.parser.node(expression);
        if(
            node.kind === 'TaggedTemplateExpression' &&
            !['SemicolonToken', 'CloseBraceToken', 'EndOfFile'].includes(this.parser.kind()) &&
            (this.parser.scanner.flags & 1) === 0
        ) {
            const template = this.parser.node(
                node.children[node.children.length - 1] ?? panic('missing tagged template'),
            );
            const scanner = new Scanner(this.parser.scanner.text);
            scanner.pos = template.pos;
            scanner.scan();
            this.parser.errorAt(
                1443,
                scanner.start,
                template.end,
                `Module declaration names may only use ' or " quoted strings.`,
            );
            return;
        }
        if(
            node.kind === 'Identifier' &&
            node.text !== '' &&
            this.parser.kind() !== 'SemicolonToken' &&
            this.parser.kind() !== 'CloseBraceToken' &&
            this.parser.kind() !== 'EndOfFile' &&
            (this.parser.scanner.flags & 1) === 0
        ) {
            if(this.keywordStatementError(node.text, tokenStart, node.end)) {
                return;
            }
            const suggestion = keywordSuggestion(node.text);
            if(suggestion !== '') {
                this.parser.errorAt(
                    1435,
                    tokenStart,
                    node.end,
                    `Unknown keyword or identifier. Did you mean '${suggestion}'?`,
                );
            }
            else if(this.parser.kind() !== 'Unknown') {
                this.parser.errorAt(1434, tokenStart, node.end, 'Unexpected keyword or identifier.');
            }
        }
        else {
            this.semicolon();
        }
    }
    keywordStatementError(text: string, start: number, end: number): boolean {
        if(text === 'declare') {
            return true;
        }
        if(['const', 'let', 'var'].includes(text)) {
            this.parser.errorAt(1440, start, end, 'Variable declaration not allowed at this location.');
            return true;
        }
        if(text === 'is') {
            this.parser.errorAt(
                1228,
                start,
                this.parser.scanner.start,
                'A type predicate is only allowed in return type position for functions and methods.',
            );
            return true;
        }
        if(['interface', 'module', 'namespace', 'type'].includes(text)) {
            const label = text === 'interface' ? 'Interface' : text === 'type' ? 'Type alias' : 'Namespace';
            const empty = this.parser.kind() === (text === 'type' ? 'EqualsToken' : 'OpenBraceToken');
            const code = empty
                ? text === 'interface'
                    ? 1438
                    : text === 'type'
                      ? 1439
                      : 1437
                : text === 'interface'
                  ? 2427
                  : text === 'type'
                    ? 2457
                    : 2819;
            this.parser.errorAt(
                code,
                this.parser.scanner.start,
                this.parser.scanner.pos,
                empty ? `${label} must be given a name.` : `${label} name cannot be '${this.parser.scanner.value}'.`,
            );
            return true;
        }
        return false;
    }
    variableList(): number {
        const pos = this.parser.scanner.fullStart;
        const flags =
            this.parser.kind() === 'ConstKeyword'
                ? 2
                : this.parser.kind() === 'LetKeyword'
                  ? 1
                  : this.parser.kind() === 'UsingKeyword'
                    ? 4
                    : this.parser.kind() === 'AwaitKeyword'
                      ? 6
                      : 0;
        if(flags === 6) {
            this.parser.next();
        }
        this.parser.next();
        const declarations: number[] = [];
        this.parser.beginList('variables');
        while(true) {
            if(this.parser.listElement('variables')) {
                const start = this.parser.scanner.fullStart;
                declarations.push(this.variableDeclaration());
                if(this.parser.kind() === 'CommaToken') {
                    this.parser.next();
                    continue;
                }
                if(this.parser.listTerminator('variables')) {
                    break;
                }
                this.parser.expect('CommaToken');
                if(start === this.parser.scanner.fullStart) {
                    this.parser.next();
                }
                continue;
            }
            if(this.parser.listTerminator('variables') || this.parser.recoverList('variables')) {
                break;
            }
        }
        this.parser.endList('variables');
        const id = this.parser.make('VariableDeclarationList', pos, declarations);
        this.parser.node(id).semantic = `${flags}`;
        return id;
    }
    variableDeclaration(): number {
        const start = this.parser.scanner.fullStart;
        const children = [this.parser.bindingName()];
        if(this.parser.kind() === 'ExclamationToken') {
            children.push(this.parser.token());
        }
        if(this.parser.kind() === 'ColonToken') {
            this.parser.next();
            children.push(this.parser.type());
        }
        if(this.parser.kind() === 'EqualsToken') {
            this.parser.next();
            children.push(this.parser.rootAssignment());
        }
        return this.parser.make('VariableDeclaration', start, children);
    }
    lookahead(offset: number): string {
        const saved = {
            pos: this.parser.scanner.pos,
            start: this.parser.scanner.start,
            fullStart: this.parser.scanner.fullStart,
            kind: this.parser.scanner.kind,
            value: this.parser.scanner.value,
            flags: this.parser.scanner.flags,
            errors: this.parser.scanner.errors.length,
        };
        for(let index = 0; index < offset; index++) {
            this.parser.scanner.scan();
        }
        const result = this.parser.kind();
        this.parser.scanner.pos = saved.pos;
        this.parser.scanner.start = saved.start;
        this.parser.scanner.fullStart = saved.fullStart;
        this.parser.scanner.kind = saved.kind;
        this.parser.scanner.value = saved.value;
        this.parser.scanner.flags = saved.flags;
        this.parser.scanner.errors.splice(saved.errors);
        return result;
    }
    decorator(): number {
        return this.parser.decorator();
    }
    canExportName(): boolean {
        return (
            this.parser.kind() === 'Identifier' ||
            this.parser.kind().endsWith('Keyword') ||
            this.parser.kind() === 'StringLiteral'
        );
    }
    specifiers(importing: boolean): number {
        const pos = this.parser.scanner.fullStart;
        const members: number[] = [];
        if(!this.parser.expect('OpenBraceToken')) {
            return this.parser.make(importing ? 'NamedImports' : 'NamedExports', pos, members);
        }
        this.parser.beginList('specifiers');
        while(true) {
            if(!this.parser.listElement('specifiers')) {
                if(this.parser.listTerminator('specifiers') || this.parser.recoverList('specifiers')) {
                    break;
                }
                continue;
            }
            const start = this.parser.scanner.fullStart;
            let name = this.parser.propertyName();
            let property = -1;
            let typeOnly = false;
            let canAs = true;
            if(this.parser.node(name).kind === 'Identifier' && this.parser.node(name).text === 'type') {
                if(this.parser.kind() === 'AsKeyword') {
                    const first = this.parser.identifier();
                    if(this.parser.kind() === 'AsKeyword') {
                        const second = this.parser.identifier();
                        if(this.canExportName()) {
                            typeOnly = true;
                            property = first;
                            name = this.parser.propertyName();
                        }
                        else {
                            property = name;
                            name = second;
                        }
                        canAs = false;
                    }
                    else if(this.canExportName()) {
                        property = name;
                        name = this.parser.propertyName();
                        canAs = false;
                    }
                    else {
                        typeOnly = true;
                        name = first;
                    }
                }
                else if(this.canExportName()) {
                    typeOnly = true;
                    name = this.parser.propertyName();
                }
            }
            if(canAs && this.parser.kind() === 'AsKeyword') {
                this.parser.next();
                property = name;
                name = this.parser.propertyName();
            }
            const children: number[] = [];
            if(property >= 0) {
                children.push(property);
            }
            children.push(name);
            const member = this.parser.make(importing ? 'ImportSpecifier' : 'ExportSpecifier', start, children);
            this.parser.node(member).semantic = typeOnly ? '1' : '0';
            members.push(member);
            if(this.parser.kind() === 'CommaToken') {
                this.parser.next();
                continue;
            }
            if(this.parser.listTerminator('specifiers')) {
                break;
            }
            this.parser.expect('CommaToken');
            if(start === this.parser.scanner.fullStart) {
                this.parser.next();
            }
        }
        this.parser.endList('specifiers');
        this.parser.expect('CloseBraceToken');
        return this.parser.make(importing ? 'NamedImports' : 'NamedExports', pos, members);
    }
    attributes(): number {
        const pos = this.parser.scanner.fullStart;
        const operator = this.parser.kind();
        this.parser.next();
        return this.attributeBody(pos, operator);
    }
    attributeBody(pos: number, operator: string): number {
        this.parser.expect('OpenBraceToken');
        const multiLine = (this.parser.scanner.flags & 1) !== 0;
        const children: number[] = [];
        let trailing = false;
        while(this.parser.kind() !== 'EndOfFile' && this.parser.kind() !== 'CloseBraceToken') {
            const start = this.parser.scanner.fullStart;
            const name = this.parser.propertyName();
            this.parser.expect('ColonToken');
            const value = this.parser.rootAssignment();
            children.push(this.parser.make('ImportAttribute', start, [name, value]));
            trailing = this.parser.kind() === 'CommaToken';
            if(!trailing) {
                break;
            }
            this.parser.next();
        }
        this.parser.expect('CloseBraceToken');
        const id = this.parser.make('ImportAttributes', pos, children);
        const node = this.parser.node(id);
        node.operator = operator;
        node.list = children.length;
        node.trailing = trailing;
        node.multiLine = multiLine;
        return id;
    }
    moduleSpecifier(): number {
        return this.parser.kind() === 'StringLiteral' ? this.parser.propertyName() : this.parser.rootExpression();
    }
    importDeclaration(pos: number, prefix: readonly number[]): number {
        const children = prefix.slice();
        this.parser.expect('ImportKeyword');
        const clausePos = this.parser.scanner.fullStart;
        let name = -1;
        let phase = 'Unknown';
        if(
            this.parser.kind() === 'Identifier' ||
            (this.parser.kind().endsWith('Keyword') && !isReservedKind(this.parser.kind()))
        ) {
            name = this.parser.identifier();
        }
        if(
            name >= 0 &&
            this.parser.node(name).text === 'type' &&
            (this.parser.kind() !== 'FromKeyword' ||
                this.lookahead(1) === 'FromKeyword' ||
                this.lookahead(1) === 'EqualsToken') &&
            (this.canExportName() || this.parser.kind() === 'OpenBraceToken' || this.parser.kind() === 'AsteriskToken')
        ) {
            phase = 'TypeKeyword';
            name = -1;
            if(this.canExportName()) {
                name = this.parser.identifier();
            }
        }
        else if(
            name >= 0 &&
            this.parser.node(name).text === 'defer' &&
            (this.parser.kind() === 'FromKeyword'
                ? this.lookahead(1) !== 'StringLiteral'
                : this.parser.kind() !== 'CommaToken' && this.parser.kind() !== 'EqualsToken')
        ) {
            phase = 'DeferKeyword';
            name = -1;
            if(this.canExportName()) {
                name = this.parser.identifier();
            }
        }
        if(name >= 0 && this.parser.kind() === 'EqualsToken') {
            children.push(name);
            this.parser.next();
            if(this.parser.kind() === 'RequireKeyword' && this.parser.peek() === 'OpenParenToken') {
                const start = this.parser.scanner.fullStart;
                this.parser.next();
                this.parser.expect('OpenParenToken');
                const path = this.parser.propertyName();
                this.parser.expect('CloseParenToken');
                children.push(this.parser.make('ExternalModuleReference', start, [path]));
            }
            else {
                children.push(this.parser.entityName());
            }
            this.semicolon();
            const id = this.parser.make('ImportEqualsDeclaration', pos, children);
            this.parser.node(id).semantic = phase === 'TypeKeyword' ? '1' : '0';
            return id;
        }
        if(name >= 0 || this.parser.kind() === 'AsteriskToken' || this.parser.kind() === 'OpenBraceToken') {
            const parts: number[] = [];
            if(name >= 0) {
                parts.push(name);
            }
            if(name < 0 || this.parser.kind() === 'CommaToken') {
                if(name >= 0) {
                    this.parser.next();
                }
                if(this.parser.kind() === 'AsteriskToken') {
                    const start = this.parser.scanner.fullStart;
                    this.parser.next();
                    this.parser.expect('AsKeyword');
                    const target = this.parser.identifier();
                    parts.push(this.parser.make('NamespaceImport', start, [target]));
                }
                else {
                    parts.push(this.specifiers(true));
                }
            }
            const clause = this.parser.make('ImportClause', clausePos, parts);
            this.parser.node(clause).semantic = phase;
            children.push(clause);
            this.parser.expect('FromKeyword');
        }
        children.push(this.moduleSpecifier());
        if(this.parser.kind() === 'WithKeyword' || this.parser.kind() === 'AssertKeyword') {
            children.push(this.attributes());
        }
        this.semicolon();
        return this.parser.make('ImportDeclaration', pos, children);
    }
    exportDeclaration(pos: number, prefix: readonly number[]): number {
        const oldAwait = this.parser.getAwait();
        this.parser.setAwait(true);
        const children = prefix.slice();
        this.parser.expect('ExportKeyword');
        if(this.parser.kind() === 'DefaultKeyword' || this.parser.kind() === 'EqualsToken') {
            const equals = this.parser.kind() === 'EqualsToken';
            this.parser.next();
            children.push(this.parser.rootAssignment());
            this.semicolon();
            const id = this.parser.make('ExportAssignment', pos, children);
            this.parser.node(id).semantic = equals ? '1' : '0';
            this.parser.setAwait(oldAwait);
            return id;
        }
        if(this.parser.kind() === 'AsKeyword') {
            this.parser.next();
            this.parser.expect('NamespaceKeyword');
            children.push(this.parser.identifier());
            this.semicolon();
            this.parser.setAwait(oldAwait);
            return this.parser.make('NamespaceExportDeclaration', pos, children);
        }
        const typeOnly = this.parser.kind() === 'TypeKeyword';
        if(typeOnly) {
            this.parser.next();
        }
        if(this.parser.kind() === 'AsteriskToken') {
            const start = this.parser.scanner.fullStart;
            this.parser.next();
            if(this.parser.kind() === 'AsKeyword') {
                this.parser.next();
                const name = this.parser.propertyName();
                children.push(this.parser.make('NamespaceExport', start, [name]));
            }
            this.parser.expect('FromKeyword');
            children.push(this.moduleSpecifier());
        }
        else {
            children.push(this.specifiers(false));
            if(this.parser.kind() === 'FromKeyword') {
                this.parser.next();
                children.push(this.moduleSpecifier());
            }
        }
        if(this.parser.kind() === 'WithKeyword' || this.parser.kind() === 'AssertKeyword') {
            children.push(this.attributes());
        }
        this.semicolon();
        const id = this.parser.make('ExportDeclaration', pos, children);
        this.parser.node(id).semantic = typeOnly ? '1' : '0';
        this.parser.setAwait(oldAwait);
        return id;
    }
    moduleDeclaration(pos: number, prefix: readonly number[], keyword: string): number {
        const children = prefix.slice();
        children.push(this.parser.propertyName());
        if(this.parser.kind() === 'WithKeyword') {
            this.parser.next();
            children.push(this.parser.typeLiteral());
        }
        if(this.parser.kind() === 'DotToken') {
            this.parser.next();
            const start = this.parser.scanner.fullStart;
            const exported = this.parser.make('ExportKeyword', start);
            this.parser.node(exported).end = start;
            children.push(this.moduleDeclaration(start, [exported], keyword));
        }
        else if(this.parser.kind() === 'OpenBraceToken') {
            const start = this.parser.scanner.fullStart;
            this.parser.next();
            const body = this.statementList('block');
            this.parser.expect('CloseBraceToken');
            children.push(this.parser.make('ModuleBlock', start, body));
        }
        else if(
            keyword === 'GlobalKeyword' ||
            this.parser.node(children[prefix.length] ?? panic('missing module name')).kind === 'StringLiteral'
        ) {
            this.semicolon();
        }
        else {
            this.parser.expect('OpenBraceToken');
            children.push(this.parser.make('ModuleBlock', this.parser.scanner.fullStart));
        }
        const id = this.parser.make('ModuleDeclaration', pos, children);
        this.parser.node(id).operator = keyword;
        return id;
    }
    exportedClauseAhead(): boolean {
        const after = this.lookahead(1);
        if(after === 'DefaultKeyword') {
            const kind = this.lookahead(2);
            return (
                kind !== 'ClassKeyword' &&
                kind !== 'FunctionKeyword' &&
                kind !== 'InterfaceKeyword' &&
                kind !== 'AbstractKeyword' &&
                kind !== 'AsyncKeyword' &&
                kind !== 'AtToken'
            );
        }
        return (
            after === 'AsteriskToken' ||
            after === 'OpenBraceToken' ||
            after === 'EqualsToken' ||
            after === 'AsKeyword' ||
            (after === 'TypeKeyword' &&
                (this.lookahead(2) === 'OpenBraceToken' || this.lookahead(2) === 'AsteriskToken'))
        );
    }
    functionDeclaration(pos: number, prefix: readonly number[], async: boolean): number {
        const children = prefix.slice();
        this.parser.expect('FunctionKeyword');
        const generator = this.parser.kind() === 'AsteriskToken';
        if(generator) {
            children.push(this.parser.token());
        }
        let defaulted = false;
        for(const modifier of prefix) {
            if(this.parser.node(modifier).kind === 'DefaultKeyword') {
                defaulted = true;
            }
        }
        if(
            !defaulted ||
            this.parser.kind() === 'Identifier' ||
            (this.parser.kind().endsWith('Keyword') && !isReservedKind(this.parser.kind()))
        ) {
            children.push(this.parser.bindingIdentifier());
        }
        const oldAwait = this.parser.getAwait();
        const oldYield = this.parser.getYield();
        this.parser.setAwait(async);
        this.parser.setYield(generator);
        const oldIn = this.parser.getIn();
        this.parser.setIn(false);
        for(const type of this.parser.typeParameters()) {
            children.push(type);
        }
        for(const parameter of this.parser.parameters()) {
            children.push(parameter);
        }
        if(this.parser.kind() === 'ColonToken') {
            this.parser.next();
            children.push(this.parser.returnType());
        }
        if(this.parser.kind() === 'OpenBraceToken') {
            children.push(this.block());
        }
        else if(
            this.parser.kind() === 'SemicolonToken' ||
            this.parser.kind() === 'CloseBraceToken' ||
            this.parser.kind() === 'EndOfFile' ||
            (this.parser.scanner.flags & 1) !== 0
        ) {
            this.semicolon();
        }
        else {
            this.parser.errorAt(1144, this.parser.scanner.start, this.parser.scanner.pos, "'{' or ';' expected.");
            children.push(this.parser.make('Block', this.parser.scanner.fullStart));
        }
        this.parser.setIn(oldIn);
        this.parser.setAwait(oldAwait);
        this.parser.setYield(oldYield);
        return this.parser.make('FunctionDeclaration', pos, children);
    }
    classDeclaration(pos: number, prefix: readonly number[], expression: boolean): number {
        const children = prefix.slice();
        this.parser.expect('ClassKeyword');
        if(
            this.parser.kind() === 'Identifier' ||
            (this.parser.kind().endsWith('Keyword') &&
                this.parser.kind() !== 'ExtendsKeyword' &&
                this.parser.kind() !== 'ImplementsKeyword')
        ) {
            children.push(this.parser.identifier());
        }
        for(const type of this.parser.typeParameters()) {
            children.push(type);
        }
        for(const heritage of this.heritageClauses(false)) {
            children.push(heritage);
        }
        this.parser.expect('OpenBraceToken');
        this.parser.beginList('class');
        while(this.parser.kind() !== 'EndOfFile' && this.parser.kind() !== 'CloseBraceToken') {
            if(!this.parser.listElement('class')) {
                if(this.parser.recoverList('class')) {
                    break;
                }
                continue;
            }
            const start = this.parser.scanner.fullStart;
            if(this.parser.kind() === 'SemicolonToken') {
                this.parser.next();
                children.push(this.parser.make('SemicolonClassElement', start));
                continue;
            }
            const member = this.parser.modifiers(true, true);
            let async = false;
            for(const modifier of member) {
                async = async || this.parser.node(modifier).kind === 'AsyncKeyword';
            }
            if(
                this.parser.kind() === 'OpenBraceToken' &&
                member.length > 0 &&
                this.parser.node(member[0] ?? -1).kind === 'StaticKeyword'
            ) {
                const block = this.block();
                children.push(this.parser.make('ClassStaticBlockDeclaration', start, [block]));
                continue;
            }
            if(this.parser.kind() === 'OpenBracketToken' && this.lookahead(2) === 'ColonToken') {
                this.parser.next();
                const parameterPos = this.parser.scanner.fullStart;
                const parameter = [this.parser.identifier()];
                this.parser.expect('ColonToken');
                parameter.push(this.parser.type());
                member.push(this.parser.make('Parameter', parameterPos, parameter));
                this.parser.expect('CloseBracketToken');
                if(this.parser.kind() === 'ColonToken') {
                    this.parser.next();
                    member.push(this.parser.type());
                }
                this.semicolon();
                children.push(this.parser.make('IndexSignature', start, member));
                continue;
            }
            let methodKind = 'MethodDeclaration';
            if(
                (this.parser.kind() === 'GetKeyword' || this.parser.kind() === 'SetKeyword') && accessorAhead(this.parser.scanner)
            ) {
                methodKind = this.parser.kind() === 'GetKeyword' ? 'GetAccessor' : 'SetAccessor';
                this.parser.next();
            }
            const generator = this.parser.kind() === 'AsteriskToken';
            if(generator) {
                member.push(this.parser.token());
            }
            const nameStart = this.parser.scanner.start;
            const constructor = this.parser.kind() === 'ConstructorKeyword';
            if(constructor) {
                this.parser.next();
                methodKind = 'Constructor';
            }
            else {
                member.push(this.parser.propertyName());
            }
            const name = member[member.length - 1] ?? -1;
            if(this.parser.kind() === 'QuestionToken' || this.parser.kind() === 'ExclamationToken') {
                member.push(this.parser.token());
            }
            if(methodKind !== 'MethodDeclaration' || this.parser.kind() === 'OpenParenToken' || this.parser.kind() === 'LessThanToken') {
                children.push(this.parser.methodBody(start, member, methodKind, async, generator));
            }
            else {
                let annotated = false;
                let initialized = false;
                if(this.parser.kind() === 'ColonToken') {
                    annotated = true;
                    this.parser.next();
                    member.push(this.parser.type());
                }
                if(this.parser.kind() === 'EqualsToken') {
                    initialized = true;
                    this.parser.next();
                    member.push(this.parser.rootAssignment());
                }
                if(!annotated && !initialized && name >= 0) {
                    this.expressionSemicolon(name, nameStart);
                }
                else {
                    this.semicolon();
                }
                children.push(this.parser.make('PropertyDeclaration', start, member));
            }
        }
        this.parser.endList('class');
        this.parser.expect('CloseBraceToken');
        return this.parser.make(expression ? 'ClassExpression' : 'ClassDeclaration', pos, children);
    }
    heritageNameValid(index: number): boolean {
        const node = this.parser.node(index);
        if(node.kind === 'Identifier') {
            return node.pos !== node.end;
        }
        return node.kind === 'PropertyAccessExpression' && !node.optional &&
            node.children.length === 2 && this.heritageNameValid(node.children[0] ?? -1) &&
            this.heritageNameValid(node.children[1] ?? -1);
    }
    heritageName(index: number): number {
        const node = this.parser.node(index);
        if(node.kind === 'PropertyAccessExpression') {
            const qualified = this.parser.make('QualifiedName', node.pos, [
                this.heritageName(node.children[0] ?? -1), node.children[1] ?? -1,
            ]);
            this.parser.node(qualified).end = node.end;
            return qualified;
        }
        return index;
    }
    heritageClauses(interface_: boolean): number[] {
        const clauses: number[] = [];
        this.parser.beginList('heritageClauses');
        while(this.parser.kind() === 'ExtendsKeyword' || this.parser.kind() === 'ImplementsKeyword') {
            const start = this.parser.scanner.fullStart;
            const operator = this.parser.kind();
            const typeMode = interface_ ? operator === 'ExtendsKeyword' : operator === 'ImplementsKeyword';
            this.parser.next();
            const bases: number[] = [];
            this.parser.beginList('heritage');
            while(true) {
                if(!this.parser.listElement('heritage')) {
                    if(this.parser.listTerminator('heritage') || this.parser.recoverList('heritage')) {
                        break;
                    }
                    continue;
                }
                const pos = this.parser.scanner.fullStart;
                const target = this.parser.suffix(this.parser.primary(), true);
                const original = this.parser.node(target);
                const parts = original.kind === 'ExpressionWithTypeArguments' ? original.children.slice() : [target];
                const expression = parts[0] ?? -1;
                if(this.parser.kind() === 'LessThanToken') {
                    for(const type of this.parser.typeArguments()) {
                        parts.push(type);
                    }
                }
                if(typeMode && this.heritageNameValid(expression)) {
                    parts[0] = this.heritageName(expression);
                    bases.push(this.parser.make('TypeReference', pos, parts));
                }
                else {
                    bases.push(
                        original.kind === 'ExpressionWithTypeArguments'
                            ? target
                            : this.parser.make('ExpressionWithTypeArguments', pos, parts),
                    );
                }
                if(!interface_ && !typeMode && this.parser.depth() === 0) {
                    this.parser.roots.push(expression);
                }
                if(this.parser.kind() === 'CommaToken') {
                    this.parser.next();
                }
                else if(this.parser.listTerminator('heritage')) {
                    break;
                }
                else {
                    this.parser.expect('CommaToken');
                }
            }
            this.parser.endList('heritage');
            const clause = this.parser.make('HeritageClause', start, bases);
            this.parser.node(clause).operator = operator;
            clauses.push(clause);
        }
        this.parser.endList('heritageClauses');
        return clauses;
    }
    decoratedExpression(pos: number): number {
        const modifiers = this.parser.modifiers(true, false);
        if(this.parser.kind() === 'ClassKeyword') {
            return this.classDeclaration(pos, modifiers, true);
        }
        this.parser.errorAt(1109, this.parser.scanner.fullStart, this.parser.scanner.fullStart, 'Expression expected.');
        return this.parser.make('MissingDeclaration', pos, modifiers);
    }
    tuple(pos: number): number {
        this.parser.next();
        const children: number[] = [];
        this.parser.beginList('tuple');
        while(!this.parser.listTerminator('tuple')) {
            if(!this.parser.listElement('tuple')) {
                if(this.parser.recoverList('tuple')) {
                    break;
                }
                continue;
            }
            const start = this.parser.scanner.fullStart;
            const rest = this.parser.kind() === 'DotDotDotToken' ? this.parser.token() : -1;
            if(this.parser.tupleNameAhead()) {
                const parts: number[] = [];
                if(rest >= 0) {
                    parts.push(rest);
                }
                parts.push(this.parser.identifier());
                if(this.parser.kind() === 'QuestionToken') {
                    parts.push(this.parser.token());
                }
                this.parser.expect('ColonToken');
                parts.push(this.parser.type());
                children.push(this.parser.make('NamedTupleMember', start, parts));
            }
            else {
                const child = this.parser.type();
                if(rest >= 0) {
                    children.push(this.parser.make('RestType', start, [child]));
                }
                else if(this.parser.kind() === 'QuestionToken') {
                    this.parser.next();
                    children.push(this.parser.make('OptionalType', start, [child]));
                }
                else {
                    children.push(child);
                }
            }
            if(this.parser.kind() === 'CommaToken') {
                this.parser.next();
            }
            else if(!this.parser.listTerminator('tuple')) {
                this.parser.expect('CommaToken');
            }
        }
        this.parser.endList('tuple');
        this.parser.expect('CloseBracketToken');
        return this.parser.make('TupleType', pos, children);
    }
    forStatement(): number {
        const pos = this.parser.scanner.fullStart;
        this.parser.next();
        const children: number[] = [];
        if(this.parser.kind() === 'AwaitKeyword') {
            children.push(this.parser.token());
        }
        this.parser.expect('OpenParenToken');
        const old = this.parser.getIn();
        this.parser.setIn(true);
        if(this.parser.kind() !== 'SemicolonToken') {
            children.push(
                this.parser.kind() === 'ConstKeyword' ||
                    this.parser.kind() === 'LetKeyword' ||
                    this.parser.kind() === 'VarKeyword' ||
                    (this.parser.kind() === 'UsingKeyword' && this.parser.nextIdentifierSameLine()) ||
                    (this.parser.kind() === 'AwaitKeyword' && this.parser.peek() === 'UsingKeyword')
                    ? this.variableList()
                    : this.parser.rootExpression(),
            );
        }
        this.parser.setIn(old);
        if(this.parser.kind() === 'OfKeyword' || this.parser.kind() === 'InKeyword') {
            const of = this.parser.kind() === 'OfKeyword';
            this.parser.next();
            children.push(of ? this.parser.rootAssignment() : this.parser.rootExpression());
            this.parser.expect('CloseParenToken');
            children.push(this.statement());
            return this.parser.make(of ? 'ForOfStatement' : 'ForInStatement', pos, children);
        }
        this.parser.expect('SemicolonToken');
        if(this.parser.kind() !== 'SemicolonToken') {
            children.push(this.parser.rootExpression());
        }
        this.parser.expect('SemicolonToken');
        if(this.parser.kind() !== 'CloseParenToken') {
            children.push(this.parser.rootExpression());
        }
        this.parser.expect('CloseParenToken');
        children.push(this.statement());
        return this.parser.make('ForStatement', pos, children);
    }
    switchStatement(): number {
        const pos = this.parser.scanner.fullStart;
        this.parser.next();
        this.parser.expect('OpenParenToken');
        const expression = this.parser.rootExpression();
        this.parser.expect('CloseParenToken');
        const start = this.parser.scanner.fullStart;
        this.parser.expect('OpenBraceToken');
        const clauses: number[] = [];
        this.parser.beginList('switch');
        while(!this.parser.listTerminator('switch')) {
            if(!this.parser.listElement('switch')) {
                if(this.parser.recoverList('switch')) {
                    break;
                }
                continue;
            }
            const clausePos = this.parser.scanner.fullStart;
            const isCase = this.parser.kind() === 'CaseKeyword';
            this.parser.next();
            const statements: number[] = [];
            if(isCase) {
                statements.push(this.parser.rootExpression());
            }
            this.parser.expect('ColonToken');
            for(const statement of this.statementList('switchStatements')) {
                statements.push(statement);
            }
            clauses.push(this.parser.make(isCase ? 'CaseClause' : 'DefaultClause', clausePos, statements));
        }
        this.parser.endList('switch');
        this.parser.expect('CloseBraceToken');
        const block = this.parser.make('CaseBlock', start, clauses);
        return this.parser.make('SwitchStatement', pos, [expression, block]);
    }
    tryStatement(): number {
        const pos = this.parser.scanner.fullStart;
        this.parser.next();
        const children = [this.block()];
        if(this.parser.kind() === 'CatchKeyword') {
            const start = this.parser.scanner.fullStart;
            this.parser.next();
            const clause: number[] = [];
            if(this.parser.kind() === 'OpenParenToken') {
                this.parser.next();
                const declarationPos = this.parser.scanner.fullStart;
                const parts = [this.parser.bindingName()];
                if(this.parser.kind() === 'ColonToken') {
                    this.parser.next();
                    parts.push(this.parser.type());
                }
                clause.push(this.parser.make('VariableDeclaration', declarationPos, parts));
                this.parser.expect('CloseParenToken');
            }
            clause.push(this.block());
            children.push(this.parser.make('CatchClause', start, clause));
        }
        if(children.length === 1 || this.parser.kind() === 'FinallyKeyword') {
            if(this.parser.kind() === 'FinallyKeyword') {
                this.parser.next();
            }
            else {
                this.parser.errorAt(
                    1472,
                    this.parser.scanner.start,
                    this.parser.scanner.pos,
                    "'catch' or 'finally' expected.",
                );
            }
            children.push(this.block());
        }
        return this.parser.make('TryStatement', pos, children);
    }
    enumDeclaration(pos: number, prefix: readonly number[]): number {
        this.parser.next();
        const children = prefix.slice();
        children.push(this.parser.bindingIdentifier());
        if(!this.parser.expect('OpenBraceToken')) {
            return this.parser.make('EnumDeclaration', pos, children);
        }
        const oldAwait = this.parser.getAwait();
        const oldYield = this.parser.getYield();
        const oldIn = this.parser.getIn();
        this.parser.setAwait(false);
        this.parser.setYield(false);
        this.parser.setIn(false);
        this.parser.beginList('enum');
        while(true) {
            if(!this.parser.listElement('enum')) {
                if(this.parser.listTerminator('enum') || this.parser.recoverList('enum')) {
                    break;
                }
                continue;
            }
            const start = this.parser.scanner.fullStart;
            const member = [this.parser.propertyName()];
            if(this.parser.kind() === 'EqualsToken') {
                this.parser.next();
                member.push(this.parser.rootAssignment());
            }
            children.push(this.parser.make('EnumMember', start, member));
            if(this.parser.kind() === 'CommaToken') {
                this.parser.next();
                continue;
            }
            if(this.parser.listTerminator('enum')) {
                break;
            }
            this.parser.errorAt(
                1357,
                this.parser.scanner.start,
                this.parser.scanner.pos,
                `An enum member name must be followed by a ',', '=', or '}'.`,
            );
            if(start === this.parser.scanner.fullStart) {
                this.parser.next();
            }
        }
        this.parser.endList('enum');
        this.parser.setAwait(oldAwait);
        this.parser.setYield(oldYield);
        this.parser.setIn(oldIn);
        this.parser.expect('CloseBraceToken');
        return this.parser.make('EnumDeclaration', pos, children);
    }
    letDeclaration(): boolean {
        const next = this.parser.peek();
        return (
            next === 'Identifier' ||
            (next.endsWith('Keyword') && !isReservedKind(next)) ||
            next === 'OpenBraceToken' ||
            next === 'OpenBracketToken'
        );
    }
    statement(): number {
        const pos = this.parser.scanner.fullStart;
        if(this.parser.kind() === 'WithKeyword') {
            this.parser.next();
            this.parser.expect('OpenParenToken');
            const expression = this.parser.rootExpression();
            this.parser.expect('CloseParenToken');
            return this.parser.make('WithStatement', pos, [expression, this.statement()]);
        }
        if(this.parser.kind() === 'ExportKeyword' && this.exportedClauseAhead()) {
            return this.exportDeclaration(pos, []);
        }
        if(this.parser.kind() === 'OpenBraceToken') {
            return this.block();
        }
        if(this.parser.kind() === 'SemicolonToken') {
            this.parser.next();
            return this.parser.make('EmptyStatement', pos);
        }
        if(this.parser.kind() === 'ReturnKeyword' || this.parser.kind() === 'ThrowKeyword') {
            const returning = this.parser.kind() === 'ReturnKeyword';
            this.parser.next();
            const children: number[] = [];
            const tokenStart = this.parser.scanner.start;
            if(!returning) {
                const expression =
                    (this.parser.scanner.flags & 1) !== 0
                        ? this.parser.make('Identifier', this.parser.scanner.fullStart)
                        : this.parser.rootExpression();
                children.push(expression);
                this.expressionSemicolon(expression, tokenStart);
            }
            else if(
                this.parser.kind() !== 'CloseBraceToken' &&
                this.parser.kind() !== 'SemicolonToken' &&
                this.parser.kind() !== 'EndOfFile' &&
                (this.parser.scanner.flags & 1) === 0
            ) {
                children.push(this.parser.rootExpression());
            }
            if(returning) {
                this.semicolon();
            }
            return this.parser.make(returning ? 'ReturnStatement' : 'ThrowStatement', pos, children);
        }
        if(
            (this.parser.kind() === 'ConstKeyword' && this.parser.peek() !== 'EnumKeyword') ||
            (this.parser.kind() === 'LetKeyword' && this.letDeclaration()) ||
            this.parser.kind() === 'VarKeyword' ||
            (this.parser.kind() === 'UsingKeyword' && this.parser.nextIdentifierSameLine()) ||
            (this.parser.kind() === 'AwaitKeyword' && this.parser.peek() === 'UsingKeyword')
        ) {
            const list = this.variableList();
            this.semicolon();
            return this.parser.make('VariableStatement', pos, [list]);
        }
        if(this.parser.kind() === 'IfKeyword' || this.parser.kind() === 'WhileKeyword') {
            const conditional = this.parser.kind() === 'IfKeyword';
            this.parser.next();
            this.parser.expect('OpenParenToken');
            const condition = this.parser.rootExpression();
            this.parser.expect('CloseParenToken');
            const body = this.statement();
            const children = [condition, body];
            if(conditional && this.parser.kind() === 'ElseKeyword') {
                this.parser.next();
                children.push(this.statement());
            }
            return this.parser.make(conditional ? 'IfStatement' : 'WhileStatement', pos, children);
        }
        if(this.parser.kind() === 'ForKeyword') {
            return this.forStatement();
        }
        if(this.parser.kind() === 'SwitchKeyword') {
            return this.switchStatement();
        }
        if(this.parser.kind() === 'TryKeyword') {
            return this.tryStatement();
        }
        if(this.parser.kind() === 'DoKeyword') {
            this.parser.next();
            const body = this.statement();
            this.parser.expect('WhileKeyword');
            this.parser.expect('OpenParenToken');
            const condition = this.parser.rootExpression();
            this.parser.expect('CloseParenToken');
            if(this.parser.kind() === 'SemicolonToken') {
                this.parser.next();
            }
            return this.parser.make('DoStatement', pos, [body, condition]);
        }
        if(
            this.parser.kind() === 'BreakKeyword' ||
            this.parser.kind() === 'ContinueKeyword' ||
            this.parser.kind() === 'DebuggerKeyword'
        ) {
            const kind =
                this.parser.kind() === 'BreakKeyword'
                    ? 'BreakStatement'
                    : this.parser.kind() === 'ContinueKeyword'
                      ? 'ContinueStatement'
                      : 'DebuggerStatement';
            this.parser.next();
            const children: number[] = [];
            if(
                kind !== 'DebuggerStatement' &&
                this.parser.kind() !== 'SemicolonToken' &&
                this.parser.kind() !== 'CloseBraceToken' &&
                this.parser.kind() !== 'EndOfFile' &&
                (this.parser.scanner.flags & 1) === 0
            ) {
                children.push(this.parser.bindingIdentifier());
            }
            this.semicolon();
            return this.parser.make(kind, pos, children);
        }
        const modifiers: number[] = [];
        let async = false;
        while(
            this.parser.kind() === 'AtToken' ||
            (this.parser.kind() === 'ExportKeyword' && !this.exportedClauseAhead()) ||
            this.parser.kind() === 'DefaultKeyword' ||
            (declarationModifierAhead(this.parser.scanner) && declarationAhead(this.parser.scanner)) ||
            (this.parser.kind() === 'ConstKeyword' && this.parser.peek() === 'EnumKeyword')
        ) {
            if(this.parser.kind() === 'AsyncKeyword') {
                async = true;
            }
            modifiers.push(this.parser.kind() === 'AtToken' ? this.decorator() : this.parser.token());
        }
        if(this.parser.kind() === 'ExportKeyword' && this.exportedClauseAhead()) {
            return this.exportDeclaration(pos, modifiers);
        }
        if(this.parser.kind() === 'FunctionKeyword') {
            return this.functionDeclaration(pos, modifiers, async);
        }
        if(this.parser.kind() === 'ClassKeyword') {
            return this.classDeclaration(pos, modifiers, false);
        }
        if(
            this.parser.kind() === 'ConstKeyword' ||
            (this.parser.kind() === 'LetKeyword' && (modifiers.length > 0 || this.letDeclaration())) ||
            this.parser.kind() === 'VarKeyword' ||
            (this.parser.kind() === 'UsingKeyword' && this.parser.nextIdentifierSameLine()) ||
            (this.parser.kind() === 'AwaitKeyword' && this.parser.peek() === 'UsingKeyword')
        ) {
            modifiers.push(this.variableList());
            this.semicolon();
            return this.parser.make('VariableStatement', pos, modifiers);
        }
        if(
            (this.parser.kind() === 'TypeKeyword' || this.parser.kind() === 'InterfaceKeyword') &&
            this.parser.nextIdentifierSameLine()
        ) {
            const interface_ = this.parser.kind() === 'InterfaceKeyword';
            this.parser.next();
            modifiers.push(this.parser.identifier());
            for(const type of this.parser.typeParameters()) {
                modifiers.push(type);
            }
            if(interface_) {
                for(const heritage of this.heritageClauses(true)) {
                    modifiers.push(heritage);
                }
                const body = this.parser.typeLiteral();
                for(const member of this.parser.node(body).children) {
                    modifiers.push(member);
                }
            }
            else {
                this.parser.expect('EqualsToken');
                modifiers.push(
                    this.parser.kind() === 'IntrinsicKeyword' && this.parser.peek() !== 'DotToken'
                        ? this.parser.token()
                        : this.parser.type(),
                );
                this.semicolon();
            }
            return this.parser.make(interface_ ? 'InterfaceDeclaration' : 'TypeAliasDeclaration', pos, modifiers);
        }
        if(this.parser.kind() === 'EnumKeyword') {
            return this.enumDeclaration(pos, modifiers);
        }
        if(
            ((this.parser.kind() === 'NamespaceKeyword' || this.parser.kind() === 'ModuleKeyword') &&
                declarationAhead(this.parser.scanner)) ||
            (this.parser.kind() === 'GlobalKeyword' && this.parser.peek() === 'OpenBraceToken')
        ) {
            const keyword = this.parser.kind();
            if(keyword !== 'GlobalKeyword') {
                this.parser.next();
            }
            return this.moduleDeclaration(pos, modifiers, keyword);
        }
        if(
            this.parser.kind() === 'ImportKeyword' &&
            this.parser.peek() !== 'OpenParenToken' &&
            this.parser.peek() !== 'DotToken'
        ) {
            return this.importDeclaration(pos, modifiers);
        }
        const tokenStart = this.parser.scanner.start;
        const expression = this.parser.rootExpression();
        if(this.parser.node(expression).kind === 'Identifier' && this.parser.kind() === 'ColonToken') {
            if(this.parser.depth() === 0) {
                this.parser.roots.pop();
            }
            this.parser.next();
            return this.parser.make('LabeledStatement', pos, [expression, this.statement()]);
        }

        this.expressionSemicolon(expression, tokenStart);
        return this.parser.make('ExpressionStatement', pos, [expression]);
    }
}
