// A checked numeric arena for JSON payload syntax; never recursive strong references.
import { panic } from 'adamic';
import { PayloadIndex } from '../../arena/arena_index.a';
export type PayloadNode = { readonly kind: 'object'; readonly fields: Map<string,PayloadIndex> }
 | { readonly kind: 'array'; readonly elements: PayloadIndex[] }
 | { readonly kind: 'string'; readonly text: string }
 | { readonly kind: 'number'; readonly value: number }
 | { readonly kind: 'boolean'; readonly value: boolean }
 | { readonly kind: 'null' };
export class Payload {
 readonly indices: PayloadIndex[] = []; readonly nodes: PayloadNode[] = [];
 readonly text: string; at = 0; private rootIndex: PayloadIndex | undefined = undefined;
 rootHandle(): PayloadIndex { return this.rootIndex ?? panic("payload root not initialized"); }
 constructor(text: string) { this.text = text; this.rootIndex = this.parse(); this.space(); if(this.at !== text.length) { panic('extra JSON payload'); } }
 read(index: PayloadIndex): PayloadNode { return this.nodes[PayloadIndex.read(this.indices,index)] ?? panic('missing payload node'); }
 add(value: PayloadNode): PayloadIndex { const index = PayloadIndex.push(this.indices); this.nodes.push(value); return index; }
 space(): void { while(' \n\r\t'.includes(this.text.charAt(this.at)) && this.at < this.text.length) { this.at++; } }
 take(text: string): void { if(this.text.slice(this.at,this.at + text.length) !== text) { panic('invalid JSON token'); } this.at += text.length; }
 string(): string {
  this.take('"'); let result = '';
  while(this.at < this.text.length) {
   const ch = this.text.charAt(this.at); this.at++;
   if(ch === '"') { return result; }
   if(ch !== '\\') { if(ch.charCodeAt(0) < 32) { panic('control in JSON string'); } result += ch; continue; }
   const code = this.text.charAt(this.at); this.at++;
   if(code === '"' || code === '\\' || code === '/') { result += code; }
   else if(code === 'n') { result += '\n'; } else if(code === 'r') { result += '\r'; } else if(code === 't') { result += '\t'; }
   else if(code === 'b') { result += '\b'; } else if(code === 'f') { result += '\f'; }
   else if(code === 'u') { const digits = this.text.slice(this.at,this.at + 4); if(digits.length !== 4 || !/^[0-9a-fA-F]{4}$/.test(digits)) { panic('invalid JSON escape'); } result += String.fromCharCode(Number.parseInt(digits,16)); this.at += 4; }
   else { panic('invalid JSON escape'); }
  }
  return panic('unterminated JSON string');
 }
 parse(): PayloadIndex {
  this.space(); const ch = this.text.charAt(this.at);
  if(ch === '"') { return this.add({kind: 'string',text: this.string()}); }
  if(ch === '{') {
   this.at++; this.space(); const fields = new Map<string,PayloadIndex>();
   if(this.text.charAt(this.at) !== '}') { while(true) { this.space(); const name = this.string(); this.space(); this.take(':'); const value = this.parse(); if(fields.has(name)) { panic('duplicate JSON field'); } fields.set(name,value); this.space(); if(this.text.charAt(this.at) !== ',') { break; } this.at++; } }
   this.take('}'); return this.add({kind: 'object',fields});
  }
  if(ch === '[') {
   this.at++; this.space(); const elements: PayloadIndex[] = [];
   if(this.text.charAt(this.at) !== ']') { while(true) { elements.push(this.parse()); this.space(); if(this.text.charAt(this.at) !== ',') { break; } this.at++; } }
   this.take(']'); return this.add({kind: 'array',elements});
  }
  if(ch === 'n') { this.take('null'); return this.add({kind: 'null'}); }
  if(ch === 't') { this.take('true'); return this.add({kind: 'boolean',value: true}); }
  if(ch === 'f') { this.take('false'); return this.add({kind: 'boolean',value: false}); }
  const start = this.at;
  while(this.at < this.text.length && '-+0123456789.eE'.includes(this.text.charAt(this.at))) { this.at++; }
  const number = this.text.slice(start,this.at);
  if(number === '' || !/^-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$/.test(number)) { return panic('invalid JSON number'); }
  return this.add({kind: 'number',value: Number(number)});
 }
 field(index: PayloadIndex,name: string): PayloadIndex { const node = this.read(index); if(node.kind !== 'object') { return panic('expected payload object'); } return node.fields.get(name) ?? panic(`missing payload field ${name}`); }
 textAt(index: PayloadIndex): string { const node = this.read(index); if(node.kind !== 'string') { return panic('expected payload string'); } return node.text; }
 numberAt(index: PayloadIndex): number { const node = this.read(index); if(node.kind !== 'number') { return panic('expected payload number'); } return node.value; }
 booleanAt(index: PayloadIndex): boolean { const node = this.read(index); if(node.kind !== 'boolean') { return panic('expected payload boolean'); } return node.value; }
 array(index: PayloadIndex): readonly PayloadIndex[] { const node = this.read(index); if(node.kind !== 'array') { return panic('expected payload array'); } return node.elements; }
 stringField(index: PayloadIndex,name: string): string { return this.textAt(this.field(index,name)); }
 numberField(index: PayloadIndex,name: string): number { return this.numberAt(this.field(index,name)); }
 booleanField(index: PayloadIndex,name: string): boolean { return this.booleanAt(this.field(index,name)); }
}
