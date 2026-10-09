// Absolute settings/tsconfig paths, read on disk by the port; canonical records are the oracle seam.
import { panic, programArguments, readTextFile } from 'adamic';
import { loadSettings, printSettings } from './settings.ts';
import { loadHouse, resolveHouse, resolveSettings } from './house.ts';
import { dir } from '../gitignore/path.ts';
import { sortedKeys } from './settings.ts';
import { quoted } from './json.ts';
import { readProjectConfig } from './tsconfig.ts';
const input = readTextFile(programArguments()[0] ?? panic('usage: loader_main.ts <cases>'));
if (input.kind === 'Error') { panic(input.message); }
for (const line of input.text.split('\n')) {
	if (line === '') { continue; }
	const fields = line.split('\t');
	const path = fields[1] ?? panic('missing config path');
	if (fields[0] === 'house') {
		console.log(`house ${path}`);
		const root = dir(path);
		const result = loadHouse(path, { reactFiles: [root + '/React.tsx'], nextFiles: [root + '/Next.tsx'], tailwindEntryPoint: '', tailwindSkipped: 'no stylesheet' });
		if (result.kind === 'Error') { console.log(`error ${result.message}`); }
		else { printSettings(result.house.whole); }
	} else if (fields[0] === 'resolve') {
		const file = fields[2] ?? '';
		console.log(`resolve ${path} ${file}`);
		if (fields[3] === 'house') {
			const root = dir(path);
			const result = loadHouse(path, { reactFiles: [root + '/React.tsx'], nextFiles: [root + '/Next.tsx'], tailwindEntryPoint: '', tailwindSkipped: 'no stylesheet' });
			if (result.kind === 'Error') { console.log(`error ${result.message}`); }
			else { const resolved = resolveHouse(result.house, file); console.log(`ignored ${resolved.ignored ? 1 : 0} ${quoted(resolved.ignoredBy)}`); for (const name of sortedKeys(resolved.rules)) { const setting = resolved.rules.get(name) ?? panic('missing resolved rule'); console.log(`resolved-rule ${quoted(name)} ${setting.severity} [${setting.options.join(',')}]`); } }
		} else {
			const result = loadSettings(path);
			if (result.kind === 'Error') { console.log(`error ${result.message}`); }
			else { const resolved = resolveSettings(result.settings, file); console.log(`ignored ${resolved.ignored ? 1 : 0} ${quoted(resolved.ignoredBy)}`); for (const name of sortedKeys(resolved.rules)) { const setting = resolved.rules.get(name) ?? panic('missing resolved rule'); console.log(`resolved-rule ${quoted(name)} ${setting.severity} [${setting.options.join(',')}]`); } }
		}
	} else if (fields[0] === 'settings') {
		console.log(`settings ${path}`);
		const result = loadSettings(path, fields.slice(2));
		if (result.kind === 'Error') { console.log(`error ${result.message}`); }
		else { printSettings(result.settings); }
	} else if (fields[0] === 'tsconfig') {
		console.log(`tsconfig ${path}`);
		const result = readProjectConfig(path);
		if (result.kind === 'Error') { console.log(`error ${result.message}`); }
		else {
			for (const file of result.project.files) { console.log(`file ${quoted(file)}`); }
			for (const reference of result.project.references) { console.log(`reference ${quoted(reference)}`); }
		}
	} else { panic('unknown loader record'); }
}
