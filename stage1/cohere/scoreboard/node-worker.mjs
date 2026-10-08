// Host tooling only. Shared source modules remain unchanged.
import { createInterface } from 'node:readline';
import { format as consoleFormat } from 'node:util';
import { pathToFileURL } from 'node:url';
import { Parser } from '../../typescript/parser/parser.ts';
const lint = await import(pathToFileURL(process.argv[2]).href);
const formatter = await import(pathToFileURL(process.argv[3]).href);
const write = process.stdout.write.bind(process.stdout);
let active, chunks, start, operation, phase = 'host', faultPhase = '';
function setPhase(value) { phase=value;process.stderr.write('scoreboard-phase\t'+value+'\n'); }
const originalFile=Parser.prototype.file;
Parser.prototype.file=function(...args) {
 setPhase('parser');
 try{return originalFile.apply(this,args);}
 catch(error){faultPhase='parser';throw error;}
 finally{setPhase(operation==='lint'?'rule':'formatter');}
};
function reply(result) { write(JSON.stringify(result) + '\n'); }
// The real runtime's panic exits 70. Preserve output collected before it exited.
process.on('exit', (code) => { if(active) reply({stdout: chunks.join(''), exit_code:code, error:`worker exited ${code}`,phase,elapsed_ns:Number(process.hrtime.bigint()-start)}); });
for await (const line of createInterface({ input: process.stdin, crlfDelay: Infinity })) {
 const q=JSON.parse(line);operation=q.Op;faultPhase='';setPhase('host');chunks=[];start=process.hrtime.bigint();active=true;
 const prior=console.log;console.log=(...args)=>{const text=consoleFormat(...args)+'\n';chunks.push(text);if(!text.startsWith('fixed\t'))process.stderr.write('scoreboard-log\t'+JSON.stringify(text)+'\n');};
 let result={exit_code:0,reusable:true};
 try {
  const source=Buffer.from(q.Source,'base64').toString('utf8');
  if(q.Op==='lint'){console.log('case 0');lint.run(q.Path+'\t'+q.Rule+'\t\t\t\t'+(q.Options??''),false,0,'','',0,source);}
  else if(q.Op==='format'){formatter.formatHost(q.Path,q.Family,source);}
  else throw Error('unknown operation '+q.Op);
 } catch(error) { result={exit_code:70,reusable:true,phase:faultPhase||phase,error:String(error),stack:error.stack,stderr:`adamic: panic: ${String(error)}\n`}; }
 finally { console.log=prior; }
 result.stdout=chunks.join('');result.phase=result.phase||phase;result.elapsed_ns=Number(process.hrtime.bigint()-start);active=false;reply(result);
}
