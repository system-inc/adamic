// Go invoked_functions.go:57. Aliases share a candidate; closure is a fixpoint.
// Map keys are Go function-local identities, not fabricated arena handles.
// Keep the ordinal: repeated table entries can reference the same arena function.
import { HIRArena, HIRFunction } from '../../core.ts';
import type { FunctionIndex, IdentifierIndex, FunctionReferenceInterface } from '../../core.ts';
import { Unit3Facts } from './facts.ts';
interface CandidateInterface { readonly functionReference: FunctionReferenceInterface; readonly mayInvoke: Map<number,FunctionIndex>; }
export function collectAssumedInvokedFunctions(arena: HIRArena,fn: HIRFunction,facts: Unit3Facts): Map<number,FunctionIndex> {
 const invoked = new Map<number,FunctionIndex>(); const candidates = new Map<IdentifierIndex,CandidateInterface>();
 for(const bid of fn.blockOrder) { for(const iid of fn.block(bid).instructions) {
  const instruction = fn.instruction(iid); const v = instruction.value;
  if(v.kind === 'FunctionExpression') { candidates.set(instruction.lvalue.identifier,{functionReference: v.functionReference,mayInvoke: new Map<number,FunctionIndex>()}); }
  else if(v.kind === 'StoreLocal') { const held = candidates.get(v.value.identifier); if(held !== undefined) { candidates.set(v.lvalue.identifier,held); } }
  else if(v.kind === 'LoadLocal') { const held = candidates.get(v.place.identifier); if(held !== undefined) { candidates.set(instruction.lvalue.identifier,held); } }
 } }
 const mark = (id: IdentifierIndex): void => { const candidate = candidates.get(id); if(candidate !== undefined) { invoked.set(candidate.functionReference.ordinal,candidate.functionReference.index); } };
 for(const bid of fn.blockOrder) {
  const block = fn.block(bid);
  for(const iid of block.instructions) {
   const instruction = fn.instruction(iid); const v = instruction.value;
   if(v.kind === 'CallExpression') { const held = candidates.get(v.callee.identifier); if(held !== undefined) { invoked.set(held.functionReference.ordinal,held.functionReference.index); } else if(facts.hook(fn,v.callee.identifier)) { for(const argument of v.args) { mark(argument.place.identifier); } } }
   else if(v.kind === 'JsxExpression') { for(const prop of v.props) { if(!prop.spread) { mark(prop.value.identifier); } } for(const child of v.children) { mark(child.identifier); } }
   else if(v.kind === 'FunctionExpression') {
    const nested = arena.read(v.functionReference.index); const holder = candidates.get(instruction.lvalue.identifier);
    if(holder === undefined || nested.context.length === 0 || collectAssumedInvokedFunctions(arena,nested,facts).size === 0) { continue; }
    const parentFunctions = new Map<number,FunctionReferenceInterface>();
    for(const parentBid of fn.blockOrder) { for(const parentIid of fn.block(parentBid).instructions) { const p = fn.instruction(parentIid); if(p.value.kind === 'FunctionExpression') { parentFunctions.set(fn.identifier(p.lvalue.identifier).id.slot,p.value.functionReference); } } }
    for(const capture of nested.context) { const reference = parentFunctions.get(nested.identifier(capture.identifier).id.slot); if(reference !== undefined) { holder.mayInvoke.set(reference.ordinal,reference.index); } }
   }
  }
  if(block.terminalPresent && block.terminal.kind === 'Return' && block.terminal.value !== undefined) { mark(block.terminal.value.identifier); }
 }
 while(true) { let grew = false; for(const candidate of candidates.values()) { if(!invoked.has(candidate.functionReference.ordinal)) { continue; } for(const reached of candidate.mayInvoke.keys()) { if(!invoked.has(reached)) { const index = candidate.mayInvoke.get(reached); if(index !== undefined) { invoked.set(reached,index); grew = true; } } } } if(!grew) { return invoked; } }
}
