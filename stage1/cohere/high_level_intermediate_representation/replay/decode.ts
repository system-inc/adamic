import { panic } from 'adamic';
import { HIRArena, ConstructedHIR, BasicBlock, Instruction, IdentifierIndex, DeclarationIndex, BlockIndex, InstructionIndex } from '../core.ts';
import type { HIRFunction, PlaceInterface, EffectType, PatternIndex, ValueType, TerminalType, ArgumentInterface, FunctionIndex, PatternElementInterface, PatternPropertyInterface } from '../core.ts';
import { IdentityGraph, integer } from './checkpoint.ts';
import type { ReplayInstruction } from './checkpoint.ts';
import { Payload } from './payload.ts';
import type { PayloadIndex } from '../../arena/arena_index.a';
function effect(text: string): EffectType {
 if(text === '<unknown>' || text === 'freeze' || text === 'read' || text === 'capture' || text === 'mutate-iterator?' || text === 'mutate?' || text === 'mutate' || text === 'store') { return text; }
 return panic('unknown place effect');
}
export function readPlace(fn: HIRFunction,text: string): PlaceInterface {
 const fields = text.split(':'); if(fields.length !== 5) { panic('invalid place'); }
 const flag = fields[2]; if(flag !== '0' && flag !== '1') { panic('invalid reactive flag'); }
 return {identifier: fn.identifierAt(integer(fields[0] ?? '')),effect: effect(fields[1] ?? ''),reactive: flag === '1',start: integer(fields[3] ?? ''),end: integer(fields[4] ?? '')};
}
function pattern(fn: HIRFunction,json: Payload,index: PayloadIndex): PatternIndex {
 const node = json.read(index); if(node.kind !== 'object') { return panic('invalid pattern'); }
 const place = node.fields.get('Place'); if(place !== undefined) { return fn.addPattern({kind: 'Place',place: readPlace(fn,json.textAt(place))}); }
 const restIndex = json.field(index,'Rest'); const rest = json.read(restIndex).kind === 'null' ? undefined : readPlace(fn,json.textAt(restIndex));
 const properties = node.fields.get('Properties');
 if(properties !== undefined) {
  const result: PatternPropertyInterface[] = [];
  for(const item of json.array(properties)) {
   const computed = json.field(item,'ComputedKey'); const fallback = json.field(item,'Default');
   result.push({key: json.stringField(item,'Key'),computedKey: json.read(computed).kind === 'null' ? undefined : readPlace(fn,json.textAt(computed)),defaultValue: json.read(fallback).kind === 'null' ? undefined : readPlace(fn,json.textAt(fallback)),value: pattern(fn,json,json.field(item,'Value'))});
  }
  return fn.addPattern({kind: 'Object',properties: result,rest});
 }
 const elements: PatternElementInterface[] = [];
 for(const item of json.array(json.field(index,'Elements'))) {
  const value = json.field(item,'Value'); const fallback = json.field(item,'Default');
  elements.push({value: json.read(value).kind === 'null' ? undefined : pattern(fn,json,value),defaultValue: json.read(fallback).kind === 'null' ? undefined : readPlace(fn,json.textAt(fallback))});
 }
 return fn.addPattern({kind: 'Array',elements,rest});
}
export function readInstructionValue(fn: HIRFunction,row: ReplayInstruction): ValueType {
 const kind = row.kind; const text = row.payload;
 if(kind === 'Primitive') { return {kind: 'Primitive',literal: text}; }
 if(kind === 'LoadLocal') { return {kind: 'LoadLocal',place: readPlace(fn,text)}; }
 if(kind === 'UnaryExpression') { const space = text.indexOf(' '); return {kind: 'UnaryExpression',operator: text.slice(0,space),value: readPlace(fn,text.slice(space + 1))}; }
 if(kind === 'BinaryExpression') { const tokens = text.split(' '); return {kind: 'BinaryExpression',left: readPlace(fn,tokens[0] ?? ''),operator: tokens[1] ?? '',right: readPlace(fn,tokens[2] ?? '')}; }
 const json = new Payload(text); const root = json.rootHandle();
 const p = (name: string): PlaceInterface => readPlace(fn,json.stringField(root,name));
 const s = (name: string): string => json.stringField(root,name);
 const n = (name: string): number => json.numberField(root,name);
 const b = (name: string): boolean => json.booleanField(root,name);
 const places = (name: string): PlaceInterface[] => json.array(json.field(root,name)).map((item) => readPlace(fn,json.textAt(item)));
 const args = (): ArgumentInterface[] => json.array(json.field(root,'Args')).map((item) => ({place: readPlace(fn,json.stringField(item,'Place')),spread: json.booleanField(item,'Spread')}));
 if(kind === 'Debugger') { return {kind: 'Debugger'}; }
 if(kind === 'LoadContext') { return {kind: 'LoadContext',place: p('Place')}; }
 if(kind === 'DeclareLocal') { return {kind: 'DeclareLocal',lvalue: p('LValue'),declarationKind: n('Kind')}; }
 if(kind === 'DeclareContext') { return {kind: 'DeclareContext',lvalue: p('LValue'),declarationKind: n('Kind')}; }
 if(kind === 'StoreLocal') { return {kind: 'StoreLocal',lvalue: p('LValue'),value: p('Value'),declarationKind: n('Kind')}; }
 if(kind === 'StoreContext') { return {kind: 'StoreContext',lvalue: p('LValue'),value: p('Value'),declarationKind: n('Kind')}; }
 if(kind === 'PrefixUpdate') { return {kind: 'PrefixUpdate',lvalue: p('LValue'),value: p('Value'),operation: s('Operation')}; }
 if(kind === 'PostfixUpdate') { return {kind: 'PostfixUpdate',lvalue: p('LValue'),value: p('Value'),operation: s('Operation')}; }
 if(kind === 'LoadGlobal') { return {kind: 'LoadGlobal',name: s('Name'),bindingKind: n('BindingKind'),source: s('Source'),imported: s('Imported')}; }
 if(kind === 'StoreGlobal') { return {kind: 'StoreGlobal',name: s('Name'),value: p('Value')}; }
 if(kind === 'PropertyLoad') { return {kind: 'PropertyLoad',object: p('Object'),property: s('Property'),optional: b('Optional')}; }
 if(kind === 'PropertyDelete') { return {kind: 'PropertyDelete',object: p('Object'),property: s('Property')}; }
 if(kind === 'ComputedLoad') { return {kind: 'ComputedLoad',object: p('Object'),property: p('Property'),optional: b('Optional')}; }
 if(kind === 'ComputedDelete') { return {kind: 'ComputedDelete',object: p('Object'),property: p('Property')}; }
 if(kind === 'PropertyStore') { return {kind: 'PropertyStore',object: p('Object'),property: s('Property'),value: p('Value')}; }
 if(kind === 'ComputedStore') { return {kind: 'ComputedStore',object: p('Object'),property: p('Property'),value: p('Value')}; }
 if(kind === 'RegExpLiteral') { return {kind: 'RegExpLiteral',pattern: s('Pattern'),flags: s('Flags')}; }
 if(kind === 'MetaProperty') { return {kind: 'MetaProperty',meta: s('Meta'),property: s('Property')}; }
 if(kind === 'JsxText') { return {kind: 'JsxText',text: s('Value')}; }
 if(kind === 'JsxFragment') { return {kind: 'JsxFragment',children: places('Children')}; }
 if(kind === 'Await') { return {kind: 'Await',value: p('Value')}; }
 if(kind === 'GetIterator') { return {kind: 'GetIterator',value: p('Value')}; }
 if(kind === 'NextPropertyOf') { return {kind: 'NextPropertyOf',value: p('Value')}; }
 if(kind === 'IteratorNext') { return {kind: 'IteratorNext',iterator: p('Iterator'),collection: p('Collection')}; }
 if(kind === 'FinishMemoize') { return {kind: 'FinishMemoize',manualMemoId: n('ManualMemoId'),value: p('Value'),pruned: b('Pruned')}; }
 if(kind === 'TypeCastExpression' || kind === 'UnsupportedNode') {
  const node = json.field(root,'Node'); const nodeKind = json.stringField(node,'kind');
  if(kind === 'TypeCastExpression') { return {kind: 'TypeCastExpression',value: p('Value'),nodeKind: nodeKind.startsWith('Kind') ? nodeKind.slice(4) : nodeKind,nodePos: json.numberField(node,'pos'),nodeEnd: json.numberField(node,'end')}; }
  return {kind: 'UnsupportedNode',reason: s('Reason'),nodeKind,nodePos: json.numberField(node,'pos'),nodeEnd: json.numberField(node,'end')};
 }
 if(kind === 'FunctionExpression' || kind === 'ObjectMethod') {
  const ordinal = n('Function'); const functionReference = {ordinal,index: fn.functions[ordinal] ?? panic('unknown nested function')};
  if(kind === 'ObjectMethod') { return {kind: 'ObjectMethod',key: s('Key'),functionReference}; }
  return {kind: 'FunctionExpression',functionReference,captures: places('Captures')};
 }
 if(kind === 'CallExpression' || kind === 'MethodCall' || kind === 'NewExpression') {
  if(kind === 'NewExpression') { return {kind: 'NewExpression',callee: p('Callee'),args: args()}; }
  const originIndex = json.field(root,'CalleeOrigin'); const origin = {module: json.stringField(originIndex,'Module'),exported: json.stringField(originIndex,'Export')};
  if(kind === 'CallExpression') { return {kind: 'CallExpression',callee: p('Callee'),args: args(),optional: b('Optional'),origin}; }
  return {kind: 'MethodCall',receiver: p('Receiver'),property: p('Property'),args: args(),optional: b('Optional'),origin};
 }
 if(kind === 'Destructure') { return {kind: 'Destructure',pattern: pattern(fn,json,json.field(root,'Pattern')),lvaluePattern: pattern(fn,json,json.field(root,'LValue')),value: p('Value'),declarationKind: n('Kind')}; }
 if(kind === 'TemplateLiteral' || kind === 'TaggedTemplateExpression') {
  const quasis = json.array(json.field(root,'Quasis')).map((item) => json.textAt(item)); const subexprs = places('Subexprs');
  if(kind === 'TemplateLiteral') { return {kind: 'TemplateLiteral',quasis,subexprs}; }
  return {kind: 'TaggedTemplateExpression',tag: p('Tag'),quasis,subexprs};
 }
 if(kind === 'ArrayExpression') { return {kind: 'ArrayExpression',elements: json.array(json.field(root,'Elements')).map((item) => ({place: readPlace(fn,json.stringField(item,'Place')),hole: json.booleanField(item,'Hole'),spread: json.booleanField(item,'Spread')}))}; }
 if(kind === 'ObjectExpression') { return {kind: 'ObjectExpression',properties: json.array(json.field(root,'Properties')).map((item) => {
  const computed = json.field(item,'ComputedKey'); return {key: json.stringField(item,'Key'),computedKey: json.read(computed).kind === 'null' ? undefined : readPlace(fn,json.textAt(computed)),value: readPlace(fn,json.stringField(item,'Value')),spread: json.booleanField(item,'Spread')};
 })}; }
 if(kind === 'JsxExpression') {
  const tagIndex = json.field(root,'Tag'); const tagPlace = json.field(tagIndex,'Place');
  return {kind: 'JsxExpression',tag: {name: json.stringField(tagIndex,'Name'),place: json.read(tagPlace).kind === 'null' ? undefined : readPlace(fn,json.textAt(tagPlace))},children: places('Children'),props: json.array(json.field(root,'Props')).map((item) => ({name: json.stringField(item,'Name'),value: readPlace(fn,json.stringField(item,'Value')),spread: json.booleanField(item,'Spread')}))};
 }
 if(kind === 'StartMemoize') {
  const deps = json.field(root,'Deps'); if(json.read(deps).kind === 'null') { return {kind: 'StartMemoize',manualMemoId: n('ManualMemoId'),deps: undefined}; }
  return {kind: 'StartMemoize',manualMemoId: n('ManualMemoId'),deps: json.array(deps).map((dep) => {
   const rootIndex = json.field(dep,'Root'); return {root: {isGlobal: json.booleanField(rootIndex,'IsGlobal'),name: json.stringField(rootIndex,'Name'),place: readPlace(fn,json.stringField(rootIndex,'Place'))},path: json.array(json.field(dep,'Path')).map((item) => ({property: json.stringField(item,'Property'),optional: json.booleanField(item,'Optional')}))};
  })};
 }
 return panic(`unknown instruction variant ${kind}`);
}
export function readTerminal(fn: HIRFunction,text: string): TerminalType {
 const words = text.split(' '); const kind = words[1]; const payload = words.slice(2).join(' ');
 if(kind === 'Return') { return {kind: 'Return',value: readPlace(fn,payload)}; }
 if(kind === 'Unreachable' || kind === 'Unsupported') { return {kind}; }
 const json = new Payload(payload); const root = json.rootHandle();
 const block = (name: string): BlockIndex => fn.blockAt(json.numberField(root,name));
 const p = (name: string): PlaceInterface => readPlace(fn,json.stringField(root,name));
 if(kind === 'Throw') { return {kind: 'Throw',value: p('Value')}; }
 if(kind === 'Goto') { return {kind: 'Goto',block: block('Block'),variant: json.numberField(root,'Variant')}; }
 if(kind === 'If' || kind === 'Branch') { return {kind,testPlace: p('Test'),consequent: block('Consequent'),alternate: block('Alternate'),fallthrough: block('Fallthrough')}; }
 // gap 1 (GAPS.md): required presence/value booleans preserve the Optional field.
 if(kind === 'Optional') { return {kind: 'Optional',testBlock: block('Test'),fallthrough: block('Fallthrough'),optionalFlag: {present: true,value: json.booleanField(root,'Optional')}}; }
 if(kind === 'Logical') { return {kind: 'Logical',testBlock: block('Test'),fallthrough: block('Fallthrough'),operator: json.stringField(root,'Operator')}; }
 if(kind === 'Ternary') { return {kind: 'Ternary',testBlock: block('Test'),fallthrough: block('Fallthrough')}; }
 if(kind === 'While' || kind === 'DoWhile') { return {kind,testBlock: block('Test'),loop: block('Loop'),fallthrough: block('Fallthrough')}; }
 if(kind === 'For') { const update = json.numberField(root,'Update'); return {kind: 'For',testBlock: block('Test'),loop: block('Loop'),fallthrough: block('Fallthrough'),init: block('Init'),update: update === 0 ? undefined : fn.blockAt(update)}; }
 if(kind === 'ForOf') { return {kind: 'ForOf',testBlock: block('Test'),loop: block('Loop'),fallthrough: block('Fallthrough'),init: block('Init')}; }
 if(kind === 'ForIn') { return {kind: 'ForIn',loop: block('Loop'),fallthrough: block('Fallthrough'),init: block('Init')}; }
 if(kind === 'Label') { return {kind: 'Label',block: block('Block'),fallthrough: block('Fallthrough')}; }
 if(kind === 'Try') { const binding = json.field(root,'HandlerBinding'); return {kind: 'Try',block: block('Block'),handler: block('Handler'),fallthrough: block('Fallthrough'),handlerBinding: json.read(binding).kind === 'null' ? undefined : readPlace(fn,json.textAt(binding))}; }
 if(kind === 'Switch') { return {kind: 'Switch',testPlace: p('Test'),fallthrough: block('Fallthrough'),cases: json.array(json.field(root,'Cases')).map((item) => { const test = json.field(item,'Test'); return {test: json.read(test).kind === 'null' ? undefined : readPlace(fn,json.textAt(test)),block: fn.blockAt(json.numberField(item,'Block'))}; })}; }
 return panic(`unknown terminal ${kind}`);
}
// Allocate a new arena on every replay. No constructor, AST parser or analysis is rerun.
export function decodeGraph(graph: IdentityGraph): ConstructedHIR {
 const arena = new HIRArena();
 for(const record of graph.functions) {
  const header = graph.rows.find((row) => row.kind === 'function' && row.functionPath === record.path) ?? panic('missing function header');
  const words = (header.body ?? '').split(' '); arena.create(words[0] ?? '');
 }
 for(let i = 0; i < graph.functions.length; i++) {
  const record = graph.functions[i] ?? panic('missing identity record'); const index = graph.indices[i] ?? panic('missing function index');
  // ArenaIndices belong to the reader, so mint distinct handles in the HIR arena.
  const fn = arena.read(arenaIndex(arena,i,graph));
  while(fn.identifierIndices.length < record.identifiers.length) { IdentifierIndex.push(fn.identifierIndices); }
  let maxDeclaration = 1; for(const id of record.declarationIds) { if(id > maxDeclaration) { maxDeclaration = id; } }
  while(fn.declarationIndices.length < maxDeclaration) { DeclarationIndex.push(fn.declarationIndices); }
  while(fn.identifiers.length > 0) { fn.identifiers.pop(); }
  const ids = record.identifierIds.slice().sort((a,b) => a - b);
  for(let slot = 0; slot < ids.length; slot++) { if(ids[slot] !== slot) { panic('identifier table is not contiguous'); } }
  let maxBlock = 1; for(const id of record.blockIds) { if(id > maxBlock) { maxBlock = id; } }
  while(fn.blockIndices.length < maxBlock) { BlockIndex.push(fn.blockIndices); }
  while(fn.blockTable.length > 0) { fn.blockTable.pop(); }
  for(const block of fn.blockIndices) { fn.blockTable.push(new BasicBlock(fn.blockIndices,block,{kind: 'Unreachable'})); }
  while(fn.blockOrder.length > 0) { fn.blockOrder.pop(); } fn.retained.clear();
  for(const child of record.children) { fn.functions.push(arenaIndex(arena,child.slot,graph)); }
  const rows = record.instructionRows.slice().sort((a,b) => a.id - b.id);
  for(const row of rows) {
   if(row.id !== fn.instructions.length) { panic('instruction table is not contiguous'); }
   const id = InstructionIndex.push(fn.instructionIndices); const range = row.range.split(':');
   const instruction = new Instruction(id,readPlace(fn,row.lvalue),readInstructionValue(fn,row),integer(range[0] ?? ''),integer(range[1] ?? '')); instruction.order = row.order; fn.instructions.push(instruction);
  }
  for(const row of graph.rows) {
   if(row.functionPath !== record.path) { continue; }
   const body = row.body ?? ''; const words = body.split(' ');
   if(row.kind === 'identifier') { const id = integer(words[0] ?? ''); if(id !== fn.identifiers.length) { panic('identifier rows are not in table order'); } fn.identifiers.push({id: fn.identifierAt(id),declaration: fn.declarationAt(integer(words[1] ?? '')),name: words.slice(2).join(' '),nodeIndex: -1,source: undefined}); }
   else if(row.kind === 'function') { fn.kind = words[1] ?? ''; fn.entry = fn.blockAt(record.entry); fn.isAsync = words[3] === 'async=1'; fn.isGenerator = words[4] === 'generator=1'; }
   else if(row.kind === 'params' || row.kind === 'context') { if(body !== '') { const places = body.split(',').map((value) => readPlace(fn,value)); for(const place of places) { if(row.kind === 'params') { fn.params.push(place); } else { fn.context.push(place); } } } }
   else if(row.kind === 'returns') { fn.returns = readPlace(fn,body); }
   else if(row.kind === 'context-declarations') { for(const value of body.slice(1,-1).split(' ')) { if(value !== '') { fn.contextDeclarations.add(fn.declarationAt(integer(value))); } } }
   else if(row.kind === 'outlined') { const outlined = fn.outlined ?? new Map<IdentifierIndex,FunctionIndex>(); outlined.set(fn.identifierAt(integer(words[0] ?? '')),fn.functions[integer(words[1] ?? '')] ?? panic('missing outlined function')); fn.outlined = outlined; }
  }
  // These handles are validated through the graph owner before their slots are used.
  graph.function(index);
 }
 const current = new Map<string,BlockIndex>();
 for(const row of graph.rows) {
  const fn = arena.read(arenaIndex(arena,graph.functionIndex(row.functionPath).slot,graph)); const body = row.body ?? ''; const words = body.split(' ');
  if(row.kind === 'block') {
   const id = fn.blockAt(integer(words[0] ?? '')); const block = new BasicBlock(fn.blockIndices,id,{kind: 'Unreachable'},words[1] ?? 'block');
   block.predecessors = (words[2] ?? '').slice(13).split(',').filter((value) => value !== '').map((value) => fn.blockAt(integer(value)));
   fn.blockTable[id.slot] = block; fn.blockOrder.push(id); fn.retained.add(id.slot + 1); current.set(row.functionPath,id);
  } else if(row.kind === 'instruction') { fn.block(current.get(row.functionPath) ?? panic('instruction outside block')).instructions.push(fn.instructionIndices[integer(words[0] ?? '')] ?? panic('missing instruction')); }
  else if(row.kind === 'terminal') { const block = fn.block(current.get(row.functionPath) ?? panic('terminal outside block')); block.terminal = readTerminal(fn,body); block.terminalOrder = integer(words[0] ?? ''); }
  else if(row.kind === 'phi') {
   const block = fn.block(current.get(row.functionPath) ?? panic('phi outside block'));
   block.phis = [...block.phis,{place: readPlace(fn,words[0] ?? ''),operands: words.slice(1).map((operand) => { const parts = operand.split('='); const predecessor = integer(parts[0] ?? ''); fn.blockAt(predecessor); return {predecessor,place: readPlace(fn,parts[1] ?? '')}; })}];
  }
 }
 return new ConstructedHIR(arena,arenaIndex(arena,0,graph));
}
// HIRArena exposes canonical handles without allowing callers to mint an index.
function arenaIndex(arena: HIRArena,slot: number,graph: IdentityGraph): FunctionIndex { graph.function(graph.indices[slot] ?? panic('unknown replay function')); return arena.indexAt(slot); }
