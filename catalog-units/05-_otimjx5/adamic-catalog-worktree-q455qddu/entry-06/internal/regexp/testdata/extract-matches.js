// Capture native RegExpBuiltinExec calls from the pinned test262 suite. This
// observes calls made by exec, test, match, matchAll, replace and split. Exotic
// objects and custom exec implementations remain tests of the JS object model.
const fs = require('fs'), vm = require('vm'), path = require('path'), zlib = require('zlib');
const root = process.argv[2], destination = process.argv[3];
if (!process.version.startsWith('v24.')) throw Error('Node 24 required');
const suite = 'test/built-ins/RegExp';
function walk(directory) {
  return fs.readdirSync(path.join(root, directory)).flatMap(name => {
    const relative = path.join(directory, name);
    return fs.statSync(path.join(root, relative)).isDirectory() ? walk(relative) : name.endsWith('.js') ? [relative] : [];
  });
}
const rows = [], seen = new Set(), report = {node: process.version, files: 0, completed: 0, failed: [], timedOut: [], calls: 0, duplicates: 0, oversized: 0, exotic: 0, negativeSyntax: 0, oversizedSources: {}};
for (const relative of walk(suite).sort()) {
  report.files++;
  let source = fs.readFileSync(path.join(root, relative), 'utf8');
  if (/negative:[\s\S]*?phase:\s*parse/.test(source)) { report.negativeSyntax++; continue; }
  if (/flags:\s*\[[^\]]*onlyStrict/.test(source)) source = '"use strict";\n'+source;
  const context = vm.createContext({print(){}, console:{log(){}}, setTimeout(){}, clearTimeout(){}});
  context.record = function(row) {
    report.calls++;
    if (row.input.length > 4096 || row.pattern.length > 4096) { report.oversized++; return; }
    row.source = relative;
    const key = JSON.stringify([row.pattern, row.flags, row.input, row.lastIndex]);
    if (seen.has(key)) { report.duplicates++; return; }
    seen.add(key); rows.push(row);
  };
  context.exotic = () => {report.exotic++;};
 context.oversized = () => {report.oversized++;report.oversizedSources[relative]=(report.oversizedSources[relative]??0)+1;};
  context.global = context;
  context.$262 = {global: context, evalScript(s){return vm.runInContext(s,context,{timeout:1000});}, createRealm(){throw Error('unavailable secondary realm');}, detachArrayBuffer(){throw Error('unavailable detach');}, gc(){}};
  let includes = ['sta.js', 'assert.js'];
  const match = /includes:\s*\[([^\]]*)\]/.exec(source);
  if (match) includes.push(...match[1].split(',').map(s=>s.trim()).filter(Boolean));
  try {
    for (const include of new Set(includes)) {
      const file = path.join(root, 'harness', include);
      if (fs.existsSync(file)) vm.runInContext(fs.readFileSync(file,'utf8'),context,{filename:include,timeout:1000});
    }
    vm.runInContext(`
      const NativeRegExp = RegExp, nativeExec = RegExp.prototype.exec;
      const sourceGetter = Object.getOwnPropertyDescriptor(RegExp.prototype,'source').get;
      const flagGetters = ['hasIndices','global','ignoreCase','multiline','dotAll','unicode','unicodeSets','sticky'].map(name=>Object.getOwnPropertyDescriptor(RegExp.prototype,name).get);
      const flagNames = 'dgimsuvy';
      const apply=Reflect.apply, from=Array.from, entries=Object.entries, fromEntries=Object.fromEntries, charCodeAt=String.prototype.charCodeAt;
      const sources=new WeakMap();
      const codeUnits = s => from({length:s.length},(_,i)=>apply(charCodeAt,s,[i]));
      Object.defineProperty(RegExp.prototype,'exec',{configurable:true,writable:true,value:({exec(input){
        // Avoid repeating user coercion or observing an exotic lastIndex.
        if(typeof input!=='string'||typeof this.lastIndex!=='number'||!Number.isSafeInteger(this.lastIndex)||this.lastIndex<0){exotic();return apply(nativeExec,this,[input]);}
        let pattern,flags='';
        try{let cached=sources.get(this);if(!cached){pattern=apply(sourceGetter,this,[]);flagGetters.forEach((getter,i)=>{if(apply(getter,this,[]))flags+=flagNames[i];});cached={pattern,flags};sources.set(this,cached);}({pattern,flags}=cached);}catch{exotic();return apply(nativeExec,this,[input]);}
        const lastIndex=this.lastIndex;
        const result=apply(nativeExec,this,[input]);
        if(input.length>4096||pattern.length>4096){oversized();return result;}
        const oracle=new NativeRegExp(pattern,flags.indexOf('d')>=0?flags:flags+'d');oracle.lastIndex=lastIndex;
        const indexed=apply(nativeExec,oracle,[input]);
        // Verify replay produces the same capture strings and lastIndex before
        // using its indices. d changes only result shape.
        if((result===null)!==(indexed===null)||this.lastIndex!==oracle.lastIndex||result&&from(result).some((s,i)=>s!==indexed[i]))throw Error('indices replay changed execution');
        record({pattern,patternUnits:codeUnits(pattern),flags,input:codeUnits(input),lastIndex,expected:{captures:indexed?from(indexed.indices,x=>x??null):null,lastIndex:oracle.lastIndex,groups:indexed?.indices.groups?fromEntries(entries(indexed.indices.groups).map(([k,v])=>[k,v??null])):null}});
        return result;
      }}).exec});
    `,context,{timeout:1000});
    vm.runInContext(source,context,{filename:relative,timeout:1000});
    report.completed++;
  } catch(error) {
    const entry = {source:relative,error:String(error)};
    if(error.code==='ERR_SCRIPT_EXECUTION_TIMEOUT')report.timedOut.push(entry);else report.failed.push(entry);
  }
}
fs.writeFileSync(destination,zlib.gzipSync(JSON.stringify(rows)));
fs.writeFileSync(destination.replace(/\.json\.gz$/, '-extraction.json'),JSON.stringify(report,null,2)+'\n');
console.log(JSON.stringify({...report,failed:report.failed.length,timedOut:report.timedOut.length,oversizedSources:Object.keys(report.oversizedSources).length,unique:rows.length}));
