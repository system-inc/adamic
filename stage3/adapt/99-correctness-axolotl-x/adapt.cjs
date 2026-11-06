'use strict';
const e = require('./engine.cjs');
const {fs, path, assert, ts} = e;
const owners = require('./owners.json');
const inventory = require('./inventory.json');
function inspect(texts) {
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
    for (const pin of inventory.imports) {
        const source = e.parse(pin.file,texts.get(pin.file));
        let count = 0;
        function visit(node) { if (ts.isImportSpecifier(node) && node.name.text === 'maybeBind') count++; ts.forEachChild(node,visit); }
        visit(source);
        assert.equal(count, states[0] ? 0 : 1, `${pin.file}: half-adapted import`);
    }
    return states[0];
}
function adapt(tree) {
    assert.equal(ts.version,'6.0.3');
    const raw = new Map(e.sources(tree).map(name => [name,fs.readFileSync(name,'utf8')]));
    const normalized = new Map([...raw].map(([name,text]) => [name,e.normalize(text)]));
    const texts = new Map([...normalized].map(([name,text]) => [path.relative(path.join(tree,'src/compiler'),name),text]));
    const adapted = inspect(texts);
    const scanned = e.scan(tree,e.program(tree,normalized));
    const ownership = require('./ownership.json');
    const handed = site => ownership.some(row=>row.file===site.file&&(row.owner===site.owner||site.owner.startsWith(row.owner+'.')));
    const left = scanned.sites.filter(handed);
    const owned = scanned.sites.filter(site=>!handed(site));
    assert.deepEqual(scanned.refused, inventory.refused, 'unexpected refused core maybeBind call');
    const signature = site => JSON.stringify([site.file,site.owner,site.before,site.after]);
    assert.deepEqual(left.map(signature).sort(),inventory.left.map(signature).sort(),'handed-off maybeBind site inventory drift');
    assert.deepEqual(owned.map(signature).sort(), adapted ? [] : inventory.sites.map(signature).sort(), 'core maybeBind site inventory drift');
    if (adapted) return {files:0, rewrites:0, left:left.length+scanned.refused.length, javascriptChanging:true};
    const plans = [];
    for (const file of new Set(owners.map(pin => pin.file))) {
        const name = path.join(tree,'src/compiler',file);
        let after = e.edit(texts.get(file),owned.filter(site => site.file === file));
        const removal = e.removableImport(file,after);
        const expected = inventory.imports.some(pin => pin.file === file);
        assert.equal(!!removal,expected, `${file}: import usage drift`);
        if (removal) after = e.edit(after,[removal]);
        texts.set(file,after);
        const original = raw.get(name);
        if (original.includes('\r\n')) after = after.replace(/\n/g,'\r\n');
        plans.push({name, before:original, after});
    }
    assert(inspect(texts), 'invalid planned owners');
    for (const [name,text] of raw) assert.equal(fs.readFileSync(name,'utf8'),text, `${name}: source changed while planning`);
    for (const plan of plans) fs.writeFileSync(plan.name,plan.after);
    return {files:plans.length, rewrites:owned.length, left:left.length+scanned.refused.length, javascriptChanging:true};
}
module.exports = {adapt, inspect, owners};
if (require.main === module) {
    assert.equal(process.argv.length,3,'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])),null,2));
}
