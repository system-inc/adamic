// Execute the unmodified selector port on cohere's own concrete Go fixtures,
// instrumenting only regex literals through a TypeScript AST transformation.
const fs=require('node:fs'), path=require('node:path'), cp=require('node:child_process'), ts=require('typescript');
const root=path.resolve(__dirname,'../../../..');
const inventory=JSON.parse(fs.readFileSync(`${__dirname}/inventory.json`,'utf8'));
const fixtures=JSON.parse(cp.execFileSync('go',['run','./internal/regexp/testdata/cohere/selector-fixtures.go'],{cwd:root,encoding:'utf8'}));
const cache=new Map(), registered=new WeakMap();
let current;
function observe(id,regex) { registered.set(regex,id); return regex; }
const original=RegExp.prototype[Symbol.split];
RegExp.prototype[Symbol.split]=function(input,limit) {
  const id=registered.get(this);
  if(id!==undefined) inventory.patterns[id].inputs.push({units:Array.from({length:input.length},(_,i)=>input.charCodeAt(i)),source:{file:current.file,line:current.line},method:'split',evidence:'observed by running the selector port on this cohere test fixture',fixture:current.text});
  return original.call(this,input,limit);
};
function load(file) {
  if(cache.has(file)) return cache.get(file).exports;
  const module={exports:{}};cache.set(file,module);
  const relative=path.relative(root,file).split(path.sep).join('/');
  const patterns=inventory.patterns.filter(p=>p.file===relative);
  const result=ts.transpileModule(fs.readFileSync(file,'utf8'),{fileName:file,compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022},transformers:{before:[context=>source=>{
    function visit(node) {
      if(ts.isRegularExpressionLiteral(node)) {
        const loc=source.getLineAndCharacterOfPosition(node.getStart(source));
        const p=patterns.find(p=>p.line===loc.line+1 && p.column===loc.character+1);
        if(!p) throw Error('unrecorded regex literal');
        return ts.factory.createCallExpression(ts.factory.createIdentifier('__observe'),undefined,[ts.factory.createNumericLiteral(p.id),node]);
      }
      return ts.visitEachChild(node,visit,context);
    }
    return ts.visitNode(source,visit);
  }]}});
  function requirePort(name) {
    if(name==='adamic') return {panic:message=>{throw Error(message)},utf8Length:text=>Buffer.byteLength(text)};
    if(!name.startsWith('./')) throw Error(`unexpected import ${name}`);
    return load(path.resolve(path.dirname(file),name));
  }
  new Function('require','module','exports','__observe',result.outputText)(requirePort,module,module.exports,observe);
  return module.exports;
}
try {
  const {parse}=load(path.join(root,'stage1/cohere/selector/parser.ts'));
  for(const fixture of fixtures) { current=fixture;parse(fixture.text); }
} finally { RegExp.prototype[Symbol.split]=original; }
fs.writeFileSync(`${__dirname}/selector-fixtures.json`,JSON.stringify(fixtures,null,2)+'\n');
fs.writeFileSync(`${__dirname}/inventory.json`,JSON.stringify(inventory,null,2)+'\n');
console.log(JSON.stringify({fixtures:fixtures.length,observedCalls:inventory.patterns.filter(p=>p.file==='stage1/cohere/selector/parser.ts').reduce((n,p)=>n+p.inputs.length,0)}));
