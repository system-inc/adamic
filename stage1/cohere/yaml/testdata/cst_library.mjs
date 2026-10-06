import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
const [library,path] = process.argv.slice(2);
const require = createRequire(join(library,'package.json'));
if(require('yaml/package.json').version !== '2.9.0') throw new Error('expected yaml 2.9.0');
const { Parser,LineCounter } = require('yaml');
const hex = text => { let result=''; for(let i=0;i<text.length;i++) result+=text.charCodeAt(i).toString(16).padStart(4,'0'); return result; };
const list = tokens => '['+(tokens ?? []).map(token=>outline(token)+',').join('')+']';
function outline(token) {
    if(token == null) return '-';
    const start = Array.isArray(token.start) ? token.start : [];
    const flowStart = token.start && !Array.isArray(token.start) ? token.start : undefined;
    const header = `${token.type}|${token.offset}|${'indent' in token ? token.indent : '-'}|${hex(token.source ?? '')}|${hex(token.message ?? '')}|`;
    const items = (token.items ?? []).map(item=>`${list(item.start)}|${'key' in item ? 1 : 0}|${outline(item.key)}|${'sep' in item ? list(item.sep) : '-'}|${outline(item.value)}|${item.explicitKey ? 1 : 0},`).join('');
    return header+`${list(start)}|${outline(flowStart)}|${outline(token.value)}|${'end' in token ? list(token.end) : '-'}|${list(token.props)}|[${items}]`;
}
const unescape = text=>text.replace(/\\([\\nrt])/g,(_,char)=>char==='n'?'\n':char==='r'?'\r':char==='t'?'\t':char);
let number=0;
for(const line of readFileSync(path,'utf8').split('\n')) {
    if(line==='') continue;
    const tab=line.indexOf('\t'); const size=Number(line.slice(0,tab)); const text=unescape(line.slice(tab+1));
    const lines=new LineCounter(); const parser=new Parser(lines.addNewLine);
    console.log(`case ${number++}`);
    function emit(source,incomplete) { for(const token of parser.parse(source,incomplete)) console.log(outline(token)); }
    if(size===0) emit(text,false);
    else { for(let start=0;start<text.length;start+=size) emit(text.slice(start,start+size),true); emit('',false); }
    console.log(`lines ${lines.lineStarts.join(',')}`);
}
