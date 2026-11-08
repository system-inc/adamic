// transpile.cjs: a program's modules as tsc ships them, for the installed-layout harness
// (internal/oracle/installed_layout_test.go). Each module is transpiled to CommonJS by TypeScript's own
// transpileModule, then either wrapped into one bundle, as _tsc.js holds tsc's modules, or written a file each.
//
//	node transpile.cjs <typescript.js> bundle <root> <entry> <out.js> <module>...
//	node transpile.cjs <typescript.js> files <root> <entry> <out directory> <module>...
//
// Modules and the entry are paths from root, with / between directories. In a bundle each module is a function
// (module, exports, require) in a table keyed by that path, and a module's require serves the table for a relative
// specifier, resolved from the module's own directory, and falls through to Node's for any other (node: modules).
// No bundler is used: every module in the bundle reads the bundle's own __filename and __dirname, as every module of
// _tsc.js does. A file each, a relative import's .ts is rewritten to .js, and each module reads its own.
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const [api, mode, root, entry, out, ...modules] = process.argv.slice(2);
const ts = require(api);

function transpile(name, rewrite) {
	const source = fs.readFileSync(path.join(root, name), 'utf8');
	const compilerOptions = { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, rewriteRelativeImportExtensions: rewrite };
	return ts.transpileModule(source, { fileName: name, compilerOptions }).outputText;
}

if (mode === 'files') {
	for (const name of modules) {
		const destination = path.join(out, name.replace(/\.ts$/, '.js'));
		fs.mkdirSync(path.dirname(destination), { recursive: true });
		fs.writeFileSync(destination, transpile(name, true));
	}
} else if (mode === 'bundle') {
	const table = modules.map(name => JSON.stringify(name) + ': function (module, exports, require) {\n' + transpile(name, false) + '\n},\n').join('');
	const bundle = [
		"'use strict';",
		'const adamicModules = {',
		table + '};',
		'const adamicLoaded = {};',
		"const adamicPath = require('node:path');",
		'function adamicRequireFrom(directory) {',
		'\treturn specifier => {',
		"\t\tif (!specifier.startsWith('.')) return require(specifier);",
		'\t\tconst name = adamicPath.posix.join(directory, specifier);',
		'\t\tif (!Object.hasOwn(adamicLoaded, name)) {',
		'\t\t\tconst module = { exports: {} };',
		'\t\t\tadamicLoaded[name] = module;',
		'\t\t\tadamicModules[name](module, module.exports, adamicRequireFrom(adamicPath.posix.dirname(name)));',
		'\t\t}',
		'\t\treturn adamicLoaded[name].exports;',
		'\t};',
		'}',
		"adamicRequireFrom('.')(" + JSON.stringify('./' + entry) + ');',
		'',
	].join('\n');
	fs.writeFileSync(out, bundle);
} else {
	throw new Error('transpile.cjs: the mode is bundle or files, not ' + mode);
}
