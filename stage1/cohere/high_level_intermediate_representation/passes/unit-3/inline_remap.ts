// Go inline_remap.go:67. Fresh record copies use the owner's typed copy API.
import { panic } from 'adamic';
import { HIRFunction } from '../../core.ts';
import type { IdentifierIndex, DeclarationIndex, BlockIndex, InstructionIndex, FunctionIndex, PlaceInterface, FunctionReferenceInterface } from '../../core.ts';
import { copyInstructionValueWithRemap, copyTerminalWithRemap, copyPattern } from '../../replay/index.ts';
export interface InlineRemapInterface {
 readonly identifiers: Map<IdentifierIndex,IdentifierIndex>;
 readonly blocks: Map<BlockIndex,BlockIndex>;
 readonly instructions: Map<InstructionIndex,InstructionIndex>;
 readonly entry: BlockIndex;
}
export function copyNestedBodyInto(parent: HIRFunction,nested: HIRFunction,captures: readonly PlaceInterface[]): InlineRemapInterface | undefined {
 if(nested.context.length !== captures.length) { return undefined; }
 const identifiers = new Map<IdentifierIndex,IdentifierIndex>(); const blocks = new Map<BlockIndex,BlockIndex>(); const instructions = new Map<InstructionIndex,InstructionIndex>();
 const declarations = new Map<DeclarationIndex,DeclarationIndex>();
 for(let i = 0; i < nested.context.length; i++) {
  const context = nested.context[i] ?? panic('missing context'); const capture = captures[i] ?? panic('missing capture');
  identifiers.set(context.identifier,capture.identifier);
  const from = nested.identifier(context.identifier).declaration; const to = parent.identifier(capture.identifier).declaration; const mapped = declarations.get(from);
  if(mapped !== undefined && mapped !== to) { return undefined; }
  declarations.set(from,to);
 }
 for(const old of nested.identifiers) {
  if(identifiers.has(old.id)) { continue; }
  const mapped = declarations.get(old.declaration);
  const copied = parent.named(old.name,0,0,mapped); const identifier = parent.identifier(copied.identifier); identifier.source = old.source; identifier.nodeIndex = old.nodeIndex;
  if(mapped === undefined) { declarations.set(old.declaration,identifier.declaration); }
  identifiers.set(old.id,copied.identifier);
 }
 for(const bid of nested.blockOrder) { const block = nested.block(bid); const copied = parent.newBlock(block.kind); copied.terminalPresent = false; blocks.set(block.id,copied.id); }
 const entry = blocks.get(nested.entry) ?? parent.entry;
 const functions = new Map<FunctionIndex,FunctionReferenceInterface>();
 for(const child of nested.functions) { const ordinal = parent.functions.length; parent.functions.push(child); functions.set(child,{index: child,ordinal}); }
 const p = (place: PlaceInterface): PlaceInterface => ({identifier: identifiers.get(place.identifier) ?? parent.identifierAt(nested.identifier(place.identifier).id.slot),effect: place.effect,reactive: place.reactive,start: place.start,end: place.end});
 const b = (block: BlockIndex): BlockIndex => blocks.get(block) ?? parent.blockAt(block.slot + 1);
 const f = (reference: FunctionReferenceInterface): FunctionReferenceInterface => functions.get(reference.index) ?? reference;
 for(const bid of nested.blockOrder) {
  const source = nested.block(bid); const target = parent.block(blocks.get(bid) ?? panic('missing copied block'));
  for(const iid of source.instructions) {
   const original = nested.instruction(iid); let value = copyInstructionValueWithRemap(nested,parent,original.value,p,f);
   // Go's visitor renames LValue; Pattern is retained syntax, not a bound place.
   if(value.kind === 'Destructure' && original.value.kind === 'Destructure') { value = {kind: 'Destructure',lvaluePattern: value.lvaluePattern,pattern: copyPattern(nested,parent,original.value.pattern),value: value.value,declarationKind: value.declarationKind}; }
   // Go EachPlace skips the zero place in global memo roots.
   if(value.kind === 'StartMemoize' && original.value.kind === 'StartMemoize' && value.deps !== undefined && original.value.deps !== undefined) {
    for(let i = 0; i < value.deps.length; i++) { const dep = value.deps[i]; const old = original.value.deps[i]; if(dep !== undefined && old !== undefined && dep.root.isGlobal) { const place = old.root.place; dep.root.place = {identifier: parent.identifierAt(place.identifier.slot),effect: place.effect,reactive: place.reactive,start: place.start,end: place.end}; } }
   }
   const id = parent.emit(p(original.lvalue),value,original.start,original.end); const copied = parent.instruction(id); copied.order = original.order; copied.source = original.source;
   target.instructions.push(id); instructions.set(iid,id);
  }
  target.terminalPresent = source.terminalPresent;
  if(source.terminalPresent) { target.terminal = copyTerminalWithRemap(source.terminal,p,b); target.terminalOrder = source.terminalOrder; }
  target.phis = source.phis.map((phi) => ({place: p(phi.place),operands: phi.operands.map((op) => ({predecessor: b(nested.blockAt(op.predecessor)).slot + 1,place: p(op.place)})).sort((left,right) => left.predecessor - right.predecessor)}));
 }
 return {identifiers,blocks,instructions,entry};
}
