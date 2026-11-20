const fs = require('node:fs');
const vm = require('node:vm');
const dir = process.env.INFER_CENSUS_DIR;
const rows = [];
function record(flags, same, control = false) {
    rows.push([flags, same, control]);
}
globalThis.__inferParameterCensus = record;
for (const method of ['runInNewContext', 'runInContext']) {
    const original = vm[method];
    vm[method] = function (code, context, ...rest) {
        if (context && typeof code === 'string' && code.includes('__inferParameterCensus')) {
            Object.defineProperty(context, '__inferParameterCensus', { value: record, configurable: true });
        }
        return Reflect.apply(original, this, [code, context, ...rest]);
    };
}
try {
    const fd = fs.openSync(dir + '/control', 'wx');
    fs.closeSync(fd);
    const value = {};
    record(0, value === value, true);
    record(0, value === {}, true);
}
catch (error) {
    if (error.code !== 'EEXIST') throw error;
}
process.on('exit', () => {
    if (rows.length) fs.writeFileSync(dir + '/' + process.pid + '.json', JSON.stringify(rows));
});
