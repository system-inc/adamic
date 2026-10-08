// Go dead_code_elimination.go:43. Tables retain swept instructions; block lists do not.
import { HIRFunction } from '../../core.ts';
import type { IdentifierIndex, ValueType } from '../../core.ts';
import { eachPlace, eachTerminalPlace } from './places.ts';
export interface DeadCodeEliminationResultInterface { readonly instructions: number; readonly phis: number; }
export function eliminateDeadCode(fn: HIRFunction): DeadCodeEliminationResultInterface {
 const used = new Set<IdentifierIndex>(); const names = new Set<string>();
 const reference = (id: IdentifierIndex): void => { used.add(id); const name = fn.identifier(id).name; if(name !== '') { names.add(name); } };
 const position = new Map<number,number>(); for(let i = 0; i < fn.blockOrder.length; i++) { const id = fn.blockOrder[i]; if(id !== undefined) { position.set(fn.block(id).id.slot + 1,i); } }
 let loops = false; for(let i = 0; i < fn.blockOrder.length; i++) { const id = fn.blockOrder[i]; if(id !== undefined) { for(const p of fn.block(id).predecessors) { const at = position.get(p.slot + 1); if(at !== undefined && at >= i) { loops = true; } } } }
 while(true) {
  const size = used.size;
  for(let i = fn.blockOrder.length - 1; i >= 0; i--) {
   const bid = fn.blockOrder[i]; if(bid === undefined) { continue; } const block = fn.block(bid);
   if(block.terminalPresent) { eachTerminalPlace(block.terminal,(p,define) => { if(!define) { reference(p.identifier); } }); }
   for(let j = block.instructions.length - 1; j >= 0; j--) {
    const iid = block.instructions[j]; if(iid === undefined) { continue; } const instruction = fn.instruction(iid); const v = instruction.value;
    if(block.kind !== 'block' && j === block.instructions.length - 1) { reference(instruction.lvalue.identifier); eachPlace(fn,v,(p,d) => { if(!d) { reference(p.identifier); } }); continue; }
    if(!liveIdentifier(fn,used,names,instruction.lvalue.identifier) && pruneableValue(fn,used,names,v)) { continue; }
    reference(instruction.lvalue.identifier);
    if(v.kind === 'StoreLocal') { if(v.declarationKind === 2 || used.has(v.lvalue.identifier)) { reference(v.value.identifier); } continue; }
    eachPlace(fn,v,(p,d) => { if(!d) { reference(p.identifier); } });
   }
   for(const phi of block.phis) { if(liveIdentifier(fn,used,names,phi.place.identifier)) { for(const operand of phi.operands) { reference(operand.place.identifier); } } }
  }
  if(used.size <= size || !loops) { break; }
 }
 let instructions = 0; let phis = 0;
 for(const bid of fn.blockOrder) { const b = fn.block(bid); const kept = b.instructions.filter((id) => { const yes = liveIdentifier(fn,used,names,fn.instruction(id).lvalue.identifier); if(!yes) { instructions++; } return yes; }); while(b.instructions.length > 0) { b.instructions.pop(); } for(const id of kept) { b.instructions.push(id); }
 b.phis = b.phis.filter((phi) => { const yes = liveIdentifier(fn,used,names,phi.place.identifier); if(!yes) { phis++; } return yes; }); }
 return {instructions,phis};
}

function liveIdentifier(fn: HIRFunction,used: Set<IdentifierIndex>,names: Set<string>,id: IdentifierIndex): boolean { return used.has(id) || (fn.identifier(id).name !== '' && names.has(fn.identifier(id).name)); }
function pruneableValue(fn: HIRFunction,used: Set<IdentifierIndex>,names: Set<string>,v: ValueType): boolean {
  if(v.kind === 'DeclareLocal') { return !liveIdentifier(fn,used,names,v.lvalue.identifier); }
  if(v.kind === 'StoreLocal') { return v.declarationKind === 2 ? !used.has(v.lvalue.identifier) : !liveIdentifier(fn,used,names,v.lvalue.identifier); }
  return v.kind === 'RegExpLiteral' || v.kind === 'MetaProperty' || v.kind === 'LoadGlobal' || v.kind === 'ArrayExpression' || v.kind === 'BinaryExpression' || v.kind === 'ComputedLoad' || v.kind === 'FunctionExpression' || v.kind === 'LoadLocal' || v.kind === 'JsxExpression' || v.kind === 'JsxFragment' || v.kind === 'JsxText' || v.kind === 'ObjectExpression' || v.kind === 'Primitive' || v.kind === 'PropertyLoad' || v.kind === 'TemplateLiteral' || v.kind === 'UnaryExpression' || v.kind === 'TypeCastExpression';

}
