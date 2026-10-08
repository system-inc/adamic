// Go drop_manual_memoization.go:198. Callee anchors, not call positions, own starts.
import { panic } from 'adamic';
import { HIRFunction } from '../../core.ts';
import type { IdentifierIndex, InstructionIndex, Instruction, PlaceInterface, ValueType, ManualMemoDependencyInterface, SourceHandleInterface } from '../../core.ts';
import { optionalSources, optionalPlaces, optionalJoins, equalPaths } from './optional_sources.ts';
import { refreshOrder } from './graph_adapter.ts';
export interface ManualMemoizationInterface { recognised: number; marked: number; dependencies: number; withoutDepsArray: number; unextractableDeps: number; notAnArrayLiteral: number; notAnInlineFunction: number; }
interface QueuedMarkerInterface { readonly after: InstructionIndex; readonly lvalue: PlaceInterface; readonly value: ValueType; readonly source: SourceHandleInterface | undefined; readonly start: number; readonly end: number; }
export function dropManualMemoization(fn: HIRFunction): ManualMemoizationInterface {
 const result: ManualMemoizationInterface = {recognised: 0,marked: 0,dependencies: 0,withoutDepsArray: 0,unextractableDeps: 0,notAnArrayLiteral: 0,notAnInlineFunction: 0};
 const functions = new Set<IdentifierIndex>(); const memos = new Map<IdentifierIndex,string>(); const react = new Set<IdentifierIndex>();
 const lists = new Map<IdentifierIndex,PlaceInterface[]>(); const deps = new Map<IdentifierIndex,ManualMemoDependencyInterface>(); const anchors = new Map<IdentifierIndex,InstructionIndex>();
 const optionals = optionalPlaces(fn); const joins = optionalJoins(fn); const queued: QueuedMarkerInterface[] = [];
 const zero: PlaceInterface = {identifier: fn.identifierAt(0),effect: '<unknown>',reactive: false,start: 0,end: 0};
 const optionalDependencies = optionalSources(fn);
 for(const id of optionalDependencies.keys()) { const dep = optionalDependencies.get(id); if(dep !== undefined) { deps.set(id,{root: {isGlobal: false,name: '',place: dep.place},path: dep.path}); } }
 const propagate = (target: IdentifierIndex,source: PlaceInterface): void => {
  const existing = deps.get(source.identifier); if(existing !== undefined) { deps.set(target,existing); }
  else if(fn.identifier(source.identifier).name !== '') { deps.set(target,{root: {isGlobal: false,name: '',place: source},path: []}); }
 };
 const collect = (instruction: Instruction): void => {
  const target = instruction.lvalue.identifier; const v = instruction.value; anchors.set(target,instruction.id);
  if(v.kind === 'FunctionExpression') { functions.add(target); }
  else if(v.kind === 'LoadGlobal') {
   if(v.name === 'useMemo' || v.name === 'useCallback') { memos.set(target,v.name); }
   else if(v.name === 'React') { react.add(target); }
   else { deps.set(target,{root: {isGlobal: true,name: v.name,place: zero},path: []}); }
  } else if(v.kind === 'Primitive') {
   if(v.literal === 'string:useMemo' || v.literal === 'string:useCallback') { memos.set(target,v.literal === 'string:useMemo' ? 'useMemo' : 'useCallback'); }
  } else if(v.kind === 'ArrayExpression') {
   if(v.elements.every((e) => !e.hole && !e.spread)) { lists.set(target,v.elements.map((e) => e.place)); }
  } else if(v.kind === 'PropertyLoad') {
   const object = deps.get(v.object.identifier); if(object !== undefined) { deps.set(target,{root: object.root,path: [...object.path,{property: v.property,optional: v.optional || optionals.has(target)}]}); }
  } else if(v.kind === 'LoadLocal' || v.kind === 'LoadContext') { propagate(target,v.place); }
  else if(v.kind === 'StoreLocal') { const source = deps.get(v.value.identifier); if(source !== undefined && fn.identifier(v.lvalue.identifier).name === '') { deps.set(v.lvalue.identifier,source); } }
 };
 for(const bid of fn.blockOrder) {
  const block = fn.block(bid);
  if(joins.has(bid)) { for(const phi of block.phis) {
   if(deps.has(phi.place.identifier)) { continue; } let resolved: ManualMemoDependencyInterface | undefined = undefined; let conflict = false;
   for(const op of phi.operands) {
    const dep = deps.get(op.place.identifier); if(dep === undefined) { continue; }
    if(resolved === undefined) { resolved = dep; } else if(resolved.root.isGlobal !== dep.root.isGlobal || (dep.root.isGlobal ? resolved.root.name !== dep.root.name : resolved.root.place.identifier !== dep.root.place.identifier) || !equalPaths(resolved.path,dep.path)) { conflict = true; break; }
   }
   if(resolved !== undefined && !conflict) { deps.set(phi.place.identifier,resolved); }
  } }
  for(const iid of block.instructions) {
   const instruction = fn.instruction(iid); const v = instruction.value;
   let kind: string | undefined = undefined; let callee: IdentifierIndex | undefined = undefined;
   if(v.kind === 'CallExpression') { kind = memos.get(v.callee.identifier); callee = v.callee.identifier; }
   else if(v.kind === 'MethodCall' && react.has(v.receiver.identifier)) { kind = memos.get(v.property.identifier); callee = v.property.identifier; }
   if(kind === undefined || callee === undefined || (v.kind !== 'CallExpression' && v.kind !== 'MethodCall')) { collect(instruction); continue; }
   const first = v.args[0]; if(first === undefined || first.spread) { continue; }
   let dependencies: ManualMemoDependencyInterface[] | undefined = undefined;
   const second = v.args[1];
   if(second !== undefined && !second.spread) {
    const elements = lists.get(second.place.identifier);
    if(elements === undefined) { result.notAnArrayLiteral++; continue; }
    dependencies = [];
    for(const element of elements) { const dependency = deps.get(element.identifier); if(dependency === undefined) { result.unextractableDeps++; dependencies = undefined; break; } dependencies.push(dependency); }
   }
   result.recognised++;
   instruction.value = kind === 'useMemo' ? {kind: 'CallExpression',callee: first.place,args: [],optional: false,origin: {module: '',exported: ''}} : {kind: 'LoadLocal',place: first.place};
   if(!functions.has(first.place.identifier)) { result.notAnInlineFunction++; continue; }
   const memoId = result.marked; result.marked++;
   const marker = (after: InstructionIndex,value: ValueType): void => { const p = fn.temporary(instruction.start,instruction.end); queued.push({after,lvalue: p,value,source: instruction.source,start: instruction.start,end: instruction.end}); };
   marker(anchors.get(callee) ?? fn.instructionIndices[0] ?? panic('missing callee anchor'),{kind: 'StartMemoize',manualMemoId: memoId,deps: dependencies});
   marker(iid,{kind: 'FinishMemoize',manualMemoId: memoId,value: kind === 'useMemo' ? instruction.lvalue : first.place,pruned: false});
   result.dependencies += dependencies?.length ?? 0; if(dependencies === undefined) { result.withoutDepsArray++; }
  }
 }
 if(queued.length > 0) {
  for(const bid of fn.blockOrder) { const block = fn.block(bid); const rebuilt: InstructionIndex[] = [];
   for(const iid of block.instructions) { rebuilt.push(iid); for(const marker of queued) { if(marker.after !== iid) { continue; } const id = fn.emit(marker.lvalue,marker.value,marker.start,marker.end); fn.instruction(id).source = marker.source; rebuilt.push(id); } }
   while(block.instructions.length > 0) { block.instructions.pop(); } for(const id of rebuilt) { block.instructions.push(id); }
  }
  refreshOrder(fn);
 }
 return result;
}
