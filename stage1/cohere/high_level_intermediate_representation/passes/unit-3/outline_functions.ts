// Go outline_functions.go:38.
import { HIRArena } from '../../core.ts';
import type { FunctionIndex, IdentifierIndex } from '../../core.ts';
export function outlineFunctions(arena: HIRArena,index: FunctionIndex): number {
 const fn = arena.read(index); let count = 0;
 for(const bid of fn.blockOrder) { for(const iid of fn.block(bid).instructions) {
  const instruction = fn.instruction(iid); const value = instruction.value;
  if(value.kind !== 'FunctionExpression') { continue; }
  const nested = arena.read(value.functionReference.index); count += outlineFunctions(arena,value.functionReference.index);
  if(nested.context.length !== 0 || value.captures.length !== 0 || nested.name !== '') { continue; }
  instruction.value = {kind: 'LoadGlobal',name: value.functionReference.ordinal === 0 ? '_temp' : `_temp${value.functionReference.ordinal + 1}`,bindingKind: 0,source: '',imported: ''};
  if(fn.outlined === undefined) { fn.outlined = new Map<IdentifierIndex,FunctionIndex>(); }
  fn.outlined.set(instruction.lvalue.identifier,value.functionReference.index); count++;
 } }
 return count;
}
