// Stock TypeScript helper/API observations, distinct from stock CLI behavior.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.STOCK_TYPESCRIPT);
if (ts.version !== '6.0.3') throw Error('Wrong stock version');
const root = __dirname;
const oldFile = path.join(root, 'input-fixtures/29_compare-old/tsconfig.json');
const newFile = path.join(root, 'input-fixtures/30_compare-new/tsconfig.json');
const oldOptions = ts.readConfigFile(oldFile, f => fs.readFileSync(f, 'utf8')).config.compilerOptions;
const newOptions = ts.readConfigFile(newFile, f => fs.readFileSync(f, 'utf8')).config.compilerOptions;
console.log('stock API compareDataObjects distinct empty paths:', ts.compareDataObjects(oldOptions, newOptions));
if (ts.compareDataObjects(oldOptions, newOptions) !== true) throw Error('Observation changed');
for (const key of ['constructor', 'toString', 'hasOwnProperty', '__proto__']) {
    const source = JSON.stringify({ compilerOptions: { paths: { [key]: ['./target'] } } });
    const parsed = ts.parseConfigFileTextToJson('tsconfig.json', source).config.compilerOptions.paths;
    console.log('config own key', key, Object.prototype.hasOwnProperty.call(parsed, key));
}
console.log('stock host getenv missing toString:', typeof ts.sys.getEnvironmentVariable('toString'));
// Coverage includes all 79 source operations. No guessed verdict may replace an unresolved one.
const global = JSON.parse(fs.readFileSync(path.join(root, 'inherited-key-global.json'), 'utf8'));
const rows = JSON.parse(fs.readFileSync(path.join(root, 'own-guard-verdicts.json'), 'utf8'));
function checkCoverage(data) {
    const ids = global.sites.filter(s => s.counted && ['user_input', 'unknown'].includes(s.classification)).map(s => s.id).sort();
    if (JSON.stringify(ids) !== JSON.stringify(data.map(s => s.id).sort())) throw Error('Coverage mismatch');
}
checkCoverage(rows);
const missing = rows.slice(1);
try { checkCoverage(missing); throw Error('Mutant survived'); }
catch (error) { if (error.message !== 'Coverage mismatch') throw error; }
console.log('CAUGHT by complete site IDs: omit one ownership-review row');
console.log('PASS: stock helper/API observations and 79-site coverage');
