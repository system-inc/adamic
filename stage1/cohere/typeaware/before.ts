import { panic } from 'adamic';
import type { Bindings } from './bindings.ts';
import { Diagnostic } from './diagnostic.ts';

export class Before {
    readonly bindings: Bindings;
    readonly edges = new Map<number, number[]>();
    readonly built: number[] = [];
    constructor(bindings: Bindings) {
        this.bindings = bindings;
    }
    typeReference(index: number): boolean {
        const parentIndex = this.bindings.parent(index);
        if(parentIndex < 0) {
            return false;
        }
        const parent = this.bindings.rules.parser.node(parentIndex);
        if(parent.kind === 'TypeReference') {
            return true;
        }
        if(parent.kind === 'ExpressionWithTypeArguments') {
            const heritage = this.bindings.parent(parentIndex);
            return heritage >= 0 && this.bindings.rules.parser.node(heritage).operator === 'ImplementsKeyword';
        }
        return false;
    }
    query(index: number): boolean {
        let current = index;
        while(current >= 0) {
            const kind = this.bindings.rules.parser.node(current).kind;
            if(kind === 'TypeQuery') {
                return true;
            }
            if(kind !== 'Identifier' && kind !== 'QualifiedName') {
                return false;
            }
            current = this.bindings.parent(current);
        }
        return false;
    }
    context(index: number): number {
        let current = index;
        while(current >= 0) {
            const node = this.bindings.rules.parser.node(current);
            if(this.bindings.functionKind(node.kind) ||
                ['ClassStaticBlockDeclaration', 'ModuleDeclaration', 'SourceFile'].includes(node.kind) ||
                (node.kind === 'PropertyDeclaration' && this.bindings.shape.initializer(current) >= 0)) {
                return current;
            }
            const parent = this.bindings.parent(current);
            if(parent >= 0 && this.bindings.rules.parser.node(parent).kind === 'ComputedPropertyName') {
                current = this.bindings.parent(parent);
            }
            current = this.bindings.parent(current);
        }
        return -1;
    }
    separate(index: number, declaration: number): boolean {
        const wanted = this.context(declaration);
        let current = this.context(index);
        while(current >= 0) {
            if(current === wanted) {
                return false;
            }
            const node = this.bindings.rules.parser.node(current);
            if(node.kind === 'ClassStaticBlockDeclaration' ||
                (node.kind === 'PropertyDeclaration' && node.children.some((child) =>
                    this.bindings.rules.parser.node(child).kind === 'StaticKeyword'))) {
                current = this.context(this.bindings.parent(current));
                continue;
            }
            return true;
        }
        return false;
    }
    inside(index: number, location: number): boolean {
        if(index < 0) {
            return false;
        }
        const node = this.bindings.rules.parser.node(index);
        return node.pos <= location && location <= node.end;
    }
    initializing(index: number, declaration: number): boolean {
        if(this.separate(index, declaration)) {
            return false;
        }
        const parser = this.bindings.rules.parser;
        const location = parser.node(index).end;
        const node = parser.node(declaration);
        if(node.kind === 'ClassDeclaration' || node.kind === 'ClassExpression') {
            if(!this.inside(declaration, location)) {
                return false;
            }
            for(const member of node.children) {
                const shape = parser.node(member);
                if(shape.kind === 'ClassStaticBlockDeclaration' && this.inside(member, location)) {
                    return false;
                }
                if(shape.kind === 'PropertyDeclaration' && shape.children.some((child) =>
                    parser.node(child).kind === 'StaticKeyword') &&
                    this.inside(this.bindings.shape.initializer(member), location)) {
                    return false;
                }
            }
            return true;
        }
        let current = declaration;
        while(current >= 0) {
            const kind = parser.node(current).kind;
            if(['VariableDeclaration', 'BindingElement', 'Parameter'].includes(kind) &&
                this.inside(this.bindings.shape.initializer(current), location)) {
                return true;
            }
            if(kind === 'VariableDeclaration') {
                const list = this.bindings.parent(current);
                const loop = this.bindings.parent(list);
                if(loop >= 0 && ['ForInStatement', 'ForOfStatement'].includes(parser.node(loop).kind)) {
                    return this.inside(parser.node(loop).children[1] ?? -1, location);
                }
                return false;
            }
            if(this.bindings.functionKind(kind) || ['ClassDeclaration', 'ClassExpression',
                'CatchClause', 'ImportDeclaration', 'ExportDeclaration'].includes(kind)) {
                return false;
            }
            current = this.bindings.parent(current);
        }
        return false;
    }
    block(kind: string): boolean {
        return ['SourceFile', 'Block', 'ModuleBlock', 'CaseClause', 'DefaultClause'].includes(kind);
    }
    statement(index: number, block = -1): number {
        let current = index;
        while(current >= 0) {
            const parent = this.bindings.parent(current);
            if(parent >= 0 && (block >= 0 ? parent === block :
                this.block(this.bindings.rules.parser.node(parent).kind))) {
                return current;
            }
            current = parent;
        }
        return -1;
    }
    safe(index: number, declaration: number): boolean {
        const parser = this.bindings.rules.parser;
        if(parser.node(declaration).kind === 'FunctionDeclaration' && this.bindings.shape.body(declaration)) {
            return true;
        }
        let current = this.bindings.parent(index);
        while(current >= 0) {
            const node = parser.node(current);
            if(node.kind === 'ArrowFunction' || node.kind === 'FunctionExpression') {
                if(node.children.some((child) => parser.node(child).kind === 'Parameter')) {
                    return false;
                }
                const callIndex = this.bindings.parent(current);
                const decorator = this.bindings.parent(callIndex);
                return callIndex >= 0 && decorator >= 0 &&
                    parser.node(callIndex).kind === 'CallExpression' && parser.node(decorator).kind === 'Decorator' &&
                    parser.node(callIndex).list > 0 &&
                    parser.node(callIndex).children.slice(-parser.node(callIndex).list).includes(current);
            }
            if(this.bindings.functionKind(node.kind) || node.kind === 'ClassStaticBlockDeclaration') {
                return false;
            }
            current = this.bindings.parent(current);
        }
        return false;
    }
    graphWalk(index: number, owner: number, block: number): void {
        const node = this.bindings.rules.parser.node(index);
        if(node.kind === 'Identifier' && !this.bindings.declaring(index)) {
            const declaration = this.bindings.earliest(index);
            const target = declaration >= 0 ? this.statement(declaration) : -1;
            if(target >= 0 && target !== owner && this.bindings.parent(target) === block) {
                if(!this.edges.has(owner)) {
                    const fresh: number[] = [];
                    this.edges.set(owner, fresh);
                }
                const edges = this.edges.get(owner) ?? panic('missing reference graph');
                edges.push(target);
            }
        }
        for(const child of node.children) {
            this.graphWalk(child, owner, block);
        }
    }
    cycle(index: number, declaration: number): boolean {
        if(!this.safe(index, declaration)) {
            return false;
        }
        const target = this.statement(declaration);
        if(target < 0) {
            return false;
        }
        const block = this.bindings.parent(target);
        const source = this.statement(index, block);
        if(source < 0 || source === target) {
            return false;
        }
        if(!this.built.includes(block)) {
            this.built.push(block);
            for(const statement of this.bindings.rules.parser.node(block).children) {
                this.graphWalk(statement, statement, block);
            }
        }
        const seen = [target];
        for(let at = 0; at < seen.length; at++) {
            const outgoing = this.edges.get(seen[at] ?? -1);
            if(outgoing === undefined) {
                continue;
            }
            for(const next of outgoing) {
                if(next === source) {
                    return true;
                }
                if(!seen.includes(next)) {
                    seen.push(next);
                }
            }
        }
        return false;
    }
    keyword(declaration: number): string {
        const parser = this.bindings.rules.parser;
        const kind = parser.node(declaration).kind;
        if(kind === 'ClassDeclaration' || kind === 'ClassExpression') {
            return 'class';
        }
        if(kind === 'EnumDeclaration') {
            return 'enum';
        }
        let current = declaration;
        while(current >= 0) {
            const node = parser.node(current);
            if(node.kind === 'VariableDeclarationList') {
                return node.semantic === '2' ? 'const' : node.semantic === '1' ? 'let' : node.semantic === '4' ? 'using' :
                    node.semantic === '6' ? 'await using' : 'var';
            }
            if(!['VariableDeclaration', 'BindingElement', 'ObjectBindingPattern', 'ArrayBindingPattern'].includes(node.kind)) {
                break;
            }
            current = this.bindings.parent(current);
        }
        return 'var';
    }
    later(index: number, declaration: number): boolean {
        let current = this.bindings.parent(index);
        while(current >= 0) {
            if(this.bindings.functionKind(this.bindings.rules.parser.node(current).kind)) {
                return !this.bindings.contains(current, declaration);
            }
            current = this.bindings.parent(current);
        }
        return false;
    }
    report(index: number, declaration: number): void {
        const parser = this.bindings.rules.parser;
        const name = `'${parser.node(index).text}'`;
        const kind = parser.node(declaration).kind;
        const repair = ' Move the declaration above its first use.';
        const order = 'the cost is reading order: the reader meets the name before learning what it is, and has to jump down the file to find out.';
        let message = '';
        if(['InterfaceDeclaration', 'TypeAliasDeclaration'].includes(kind) || this.typeReference(index)) {
            message = `${name} is used as a type above the line that declares it. A type is erased before the program runs, so nothing fails at runtime; ${order}`;
        }
        else if(kind === 'FunctionDeclaration') {
            message = `${name} is read above the function declaration that defines it. A function declaration is hoisted together with its body, so the read works at runtime; ${order}`;
        }
        else if(['ClassDeclaration', 'ClassExpression', 'VariableDeclaration', 'BindingElement', 'Parameter', 'EnumDeclaration'].includes(kind)) {
            const keyword = this.keyword(declaration);
            const tdz = keyword !== 'var' && keyword !== 'enum';
            if(this.later(index, declaration)) {
                message = tdz ? `${name} is read inside a function written above the \`${keyword}\` that declares it. That is safe only while nothing calls the function before the declaration runs: an earlier call, such as one made while the module is still loading, reaches the binding in its temporal dead zone and throws a ReferenceError.` :
                    `${name} is read inside a function written above the \`${keyword}\` that declares it. The binding is hoisted without its value, so a call made before the declaration runs reads \`undefined\` silently, and the failure surfaces later and somewhere else.`;
            }
            else {
                message = tdz ? `${name} is read above the \`${keyword}\` that declares it. The binding exists from the top of its block but cannot be touched until its declaration runs, so this read lands in the temporal dead zone and throws a ReferenceError.` :
                    `${name} is read above the \`${keyword}\` that declares it. The binding is hoisted without its value, so this read silently yields \`undefined\`, and the failure surfaces later and somewhere else.`;
            }
        }
        else {
            message = `${name} is read above the line that declares it, so the reader meets the name before learning what it is.`;
        }
        const node = parser.node(index);
        this.bindings.rules.findings.push(new Diagnostic('no-use-before-define', 'usedBeforeDefined', message + repair,
            this.bindings.rules.byte(this.bindings.rules.start(node)), this.bindings.rules.byte(node.end), ''));
    }
    run(): void {
        const parser = this.bindings.rules.parser;
        for(let index = 0; index < parser.nodes.length; index++) {
            if(this.bindings.parent(index) < 0 || parser.node(index).kind !== 'Identifier' || this.bindings.declaring(index) ||
                this.typeReference(index) || this.query(index)) {
                continue;
            }
            const declaration = this.bindings.earliest(index);
            if(declaration < 0) {
                continue;
            }
            const name = this.bindings.shape.name(declaration);
            if(name < 0) {
                continue;
            }
            if(['ClassDeclaration', 'ClassExpression'].includes(parser.node(declaration).kind) &&
                parser.node(declaration).children.some((child) => parser.node(child).kind === 'Decorator' &&
                    this.bindings.contains(child, index))) {
                continue;
            }
            const above = parser.node(index).end < parser.node(name).end;
            const initializing = this.initializing(index, declaration);
            if(!above && !initializing) {
                continue;
            }
            if(above && !initializing && this.cycle(index, declaration)) {
                continue;
            }
            this.report(index, declaration);
        }
    }
}
