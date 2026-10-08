// Shared checkpoint framing and checked identity arenas. Pass payload codecs live with their pass.
import { panic } from 'adamic';
import { FunctionIndex, BlockIndex, InstructionIndex, IdentifierIndex, DeclarationIndex, ScopeIndex, ReactiveIndex } from '../../arena/arena_index.a';
export { FunctionIndex, BlockIndex, InstructionIndex, IdentifierIndex, DeclarationIndex, ScopeIndex, ReactiveIndex } from '../../arena/arena_index.a';
export type AnchorKind = 'function' | 'identifier' | 'declaration' | 'block' | 'instruction' | 'scope' | 'reactive';
export interface GraphRow { readonly kind: string; readonly body: string | undefined; readonly functionPath: string; }
export interface ReplayInstruction {
    readonly index: InstructionIndex; readonly id: number; readonly order: number;
    readonly lvalue: string; readonly range: string; readonly kind: string;
    payload: string; readonly orphan: boolean;
}
export interface SidecarRow {
    readonly namespace: string; readonly functionPath: string; readonly anchorKind: AnchorKind;
    readonly anchorId: number | undefined; readonly key: string; readonly payload: string;
}
export interface ExtraIdentity { readonly functionPath: string; readonly kind: 'scope' | 'reactive' | 'block'; readonly id: number; }
export function integer(text: string): number {
    const value = Number(text);
    if(text === '' || !Number.isInteger(value) || value < 0 || value > 9007199254740991 || String(value) !== text) { panic(`invalid checkpoint integer: ${text}`); }
    return value;
}
function word(words: readonly string[], at: number): string { return words[at] ?? panic('truncated checkpoint row'); }
// Scalar order agrees with Go's UTF-8 byte order, including supplementary keys.
function compare(a: string,b: string): number {
 let x = 0; let y = 0;
 while(x < a.length && y < b.length) {
  let ac = a.charCodeAt(x); x++; let bc = b.charCodeAt(y); y++;
  if(ac >= 55296 && ac <= 56319 && x < a.length) { const low = a.charCodeAt(x); if(low >= 56320 && low <= 57343) { ac = 65536 + (ac - 55296) * 1024 + low - 56320; x++; } }
  if(bc >= 55296 && bc <= 56319 && y < b.length) { const low = b.charCodeAt(y); if(low >= 56320 && low <= 57343) { bc = 65536 + (bc - 55296) * 1024 + low - 56320; y++; } }
  if(ac !== bc) { return ac < bc ? -1 : 1; }
 }
 return x < a.length ? 1 : y < b.length ? -1 : 0;
}
// Escaped framing fields remain one physical line. Payload strings are codec-owned.
export function escapeField(value: string): string {
    let result = '';
    for(let i = 0; i < value.length; i++) {
        const code = value.charCodeAt(i);
        if(code === 37 || code === 9 || code === 10 || code === 13 || code >= 128) { result += '%' + code.toString(16).padStart(4,'0'); }
        else { result += value.charAt(i); }
    }
    return result;
}
export function unescapeField(value: string): string {
    let result = '';
    for(let i = 0; i < value.length; i++) {
        if(value.charAt(i) !== '%') { result += value.charAt(i); continue; }
        const digits = value.slice(i + 1,i + 5); const code = Number.parseInt(digits,16);
        if(digits.length !== 4 || (code !== 37 && code !== 9 && code !== 10 && code !== 13 && code < 128) || code.toString(16).padStart(4,'0') !== digits) { panic('noncanonical checkpoint escape'); }
        result += String.fromCharCode(code); i += 4;
    }
    return result;
}
export class ReplayFunction {
    readonly path: string;
    readonly identifiers: IdentifierIndex[] = []; readonly identifierIds: number[] = []; readonly identifierMap = new Map<number, IdentifierIndex>();
    readonly declarations: DeclarationIndex[] = []; readonly declarationIds: number[] = []; readonly declarationMap = new Map<number, DeclarationIndex>();
    readonly blocks: BlockIndex[] = []; readonly blockIds: number[] = []; readonly blockMap = new Map<number, BlockIndex>();
    readonly instructions: InstructionIndex[] = []; readonly instructionRows: ReplayInstruction[] = []; readonly instructionMap = new Map<number, InstructionIndex>();
    readonly scopes: ScopeIndex[] = []; readonly scopeIds: number[] = []; readonly scopeMap = new Map<number, ScopeIndex>();
    readonly reactive: ReactiveIndex[] = []; readonly reactiveIds: number[] = []; readonly reactiveMap = new Map<number, ReactiveIndex>();
    readonly children: FunctionIndex[] = [];
    readonly places: string[] = []; readonly blockReferences: number[] = [];
    entry = 0;
    constructor(path: string) { this.path = path; }
    identifier(id: number): IdentifierIndex { const index = this.identifierMap.get(id) ?? panic(`unknown identifier ${this.path}:${id}`); IdentifierIndex.read(this.identifiers,index); return index; }
    declaration(id: number): DeclarationIndex { const index = this.declarationMap.get(id) ?? panic(`unknown declaration ${this.path}:${id}`); DeclarationIndex.read(this.declarations,index); return index; }
    block(id: number): BlockIndex { const index = this.blockMap.get(id) ?? panic(`unknown block ${this.path}:${id}`); BlockIndex.read(this.blocks,index); return index; }
    instruction(id: number): InstructionIndex { const index = this.instructionMap.get(id) ?? panic(`unknown instruction ${this.path}:${id}`); InstructionIndex.read(this.instructions,index); return index; }
    instructionRow(index: InstructionIndex): ReplayInstruction { return this.instructionRows[InstructionIndex.read(this.instructions,index)] ?? panic('missing instruction row'); }
    scope(id: number): ScopeIndex { const index = this.scopeMap.get(id) ?? panic(`unknown scope ${this.path}:${id}`); ScopeIndex.read(this.scopes,index); return index; }
    reactiveNode(id: number): ReactiveIndex { const index = this.reactiveMap.get(id) ?? panic(`unknown reactive ${this.path}:${id}`); ReactiveIndex.read(this.reactive,index); return index; }
    validate(): void {
        this.block(this.entry);
        for(const id of this.blockReferences) { this.block(id); }
        for(const place of this.places) { this.identifier(integer(word(place.split(':'),0))); }
    }
}
export class IdentityGraph {
    readonly indices: FunctionIndex[] = []; readonly functions: ReplayFunction[] = [];
    readonly paths = new Map<string,FunctionIndex>(); readonly rows: GraphRow[] = [];
    functionIndex(path: string): FunctionIndex { const index = this.paths.get(path) ?? panic(`unknown function path ${path}`); FunctionIndex.read(this.indices,index); return index; }
    function(index: FunctionIndex): ReplayFunction { return this.functions[FunctionIndex.read(this.indices,index)] ?? panic('missing function identity'); }
    at(path: string): ReplayFunction { return this.function(this.functionIndex(path)); }
    addFunction(path: string): FunctionIndex {
        if(this.paths.has(path)) { panic(`duplicate function path ${path}`); }
        const index = FunctionIndex.push(this.indices); this.paths.set(path,index); this.functions.push(new ReplayFunction(path)); return index;
    }
    registerIdentity(identity: ExtraIdentity): void {
        const fn = this.at(identity.functionPath);
        integer(String(identity.id));
        if(identity.kind === 'block') { if(fn.blockMap.has(identity.id)) { panic('duplicate block identity'); } const index = BlockIndex.push(fn.blocks); fn.blockMap.set(identity.id,index); fn.blockIds.push(identity.id); }
        else if(identity.kind === 'scope') {
            if(fn.scopeMap.has(identity.id)) { panic('duplicate scope identity'); }
            const index = ScopeIndex.push(fn.scopes); fn.scopeMap.set(identity.id,index); fn.scopeIds.push(identity.id);
        } else {
            if(fn.reactiveMap.has(identity.id)) { panic('duplicate reactive identity'); }
            const index = ReactiveIndex.push(fn.reactive); fn.reactiveMap.set(identity.id,index); fn.reactiveIds.push(identity.id);
        }
    }
    validateAnchor(row: SidecarRow): void {
        const fn = this.at(row.functionPath);
        if(row.anchorKind === 'function') { if(row.anchorId !== undefined) { panic('function anchor has an id'); } return; }
        const id = row.anchorId ?? panic('non-function anchor requires an id'); integer(String(id));
        if(row.anchorKind === 'identifier') { fn.identifier(id); }
        else if(row.anchorKind === 'declaration') { fn.declaration(id); }
        else if(row.anchorKind === 'block') { fn.block(id); }
        else if(row.anchorKind === 'instruction') { fn.instruction(id); }
        else if(row.anchorKind === 'scope') { fn.scope(id); }
        else { fn.reactiveNode(id); }
    }
}
// This decodes graph rows and identities, not pass-specific InstructionValue payloads.
export function readGraph(text: string): IdentityGraph {
    if(!text.endsWith('\n')) { panic('graph must end with a newline'); }
    const graph = new IdentityGraph(); const stack: FunctionIndex[] = [];
    let pending: number | undefined; let needFunction = false;
    const lines = text.slice(0,-1).split('\n');
    for(const line of lines) {
        const space = line.indexOf(' '); const kind = space < 0 ? line : line.slice(0,space); const body = space < 0 ? undefined : line.slice(space + 1);
        if(kind === 'hir-v1') {
            if(body !== undefined || needFunction || (stack.length > 0 && pending === undefined) || (stack.length === 0 && graph.functions.length > 0)) { panic('unexpected graph header'); }
            needFunction = true;
        } else if(kind === 'function') {
            if(!needFunction) { panic('function without graph header'); }
            let path = '$'; const parent = stack[stack.length - 1];
            if(parent !== undefined) { path = graph.function(parent).path + '/' + String(pending ?? panic('nested function without ordinal')); }
            const index = graph.addFunction(path);
            if(parent !== undefined) { graph.function(parent).children.push(index); }
            stack.push(index); pending = undefined; needFunction = false;
            const words = (body ?? panic('missing function body')).split(' ');
            const entry = word(words,2); if(!entry.startsWith('entry=')) { panic('missing entry'); }
            graph.function(index).entry = integer(entry.slice(6));
        } else {
            if(needFunction || pending !== undefined) { panic('missing nested function header'); }
            const index = stack[stack.length - 1] ?? panic('graph row outside a function'); const fn = graph.function(index); const fields = (body ?? '').split(' ');
            if(kind === 'identifier') {
                const id = integer(word(fields,0)); const decl = integer(word(fields,1));
                if(fn.identifierMap.has(id)) { panic('duplicate identifier'); }
                const handle = IdentifierIndex.push(fn.identifiers); fn.identifierMap.set(id,handle); fn.identifierIds.push(id);
                if(!fn.declarationMap.has(decl)) { const d = DeclarationIndex.push(fn.declarations); fn.declarationMap.set(decl,d); fn.declarationIds.push(decl); }
            } else if(kind === 'block') {
                const id = integer(word(fields,0)); if(fn.blockMap.has(id)) { panic('duplicate block'); }
                const handle = BlockIndex.push(fn.blocks); fn.blockMap.set(id,handle); fn.blockIds.push(id);
                const preds = word(fields,2); if(!preds.startsWith('predecessors=')) { panic('missing predecessors'); }
                for(const pred of preds.slice(13).split(',')) { if(pred !== '') { fn.blockReferences.push(integer(pred)); } }
            } else if(kind === 'instruction' || kind === 'orphan') {
                const raw = kind === 'orphan' ? (body ?? panic('missing orphan')).slice(12) : body ?? panic('missing instruction');
                if(kind === 'orphan' && !(body ?? '').startsWith('instruction ')) { panic('invalid orphan'); }
                const tokens = raw.split(' '); const id = integer(word(tokens,0));
                if(fn.instructionMap.has(id)) { panic('duplicate instruction'); }
                const handle = InstructionIndex.push(fn.instructions); fn.instructionMap.set(id,handle);
                const lvalue = word(tokens,2); fn.places.push(lvalue);
                fn.instructionRows.push({index: handle,id,order: integer(word(tokens,1)),lvalue,range: word(tokens,3),kind: word(tokens,4),payload: tokens.slice(5).join(' '),orphan: kind === 'orphan'});
            } else if(kind === 'params' || kind === 'context') { if(body !== '') { for(const place of (body ?? panic('missing places')).split(',')) { fn.places.push(place); } } }
            else if(kind === 'returns') { fn.places.push(body ?? panic('missing return place')); }
            else if(kind === 'phi') { fn.places.push(word(fields,0)); for(const operand of fields.slice(1)) { const parts = operand.split('='); fn.blockReferences.push(integer(word(parts,0))); fn.places.push(word(parts,1)); } }
            else if(kind === 'terminal') { if(word(fields,1) === 'Return') { fn.places.push(word(fields,2)); } }
            else if(kind === 'nested') { pending = integer(body ?? ''); if(pending !== fn.children.length) { panic('noncanonical nested ordinal'); } }
            else if(kind === 'end') { fn.validate(); stack.pop(); }
            else if(kind !== 'scopes' && kind !== 'context-declarations' && kind !== 'outlined') { panic(`unknown graph row ${kind}`); }
        }
        const current = stack[stack.length - 1];
        graph.rows.push({kind,body,functionPath: current === undefined ? '$' : graph.function(current).path});
    }
    if(stack.length !== 0 || pending !== undefined || needFunction || graph.functions.length === 0) { panic('incomplete graph'); }
    return graph;
}
export function writeGraph(graph: IdentityGraph): string {
    let result = '';
    for(const row of graph.rows) {
        if(row.kind === 'instruction' || row.kind === 'orphan') {
            const raw = row.kind === 'orphan' ? (row.body ?? '').slice(12) : row.body ?? '';
            const fn = graph.at(row.functionPath); const instruction = fn.instructionRow(fn.instruction(integer(word(raw.split(' '),0))));
            result += `${instruction.orphan ? 'orphan ' : ''}instruction ${instruction.id} ${instruction.order} ${instruction.lvalue} ${instruction.range} ${instruction.kind} ${instruction.payload}\n`;
        } else { result += row.kind + (row.body === undefined ? '' : ' ' + row.body) + '\n'; }
    }
    return result;
}
function sidecarOrder(a: SidecarRow,b: SidecarRow): number {
    let order = compare(a.namespace,b.namespace); if(order !== 0) { return order; }
    order = compare(a.functionPath,b.functionPath); if(order !== 0) { return order; }
    order = compare(a.anchorKind,b.anchorKind); if(order !== 0) { return order; }
    order = (a.anchorId ?? -1) - (b.anchorId ?? -1); if(order !== 0) { return order; }
    return compare(a.key,b.key);
}
function identityOrder(a: ExtraIdentity,b: ExtraIdentity): number { let order = compare(a.functionPath,b.functionPath); if(order !== 0) { return order; } order = compare(a.kind,b.kind); if(order !== 0) { return order; } return a.id - b.id; }
function kind(text: string): AnchorKind {
    if(text === 'function' || text === 'identifier' || text === 'declaration' || text === 'block' || text === 'instruction' || text === 'scope' || text === 'reactive') { return text; }
    return panic('unknown anchor kind');
}
export class Checkpoint {
    readonly key: string; readonly pass: string; readonly graph: IdentityGraph;
    readonly identities: ExtraIdentity[] = []; readonly sidecars: SidecarRow[] = [];
    private readonly sidecarKeys = new Set<string>();
    constructor(key: string,pass: string,graph: IdentityGraph) { this.key = key; this.pass = pass; this.graph = graph; }
    addIdentity(identity: ExtraIdentity): void { this.graph.registerIdentity(identity); this.identities.push(identity); }
    addSidecar(row: SidecarRow): void {
        if(row.namespace === '' || row.key === '') { panic('empty sidecar namespace or key'); }
        this.graph.validateAnchor(row);
        const key = escapeField(row.namespace) + '\t' + escapeField(row.functionPath) + '\t' + row.anchorKind + '\t' + String(row.anchorId ?? -1) + '\t' + escapeField(row.key);
        if(this.sidecarKeys.has(key)) { panic('duplicate sidecar key'); }
        this.sidecarKeys.add(key);
        this.sidecars.push(row);
    }
    select(namespace: string,path: string): SidecarRow[] { this.graph.at(path); return this.sidecars.filter((row) => row.namespace === namespace && row.functionPath === path); }
}
export function writeCheckpoint(checkpoint: Checkpoint): string {
    const rows = writeGraph(checkpoint.graph).slice(0,-1).split('\n');
    let result = `hir-checkpoint-v1\ncase\t${escapeField(checkpoint.key)}\t${escapeField(checkpoint.pass)}\ngraph-lines ${rows.length}\n` + rows.join('\n') + '\n';
    const identities = checkpoint.identities.slice().sort(identityOrder);
    result += `identities ${identities.length}\n`;
    for(const identity of identities) { result += `identity\t${escapeField(identity.functionPath)}\t${identity.kind}\t${identity.id}\n`; }
    const sidecars = checkpoint.sidecars.slice().sort(sidecarOrder); result += `sidecars ${sidecars.length}\n`;
    for(const row of sidecars) { checkpoint.graph.validateAnchor(row); result += `sidecar\t${escapeField(row.namespace)}\t${escapeField(row.functionPath)}\t${row.anchorKind}\t${row.anchorId === undefined ? '-' : row.anchorId}\t${escapeField(row.key)}\t${escapeField(row.payload)}\n`; }
    return result + 'end-checkpoint\n';
}
export function readCheckpoint(text: string): Checkpoint {
    const lines = text.split('\n'); if(lines[0] !== 'hir-checkpoint-v1') { panic('unknown checkpoint version'); }
    const header = word(lines,1).split('\t'); if(header.length !== 3 || header[0] !== 'case') { panic('invalid case row'); }
    const graphHeader = word(lines,2); if(!graphHeader.startsWith('graph-lines ')) { panic('missing graph length'); }
    const count = integer(graphHeader.slice(12)); const graph = readGraph(lines.slice(3,3 + count).join('\n') + '\n');
    const checkpoint = new Checkpoint(unescapeField(word(header,1)),unescapeField(word(header,2)),graph); let at = 3 + count;
    const identityHeader = word(lines,at); at++; if(!identityHeader.startsWith('identities ')) { panic('missing identities'); }
    const identityCount = integer(identityHeader.slice(11));
    for(let i = 0; i < identityCount; i++) {
        const fields = word(lines,at).split('\t'); at++; const tag = word(fields,2);
        if(fields.length !== 4 || fields[0] !== 'identity' || (tag !== 'scope' && tag !== 'reactive' && tag !== 'block')) { panic('invalid identity row'); }
        checkpoint.addIdentity({functionPath: unescapeField(word(fields,1)),kind: tag,id: integer(word(fields,3))});
    }
    const sidecarHeader = word(lines,at); at++; if(!sidecarHeader.startsWith('sidecars ')) { panic('missing sidecars'); }
    const sidecarCount = integer(sidecarHeader.slice(9));
    for(let i = 0; i < sidecarCount; i++) {
        const fields = word(lines,at).split('\t'); at++; if(fields.length !== 7 || fields[0] !== 'sidecar') { panic('invalid sidecar row'); }
        const id = word(fields,4);
        checkpoint.addSidecar({namespace: unescapeField(word(fields,1)),functionPath: unescapeField(word(fields,2)),anchorKind: kind(word(fields,3)),anchorId: id === '-' ? undefined : integer(id),key: unescapeField(word(fields,5)),payload: unescapeField(word(fields,6))});
    }
    if(word(lines,at) !== 'end-checkpoint' || word(lines,at + 1) !== '' || at + 2 !== lines.length) { panic('invalid checkpoint end'); }
    if(writeCheckpoint(checkpoint) !== text) { panic('noncanonical checkpoint'); }
    return checkpoint;
}
