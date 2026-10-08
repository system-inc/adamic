// Go optional_chains.go: source dependency projection used inside DropManualMemoization.
// This owns only source-path recovery; it does not supply inferred scope answers.
import { panic } from 'adamic';
import type { HIRFunction, BlockIndex, IdentifierIndex, PlaceInterface, BasicBlock, TerminalType, DependencyPathEntryInterface } from '../../core.ts';
export interface OptionalSourceInterface { readonly place: PlaceInterface; readonly path: DependencyPathEntryInterface[]; }
export function equalPaths(a: readonly DependencyPathEntryInterface[],b: readonly DependencyPathEntryInterface[]): boolean { return a.length === b.length && a.every((p,i) => { const q = b[i]; return q !== undefined && p.property === q.property && p.optional === q.optional; }); }
export function optionalJoins(fn: HIRFunction): Set<BlockIndex> { const joins = new Set<BlockIndex>(); for(const id of fn.blockOrder) { const b = fn.block(id); if(b.terminalPresent && b.terminal.kind === 'Optional' && b.terminal.fallthrough !== undefined) { joins.add(b.terminal.fallthrough); } } return joins; }
export function optionalSources(fn: HIRFunction): Map<IdentifierIndex,OptionalSourceInterface> {
 const result = new Map<IdentifierIndex,OptionalSourceInterface>(); const seen = new Set<BlockIndex>();
 const lookup = (id: BlockIndex | undefined): BasicBlock | undefined => id === undefined ? undefined : fn.blockOrUndefined(id);
 for(const id of fn.blockOrder) { const b = fn.block(id); if(b.terminal.kind === 'Optional' && !seen.has(id)) { traverseSource(fn,seen,result,b,undefined); } }
 const joins = optionalJoins(fn);
 for(const id of fn.blockOrder) {
  if(!joins.has(id)) { continue; }
  for(const phi of fn.block(id).phis) {
   if(result.has(phi.place.identifier)) { continue; }
   let resolved: OptionalSourceInterface | undefined = undefined; let conflict = false;
   for(const op of phi.operands) { const dep = result.get(op.place.identifier); if(dep === undefined) { continue; } if(resolved === undefined) { resolved = dep; } else if(resolved.place.identifier !== dep.place.identifier || !equalPaths(resolved.path,dep.path)) { conflict = true; break; } }
   if(resolved !== undefined && !conflict) { result.set(phi.place.identifier,resolved); }
  }
 }
 return result;
}
export function optionalPlaces(fn: HIRFunction): Set<IdentifierIndex> {
 const result = new Set<IdentifierIndex>();
 for(const id of fn.blockOrder) {
  const optional = fn.block(id).terminal; if(optional.kind !== 'Optional' || optional.optionalFlag?.value !== true || optional.testBlock === undefined) { continue; }
  let test = fn.blockOrUndefined(optional.testBlock); const seen = new Set<BlockIndex>();
  while(test !== undefined && !seen.has(test.id)) {
   seen.add(test.id); const terminal = test.terminal;
   if(terminal.kind === 'Branch' && terminal.fallthrough === optional.fallthrough) {
    const consequent = terminal.consequent === undefined ? undefined : fn.blockOrUndefined(terminal.consequent); const last = consequent === undefined ? undefined : consequent.instructions[consequent.instructions.length - 1];
    if(last !== undefined) { const v = fn.instruction(last).value; if(v.kind === 'StoreLocal') { result.add(v.value.identifier); } } break;
   }
   if(terminal.kind !== 'Branch' && terminal.kind !== 'Optional' && terminal.kind !== 'Logical' && terminal.kind !== 'Ternary') { break; }
   test = terminal.fallthrough === undefined ? undefined : fn.blockOrUndefined(terminal.fallthrough);
  }
 }
 return result;
}

// Plain supported recursion with explicit traversal state.
function traverseSource(fn: HIRFunction,seen: Set<BlockIndex>,result: Map<IdentifierIndex,OptionalSourceInterface>,block: BasicBlock,outerAlternate: BlockIndex | undefined): IdentifierIndex | undefined {
 const lookup = (id: BlockIndex | undefined): BasicBlock | undefined => id === undefined ? undefined : fn.blockOrUndefined(id);
  seen.add(block.id); const optional = block.terminal; if(optional.kind !== 'Optional') { return undefined; }
  const maybeTest = lookup(optional.testBlock); if(maybeTest === undefined) { return undefined; }
  let base: OptionalSourceInterface; let test: TerminalType;
  if(maybeTest.terminal.kind === 'Branch') {
   const firstId = maybeTest.instructions[0]; if(firstId === undefined) { return undefined; }
   const first = fn.instruction(firstId); const value = first.value;
   if(value.kind !== 'LoadLocal' && value.kind !== 'LoadContext') { return undefined; }
   const path: DependencyPathEntryInterface[] = [];
   for(let i = 1; i < maybeTest.instructions.length; i++) {
    const iid = maybeTest.instructions[i] ?? panic('missing property'); const previous = maybeTest.instructions[i - 1] ?? panic('missing previous'); const load = fn.instruction(iid).value;
    if(load.kind !== 'PropertyLoad' || load.object.identifier !== fn.instruction(previous).lvalue.identifier) { return undefined; }
    path.push({property: load.property,optional: false});
   }
   base = {place: value.place,path}; test = maybeTest.terminal;
  } else if(maybeTest.terminal.kind === 'Optional') {
   const testBlock = lookup(maybeTest.terminal.fallthrough); if(testBlock === undefined || testBlock.terminal.kind !== 'Branch') { return undefined; }
   test = testBlock.terminal; const inner = traverseSource(fn,seen,result,maybeTest,test.alternate);
   if(inner === undefined || inner.slot === 0 || test.testPlace?.identifier !== inner) { return undefined; }
   const dependency = result.get(inner); if(dependency === undefined) { return undefined; } base = dependency;
  } else { return undefined; }
  if(outerAlternate !== undefined && test.alternate === outerAlternate && block.instructions.length !== 0) { return undefined; }
  const consequent = lookup(test.consequent); const alternate = lookup(test.alternate);
  if(consequent === undefined || alternate === undefined || consequent.instructions.length !== 2 || alternate.instructions.length !== 2) { return undefined; }
  const load = fn.instruction(consequent.instructions[0] ?? panic('missing optional load'));
  const store = fn.instruction(consequent.instructions[1] ?? panic('missing optional store')).value;
  if(load.value.kind !== 'PropertyLoad' || store.kind !== 'StoreLocal' || load.value.object.identifier !== test.testPlace?.identifier || store.value.identifier !== load.lvalue.identifier) { return undefined; }
  if(consequent.terminal.kind !== 'Goto' || consequent.terminal.variant !== 0 || consequent.terminal.block !== optional.fallthrough) { return undefined; }
  if(fn.instruction(alternate.instructions[0] ?? panic('missing alternate primitive')).value.kind !== 'Primitive' || fn.instruction(alternate.instructions[1] ?? panic('missing alternate store')).value.kind !== 'StoreLocal') { return undefined; }
  const dependency: OptionalSourceInterface = {place: {identifier: base.place.identifier,effect: '<unknown>',reactive: base.place.reactive,start: 0,end: 0},path: [...base.path,{property: load.value.property,optional: optional.optionalFlag?.value === true}]};
  result.set(store.lvalue.identifier,dependency); result.set(load.lvalue.identifier,dependency); return store.lvalue.identifier;

}
