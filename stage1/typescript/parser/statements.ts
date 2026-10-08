// Statement and declaration descent follows typescript-go's parser.
import { panic } from 'adamic';
import type { Scanner } from '../scanner/scanner.ts';
import type { ParseNode } from './nodes.ts';

export interface StatementContextInterface {
    readonly scanner: Scanner;
    readonly path: string;
    readonly roots: number[];
    readonly kind: () => string;
    readonly next: () => void;
    readonly expect: (kind: string) => void;
    readonly node: (index: number) => ParseNode;
    readonly make: (kind: string, pos: number, children: number[]) => number;
    readonly entityName: () => number;
    readonly identifier: () => number;
    readonly token: () => number;
    readonly type: (minimum: number, conditional: boolean) => number;
    readonly typeArguments: () => number[];
    readonly typeParameters: () => number[];
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
    make(kind: string, pos: number, children: number[] = []): number {
        return this.parser.make(kind, pos, children);
    }
    type(minimum = 0, conditional = true): number {
        return this.parser.type(minimum, conditional);
    }
    block(): number {
        const pos = this.parser.scanner.fullStart;
        this.parser.expect('OpenBraceToken');
        const children: number[] = [];
        while(this.parser.kind() !== 'CloseBraceToken') {
            children.push(this.statement());
        }
        this.parser.expect('CloseBraceToken');
        return this.make('Block', pos, children);
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
            panic(`parser slice expected semicolon at ${this.parser.scanner.start} in ${this.parser.path}`);
        }
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
        while(true) {
            const start = this.parser.scanner.fullStart;
            const children = [this.parser.bindingName()];
            if(this.parser.kind() === 'ExclamationToken') {
                children.push(this.parser.token());
            }
            if(this.parser.kind() === 'ColonToken') {
                this.parser.next();
                children.push(this.type());
            }
            if(this.parser.kind() === 'EqualsToken') {
                this.parser.next();
                children.push(this.parser.rootAssignment());
            }
            declarations.push(this.make('VariableDeclaration', start, children));
            if(this.parser.kind() !== 'CommaToken') {
                break;
            }
            this.parser.next();
        }
        const id = this.make('VariableDeclarationList', pos, declarations);
        this.parser.node(id).semantic = `${flags}`;
        return id;
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
            this.parser.next();
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
        this.parser.expect('OpenBraceToken');
        const members: number[] = [];
        while(this.parser.kind() !== 'CloseBraceToken') {
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
            const member = this.make(importing ? 'ImportSpecifier' : 'ExportSpecifier', start, children);
            this.parser.node(member).semantic = typeOnly ? '1' : '0';
            members.push(member);
            if(this.parser.kind() !== 'CommaToken') {
                break;
            }
            this.parser.next();
        }
        this.parser.expect('CloseBraceToken');
        return this.make(importing ? 'NamedImports' : 'NamedExports', pos, members);
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
        while(this.parser.kind() !== 'CloseBraceToken') {
            const start = this.parser.scanner.fullStart;
            const name = this.parser.propertyName();
            this.parser.expect('ColonToken');
            const value = this.parser.rootAssignment();
            children.push(this.make('ImportAttribute', start, [name, value]));
            trailing = this.parser.kind() === 'CommaToken';
            if(!trailing) {
                break;
            }
            this.parser.next();
        }
        this.parser.expect('CloseBraceToken');
        const id = this.make('ImportAttributes', pos, children);
        const node = this.parser.node(id);
        node.operator = operator;
        node.list = children.length;
        node.trailing = trailing;
        node.multiLine = multiLine;
        return id;
    }
    importDeclaration(pos: number, prefix: readonly number[]): number {
        const children = prefix.slice();
        this.parser.expect('ImportKeyword');
        const clausePos = this.parser.scanner.fullStart;
        let name = -1;
        let phase = 'Unknown';
        if(
            this.parser.kind() !== 'StringLiteral' &&
            this.parser.kind() !== 'AsteriskToken' &&
            this.parser.kind() !== 'OpenBraceToken'
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
                children.push(this.make('ExternalModuleReference', start, [path]));
            }
            else {
                children.push(this.parser.entityName());
            }
            this.semicolon();
            const id = this.make('ImportEqualsDeclaration', pos, children);
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
                    parts.push(this.make('NamespaceImport', start, [target]));
                }
                else {
                    parts.push(this.specifiers(true));
                }
            }
            const clause = this.make('ImportClause', clausePos, parts);
            this.parser.node(clause).semantic = phase;
            children.push(clause);
            this.parser.expect('FromKeyword');
        }
        children.push(this.parser.propertyName());
        if(this.parser.kind() === 'WithKeyword' || this.parser.kind() === 'AssertKeyword') {
            children.push(this.attributes());
        }
        this.semicolon();
        return this.make('ImportDeclaration', pos, children);
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
            const id = this.make('ExportAssignment', pos, children);
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
            return this.make('NamespaceExportDeclaration', pos, children);
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
                children.push(this.make('NamespaceExport', start, [name]));
            }
            this.parser.expect('FromKeyword');
            children.push(this.parser.propertyName());
        }
        else {
            children.push(this.specifiers(false));
            if(this.parser.kind() === 'FromKeyword') {
                this.parser.next();
                children.push(this.parser.propertyName());
            }
        }
        if(this.parser.kind() === 'WithKeyword' || this.parser.kind() === 'AssertKeyword') {
            children.push(this.attributes());
        }
        this.semicolon();
        const id = this.make('ExportDeclaration', pos, children);
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
            const exported = this.make('ExportKeyword', start);
            this.parser.node(exported).end = start;
            children.push(this.moduleDeclaration(start, [exported], keyword));
        }
        else if(this.parser.kind() === 'OpenBraceToken') {
            const start = this.parser.scanner.fullStart;
            this.parser.next();
            const body: number[] = [];
            while(this.parser.kind() !== 'CloseBraceToken') {
                body.push(this.statement());
            }
            this.parser.expect('CloseBraceToken');
            children.push(this.make('ModuleBlock', start, body));
        }
        else {
            this.semicolon();
        }
        const id = this.make('ModuleDeclaration', pos, children);
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
        if(this.parser.kind() !== 'OpenParenToken' && this.parser.kind() !== 'LessThanToken') {
            children.push(this.parser.identifier());
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
        else {
            this.semicolon();
        }
        this.parser.setIn(oldIn);
        this.parser.setAwait(oldAwait);
        this.parser.setYield(oldYield);
        return this.make('FunctionDeclaration', pos, children);
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
        while(this.parser.kind() === 'ExtendsKeyword' || this.parser.kind() === 'ImplementsKeyword') {
            const start = this.parser.scanner.fullStart;
            const extending = this.parser.kind() === 'ExtendsKeyword';
            const operator = this.parser.kind();
            this.parser.next();
            const types: number[] = [];
            while(true) {
                const target = extending ? this.parser.suffix(this.parser.primary(), true) : this.type();
                let item = target;
                if(extending && this.parser.node(target).kind !== 'ExpressionWithTypeArguments') {
                    const arguments_ = [target];
                    if(this.parser.kind() === 'LessThanToken') {
                        for(const type of this.parser.typeArguments()) {
                            arguments_.push(type);
                        }
                    }
                    item = this.make('ExpressionWithTypeArguments', this.parser.node(target).pos, arguments_);
                }
                types.push(item);
                if(extending && this.parser.depth() === 0) {
                    this.parser.roots.push(
                        this.parser.node(target).kind === 'ExpressionWithTypeArguments'
                            ? (this.parser.node(target).children[0] ?? panic('empty heritage'))
                            : target,
                    );
                }
                if(this.parser.kind() !== 'CommaToken') {
                    break;
                }
                this.parser.next();
            }
            const heritage = this.make('HeritageClause', start, types);
            this.parser.node(heritage).operator = operator;
            children.push(heritage);
        }
        this.parser.expect('OpenBraceToken');
        while(this.parser.kind() !== 'CloseBraceToken') {
            const start = this.parser.scanner.fullStart;
            if(this.parser.kind() === 'SemicolonToken') {
                this.parser.next();
                children.push(this.make('SemicolonClassElement', start));
                continue;
            }
            const member: number[] = [];
            while(this.parser.kind() === 'AtToken') {
                member.push(this.decorator());
            }
            let async = false;
            while(
                this.parser.kind() === 'StaticKeyword' ||
                this.parser.kind() === 'PublicKeyword' ||
                this.parser.kind() === 'PrivateKeyword' ||
                this.parser.kind() === 'ProtectedKeyword' ||
                this.parser.kind() === 'ReadonlyKeyword' ||
                this.parser.kind() === 'AbstractKeyword' ||
                this.parser.kind() === 'DeclareKeyword' ||
                this.parser.kind() === 'OverrideKeyword' ||
                this.parser.kind() === 'AsyncKeyword' ||
                this.parser.kind() === 'AccessorKeyword'
            ) {
                if(
                    this.parser.peek() === 'ColonToken' ||
                    this.parser.peek() === 'EqualsToken' ||
                    this.parser.peek() === 'OpenParenToken' ||
                    this.parser.peek() === 'SemicolonToken'
                ) {
                    break;
                }
                if(this.parser.kind() === 'AsyncKeyword') {
                    async = true;
                }
                member.push(this.parser.token());
            }
            if(
                this.parser.kind() === 'OpenBraceToken' &&
                member.length > 0 &&
                this.parser.node(member[0] ?? -1).kind === 'StaticKeyword'
            ) {
                const block = this.block();
                children.push(this.make('ClassStaticBlockDeclaration', start, [block]));
                continue;
            }
            if(this.parser.kind() === 'OpenBracketToken' && this.lookahead(2) === 'ColonToken') {
                this.parser.next();
                const parameterPos = this.parser.scanner.fullStart;
                const parameter = [this.parser.identifier()];
                this.parser.expect('ColonToken');
                parameter.push(this.type());
                member.push(this.make('Parameter', parameterPos, parameter));
                this.parser.expect('CloseBracketToken');
                if(this.parser.kind() === 'ColonToken') {
                    this.parser.next();
                    member.push(this.type());
                }
                this.semicolon();
                children.push(this.make('IndexSignature', start, member));
                continue;
            }
            let methodKind = 'MethodDeclaration';
            if(
                (this.parser.kind() === 'GetKeyword' || this.parser.kind() === 'SetKeyword') &&
                this.parser.peek() !== 'OpenParenToken' &&
                this.parser.peek() !== 'LessThanToken' &&
                this.parser.peek() !== 'SemicolonToken' &&
                this.parser.peek() !== 'QuestionToken' &&
                this.parser.peek() !== 'ExclamationToken' &&
                this.parser.peek() !== 'CloseBraceToken' &&
                this.parser.peek() !== 'ColonToken' &&
                this.parser.peek() !== 'EqualsToken'
            ) {
                methodKind = this.parser.kind() === 'GetKeyword' ? 'GetAccessor' : 'SetAccessor';
                this.parser.next();
            }
            const generator = this.parser.kind() === 'AsteriskToken';
            if(generator) {
                member.push(this.parser.token());
            }
            const constructor = this.parser.kind() === 'ConstructorKeyword';
            if(constructor) {
                this.parser.next();
                methodKind = 'Constructor';
            }
            else {
                member.push(this.parser.propertyName());
            }
            if(this.parser.kind() === 'QuestionToken' || this.parser.kind() === 'ExclamationToken') {
                member.push(this.parser.token());
            }
            if(this.parser.kind() === 'OpenParenToken' || this.parser.kind() === 'LessThanToken') {
                children.push(this.parser.methodBody(start, member, methodKind, async, generator));
            }
            else {
                if(this.parser.kind() === 'ColonToken') {
                    this.parser.next();
                    member.push(this.type());
                }
                if(this.parser.kind() === 'EqualsToken') {
                    this.parser.next();
                    member.push(this.parser.rootAssignment());
                }
                this.semicolon();
                children.push(this.make('PropertyDeclaration', start, member));
            }
        }
        this.parser.expect('CloseBraceToken');
        return this.make(expression ? 'ClassExpression' : 'ClassDeclaration', pos, children);
    }
    forStatement(): number {
        const pos = this.parser.scanner.fullStart;
        this.parser.next();
        const children: number[] = [];
        let initializer = -1; let condition = -1; let incrementor = -1;
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
        if(children.length > 0 && this.parser.node(children[children.length - 1] ?? -1).kind !== 'AwaitKeyword') { initializer = children[children.length - 1] ?? -1; }
        this.parser.setIn(old);
        if(this.parser.kind() === 'OfKeyword' || this.parser.kind() === 'InKeyword') {
            const of = this.parser.kind() === 'OfKeyword';
            this.parser.next();
            children.push(of ? this.parser.rootAssignment() : this.parser.rootExpression());
            this.parser.expect('CloseParenToken');
            children.push(this.statement());
            return this.make(of ? 'ForOfStatement' : 'ForInStatement', pos, children);
        }
        this.parser.expect('SemicolonToken');
        if(this.parser.kind() !== 'SemicolonToken') {
            children.push(this.parser.rootExpression());
            condition = children[children.length - 1] ?? -1;
        }
        this.parser.expect('SemicolonToken');
        if(this.parser.kind() !== 'CloseParenToken') {
            children.push(this.parser.rootExpression());
            incrementor = children[children.length - 1] ?? -1;
        }
        this.parser.expect('CloseParenToken');
        children.push(this.statement());
        const id = this.make('ForStatement', pos, children);
        this.parser.node(id).slots = [initializer, condition, incrementor, children[children.length - 1] ?? -1];
        return id;
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
        while(this.parser.kind() !== 'CloseBraceToken') {
            const clausePos = this.parser.scanner.fullStart;
            const isCase = this.parser.kind() === 'CaseKeyword';
            this.parser.next();
            const statements: number[] = [];
            if(isCase) {
                statements.push(this.parser.rootExpression());
            }
            this.parser.expect('ColonToken');
            while(
                this.parser.kind() !== 'CaseKeyword' &&
                this.parser.kind() !== 'DefaultKeyword' &&
                this.parser.kind() !== 'CloseBraceToken'
            ) {
                statements.push(this.statement());
            }
            clauses.push(this.make(isCase ? 'CaseClause' : 'DefaultClause', clausePos, statements));
        }
        this.parser.expect('CloseBraceToken');
        const block = this.make('CaseBlock', start, clauses);
        return this.make('SwitchStatement', pos, [expression, block]);
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
                    parts.push(this.type());
                }
                clause.push(this.make('VariableDeclaration', declarationPos, parts));
                this.parser.expect('CloseParenToken');
            }
            clause.push(this.block());
            children.push(this.make('CatchClause', start, clause));
        }
        if(this.parser.kind() === 'FinallyKeyword') {
            this.parser.next();
            children.push(this.block());
        }
        return this.make('TryStatement', pos, children);
    }
    statement(): number {
        const pos = this.parser.scanner.fullStart;
        if(this.parser.kind() === 'WithKeyword') {
            this.parser.next();
            this.parser.expect('OpenParenToken');
            const expression = this.parser.rootExpression();
            this.parser.expect('CloseParenToken');
            return this.make('WithStatement', pos, [expression, this.statement()]);
        }
        if(this.parser.kind() === 'ExportKeyword' && this.exportedClauseAhead()) {
            return this.exportDeclaration(pos, []);
        }
        if(this.parser.kind() === 'OpenBraceToken') {
            return this.block();
        }
        if(this.parser.kind() === 'SemicolonToken') {
            this.parser.next();
            return this.make('EmptyStatement', pos);
        }
        if(this.parser.kind() === 'ReturnKeyword' || this.parser.kind() === 'ThrowKeyword') {
            const returning = this.parser.kind() === 'ReturnKeyword';
            this.parser.next();
            const children: number[] = [];
            if(
                this.parser.kind() !== 'CloseBraceToken' &&
                this.parser.kind() !== 'SemicolonToken' &&
                this.parser.kind() !== 'EndOfFile' &&
                (this.parser.scanner.flags & 1) === 0
            ) {
                children.push(this.parser.rootExpression());
            }
            this.semicolon();
            return this.make(returning ? 'ReturnStatement' : 'ThrowStatement', pos, children);
        }
        if(
            (this.parser.kind() === 'ConstKeyword' && this.parser.peek() !== 'EnumKeyword') ||
            this.parser.kind() === 'LetKeyword' ||
            this.parser.kind() === 'VarKeyword' ||
            (this.parser.kind() === 'UsingKeyword' && this.parser.nextIdentifierSameLine()) ||
            (this.parser.kind() === 'AwaitKeyword' && this.parser.peek() === 'UsingKeyword')
        ) {
            const list = this.variableList();
            this.semicolon();
            return this.make('VariableStatement', pos, [list]);
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
            return this.make(conditional ? 'IfStatement' : 'WhileStatement', pos, children);
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
            return this.make('DoStatement', pos, [body, condition]);
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
                this.parser.kind() === 'Identifier' &&
                (this.parser.scanner.flags & 1) === 0
            ) {
                children.push(this.parser.identifier());
            }
            this.semicolon();
            return this.make(kind, pos, children);
        }
        if(this.parser.kind() === 'Identifier' && this.parser.peek() === 'ColonToken') {
            const name = this.parser.identifier();
            this.parser.next();
            const body = this.statement();
            return this.make('LabeledStatement', pos, [name, body]);
        }
        const modifiers: number[] = [];
        let async = false;
        while(
            this.parser.kind() === 'AtToken' ||
            this.parser.kind() === 'ExportKeyword' ||
            this.parser.kind() === 'DefaultKeyword' ||
            this.parser.kind() === 'DeclareKeyword' ||
            this.parser.kind() === 'AbstractKeyword' ||
            (this.parser.kind() === 'AsyncKeyword' && this.parser.peek() === 'FunctionKeyword') ||
            (this.parser.kind() === 'ConstKeyword' && this.parser.peek() === 'EnumKeyword')
        ) {
            if(this.parser.kind() === 'AsyncKeyword') {
                async = true;
            }
            modifiers.push(this.parser.kind() === 'AtToken' ? this.decorator() : this.parser.token());
        }
        if(this.parser.kind() === 'FunctionKeyword') {
            return this.functionDeclaration(pos, modifiers, async);
        }
        if(this.parser.kind() === 'ClassKeyword') {
            return this.classDeclaration(pos, modifiers, false);
        }
        if(
            this.parser.kind() === 'ConstKeyword' ||
            this.parser.kind() === 'LetKeyword' ||
            this.parser.kind() === 'VarKeyword' ||
            (this.parser.kind() === 'UsingKeyword' && this.parser.nextIdentifierSameLine()) ||
            (this.parser.kind() === 'AwaitKeyword' && this.parser.peek() === 'UsingKeyword')
        ) {
            modifiers.push(this.variableList());
            this.semicolon();
            return this.make('VariableStatement', pos, modifiers);
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
                if(this.parser.kind() === 'ExtendsKeyword') {
                    const start = this.parser.scanner.fullStart;
                    this.parser.next();
                    const bases: number[] = [];
                    while(true) {
                        const startType = this.parser.scanner.fullStart;
                        const parts = [this.parser.entityName()];
                        if(this.parser.kind() === 'LessThanToken') {
                            for(const type of this.parser.typeArguments()) {
                                parts.push(type);
                            }
                        }
                        bases.push(this.make('TypeReference', startType, parts));
                        if(this.parser.kind() !== 'CommaToken') {
                            break;
                        }
                        this.parser.next();
                    }
                    const heritage = this.make('HeritageClause', start, bases);
                    this.parser.node(heritage).operator = 'ExtendsKeyword';
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
                        : this.type(),
                );
                this.semicolon();
            }
            return this.make(interface_ ? 'InterfaceDeclaration' : 'TypeAliasDeclaration', pos, modifiers);
        }
        if(this.parser.kind() === 'EnumKeyword') {
            this.parser.next();
            modifiers.push(this.parser.identifier());
            this.parser.expect('OpenBraceToken');
            while(this.parser.kind() !== 'CloseBraceToken') {
                const start = this.parser.scanner.fullStart;
                const member = [this.parser.propertyName()];
                if(this.parser.kind() === 'EqualsToken') {
                    this.parser.next();
                    member.push(this.parser.rootAssignment());
                }
                modifiers.push(this.make('EnumMember', start, member));
                if(this.parser.kind() !== 'CommaToken') {
                    break;
                }
                this.parser.next();
            }
            this.parser.expect('CloseBraceToken');
            return this.make('EnumDeclaration', pos, modifiers);
        }
        if(
            ((this.parser.kind() === 'NamespaceKeyword' || this.parser.kind() === 'ModuleKeyword') &&
                (this.parser.nextIdentifierSameLine() || this.parser.peek() === 'StringLiteral')) ||
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
        const expression = this.parser.rootExpression();
        this.semicolon();
        return this.make('ExpressionStatement', pos, [expression]);
    }
}
