// No SSA algorithms live here: adapt HIR records to the existing imported module.
import { panic } from 'adamic';
import { reversePostorder, markPredecessors, markEvaluationOrder } from '../../../static_single_assignment/graph.ts';
import type { GraphInterface } from '../../../static_single_assignment/static_single_assignment.ts';
import { HIRFunction, BasicBlock, BlockIndex } from '../../core.ts';
import type { PlaceInterface } from '../../core.ts';
import { eachEdge } from './places.ts';
function graphFor(owner: HIRFunction): GraphInterface<HIRFunction,BasicBlock,PlaceInterface> { return {
 entry: (fn) => BlockIndex.read(fn.blockIndices,fn.entry) + 1,
 blockBound: (fn) => fn.nextBlock,
 block: (fn,id) => id > 0 && id < fn.nextBlock ? fn.blockOrUndefined(fn.blockAt(id)) : undefined,
 blocks: (fn) => fn.blockOrder.map((id) => fn.block(id)),
 setBlocks: (fn,blocks) => { fn.blockOrder = blocks.map((b) => b.id); },
 retain: (fn,keep) => { for(const b of fn.blockTable) { if(!keep(b.id.slot + 1)) { b.present = false; fn.retained.delete(b.id.slot + 1); } } },
 placeholder: (fn,block) => { const b = new BasicBlock(fn.blockIndices,block.id,{kind: 'Unreachable'},block.kind); b.predecessors = [...block.predecessors]; fn.blockTable[block.id.slot] = b; return b; },
 id: (b) => BlockIndex.read(b.arenaIndices,b.id) + 1,
 predecessors: (b) => b.predecessors.map((id) => BlockIndex.read(b.arenaIndices,id) + 1),
 setPredecessors: (b: BasicBlock,ids: readonly number[]) => { b.predecessors = ids.map((id) => { const index = b.arenaIndices[id - 1] ?? panic('unknown predecessor'); BlockIndex.read(b.arenaIndices,index); return index; }); },
 phis: (b) => b.phis, setPhis: (b,phis) => { b.phis = phis; },
 eachEdge: (b,visit) => { if(b.terminalPresent) { eachEdge(b.terminal,(id,kind) => visit(BlockIndex.read(b.arenaIndices,id) + 1,kind)); } },
 endsInReturn: (b) => b.terminal.kind === 'Return',
 instructionCount: (_fn,b) => b.instructions.length,
 // The three graph-maintenance algorithms do not rename places or construct SSA.
 eachInstructionPlace: () => { panic('unit-3 maintenance adapter cannot construct SSA'); },
 eachTerminalPlace: () => { panic('unit-3 maintenance adapter cannot construct SSA'); },
 isContextStore: () => false, contextStoreDefines: () => false,
 setInstructionOrder: (fn,b,i,order) => { fn.instruction(b.instructions[i] ?? panic('missing instruction')).order = order; },
 setTerminalOrder: (b,order) => { if(b.terminalPresent) { b.terminalOrder = order; } },
 params: (fn) => fn.params, returns: (fn) => fn.returns, setReturns: (fn,p) => { fn.returns = p; },
 declaration: (fn,id) => fn.identifier(fn.identifierAt(id)).declaration.slot + 1,
 contextual: (fn,d) => fn.contextDeclarations.has(fn.declarationAt(d)),
 mint: (fn,id) => fn.mint(fn.identifierAt(id)).slot,
 named: (fn,id) => fn.identifier(fn.identifierAt(id)).name !== '',
 placeString: (_fn,id) => `$${id}`, identifierOf: (p) => owner.identifier(p.identifier).id.slot,
 withIdentifier: (p,id) => ({identifier: owner.identifierAt(id),effect: p.effect,reactive: p.reactive,start: p.start,end: p.end}),
}; }
export function normalizeGraph(fn: HIRFunction): void { const graph = graphFor(fn); reversePostorder(graph,fn); markPredecessors(graph,fn); markEvaluationOrder(graph,fn); }
export function refreshPredecessors(fn: HIRFunction): void { markPredecessors(graphFor(fn),fn); }
export function refreshOrder(fn: HIRFunction): void { markEvaluationOrder(graphFor(fn),fn); }
