// House settings reader. Program-derived detection is a typed input, just as in Go LoadHouse.
import { fileStatus, panic } from 'adamic';
import { clean, dir } from '../gitignore/path.ts';
import { LintGlob } from './glob.ts';
import { Layer, loadLayers, readLayers, type RuleSetting, type Settings } from './settings.ts';
export interface HouseDetection { readonly reactFiles: readonly string[]; readonly nextFiles: readonly string[]; readonly tailwindEntryPoint: string; readonly tailwindSkipped: string; }
export class HouseSettings {
	readonly whole: Settings;
	readonly detection: HouseDetection;
	readonly variants: ReadonlyMap<string, Settings>;
	constructor(whole: Settings, detection: HouseDetection, variants: ReadonlyMap<string, Settings>) { this.whole = whole; this.detection = detection; this.variants = variants; }
}
export type HouseResult = { readonly kind: 'Ok'; readonly house: HouseSettings } | { readonly kind: 'Error'; readonly message: string };
function variantLayers(ownLayers: readonly Layer[], sets: readonly string[]): { readonly kind: 'Ok'; readonly layers: readonly Layer[] } | { readonly kind: 'Error'; readonly message: string } {
	const layers: Layer[] = [];
	const seen = new Map<string, boolean>();
	for (const set of sets) {
		const base = readLayers(set, []);
		if (base.kind === 'Error') { return base; }
		for (const layer of base.layers) { if (!seen.has(layer.path)) { layers.push(layer); seen.set(layer.path, true); } }
	}
	for (const layer of ownLayers) {
		const own = new Layer(layer.path, layer.document);
		for (const base of layer.extended) { own.extended.push(base); }
		for (const set of sets) { own.extended.push(set); }
		layers.push(own);
	}
	return { kind: 'Ok', layers };
}
export function loadHouse(path: string, detection: HouseDetection, registeredNames: readonly string[] = []): HouseResult {
	let ownLayers: readonly Layer[] = [];
	if (fileStatus(path).kind === 'Ok') {
		const own = readLayers(path, []);
		if (own.kind === 'Error') { return own; }
		ownLayers = own.layers;
	}
	const root = dir(clean(path));
	const sets = ['cohere:typescript', 'cohere:react', 'cohere:next', 'cohere:tailwind'];
	const wholeLayers = variantLayers(ownLayers, sets);
	if (wholeLayers.kind === 'Error') { return wholeLayers; }
	const whole = loadLayers(wholeLayers.layers, root, registeredNames);
	if (whole.kind === 'Error') { return whole; }
	const variants = new Map<string, Settings>();
	for (const key of ['00', '10', '01', '11']) {
		const selected = ['cohere:typescript'];
		if (key[0] === '1') { selected.push('cohere:react'); }
		if (key[1] === '1') { selected.push('cohere:next'); }
		if (detection.tailwindEntryPoint !== '') { selected.push('cohere:tailwind'); }
		const layers = variantLayers(ownLayers, selected);
		if (layers.kind === 'Error') { return layers; }
		const loaded = loadLayers(layers.layers, root, registeredNames);
		if (loaded.kind === 'Error') { return loaded; }
		const variant = loaded.settings;
		variant.ignores.splice(0, variant.ignores.length);
		for (const pattern of whole.settings.ignores) { variant.ignores.push(pattern); }
		const have = new Map<string, boolean>();
		for (const name of variant.rules.keys()) { have.set(name, true); }
		for (const override of variant.overrides) { for (const name of override.rules.keys()) { have.set(name, true); } }
		for (const name of whole.settings.rules.keys()) { if (!have.has(name)) { variant.rules.set(name, { severity: 'off', options: [] }); } }
		for (const override of whole.settings.overrides) { for (const name of override.rules.keys()) { if (!have.has(name)) { variant.rules.set(name, { severity: 'off', options: [] }); } } }
		variants.set(key, variant);
	}
	return { kind: 'Ok', house: new HouseSettings(whole.settings, detection, variants) };
}
export interface Resolved { readonly ignored: boolean; readonly ignoredBy: string; readonly rules: ReadonlyMap<string, RuleSetting>; }
export function relativePath(root: string, path: string): string {
	if (!path.startsWith('/')) { return path.startsWith('./') ? path.slice(2) : path; }
	const left = clean(root).split('/');
	const right = clean(path).split('/');
	let common = 0;
	while (common < left.length && common < right.length && left[common] === right[common]) { common++; }
	const parts: string[] = [];
	for (let index = common; index < left.length; index++) { parts.push('..'); }
	for (const part of right.slice(common)) { parts.push(part); }
	return parts.length === 0 ? '.' : parts.join('/');
}
export function resolveSettings(settings: Settings, path: string): Resolved {
	const relative = relativePath(settings.root, path);
	for (const pattern of settings.ignores) { if (new LintGlob(pattern).matches(relative)) { return { ignored: true, ignoredBy: pattern, rules: new Map<string, RuleSetting>() }; } }
	const effective = new Map(settings.rules);
	for (const override of settings.overrides) {
		if (!override.files.some((pattern) => new LintGlob(pattern).matches(relative))) { continue; }
		for (const [name, setting] of override.rules) {
			const inForce = effective.get(name);
			effective.set(name, setting.options.length === 0 && inForce !== undefined ? { severity: setting.severity, options: inForce.options } : setting);
		}
	}
	return { ignored: false, ignoredBy: '', rules: effective };
}
export function resolveHouse(house: HouseSettings, path: string): Resolved {
	const key = (house.detection.reactFiles.includes(path) ? '1' : '0') + (house.detection.nextFiles.includes(path) ? '1' : '0');
	return resolveSettings(house.variants.get(key) ?? panic('missing house variant'), path);
}
