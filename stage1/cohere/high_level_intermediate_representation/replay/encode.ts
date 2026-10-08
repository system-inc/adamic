import { panic } from 'adamic';
import type { FunctionIndex } from '../../arena/arena_index.a';
import type { ConstructedHIR, SourceHandleInterface } from '../core.ts';
import { dump, placeText, terminalText } from '../dump.ts';
import { Checkpoint, readGraph } from './checkpoint.ts';
import type { SidecarRow, ExtraIdentity, AnchorKind } from './checkpoint.ts';
import { InputFacts } from './facts.ts';
function sourceKey(source: SourceHandleInterface): string { return `${source.kind}:${source.start}:${source.end}`; }
interface TypeInput { readonly valuePresent: string | undefined; readonly flags: string | undefined; readonly aliasPresent: string; readonly typePresent: string; readonly alias: string; readonly symbol: string; }
// Pass outputs supply their own sidecars. Shared input facts and allocator state
// are rebuilt from the changed arena; they are never copied as expected answers.
export function encodeCheckpoint(graph: ConstructedHIR,base: Checkpoint,pass: string,sidecars: readonly SidecarRow[],identities: readonly ExtraIdentity[]): Checkpoint {
 const output = new Checkpoint(base.key,pass,readGraph(dump(graph))); const facts = new InputFacts(base);
 const hasValueFacts = base.sidecars.some((row) => row.namespace === 'input.types' && row.key === 'value-present');
 const checker = facts.checkerAvailable('$'); const types = new Map<string,TypeInput>();
 for(const fn of base.graph.functions) {
  for(const id of fn.identifierIds) {
   const source = facts.node(fn.path,'identifier',id);
   if(source !== undefined) { types.set(sourceKey(source),{valuePresent: facts.optional('input.types',fn.path,'identifier',id,'value-present'),flags: facts.optional('input.types',fn.path,'identifier',id,'flags'),aliasPresent: facts.require('input.types',fn.path,'identifier',id,'alias-present'),typePresent: facts.require('input.types',fn.path,'identifier',id,'type-present'),alias: facts.require('input.types',fn.path,'identifier',id,'alias-symbol'),symbol: facts.require('input.types',fn.path,'identifier',id,'type-symbol')}); }
  }
 }
 const add = (namespace: string,path: string,kind: AnchorKind,id: number | undefined,key: string,payload: string): void => output.addSidecar({namespace,functionPath: path,anchorKind: kind,anchorId: id,key,payload});
 const node = (path: string,kind: AnchorKind,id: number | undefined,source: SourceHandleInterface | undefined): void => add('input.ast',path,kind,id,'node',source === undefined ? 'nil' : `${source.kind}\t${source.start}\t${source.end}`);
 const ordered: FunctionIndex[] = []; const pending: FunctionIndex[] = [graph.root];
 while(pending.length > 0) { const index = pending.pop() ?? panic('missing function index'); ordered.push(index); const fn = graph.arena.read(index); for(let i = fn.functions.length - 1; i >= 0; i--) { pending.push(fn.functions[i] ?? panic('missing nested function')); } }
 for(let slot = 0; slot < output.graph.functions.length; slot++) {
  const record = output.graph.functions[slot] ?? panic('missing function'); const fn = graph.arena.read(ordered[slot] ?? panic('missing arena function'));
  for(const id of fn.scopeIds) { output.addIdentity({functionPath: record.path,kind: 'scope',id}); }
  for(let id = 1; id < fn.nextBlock; id++) { if(!record.blockMap.has(id)) { output.addIdentity({functionPath: record.path,kind: 'block',id}); } }
  add('input.checker',record.path,'function',undefined,'available',checker ? '1' : '0'); node(record.path,'function',undefined,fn.source);
  add('input.identity',record.path,'function',undefined,'next-block',String(fn.nextBlock));
  for(const identifier of fn.identifiers) {
   const id = identifier.id.slot; node(record.path,'identifier',id,identifier.source);
   let type: TypeInput = {valuePresent: hasValueFacts ? '0' : undefined,flags: hasValueFacts ? '0' : undefined,aliasPresent: '0',typePresent: '0',alias: '',symbol: ''};
   const source = identifier.source;
   if(checker && source !== undefined) { type = types.get(sourceKey(source)) ?? panic('new AST source requires an actual checker fact'); }
   if(type.valuePresent !== undefined) { add('input.types',record.path,'identifier',id,'value-present',type.valuePresent); }
   if(type.flags !== undefined) { add('input.types',record.path,'identifier',id,'flags',type.flags); }
   add('input.types',record.path,'identifier',id,'alias-present',type.aliasPresent); add('input.types',record.path,'identifier',id,'type-present',type.typePresent);
   add('input.types',record.path,'identifier',id,'alias-symbol',type.alias); add('input.types',record.path,'identifier',id,'type-symbol',type.symbol);
  }
  for(const instruction of fn.instructions) { const id = instruction.id.slot; const value = instruction.value; node(record.path,'instruction',id,instruction.source); if(value.kind === 'Destructure') { add('input.pattern',record.path,'instruction',id,'lvalue-shares-pattern',value.pattern === value.lvaluePattern ? '1' : '0'); } }
  for(const block of fn.blockTable) {
   if(fn.retained.has(block.id.slot + 1) || !block.present) { continue; }
   let body = `block ${block.id.slot + 1} ${block.kind} predecessors=${block.predecessors.map((id) => id.slot + 1).join(',')}\n`;
   for(const phi of block.phis) { body += `phi ${placeText(phi.place)}`; for(const operand of phi.operands) { body += ` ${operand.predecessor}=${placeText(operand.place)}`; } body += '\n'; }
   for(const id of block.instructions) { body += `instruction-ref ${id.slot}\n`; }
   body += block.terminalPresent ? `terminal ${block.terminalOrder} ${block.terminal.kind} ${terminalText(block.terminal,fn)}\n` : 'terminal-null\n';
   add('input.blocks',record.path,'function',undefined,String(block.id.slot + 1),body);
  }
 }
 for(const row of base.select('input.source','$')) { output.addSidecar(row); }
 for(const identity of identities) { if(identity.kind === 'scope' && output.graph.at(identity.functionPath).scopeMap.has(identity.id)) { continue; } output.addIdentity(identity); }
 for(const row of sidecars) { output.addSidecar(row); }
 return output;
}
