'use strict';
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm'),assert=require('node:assert/strict'),ts=require('typescript');
const beforeRoot=process.argv[2],afterRoot=process.argv[3];
const calls=JSON.parse(fs.readFileSync(path.join(__dirname,'calls.main.json'))).filter(x=>x.after.includes('.call(timerHost,'));
assert.equal(calls.length,9);
const results=[];
for(const [index,site]of calls.entries()) {
 const source=fs.readFileSync(path.join(beforeRoot,site.file),'utf8').replace(/\r\n/g,'\n');assert(source.includes(site.after),'actual source block: '+site.file+':'+site.originalLine);
 const traces=[];
 for(const amended of [false,true]) {
  const queue={},state={},handle={},cb=()=>{},request={projectName:'before',kind:'discover'};
  const event=[],pending=[];
  const callback=name=>(...args)=>event.push([name,...args.map(identify)]);
  const host={setTimeout(fn,delay,label){assert.equal(this,host,'receiver identity');pending.push({fn,delay,label});return handle;}};
  const self={host,delay:19,logger:{hasLevel:()=>false},installer:{send:r=>event.push(['send',identify(r)])},operationHost:{getServerHost:()=>host,executeWithRequestId:(id,action)=>{event.push(['request',id]);action();}},executeAction:action=>action()};
  function identify(value){if(value===queue)return 'queue identity';if(value===state)return 'state identity';if(value===self)return 'self identity';if(value===cb)return 'cb identity';if(value===request)return 'request identity';return value;}
  const sandbox={result:undefined,LogLevel:{verbose:3},host,PollingInterval:{Low:250,Medium:500,High:2000},pollingInterval:250,pollingIntervalQueue:()=>queue,pollLowPollingIntervalQueue:callback('pollLowPollingIntervalQueue'),pollPollingIntervalQueue:callback('pollPollingIntervalQueue'),pollQueue:callback('pollQueue'),invalidateResolutionsOfFailedLookup:callback('invalidateResolutionsOfFailedLookup'),updateProgramWithWatchStatus:callback('updateProgramWithWatchStatus'),hostWithWatch:host,buildNextInvalidatedProject:callback('buildNextInvalidatedProject'),state,changeDetected:true,time:23,ThrottledOperations:{run:callback('ThrottledOperations.run')},GcTimer:{run:callback('GcTimer.run')},delay:17,operationId:'operation',cb,ms:31,actionType:'action',requestId:11,action:()=>event.push(['action']),TypingsInstallerAdapter:{requestDelayMillis:100},request,self};
  let block=site.after;
  if(amended)block=block.replace(/^([ \t]*)const \{ setTimeout \} = timerHost;\n/m,'').replace(/setTimeout(!?)\.call\(timerHost, /g,'timerHost.setTimeout$1(');
  if(amended)assert(fs.readFileSync(path.join(afterRoot,site.file),'utf8').replace(/\r\n/g,'\n').includes(block),'actual amended block');
  if(amended && index===0 && process.argv[4]==='delay-mutant')block=block.replace('const timeoutDelay = pollingInterval;','const timeoutDelay = pollingInterval + 1;');
  const js=ts.transpileModule('result = function(){ return '+block+'; }.call(self);',{compilerOptions:{target:ts.ScriptTarget.ES2024}}).outputText;
  vm.runInNewContext(js,sandbox);assert.equal(sandbox.result,handle,'returned handle identity');assert.equal(pending.length,1);
  const scheduled=pending[0];assert.equal(typeof scheduled.fn,'function');request.projectName='after';scheduled.fn(scheduled.label);
  traces.push({delay:scheduled.delay,label:identify(scheduled.label),events:event});
 }
 assert.deepEqual(traces[1],traces[0]);results.push({file:site.file,originalLine:site.originalLine,...traces[0]});
}
console.log(JSON.stringify({sites:results.length,results},null,2));
