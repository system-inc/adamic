// ReadProjectConfig's file discovery and extends. Source files are never decoded by a host bridge.
import { fileStatus, panic, readDirectory, readTextFile, realPath } from 'adamic';
import { clean, dir } from '../gitignore/path.ts';
import { compareCodePoints, join } from '../formatfiles/golang.ts';
import { JsonDocument, parseJson } from './json.ts';
import { absolutePath, TsGlob } from './tsglob.ts';
class Specification {
	files: string[] = [];
	includes: string[] = [];
	excludes: string[] = [];
	sourceExtensions: string[] = [];
	sourceExtensionsPresent = false;
	filesPresent = false;
	includePresent = false;
	excludePresent = false;
	referencesPresent = false;
	extendsPresent = false;
	readonly references: string[] = [];
	readonly options = new Map<string, string>();
	readonly diagnostics: number[] = [];
	gap = '';
}
export interface ProjectConfig { readonly files: readonly string[]; readonly references: readonly string[]; }
export type ProjectResult = { readonly kind: 'Ok'; readonly project: ProjectConfig } | { readonly kind: 'Error'; readonly message: string };
function exists(path: string): boolean { const status = fileStatus(path); return status.kind === 'Ok' && status.type === 'file'; }
function template(path: string, root: string): string { return path.startsWith('${configDir}') ? root + path.slice(12) : path; }
function invalidSpec(pattern: string, disallowTrailing: boolean): number {
    const withoutSlash = pattern.endsWith('/') ? pattern.slice(0, -1) : pattern;
    if (disallowTrailing && (withoutSlash === '**' || withoutSlash.endsWith('/**'))) { return 5010; }
    const segments = pattern.split('/');
    let recursive = false;
    for (const segment of segments) { if (recursive && segment === '..') { return 5065; } if (segment === '**') { recursive = true; } }
    return 0;
}
function specs(document: JsonDocument, index: number, base: string, key: string, diagnostics: number[]): string[] {
	const result: string[] = [];
	if (document.kind(index) !== 'array') { return result; }
	for (const child of document.node(index).children) {
		if (document.kind(child) === 'string') {
			const text = document.string(child);
			const invalid = key === 'files' ? 0 : invalidSpec(text, key === 'include');
			if (invalid !== 0) { diagnostics.push(invalid); continue; }
			result.push(text.startsWith('${configDir}') ? text : absolutePath(base, text));
		}
	}
	return result;
}
function resolveBase(name: string, root: string): string {
	if (name.startsWith('/') || name.startsWith('./') || name.startsWith('../')) {
		const path = absolutePath(root, name);
		return exists(path) || path.endsWith('.json') ? path : path + '.json';
	}
	for (let directory = root; ; directory = dir(directory)) {
		const candidate = join(directory, 'node_modules/' + name);
		if (exists(candidate)) { return candidate; }
		if (exists(candidate + '.json')) { return candidate + '.json'; }
		const manifest = readTextFile(join(candidate, 'package.json'));
		if (manifest.kind === 'Ok') {
			const parsed = parseJson(manifest.text, false);
			const tsconfig = parsed.string(parsed.get(0, 'tsconfig'));
			if (tsconfig !== '' && exists(absolutePath(candidate, tsconfig))) { return absolutePath(candidate, tsconfig); }
		}
		if (exists(join(candidate, 'tsconfig.json'))) { return join(candidate, 'tsconfig.json'); }
		if (dir(directory) === directory) { return ''; }
	}
}
function readSpecification(path: string, chain: readonly string[]): Specification {
	const result = new Specification();
	if (chain.includes(path)) { result.diagnostics.push(18000); return result; }
	const read = readTextFile(path);
	if (read.kind === 'Error') { result.diagnostics.push(5083); return result; }
	const document = parseJson(read.text, true);
	if (document.error !== '') {
		result.gap = `stage1 tsconfig syntax diagnostics not represented: ${path}: ${document.error}`;
		return result;
	}
	const root = dir(path);
	const ownSources = document.get(0, 'sourceExtensions');
	result.sourceExtensionsPresent = ownSources >= 0;
	if (document.kind(ownSources) === 'array') {
		for (const child of document.node(ownSources).children) {
			if (document.kind(child) === 'string') { result.sourceExtensions.push(document.string(child)); }
			else { result.diagnostics.push(5024); }
		}
	} else if (ownSources >= 0 && document.kind(ownSources) !== 'null') { result.diagnostics.push(5024); }
	result.extendsPresent = document.get(0, 'extends') >= 0;
	const options = document.get(0, 'compilerOptions');
	if (document.kind(options) === 'object') {
		for (const key of document.node(options).keys) {
			const index = document.get(options, key);
			const kind = document.kind(index);
			if ((key === 'allowJs' || key === 'checkJs' || key === 'resolveJsonModule') && kind === 'bool') { result.options.set(key, document.string(index)); }
			if ((key === 'outDir' || key === 'declarationDir') && kind === 'string') { const value = document.string(index); result.options.set(key, value.startsWith('${configDir}') ? value : absolutePath(root, value)); }
		}
	}
	const hasOwn = new Map<string, boolean>();
	for (const key of ['files', 'include', 'exclude']) { hasOwn.set(key, document.get(0, key) >= 0); }
	const extendsIndex = document.get(0, 'extends');
	const bases = document.kind(extendsIndex) === 'string' ? [document.string(extendsIndex)] : document.strings(extendsIndex);
	for (const name of bases) {
		const basePath = resolveBase(name, root);
		if (basePath === '') { result.diagnostics.push(6053); continue; }
		const inherited = readSpecification(basePath, chain.concat([path]));
		for (const diagnostic of inherited.diagnostics) { result.diagnostics.push(diagnostic); }
		if (inherited.gap !== '') { result.gap = inherited.gap; }
		if (ownSources < 0 && inherited.sourceExtensionsPresent) { result.sourceExtensions = inherited.sourceExtensions; result.sourceExtensionsPresent = true; }
		if (hasOwn.get('files') !== true && inherited.filesPresent) { result.files = inherited.files; result.filesPresent = true; }
		if (hasOwn.get('include') !== true && inherited.includePresent) { result.includes = inherited.includes; result.includePresent = true; }
		if (hasOwn.get('exclude') !== true && inherited.excludePresent) { result.excludes = inherited.excludes; result.excludePresent = true; }
		for (const [key, value] of inherited.options) { if (document.get(options, key) < 0) { result.options.set(key, value); } }
	}
	for (const key of ['files', 'include', 'exclude']) {
		const index = document.get(0, key);
		if (document.kind(index) !== 'array') { continue; }
		const values = specs(document, index, root, key, result.diagnostics);
		if (key === 'files') { result.files = values; result.filesPresent = true; }
		if (key === 'include') { result.includes = values; result.includePresent = true; }
		if (key === 'exclude') { result.excludes = values; result.excludePresent = true; }
	}
	const references = document.get(0, 'references');
	result.referencesPresent = references >= 0;
	if (document.kind(references) === 'array') {
		for (const child of document.node(references).children) {
			const index = document.get(child, 'path');
			if (document.kind(index) !== 'string' || document.string(index) === '') { result.diagnostics.push(document.string(index) === '' && document.kind(index) === 'string' ? 18051 : 5024); continue; }
			let reference = absolutePath(root, document.string(index));
			if (!reference.endsWith('.json')) { reference = join(reference, 'tsconfig.json'); }
			result.references.push(reference);
		}
	}
	if (document.get(0, 'contentMappers') >= 0) { result.gap = `stage1 tsconfig contentMappers not represented: ${path}`; }
	return result;
}
class Bucket { readonly files: string[] = []; }
function includedIndex(path: string, includes: readonly TsGlob[], excludes: readonly TsGlob[], prefix: boolean): number {
	for (const exclude of excludes) { if (exclude.matches(path, false)) { return -1; } }
	for (let index = 0; index < includes.length; index++) { if ((includes[index] ?? panic('missing glob')).matches(path, prefix)) { return index; } }
	return -1;
}
function visit(path: string, includes: readonly TsGlob[], excludes: readonly TsGlob[], extensions: readonly string[], buckets: readonly Bucket[], visited: Map<string, boolean>): string {
	const resolved = realPath(path);
	const canonical = resolved.kind === 'Ok' ? resolved.path : path;
	if (visited.has(canonical)) { return ''; }
	visited.set(canonical, true);
	const entries = readDirectory(path);
	if (entries.kind === 'Error') { return ''; }
	const directories: string[] = [];
	for (const name of entries.names) {
		const absolute = join(path, name);
		const status = fileStatus(absolute);
		if (status.kind === 'Error') { continue; }
		if (status.type === 'directory') {
			if (includedIndex(absolute, includes, excludes, true) < 0) { continue; }

			directories.push(name);
		} else if (status.type === 'file' && extensions.some((extension) => name.endsWith(extension))) {
			const index = includedIndex(absolute, includes, excludes, false);
			if (index >= 0) { (buckets[index] ?? panic('missing include bucket')).files.push(absolute); }
		}
	}
	for (const name of directories) { const error = visit(join(path, name), includes, excludes, extensions, buckets, visited); if (error !== '') { return error; } }
	return '';
}
function includeBase(path: string): string {
	let wildcard = path.indexOf('*');
	const question = path.indexOf('?');
	if (wildcard < 0 || (question >= 0 && question < wildcard)) { wildcard = question; }
	if (wildcard >= 0) { const prefix = path.slice(0, path.slice(0, wildcard + 1).lastIndexOf('/')); return prefix === '' ? '/' : prefix; }
	return (path.slice(path.lastIndexOf('/') + 1).includes('.')) ? dir(path) : path;
}
function extensionGroup(file: string, groups: readonly (readonly string[])[]): readonly string[] {
	for (const group of groups) { if (group.some((extension) => file.endsWith(extension))) { return group; } }
	return [];
}
function extension(file: string): string {
	for (const suffix of ['.d.ts', '.d.mts', '.d.cts', '.tsx', '.ts', '.mts', '.cts', '.jsx', '.js', '.mjs', '.cjs']) { if (file.endsWith(suffix)) { return suffix; } }
	return '';
}
function higherPriority(file: string, groups: readonly (readonly string[])[], literal: ReadonlyMap<string, string>, wildcard: ReadonlyMap<string, string>): boolean {
	const suffix = extension(file);
	const stem = file.slice(0, file.length - suffix.length);
	for (const candidate of extensionGroup(file, groups)) {
		if (file.endsWith(candidate) && (candidate !== '.ts' || !file.endsWith('.d.ts'))) { return false; }
		if (candidate === '.d.ts' && (suffix === '.js' || suffix === '.jsx')) { continue; }
		if (literal.has(stem + candidate) || wildcard.has(stem + candidate)) { return true; }
	}
	return false;
}
export function readProjectConfig(configPath: string): ProjectResult {
	const path = clean(configPath);
	if (!exists(path)) { return { kind: 'Error', message: `no tsconfig at ${path}` }; }
	const root = dir(path);
	const specification = readSpecification(path, []);
	if (specification.gap !== '') { return { kind: 'Error', message: specification.gap }; }
	if (specification.filesPresent && specification.files.length === 0 && !specification.referencesPresent && !specification.extendsPresent) { specification.diagnostics.push(18002); }
	if (!specification.filesPresent && !specification.includePresent) { specification.includes = ['**/*']; }
	const includes = specification.includes.map((pattern) => template(pattern, root));
	let excludes = specification.excludes.map((pattern) => template(pattern, root));
	if (!specification.excludePresent) { excludes = ['node_modules', 'bower_components', 'jspm_packages']; for (const key of ['outDir', 'declarationDir']) { const value = specification.options.get(key) ?? ''; if (value !== '') { excludes.push(template(value, root)); } } }

	const includeGlobs = includes.map((pattern) => new TsGlob(pattern, root, false));
	const excludeGlobs = excludes.map((pattern) => new TsGlob(pattern, root, true));
	const allowJs = specification.options.get('allowJs') === 'true' || (!specification.options.has('allowJs') && specification.options.get('checkJs') === 'true');
	const groups: (readonly string[])[] = allowJs ? [['.ts', '.tsx', '.d.ts', '.js', '.jsx'], ['.cts', '.d.cts', '.cjs'], ['.mts', '.d.mts', '.mjs']] : [['.ts', '.tsx', '.d.ts'], ['.cts', '.d.cts'], ['.mts', '.d.mts']];
	const extensions: string[] = [];
	for (const group of groups) { for (const extension of group) { extensions.push(extension); } }
	// The pinned TypeScript reader opts extra extensions in per config, inherits
	// the list through extends, and never groups them with built-in priorities.
	const builtinExtensions = ['.ts', '.tsx', '.d.ts', '.cts', '.d.cts', '.mts', '.d.mts', '.js', '.jsx', '.cjs', '.mjs', '.json'];
	const seenSources = new Map<string, boolean>();
	for (const suffix of specification.sourceExtensions) {
		if (!suffix.startsWith('.') || suffix.length < 2 || builtinExtensions.includes(suffix.toLowerCase()) || seenSources.has(suffix)) { specification.diagnostics.push(6046); continue; }
		seenSources.set(suffix, true);
		if (!extensions.includes(suffix)) { groups.push([suffix]); extensions.push(suffix); }
	}

	const json = specification.options.get('resolveJsonModule') === 'true';
	if (json) { extensions.push('.json'); }
	const buckets = includeGlobs.map(() => new Bucket());
	const bases: string[] = [root];
	for (const base of includes.map((pattern) => includeBase(absolutePath(root, pattern))).sort(compareCodePoints)) { if (!bases.some((parent) => base === parent || base.startsWith(parent + '/'))) { bases.push(base); } }
	const visited = new Map<string, boolean>();
	if (includes.length > 0) { for (const base of bases) { const error = visit(base, includeGlobs, excludeGlobs, extensions, buckets, visited); if (error !== '') { return { kind: 'Error', message: error }; } } }
	const literal = new Map<string, string>();
	for (const file of specification.files) { const normalized = absolutePath(root, template(file, root)); if (!literal.has(normalized)) { literal.set(normalized, normalized); } }
	const wildcard = new Map<string, string>();
	const jsonFiles = new Map<string, string>();
	for (const bucket of buckets) {
		for (const file of bucket.files) {
			if (file.endsWith('.json')) {
				if (includes.some((pattern) => pattern.endsWith('.json') && new TsGlob(pattern, root, false).matches(file, false)) && !literal.has(file)) { jsonFiles.set(file, file); }
				continue;
			}
			if (higherPriority(file, groups, literal, wildcard)) { continue; }
			const group = extensionGroup(file, groups);
			const suffix = extension(file);
			const stem = file.slice(0, file.length - suffix.length);
			let remove = false;
			for (const candidate of group) { if (remove) { wildcard.delete(stem + candidate); } if (candidate === suffix) { remove = true; } }
			if (!literal.has(file)) { wildcard.set(file, file); }
		}
	}
	const files = [...literal.values()];
	for (const file of wildcard.values()) { files.push(file); }
	for (const file of jsonFiles.values()) { files.push(file); }
	if (files.length === 0 && !specification.filesPresent && !specification.referencesPresent) { specification.diagnostics.push(18003); }
	const errors = specification.diagnostics;
	// Go ReadProjectConfig permits option-only findings while retaining valid
	// values; cohere reports those findings later in its options phase.
	if (errors.length > 0 && !errors.every((code) => [5023, 5024, 5025, 6046].includes(code))) { return { kind: 'Error', message: `reading ${path}: ${errors.slice(0, 5).map((code) => `TS${code}: `).join('\n')}${errors.length > 5 ? `\nand ${errors.length - 5} more` : ''}` }; }
	return { kind: 'Ok', project: { files, references: specification.references } };
}
