// Expose unchanged private dispatch and whitespace functions from the pinned fork in memory.
import fs from 'node:fs';
import vm from 'node:vm';
const source = fs.readFileSync(process.argv[2] + '/plugins/markdown.js', 'utf8');
const exposures = [['function ou(', '__white'], ['function lu(', '__dispatch']];
let exposed = source;
for (const [anchor, name] of exposures) {
    const offset = exposed.indexOf(anchor);
    if (offset < 0 || exposed.indexOf(anchor, offset + 1) >= 0) throw new Error('pinned function anchor changed');
    const declaration = /^function ([A-Za-z_$][\w$]*)\(/.exec(exposed.slice(offset));
    exposed = exposed.slice(0, offset) + `globalThis.${name} = ${declaration[1]};` + exposed.slice(offset);
}
const context = {module:{exports:{}}, exports:{}};
vm.runInNewContext(exposed, context);
const decode = text => text.replace(/\\(.)/g, (_, c) => c === 'n' ? '\n' : c === 'r' ? '\r' : c === 't' ? '\t' : c);
const token = (f, i) => f[i] !== '1' ? undefined : ({type:f[i+1], value:decode(f[i+2]), kind:f[i+3], isCJ:(Number(f[i+4])&1)!==0, hasLeadingPunctuation:(Number(f[i+4])&2)!==0, hasTrailingPunctuation:(Number(f[i+4])&4)!==0});
const classify = doc => typeof doc === 'string' ? 'T'+doc : Array.isArray(doc) ? 'H' : doc.type === 'line' ? doc.soft ? 'S' : doc.hard ? 'H' : 'L' : (()=>{throw new Error('unknown whitespace document')})();
const output = [];
for (const line of fs.readFileSync(process.argv[3], 'utf8').split('\n')) {
    if (!line.startsWith('S\t')) continue;
    const f = line.split('\t');
    const node = {type:'whitespace', value:decode(f[1])};
    const kinds = f[19] === '' ? [] : f[19].split(',');
    const setext = f[20].split(',');
    const ancestors = kinds.map((type,i)=>({type, referenceType:'full', position:{start:{line:1},end:{line:setext[i]==='1'?2:1}}}));
    const parent = {type:'sentence',children:[]};
    for (const item of f[21].split(';')) {
        if (!item) continue;
        const [type,value,previousKind,nextKind] = item.split(',');
        parent.children.push({type:'word',kind:previousKind},{type,value:value==='0'?'':value==='1'?' ':'x'},{type:'word',kind:nextKind});
    }
    const path = {node, parent, previous:token(f,4), next:token(f,9), siblings:[node,token(f,9),token(f,14)], index:0, hasAncestor:predicate=>ancestors.some(predicate), findAncestor:predicate=>ancestors.find(predicate)};
    const observed = [];
    for (const mode of ['preserve','always','never']) {
        const options = {parser:'markdown',proseWrap:mode,originalText:''};
        observed.push(classify(context.__dispatch(path, options, ()=>{throw new Error('unexpected child print')})));
        observed.push(classify(context.__white(path, node.value, mode, true, options)));
    }
    output.push(observed.join(','));
}
process.stdout.write(output.join('\n') + '\n');
