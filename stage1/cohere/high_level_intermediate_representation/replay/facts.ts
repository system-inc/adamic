// Syntax/checker inputs are distinct from pass-owned effect/reactive answers.
import { panic } from 'adamic';
import { Checkpoint, escapeField, integer } from './checkpoint.ts';
import type { AnchorKind } from './checkpoint.ts';
import type { SourceHandleInterface, ConstructedHIR, ModuleExportOriginInterface } from '../core.ts';
import { BasicBlock } from '../core.ts';
import { SymbolSnapshot } from '../symbol.ts';
import { decodeGraph, readPlace, readTerminal } from './decode.ts';
import { Payload } from './payload.ts';
function key(namespace: string,path: string,kind: AnchorKind,id: number | undefined,name: string): string { return `${escapeField(namespace)}\t${escapeField(path)}\t${kind}\t${id === undefined ? '-' : id}\t${escapeField(name)}`; }
export class InputFacts {
 readonly checkpoint: Checkpoint; private readonly values = new Map<string,string>();
 constructor(checkpoint: Checkpoint) { this.checkpoint = checkpoint; for(const row of checkpoint.sidecars) { this.values.set(key(row.namespace,row.functionPath,row.anchorKind,row.anchorId,row.key),row.payload); } }
 optional(namespace: string,path: string,kind: AnchorKind,id: number | undefined,name: string): string | undefined { this.checkpoint.graph.validateAnchor({namespace,functionPath: path,anchorKind: kind,anchorId: id,key: name,payload: ''}); return this.values.get(key(namespace,path,kind,id,name)); }
 require(namespace: string,path: string,kind: AnchorKind,id: number | undefined,name: string): string { return this.optional(namespace,path,kind,id,name) ?? panic(`missing input fact ${namespace}:${path}:${kind}:${id}:${name}`); }
 checkerAvailable(path: string): boolean { const value = this.require('input.checker',path,'function',undefined,'available'); if(value !== '0' && value !== '1') { panic('invalid checker availability'); } return value === '1'; }
 node(path: string,kind: AnchorKind,id: number | undefined): SourceHandleInterface | undefined {
  const value = this.require('input.ast',path,kind,id,'node'); if(value === 'nil') { return undefined; }
  const fields = value.split('\t'); if(fields.length !== 3) { panic('invalid source handle'); }
  return {kind: fields[0] ?? '',start: integer(fields[1] ?? ''),end: integer(fields[2] ?? '')};
 }
 valueTypePresent(path: string,id: number): boolean {
  if(!this.checkerAvailable(path) || this.node(path,'identifier',id) === undefined) { return false; }
  const value = this.require('input.types',path,'identifier',id,'value-present');
  if(value !== '0' && value !== '1') { panic('invalid value type presence'); } return value === '1';
 }
 typeFlags(path: string,id: number): number | undefined {
  if(!this.valueTypePresent(path,id)) { return undefined; } return integer(this.require('input.types',path,'identifier',id,'flags'));
 }
 // Exactly reactive.go:662: alias's present symbol wins, including an empty name.
 stableTypeName(path: string,id: number): string {
  if(!this.checkerAvailable(path)) { return ''; }
  if(this.node(path,'identifier',id) === undefined) { return ''; }
  if(this.require('input.types',path,'identifier',id,'alias-present') === '1') { return this.require('input.types',path,'identifier',id,'alias-symbol'); }
  return this.require('input.types',path,'identifier',id,'type-symbol');
 }
 source(): string { return this.require('input.source','$','function',undefined,'text'); }
 fileKind(): string { return this.require('input.source','$','function',undefined,'kind'); }
 symbols(): SymbolSnapshot { return new SymbolSnapshot(this.require('input.source','$','function',undefined,'symbols')); }
 // Callee origins are construction INPUT, retained in the instruction payload.
 calleeOrigin(path: string,id: number): ModuleExportOriginInterface | undefined {
  const fn = this.checkpoint.graph.at(path); const row = fn.instructionRow(fn.instruction(id));
  if(row.kind !== 'CallExpression' && row.kind !== 'MethodCall') { return undefined; }
  const json = new Payload(row.payload); const origin = json.field(json.rootHandle(),'CalleeOrigin');
  return {module: json.stringField(origin,'Module'),exported: json.stringField(origin,'Export')};
 }
 // Unit 4 supplies its signature-input codec in its own namespace, never effects answers.
 calleeSignature(path: string,id: number,name: string): string { return this.require('input.callee-signatures',path,'instruction',id,name); }
}
export function decodeCheckpoint(checkpoint: Checkpoint): ConstructedHIR {
 const graph = decodeGraph(checkpoint.graph); const facts = new InputFacts(checkpoint);
 for(let slot = 0; slot < checkpoint.graph.functions.length; slot++) {
  const record = checkpoint.graph.functions[slot] ?? panic('missing replay function'); const fn = graph.arena.read(graph.arena.indexAt(slot));
  fn.source = facts.node(record.path,'function',undefined);
  for(const identifier of fn.identifiers) { identifier.source = facts.node(record.path,'identifier',identifier.id.slot); }
  for(const instruction of fn.instructions) { instruction.source = facts.node(record.path,'instruction',instruction.id.slot); }
  for(const block of fn.blockTable) { block.present = fn.retained.has(block.id.slot + 1); }
  for(const row of checkpoint.select('input.blocks',record.path)) {
   const lines = row.payload.split('\n'); const header = (lines[0] ?? '').split(' '); const id = fn.blockAt(integer(header[1] ?? ''));
   const block = new BasicBlock(fn.blockIndices,id,{kind: 'Unreachable'},header[2] ?? 'block');
   block.predecessors = (header[3] ?? '').slice(13).split(',').filter((value) => value !== '').map((value) => fn.blockAt(integer(value)));
   for(const line of lines.slice(1)) {
    if(line.startsWith('instruction-ref ')) { block.instructions.push(fn.instructionIndices[integer(line.slice(16))] ?? panic('unknown detached instruction')); }
    else if(line.startsWith('terminal ')) { block.terminal = readTerminal(fn,line.slice(9)); block.terminalOrder = integer(line.slice(9).split(' ')[0] ?? ''); }
    else if(line === 'terminal-null') { block.terminalPresent = false; }
    else if(line.startsWith('phi ')) { const words = line.slice(4).split(' '); block.phis = [...block.phis,{place: readPlace(fn,words[0] ?? ''),operands: words.slice(1).map((operand) => { const parts = operand.split('='); return {predecessor: integer(parts[0] ?? ''),place: readPlace(fn,parts[1] ?? '')}; })}]; }
    else if(line !== '') { panic('unknown detached block row'); }
   }
   fn.blockTable[id.slot] = block;
  }
  // A real Go allocator value, including missing map slots, is checked explicitly.
  const nextBlock = integer(facts.require('input.identity',record.path,'function',undefined,'next-block'));
  if(fn.nextBlock !== nextBlock) { panic('block allocator high-water mark differs'); }
  for(const instruction of fn.instructions) {
   const value = instruction.value;
   if(value.kind === 'Destructure') {
    const shares = facts.require('input.pattern',record.path,'instruction',instruction.id.slot,'lvalue-shares-pattern');
    if(shares === '1') { value.lvaluePattern = value.pattern; } else if(shares !== '0') { panic('invalid pattern alias fact'); }
   }
  }
 }
 return graph;
}
