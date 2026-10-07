// Adapted from the accepted codex/wasm-requests-profile harness at a4e0902.
import assert from 'node:assert/strict';
import {readFile,writeFile} from 'node:fs/promises';
import {stripTypeScriptTypes} from 'node:module';
import {WASI} from 'node:wasi';
import {loadavg,cpus} from 'node:os';
import {execFileSync,spawn} from 'node:child_process';
import {performance} from 'node:perf_hooks';
import {createHash} from 'node:crypto';
import inspector from 'node:inspector';
const scratch='/tmp/decode-ascii', mode=process.argv[2]??'measure';
const host=await readFile(`${scratch}/service/host.mjs`,'utf8');
const generator=host.slice(host.indexOf('let seed ='),host.indexOf('// Independent semantic pins'))+'\nexport {requests};';
const {requests}=await import(`data:text/javascript;base64,${Buffer.from(generator).toString('base64')}`);
const stripped=stripTypeScriptTypes(await readFile(`${scratch}/service/service.a`,'utf8'),{sourceUrl:'service.a'});
const plain=await import(`data:text/javascript;base64,${Buffer.from(stripped).toString('base64')}`);
const expected=requests.map(plain.handleRequest), checksum=expected.reduce((sum,x)=>sum+x.length,0);
const jsonl=requests.join('\n')+'\n';
const corpus={requests:requests.length,bytes:Buffer.byteLength(jsonl),sha256:createHash('sha256').update(jsonl).digest('hex'),checksum};
await writeFile(`${scratch}/requests.jsonl`,jsonl);
const machine=()=>({load1:loadavg()[0],nproc:Number(execFileSync('nproc',{encoding:'utf8'}).trim()),cpu:cpus()[0].model});
const encoder=new TextEncoder(),decoder=new TextDecoder();
async function wasmHandler(stage){
 const wasi=new WASI({version:'preview1',args:[],env:{},preopens:{},returnOnExit:true});
 const module=await WebAssembly.compile(await readFile(`${scratch}/service-${stage}.wasm`));
 const instance=await WebAssembly.instantiate(module,wasi.getImportObject());wasi.initialize(instance);const api=instance.exports;
 return request=>{
  const bytes=encoder.encode(request),input=api.malloc(Math.max(1,bytes.length));assert.notEqual(input,0);
  new Uint8Array(api.memory.buffer,input,bytes.length).set(bytes);
  const response=api.adamic_request(input,bytes.length);
  const result=decoder.decode(new Uint8Array(api.memory.buffer,api.adamic_response_bytes(response),api.adamic_response_length(response)));
  api.adamic_release(response);api.free(input);return result;
 };
}
function batch(handler,verify=false){
 let sum=0;for(let i=0;i<requests.length;i++){const response=handler(requests[i]);if(verify)assert.equal(response,expected[i],`response ${i}`);sum+=response.length;}assert.equal(sum,checksum);
}
async function native(stage){
 const begin=performance.now();return await new Promise((resolve,reject)=>{
  const child=spawn(`${scratch}/native-${stage}`,[`${scratch}/requests.jsonl`,'warm'],{stdio:['ignore','pipe','pipe']});let stdout='',stderr='',start,stop;
  child.stdout.on('data',x=>stdout+=x);child.stderr.on('data',x=>{stderr+=x;if(start===undefined&&stderr.includes('serve:start\n'))start=performance.now();if(stop===undefined&&stderr.includes('serve:stop\n'))stop=performance.now();});
  child.on('error',reject);child.on('close',code=>{if(code!==0)return reject(new Error(`native ${stage}: ${stderr}`));assert.ok(stop>start);assert.equal(Number(stdout.trim()),checksum);resolve({milliseconds:stop-start,wholeMilliseconds:performance.now()-begin});});
 });
}
const handlers={};
for(const stage of mode==='profile'?['after']:['before','after']){handlers[stage]=await wasmHandler(stage);batch(handlers[stage],true);}
if(mode==='profile'){
 const before=machine(),session=new inspector.Session();session.connect();const post=(method,params={})=>new Promise((resolve,reject)=>session.post(method,params,(e,v)=>e?reject(e):resolve(v)));
 await post('Profiler.enable');await post('Profiler.setSamplingInterval',{interval:1000});await post('Profiler.start');
 const startMicros=Number(process.hrtime.bigint()/1000n),start=performance.now();batch(handlers.after);const milliseconds=performance.now()-start,endMicros=Number(process.hrtime.bigint()/1000n);
 const {profile}=await post('Profiler.stop');session.disconnect();profile.measurementWindow={startMicros,endMicros};
 await writeFile(`${scratch}/after.cpuprofile`,JSON.stringify(profile));console.log(JSON.stringify({milliseconds,before,after:machine()}));process.exit(0);
}
for(const stage of ['before','after']){
 const lines=execFileSync(`${scratch}/native-${stage}`,[`${scratch}/requests.jsonl`,'verify'],{encoding:'utf8',maxBuffer:64*1024*1024,stdio:['ignore','pipe','pipe']}).trimEnd().split('\n');
 assert.equal(lines.length,requests.length+1);assert.equal(Number(lines.pop()),checksum);assert.deepEqual(lines,expected);
}
await native('before');await native('after');
const orders=[['wasm-before','wasm-after','native-before','native-after'],['native-after','native-before','wasm-after','wasm-before'],['wasm-after','native-before','wasm-before','native-after'],['native-before','wasm-before','native-after','wasm-after'],['wasm-before','native-after','wasm-after','native-before']];
const rounds=[];
for(let i=0;i<5;i++){
 const before=machine(),observations={};
 for(const entry of orders[i]){const [engine,stage]=entry.split('-');if(engine==='native')observations[entry]=await native(stage);else{const start=performance.now();batch(handlers[stage]);observations[entry]={milliseconds:performance.now()-start};}}
 const round={round:i+1,order:orders[i],before,after:machine(),observations};rounds.push(round);console.log(JSON.stringify(round));
}
const best=Object.fromEntries(orders[0].map(entry=>[entry,100000/Math.min(...rounds.map(r=>r.observations[entry].milliseconds))*1000]));
await writeFile(process.env.ADAMIC_DECODE_RESULTS??'cloud/reports/decode-ascii/results.json',JSON.stringify({corpus,node:process.version,rounds,best},null,2)+'\n');console.log(JSON.stringify({best}));
