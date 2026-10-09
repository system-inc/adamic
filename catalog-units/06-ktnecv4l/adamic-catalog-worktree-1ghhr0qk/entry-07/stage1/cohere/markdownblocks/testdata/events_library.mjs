import fs from 'node:fs';import vm from 'node:vm';
let bundle=fs.readFileSync(process.argv[2]+'/plugins/markdown.js','utf8');const anchor='function _i(e){';
if(bundle.split(anchor).length!==2||!bundle.includes('function zu(e,t,r){'))throw new Error('pinned tokenizer anchor');
bundle=bundle.replace(anchor,'globalThis.adamicCreateTokenizer=zu;globalThis.adamicPreprocess=li;'+anchor);
const global={module:{exports:{}},exports:{}};vm.runInNewContext(bundle,global);
const encode=text=>text.replace(/[\\\n\r\t]/g,c=>c==='\\'?'\\\\':c==='\n'?'\\n':c==='\r'?'\\r':'\\t');
const point=p=>[p.line,p.column,p.offset,p._index,p._bufferIndex].join(',');const output=[];
for(const line of fs.readFileSync(process.argv[3],'utf8').split('\n')){
 if(!line)continue;const units=line==='-'?[]:line.split(',').map(Number);let source='';for(let i=0;i<units.length;i+=8192)source+=String.fromCharCode(...units.slice(i,i+8192));
 const before={name:'before'},changed={name:'changed'},fields={_spread:true,_align:['left']};
 const check={partial:true,tokenize:function(e,ok){e.enter('rollback');this.currentConstruct=changed;return code=>{e.consume(code);e.exit('rollback');return ok}}};
 let context;const initial={tokenize:function(e){const self=this;self.defineSkip({line:2,column:3});self.defineSkip({line:3,column:5});self.currentConstruct=before;e.enter('document');let data=false;
  function after(code){if((code===null||code<0)&&data){e.exit('data');data=false}if(code!==null&&code<0){e.enter('virtual',fields);e.consume(code);e.exit('virtual')}else if(code!==null&&code>0){if(!data){e.enter('data');data=true}e.consume(code)}else{e.exit('document');e.consume(code)}return state}
  function state(code){return code!==null&&code<0?e.check(check,after,after)(code):after(code)}return state}};
 context=global.adamicCreateTokenizer({constructs:{disable:{null:[]}}},initial);
 const events=context.write(global.adamicPreprocess()(source,undefined,true));const rows=[];
 for(const [direction,t]of events){let text='-',expanded='-';if(t.start._index!==t.end._index||(t.start._bufferIndex>=0&&t.end._bufferIndex>=0)){text=context.sliceSerialize(t,false);expanded=context.sliceSerialize(t,true)}rows.push([direction==='enter',t.type,point(t.start),point(t.end),t._spread??false,(t._align??[]).join(','),encode(text),encode(expanded)].join('\t'))}
 // consumed and empty stack are invariants of the completed original state program.
 rows.push(['now',point(context.now()),context.previous??0,context.currentConstruct===before?7:9,true,0].join('\t'));output.push(encode(rows.join('\n')));
}
process.stdout.write(output.join('\n')+'\n');
