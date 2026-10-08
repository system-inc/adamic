// Lane-owned syntax facts: cooked identifier text is input, not hook classification.
import { panic } from 'adamic';
import { Checkpoint } from '../../replay/index.ts';
import type { SidecarRow } from '../../replay/index.ts';
import type { ConstructedHIR, HIRFunction, IdentifierIndex, FunctionIndex, SourceHandleInterface } from '../../core.ts';
function sourceKey(s: SourceHandleInterface): string { return `${s.kind}:${s.start}:${s.end}`; }
export class Unit3Facts {
 private readonly texts = new Map<string,string>();
 constructor(checkpoint: Checkpoint) {
  for(const row of checkpoint.sidecars) {
   if(row.namespace !== 'unit3.ast' || row.anchorId === undefined) { continue; }
   const ast = checkpoint.sidecars.find((r) => r.namespace === 'input.ast' && r.functionPath === row.functionPath && r.anchorKind === 'identifier' && r.anchorId === row.anchorId && r.key === 'node');
   if(ast === undefined || ast.payload === 'nil') { panic('cooked identifier without source handle'); }
   const fields = ast.payload.split('\t'); this.texts.set(`${fields[0]}:${fields[1]}:${fields[2]}`,row.payload);
  }
 }
 hook(fn: HIRFunction,id: IdentifierIndex): boolean {
  const s = fn.identifier(id).source; if(s === undefined || s.kind !== 'KindIdentifier') { return false; }
  const text = this.texts.get(sourceKey(s)) ?? panic('missing cooked identifier input'); const next = text.charCodeAt(3);
  return text.startsWith('use') && text.length >= 4 && ((next >= 65 && next <= 90) || (next >= 48 && next <= 57));
 }
 rows(graph: ConstructedHIR): SidecarRow[] {
  const rows: SidecarRow[] = [];
  visitFacts(graph,this.texts,rows,graph.root,'$'); return rows;
 }
}
// A plain module function carries the former closure's captured state explicitly.
function visitFacts(graph: ConstructedHIR,texts: Map<string,string>,rows: SidecarRow[],index: FunctionIndex,path: string): void {
   const fn = graph.arena.read(index);
   for(const identifier of fn.identifiers) { const s = identifier.source; if(s === undefined || s.kind !== 'KindIdentifier') { continue; } rows.push({namespace: 'unit3.ast',functionPath: path,anchorKind: 'identifier',anchorId: identifier.id.slot,key: 'text',payload: texts.get(sourceKey(s)) ?? panic('missing source text for copied identifier')}); }
   for(let i = 0; i < fn.functions.length; i++) { visitFacts(graph,texts,rows,fn.functions[i] ?? panic('missing child'),`${path}/${i}`); }
}
