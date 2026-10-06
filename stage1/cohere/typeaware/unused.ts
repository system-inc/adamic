import { panic } from 'adamic';
import { Bindings } from './bindings.ts';
import { Reassign } from './reassign.ts';
import { Frames, header } from './frames.ts';
import { Diagnostic } from './diagnostic.ts';
import { Suggestion } from './suggestion.ts';
import { Repair } from './repair.ts';

export class SymbolIdentities {
    readonly own: number;
    readonly alias: number;
    readonly unknownAlias: boolean;
    readonly local: number;
    readonly shorthand: number;
    constructor(own: number, alias: number, unknown: boolean, local: number, shorthand: number) {
        this.own = own; this.alias = alias; this.unknownAlias = unknown; this.local = local; this.shorthand = shorthand;
    }
}
export class UnusedCandidate {
    readonly name: number;
    readonly declaration: number;
    readonly kind: string;
    readonly symbol: number;
    constructor(name: number, declaration: number, kind: string, symbol: number) {
        this.name = name; this.declaration = declaration; this.kind = kind; this.symbol = symbol;
    }
}
export class Unused {
    readonly bindings: Bindings;
    readonly writes: Reassign;
    readonly identities = new Map<number, SymbolIdentities>();
    readonly candidates: UnusedCandidate[] = [];
    readonly read = new Map<number, boolean>();
    readonly typeRead = new Map<number, boolean>();
    readonly written = new Map<number, number[]>();
    readonly reported: number[] = [];
    constructor(bindings: Bindings) { this.bindings = bindings; this.writes = new Reassign(bindings.rules); }
    identity(index: number): SymbolIdentities {
        if(!this.identities.has(index)) {
            const frames = new Frames(this.bindings.rules.ask(index, 'symbol-identities')); header(frames, 'symbol-identities');
            const result = new SymbolIdentities(frames.natural(), frames.natural(), frames.yes(), frames.natural(), frames.natural());
            frames.end(); this.identities.set(index, result);
        }
        return this.identities.get(index) ?? panic('missing symbol identity');
    }
    nodeKind(index: number): string { return index < 0 ? '' : this.bindings.rules.parser.node(index).kind; }
    owner(index: number): number {
        let current = index;
        while(['BindingElement', 'ObjectBindingPattern', 'ArrayBindingPattern'].includes(this.nodeKind(current))) { current = this.bindings.parent(current); }
        return current;
    }
    scope(index: number): number {
        let current = index;
        while(current >= 0) {
            const kind = this.nodeKind(current);
            if(this.bindings.functionKind(kind) || ['SourceFile', 'ModuleDeclaration', 'ClassStaticBlockDeclaration'].includes(kind)) { return current; }
            current = this.bindings.parent(current);
        }
        return -1;
    }
    initialized(index: number): boolean {
        let current = this.bindings.parent(index);
        while(current >= 0) {
            const kind = this.nodeKind(current);
            if(['BindingElement', 'Parameter', 'VariableDeclaration'].includes(kind)) {
                if(this.bindings.shape.initializer(current) >= 0) { return true; }
                if(kind === 'Parameter') { return false; }
                if(kind === 'VariableDeclaration') {
                    return ['ForOfStatement', 'ForInStatement'].includes(this.nodeKind(this.bindings.parent(this.bindings.parent(current))));
                }
            }
            else if(!['ObjectBindingPattern', 'ArrayBindingPattern'].includes(kind)) { return false; }
            current = this.bindings.parent(current);
        }
        return false;
    }
    collect(index: number): void {
        const node = this.bindings.rules.parser.node(index); const name = this.bindings.shape.name(index);
        let kind = '';
        if(['VariableDeclaration', 'FunctionDeclaration', 'ClassDeclaration'].includes(node.kind)) { kind = 'variable'; }
        else if(node.kind === 'BindingElement') {
            const owner = this.owner(index);
            kind = this.nodeKind(owner) === 'Parameter' ? 'parameter' : this.nodeKind(this.bindings.parent(owner)) === 'CatchClause' ? 'caught' : 'variable';
        }
        else if(node.kind === 'Parameter') { kind = 'parameter'; }
        else if(['ImportEqualsDeclaration', 'ImportSpecifier', 'ImportClause', 'NamespaceImport'].includes(node.kind)) { kind = 'import'; }
        else if(['InterfaceDeclaration', 'TypeAliasDeclaration', 'EnumDeclaration', 'ModuleDeclaration', 'TypeParameter'].includes(node.kind)) { kind = 'type'; }
        if(node.kind === 'VariableDeclaration' && this.nodeKind(this.bindings.parent(index)) === 'CatchClause') { kind = 'caught'; }
        if(kind !== '' && name >= 0 && this.nodeKind(name) === 'Identifier') {
            this.candidates.push(new UnusedCandidate(name, index, kind, this.identity(name).own));
        }
        for(const child of node.children) { this.collect(child); }
    }
    loop(index: number): boolean {
        let current = index;
        while(current >= 0) {
            const kind = this.nodeKind(current);
            if(['ForStatement', 'ForInStatement', 'ForOfStatement', 'WhileStatement', 'DoStatement'].includes(kind)) { return true; }
            if(['FunctionDeclaration', 'FunctionExpression', 'ArrowFunction', 'MethodDeclaration', 'SourceFile'].includes(kind)) { return false; }
            current = this.bindings.parent(current);
        }
        return false;
    }
    discarded(index: number, name: string): boolean {
        let current = index;
        while(current >= 0) {
            const parent = this.bindings.parent(current); if(parent < 0) { return false; }
            const node = this.bindings.rules.parser.node(parent);
            if(['ParenthesizedExpression', 'AsExpression', 'NonNullExpression', 'SatisfiesExpression'].includes(node.kind)) { current = parent; continue; }
            if(node.kind === 'ExpressionStatement') { return true; }
            if(node.kind === 'BinaryExpression') {
                const operator = this.nodeKind(node.children[1] ?? -1);
                if(operator === 'CommaToken') {
                    if(node.children[0] === current) { return true; }
                    current = parent; continue;
                }
                if(this.writes.assignment(operator) && node.children[2] === current) {
                    const target = this.bindings.rules.parser.node(this.bindings.rules.skip(node.children[0] ?? -1));
                    if(target.kind === 'Identifier' && target.text === name) { current = parent; continue; }
                }
            }
            if(node.kind === 'ForStatement') {
                return node.children[0] === current || node.children[node.children.length - 2] === current;
            }
            return false;
        }
        return false;
    }
    update(index: number): boolean {
        let current = index;
        while(current >= 0) {
            const parent = this.bindings.parent(current); if(parent < 0) { return false; }
            const node = this.bindings.rules.parser.node(parent);
            if(node.kind === 'ParenthesizedExpression') { current = parent; continue; }
            if(['PrefixUnaryExpression', 'PostfixUnaryExpression'].includes(node.kind)) { return ['PlusPlusToken', 'MinusMinusToken'].includes(node.operator); }
            return node.kind === 'BinaryExpression' && node.children[0] === current && this.nodeKind(node.children[1] ?? -1) !== 'EqualsToken' &&
                this.writes.assignment(this.nodeKind(node.children[1] ?? -1));
        }
        return false;
    }
    selfAssignment(index: number, scope: number): boolean {
        const name = this.bindings.rules.parser.node(index).text;
        let current = index;
        while(current >= 0) {
            const parent = this.bindings.parent(current); if(parent < 0) { return false; }
            const node = this.bindings.rules.parser.node(parent);
            if(['ParenthesizedExpression', 'PropertyAccessExpression', 'ElementAccessExpression', 'NonNullExpression', 'AsExpression', 'CallExpression'].includes(node.kind)) {
                if(node.kind === 'CallExpression' && node.children[0] !== current) { return false; }
                current = parent; continue;
            }
            if(['PrefixUnaryExpression', 'PostfixUnaryExpression'].includes(node.kind)) {
                return current === index && ['PlusPlusToken', 'MinusMinusToken'].includes(node.operator) && this.discarded(parent, name) && !this.loop(parent);
            }
            if(node.kind !== 'BinaryExpression') { return false; }
            const operator = this.nodeKind(node.children[1] ?? -1);
            if(!this.writes.assignment(operator)) { current = parent; continue; }
            const target = this.bindings.rules.parser.node(this.bindings.rules.skip(node.children[0] ?? -1));
            return target.kind === 'Identifier' && target.text === name && this.scope(index) === scope && this.discarded(parent, name) && !this.loop(parent);
        }
        return false;
    }
    excludedRead(index: number, symbol: number): boolean {
        const name = this.bindings.rules.parser.node(index).text;
        let current = this.bindings.parent(index);
        let sequence = index;
        while(sequence >= 0) {
            const parent = this.bindings.parent(sequence); if(parent < 0) { break; }
            const node = this.bindings.rules.parser.node(parent);
            if(node.kind === 'ParenthesizedExpression') { sequence = parent; continue; }
            if(node.kind !== 'BinaryExpression' || this.nodeKind(node.children[1] ?? -1) !== 'CommaToken') { break; }
            if(node.children[0] === sequence) { return true; }
            sequence = parent;
        }
        let closure = true;
        while(current >= 0) {
            const node = this.bindings.rules.parser.node(current);
            if(['TypeAliasDeclaration', 'InterfaceDeclaration', 'FunctionDeclaration', 'FunctionExpression', 'ClassDeclaration', 'ClassExpression', 'ArrowFunction'].includes(node.kind)) {
                const declared = node.kind === 'ArrowFunction' ? -1 : this.bindings.shape.name(current);
                if(declared >= 0 && this.nodeKind(declared) === 'Identifier' && this.identity(declared).own === symbol) { return true; }
                const parent = this.bindings.parent(current);
                if(this.nodeKind(parent) === 'VariableDeclaration') {
                    const variable = this.bindings.shape.name(parent);
                    if(variable >= 0 && this.nodeKind(variable) === 'Identifier' && this.identity(variable).own === symbol) { return true; }
                }
            }
            if(closure) {
                if(node.kind === 'Parameter' || ['FunctionDeclaration', 'MethodDeclaration', 'SourceFile'].includes(node.kind)) { closure = false; }
                else if(['ArrowFunction', 'FunctionExpression'].includes(node.kind)) {
                    const parent = this.bindings.parent(current); const assignment = parent >= 0 ? this.bindings.rules.parser.node(parent) : node;
                    if(assignment.kind === 'BinaryExpression' && this.nodeKind(assignment.children[1] ?? -1) === 'EqualsToken' && assignment.children[2] === current) {
                        const target = this.bindings.rules.parser.node(this.bindings.rules.skip(assignment.children[0] ?? -1));
                        if(target.kind === 'Identifier' && target.text === name) { return true; }
                    }
                    closure = false;
                }
            }
            current = this.bindings.parent(current);
        }
        const candidate = this.candidates.find((candidate) => candidate.symbol === symbol);
        return this.selfAssignment(index, candidate === undefined ? -1 : this.scope(candidate.declaration));
    }
    typeQuery(index: number): boolean {
        let current = this.bindings.parent(index);
        while(current >= 0 && this.nodeKind(current) === 'QualifiedName') { current = this.bindings.parent(current); }
        return this.nodeKind(current) === 'TypeQuery';
    }
    gather(index: number, interesting: readonly string[]): void {
        const node = this.bindings.rules.parser.node(index);
        if(node.kind === 'Identifier' && interesting.includes(node.text)) {
            const ids = this.identity(index); const symbols: number[] = [];
            if(ids.shorthand !== 0) { symbols.push(ids.shorthand); }
            else {
                if(ids.own !== 0) { symbols.push(ids.own); }
                if(this.nodeKind(this.bindings.parent(index)) === 'ExportSpecifier') {
                    if(ids.alias !== 0 && !ids.unknownAlias) { symbols.push(ids.alias); }
                    if(ids.local !== 0 && !symbols.includes(ids.local)) { symbols.push(ids.local); }
                }
            }
            const declaring = this.bindings.declaring(index);
            const write = !declaring && this.writes.writes(index);
            for(const symbol of symbols) {
                if(this.nodeKind(this.bindings.parent(index)) === 'TypePredicate') { this.typeRead.set(symbol, true); }
                if(write) {
                    if(!this.written.has(symbol)) { this.written.set(symbol, []); }
                    const writes = this.written.get(symbol) ?? panic('missing writes'); writes.push(index);
                }
                if(declaring || (write && !this.update(index)) || this.excludedRead(index, symbol)) { continue; }
                if(this.typeQuery(index)) { this.typeRead.set(symbol, true); }
                else { this.read.set(symbol, true); }
            }
        }
        for(const child of node.children) { this.gather(child, interesting); }
    }
    exported(index: number): boolean {
        let current = index;
        while(current >= 0) {
            const node = this.bindings.rules.parser.node(current);
            if(['Parameter', 'TypeParameter'].includes(node.kind)) { return false; }
            if(['FunctionExpression', 'ArrowFunction', 'MethodDeclaration', 'Constructor', 'GetAccessor', 'SetAccessor', 'Block', 'ModuleBlock'].includes(node.kind) && current !== index) { return false; }
            if(node.kind === 'ImportEqualsDeclaration') { return node.children.some((child) => ['ExportKeyword', 'DefaultKeyword'].includes(this.nodeKind(child))); }
            if(['VariableStatement', 'FunctionDeclaration', 'ClassDeclaration', 'InterfaceDeclaration', 'TypeAliasDeclaration', 'EnumDeclaration', 'ModuleDeclaration'].includes(node.kind) &&
                node.children.some((child) => ['ExportKeyword', 'DefaultKeyword'].includes(this.nodeKind(child)))) { return true; }
            current = this.bindings.parent(current);
        }
        return false;
    }
    signature(index: number): boolean {
        let current = index;
        while(current >= 0) {
            const node = this.bindings.rules.parser.node(current);
            if(node.children.some((child) => this.nodeKind(child) === 'DeclareKeyword')) { return this.nodeKind(index) !== 'TypeParameter'; }
            if(['MethodSignature', 'CallSignature', 'ConstructSignature', 'IndexSignature', 'FunctionType', 'ConstructorType'].includes(node.kind)) { return true; }
            if(['FunctionDeclaration', 'MethodDeclaration', 'Constructor'].includes(node.kind)) {
                return current !== index && !this.bindings.shape.body(current);
            }
            if(node.kind === 'SourceFile') { return false; }
            current = this.bindings.parent(current);
        }
        return false;
    }
    required(parameter: number): boolean {
        const owner = this.bindings.parent(parameter);
        if(this.nodeKind(owner) === 'SetAccessor') { return true; }
        return this.nodeKind(owner) === 'Constructor' && this.bindings.rules.parser.node(parameter).children.some((child) =>
            ['PublicKeyword', 'PrivateKeyword', 'ProtectedKeyword', 'ReadonlyKeyword', 'OverrideKeyword'].includes(this.nodeKind(child)));
    }
    defaulted(declaration: number): boolean {
        let current = declaration;
        while(current >= 0 && ['BindingElement', 'Parameter', 'ObjectBindingPattern', 'ArrayBindingPattern'].includes(this.nodeKind(current))) {
            if(this.bindings.shape.initializer(current) >= 0) { return true; }
            if(this.nodeKind(current) === 'Parameter') { return false; }
            current = this.bindings.parent(current);
        }
        return false;
    }
    leadingReturn(index: number): boolean {
        let current = index;
        while(current >= 0) {
            const node = this.bindings.rules.parser.node(current);
            if(['ForInStatement', 'ForOfStatement'].includes(node.kind)) {
                let body = node.children[node.children.length - 1] ?? -1;
                if(this.nodeKind(body) === 'Block') { body = this.bindings.rules.parser.node(body).children[0] ?? -1; }
                return this.nodeKind(body) === 'ReturnStatement';
            }
            if(['FunctionDeclaration', 'FunctionExpression', 'ArrowFunction', 'SourceFile'].includes(node.kind)) { return false; }
            current = this.bindings.parent(current);
        }
        return false;
    }
    exempt(candidate: UnusedCandidate): boolean {
        const node = this.bindings.rules.parser.node(candidate.declaration);
        const name = this.bindings.rules.parser.node(candidate.name).text;
        if(candidate.symbol === 0 || this.read.get(candidate.symbol) === true || this.signature(candidate.declaration)) { return true; }
        if(candidate.kind === 'import') {
            const ids = this.identity(candidate.name);
            if(node.kind !== 'ImportEqualsDeclaration' && !ids.unknownAlias && this.read.get(ids.alias) === true) { return true; }
            if(['React', 'h'].includes(name) && (this.bindings.rules.path.endsWith('.tsx') || this.bindings.rules.path.endsWith('.jsx'))) { return true; }
        }
        if(['ClassDeclaration', 'Parameter'].includes(node.kind) && node.children.some((child) => this.nodeKind(child) === 'Decorator')) { return true; }
        if(candidate.kind === 'parameter') {
            if(name === 'this' || this.required(this.owner(candidate.declaration))) { return true; }
            if(node.kind === 'Parameter') {
                const functionIndex = this.bindings.parent(candidate.declaration);
                let seen = false;
                for(const other of this.candidates) {
                    if(other.declaration === candidate.declaration) { seen = true; continue; }
                    if(!seen || other.kind !== 'parameter' || this.bindings.parent(this.owner(other.declaration)) !== functionIndex) { continue; }
                    const parameter = this.owner(other.declaration);
                    if(this.required(parameter) || this.bindings.rules.parser.node(parameter).children.some((child) => this.nodeKind(child) === 'Decorator') ||
                        this.defaulted(other.declaration) || this.read.get(other.symbol) === true) { return true; }
                }
            }
        }
        if(candidate.kind === 'type' && this.nodeKind(this.bindings.parent(candidate.declaration)) === 'MappedType') { return true; }
        if(node.kind === 'ModuleDeclaration' && node.children.some((child) => this.nodeKind(child) === 'ModuleDeclaration')) { return true; }
        if(this.leadingReturn(candidate.declaration)) { return true; }
        const writes = this.written.get(candidate.symbol);
        if(writes !== undefined && writes.some((write) => this.leadingReturn(write))) { return true; }
        return this.bindings.list(candidate.name, 'binding-declarations', false).some((declaration) => this.exported(declaration)) || this.exported(candidate.declaration);
    }
    importRemoval(candidate: UnusedCandidate, finding: Diagnostic): void {
        if(this.bindings.list(candidate.name, 'binding-declarations', false).length > 1) { return; }
        const rules = this.bindings.rules; const node = rules.parser.node(candidate.declaration);
        let statement = candidate.declaration;
        while(statement >= 0 && !['ImportDeclaration', 'ImportEqualsDeclaration'].includes(this.nodeKind(statement))) { statement = this.bindings.parent(statement); }
        if(statement < 0) { return; }
        const peers = this.candidates.filter((other) => other.kind === 'import' && this.bindings.contains(statement, other.declaration));
        const all = peers.every((other) => this.reported.includes(other.name));
        let start = -1; let end = -1; let whole = false;
        if(all || peers.length === 1 || node.kind === 'ImportEqualsDeclaration') {
            whole = true; const declaration = rules.parser.node(statement);
            start = rules.start(declaration); end = declaration.end;
            const lineStart = rules.scanner.text.slice(0, start).lastIndexOf('\n') + 1;
            const after = rules.scanner.text.indexOf('\n', end); const lineEnd = after < 0 ? rules.scanner.text.length : after + 1;
            if(rules.scanner.text.slice(start, end) === rules.scanner.text.slice(lineStart, lineEnd).trim()) { start = lineStart; end = lineEnd; }
        }
        else if(node.kind === 'NamespaceImport') { return; }
        else {
            const declaration = rules.parser.node(statement);
            const scanner = rules.scanner; scanner.pos = rules.start(declaration);
            const kinds: string[] = []; const starts: number[] = []; const ends: number[] = [];
            while(scanner.pos < declaration.end) {
                const kind = scanner.scan(); if(kind === 'EndOfFileToken') { break; }
                kinds.push(kind); starts.push(scanner.start); ends.push(scanner.pos);
            }
            if(node.kind === 'ImportClause') {
                start = rules.start(rules.parser.node(candidate.name)); end = rules.parser.node(candidate.name).end;
                const comma = starts.findIndex((value) => value >= end);
                if(kinds[comma] !== 'CommaToken') { return; } end = ends[comma] ?? end;
            }
            else if(node.kind === 'ImportSpecifier') {
                const unusedNamed = peers.filter((other) => this.nodeKind(other.declaration) === 'ImportSpecifier' && !this.reported.includes(other.name));
                if(unusedNamed.length === 0) {
                    const curly = kinds.indexOf('OpenBraceToken'); const right = kinds.indexOf('CloseBraceToken');
                    if(curly < 1 || kinds[curly - 1] !== 'CommaToken' || right < 0) { return; }
                    start = starts[curly - 1] ?? -1; end = ends[right] ?? -1;
                }
                else {
                    start = rules.start(node); end = node.end;
                    let before = -1; let after = -1;
                    for(let at = 0; at < kinds.length; at++) {
                        if((ends[at] ?? -1) <= start) { before = at; }
                        if(after < 0 && (starts[at] ?? -1) >= end) { after = at; }
                    }
                    const comma = kinds[before] === 'CommaToken' ? before : after;
                    if(kinds[comma] !== 'CommaToken') { return; }
                    start = Math.min(start, starts[comma] ?? start); end = Math.max(end, ends[comma] ?? end);
                }
            }
        }
        if(start < 0 || end < 0) { return; }
        const name = rules.parser.node(candidate.name).text;
        const id = whole ? 'removeUnusedImportDeclaration' : 'removeUnusedVar';
        const message = whole ? 'Remove the import declaration: nothing it binds is used.' : `Remove '${name}' from the import, keeping the bindings beside it that are used.`;
        finding.suggestions.push(new Suggestion(id, message, [new Repair(rules.byte(start), rules.byte(end), '')]));
    }
    run(): void {
        const rules = this.bindings.rules;
        if(['.d.ts', '.vue', '.svelte', '.astro'].some((suffix) => rules.path.endsWith(suffix))) { return; }
        const root = rules.parser.nodes.findIndex((node) => node.kind === 'SourceFile'); this.collect(root);
        this.gather(root, this.candidates.map((candidate) => rules.parser.node(candidate.name).text));
        for(const candidate of this.candidates) {
            if(this.exempt(candidate)) { continue; }
            const typeOnly = this.typeRead.get(candidate.symbol) === true;
            if(typeOnly && candidate.kind === 'import') { continue; }
            let index = candidate.name; let assigned = this.initialized(candidate.name);
            const writes = this.written.get(candidate.symbol);
            if(writes !== undefined) {
                if(writes.length > 0) { assigned = true; }
                for(const write of writes) {
                    if(this.scope(write) === this.scope(candidate.declaration) && rules.parser.node(write).pos > rules.parser.node(index).pos) { index = write; }
                }
            }
            const name = rules.parser.node(candidate.name).text; const action = assigned ? 'assigned a value' : 'declared';
            const message = typeOnly ? `'${name}' is ${action} but only used as a type. Every reference is a \`typeof\`, which reads the binding's type while compiling and nothing at runtime, so the value itself is built for no reader. Write the type out and delete the value, or keep the value if something outside this file is meant to read it and export it.` :
                `'${name}' is ${action} and nothing ever reads it. A name that is written but never read is almost always the residue of an edit that moved on: an import whose call site was deleted, a parameter left behind when a signature changed, a variable holding a value nobody asked for. It costs nothing at runtime, which is what lets it accumulate, and it costs the next reader real time, because an unused name reads exactly like a used one until you search the file and find nothing. Delete it.`;
            const node = rules.parser.node(index); const finding = new Diagnostic('no-unused-vars', typeOnly ? 'usedOnlyAsType' : 'noUnusedVars', message, rules.byte(rules.start(node)), rules.byte(node.end));
            this.reported.push(candidate.name); if(candidate.kind === 'import' && !typeOnly) { this.importRemoval(candidate, finding); }
            rules.findings.push(finding);
        }
    }
}
