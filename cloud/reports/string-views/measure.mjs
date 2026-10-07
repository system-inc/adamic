import assert from 'node:assert/strict';
import { readFile, writeFile } from 'node:fs/promises';
import { stripTypeScriptTypes } from 'node:module';
import { performance } from 'node:perf_hooks';
import { loadavg, cpus } from 'node:os';
import { spawn, execFileSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import inspector from 'node:inspector';

const scratch = '/tmp/wasm-requests-profile';
const mode = process.argv[2] ?? 'measure';
const host = await readFile('cloud/reports/string-views/host.mjs', 'utf8');
// Reuse the exact accepted generator without changing it or duplicating its algorithm.
const generatorSource = host.slice(host.indexOf('let seed ='), host.indexOf('// Independent semantic pins')) + '\nexport { requests };';
const { requests } = await import(`data:text/javascript;base64,${Buffer.from(generatorSource).toString('base64')}`);
const source = stripTypeScriptTypes(await readFile('cloud/reports/string-views/service.a', 'utf8'), {sourceUrl:'service.a'});
const plain = await import(`data:text/javascript;base64,${Buffer.from(source).toString('base64')}`);
const expected = requests.map(request=>plain.handleRequest(request));
const expectedChecksum = expected.reduce((sum,response)=>sum+response.length,0);
const jsonl = requests.join('\n')+'\n';
if (mode==='prepare') {
	await writeFile(`${scratch}/requests.jsonl`,jsonl);
	await writeFile(`${scratch}/responses.jsonl`,expected.join('\n')+'\n');
	console.log(JSON.stringify({requests:requests.length,inputBytes:Buffer.byteLength(jsonl),checksum:expectedChecksum,sha256:createHash('sha256').update(jsonl).digest('hex')}));
	process.exit(0);
}

function machine() { return {load1:loadavg()[0],nproc:Number(execFileSync('nproc',{encoding:'utf8'}).trim()),cpu:cpus()[0].model}; }
function batch(handler,verify=false) {
	let checksum=0;
	for(let index=0;index<requests.length;index++) {
		const response=handler(requests[index]);
		if(verify) assert.equal(response,expected[index],`response ${index}`);
		checksum+=response.length;
	}
	assert.equal(checksum,expectedChecksum);
	return checksum;
}
async function native(binary='service-native',argument='warm') {
	const before=performance.now();
	return await new Promise((resolve,reject)=>{
		const child=spawn(`${scratch}/${binary}`,[`${scratch}/requests.jsonl`,argument],{stdio:['ignore','pipe','pipe']});
		let stdout='',stderr='',start,stop;
		child.stdout.on('data',chunk=>stdout+=chunk);
		child.stderr.on('data',chunk=>{
			stderr+=chunk.toString();
			if(start===undefined && stderr.includes('serve:start\n')) start=performance.now();
			if(stop===undefined && stderr.includes('serve:stop\n')) stop=performance.now();
		});
		child.on('error',reject);
		child.on('close',code=>{
			if(code!==0) return reject(new Error(`native exit ${code}: ${stderr}`));
			assert.ok(start!==undefined && stop!==undefined && stop>start,'separate native timing markers');
			assert.equal(Number(stdout.trim()),argument==='control'?requests.length:expectedChecksum);
			resolve({milliseconds:stop-start,wholeMilliseconds:performance.now()-before,stderr});
		});
	});
}

// Verify outside the measured interval and warm both persistent engines with the whole corpus.

batch(plain.handleRequest,true);
if(mode==='profile-node') {
	const session=new inspector.Session();session.connect();
	const post=(name,args={})=>new Promise((resolve,reject)=>session.post(name,args,(error,value)=>error?reject(error):resolve(value)));
	await post('Profiler.enable');await post('Profiler.setSamplingInterval',{interval:1000});
	await post('Profiler.start');
	const startMicros=Number(process.hrtime.bigint()/1000n);
	const start=performance.now();batch(plain.handleRequest);
	const milliseconds=performance.now()-start;
	const endMicros=Number(process.hrtime.bigint()/1000n);
	const {profile}=await post('Profiler.stop');session.disconnect();
	profile.measurementWindow={startMicros,endMicros};
	await writeFile(`${scratch}/${mode}.cpuprofile`,JSON.stringify(profile));
	console.log(JSON.stringify({mode,milliseconds,requests:requests.length,machine:machine()}));
	process.exit(0);
}

const warmNative=await native();
const rounds=[];
const orders=[['native','node'],['node','native'],['native','node'],['node','native'],['native','node']];
for(let index=0;index<5;index++) {
	const round={round:index+1,before:machine(),order:orders[index],rates:{},milliseconds:{}};
	for(const engine of orders[index]) {
		let milliseconds;
		if(engine==='native') {
			const observation=await native();
			milliseconds=observation.milliseconds;
			round.nativeWholeMilliseconds=observation.wholeMilliseconds;
		} else {
			const start=performance.now();batch(plain.handleRequest);
			milliseconds=performance.now()-start;
		}
		round.milliseconds[engine]=milliseconds;
		round.rates[engine]=requests.length*1000/milliseconds;
	}
	round.after=machine();rounds.push(round);console.log(JSON.stringify(round));
}
const best=Object.fromEntries(['native','node'].map(engine=>[engine,Math.max(...rounds.map(round=>round.rates[engine]))]));
const result={base:'8cb9d252a9dc79d563b2644e0368e8b2edb114c7',requests:requests.length,inputBytes:Buffer.byteLength(jsonl),inputSha256:createHash('sha256').update(jsonl).digest('hex'),checksum:expectedChecksum,warmNative,rounds,best,node:process.version};
await writeFile(process.argv[3],JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify({best,requests:requests.length}));
