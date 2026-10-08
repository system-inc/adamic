import fs from 'node:fs';
import path from 'node:path';
const root=path.dirname(new URL(import.meta.url).pathname);
const fixtures=JSON.parse(fs.readFileSync(path.join(root,'fixtures.json'),'utf8'));
const helper=`export function observe(regex: RegExp, text: string, input: number): void {
 regex.lastIndex = 0;
 let previousEnd = -1;
 let match = regex.exec(text);
 while(match !== null) {
  const value = match[0] ?? '';
  const start = match.index;
  const end = start + value.length;
  if(value.length > 0 || start !== previousEnd) {
   let units = '';
   for(let i = 0; i < value.length; i++) units += \`\${value.charCodeAt(i)},\`;
   console.log(\`\${input}:\${start}:\${end}:\${units}\`);
   previousEnd = end;
  }
  if(value.length === 0) {
   if(regex.lastIndex >= text.length) break;
   const point = text.codePointAt(regex.lastIndex) ?? 0;
   regex.lastIndex += point > 65535 ? 2 : 1;
  }
  match = regex.exec(text);
 }
}
`;
fs.writeFileSync(path.join(root,'observe.a'),helper);
for(let index=0;index<fixtures.length;index++) {
 const f=fixtures[index];
 f.native_expectation='required sanitized native after area/library merge';
 f.native_when_accepted='required identical matches and spans';
 // All patterns and flags are strings; no literal or RegExp-object constructor argument.
 for(const sample of f.samples) {
  const regex=new RegExp(f.pattern,f.flags);
  let previousEnd=-1;
  const matches=[];
  for(let match=regex.exec(sample.input);match!==null;match=regex.exec(sample.input)) {
   const value=match[0],start=match.index,end=start+value.length;
   if(value.length>0 || start!==previousEnd){matches.push({start,end,text:value});previousEnd=end;}
   if(value.length===0){if(regex.lastIndex>=sample.input.length)break;regex.lastIndex+=(sample.input.codePointAt(regex.lastIndex)>65535?2:1);}
  }
  sample.node_matches=matches;
  if(JSON.stringify(matches)!==JSON.stringify(sample.go_matches))throw new Error(`${f.id}: Go/Node spans differ on ${JSON.stringify(sample.input)}`);
 }
 const shape=f.shape==='dynamic'?'dynamic':f.shape==='plain'?'plain':f.shape.slice(2,3);
 const directory=path.join(root,shape);fs.mkdirSync(directory,{recursive:true});
 f.file=`${shape}/${String(index).padStart(3,'0')}.a`;
 const prefix=`// ${f.id}; ${f.expression}\nimport { observe } from '../observe.a';\n`;
 const samples=f.samples.map((s,i)=>`observe(regex, ${JSON.stringify(s.input)}, ${i});`).join('\n')+'\n';
 fs.writeFileSync(path.join(root,f.file),prefix+`const pattern = ${JSON.stringify(f.pattern)};\nconst flags = ${JSON.stringify(f.flags)};\nconst regex = new RegExp(pattern, flags);\n`+samples);
 // Runtime companion gets each string from input, so folding cannot hide the pending path.
 f.runtime_file=f.file.replace('.a','_runtime.a');
 fs.writeFileSync(path.join(root,f.runtime_file),prefix+`import { programArguments } from 'adamic';\nconst argumentsList = programArguments();\nconst pattern = argumentsList[0] ?? '';\nconst flags = argumentsList[1] ?? '';\nconst regex = new RegExp(pattern, flags);\n`+samples);
}
fs.writeFileSync(path.join(root,'fixtures.json'),JSON.stringify(fixtures,null,2)+'\n');
console.log(`${fixtures.length} fixtures agree on whole matches and UTF-16 spans; ${fixtures.length} runtime companions written`);
