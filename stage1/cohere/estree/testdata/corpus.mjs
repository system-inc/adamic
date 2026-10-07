import fs from 'node:fs';
import { registerHooks, stripTypeScriptTypes } from 'node:module';
import { fileURLToPath } from 'node:url';
const runtime = new URL('./audit-runtime.mjs', import.meta.url).href;
registerHooks({
  resolve(specifier, context, next) { return specifier === 'adamic' ? {url:runtime,shortCircuit:true} : next(specifier,context); },
  load(url, context, next) {
    if(!url.endsWith('.ts')) return next(url,context);
    let source=stripTypeScriptTypes(fs.readFileSync(fileURLToPath(url),'utf8'));
    return {format:'module',source,shortCircuit:true};
  },
});
const { answer } = await import('../pipeline.ts');
const records = fs.readFileSync(process.argv[2],'utf8').trim().split('\n').map(JSON.parse);
const manifest = [];
const counts = {};
const output = fs.openSync(process.argv[3],'w');
for(let index=0;index<records.length;index++) {
  const record = records[index];
  let status;
  try {
    const got = answer(record.path,fs.readFileSync(record.path,'utf8'));
    status = record.status !== 'ok' ? 'port-accepts-go-refusal' : got === fs.readFileSync(record.answer,'utf8') ? 'identical' : 'different';
    if(status==='identical') manifest.push(record.path);
    if(status==='different') {
      const want=fs.readFileSync(record.answer,'utf8').split('\n'), lines=got.split('\n');
      let line=0;while(line<Math.min(want.length,lines.length)&&want[line]===lines[line]) line++;
      record.difference={line:line+1,go:want[line],port:lines[line]};
    }
  } catch(error) { status=record.status==='ok' ? 'port-refuses-go-answer' : 'both-refuse';record.portError=String(error.message); }
  record.portStatus=status;
  counts[status]=(counts[status]??0)+1;
  fs.writeSync(output,JSON.stringify(record)+'\n');
  if(index%500===0) fs.writeSync(process.stdout.fd,JSON.stringify({seen:index+1,counts})+'\n');
}
fs.closeSync(output);
fs.writeFileSync(process.argv[4],manifest.join('\n')+'\n');
process.stdout.write(JSON.stringify({total:records.length,counts})+'\n');
