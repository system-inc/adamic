import { panic } from 'adamic';
import type { Rules } from './rules.ts';
import type { Types } from './types.ts';
import type { TypeFact } from './type_fact.ts';
import { types } from './facts.ts';
import { CheckerFacts, Properties, TypeMetadata } from './checker_facts.ts';
import { Frames, header } from './frames.ts';
import { Bindings } from './bindings.ts';
import { Diagnostic } from './diagnostic.ts';

export class FlowPair {
    readonly source: Types;
    readonly target: Types;
    readonly s: TypeFact;
    readonly t: TypeFact;
    readonly path: string;
    readonly mutable: boolean;
    property = '';
    propertyType = '';
    constructor(source: Types, s: TypeFact, target: Types, t: TypeFact, path = '', mutable = false) {
        this.source = source; this.target = target; this.s = s; this.t = t; this.path = path; this.mutable = mutable;
    }
}
export class FlowSignatures {
    readonly count: number;
    readonly names: readonly string[];
    readonly rests: readonly boolean[];
    readonly graph: Types;
    constructor(count: number, names: readonly string[], rests: readonly boolean[], graph: Types) {
        this.count = count; this.names = names; this.rests = rests; this.graph = graph;
    }
}
export class Flow {
    readonly rules: Rules;
    readonly bindings: Bindings;
    readonly facts: CheckerFacts;
    readonly propertyCache = new Map<number, Properties>();
    readonly metadataCache = new Map<number, TypeMetadata>();
    readonly signatureCache = new Map<number, FlowSignatures>();
    readonly originCache = new Map<number, string>();
    readonly referenceCache = new Map<number, Types>();
    readonly assignableCache = new Map<string, boolean>();
    readonly visited: string[] = [];
    siteIndex = 0;
    newContainer = false;
    exact = false;
    optional = false;
    missingName = '';
    missingType = '';
    constructor(bindings: Bindings) { this.bindings = bindings; this.rules = bindings.rules; this.facts = new CheckerFacts(this.rules); }
    properties(type: TypeFact): Properties {
        if(!this.propertyCache.has(type.id)) { this.propertyCache.set(type.id, this.facts.properties(this.siteIndex, type)); }
        return this.propertyCache.get(type.id) ?? panic('missing flow properties');
    }
    metadata(type: TypeFact): TypeMetadata {
        if(!this.metadataCache.has(type.id)) { this.metadataCache.set(type.id, this.facts.metadata(this.siteIndex, type)); }
        return this.metadataCache.get(type.id) ?? panic('missing flow metadata');
    }
    assignable(source: number, target: number): boolean {
        const key = `${source}:${target}`;
        if(!this.assignableCache.has(key)) { this.assignableCache.set(key, this.facts.compare(this.siteIndex, source, target)); }
        return this.assignableCache.get(key) ?? panic('missing flow relation');
    }
    library(type: TypeFact): string {
        if(type.target === 0 || (this.metadata(type).objectFlags & 4) === 0) { return ''; }
        if(!this.originCache.has(type.target)) {
            const frames = new Frames(this.rules.ask(this.siteIndex, `type-origin\n${type.target}`));
            header(frames, 'type-origin'); let name = ''; let declarationFile = false;
            if(frames.yes()) {
                name = frames.field(); const count = frames.natural();
                for(let at = 0; at < count; at++) { frames.field(); if(frames.yes()) { declarationFile = true; } frames.yes(); }
            }
            frames.end(); this.originCache.set(type.target, declarationFile ? name : '');
        }
        return this.originCache.get(type.target) ?? '';
    }
    signatures(type: TypeFact): FlowSignatures {
        if(!this.signatureCache.has(type.id)) {
            const frames = new Frames(this.rules.ask(this.siteIndex, `function-signatures\n${type.id}`));
            header(frames, 'function-signatures'); const count = frames.natural();
            const names: string[] = []; const rests: boolean[] = [];
            for(let at = 0; at < count; at++) {
                const parameters = frames.natural();
                for(let p = 0; p < parameters; p++) { names.push(frames.field()); rests.push(frames.yes()); }
            }
            const graph = types(`1\n119\nfunction-signatures${frames.text.slice(frames.cursor)}`, 'function-signatures');
            this.signatureCache.set(type.id, new FlowSignatures(count, names, rests, graph));
        }
        return this.signatureCache.get(type.id) ?? panic('missing flow signatures');
    }
    propertyExists(type: TypeFact, key: number): boolean {
        const frames = new Frames(this.rules.ask(this.siteIndex, `property-exists\n${type.id}\n${key}`));
        header(frames, 'property-exists'); const result = frames.yes(); frames.end(); return result;
    }
    expanded(type: TypeFact): Types {
        if(!this.referenceCache.has(type.id)) {
            this.referenceCache.set(type.id, types(this.rules.ask(this.siteIndex, `reference-shape\n${type.id}`), 'reference-shape'));
        }
        return this.referenceCache.get(type.id) ?? panic('missing reference arguments');
    }
    relate(pair: FlowPair, depth: number): FlowPair | undefined {
        if(pair.s.id === pair.t.id || depth > 8) { return undefined; }
        const key = `${pair.s.id}:${pair.t.id}`;
        if(this.visited.includes(key)) { return undefined; }
        this.visited.push(key);
        if((pair.t.flags & 1048576) !== 0 && this.metadata(pair.t).isClass()) { return undefined; }
        if(!this.optional) {
            if(pair.mutable && !this.assignable(pair.t.id, pair.s.id)) { return pair; }
        }
        else if((pair.t.flags & 1048576) !== 0 && (pair.s.flags & (134217728 | 1)) === 0 &&
            !pair.t.array && pair.t.tuple < 0 && !(this.exact && pair.path === '')) {
            const apparent = types(this.rules.ask(this.siteIndex, `apparent-shape\n${pair.s.id}`), 'apparent-shape');
            if(apparent.present && (apparent.root().flags & (1048576 | 268435456)) !== 0) {
                const target = this.properties(pair.t);
                for(let at = 0; at < target.names.length; at++) {
                    if(((target.flags[at] ?? 0) & 16777216) !== 0 && !this.propertyExists(apparent.root(), target.keys[at] ?? 0)) {
                        pair.property = target.names[at] ?? '';
                        this.missingName = pair.property;
                        pair.propertyType = this.rules.name(target.types.root(at), this.siteIndex);
                        this.missingType = pair.propertyType;
                        return pair;
                    }
                }
            }
        }
        if((pair.s.flags & 134217728) !== 0) {
            for(const member of pair.s.parts) {
                const found = this.relate(new FlowPair(pair.source, pair.source.type(member), pair.target, pair.t, pair.path), depth + 1);
                if(found !== undefined) { return found; }
            }
            return undefined;
        }
        if((pair.t.flags & 134217728) !== 0) {
            let first: FlowPair | undefined;
            for(const member of pair.t.parts) {
                if(!this.assignable(pair.s.id, member)) { continue; }
                const found = this.relate(new FlowPair(pair.source, pair.s, pair.target, pair.target.type(member), pair.path), depth + 1);
                if(found === undefined) { return undefined; }
                if(first === undefined) { first = found; }
            }
            return first;
        }
        if(pair.s.target === 0 && (pair.s.flags & 1048576) !== 0 && (this.metadata(pair.s).objectFlags & 4) !== 0) {
            const expanded = this.expanded(pair.s); pair = new FlowPair(expanded, expanded.root(), pair.target, pair.t, pair.path, pair.mutable);
        }
        if(pair.t.target === 0 && (pair.t.flags & 1048576) !== 0 && (this.metadata(pair.t).objectFlags & 4) !== 0) {
            const expanded = this.expanded(pair.t); pair = new FlowPair(pair.source, pair.s, expanded, expanded.root(), pair.path, pair.mutable);
        }
        const shared = !(this.newContainer && pair.path === '');
        if(pair.t.array) {
            const element = pair.t.arguments[0] ?? 0;
            if(element === 0) { return undefined; }
            const mutable = shared && this.library(pair.t) !== 'ReadonlyArray';
            const sourceElements: number[] = [];
            if(pair.s.array) { for(const id of pair.s.arguments.slice(0, 1)) { sourceElements.push(id); } }
            else if(pair.s.tuple >= 0) { for(const id of pair.s.arguments) { sourceElements.push(id); } }
            for(const source of sourceElements) {
                const found = this.relate(new FlowPair(pair.source, pair.source.type(source), pair.target, pair.target.type(element), `${pair.path}[]`, mutable), depth + 1);
                if(found !== undefined) { return found; }
            }
            return undefined;
        }
        if(pair.t.tuple >= 0) {
            if(pair.s.tuple < 0) { return undefined; }
            const mutable = shared && !this.metadata(pair.t).readonlyTuple;
            for(let at = 0; at < Math.min(pair.s.arguments.length, pair.t.arguments.length); at++) {
                const found = this.relate(new FlowPair(pair.source, pair.source.type(pair.s.arguments[at] ?? 0), pair.target,
                    pair.target.type(pair.t.arguments[at] ?? 0), `${pair.path}[${at}]`, mutable), depth + 1);
                if(found !== undefined) { return found; }
            }
            return undefined;
        }
        const container = this.library(pair.t);
        if(['Map', 'Set', 'WeakMap', 'WeakSet', 'ReadonlyMap', 'ReadonlySet'].includes(container)) {
            const source = this.library(pair.s);
            const base = container.startsWith('Readonly') ? container.slice(8) : container;
            const sourceBase = source.startsWith('Readonly') ? source.slice(8) : source;
            if(sourceBase !== base) { return undefined; }
            for(let at = 0; at < Math.min(pair.s.arguments.length, pair.t.arguments.length); at++) {
                const label = ['Map', 'WeakMap'].includes(base) ? at === 0 ? 'key' : 'value' : 'member';
                const found = this.relate(new FlowPair(pair.source, pair.source.type(pair.s.arguments[at] ?? 0), pair.target,
                    pair.target.type(pair.t.arguments[at] ?? 0), `${pair.path}<${container} ${label}>`, shared && !container.startsWith('Readonly')), depth + 1);
                if(found !== undefined) { return found; }
            }
            return undefined;
        }
        if((pair.s.flags & 1048576) === 0 || (pair.t.flags & 1048576) === 0) { return undefined; }
        const ss = this.signatures(pair.s); const ts = this.signatures(pair.t);
        if(ss.count === 1 && ts.count === 1) {
            const found = this.relate(new FlowPair(ss.graph, ss.graph.root(), ts.graph, ts.graph.root(), `${pair.path}()`), depth + 1);
            if(found !== undefined) { return found; }
            for(let at = 0; at < Math.min(ss.names.length, ts.names.length); at++) {
                if(ss.rests[at] === true || ts.rests[at] === true) { break; }
                const parameter = this.relate(new FlowPair(ts.graph, ts.graph.root(at + 1), ss.graph,
                    ss.graph.root(at + 1), `${pair.path}(${ss.names[at] ?? ''})`), depth + 1);
                if(parameter !== undefined) { return parameter; }
            }
        }
        const sp = this.properties(pair.s); const tp = this.properties(pair.t);
        for(let at = 0; at < tp.names.length; at++) {
            if(((tp.flags[at] ?? 0) & 8192) !== 0) { continue; }
            const match = sp.names.indexOf(tp.names[at] ?? ''); if(match < 0) { continue; }
            const found = this.relate(new FlowPair(sp.types, sp.types.root(match), tp.types, tp.types.root(at),
                `${pair.path}.${tp.names[at] ?? ''}`, shared && tp.readonlys[at] !== true), depth + 1);
            if(found !== undefined) { return found; }
        }
        return undefined;
    }
    annotation(index: number): number {
        const node = this.rules.parser.node(index);
        const name = node.kind === 'PropertyDeclaration' ? node.children.find((child) =>
            ['Identifier', 'PrivateIdentifier', 'StringLiteral', 'NumericLiteral', 'ComputedPropertyName'].includes(
                this.rules.parser.node(child).kind)) ?? -1 : this.bindings.shape.name(index);
        const initializer = this.bindings.shape.initializer(index);
        const candidates = node.children.filter((child) => child !== name && child !== initializer &&
            !['QuestionToken', 'ExclamationToken', 'DotDotDotToken', 'PublicKeyword', 'PrivateKeyword', 'ProtectedKeyword',
                'ReadonlyKeyword', 'StaticKeyword', 'AbstractKeyword', 'DeclareKeyword', 'OverrideKeyword', 'Decorator'].includes(this.rules.parser.node(child).kind));
        return candidates[0] ?? -1;
    }
    fresh(index: number, nothing = false): boolean {
        const node = this.rules.parser.node(this.rules.skip(index));
        if(['ObjectLiteralExpression', 'ArrayLiteralExpression'].includes(node.kind)) { return true; }
        if(nothing && (node.kind === 'NullKeyword' || (node.kind === 'Identifier' && node.text === 'undefined'))) { return true; }
        return node.kind === 'ConditionalExpression' && this.fresh(node.children[2] ?? -1, true) && this.fresh(node.children[4] ?? -1, true);
    }
    made(index: number): boolean {
        const node = this.rules.parser.node(this.rules.skip(index));
        if(node.kind === 'NewExpression') { return true; }
        if(node.kind !== 'CallExpression') { return false; }
        const callee = this.rules.parser.node(this.rules.skip(node.children[0] ?? -1));
        if(callee.kind !== 'PropertyAccessExpression') { return false; }
        const name = this.rules.parser.node(callee.children[callee.children.length - 1] ?? -1).text;
        const receiver = this.rules.parser.node(this.rules.skip(callee.children[0] ?? -1));
        if(receiver.kind === 'Identifier' && ['Array', 'Object'].includes(receiver.text)) {
            return receiver.text === 'Array' ? ['from', 'of'].includes(name) : ['keys', 'values', 'entries', 'fromEntries'].includes(name);
        }
        return ['map', 'filter', 'slice', 'concat', 'flat', 'flatMap', 'toSorted', 'toReversed', 'toSpliced', 'with'].includes(name);
    }
    literalAlias(index: number): boolean {
        const identifier = this.rules.skip(index); if(this.rules.parser.node(identifier).kind !== 'Identifier') { return false; }
        const declarations = this.bindings.list(identifier, 'binding-declarations');
        if(declarations.length !== 1) { return false; }
        const declaration = declarations[0] ?? -1; const node = this.rules.parser.node(declaration);
        const parent = this.bindings.parent(declaration);
        const initializer = this.bindings.shape.initializer(declaration);
        return node.kind === 'VariableDeclaration' && parent >= 0 && this.rules.parser.node(parent).semantic === '2' &&
            this.annotation(declaration) < 0 && initializer >= 0 && this.rules.parser.node(this.rules.skip(initializer)).kind === 'ObjectLiteralExpression';
    }
    offer(index: number, target: Types, upcast = false): void {
        if(!target.present || !target.has(target.root(), 1048576)) { return; }
        const source = types(this.rules.ask(index, 'raw-shape'), 'raw-shape');
        if(source.root().id === target.root().id) { return; }
        this.siteIndex = index;
        if(upcast && !this.assignable(source.root().id, target.root().id)) { return; }
        this.newContainer = this.made(index); this.exact = this.literalAlias(index);
        const fresh = this.fresh(index); const pair = new FlowPair(source, source.root(), target, target.root());
        for(const optional of [false, true]) {
            if(optional && fresh) { continue; }
            this.optional = optional; this.missingName = ''; this.missingType = ''; this.visited.splice(0, this.visited.length);
            let found: FlowPair | undefined;
            if(!fresh) { found = this.relate(pair, 0); }
            // The mutable rule's top judge sees a non-mutable pair, so a fresh site passes.
            if(found === undefined) { continue; }
            let expression = this.rules.text(this.rules.parser.node(index));
            if(this.rules.byte(this.rules.parser.node(index).end) - this.rules.byte(this.rules.start(this.rules.parser.node(index))) > 40 ||
                expression.includes('\n') || expression.includes('\r')) { expression = 'value'; }
            const slot = `${expression}${found.path}`;
            const sourceName = this.rules.name(optional ? found.s : source.root(), index);
            const targetName = this.rules.name(optional ? found.t : target.root(), index);
            const message = optional ?
                `\`${slot}\` is a '${sourceName}', seen here as '${targetName}', which adds the optional property \`${this.missingName}\` that '${sourceName}' does not declare. A '${sourceName}' can still carry a \`${this.missingName}\` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '${this.missingType}' and holds something else. Declare \`${this.missingName}\` in the source's type, or build the value with the property named.` :
                `'${sourceName}' is seen here as '${targetName}', and \`${slot}\` is mutable. Through the wider type anything that is a '${this.rules.name(found.t, index)}' can be written there, and the original, still typed '${this.rules.name(found.s, index)}', then holds it: TypeScript accepts this, and it is a TypeError at runtime. Make the slot read-only in the target type (\`readonly T[]\`, a \`readonly\` property, \`ReadonlyMap\`), so the narrower value is only read through it, or copy the value ([...items], { ...record }).`;
            const node = this.rules.parser.node(index);
            this.rules.findings.push(new Diagnostic(optional ? 'adamic/no-optional-widening' : 'adamic/invariant-mutable',
                optional ? 'optionalWidening' : 'mutableWidening', message, this.rules.byte(this.rules.start(node)), this.rules.byte(node.end), ''));
        }
    }
    contextual(index: number): void { this.offer(index, types(this.rules.ask(index, 'contextual-shape'), 'contextual-shape')); }
    destructuring(index: number): boolean {
        let current = index;
        while(current >= 0) {
            const parent = this.bindings.parent(current); if(parent < 0) { return false; }
            const node = this.rules.parser.node(parent);
            if(node.kind === 'BinaryExpression') { return this.rules.parser.node(node.children[1] ?? -1).kind === 'EqualsToken' && node.children[0] === current; }
            if(['ForOfStatement', 'ForInStatement'].includes(node.kind)) { return node.children[0] === current; }
            if(!['PropertyAssignment', 'ShorthandPropertyAssignment', 'ParenthesizedExpression', 'ArrayLiteralExpression',
                'ObjectLiteralExpression', 'SpreadElement', 'SpreadAssignment'].includes(node.kind)) { return false; }
            current = parent;
        }
        return false;
    }
    returned(functionIndex: number, expression: number): void {
        if(functionIndex < 0) { return; }
        const node = this.rules.parser.node(functionIndex);
        if(node.children.some((child) => ['AsyncKeyword', 'AsteriskToken'].includes(this.rules.parser.node(child).kind))) { return; }
        const body = node.children[node.children.length - 1] ?? -1;
        const before = node.children[node.children.length - 2] ?? -1;
        let annotation = before;
        if(node.kind === 'ArrowFunction') { annotation = node.children[node.children.length - 3] ?? -1; }
        if(annotation < 0 || body < 0 || ['Identifier', 'Parameter', 'TypeParameter', 'EqualsGreaterThanToken'].includes(this.rules.parser.node(annotation).kind)) { return; }
        this.offer(expression, types(this.rules.ask(functionIndex, 'annotated-return-shape'), 'annotated-return-shape'));
    }
    order(index: number, out: number[]): void {
        out.push(index);
        for(const child of this.rules.parser.node(index).children) { this.order(child, out); }
    }
    run(): void {
        const ordered: number[] = [];
        const root = this.rules.parser.nodes.findIndex((node) => node.kind === 'SourceFile');
        this.order(root, ordered);
        for(const index of ordered) {
            if(this.bindings.parent(index) < 0) { continue; }
            const node = this.rules.parser.node(index);
            if(['VariableDeclaration', 'PropertyDeclaration', 'Parameter'].includes(node.kind)) {
                const initializer = this.bindings.shape.initializer(index); const annotation = this.annotation(index);
                if(initializer >= 0 && annotation >= 0) { this.offer(initializer, types(this.rules.ask(annotation, 'annotation-shape'), 'annotation-shape')); }
            }
            else if(node.kind === 'BinaryExpression' && this.rules.parser.node(node.children[1] ?? -1).kind === 'EqualsToken') {
                const left = node.children[0] ?? -1;
                if(!['ObjectLiteralExpression', 'ArrayLiteralExpression'].includes(this.rules.parser.node(this.rules.skip(left)).kind)) {
                    this.offer(node.children[2] ?? -1, types(this.rules.ask(left, 'raw-shape'), 'raw-shape'));
                }
            }
            else if(['CallExpression', 'NewExpression'].includes(node.kind)) {
                let arguments_: number[] = [];
                if(node.list > 0) { arguments_ = node.children.slice(node.children.length - node.list); }
                for(let at = 0; at < arguments_.length; at++) {
                    const argument = arguments_[at] ?? -1;
                    if(this.rules.parser.node(argument).kind !== 'SpreadElement') {
                        this.offer(argument, types(this.rules.ask(index, `contextual-argument\n${at}`), 'contextual-argument'));
                    }
                }
            }
            else if(node.kind === 'ReturnStatement' && node.children.length > 0) {
                let current = this.bindings.parent(index);
                while(current >= 0 && !this.bindings.functionKind(this.rules.parser.node(current).kind)) { current = this.bindings.parent(current); }
                this.returned(current, node.children[0] ?? -1);
            }
            else if(node.kind === 'ArrowFunction') {
                const body = node.children[node.children.length - 1] ?? -1;
                if(this.rules.parser.node(body).kind !== 'Block') { this.returned(index, body); }
            }
            else if(node.kind === 'PropertyAssignment' && !this.destructuring(this.bindings.parent(index))) { this.contextual(node.children[node.children.length - 1] ?? -1); }
            else if(node.kind === 'ShorthandPropertyAssignment' && !this.destructuring(this.bindings.parent(index)) && node.children.length === 1) { this.contextual(node.children[0] ?? -1); }
            else if(node.kind === 'ArrayLiteralExpression' && !this.destructuring(index)) {
                for(const element of node.children) {
                    if(!['SpreadElement', 'OmittedExpression'].includes(this.rules.parser.node(element).kind)) { this.contextual(element); }
                }
            }
            else if(['AsExpression', 'TypeAssertionExpression'].includes(node.kind)) {
                const annotation = node.kind === 'AsExpression' ? node.children[1] ?? -1 : node.children[0] ?? -1;
                const expression = node.kind === 'AsExpression' ? node.children[0] ?? -1 : node.children[1] ?? -1;
                const typeNode = this.rules.parser.node(annotation);
                if(typeNode.kind === 'TypeReference' && typeNode.children.length === 1 && this.rules.parser.node(typeNode.children[0] ?? -1).text === 'const') { continue; }
                this.offer(expression, types(this.rules.ask(annotation, 'annotation-shape'), 'annotation-shape'), true);
            }
        }
    }
}
