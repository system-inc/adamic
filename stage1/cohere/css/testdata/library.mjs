import {createRequire} from 'node:module';
import {readFileSync} from 'node:fs';
const require = createRequire(`${process.argv[2]}/package.json`);
const postcss = require('postcss');
const scss = require('postcss-scss');
if(require('postcss/package.json').version !== '8.5.16' || require('postcss-scss/package.json').version !== '4.0.9') throw new Error('oracle version drift');
const decode = line => line.replace(/\\([\\nrt])/g, (_, c) => ({'\\':'\\',n:'\n',r:'\r',t:'\t'})[c]);
function canonical(value) {
 if(Array.isArray(value)) return value.map(canonical);
 if(value && typeof value === 'object') return Object.fromEntries(Object.keys(value).sort().map(key => [key, canonical(value[key])]));
 return value;
}
function convert(value, css) {
 if(value === undefined) return '<undefined>';
 if(Array.isArray(value)) return value.map(child => convert(child, css));
 if(value && typeof value === 'object') {
  const result = {};
  for(const key of Object.keys(value)) {
   if(key === 'parent' || key === 'input' || key === 'type') continue;
   result[key] = key === 'offset' ? Buffer.byteLength(css.slice(0,value[key])) + Math.max(0,value[key]-css.length) : convert(value[key],css);
  }
  if(value.type) {result.type=value.type; result.keys=Object.keys(value).filter(key=>key!=='type' && key!=='parent');}
  return result;
 }
 return value;
}
if(process.argv[4] === 'count') {
 const inputs = readFileSync(process.argv[3],'utf8').split('\n').filter(Boolean).map(line => ({text:decode(line.slice(2)),scss:line[1]==='S'}));
 let parsed=0, children=0;
 for(let round=0;round<10;round++) {
  for(const input of inputs) {
   try {
    const root=input.scss?scss.parse(input.text,{map:false}):postcss.parse(input.text,{map:false});
    parsed++;children+=root.nodes.length;
   } catch(error) {if(!['CssSyntaxError','TypeError'].includes(error.name)) throw error;}
  }
 }
 console.log(`${parsed} of ${inputs.length*10} stylesheets parsed, ${children} nodes`);
 process.exit(0);
}
let number = 0;
for(const line of readFileSync(process.argv[3],'utf8').split('\n')) {
 if(!line) continue;
 const text = decode(line.slice(2));
 let answer;
 let error = false;
 try {
  const root = line[1] === 'S' ? scss.parse(text,{map:false}) : postcss.parse(text,{map:false});
  answer = convert(root,root.source.input.css);
 } catch(e) {
  error = true;
  if(e.name === 'CssSyntaxError') {
   answer={name:e.name,reason:e.reason,line:e.line,column:e.column,offset:e.input.offset};
   for(const key of ['endLine','endColumn','endOffset']) if(e.input[key] !== undefined) answer[key]=e.input[key];
   const css=text.replace(/^[\ufeff\ufffe]/,'');
   for(const key of ['offset','endOffset']) if(answer[key] !== undefined) answer[key]=Buffer.byteLength(css.slice(0,answer[key]))+Math.max(0,answer[key]-css.length);
  } else answer={notCssSyntaxError:`${e.name}: ${e.message}`};
 }
 console.log(`case ${number++}`);
 console.log(`${error?'error ':''}${JSON.stringify(canonical(answer)).replace(/\u2028/g,'\\u2028').replace(/\u2029/g,'\\u2029')}`);
}
