// Go merge_consecutive_blocks.go:62. Only block lookup entries are deleted.
import { HIRFunction } from '../../core.ts';
import type { BlockIndex } from '../../core.ts';
import { copyTerminalWithRemap } from '../../replay/index.ts';
import { refreshPredecessors } from './graph_adapter.ts';
export function mergeConsecutiveBlocks(fn: HIRFunction): number {
 const mergedInto = new Map<BlockIndex,BlockIndex>(); const fallthroughs = new Set<BlockIndex>(); const survivors: BlockIndex[] = []; let count = 0;
 const resolve = (id: BlockIndex): BlockIndex => resolveBlock(fn,mergedInto,id);
 for(const bid of fn.blockOrder) {
  const block = fn.block(bid); if(block.terminalPresent && block.terminal.fallthrough !== undefined) { fallthroughs.add(block.terminal.fallthrough); }
  const pred = block.predecessors[0];
  if(block.predecessors.length !== 1 || block.kind !== 'block' || fallthroughs.has(block.id) || pred === undefined) { survivors.push(bid); continue; }
  const predecessor = fn.blockOrUndefined(resolve(pred));
  if(predecessor === undefined || !predecessor.terminalPresent || predecessor.terminal.kind !== 'Goto' || predecessor.kind !== 'block') { survivors.push(bid); continue; }
  for(const phi of block.phis) { const op = phi.operands[0]; if(phi.operands.length !== 1 || op === undefined) { continue; } predecessor.instructions.push(fn.emit(phi.place,{kind: 'LoadLocal',place: op.place},0,0)); }
  for(const id of block.instructions) { predecessor.instructions.push(id); }
  predecessor.terminal = block.terminal; predecessor.terminalPresent = block.terminalPresent; predecessor.terminalOrder = block.terminalOrder;
  mergedInto.set(block.id,predecessor.id); block.present = false; fn.retained.delete(block.id.slot + 1); count++;
 }
 if(count === 0) { return 0; }
 fn.blockOrder = survivors;
 for(const bid of survivors) {
  const block = fn.block(bid);
  for(const phi of block.phis) {
   const old = [...phi.operands];
   for(const entry of old) { const mapped = resolve(fn.blockAt(entry.predecessor)).slot + 1; if(mapped === entry.predecessor) { continue; }
    phi.operands = phi.operands.filter((op) => op.predecessor !== entry.predecessor && op.predecessor !== mapped);
    phi.operands.push({predecessor: mapped,place: entry.place}); phi.operands.sort((a,b) => a.predecessor - b.predecessor);
   }
  }
  if(block.terminalPresent) { block.terminal = copyTerminalWithRemap(block.terminal,(p) => p,resolve); }
 }
 fn.entry = resolve(fn.entry); refreshPredecessors(fn); return count;
}

function resolveBlock(fn: HIRFunction, mergedInto: Map<BlockIndex,BlockIndex>, id: BlockIndex): BlockIndex {
 let slot = id.slot;
 while(true) { const current = fn.blockAt(slot + 1); const next = mergedInto.get(current); if(next === undefined || next.slot === slot) { return fn.blockAt(slot + 1); } slot = next.slot; }
}
