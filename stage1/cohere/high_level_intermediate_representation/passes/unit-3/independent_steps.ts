// Subpasses independent of the refused graph-maintenance callback.
import { panic } from 'adamic';
import { InputFacts, readPlace, writeCheckpoint, encodeCheckpoint } from '../../replay/index.ts';
import type { Checkpoint, SidecarRow } from '../../replay/index.ts';
import type { ConstructedHIR, HIRFunction, PlaceInterface } from '../../core.ts';
import { outlineFunctions } from './outline_functions.ts';
import { eliminateDeadCode } from './dead_code_elimination.ts';
import { collectAssumedInvokedFunctions } from './invoked_functions.ts';
import { copyNestedBodyInto } from './inline_remap.ts';
import { Unit3Facts } from './facts.ts';
export function independentResult(graph: ConstructedHIR,before: Checkpoint,facts: Unit3Facts): string | undefined {
 const fn = graph.arena.read(graph.root); const pass = before.pass.split(':')[0] ?? panic('missing pass name'); let result: string;
 if(pass === 'outline_functions') { result = String(outlineFunctions(graph.arena,graph.root)); }
 else if(pass === 'dead_code_elimination') { const r = eliminateDeadCode(fn); result = `${r.instructions},${r.phis}`; }
 else if(pass === 'invoked_functions') { result = [...collectAssumedInvokedFunctions(graph.arena,fn,facts).keys()].sort((a,b) => a - b).join(','); }
 else if(pass === 'inline_remap') {
  const input = new InputFacts(before); const ordinal = input.require('unit3.input','$','function',undefined,'nested');
  let nested: HIRFunction | undefined = undefined;
  if(ordinal !== 'absent') { nested = graph.arena.read(fn.functions[Number(ordinal)] ?? panic('missing nested input')); }
  const captures: PlaceInterface[] = input.require('unit3.input','$','function',undefined,'captures').split(' ').filter((p) => p !== '').map((p) => readPlace(fn,p));
  const remap = nested === undefined ? undefined : copyNestedBodyInto(fn,nested,captures);
  if(remap === undefined) { result = 'false'; }
  else {
   result = `true\nentry ${remap.entry.slot + 1}`;
   for(const key of [...remap.identifiers.keys()].sort((a,b) => a.slot - b.slot)) { result += `\nidentifier ${key.slot}=${remap.identifiers.get(key)?.slot}`; }
   for(const key of [...remap.blocks.keys()].sort((a,b) => a.slot - b.slot)) { const mapped = remap.blocks.get(key) ?? panic('missing remapped block'); result += `\nblock ${key.slot + 1}=${mapped.slot + 1}`; }
   for(const key of [...remap.instructions.keys()].sort((a,b) => a.slot - b.slot)) { result += `\ninstruction ${key.slot}=${remap.instructions.get(key)?.slot}`; }
  }
 } else { return undefined; }
 return result;
}
export function encodeResult(graph: ConstructedHIR,before: Checkpoint,facts: Unit3Facts,result: string): string {
 const pass = before.pass.split(':')[0] ?? panic('missing pass name');
 const rows: SidecarRow[] = facts.rows(graph);
 for(const row of before.sidecars) { if(row.namespace === 'unit3.input') { rows.push(row); } }
 rows.push({namespace: 'unit3.result',functionPath: '$',anchorKind: 'function',anchorId: undefined,key: 'result',payload: result});
 return writeCheckpoint(encodeCheckpoint(graph,before,before.pass,rows,[]));
}