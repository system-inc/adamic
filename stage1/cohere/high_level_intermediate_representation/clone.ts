// Go clone.go: clone owned graph storage; immutable place records may be shared.
import { panic } from 'adamic';
import { HIRArena, ConstructedHIR, BasicBlock, Instruction, BlockIndex, InstructionIndex, IdentifierIndex, DeclarationIndex, PatternIndex } from './core.ts';
import type { HIRFunction, FunctionIndex, PlaceInterface, ValueType, TerminalType } from './core.ts';
function clonedPlace(fn: HIRFunction, place: PlaceInterface): PlaceInterface {
 return {identifier: fn.identifierAt(place.identifier.slot),effect: place.effect,reactive: place.reactive,start: place.start,end: place.end};
}
function pattern(src: HIRFunction, dst: HIRFunction, index: PatternIndex): PatternIndex {
 const value = src.pattern(index);
 if(value.kind === 'Place') { return dst.addPattern({kind: 'Place',place: clonedPlace(dst,value.place)}); }
 const rest = value.rest === undefined ? undefined : clonedPlace(dst,value.rest);
 if(value.kind === 'Object') { return dst.addPattern({kind: 'Object',rest,properties: value.properties.map((item) => ({key: item.key,computedKey: item.computedKey === undefined ? undefined : clonedPlace(dst,item.computedKey),defaultValue: item.defaultValue === undefined ? undefined : clonedPlace(dst,item.defaultValue),value: pattern(src,dst,item.value)}))}); }
 return dst.addPattern({kind: 'Array',rest,elements: value.elements.map((item) => ({value: item.value === undefined ? undefined : pattern(src,dst,item.value),defaultValue: item.defaultValue === undefined ? undefined : clonedPlace(dst,item.defaultValue)}))});
}
function value(src: HIRFunction, dst: HIRFunction, item: ValueType): ValueType {
 const p = (place: PlaceInterface): PlaceInterface => clonedPlace(dst,place);
 if(item.kind === 'Destructure') { return {kind: 'Destructure',pattern: pattern(src,dst,item.pattern),lvaluePattern: pattern(src,dst,item.lvaluePattern),value: p(item.value),declarationKind: item.declarationKind}; }
 if(item.kind === 'FunctionExpression' || item.kind === 'ObjectMethod') {
  const reference = {index: dst.functions[item.functionReference.ordinal] ?? panic('missing cloned function'),ordinal: item.functionReference.ordinal};
  if(item.kind === 'ObjectMethod') { return {...item,functionReference: reference}; }
  return {...item,functionReference: reference,captures: item.captures.map(p)};
 }
 if(item.kind === 'LoadLocal' || item.kind === 'LoadContext') { return {...item,place: p(item.place)}; }
 if(item.kind === 'DeclareLocal') { return {...item,lvalue: p(item.lvalue)}; }
 if(item.kind === 'StoreLocal' || item.kind === 'StoreContext' || item.kind === 'PrefixUpdate' || item.kind === 'PostfixUpdate') { return {...item,lvalue: p(item.lvalue),value: p(item.value)}; }
 if(item.kind === 'StoreGlobal' || item.kind === 'UnaryExpression' || item.kind === 'Await' || item.kind === 'TypeCastExpression' || item.kind === 'GetIterator' || item.kind === 'NextPropertyOf') { return {...item,value: p(item.value)}; }
 if(item.kind === 'BinaryExpression') { return {...item,left: p(item.left),right: p(item.right)}; }
 if(item.kind === 'PropertyLoad' || item.kind === 'PropertyDelete') { return {...item,object: p(item.object)}; }
 if(item.kind === 'ComputedLoad' || item.kind === 'ComputedDelete') { return {...item,object: p(item.object),property: p(item.property)}; }
 if(item.kind === 'PropertyStore') { return {...item,object: p(item.object),value: p(item.value)}; }
 if(item.kind === 'ComputedStore') { return {...item,object: p(item.object),property: p(item.property),value: p(item.value)}; }
 if(item.kind === 'CallExpression' || item.kind === 'NewExpression') { return {...item,callee: p(item.callee),args: item.args.map((arg) => ({place: p(arg.place),spread: arg.spread}))}; }
 if(item.kind === 'MethodCall') { return {...item,receiver: p(item.receiver),property: p(item.property),args: item.args.map((arg) => ({place: p(arg.place),spread: arg.spread}))}; }
 if(item.kind === 'IteratorNext') { return {...item,iterator: p(item.iterator),collection: p(item.collection)}; }
 if(item.kind === 'ArrayExpression') { return {...item,elements: item.elements.map((element) => ({place: p(element.place),spread: element.spread,hole: element.hole}))}; }
 if(item.kind === 'ObjectExpression') { return {...item,properties: item.properties.map((property) => ({key: property.key,computedKey: property.computedKey === undefined ? undefined : p(property.computedKey),value: p(property.value),spread: property.spread}))}; }
 if(item.kind === 'TemplateLiteral') { return {...item,quasis: [...item.quasis],subexprs: item.subexprs.map(p)}; }
 if(item.kind === 'TaggedTemplateExpression') { return {...item,tag: p(item.tag),quasis: [...item.quasis],subexprs: item.subexprs.map(p)}; }
 if(item.kind === 'JsxExpression') { return {...item,tag: {name: item.tag.name,place: item.tag.place === undefined ? undefined : p(item.tag.place)},props: item.props.map((prop) => ({name: prop.name,value: p(prop.value),spread: prop.spread})),children: item.children.map(p)}; }
 if(item.kind === 'JsxFragment') { return {...item,children: item.children.map(p)}; }
 return {...item};
}
function terminal(dst: HIRFunction, item: TerminalType): TerminalType {
 const b = (index: BlockIndex | undefined): BlockIndex | undefined => index === undefined ? undefined : dst.blockAt(index.slot + 1);
 return {...item,value: item.value === undefined ? undefined : clonedPlace(dst,item.value),testPlace: item.testPlace === undefined ? undefined : clonedPlace(dst,item.testPlace),handlerBinding: item.handlerBinding === undefined ? undefined : clonedPlace(dst,item.handlerBinding),block: b(item.block),testBlock: b(item.testBlock),consequent: b(item.consequent),alternate: b(item.alternate),fallthrough: b(item.fallthrough),loop: b(item.loop),init: b(item.init),update: b(item.update),handler: b(item.handler),cases: item.cases === undefined ? undefined : item.cases.map((clause) => ({test: clause.test === undefined ? undefined : clonedPlace(dst,clause.test),block: dst.blockAt(clause.block.slot + 1)}))};
}
function cloneFunction(source: HIRArena, target: HIRArena, index: FunctionIndex): FunctionIndex {
 const src = source.read(index); const result = target.create(src.name); const dst = target.read(result);
 dst.nodeIndex = src.nodeIndex; dst.isAsync = src.isAsync; dst.isGenerator = src.isGenerator;
 while(dst.identifierIndices.length < src.identifierIndices.length) { IdentifierIndex.push(dst.identifierIndices); }
 while(dst.declarationIndices.length < src.declarationIndices.length) { DeclarationIndex.push(dst.declarationIndices); }
 while(dst.identifiers.length > 0) { dst.identifiers.pop(); }
 for(const id of src.identifiers) { dst.identifiers.push({id: dst.identifierAt(id.id.slot),declaration: dst.declarationAt(id.declaration.slot + 1),name: id.name,nodeIndex: id.nodeIndex}); }
 while(dst.blockIndices.length < src.blockIndices.length) { BlockIndex.push(dst.blockIndices); }
 for(const child of src.functions) { dst.functions.push(cloneFunction(source,target,child)); }
 dst.returns = clonedPlace(dst,src.returns);
 for(const place of src.params) { dst.params.push(clonedPlace(dst,place)); }
 for(const place of src.context) { dst.context.push(clonedPlace(dst,place)); }
 for(const declaration of src.contextDeclarations) { dst.contextDeclarations.add(dst.declarationAt(declaration.slot + 1)); }
 for(const instruction of src.instructions) {
  const id = InstructionIndex.push(dst.instructionIndices); const copied = new Instruction(id,clonedPlace(dst,instruction.lvalue),value(src,dst,instruction.value),instruction.start,instruction.end); copied.order = instruction.order; dst.instructions.push(copied);
 }
 while(dst.blockTable.length > 0) { dst.blockTable.pop(); }
 for(const block of src.blockTable) {
  const copied = new BasicBlock(dst.blockIndices,dst.blockAt(block.id.slot + 1),terminal(dst,block.terminal),block.kind); copied.terminalOrder = block.terminalOrder;
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
