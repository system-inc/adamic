// Go visitor.go: read-only place and structural-edge adapters for unit 3.
import { panic } from 'adamic';
import type { HIRFunction, PlaceInterface, PatternIndex, ValueType, TerminalType, BlockIndex } from '../../core.ts';
import type { EdgeType } from '../../../static_single_assignment/static_single_assignment.ts';
export function eachPattern(fn: HIRFunction,index: PatternIndex,visit: (place: PlaceInterface,define: boolean) => void): void {
 const p = fn.pattern(index);
 if(p.kind === 'Place') { visit(p.place,true); return; }
 if(p.kind === 'Object') { for(const item of p.properties) { if(item.computedKey !== undefined) { visit(item.computedKey,false); } if(item.defaultValue !== undefined) { visit(item.defaultValue,false); } eachPattern(fn,item.value,visit); } }
 else { for(const item of p.elements) { if(item.value === undefined) { continue; } if(item.defaultValue !== undefined) { visit(item.defaultValue,false); } eachPattern(fn,item.value,visit); } }
 if(p.rest !== undefined) { visit(p.rest,true); }
}
export function eachPlace(fn: HIRFunction,v: ValueType,visit: (place: PlaceInterface,define: boolean) => void): void {
 if(v.kind === 'LoadLocal' || v.kind === 'LoadContext') { visit(v.place,false); }
 if(v.kind === 'DeclareLocal' || v.kind === 'DeclareContext' || v.kind === 'StoreLocal' || v.kind === 'StoreContext' || v.kind === 'PrefixUpdate' || v.kind === 'PostfixUpdate') { visit(v.lvalue,true); }
 if(v.kind === 'StoreLocal' || v.kind === 'StoreContext' || v.kind === 'PrefixUpdate' || v.kind === 'PostfixUpdate' || v.kind === 'StoreGlobal' || v.kind === 'UnaryExpression' || v.kind === 'TypeCastExpression' || v.kind === 'Await' || v.kind === 'GetIterator' || v.kind === 'NextPropertyOf') { visit(v.value,false); }
 if(v.kind === 'BinaryExpression') { visit(v.left,false); visit(v.right,false); }
 if(v.kind === 'Destructure') { eachPattern(fn,v.lvaluePattern,visit); visit(v.value,false); }
 if(v.kind === 'StartMemoize' && v.deps !== undefined) { for(const dep of v.deps) { if(!dep.root.isGlobal) { visit(dep.root.place,false); } } }
 if(v.kind === 'FinishMemoize') { visit(v.value,false); }
 if(v.kind === 'PropertyLoad' || v.kind === 'ComputedLoad' || v.kind === 'PropertyStore' || v.kind === 'ComputedStore' || v.kind === 'PropertyDelete' || v.kind === 'ComputedDelete') { visit(v.object,false); }
 if(v.kind === 'ComputedLoad' || v.kind === 'ComputedStore' || v.kind === 'ComputedDelete') { visit(v.property,false); }
 if(v.kind === 'PropertyStore' || v.kind === 'ComputedStore') { visit(v.value,false); }
 if(v.kind === 'CallExpression' || v.kind === 'NewExpression') { visit(v.callee,false); }
 if(v.kind === 'MethodCall') { visit(v.receiver,false); visit(v.property,false); }
 if(v.kind === 'CallExpression' || v.kind === 'MethodCall' || v.kind === 'NewExpression') { for(const arg of v.args) { visit(arg.place,false); } }
 if(v.kind === 'ArrayExpression') { for(const e of v.elements) { if(!e.hole) { visit(e.place,false); } } }
 if(v.kind === 'ObjectExpression') { for(const e of v.properties) { if(e.computedKey !== undefined) { visit(e.computedKey,false); } visit(e.value,false); } }
 if(v.kind === 'IteratorNext') { visit(v.iterator,false); visit(v.collection,false); }
 if(v.kind === 'FunctionExpression') { for(const p of v.captures) { visit(p,false); } }
 if(v.kind === 'TemplateLiteral' || v.kind === 'TaggedTemplateExpression') { if(v.kind === 'TaggedTemplateExpression') { visit(v.tag,false); } for(const p of v.subexprs) { visit(p,false); } }
 if(v.kind === 'JsxExpression') { if(v.tag.place !== undefined) { visit(v.tag.place,false); } for(const prop of v.props) { visit(prop.value,false); } }
 if(v.kind === 'JsxExpression' || v.kind === 'JsxFragment') { for(const p of v.children) { visit(p,false); } }
}
export function eachTerminalPlace(t: TerminalType,visit: (place: PlaceInterface,define: boolean) => void): void {
 if(t.kind === 'Return' || t.kind === 'Throw') { visit(t.value ?? panic('missing return/throw value'),false); }
 if(t.kind === 'If' || t.kind === 'Branch' || t.kind === 'Switch') { visit(t.testPlace ?? panic('missing terminal test'),false); }
 if(t.kind === 'Switch' && t.cases !== undefined) { for(const c of t.cases) { if(c.test !== undefined) { visit(c.test,false); } } }
 if(t.kind === 'Try' && t.handlerBinding !== undefined) { visit(t.handlerBinding,true); }
}
export function eachEdge(t: TerminalType,visit: (id: BlockIndex,edge: EdgeType) => void): void {
 if(t.fallthrough !== undefined) { visit(t.fallthrough,'Fallthrough'); }
 if(t.kind === 'Goto') { visit(t.block ?? panic('missing goto'),'Real'); }
 else if(t.kind === 'If' || t.kind === 'Branch') { visit(t.consequent ?? panic('missing consequent'),'Real'); visit(t.alternate ?? panic('missing alternate'),'Real'); }
 else if(t.kind === 'Optional' || t.kind === 'Logical' || t.kind === 'Ternary' || t.kind === 'While') { visit(t.testBlock ?? panic('missing test block'),'Real'); }
 else if(t.kind === 'DoWhile') { visit(t.loop ?? panic('missing loop'),'Real'); }
 else if(t.kind === 'For' || t.kind === 'ForOf' || t.kind === 'ForIn') { visit(t.init ?? panic('missing init'),'Real'); }
 else if(t.kind === 'Label' || t.kind === 'Try') { visit(t.block ?? panic('missing body'),'Real'); if(t.kind === 'Try') { visit(t.handler ?? panic('missing handler'),'Real'); } }
 else if(t.kind === 'Switch') { const cases = t.cases; if(cases === undefined) { panic('missing switch cases'); } for(const c of cases) { visit(c.block,'Real'); } if(!cases.some((c) => c.test === undefined)) { visit(t.fallthrough ?? panic('missing switch exit'),'Real'); } }
}
