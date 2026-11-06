'use strict';
const e = require('./engine.cjs');
const {fs, path, assert, ts} = e;
const owners = require('./owners.json');
const inventory = require('./inventory.json');
function inspect(texts, coreCalls) {
    assert(Array.isArray(coreCalls), "inspect requires symbol-resolved core calls");
    const states = [];
    for (const file of new Set(owners.map(pin => pin.file))) {
        const found = e.ownerMap(e.parse(file,texts.get(file)));
        for (const pin of owners.filter(pin => pin.file === file)) {
            const owner = found.get(pin.owner);
            assert(owner, `${file}: missing owner ${pin.owner}`);
            const hash = e.digest(owner.text);
            assert(hash === pin.ownerBeforeSha256 || hash === pin.ownerAfterSha256, `${file}: changed or half-adapted owner ${pin.owner}`);
            const state = hash === pin.ownerAfterSha256;
            assert.equal(owner.text, state ? pin.after : pin.before, `${file}: exact owner text drift ${pin.owner}`);
            states.push(state);
        }
    }
    assert(states.length && states.every(state => state === states[0]), 'half-adapted file or tree');
    for (const file of new Set(owners.map(pin => pin.file))) {
        const source = e.parse(file,texts.get(file));
        let count = 0;
        function visit(node) { if (ts.isImportSpecifier(node) && node.name.text === 'maybeBind') count++; ts.forEachChild(node,visit); }
        visit(source);
        const calls = coreCalls.filter(site => site.file === file).length;
        assert(states[0] || calls, `${file}: missing core maybeBind binding in before owner`);
        assert.equal(count, calls ? 1 : 0, `${file}: half-adapted import: ${calls} core maybeBind calls`);
    }
    return states[0];
}
function adapt(tree) {
    assert.equal(ts.version,'6.0.3');
    const raw = new Map(e.sources(tree).map(name => [name,fs.readFileSync(name,'utf8')]));
    const normalized = new Map([...raw].map(([name,text]) => [name,e.normalize(text)]));
    const texts = new Map([...normalized].map(([name,text]) => [path.relative(path.join(tree,'src/compiler'),name),text]));
    const scanned = e.scan(tree,e.program(tree,normalized));
    const adapted = inspect(texts,scanned.coreCalls);
    const ownership = require('./ownership.json');
    const handed = site => ownership.some(row=>row.file===site.file&&(row.owner===site.owner||site.owner.startsWith(row.owner+'.')));
    const signature = site => JSON.stringify([site.file,site.owner,site.before,site.after]);
    const handedSignatures = new Set(inventory.left.map(signature));
    const left = scanned.sites.filter(site => handed(site) && handedSignatures.has(signature(site)));
    const owned = scanned.sites.filter(site=>!handed(site));
    const refused = scanned.refused.filter(site => !handed(site));
    assert.deepEqual(refused, inventory.refused, 'unexpected refused core maybeBind call');
    const under = (site,row) => site.file === row.file && (site.owner === row.owner || site.owner.startsWith(row.owner+'.'));
    for (const row of ownership) {
        const present = left.filter(site => under(site,row));
        const expected = inventory.left.filter(site => under(site,row));
        assert(expected.length, `empty ownership group: ${row.file}:${row.owner}`);
        if (present.length) assert.deepEqual(present.map(signature).sort(),expected.map(signature).sort(),`handed-off owner half folded or inventory drift: ${row.file}:${row.owner}`);
    }
    assert.deepEqual(owned.map(signature).sort(), adapted ? [] : inventory.sites.map(signature).sort(), 'core maybeBind site inventory drift');
    if (adapted) return {files:0, rewrites:0, left:left.length+refused.length, javascriptChanging:true};
    const plans = [];
    for (const file of new Set(owners.map(pin => pin.file))) {
        const name = path.join(tree,'src/compiler',file);
        let after = e.edit(texts.get(file),owned.filter(site => site.file === file));
        const calls = scanned.coreCalls.filter(site => site.file === file).length - owned.filter(site => site.file === file).length;
        const removal = e.removableImport(file,after,calls);
        assert.equal(!!removal,calls === 0, `${file}: import usage drift`);
        if (removal) after = e.edit(after,[removal]);
        texts.set(file,after);
        const original = raw.get(name);
        if (original.includes('\r\n')) after = after.replace(/\n/g,'\r\n');
        plans.push({name, before:original, after});
    }
    const planned = new Map([...normalized].map(([name,text]) => [name,texts.get(path.relative(path.join(tree,'src/compiler'),name)) ?? text]));
    const afterScan = e.scan(tree,e.program(tree,planned));
    assert(inspect(texts,afterScan.coreCalls), 'invalid planned owners');
    for (const [name,text] of raw) assert.equal(fs.readFileSync(name,'utf8'),text, `${name}: source changed while planning`);
    for (const plan of plans) fs.writeFileSync(plan.name,plan.after);
    return {files:plans.length, rewrites:owned.length, left:left.length+refused.length, javascriptChanging:true};
}
module.exports = {adapt, inspect, owners};
if (require.main === module) {
    assert.equal(process.argv.length,3,'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])),null,2));
}
