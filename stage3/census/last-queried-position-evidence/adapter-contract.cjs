const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');

const [adapterPath, parserPath] = process.argv.slice(2);
const adapter = require(adapterPath);
const source = fs.readFileSync(parserPath, 'utf8');
const original = 'let lastQueriedPosition = InvalidPosition.Value;';
const adapted = 'let lastQueriedPosition: number = InvalidPosition.Value;';
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'parser-adapter-'));
const target = path.join(scratch, 'src/compiler/parser.ts');

try {
    fs.mkdirSync(path.dirname(target), {recursive: true});
    for (const [label, text] of [['CRLF', source], ['LF', source.replaceAll('\r\n', '\n')]]) {
        fs.writeFileSync(target, text);
        assert.equal(adapter.adapt(scratch).edits, 1);
        assert.equal(fs.readFileSync(target, 'utf8'), text.replace(original, adapted));
        const mtime = fs.statSync(target, {bigint: true}).mtimeNs;
        assert.equal(adapter.adapt(scratch).edits, 0);
        assert.equal(fs.statSync(target, {bigint: true}).mtimeNs, mtime);
        assert.equal(adapter.rewrite(scratch, true).edits, 1);
        assert.equal(fs.readFileSync(target, 'utf8'), text);
        console.log(label, 'exact edit, replay without write, exact mutant');
    }
    const cases = [
        ['mixed endings', source.replace('\r\n', '\n'), /mixed line endings/],
        ['parse error', source + '\r\nlet = ;\r\n', /parse error/],
        ['drift', source.replace(original, 'let lastQueriedPosition = -1;'), /changed or half-adapted owner/],
        ['half-adapted', source.replace(original, 'let lastQueriedPosition: number = -1;'), /changed or half-adapted owner/],
    ];
    for (const [label, text, error] of cases) {
        assert.notEqual(text, source, label);
        fs.writeFileSync(target, text);
        assert.throws(() => adapter.adapt(scratch), error);
        assert.equal(fs.readFileSync(target, 'utf8'), text);
        console.log(label, 'refused without writing');
    }
} finally {
    fs.rmSync(scratch, {recursive: true, force: true});
}
