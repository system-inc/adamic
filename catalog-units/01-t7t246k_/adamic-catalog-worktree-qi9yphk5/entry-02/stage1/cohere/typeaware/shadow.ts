import { panic } from 'adamic';
import type { Rules } from './rules.ts';
import { Binding } from './binding.ts';
import { Scope } from './scope.ts';
import { Frames, header } from './frames.ts';
import { Diagnostic } from './diagnostic.ts';

export class Shadow {
    readonly rules: Rules;
    readonly ordered: number[] = [];
    readonly scopes: Scope[] = [];
    readonly namedScopes = new Map<string, number[]>();
    readonly lineStarts: number[] = [];
    constructor(rules: Rules) {
        this.rules = rules;
    }
    lookup(kind: string, start: number, end: number): number {
        let first = 0;
        let last = this.ordered.length;
        while(first < last) {
            const mid = Math.floor((first + last) / 2);
            const node = this.rules.parser.node(this.ordered[mid] ?? -1);
            if(this.rules.byte(node.pos) < start) {
                first = mid + 1;
            }
            else {
                last = mid;
            }
        }
        for(let at = first; at < this.ordered.length; at++) {
            const index = this.ordered[at] ?? -1;
            const node = this.rules.parser.node(index);
            if(this.rules.byte(node.pos) !== start) {
                break;
            }
            if(this.rules.byte(node.end) === end && node.kind === kind) {
                return index;
            }
        }
        panic('binder declaration absent from Adamic parser');
    }
    name(index: number): number {
        const node = this.rules.parser.node(index);
        if(node.kind === 'BindingElement') {
            const initializer = this.initializer(index);
            const names = node.children.filter(
                (child) => child !== initializer && this.rules.parser.node(child).kind !== 'DotDotDotToken',
            );
            return names[names.length - 1] ?? -1;
        }
        if(['ImportSpecifier', 'NamespaceImport'].includes(node.kind)) {
            return node.children[node.children.length - 1] ?? -1;
        }
        for(const child of node.children) {
            const kind = this.rules.parser.node(child).kind;
            if(['Identifier', 'ThisKeyword', 'ArrayBindingPattern', 'ObjectBindingPattern'].includes(kind)) {
                return child;
            }
        }
        return -1;
    }
    value(binding: Binding): boolean {
        if(binding.flags === 0) {
            return true;
        }
        if((binding.flags & 2097152) !== 0) {
            let current = binding.declarations[0] ?? -1;
            while(current >= 0) {
                const node = this.rules.parser.node(current);
                if(node.kind === 'ImportSpecifier' && node.semantic === '1') {
                    return false;
                }
                if(node.kind === 'ImportClause') {
                    return node.semantic !== 'TypeKeyword';
                }
                if(node.kind === 'ImportDeclaration') {
                    break;
                }
                current = this.rules.parents[current] ?? -1;
            }
            return true;
        }
        return (binding.flags & (3 | 4 | 8 | 16 | 32 | 128 | 256 | 512 | 4096 | 8192 | 32768 | 65536)) !== 0;
    }
    body(index: number): boolean {
        return this.rules.parser.node(index).children.some((child) => this.rules.parser.node(child).kind === 'Block');
    }
    initializer(index: number): number {
        const node = this.rules.parser.node(index);
        if(!['VariableDeclaration', 'PropertyDeclaration', 'Parameter', 'BindingElement'].includes(node.kind)) {
            return -1;
        }
        const child = node.children[node.children.length - 1] ?? -1;
        if(child < 0) {
            return -1;
        }
        this.rules.scanner.pos = this.rules.parser.node(child).pos;
        // The equals is consumed before an initializer's leading trivia.
        let at = this.rules.parser.node(child).pos - 1;
        while(at >= node.pos && [' ', '\t', '\r', '\n'].includes(this.rules.scanner.text.slice(at, at + 1))) {
            at--;
        }
        return this.rules.scanner.text.slice(at, at + 1) === '=' ? child : -1;
    }
    report(scopeIndex: number, binding: Binding): void {
        if(binding.name === 'this') {
            return;
        }
        const container = this.rules.parser.node(this.scopes[scopeIndex]?.container ?? -1);
        let outerScope = -1;
        let outerIndex = -1;
        const candidates = this.namedScopes.get(binding.name) ?? panic('missing binding index');
        for(let at = candidates.length - 2; at >= 0; at -= 2) {
            const candidate = candidates[at] ?? -1;
            if(candidate >= scopeIndex) {
                continue;
            }
            const scope = this.scopes[candidate] ?? panic('missing scope');
            const node = this.rules.parser.node(scope.container);
            if(node.pos > container.pos || node.end < container.end) {
                continue;
            }
            outerScope = candidate;
            outerIndex = candidates[at + 1] ?? -1;
            break;
        }
        if(outerIndex < 0) {
            const own = this.scopes[scopeIndex] ?? panic('missing scope');
            for(let at = 0; at < own.bindings.length; at++) {
                const candidate = own.bindings[at] ?? panic('missing binding');
                if(
                    candidate.name === binding.name &&
                    candidate.identifier !== binding.identifier &&
                    candidate.flags === 0
                ) {
                    outerScope = scopeIndex;
                    outerIndex = at;
                    break;
                }
            }
        }
        if(outerIndex < 0) {
            return;
        }
        const enclosing = this.scopes[outerScope] ?? panic('missing enclosing scope');
        const outer = enclosing.bindings[outerIndex] ?? panic('missing enclosing binding');
        const declaration = binding.declarations[0] ?? -1;
        const oldDeclaration = outer.declarations[0] ?? -1;
        if(
            (binding.flags === 0 || (binding.flags & 2997247) !== 0) &&
            (outer.flags === 0 || (outer.flags & 2997247) !== 0) &&
            this.value(binding) !== this.value(outer)
        ) {
            return;
        }
        if(
            this.value(outer) &&
            binding.declarations.every((decl) => {
                const node = this.rules.parser.node(decl);
                const functionIndex =
                    node.kind === 'Parameter'
                        ? (this.rules.parents[decl] ?? -1)
                        : node.kind === 'FunctionDeclaration'
                          ? decl
                          : -1;
                if(functionIndex < 0) {
                    return false;
                }
                const kind = this.rules.parser.node(functionIndex).kind;
                return (
                    [
                        'FunctionType',
                        'ConstructorType',
                        'CallSignature',
                        'ConstructSignature',
                        'MethodSignature',
                    ].includes(kind) ||
                    (['FunctionDeclaration', 'MethodDeclaration', 'Constructor', 'GetAccessor', 'SetAccessor'].includes(
                        kind,
                    ) &&
                        !this.body(functionIndex))
                );
            })
        ) {
            return;
        }
        if(
            this.rules.parser.node(declaration).kind === 'TypeParameter' &&
            this.rules.parser.node(oldDeclaration).kind === 'TypeParameter'
        ) {
            const method = this.rules.parents[declaration] ?? -1;
            const classIndex = this.rules.parents[oldDeclaration] ?? -1;
            if(
                method >= 0 &&
                classIndex >= 0 &&
                this.rules.parser.node(method).kind === 'MethodDeclaration' &&
                this.rules.parser
                    .node(method)
                    .children.some((child) => this.rules.parser.node(child).kind === 'StaticKeyword') &&
                ['ClassDeclaration', 'ClassExpression'].includes(this.rules.parser.node(classIndex).kind)
            ) {
                return;
            }
            if(
                this.rules.parser.node(method).kind === 'InferType' &&
                this.rules.parser.node(classIndex).kind === 'InferType'
            ) {
                return;
            }
        }
        if(binding.flags === 0) {
            let current = declaration;
            while((this.rules.parents[current] ?? -1) >= 0) {
                const parentIndex = this.rules.parents[current] ?? -1;
                const parent = this.rules.parser.node(parentIndex);
                const transparent =
                    parent.kind === 'ParenthesizedExpression' ||
                    (parent.kind === 'ConditionalExpression' && parent.children[0] !== current) ||
                    (parent.kind === 'BinaryExpression' &&
                        ['AmpersandAmpersandToken', 'BarBarToken', 'QuestionQuestionToken'].includes(
                            this.rules.parser.node(parent.children[1] ?? -1).kind,
                        ));
                if(!transparent) {
                    break;
                }
                current = parentIndex;
            }
            if(this.initializer(oldDeclaration) === current) {
                return;
            }
        }
        const identifier = this.rules.parser.node(binding.identifier);
        const previous = this.rules.parser.node(outer.identifier);
        const old = this.rules.parser.node(oldDeclaration);
        if(
            identifier.end < previous.pos &&
            !(
                (old.kind === 'FunctionDeclaration' && this.body(oldDeclaration)) ||
                ['InterfaceDeclaration', 'TypeAliasDeclaration'].includes(old.kind)
            )
        ) {
            return;
        }
        let current = this.scopes[scopeIndex]?.container ?? -1;
        while(current >= 0) {
            const node = this.rules.parser.node(current);
            if(
                node.kind === 'ModuleDeclaration' &&
                node.children.some((child) => this.rules.parser.node(child).text === 'global')
            ) {
                return;
            }
            current = this.rules.parents[current] ?? -1;
        }
        const previousStart = this.rules.start(previous);
        let low = 0;
        let high = this.lineStarts.length;
        while(low < high) {
            const mid = Math.floor((low + high) / 2);
            if((this.lineStarts[mid] ?? 0) <= previousStart) {
                low = mid + 1;
            }
            else {
                high = mid;
            }
        }
        const line = low;
        const column = previousStart - (this.lineStarts[low - 1] ?? 0) + 1;
        let end = identifier.end;
        const parent = this.rules.parser.node(declaration);
        if(
            ['VariableDeclaration', 'Parameter'].includes(parent.kind) &&
            !parent.children.some((child) => this.rules.parser.node(child).kind === 'DotDotDotToken')
        ) {
            const initialized = this.initializer(declaration);
            for(const child of parent.children.slice(parent.children.indexOf(binding.identifier) + 1)) {
                if(child === initialized) {
                    break;
                }
                end = this.rules.parser.node(child).end;
            }
        }
        const isEnum = outer.declarations.some((decl) => this.rules.parser.node(decl).kind === 'EnumDeclaration');
        const message = isEnum
            ? "This enum member takes a name that is already bound in an enclosing scope. Enum members are added to the enum's own scope, so a later member initializer that names it resolves to this member rather than to the outer binding. Rename the member."
            : 'This binding takes a name that is already bound in an enclosing scope. Every read of the name inside this scope now resolves here instead of to the outer binding, so code written against the outer one silently reads the inner one. Rename the inner binding.';
        this.rules.findings.push(
            new Diagnostic(
                'no-shadow',
                isEnum ? 'noEnumShadow' : 'noShadow',
                `${message} The name '${binding.name}' is already declared in the upper scope on line ${line} column ${column}.`,
                this.rules.byte(this.rules.start(identifier)),
                this.rules.byte(end),
            ),
        );
    }
    run(): void {
        this.lineStarts.push(0);
        for(let at = 0; at < this.rules.scanner.text.length; at++) {
            const code = this.rules.scanner.text.charCodeAt(at);
            if(code === 13 || code === 10 || code === 8232 || code === 8233) {
                if(code === 13 && this.rules.scanner.text.charCodeAt(at + 1) === 10) {
                    at++;
                }
                this.lineStarts.push(at + 1);
            }
        }
        for(let index = 0; index < this.rules.parser.nodes.length; index++) {
            this.ordered.push(index);
        }
        this.ordered.sort((a, b) => this.rules.parser.node(a).pos - this.rules.parser.node(b).pos);
        const frames = new Frames(this.rules.ask(this.rules.parser.nodes.length - 1, 'scope-locals'));
        header(frames, 'scope-locals');
        const count = frames.natural();
        for(let at = 0; at < count; at++) {
            const table = frames.field();
            const kind = frames.field();
            const start = frames.natural();
            const end = frames.natural();
            const container = this.lookup(kind, start, end);
            const bindings: Binding[] = [];
            const length = frames.natural();
            for(let symbol = 0; symbol < length; symbol++) {
                const name = frames.field();
                const flags = frames.natural();
                const declarationCount = frames.natural();
                const declarations: number[] = [];
                let identifier = -1;
                for(let decl = 0; decl < declarationCount; decl++) {
                    const declarationKind = frames.field();
                    const first = frames.natural();
                    const last = frames.natural();
                    const index = this.lookup(declarationKind, first, last);
                    declarations.push(index);
                    const candidate = this.name(index);
                    if(
                        identifier < 0 &&
                        candidate >= 0 &&
                        this.rules.parser.node(candidate).kind === 'Identifier' &&
                        !(
                            this.rules.parser.node(index).kind === 'Parameter' &&
                            this.rules.parser.node(this.rules.parents[index] ?? -1).kind === 'IndexSignature'
                        ) &&
                        (table !== 'exports' || (first >= start && last <= end))
                    ) {
                        identifier = candidate;
                    }
                }
                if(identifier >= 0) {
                    bindings.push(new Binding(name, identifier, declarations, flags));
                }
            }
            this.scopes.push(new Scope(container, bindings));
        }
        frames.end();
        // Type parameters on classes and interfaces are absent from locals.
        for(let index = 0; index < this.rules.parser.nodes.length; index++) {
            const node = this.rules.parser.node(index);
            if(['ClassDeclaration', 'ClassExpression', 'InterfaceDeclaration'].includes(node.kind)) {
                const scope = this.scopes.find((candidate) => candidate.container === index);
                const bindings: Binding[] = [];
                for(const child of node.children) {
                    if(this.rules.parser.node(child).kind === 'TypeParameter') {
                        const name = this.name(child);
                        if(name >= 0) {
                            const binding = new Binding(this.rules.parser.node(name).text, name, [child], 262144);
                            if(scope === undefined) {
                                bindings.push(binding);
                            }
                            else {
                                scope.bindings.push(binding);
                            }
                        }
                    }
                }
                if(bindings.length > 0) {
                    this.scopes.push(new Scope(index, bindings));
                }
            }
            if(node.kind === 'FunctionExpression' || node.kind === 'ClassExpression') {
                const name = this.name(index);
                if(name >= 0 && this.rules.parser.node(name).kind === 'Identifier') {
                    const scope = this.scopes.find((candidate) => candidate.container === index);
                    if(scope !== undefined) {
                        scope.bindings.push(new Binding(this.rules.parser.node(name).text, name, [index], 0));
                    }
                }
            }
        }
        this.scopes.sort((a, b) => {
            const x = this.rules.parser.node(a.container);
            const y = this.rules.parser.node(b.container);
            return x.pos === y.pos ? y.end - x.end : x.pos - y.pos;
        });
        // Keep the first binding in each scope, matching the original scan.
        // Candidate order is scope order, not declaration order or a name sort.
        for(let at = 0; at < this.scopes.length; at++) {
            const scope = this.scopes[at] ?? panic('missing scope');
            for(let slot = 0; slot < scope.bindings.length; slot++) {
                const binding = scope.bindings[slot] ?? panic('missing binding');
                if(!this.namedScopes.has(binding.name)) {
                    const fresh: number[] = [];
                    this.namedScopes.set(binding.name, fresh);
                }
                const candidates = this.namedScopes.get(binding.name) ?? panic('missing binding index');
                if(candidates[candidates.length - 2] !== at) {
                    candidates.push(at);
                    candidates.push(slot);
                }
            }
        }
        for(let at = 0; at < this.scopes.length; at++) {
            const scope = this.scopes[at] ?? panic('missing scope');
            for(const binding of scope.bindings) {
                this.report(at, binding);
            }
        }
    }
}
