// configuration.Load: strict JSON layers, embedded sets, validation, merge and plugin defaults.
import { panic, readTextFile } from 'adamic';
import { clean, dir } from '../gitignore/path.ts';
import { compareCodePoints, join } from '../formatfiles/golang.ts';
import { compact, JsonDocument, JsonNode, parseJson, quoted } from './json.ts';
import { versionRangeError } from './version.ts';
import { pluginDefaults, setNames, setText } from './sets.ts';

export interface RuleSetting { readonly severity: string; readonly options: readonly string[]; }
export interface Reason { readonly file: string; readonly reason: string; }
export interface Override { readonly files: readonly string[]; readonly rules: ReadonlyMap<string, RuleSetting>; readonly reason: string; readonly file: string; }
export class Settings {
	readonly root: string;
	readonly rules = new Map<string, RuleSetting>();
	readonly ignores: string[] = [];
	readonly overrides: Override[] = [];
	readonly plugins: string[] = [];
	readonly sources: string[] = [];
	readonly departures = new Map<string, Reason>();
	readonly offReasons = new Map<string, Reason>();
	version = '';
	constructor(root: string) { this.root = root; }
}
export class Layer {
	readonly path: string;
	readonly document: JsonDocument;
	readonly extended: string[] = [];
	constructor(path: string, document: JsonDocument) { this.path = path; this.document = document; }
}
export type SettingsResult = { readonly kind: 'Ok'; readonly settings: Settings } | { readonly kind: 'Error'; readonly message: string };
type LayersResult = { readonly kind: 'Ok'; readonly layers: readonly Layer[] } | { readonly kind: 'Error'; readonly message: string };
type RuleResult = { readonly kind: 'Ok'; readonly setting: RuleSetting } | { readonly kind: 'Error'; readonly message: string };
export function sortedKeys<T>(map: ReadonlyMap<string, T>): string[] { return [...map.keys()].sort(compareCodePoints); }
function nodeKeys(document: JsonDocument, index: number): string[] {
	if (index < 0 || document.kind(index) === 'null') { return []; }
	return [...new Map(document.node(index).keys.map((key) => [key, true] as const)).keys()].sort(compareCodePoints);
}
function field(document: JsonDocument, key: string): number { return document.get(0, key); }
function fieldText(document: JsonDocument, key: string): string { return document.string(field(document, key)); }
function rawType(document: JsonDocument, index: number): string {
	const kind = document.kind(index);
	return kind === 'bool' ? 'bool' : kind;
}
function unmarshalError(document: JsonDocument, index: number, target: string, fieldName: string): string {
	return `json: cannot unmarshal ${rawType(document, index)} into Go struct field ${fieldName} of type ${target}`;
}
function validateRaw(document: JsonDocument): string {
	if (document.kind(0) !== 'object' && document.kind(0) !== 'null') { return `json: cannot unmarshal ${rawType(document, 0)} into Go value of type configuration.rawConfig`; }
	for (const key of document.node(0).keys) {
		const index = field(document, key);
		const kind = document.kind(index);
		if (kind === 'null') { continue; }
		if (key === 'cohere' && kind !== 'string') { return unmarshalError(document, index, 'string', 'rawConfig.cohere'); }
		if (key === 'plugins' || key === 'ignorePatterns') {
			if (kind !== 'array') { return unmarshalError(document, index, '[]string', 'rawConfig.' + key); }
			for (let at = 0; at < document.node(index).children.length; at++) {
				const child = document.node(index).children[at] ?? panic('missing list element');
				if (document.kind(child) !== 'string' && document.kind(child) !== 'null') { return `json: cannot unmarshal ${rawType(document, child)} into rawConfig.${key}.${at} of type string`; }
			}
		}
		if (key === 'rules' || key === 'reasons' || key === 'departures') {
			if (kind !== 'object') { return unmarshalError(document, index, key === 'rules' ? 'map[string]jsontext.Value' : 'map[string]string', 'rawConfig.' + key); }
			if (key !== 'rules') {
				for (let at = 0; at < document.node(index).children.length; at++) {
					const child = document.node(index).children[at] ?? panic('missing map value');
					if (document.kind(child) !== 'string' && document.kind(child) !== 'null') { return unmarshalError(document, child, 'string', 'rawConfig.' + key + '.' + (document.node(index).keys[at] ?? '')); }
				}
			}
		}
        if (key === 'overrides') {
            if (kind !== 'array') { return unmarshalError(document, index, '[]configuration.rawOverride', 'rawConfig.overrides'); }
            for (let at = 0; at < document.node(index).children.length; at++) {
                const block = document.node(index).children[at] ?? panic('missing override');
                if (document.kind(block) === 'null') { continue; }
                if (document.kind(block) !== 'object') { return `json: cannot unmarshal ${rawType(document, block)} into rawConfig.overrides.${at} of type configuration.rawOverride`; }
                for (const field of ['files', 'rules', 'reason']) {
                    const value = document.get(block, field);
                    const type = document.kind(value);
                    if (type === 'null' || type === 'missing') { continue; }
                    const target = field === 'files' ? '[]string' : field === 'rules' ? 'map[string]jsontext.Value' : 'string';
                    const expected = field === 'files' ? 'array' : field === 'rules' ? 'object' : 'string';
                    if (type !== expected) { return `json: cannot unmarshal ${rawType(document, value)} into Go struct field rawConfig.overrides.${at}.${field} of type ${target}`; }
                    if (field === 'files') {
                        for (let child = 0; child < document.node(value).children.length; child++) {
                            const item = document.node(value).children[child] ?? panic('missing files element');
                            if (document.kind(item) !== 'string' && document.kind(item) !== 'null') { return `json: cannot unmarshal ${rawType(document, item)} into rawConfig.overrides.${at}.files.${child} of type string`; }
                        }
                    }
                }
            }
        }
	}
	return '';
}
// encoding/json merges duplicate map-valued struct fields rather than replacing their maps.
function mergeDuplicateMaps(document: JsonDocument, parent: number, fields: readonly string[]): void {
    const node = document.node(parent);
    for (const field of fields) {
        let previous = -1;
        for (let at = 0; at < node.keys.length; at++) {
            if (node.keys[at] !== field) { continue; }
            const index = node.children[at] ?? -1;
            if (document.kind(index) !== 'object') { previous = -1; continue; }
            if (previous >= 0) {
                const merged = new JsonNode(); merged.kind = 'object'; merged.raw = document.node(index).raw;
                for (const source of [previous, index]) {
                    const object = document.node(source);
                    for (const key of object.keys) { merged.keys.push(key); }
                    for (const child of object.children) { merged.children.push(child); }
                }
                const mergedIndex = document.nodes.length; document.nodes.push(merged); node.children[at] = mergedIndex; previous = mergedIndex;
            } else { previous = index; }
        }
    }
}
export function readLayers(path: string, chain: readonly string[]): LayersResult {
	const absolute = path.startsWith('cohere:') ? path : clean(path);
	if (chain.includes(absolute)) { return { kind: 'Error', message: `lint config ${absolute} extends itself: ${chain.concat([absolute]).join(' -> ')}` }; }
	let text = '';
	if (absolute.startsWith('cohere:')) {
		text = setText(absolute);
		if (text === '') { return { kind: 'Error', message: `reading lint config ${absolute}: ${absolute} is not a rule set cohere carries; the sets are ${setNames.join(', ')}` }; }
	} else {
		const read = readTextFile(absolute);
		if (read.kind === 'Error') {
			const reason = read.message.endsWith(': no such file') ? 'no such file or directory' : read.message.endsWith(': permission denied') ? 'permission denied' : read.message.endsWith(': is a directory') ? 'is a directory' : 'unrepresented filesystem error';
			return { kind: 'Error', message: `reading lint config ${absolute}: open ${absolute}: ${reason}` };
		}
		text = read.text;
	}
	const document = parseJson(text, false);
	if (document.error !== '') { return { kind: 'Error', message: `parsing lint config ${absolute}: ${document.error}` }; }
	const allowed = ['extends', 'departures', 'reasons', 'plugins', 'rules', 'ignorePatterns', 'overrides', 'cohere', '$schema', 'jsPlugins', 'format', 'output', 'settings'];
	const unknown = nodeKeys(document, 0).filter((key) => !allowed.includes(key));
	if (unknown.length > 0) { return { kind: 'Error', message: `lint config ${absolute} declares ${unknown.map(quoted).join(', ')}, which this loader does not implement: the key would be discarded silently and whatever it configures would never take effect. Either implement it in \`rawConfig\` and act on it, or add it to \`ignoredTopLevelKeys\` with a note saying why ignoring it is correct` }; }
	const overrides = field(document, 'overrides');
	if (document.kind(overrides) === 'array') {
		for (let index = 0; index < document.node(overrides).children.length; index++) {
			const block = document.node(overrides).children[index] ?? -1;
			const unknown = nodeKeys(document, block).filter((key) => !['files', 'rules', 'reason'].includes(key));
			if (unknown.length > 0) { return { kind: 'Error', message: `lint config ${absolute}: override ${index} declares ${unknown.map(quoted).join(', ')}, which this loader does not read: the key would be discarded silently. An override takes "files", "rules" and "reason"` }; }
		}
	}
	const rawError = validateRaw(document);
	if (rawError !== '') { return { kind: 'Error', message: `parsing lint config ${absolute}: ${rawError}` }; }
	mergeDuplicateMaps(document, 0, ['rules', 'reasons', 'departures']);
	if (document.kind(overrides) === 'array') { for (const block of document.node(overrides).children) { mergeDuplicateMaps(document, block, ['rules']); } }
	const layer = new Layer(absolute, document);
	const layers: Layer[] = [];
	const read = new Map<string, boolean>();
	const extendsIndex = field(document, 'extends');
	const extendsKind = document.kind(extendsIndex);
	let bases: string[] = [];
	if (extendsKind === 'string') { const base = document.string(extendsIndex); if (base !== '') { bases = [base]; } }
	else if (extendsKind === 'array') {
		for (const index of document.node(extendsIndex).children) {
			if (document.kind(index) !== 'string' && document.kind(index) !== 'null') { return { kind: 'Error', message: `parsing lint config ${absolute}: "extends" names a path or a rule set, or a list of them: json: cannot unmarshal ${rawType(document, index)} into Go value of type string` }; }
			const base = document.kind(index) === 'null' ? '' : document.string(index);
			if (base === '') { return { kind: 'Error', message: `parsing lint config ${absolute}: "extends" lists an empty entry` }; }
			bases.push(base);
		}
	} else if (extendsKind !== 'missing' && extendsKind !== 'null') { return { kind: 'Error', message: `parsing lint config ${absolute}: "extends" names a path or a rule set, or a list of them: json: cannot unmarshal ${rawType(document, extendsIndex)} into Go value of type []string` }; }
	for (const base of bases) {
		if (absolute.startsWith('cohere:') && !base.startsWith('cohere:')) { return { kind: 'Error', message: `rule set ${absolute} extends ${base}, which is a path: a set may extend only other sets` }; }
		const resolved = base.startsWith('cohere:') || base.startsWith('/') ? base : join(dir(absolute), base);
		const inherited = readLayers(resolved, chain.concat([absolute]));
		if (inherited.kind === 'Error') { return { kind: 'Error', message: `lint config ${absolute} extends ${base}: ${inherited.message}` }; }
		for (const ancestor of inherited.layers) {
			if (!read.has(ancestor.path)) { layers.push(ancestor); read.set(ancestor.path, true); }
		}
		layer.extended.push(inherited.layers[inherited.layers.length - 1]?.path ?? panic('empty base layers'));
	}
	layers.push(layer);
	return { kind: 'Ok', layers };
}
function severity(text: string): string {
	const name = text.toLowerCase();
	if (name === 'off' || name === 'allow') { return 'off'; }
	if (name === 'warn' || name === 'warning') { return 'warn'; }
	if (name === 'error' || name === 'deny') { return 'error'; }
	return '';
}
function truncated(raw: string): string { return raw.length <= 60 ? raw : raw.slice(0, 60) + '...'; }
function parseRule(document: JsonDocument, index: number): RuleResult {
	let name = '';
	let options: string[] = [];
	if (document.kind(index) === 'string' || document.kind(index) === 'null') { name = document.kind(index) === 'null' ? '' : document.string(index); }
	else if (document.kind(index) === 'array') {
		const children = document.node(index).children;
		if (children.length === 0) { return { kind: 'Error', message: 'empty rule configuration array' }; }
		const first = children[0] ?? -1;
		if (document.kind(first) !== 'string' && document.kind(first) !== 'null') { return { kind: 'Error', message: `first element is not a severity string: ${truncated(document.node(first).raw)}` }; }
		name = document.kind(first) === 'null' ? '' : document.string(first);
		options = children.slice(1).map((child) => compact(document.node(child).raw));
	} else { return { kind: 'Error', message: `expected a severity string or a [severity, ...options] array, got ${truncated(document.node(index).raw)}` }; }
	const level = severity(name);
	if (level === '') { return { kind: 'Error', message: `unknown severity ${quoted(name)} (expected off, warn, or error)` }; }
	return { kind: 'Ok', setting: { severity: level, options } };
}
function ancestors(path: string, layers: readonly Layer[], result: Map<string, boolean>): void {
	for (const layer of layers) {
		if (layer.path !== path) { continue; }
		for (const extended of layer.extended) {
			if (!result.has(extended)) { result.set(extended, true); ancestors(extended, layers, result); }
		}
	}
}
function wholeTree(patterns: readonly string[]): boolean {
	if (patterns.length === 0) { return false; }
	for (const pattern of patterns) {
		if (pattern === '**/*') { continue; }
		if (!pattern.startsWith('**/*.')) { return false; }
		let extension = pattern.slice(5);
		if (extension.startsWith('{') && extension.endsWith('}')) { extension = extension.slice(1, -1).split(',').join(''); }
		if (extension === '') { return false; }
		for (const character of extension) { if (!((character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9'))) { return false; } }
	}
	return true;
}
function equalRule(left: RuleSetting, right: RuleSetting): boolean { return left.severity === right.severity && left.options.length === right.options.length && left.options.every((option, index) => option === right.options[index]); }
function reaches(key: string, registered: string): boolean { return key === registered || key.endsWith('/' + registered); }
function inheritedKey(rules: ReadonlyMap<string, RuleSetting>, name: string, registeredNames: readonly string[]): string {
    if (rules.has(name) || registeredNames.length === 0) { return name; }
    for (const key of sortedKeys(rules)) {
        if (registeredNames.some((registered) => reaches(key, registered) && reaches(name, registered))) { return key; }
    }
    return name;
}
function replacesEvery(name: string, inherited: string, registeredNames: readonly string[]): boolean {
    if (name === inherited) { return true; }
    return registeredNames.every((registered) => !reaches(inherited, registered) || reaches(name, registered));
}
export function loadLayers(layers: readonly Layer[], root: string, registeredNames: readonly string[] = []): SettingsResult {
	const settings = new Settings(root);
	const own = layers[layers.length - 1] ?? panic('settings: empty layers');
	const inOurTiers = layers.some((layer) => layer.path.startsWith('cohere:system-inc/'));
	const writers = new Map<string, string>();
	for (let layerIndex = 0; layerIndex < layers.length; layerIndex++) {
		const layer = layers[layerIndex] ?? panic('missing layer');
		const document = layer.document;
		if (layerIndex < layers.length - 1) {
			for (const key of ['cohere', 'settings', 'output']) {
				if (field(document, key) < 0) { continue; }
				if (key === 'cohere') { return { kind: 'Error', message: `lint config ${layer.path} pins "cohere", and it is extended by ${own.path}: only the project's own file may pin a release, since every project extending ${layer.path} would inherit the pin. Move it to the project's own file` }; }
				const explanation = key === 'settings' ? 'the reader of "settings" does not follow `extends`, so the value would be ignored silently.' : 'the run reads "output" from the project\'s own file only, so the value would be ignored silently.';
				return { kind: 'Error', message: `lint config ${layer.path} declares ${quoted(key)}, and it is extended by ${own.path}: ${explanation} Keep ${quoted(key)} in the project's own file` };
			}
		}
		const baseRules = new Map(settings.rules);
		const departed = new Map<string, boolean>();
		const reasoned = inOurTiers || layer.path.startsWith('cohere:');
		const ancestorPaths = new Map<string, boolean>();
		ancestors(layer.path, layers, ancestorPaths);
		const rules = field(document, 'rules');
		const reasons = field(document, 'reasons');
		const departures = field(document, 'departures');
		for (const name of nodeKeys(document, reasons)) {
			const reason = document.string(document.get(reasons, name)).trim();
			if (reason === '') { return { kind: 'Error', message: `lint config ${layer.path} names ${quoted(name)} under "reasons" and gives no reason: say why the rule is off, in a sentence` }; }
			const value = rules < 0 ? -1 : document.get(rules, name);
			const parsed = value < 0 ? { kind: 'Error', message: '' } as const : parseRule(document, value);
			if (parsed.kind !== 'Ok' || parsed.setting.severity !== 'off') { return { kind: 'Error', message: `lint config ${layer.path} names ${quoted(name)} under "reasons", but it does not turn that rule off in its own top-level rules: a reason belongs to the file whose off it explains, so remove the entry` }; }
			if (departures >= 0 && document.get(departures, name) >= 0) { return { kind: 'Error', message: `lint config ${layer.path} names ${quoted(name)} under both "reasons" and "departures": the departure's reason already says why this file turns it off, so remove the "reasons" entry` }; }
		}
		for (const name of nodeKeys(document, rules)) {
			const parsed = parseRule(document, document.get(rules, name));
			if (parsed.kind === 'Error') { return { kind: 'Error', message: `rule ${quoted(name)} in ${layer.path}: ${parsed.message}` }; }
			const inheritedName = inheritedKey(baseRules, name, registeredNames);
			const inherited = baseRules.get(inheritedName);
			let setting = parsed.setting;
			const inheritedReason = settings.offReasons.get(inheritedName);
			if (inherited !== undefined) {
				if (!ancestorPaths.has(writers.get(inheritedName) ?? '')) { return { kind: 'Error', message: `rule ${quoted(name)} is configured by both ${writers.get(inheritedName) ?? ''} and ${layer.path}, and neither extends the other: a rule belongs to one set, so configure it in one of them or in a file that extends both` }; }
				if (setting.options.length === 0) { setting = { severity: setting.severity, options: inherited.options }; }
				if (!equalRule(setting, inherited)) {
					const reason = departures < 0 ? '' : document.string(document.get(departures, name)).trim();
					if (reason !== '') { departed.set(name, true); settings.departures.set(name, { file: layer.path, reason }); }
					else if (reasoned) { return { kind: 'Error', message: `lint config ${layer.path} sets ${quoted(name)} differently from the file it extends and gives no reason: name it under "departures" with why this project differs, or remove the line so the house ruling applies` }; }
				}
				if (replacesEvery(name, inheritedName, registeredNames)) { settings.rules.delete(inheritedName); settings.offReasons.delete(inheritedName); }
			}
			settings.rules.set(name, setting);
			writers.set(name, layer.path);
			const reason = reasons < 0 ? '' : document.string(document.get(reasons, name)).trim();
			if (setting.severity !== 'off') { settings.offReasons.delete(name); }
			else if (reason !== '') { settings.offReasons.set(name, { file: layer.path, reason }); }
			else if (departed.has(name)) { settings.offReasons.set(name, settings.departures.get(name) ?? panic('missing departure')); }
			else if (inheritedReason !== undefined) { settings.offReasons.set(name, inheritedReason); }
			else { settings.offReasons.delete(name); }
		}
		for (const pattern of document.strings(field(document, 'ignorePatterns'))) { settings.ignores.push(pattern); }
		for (const plugin of document.strings(field(document, 'plugins'))) { if (!settings.plugins.includes(plugin)) { settings.plugins.push(plugin); } }
		const overrides = field(document, 'overrides');
		if (document.kind(overrides) === 'array') {
			for (let index = 0; index < document.node(overrides).children.length; index++) {
				const block = document.node(overrides).children[index] ?? panic('missing override');
				const files = document.strings(document.get(block, 'files'));
				const scopedRules = new Map<string, RuleSetting>();
				const rawRules = document.get(block, 'rules');
				for (const name of nodeKeys(document, rawRules)) {
					const parsed = parseRule(document, document.get(rawRules, name));
					if (parsed.kind === 'Error') { return { kind: 'Error', message: `override ${index}, rule ${quoted(name)} in ${layer.path}: ${parsed.message}` }; }
					scopedRules.set(name, parsed.setting);
					const inherited = baseRules.get(inheritedKey(baseRules, name, registeredNames));
					if (!wholeTree(files) || inherited === undefined) { continue; }
					const compared = parsed.setting.options.length === 0 ? { severity: parsed.setting.severity, options: inherited.options } : parsed.setting;
					if (equalRule(compared, inherited)) { continue; }
					const reason = departures < 0 ? '' : document.string(document.get(departures, name)).trim();
					if (reason === '' && !reasoned) { continue; }
					departed.set(name, true);
					if (reason === '') { return { kind: 'Error', message: `lint config ${layer.path} overrides ${quoted(name)} for every file (${files.join(', ')}) differently from the file it extends and gives no reason: an override that matches every file is a top-level rule, so name it under "departures" with why this project differs` }; }
					settings.departures.set(name, { file: layer.path, reason });
				}
				settings.overrides.push({ files, rules: scopedRules, file: layer.path, reason: document.string(document.get(block, 'reason')).trim() });
			}
		}
		for (const name of nodeKeys(document, departures)) { if (!departed.has(name)) { return { kind: 'Error', message: `lint config ${layer.path} names ${quoted(name)} under "departures", but it sets that rule the same as the file it extends, or not at all: remove the entry` }; } }
	}
	const unreasoned: string[] = [];
	for (const name of sortedKeys(settings.rules)) {
		if (settings.rules.get(name)?.severity !== 'off' || settings.offReasons.has(name)) { continue; }
		const writer = writers.get(name) ?? '';
		if (inOurTiers || writer.startsWith('cohere:')) { unreasoned.push(`${quoted(name)} in ${writer}`); }
	}
	unreasoned.sort(compareCodePoints);
	if (unreasoned.length > 0) { return { kind: 'Error', message: `lint config turns off ${unreasoned.join(', ')} with no reason: an off that says nothing reads as an allowance, so name each under "reasons" in the file that turns it off, with why, in a sentence` }; }
	for (const pair of pluginDefaults) {
		const name = pair[0] ?? '';
		const plugin = pair[1] ?? '';
		if ((plugin === 'eslint' || settings.plugins.includes(plugin)) && !settings.rules.has(name)) { settings.rules.set(name, { severity: 'warn', options: [] }); }
	}
	for (let index = layers.length - 1; index >= 0; index--) { settings.sources.push(layers[index]?.path ?? panic('missing source')); }
	if (field(own.document, 'cohere') >= 0) {
		settings.version = fieldText(own.document, 'cohere').trim();
		const error = versionRangeError(settings.version);
		if (error !== '') { return { kind: 'Error', message: `lint config ${own.path}: ${error}` }; }
	}
	return { kind: 'Ok', settings };
}
export function loadSettings(path: string, registeredNames: readonly string[] = []): SettingsResult {
	const layers = readLayers(path, []);
	if (layers.kind === 'Error') { return layers; }
	return loadLayers(layers.layers, dir(clean(path)), registeredNames);
}
export function printSettings(settings: Settings): void {
	console.log(`settings-root ${quoted(settings.root)}`);
	for (const source of settings.sources) { console.log(`source ${quoted(source)}`); }
	for (const pattern of settings.ignores) { console.log(`ignore ${quoted(pattern)}`); }
	for (const plugin of settings.plugins) { console.log(`plugin ${quoted(plugin)}`); }
	for (const name of sortedKeys(settings.rules)) { const setting = settings.rules.get(name) ?? panic('missing rule'); console.log(`rule ${quoted(name)} ${setting.severity} [${setting.options.join(',')}]`); }
	for (const name of sortedKeys(settings.departures)) { const reason = settings.departures.get(name) ?? panic('missing reason'); console.log(`departure ${quoted(name)} ${quoted(reason.file)} ${quoted(reason.reason)}`); }
	for (const name of sortedKeys(settings.offReasons)) { const reason = settings.offReasons.get(name) ?? panic('missing reason'); console.log(`off-reason ${quoted(name)} ${quoted(reason.file)} ${quoted(reason.reason)}`); }
	for (const override of settings.overrides) {
		console.log(`override ${quoted(override.file)} ${quoted(override.reason)} [${override.files.map(quoted).join(',')}]`);
		for (const name of sortedKeys(override.rules)) { const setting = override.rules.get(name) ?? panic('missing override rule'); console.log(`override-rule ${quoted(name)} ${setting.severity} [${setting.options.join(',')}]`); }
	}
	console.log(`cohere-version ${quoted(settings.version)}`);
}

export type StringsResult = { readonly kind: 'Ok'; readonly values: readonly string[] } | { readonly kind: 'Error'; readonly message: string };
// These readers deliberately do not parse rules: a bad rule must not hide formatter/discovery inputs.
export function sourcesOf(path: string): StringsResult {
	const layers = readLayers(path, []);
	if (layers.kind === 'Error') { return layers; }
	const values: string[] = [];
	for (let index = layers.layers.length - 1; index >= 0; index--) { values.push(layers.layers[index]?.path ?? panic('missing source layer')); }
	return { kind: 'Ok', values };
}
export function ignorePatternsOf(path: string): StringsResult {
	const layers = readLayers(path, []);
	if (layers.kind === 'Error') { return layers; }
	const values: string[] = [];
	for (const layer of layers.layers) { for (const value of layer.document.strings(field(layer.document, 'ignorePatterns'))) { values.push(value); } }
	return { kind: 'Ok', values };
}
