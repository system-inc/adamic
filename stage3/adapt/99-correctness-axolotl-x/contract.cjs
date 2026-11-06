'use strict';
const {fs,path,assert,normalize} = require('./engine.cjs');
const os = require('node:os');
const {adapt,inspect,owners} = require('./adapt.cjs');
const inventory = require('./inventory.json');
const tree = path.resolve(process.argv[2]);
const files = [...new Set(owners.map(pin=>pin.file))];
const original = new Map(files.map(file=>[file,normalize(fs.readFileSync(path.join(tree,'src/compiler',file),'utf8'))]));
assert.equal(inspect(original),false,'contract requires baseline');
const rows=[];
function rejected(name,change) {
    const texts = new Map(original); change(texts);
    assert.throws(()=>inspect(texts)); rows.push({name,result:'refused'});
}
const pin = owners[0];
rejected('changed owner',m=>m.set(pin.file,m.get(pin.file).replace(pin.before,pin.before.replace('{','{ /* drift */'))));
rejected('missing owner',m=>m.set(pin.file,m.get(pin.file).replace(pin.before,'')));
rejected('duplicate owner',m=>m.set(pin.file,m.get(pin.file).replace(pin.before,pin.before+'\n'+pin.before)));
rejected('half-adapted tree',m=>m.set(pin.file,m.get(pin.file).replace(pin.before,pin.after)));
const site = inventory.sites[0];
rejected('half-adapted owner',m=>m.set(site.file,m.get(site.file).replace(site.before,site.after)));
const imp=inventory.imports[0];
if(imp) rejected('half-adapted import',m=>m.set(imp.file,m.get(imp.file).replace(imp.before,'')));
const scratch=fs.mkdtempSync(path.join(os.tmpdir(),'bind-contract-'));
try {
    fs.cpSync(path.join(tree,'src'),path.join(scratch,'src'),{recursive:true});
    const changed = path.join(scratch,'src/compiler',pin.file);
    fs.writeFileSync(changed,fs.readFileSync(changed,'utf8')+'\n// unrelated contract edit\n');
    const applied=adapt(scratch); assert.equal(applied.rewrites,inventory.sites.length);
    const before=files.map(file=>{const name=path.join(scratch,'src/compiler',file);return [name,fs.readFileSync(name),fs.statSync(name).mtimeNs];});
    // BigInt stats retain nanosecond precision.
    const stamps=before.map(([name])=>fs.statSync(name,{bigint:true}).mtimeNs);
    assert.equal(adapt(scratch).files,0);
    for(let i=0;i<before.length;i++) {const [name,bytes]=before[i];assert.deepEqual(fs.readFileSync(name),bytes);assert.equal(fs.statSync(name,{bigint:true}).mtimeNs,stamps[i]);}
    rows.push({name:'foreign edit and replay',result:'pass',writes:0});
    const adapted=fs.readFileSync(changed,'utf8'); assert(normalize(adapted).endsWith('// unrelated contract edit\n'));
    // Restore, then exercise the same adapter against CRLF input.
    for(const [file,text] of original)fs.writeFileSync(path.join(scratch,'src/compiler',file),text.replace(/\n/g,'\r\n'));
    assert.equal(adapt(scratch).rewrites,inventory.sites.length);
    for(const file of files) {const text=fs.readFileSync(path.join(scratch,'src/compiler',file),'utf8');assert(!/(?<!\r)\n/.test(text));}
    rows.push({name:'CRLF preservation',result:'pass'});
} finally {fs.rmSync(scratch,{recursive:true,force:true});}
console.log(JSON.stringify(rows,null,2));
