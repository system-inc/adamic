// A stock-accepted genuine Node can have its kind rewritten through a mutable view.
// This tests the global invariant required by a prospective inserted kind check.
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const assert = require('node:assert/strict');
const {spawnSync} = require('node:child_process');
const ts = require('typescript');
const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'assertions-kind-'));
const filename = path.join(directory,'probe.a');
const source = `import ts from ${JSON.stringify(require.resolve('typescript'))};
type Mutable<T> = { -readonly [P in keyof T]: T[P] };
const token = ts.factory.createToken(ts.SyntaxKind.PlusToken);
const node: ts.Node = token;
(node as Mutable<ts.Node>).kind = ts.SyntaxKind.Identifier;
const acceptedTag = node.kind === ts.SyntaxKind.Identifier;
console.log(String(acceptedTag));
console.log(String((node as ts.Identifier).escapedText));
`;
fs.writeFileSync(filename,source);
const options = {allowNonTsExtensions:true,strict:true,noEmit:true,skipLibCheck:true,target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,esModuleInterop:true};
const host = ts.createCompilerHost(options);
const original = host.getSourceFile;
host.getSourceFile = (file, language, onError, fresh) => file === filename ? ts.createSourceFile(file,source,language,true,ts.ScriptKind.TS) : original(file,language,onError,fresh);
const program = ts.createProgram([filename],options,host);
// allowNonTsExtensions is necessary because the test's artifact is Adamic source.
// It changes extension admission only; it does not weaken semantic checks.

const diagnostics = ts.getPreEmitDiagnostics(program).map(d=>({code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
assert.deepEqual(diagnostics,[]);
const observed = spawnSync('node',['--disable-warning=ExperimentalWarning','oracle/node.mjs',filename],{encoding:'utf8'});
assert.equal(observed.status,0,observed.stderr);
assert.equal(observed.stdout,'true\nundefined\n');
assert.equal(observed.stderr,'');
console.log(JSON.stringify({typescript:ts.version,source,diagnostics,node:{stdout:observed.stdout,stderr:observed.stderr,exit:observed.status}},null,2));
fs.unlinkSync(filename);
fs.rmdirSync(directory);
