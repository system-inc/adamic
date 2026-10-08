// Go inline_iife.go:81/116. The guarded and memo-inclusive entry points stay distinct.
import { panic } from 'adamic';
import { HIRArena, HIRFunction } from '../../core.ts';
import type { IdentifierIndex, InstructionIndex, PlaceInterface, SourceHandleInterface, BasicBlock } from '../../core.ts';
import { copyNestedBodyInto } from './inline_remap.ts';
import type { InlineRemapInterface } from './inline_remap.ts';
import { eachPlace } from './places.ts';
import { normalizeGraph } from './graph_adapter.ts';
export function inlineIIFE(arena: HIRArena,fn: HIRFunction,includeMemoCallbacks: boolean): number {
 const memoized = new Set<IdentifierIndex>();
 if(!includeMemoCallbacks) { for(const instruction of fn.instructions) { if(instruction.value.kind === 'FinishMemoize') { memoized.add(instruction.value.value.identifier); } } }
 const spliced = new Set<IdentifierIndex>(); const queue = [...fn.blockOrder];
 // Candidate records are instruction handles, never references into graph nodes.
 const functions = new Map<IdentifierIndex,InstructionIndex>(); let count = 0;

 for(let position = 0; position < queue.length; position++) {
  const block = fn.block(queue[position] ?? panic('missing queued block'));
  if(block.kind === 'value' || block.kind === 'sequence') { continue; }
  for(let i = 0; i < block.instructions.length; i++) {
   const iid = block.instructions[i] ?? panic('missing instruction'); const instruction = fn.instruction(iid); const value = instruction.value;
   if(value.kind === 'FunctionExpression') { if(fn.identifier(instruction.lvalue.identifier).name === '') { functions.set(instruction.lvalue.identifier,iid); } continue; }
   if(value.kind !== 'CallExpression') { forgetCandidate(fn,functions,iid); continue; }
   const candidate = functions.get(value.callee.identifier);
   if(candidate === undefined || memoized.has(instruction.lvalue.identifier) || value.args.length !== 0) { forgetCandidate(fn,functions,iid); continue; }
   const expression = fn.instruction(candidate).value; if(expression.kind !== 'FunctionExpression') { panic('candidate is not a function'); }
   const nested = arena.read(expression.functionReference.index);
   if(nested.params.length !== 0 || nested.isAsync || nested.isGenerator) { forgetCandidate(fn,functions,iid); continue; }
   const remap = copyNestedBodyInto(fn,nested,expression.captures);
   if(remap === undefined || fn.blockOrUndefined(remap.entry) === undefined) { forgetCandidate(fn,functions,iid); continue; }
   const continuation = fn.newBlock(block.kind);
   for(const id of block.instructions.slice(i + 1)) { continuation.instructions.push(id); }
   continuation.terminal = block.terminal; continuation.terminalOrder = block.terminalOrder; continuation.terminalPresent = block.terminalPresent;
   while(block.instructions.length > i) { block.instructions.pop(); }
   const direct = singleReturnExit(fn,remap);
   block.terminal = direct ? {kind: 'Goto',block: remap.entry,variant: 0} : {kind: 'Label',block: remap.entry,fallthrough: continuation.id}; block.terminalOrder = 0; block.terminalPresent = true;
   if(!direct) { append(fn,block,fn.temporary(0,0),{kind: 'DeclareLocal',lvalue: instruction.lvalue,declarationKind: 1},instruction.source); }
   for(const id of remap.blocks.values()) {
    const copied = fn.block(id); if(copied.terminal.kind !== 'Return' || !copied.terminalPresent) { continue; }
    let returned = copied.terminal.value ?? panic('missing return value');
    const last = copied.instructions[copied.instructions.length - 1];
    if(last !== undefined) { const v = fn.instruction(last).value; if(v.kind === 'LoadLocal' && v.place.identifier.slot !== 0) { returned = v.place; } }
    if(direct) { append(fn,copied,instruction.lvalue,{kind: 'LoadLocal',place: returned},instruction.source); }
    else { append(fn,copied,fn.temporary(0,0),{kind: 'StoreLocal',lvalue: instruction.lvalue,value: returned,declarationKind: 2},instruction.source); }
    copied.terminal = {kind: 'Goto',block: continuation.id,variant: 0}; copied.terminalOrder = 0;
   }
   queue.push(continuation.id); spliced.add(value.callee.identifier); count++; break;
  }
 }
 if(count > 0) {
  for(const bid of fn.blockOrder) { const b = fn.block(bid); const kept = b.instructions.filter((id) => !(spliced.has(fn.instruction(id).lvalue.identifier) && fn.instruction(id).value.kind === 'FunctionExpression')); while(b.instructions.length > 0) { b.instructions.pop(); } for(const id of kept) { b.instructions.push(id); } }
  normalizeGraph(fn);
 }
 return count;
}
function singleReturnExit(fn: HIRFunction,remap: InlineRemapInterface): boolean {
 let exits = 0; let returns = 0;
 for(const id of remap.blocks.values()) { const block = fn.block(id); if(!block.terminalPresent) { continue; } if(block.terminal.kind === 'Return') { returns++; exits++; } else if(block.terminal.kind === 'Throw') { exits++; } }
 return exits === 1 && returns === 1;
}
import type { ValueType } from '../../core.ts';
function append(fn: HIRFunction,b: BasicBlock,p: PlaceInterface,v: ValueType,source: SourceHandleInterface | undefined): void { const id = fn.emit(p,v,0,0); fn.instruction(id).source = source; b.instructions.push(id); }

function forgetCandidate(fn: HIRFunction,functions: Map<IdentifierIndex,InstructionIndex>,iid: InstructionIndex): void { eachPlace(fn,fn.instruction(iid).value,(p) => { functions.delete(p.identifier); }); }
