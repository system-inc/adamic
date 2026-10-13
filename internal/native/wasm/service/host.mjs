import assert from 'node:assert/strict';
import { readFile, writeFile } from 'node:fs/promises';
import { stripTypeScriptTypes } from 'node:module';
import { performance } from 'node:perf_hooks';
import { cpus, loadavg, totalmem } from 'node:os';
import { WASI } from 'node:wasi';

const [modulePath, sourcePath, requestPath, responsePath, counted = '1', regionsUsed = '0'] = process.argv.slice(2);
const source = stripTypeScriptTypes(await readFile(sourcePath, 'utf8'));
const plain = await import(`data:text/javascript;base64,${Buffer.from(source).toString('base64')}`);
const loadStart = loadavg();
let seed = 0x6a09e667;
function random() { seed = (Math.imul(seed, 1664525) + 1013904223) >>> 0; return seed; }
const names = ['Ahra', '世界 🌍', 'e\u0301', '梨', 'quote" slash\\ newline\n NUL\0', '\ud800', '\uffff', '🌍'];
function generate(index) {
	const n = random();
	const name = names[(n >>> 16) % names.length] + index;
	switch (index % 20) {
		case 0: return '{"method":"GET","path":"/health"}';
		case 1: return JSON.stringify({method:'GET',path:'/catalog'});
		case 2: case 3: case 4: return JSON.stringify({method:'POST',path:'/orders',body:{customer:name,items:Array.from({length:1+n%16},(_,i)=>({sku:['tea','bread','coffee','梨'][i%4],quantity:1+(n+i)%100}))}});
		case 5: return JSON.stringify({method:'POST',path:'/summary',body:{text:`${name} tea tea 梨 🌍`}});
		case 6: return JSON.stringify({method:'POST',path:'/listing',body:{names:[...names,name,'10','2','a','a']}});
		case 7: return JSON.stringify({method:'POST',path:'/quote',body:{amount:(n%100000000)/1000}});
		case 8: return JSON.stringify({method:'POST',path:'/orders',body:{customer:name,items:[{sku:'unknown',quantity:2}]}});
		case 9: return JSON.stringify({method:'POST',path:'/orders',body:{customer:name,items:[{sku:'tea',quantity:0.5}]}});
		case 10: return '{"method":"POST","path":"/summary","body":{"text":"\\u4e16\\u754c \\ud83c\\udf0d \\n \\" \\\\"}}';
		case 11: return ['{','{"x":01}','{"x":1e}','{"x":"\\q"}','[1,]','{"x":true,}', 'null trailing','{"x":"\t"}'][(n >>> 8)%8];
		case 12: return JSON.stringify({method:'DELETE',path:'/orders',body:{}});
		case 13: return JSON.stringify({method:'POST',path:'/listing',body:{names:[1,false,null]}});
		case 14: return JSON.stringify({method:'POST',path:'/summary',body:{text:('世界 🌍 tea ').repeat(512+n%512)}});
		case 15: return JSON.stringify({method:'POST',path:'/orders',body:{customer:name,items:Array.from({length:128},(_,i)=>({sku:'tea',quantity:1+i%100}))}});
		case 16: return JSON.stringify({method:'POST',path:'/quote',body:{amount:-1}});
		case 17: return JSON.stringify({method:n,path:'/health',body:[]});
		case 18: return '{"method":"POST","path":"/quote","body":{"amount":1.25e2}}';
		default: return JSON.stringify({method:'POST',path:'/summary',body:{text:'x'.repeat(65537)}});
	}
}
const requests = Array.from({length:100000},(_,index)=>generate(index));
// Independent semantic pins keep a shared parser/handler bug from becoming its own oracle.
const pins = [
	['{"method":"GET","path":"/health"}', '{"status":200,"service":"orders"}'],
	['{"method":"POST","path":"/listing","body":{"names":["\\u4e16\\u754c","\\ud83c\\udf0d","\\n","\\\"","\\\\","\\ud800"]}}', JSON.stringify({status:200,names:['世界','🌍','\n','"','\\','\ud800'].sort()})],
	[JSON.stringify({method:'POST',path:'/orders',body:{customer:'世界 🌍',items:[{sku:'tea',quantity:2},{sku:'bread',quantity:3}]}}), JSON.stringify({status:201,customer:'世界 🌍',lines:2,subtotal:1747,tax:144,total:1891,summary:'ありがとう 世界 🌍 🌍'})],
	['{"method":"POST","path":"/quote","body":{"amount":01}}', '{"status":400,"error":"invalid JSON"}'],
];
for (const [request, expected] of pins) assert.equal(plain.handleRequest(request), expected, 'source semantic pin');
assert.ok(requests.every(request=>!request.includes('\n') && !request.includes('\r')), 'each generated request must occupy one JSONL line');
await writeFile(requestPath, requests.join('\n') + '\n');
const expected = requests.map(request=>plain.handleRequest(request));
for (let index=0;index<requests.length;index++) {
	let input;
	let malformed=false;
	try { input=JSON.parse(requests[index]); } catch { malformed=true; }
	const response=JSON.parse(expected[index]);
	if (malformed) assert.equal(response.status,400,`Node JSON.parse rejected request ${index}`);
	if (input?.method==='POST' && input?.path==='/listing' && response.status===200) {
		assert.deepEqual(response.names,[...input.body.names].sort(),`Node JSON.parse/string sort ${index}`);
	}
}

await writeFile(responsePath, expected.join('\n') + '\n');
const bytes = await readFile(modulePath);
const wasi = new WASI({version:'preview1',args:[],env:{},preopens:{},returnOnExit:true});
const instance = await WebAssembly.instantiate(bytes, wasi.getImportObject());
wasi.initialize(instance.instance);
const api = instance.instance.exports;
const encoder = new TextEncoder(), decoder = new TextDecoder();
function handle(request) {
	const bytes = encoder.encode(request);
	const input = api.malloc(Math.max(1,bytes.length));
	assert.notEqual(input,0);
	try {
		new Uint8Array(api.memory.buffer,input,bytes.length).set(bytes);
		const response = api.adamic_request(input,bytes.length);
		try { return decoder.decode(new Uint8Array(api.memory.buffer,api.adamic_response_bytes(response),api.adamic_response_length(response))); }
		finally { api.adamic_release(response); }
	} finally { api.free(input); }
}
// Warm every generated size/shape, including the maximum body, in this same instance.
for (const [request,response] of pins) assert.equal(handle(request),response,'wasm semantic pin');
for (let index=0;index<requests.length;index++) {
	assert.equal(handle(requests[index]),expected[index],`warmup response ${index}`);
	if (counted==='1') assert.equal(api.adamic_live(),0,`warmup ${index} retained a live allocation`);
}
const baseline = api.memory.buffer.byteLength;
const regionStart = counted==='1' ? api.adamic_regions() : 0;
let checksum=0;
const wasmStarted = performance.now();
for (let index=0;index<requests.length;index++) {
	const response=handle(requests[index]);
	assert.equal(response,expected[index],`response ${index}`);
	if (counted==='1') assert.equal(api.adamic_live(),0,`request ${index} retained a live allocation`);
	assert.equal(api.memory.buffer.byteLength,baseline,`linear memory grew at request ${index}`);
	checksum+=response.length;
}
const wasmMilliseconds=performance.now()-wasmStarted;
const plainStarted=performance.now();
let plainChecksum=0;
for (let index=0;index<requests.length;index++) {
	const response=plain.handleRequest(requests[index]);
	assert.equal(response,expected[index],`plain response ${index}`);
	plainChecksum+=response.length;
}
const plainMilliseconds=performance.now()-plainStarted;
assert.equal(checksum,plainChecksum);
const regions = counted==='1' ? api.adamic_regions()-regionStart : 0;
if (counted==='1' && regionsUsed==='1') assert.ok(regions>0,'statement regions must advance');
console.log(JSON.stringify({requests:requests.length,seed:'0x6a09e667',counted:counted==='1',moduleBytes:bytes.length,memoryMinimum:baseline,memoryMaximum:api.memory.buffer.byteLength,live:counted==='1'?api.adamic_live():null,regions,wasmRequestsPerSecond:requests.length*1000/wasmMilliseconds,plainRequestsPerSecond:requests.length*1000/plainMilliseconds,node:process.version,cpu:cpus()[0].model,logicalCPUs:cpus().length,totalMemory:totalmem(),loadStart,loadEnd:loadavg()}));
