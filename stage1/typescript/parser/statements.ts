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
        return this.make('VariableDeclarationList', pos, declarations);
    }
    skipDeclaration(interfaceBody: boolean): void {
        let braces = 0;
        let parens = 0;
        let brackets = 0;
        let templates = 0;
        while(this.parser.kind() !== 'EndOfFile') {
            if(
                this.parser.kind() === 'TypeOfKeyword' &&
                this.parser.peek() !== 'ImportKeyword' &&
                (this.parser.peek() === 'Identifier' || this.parser.peek().endsWith('Keyword'))
            ) {
                this.parser.next();
                const name = this.parser.entityName();
                if(this.parser.depth() === 0) {
                    this.parser.roots.push(name);
                }
                continue;
            }
            if(this.parser.kind() === 'TemplateHead') {
                templates++;
                this.parser.next();
                continue;
            }
            if(this.parser.kind() === 'CloseBraceToken' && templates > 0 && braces === 0) {
                this.parser.scanner.rescanTemplate();
                if(this.parser.kind() === 'TemplateTail') {
                    templates--;
                }
                this.parser.next();
                continue;
            }
            if(this.parser.kind() === 'OpenBraceToken') {
                braces++;
            }
            else if(this.parser.kind() === 'CloseBraceToken') {
                braces--;
                this.parser.next();
                if(interfaceBody && braces === 0 && parens === 0 && brackets === 0) {
                    if(this.parser.kind() === 'SemicolonToken') {
                        this.parser.next();
                    }
                    return;
                }
                continue;
            }
            else if(this.parser.kind() === 'OpenParenToken') {
                parens++;
            }
            else if(this.parser.kind() === 'CloseParenToken') {
                parens--;
            }
            else if(this.parser.kind() === 'OpenBracketToken') {
                brackets++;
            }
            else if(this.parser.kind() === 'CloseBracketToken') {
                brackets--;
            }
            else if(this.parser.kind() === 'SemicolonToken' && braces === 0 && parens === 0 && brackets === 0) {
                this.parser.next();
                return;
            }
            this.parser.next();
        }
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
            let methodKind = 'MethodDeclaration';
            if(
                (this.parser.kind() === 'GetKeyword' || this.parser.kind() === 'SetKeyword') &&
                this.parser.peek() !== 'OpenParenToken' &&
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
                    this.parser.kind() === 'VarKeyword'
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
            return this.make(of ? 'ForOfStatement' : 'ForInStatement', pos, children);
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
        return this.make('ForStatement', pos, children);
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
            this.parser.kind() === 'VarKeyword'
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
            modifiers.push(this.parser.token());
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
            this.parser.kind() === 'VarKeyword'
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
                        bases.push(this.make('ExpressionWithTypeArguments', startType, parts));
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
                modifiers.push(this.type());
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
            (this.parser.kind() === 'NamespaceKeyword' || this.parser.kind() === 'ModuleKeyword') &&
            (this.parser.nextIdentifierSameLine() || this.parser.peek() === 'StringLiteral')
        ) {
            this.parser.next();
            while(this.parser.kind() !== 'OpenBraceToken' && this.parser.kind() !== 'EndOfFile') {
                this.parser.next();
            }
            modifiers.push(this.block());
            return this.make('ModuleDeclaration', pos, modifiers);
        }
        if(
            (this.parser.kind() === 'ImportKeyword' &&
                this.parser.peek() !== 'OpenParenToken' &&
                this.parser.peek() !== 'DotToken') ||
            (modifiers.length > 0 &&
                (this.parser.kind() === 'OpenBraceToken' || this.parser.kind() === 'AsteriskToken'))
        ) {
            this.skipDeclaration(false);
            return this.make('ImportDeclaration', pos);
        }
        if(modifiers.length > 0 && this.parser.kind() === 'EqualsToken') {
            this.parser.next();
        }
        const expression = this.parser.rootExpression();
        this.semicolon();
        return this.make('ExpressionStatement', pos, [expression]);
    }
}
