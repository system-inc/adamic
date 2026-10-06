import { panic, tsgoInspect } from 'adamic';
import { Bindings } from './bindings.ts';
import { Unused } from './unused.ts';
import { Reassign } from './reassign.ts';
import { Frames, header } from './frames.ts';
import { Diagnostic } from './diagnostic.ts';
import { types } from './facts.ts';
import type { TypeFact } from './type_fact.ts';

export class DocTag {
    readonly kind: string; readonly name: string; readonly start: number; readonly end: number; readonly text: string;
    constructor(kind: string, name: string, start: number, end: number, text: string) {
        this.kind = kind; this.name = name; this.start = start; this.end = end; this.text = text;
    }
    reason(): string {
        const at = this.text.indexOf(`@${this.name}`);
        let text = at < 0 ? '' : this.text.slice(at + this.name.length + 1).trim();
        if(text.endsWith('*/')) { text = text.slice(0, -2); }
        const lines: string[] = [];
        for(const line of text.split('\n')) {
            let stripped = line.trim(); while(stripped.startsWith('*')) { stripped = stripped.slice(1); }
            stripped = stripped.trim(); if(stripped !== '') { lines.push(stripped); }
        }
        return lines.join(' ');
    }
}
export class DeclarationFact {
    readonly file: string; readonly kind: string; readonly start: number; readonly end: number;
    readonly declarationFile: boolean; readonly defaultLibrary: boolean;
    readonly parentKind: string; readonly parentName: string; readonly parentStart: number; readonly parentEnd: number;
    readonly tags: readonly DocTag[]; readonly parameters: readonly string[];
    constructor(file: string, kind: string, start: number, end: number, declarationFile: boolean, defaultLibrary: boolean,
        parentKind: string, parentName: string, parentStart: number, parentEnd: number, tags: readonly DocTag[], parameters: readonly string[]) {
        this.file = file; this.kind = kind; this.start = start; this.end = end; this.declarationFile = declarationFile; this.defaultLibrary = defaultLibrary;
        this.parentKind = parentKind; this.parentName = parentName; this.parentStart = parentStart; this.parentEnd = parentEnd; this.tags = tags; this.parameters = parameters;
    }
}
export class SymbolDetail {
    readonly present: boolean; readonly identity: number; readonly flags: number; readonly name: string; readonly declarations: readonly DeclarationFact[];
    constructor(present: boolean, identity: number, flags: number, name: string, declarations: readonly DeclarationFact[]) {
        this.present = present; this.identity = identity; this.flags = flags; this.name = name; this.declarations = declarations;
    }
}
export class Caller {
    readonly bindings: Bindings;
    readonly symbols: Unused;
    readonly writes: Reassign;
    readonly parameterCache = new Map<number, number>();
    constructor(bindings: Bindings) { this.bindings = bindings; this.symbols = new Unused(bindings); this.writes = new Reassign(bindings.rules); }
    declaration(frames: Frames): DeclarationFact {
        const file = frames.field(); const kind = frames.field(); const start = frames.natural(); const end = frames.natural();
        const declarationFile = frames.yes(); const library = frames.yes(); const parentKind = frames.field(); const parentName = frames.field();
        const parentStart = frames.natural(); const parentEnd = frames.natural(); const tags: DocTag[] = [];
        const count = frames.natural();
        for(let at = 0; at < count; at++) { tags.push(new DocTag(frames.field(), frames.field(), frames.natural(), frames.natural(), frames.field())); }
        const parameters: string[] = []; const length = frames.natural();
        for(let at = 0; at < length; at++) { parameters.push(frames.field()); }
        return new DeclarationFact(file, kind, start, end, declarationFile, library, parentKind, parentName, parentStart, parentEnd, tags, parameters);
    }
    symbol(frames: Frames): SymbolDetail {
        const present = frames.yes(); const declarations: DeclarationFact[] = [];
        let identity = 0; let flags = 0; let name = '';
        if(present) {
            identity = frames.natural(); flags = frames.natural(); name = frames.field();
            const count = frames.natural(); for(let at = 0; at < count; at++) { declarations.push(this.declaration(frames)); }
        }
        return new SymbolDetail(present, identity, flags, name, declarations);
    }
    nodeSymbol(index: number): SymbolDetail {
        const frames = new Frames(this.bindings.rules.ask(index, 'node-symbol-details')); header(frames, 'node-symbol-details');
        const result = this.symbol(frames); frames.end(); return result;
    }
    typeSymbols(index: number, type: TypeFact): SymbolDetail[] {
        const frames = new Frames(this.bindings.rules.ask(index, `type-symbol-details\n${type.id}`)); header(frames, 'type-symbol-details');
        const result = [this.symbol(frames), this.symbol(frames)]; frames.end(); return result;
    }
    nodeDeclaration(index: number): DeclarationFact {
        const frames = new Frames(this.bindings.rules.ask(index, 'declaration-details')); header(frames, 'declaration-details');
        const result = this.declaration(frames); frames.end(); return result;
    }
    identity(index: number): number {
        const ids = this.symbols.identity(index); return ids.shorthand === 0 ? ids.own : ids.shorthand;
    }
    root(index: number): number {
        const rules = this.bindings.rules; let current = index;
        while(current >= 0) {
            current = rules.skip(current); const node = rules.parser.node(current);
            if(['PropertyAccessExpression', 'ElementAccessExpression', 'NonNullExpression'].includes(node.kind)) { current = node.children[0] ?? -1; }
            else { return node.kind === 'Identifier' ? current : -1; }
        }
        return -1;
    }
    callback(index: number): boolean {
        const kind = this.bindings.rules.parser.node(index).kind;
        if(!['FunctionExpression', 'ArrowFunction'].includes(kind)) { return false; }
        let parent = this.bindings.parent(index);
        while(parent >= 0 && this.bindings.rules.parser.node(parent).kind === 'ParenthesizedExpression') { parent = this.bindings.parent(parent); }
        return parent >= 0 && ['CallExpression', 'NewExpression'].includes(this.bindings.rules.parser.node(parent).kind);
    }
    reassigned(index: number, symbol: number): boolean {
        const rules = this.bindings.rules; const node = rules.parser.node(index);
        if(node.kind === 'Identifier' && this.writes.writes(index) && this.identity(index) === symbol) { return true; }
        return node.children.some((child) => this.reassigned(child, symbol));
    }
    parameter(index: number): number {
        const root = this.root(index); if(root < 0) { return -1; }
        const symbol = this.identity(root); if(symbol === 0) { return -1; }
        if(!this.parameterCache.has(symbol)) {
            let parameter = -1;
            for(const declaration of this.bindings.declarations(root)) {
                let current = declaration;
                while(current >= 0 && ['BindingElement', 'ObjectBindingPattern', 'ArrayBindingPattern'].includes(this.bindings.rules.parser.node(current).kind)) { current = this.bindings.parent(current); }
                if(current >= 0 && this.bindings.rules.parser.node(current).kind === 'Parameter') { parameter = current; break; }
            }
            if(parameter >= 0) {
                const functionIndex = this.bindings.parent(parameter);
                if(functionIndex < 0 || this.callback(functionIndex) || this.reassigned(functionIndex, symbol)) { parameter = -1; }
            }
            this.parameterCache.set(symbol, parameter);
        }
        return this.parameterCache.get(symbol) ?? -1;
    }
    writesData(index: number): boolean {
        const rules = this.bindings.rules; const node = rules.parser.node(index);
        if(!['PropertyAccessExpression', 'ElementAccessExpression'].includes(node.kind)) { return false; }
        const key = node.children[node.children.length - 1] ?? -1; const receiver = node.children[0] ?? -1;
        if(node.kind === 'PropertyAccessExpression' || ['StringLiteral', 'NumericLiteral', 'NoSubstitutionTemplateLiteral'].includes(rules.parser.node(key).kind)) {
            const symbol = this.nodeSymbol(key);
            if(symbol.present && symbol.declarations.length > 0) {
                return (symbol.flags & 65536) === 0 && symbol.declarations.some((declaration) => !declaration.declarationFile);
            }
        }
        let owner = rules.skip(receiver);
        while(rules.parser.node(owner).kind === 'NonNullExpression') { owner = rules.skip(rules.parser.node(owner).children[0] ?? -1); }
        if(rules.parser.node(owner).kind !== 'Identifier' && !this.writesData(owner)) { return false; }
        const raw = types(rules.ask(receiver, 'raw-shape'), 'raw-shape');
        const apparent = types(rules.ask(receiver, `apparent-shape\n${raw.root().id}`), 'apparent-shape');
        if(!apparent.present) { return false; }
        const type = apparent.root(); if(type.array || type.tuple >= 0) { return true; }
        const symbols = this.typeSymbols(receiver, type);
        const own = symbols[0] ?? panic('missing type symbol'); const alias = symbols[1] ?? panic('missing alias symbol');
        if(['Record', 'Partial', 'Required', 'Pick', 'Omit'].includes(alias.name) && alias.declarations.some((declaration) => declaration.defaultLibrary)) { return true; }
        return own.declarations.some((declaration) => !declaration.declarationFile);
    }
    writer(method: number): boolean {
        const name = this.bindings.rules.parser.node(method).text; const symbol = this.nodeSymbol(method);
        if(!symbol.present || symbol.declarations.length === 0) { return false; }
        return symbol.declarations.every((declaration) => {
            if(!declaration.defaultLibrary || declaration.parentKind !== 'InterfaceDeclaration') { return false; }
            const owner = declaration.parentName;
            if(owner === 'Array') { return ['push', 'unshift', 'pop', 'shift', 'splice', 'sort', 'reverse', 'fill', 'copyWithin'].includes(name); }
            if(owner === 'Map') { return ['set', 'delete', 'clear'].includes(name); }
            if(owner === 'Set') { return ['add', 'delete', 'clear'].includes(name); }
            if(owner === 'WeakMap') { return ['set', 'delete'].includes(name); }
            return owner === 'WeakSet' && ['add', 'delete'].includes(name);
        });
    }
    processState(index: number): boolean {
        const rules = this.bindings.rules; const root = this.root(index); if(root < 0) { return false; }
        const declared = types(rules.ask(root, 'symbol-shape'), 'symbol-shape');
        if(!declared.present) { return false; }
        return this.typeSymbols(root, declared.root()).some((symbol) => symbol.declarations.some((declaration) =>
            ['InterfaceDeclaration', 'TypeAliasDeclaration', 'ClassDeclaration'].includes(declaration.kind) &&
            declaration.tags.some((tag) => tag.kind === 'JSDocUnknownTag' && tag.name === 'processState' && tag.reason() !== '')));
    }
    contract(declaration: DeclarationFact, position: number): boolean {
        return declaration.tags.some((tag) => {
            if(tag.kind !== 'JSDocUnknownTag' || tag.name !== 'mutates') { return false; }
            const reason = tag.reason(); const split = reason.indexOf(' ');
            return split >= 0 && reason.slice(split + 1).trim() !== '' && declaration.parameters.indexOf(reason.slice(0, split)) === position;
        });
    }
    outParameter(parameter: number): boolean {
        const rules = this.bindings.rules; const method = this.bindings.parent(parameter);
        if(method < 0 || rules.parser.node(method).kind !== 'MethodDeclaration') { return false; }
        const parameters = rules.parser.node(method).children.filter((child) => rules.parser.node(child).kind === 'Parameter');
        const position = parameters.indexOf(parameter); const own = this.nodeDeclaration(method);
        if(this.contract(own, position)) { return true; }
        const nameIndex = this.bindings.shape.name(method); if(nameIndex < 0) { return false; }
        const nameNode = rules.parser.node(nameIndex);
        if(!['Identifier', 'StringLiteral'].includes(nameNode.kind)) { return false; }
        const queue: DeclarationFact[] = [own]; const seen: string[] = [];
        for(let at = 0; at < queue.length; at++) {
            const declaration = queue[at] ?? panic('missing override container');
            const key = `${declaration.file}:${declaration.parentKind}:${declaration.parentStart}:${declaration.parentEnd}`;
            if(seen.includes(key) || !['ClassDeclaration', 'InterfaceDeclaration'].includes(declaration.parentKind)) { continue; }
            seen.push(key);
            const wire = tsgoInspect(rules.program, declaration.file, declaration.parentStart, declaration.parentEnd, declaration.parentKind, 'container-bases');
            const bases = types(wire, 'container-bases');
            for(const id of bases.roots) {
                const frames = new Frames(rules.ask(method, `property-declarations\n${id}\n${nameNode.text}`)); header(frames, 'property-declarations');
                const member = this.symbol(frames); frames.end();
                for(const inherited of member.declarations) {
                    if(!['MethodDeclaration', 'MethodSignature'].includes(inherited.kind)) { continue; }
                    if(this.contract(inherited, position)) { return true; }
                    queue.push(inherited);
                }
            }
        }
        return false;
    }
    report(index: number): void {
        const parameter = this.parameter(index);
        if(parameter < 0 || this.processState(index) || this.outParameter(parameter)) { return; }
        const rules = this.bindings.rules; const node = rules.parser.node(index);
        rules.findings.push(new Diagnostic('nexus/correctness-no-caller-data-mutation', 'callerDataMutation',
            'This writes into data the caller passed in. The change reaches the caller, and nothing at the call site says it will: the call reads like it hands data over, and the caller\'s data comes back different. Return what changed and let the caller write it, or build a new value from the parameter.',
            rules.byte(rules.start(node)), rules.byte(node.end), ''));
    }
    tags(index: number): void {
        const declaration = this.nodeDeclaration(index); const kind = declaration.kind;
        for(const tag of declaration.tags) {
            if(tag.kind !== 'JSDocUnknownTag') { continue; }
            let id = ''; let message = '';
            if(['InterfaceDeclaration', 'TypeAliasDeclaration', 'ClassDeclaration'].includes(kind) && tag.name === 'processState' && tag.reason() === '') {
                id = 'processStateWithoutReason'; message = 'This `@processState` tag gives no reason. The tag exempts every write through a parameter of this type from nexus/correctness-no-caller-data-mutation, so it has to say why the type is state the process shares rather than data a caller hands over: `@processState <why>`. Until it does, it exempts nothing.';
            }
            if(['MethodDeclaration', 'MethodSignature'].includes(kind) && tag.name === 'mutates') {
                const reason = tag.reason(); const split = reason.indexOf(' '); const parameter = split < 0 ? reason : reason.slice(0, split);
                if(declaration.parameters.indexOf(parameter) < 0) {
                    id = 'mutatesUnknownParameter'; message = 'This `@mutates` tag names no parameter of its method. It reads `@mutates <parameter> <why>`, and the first word has to be one of this method\'s parameter names, so as written it exempts nothing. A tag left behind by a rename is the usual cause.';
                }
                else if(split < 0 || reason.slice(split + 1).trim() === '') {
                    id = 'mutatesWithoutReason'; message = 'This `@mutates` tag names its parameter but gives no reason. The tag exempts every write through that parameter, in this method and in every method that overrides or implements it, from nexus/correctness-no-caller-data-mutation, so it has to say why the parameter is an out-parameter by contract: `@mutates <parameter> <why>`. Until it does, it exempts nothing.';
                }
            }
            if(id !== '') { this.bindings.rules.findings.push(new Diagnostic('nexus/correctness-no-caller-data-mutation', id, message, tag.start, tag.end, '')); }
        }
    }
    run(): void {
        const rules = this.bindings.rules;
        for(let index = 0; index < rules.parser.nodes.length; index++) {
            if(this.bindings.parent(index) < 0) { continue; }
            const node = rules.parser.node(index); let target = -1;
            if(['InterfaceDeclaration', 'TypeAliasDeclaration', 'ClassDeclaration', 'MethodDeclaration', 'MethodSignature'].includes(node.kind)) { this.tags(index); }
            if(node.kind === 'BinaryExpression' && this.writes.assignment(rules.parser.node(node.children[1] ?? -1).kind)) { target = node.children[0] ?? -1; }
            else if(node.kind === 'DeleteExpression' || (['PrefixUnaryExpression', 'PostfixUnaryExpression'].includes(node.kind) && ['PlusPlusToken', 'MinusMinusToken'].includes(node.operator))) { target = node.children[0] ?? -1; }
            if(target >= 0) {
                target = rules.skip(target);
                if(['PropertyAccessExpression', 'ElementAccessExpression'].includes(rules.parser.node(target).kind) && this.parameter(target) >= 0 && this.writesData(target)) { this.report(target); }
            }
            if(node.kind === 'CallExpression') {
                const calleeIndex = rules.skip(node.children[0] ?? -1); const callee = rules.parser.node(calleeIndex);
                if(callee.kind !== 'PropertyAccessExpression') { continue; }
                const method = callee.children[callee.children.length - 1] ?? -1;
                if(!['push', 'unshift', 'pop', 'shift', 'splice', 'sort', 'reverse', 'fill', 'copyWithin', 'set', 'delete', 'clear', 'add'].includes(rules.parser.node(method).text)) { continue; }
                const receiver = rules.skip(callee.children[0] ?? -1);
                if(this.parameter(receiver) < 0 || !this.writer(method)) { continue; }
                if(rules.parser.node(receiver).kind !== 'Identifier' && !this.writesData(receiver)) { continue; }
                this.report(calleeIndex);
            }
        }
    }
}
