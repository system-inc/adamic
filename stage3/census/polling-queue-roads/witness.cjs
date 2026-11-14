'use strict';
const fs=require('fs'),vm=require('vm'),assert=require('assert/strict'),ts=require('../../api/node_modules/typescript'),{roads}=require('./roads.cjs');
const sys=fs.readFileSync(process.argv[2]+'/src/compiler/sys.ts','utf8');
const sf=ts.createSourceFile('sys.ts',sys,ts.ScriptTarget.Latest,true);
function fn(name){const a=sf.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text===name);assert.equal(a.length,1);return a[0].getText(sf).replace(/^export /,'');}
const setup=`const PollingInterval={Low:250,Medium:500,High:2000};
const FileWatcherEventKind={Created:0,Changed:1,Deleted:2};
const missingFileModifiedTime=new Date(0);
const pollingChunkSize={250:2,500:2,2000:2};
const unchangedPollThresholds={250:1,500:2,2000:3};
const Debug={assert:(x)=>{if(!x)throw Error('assert')}};
function unorderedRemoveItem(a,x){const i=a.indexOf(x);if(i>=0){a[i]=a[a.length-1];a.pop();}}
`;
const dependencies=['getModifiedTime','pollWatchedFileQueue','onWatchedFileStat','getFileWatcherEventKind'].map(fn).join('\n');
const scenario=`
const events=[],pending=[],times=new Map();let id=0,now=0;
const host={getModifiedTime:n=>{events.push(['stat',now,n]);return new Date(times.get(n)||100);},setTimeout(cb,ms,label){pending.push({cb,at:now+ms,label,id:++id});return id;}};
const watch=createDynamicPriorityPollingWatchFile(host);
const handles=[];
for(const [n,p]of [['low',250],['medium',500],['high',2000]]){times.set(n,100);handles.push(watch(n,(n,k,t)=>events.push(['change',now,n,k,t.getTime()]),p));events.push(['add',n,p]);}
for(let step=0;step<65;step++){
 pending.sort((a,b)=>a.at-b.at||a.id-b.id);const t=pending.shift();if(!t)break;now=t.at;
 if(step===14||step===30){for(const n of times.keys())times.set(n,100+step);events.push(['modify-all',now]);}
 if(step===48){handles[0].close();events.push(['close-low',now]);}
 events.push(['fire',now,t.label]);t.cb(t.label);
}
result=events;
`;
let base;
for(const [road,owner]of Object.entries(roads)){
const instrumented=owner.replace('function addToPollingIntervalQueue(file: WatchedFileWithUnchangedPolls, pollingInterval: PollingInterval) {', 'function addToPollingIntervalQueue(file: WatchedFileWithUnchangedPolls, pollingInterval: PollingInterval) { events.push(["enqueue", now, file.fileName, pollingInterval]);').replace('function addChangedFileToLowPollingIntervalQueue(file: WatchedFileWithUnchangedPolls) {', 'function addChangedFileToLowPollingIntervalQueue(file: WatchedFileWithUnchangedPolls) { events.push(["changed-list", now, file.fileName]);');
const context={result:null};vm.createContext(context);
vm.runInContext(ts.transpileModule(setup+dependencies+'\n'+instrumented+'\n'+scenario,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.None}}).outputText,context,{timeout:10000});
const events=JSON.parse(JSON.stringify(context.result));if(base)assert.deepEqual(events,base,road);else base=events;
fs.writeFileSync(__dirname+`/evidence/witness.${road}.json`,JSON.stringify(events,null,2)+'\n');console.log(road,events.length,'events',events.filter(x=>x[0]==='change').length,'changes');console.log(JSON.stringify(events));
}
