// Go clone.go: clone owned graph storage; immutable place records may be shared.
import { panic } from 'adamic';
import { HIRArena, ConstructedHIR, BasicBlock, Instruction, BlockIndex, InstructionIndex, IdentifierIndex, DeclarationIndex, PatternIndex } from './core.ts';
import type { HIRFunction, FunctionIndex, PlaceInterface, ValueType, TerminalType, FunctionReferenceInterface } from './core.ts';
function clonedPlace(fn: HIRFunction, place: PlaceInterface): PlaceInterface {
 return {identifier: fn.identifierAt(place.identifier.slot),effect: place.effect,reactive: place.reactive,start: place.start,end: place.end};
}
export function copyPattern(src: HIRFunction,dst: HIRFunction,index: PatternIndex): PatternIndex { return copyPatternWithRemap(src,dst,index,(place) => clonedPlace(dst,place)); }
export function copyPatternWithRemap(src: HIRFunction, dst: HIRFunction, index: PatternIndex,mapPlace: (place: PlaceInterface) => PlaceInterface): PatternIndex {
 const value = src.pattern(index);
 if(value.kind === 'Place') { return dst.addPattern({kind: 'Place',place: mapPlace(value.place)}); }
 const rest = value.rest === undefined ? undefined : mapPlace(value.rest);
 if(value.kind === 'Object') { return dst.addPattern({kind: 'Object',rest,properties: value.properties.map((item) => ({key: item.key,computedKey: item.computedKey === undefined ? undefined : mapPlace(item.computedKey),defaultValue: item.defaultValue === undefined ? undefined : mapPlace(item.defaultValue),value: copyPatternWithRemap(src,dst,item.value,mapPlace)}))}); }
 return dst.addPattern({kind: 'Array',rest,elements: value.elements.map((item) => ({value: item.value === undefined ? undefined : copyPatternWithRemap(src,dst,item.value,mapPlace),defaultValue: item.defaultValue === undefined ? undefined : mapPlace(item.defaultValue)}))});
}
// gap 2 (GAPS.md, @system_adamic ruling): explicit records replace object spreads.
export function copyInstructionValue(src: HIRFunction,dst: HIRFunction,item: ValueType): ValueType { return copyInstructionValueWithRemap(src,dst,item,(place) => clonedPlace(dst,place),(reference) => ({index: dst.functions[reference.ordinal] ?? panic('missing cloned function'),ordinal: reference.ordinal})); }
export function copyInstructionValueWithRemap(src: HIRFunction, dst: HIRFunction, item: ValueType,p: (place: PlaceInterface) => PlaceInterface,mapFunction: (reference: FunctionReferenceInterface) => FunctionReferenceInterface): ValueType {
 if(item.kind === 'Destructure') { return {kind: 'Destructure',pattern: copyPatternWithRemap(src,dst,item.pattern,p),lvaluePattern: copyPatternWithRemap(src,dst,item.lvaluePattern,p),value: p(item.value),declarationKind: item.declarationKind}; }
 if(item.kind === 'LoadLocal') { return {kind: 'LoadLocal',place: p(item.place)}; }
 if(item.kind === 'LoadContext') { return {kind: 'LoadContext',place: p(item.place)}; }
 if(item.kind === 'DeclareLocal') { return {kind: 'DeclareLocal',lvalue: p(item.lvalue),declarationKind: item.declarationKind}; }
 if(item.kind === 'DeclareContext') { return {kind: 'DeclareContext',lvalue: p(item.lvalue),declarationKind: item.declarationKind}; }
 if(item.kind === 'StoreLocal') { return {kind: 'StoreLocal',lvalue: p(item.lvalue),value: p(item.value),declarationKind: item.declarationKind}; }
 if(item.kind === 'StoreContext') { return {kind: 'StoreContext',lvalue: p(item.lvalue),value: p(item.value),declarationKind: item.declarationKind}; }
 if(item.kind === 'PrefixUpdate') { return {kind: 'PrefixUpdate',lvalue: p(item.lvalue),value: p(item.value),operation: item.operation}; }
 if(item.kind === 'PostfixUpdate') { return {kind: 'PostfixUpdate',lvalue: p(item.lvalue),value: p(item.value),operation: item.operation}; }
 if(item.kind === 'StoreGlobal') { return {kind: 'StoreGlobal',name: item.name,value: p(item.value)}; }
 if(item.kind === 'UnaryExpression') { return {kind: 'UnaryExpression',operator: item.operator,value: p(item.value)}; }
 if(item.kind === 'Await') { return {kind: 'Await',value: p(item.value)}; }
 if(item.kind === 'TypeCastExpression') { return {kind: 'TypeCastExpression',nodeKind: item.nodeKind,nodePos: item.nodePos,nodeEnd: item.nodeEnd,value: p(item.value)}; }
 if(item.kind === 'GetIterator') { return {kind: 'GetIterator',value: p(item.value)}; }
 if(item.kind === 'NextPropertyOf') { return {kind: 'NextPropertyOf',value: p(item.value)}; }
 if(item.kind === 'BinaryExpression') { return {kind: 'BinaryExpression',left: p(item.left),operator: item.operator,right: p(item.right)}; }
 if(item.kind === 'PropertyLoad') { return {kind: 'PropertyLoad',object: p(item.object),property: item.property,optional: item.optional}; }
 if(item.kind === 'PropertyDelete') { return {kind: 'PropertyDelete',object: p(item.object),property: item.property}; }
 if(item.kind === 'ComputedLoad') { return {kind: 'ComputedLoad',object: p(item.object),property: p(item.property),optional: item.optional}; }
 if(item.kind === 'ComputedDelete') { return {kind: 'ComputedDelete',object: p(item.object),property: p(item.property)}; }
 if(item.kind === 'PropertyStore') { return {kind: 'PropertyStore',object: p(item.object),property: item.property,value: p(item.value)}; }
 if(item.kind === 'ComputedStore') { return {kind: 'ComputedStore',object: p(item.object),property: p(item.property),value: p(item.value)}; }
 if(item.kind === 'CallExpression') { return {kind: 'CallExpression',callee: p(item.callee),args: item.args.map((arg) => ({place: p(arg.place),spread: arg.spread})),optional: item.optional,origin: item.origin}; }
 if(item.kind === 'NewExpression') { return {kind: 'NewExpression',callee: p(item.callee),args: item.args.map((arg) => ({place: p(arg.place),spread: arg.spread}))}; }
 if(item.kind === 'MethodCall') { return {kind: 'MethodCall',receiver: p(item.receiver),property: p(item.property),args: item.args.map((arg) => ({place: p(arg.place),spread: arg.spread})),optional: item.optional,origin: item.origin}; }
 if(item.kind === 'IteratorNext') { return {kind: 'IteratorNext',iterator: p(item.iterator),collection: p(item.collection)}; }
 if(item.kind === 'ArrayExpression') { return {kind: 'ArrayExpression',elements: item.elements.map((element) => ({place: p(element.place),spread: element.spread,hole: element.hole}))}; }
 if(item.kind === 'ObjectExpression') { return {kind: 'ObjectExpression',properties: item.properties.map((property) => ({key: property.key,computedKey: property.computedKey === undefined ? undefined : p(property.computedKey),value: p(property.value),spread: property.spread}))}; }
 if(item.kind === 'TemplateLiteral') { return {kind: 'TemplateLiteral',quasis: item.quasis.map((q) => q),subexprs: item.subexprs.map(p)}; }
 if(item.kind === 'TaggedTemplateExpression') { return {kind: 'TaggedTemplateExpression',tag: p(item.tag),quasis: item.quasis.map((q) => q),subexprs: item.subexprs.map(p)}; }
 if(item.kind === 'JsxExpression') { return {kind: 'JsxExpression',tag: {name: item.tag.name,place: item.tag.place === undefined ? undefined : p(item.tag.place)},props: item.props.map((prop) => ({name: prop.name,value: p(prop.value),spread: prop.spread})),children: item.children.map(p)}; }
 if(item.kind === 'JsxFragment') { return {kind: 'JsxFragment',children: item.children.map(p)}; }
 if(item.kind === 'Primitive') { return {kind: 'Primitive',literal: item.literal}; }
 if(item.kind === 'RegExpLiteral') { return {kind: 'RegExpLiteral',pattern: item.pattern,flags: item.flags}; }
 if(item.kind === 'JsxText') { return {kind: 'JsxText',text: item.text}; }
 if(item.kind === 'LoadGlobal') { return {kind: 'LoadGlobal',name: item.name,bindingKind: item.bindingKind,source: item.source,imported: item.imported}; }
 if(item.kind === 'MetaProperty') { return {kind: 'MetaProperty',meta: item.meta,property: item.property}; }
 if(item.kind === 'UnsupportedNode') { return {kind: 'UnsupportedNode',nodeKind: item.nodeKind,nodePos: item.nodePos,nodeEnd: item.nodeEnd,reason: item.reason}; }
 if(item.kind === 'StartMemoize') { return {kind: 'StartMemoize',manualMemoId: item.manualMemoId,deps: item.deps === undefined ? undefined : item.deps.map((dep) => ({root: {isGlobal: dep.root.isGlobal,name: dep.root.name,place: p(dep.root.place)},path: dep.path.map((entry) => ({property: entry.property,optional: entry.optional}))}))}; }
 if(item.kind === 'FinishMemoize') { return {kind: 'FinishMemoize',manualMemoId: item.manualMemoId,value: p(item.value),pruned: item.pruned}; }
 if(item.kind === 'FunctionExpression' || item.kind === 'ObjectMethod') {
  const reference = mapFunction(item.functionReference);
  if(item.kind === 'ObjectMethod') { return {kind: 'ObjectMethod',key: item.key,functionReference: reference}; }
  return {kind: 'FunctionExpression',functionReference: reference,captures: item.captures.map(p)};
 }
 return {kind: 'Debugger'};
}
// gap 2 (GAPS.md): copy terminal fields explicitly until static-constructor spread lowers.
export function copyTerminal(dst: HIRFunction,item: TerminalType): TerminalType { return copyTerminalWithRemap(item,(place) => clonedPlace(dst,place),(block) => dst.blockAt(block.slot + 1)); }
export function copyTerminalWithRemap(item: TerminalType,p: (place: PlaceInterface) => PlaceInterface,mapBlock: (block: BlockIndex) => BlockIndex): TerminalType {
 const b = (index: BlockIndex | undefined): BlockIndex | undefined => index === undefined ? undefined : mapBlock(index);
 return {kind: item.kind,variant: item.variant,operator: item.operator,optionalFlag: item.optionalFlag,value: item.value === undefined ? undefined : p(item.value),testPlace: item.testPlace === undefined ? undefined : p(item.testPlace),handlerBinding: item.handlerBinding === undefined ? undefined : p(item.handlerBinding),block: b(item.block),testBlock: b(item.testBlock),consequent: b(item.consequent),alternate: b(item.alternate),fallthrough: b(item.fallthrough),loop: b(item.loop),init: b(item.init),update: b(item.update),handler: b(item.handler),cases: item.cases === undefined ? undefined : item.cases.map((clause) => ({test: clause.test === undefined ? undefined : p(clause.test),block: mapBlock(clause.block)}))};
}
function cloneFunction(source: HIRArena, target: HIRArena, index: FunctionIndex): FunctionIndex {
 const src = source.read(index); const result = target.create(src.name); const dst = target.read(result);
 dst.source = src.source; dst.nodeIndex = src.nodeIndex; dst.isAsync = src.isAsync; dst.isGenerator = src.isGenerator;
 while(dst.identifierIndices.length < src.identifierIndices.length) { IdentifierIndex.push(dst.identifierIndices); }
 while(dst.declarationIndices.length < src.declarationIndices.length) { DeclarationIndex.push(dst.declarationIndices); }
 while(dst.identifiers.length > 0) { dst.identifiers.pop(); }
 for(const id of src.identifiers) { dst.identifiers.push({id: dst.identifierAt(id.id.slot),declaration: dst.declarationAt(id.declaration.slot + 1),name: id.name,nodeIndex: id.nodeIndex,source: id.source}); }
 while(dst.blockIndices.length < src.blockIndices.length) { BlockIndex.push(dst.blockIndices); }
 for(const child of src.functions) { dst.functions.push(cloneFunction(source,target,child)); }
 if(src.outlined !== undefined) { const outlined = new Map<IdentifierIndex,FunctionIndex>(); for(const key of src.outlined.keys()) { const old = src.outlined.get(key) ?? panic('missing outlined target'); const ordinal = src.functions.indexOf(old); outlined.set(dst.identifierAt(key.slot),dst.functions[ordinal] ?? panic('outlined function absent')); } dst.outlined = outlined; }
 dst.returns = clonedPlace(dst,src.returns);
 for(const place of src.params) { dst.params.push(clonedPlace(dst,place)); }
 for(const place of src.context) { dst.context.push(clonedPlace(dst,place)); }
 for(const declaration of src.contextDeclarations) { dst.contextDeclarations.add(dst.declarationAt(declaration.slot + 1)); }
 for(const instruction of src.instructions) {
  const id = InstructionIndex.push(dst.instructionIndices); const copied = new Instruction(id,clonedPlace(dst,instruction.lvalue),copyInstructionValue(src,dst,instruction.value),instruction.start,instruction.end); copied.order = instruction.order; copied.source = instruction.source; dst.instructions.push(copied);
 }
 while(dst.blockTable.length > 0) { dst.blockTable.pop(); }
 for(const block of src.blockTable) {
  const copied = new BasicBlock(dst.blockIndices,dst.blockAt(block.id.slot + 1),copyTerminal(dst,block.terminal),block.kind); copied.terminalOrder = block.terminalOrder; copied.present = block.present; copied.terminalPresent = block.terminalPresent;
  for(const id of block.instructions) { copied.instructions.push(dst.instructionIndices[id.slot] ?? panic('missing cloned instruction')); }
  copied.predecessors = block.predecessors.map((id) => dst.blockAt(id.slot + 1));
  copied.phis = block.phis.map((phi) => ({place: clonedPlace(dst,phi.place),operands: phi.operands.map((entry) => ({predecessor: entry.predecessor,place: clonedPlace(dst,entry.place)}))})); dst.blockTable.push(copied);
 }
 dst.blockOrder = src.blockOrder.map((id) => dst.blockAt(id.slot + 1)); dst.retained.clear(); for(const id of src.retained) { dst.retained.add(id); }
 return result;
}
export function CloneFunction(graph: ConstructedHIR): ConstructedHIR {
 const arena = new HIRArena(); return new ConstructedHIR(arena,cloneFunction(graph.arena,arena,graph.root));
}
