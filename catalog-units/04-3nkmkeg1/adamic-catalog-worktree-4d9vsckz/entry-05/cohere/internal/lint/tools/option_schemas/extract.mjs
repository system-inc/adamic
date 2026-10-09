/*
 * Write every ported rule's upstream options schema, as ESLint normalizes it, and the samples that hold
 * cohere's validator to ajv's verdict (#pd2chkx).
 *
 *     node internal/lint/tools/option_schemas/extract.mjs \
 *         --home <directory> [--home <directory> ...] \
 *         --schemas internal/lint/optionschema/schemas.json \
 *         --samples internal/lint/optionschema/testdata/samples.json
 *
 * Each `--home` is a directory whose node_modules resolves some of the plugins. A package is taken from
 * the first home that has it, so ahra serves ESLint and every plugin but boundaries, which only
 * api-phi-health installs. The rule list is cohere's own docs/data/rules.json: every row whose origin is
 * an ESLint plugin or ESLint core.
 *
 * This is a regeneration tool, and nothing runs it. Its two outputs are committed. schemas.json is
 * embedded in the binary and is what OptionsRegistry.Decode validates against; samples.json is test data,
 * and `TestValidatorAgreesWithAjv` replays it. Regenerating for a new plugin version is a deliberate act:
 * the file records the version of every package it read, and `TestSchemasCoverEveryPortedRule` refuses a
 * file that is missing a ported rule or holds one cohere no longer registers.
 *
 * # What "as ESLint normalizes it" means
 *
 * ESLint's `getRuleOptionsSchema` (lib/config/config.js): no `meta.schema`, or an empty array, is
 * `{type: "array", minItems: 0, maxItems: 0}`; an array of element schemas is wrapped as
 * `{type: "array", items: [...], minItems: 0, maxItems: n}`; an object is used as written; `false`
 * opts out of validation. The schema is the JSON a plugin's own `meta.schema` serializes to, so a
 * plugin's schema written with functions or regular expressions in it would not survive; the tool
 * refuses one rather than writing a schema that says less than the plugin's.
 *
 * # The samples
 *
 * Two generators, one per direction of the agreement gate.
 *
 * Accepted shapes are drawn as the first sweep drew them (#e06zm4b): every enum value, each property
 * alone, each anyOf and oneOf branch, an empty object, each element position, kept where ajv accepts,
 * each both raw and with ajv's defaults filled in.
 *
 * Refused shapes are option lists valid except for one keyword at one place: a required key removed, a
 * number under its minimum or over its maximum, an array under minItems or over maxItems or holding a
 * duplicate where uniqueItems is set, a string under minLength or over maxLength or failing its pattern,
 * an enum value that is not one (unknown, and each listed string with its first letter's case flipped),
 * and a value of every other type. Kept where ajv refuses.
 *
 * Each sample records ajv's verdict, which is the only thing the test reads as an answer. ajv is
 * configured as ESLint 10's lib/shared/ajv.js configures it.
 */

import { existsSync, readFileSync, readdirSync, realpathSync, writeFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import * as NodePath from 'node:path';
import { pathToFileURL } from 'node:url';
import { parseArgs } from 'node:util';

const { values: argumentValues } = parseArgs({
    options: {
        home: { type: 'string', multiple: true },
        rules: { type: 'string', default: 'docs/data/rules.json' },
        schemas: { type: 'string' },
        samples: { type: 'string' },
    },
});
if(!argumentValues.home || !argumentValues.schemas || !argumentValues.samples) {
    throw new Error('name --home (directories that resolve ESLint and its plugins), --schemas and --samples');
}

// packageDirectory finds a package in the first home that installs it, hoisted or in pnpm's store
function packageDirectory(name) {
    for(const home of argumentValues.home) {
        const hoisted = NodePath.join(home, 'node_modules', name);
        if(existsSync(NodePath.join(hoisted, 'package.json'))) return realpathSync(hoisted);
        const store = NodePath.join(home, 'node_modules', '.pnpm');
        if(!existsSync(store)) continue;
        const prefix = `${name.replace('/', '+')}@`;
        const entry = readdirSync(store).filter((candidate) => candidate.startsWith(prefix)).sort().pop();
        if(entry) return realpathSync(NodePath.join(store, entry, 'node_modules', name));
    }
    throw new Error(`no --home installs ${name}`);
}
const versions = {};
function packageRequire(name) {
    const directory = packageDirectory(name);
    const required = createRequire(NodePath.join(directory, 'package.json'));
    versions[name] = required('./package.json').version;
    return { required, directory };
}

// loadPlugin returns a package's rules, through require or, for an ES module, import
async function loadPlugin(name) {
    const { required, directory } = packageRequire(name);
    let loaded;
    try {
        // By its own name, which a package with an exports map resolves to itself; by directory otherwise
        try {
            loaded = required(name);
        } catch(error) {
            if(error.code !== 'MODULE_NOT_FOUND' && error.code !== 'ERR_PACKAGE_PATH_NOT_EXPORTED') throw error;
            loaded = required(directory);
        }
    } catch(error) {
        if(error.code !== 'ERR_REQUIRE_ESM') throw error;
        const manifest = JSON.parse(readFileSync(NodePath.join(directory, 'package.json'), 'utf8'));
        const pick = (value) => (typeof value === 'string' ? value : value && (pick(value.import) || pick(value.default)));
        const entry = pick(manifest.exports && (manifest.exports['.'] ?? manifest.exports)) ?? manifest.main ?? 'index.js';
        loaded = await import(pathToFileURL(NodePath.join(directory, entry)).href);
    }
    const plugin = loaded.default ?? loaded;
    return plugin.rules ?? plugin.default?.rules;
}

// Each origin, the package that holds its rules, and the prefix cohere writes before an upstream name
const origins = {
    core: { package: 'eslint', prefix: '' },
    'typescript-eslint': { package: '@typescript-eslint/eslint-plugin', prefix: '@typescript-eslint/' },
    react: { package: 'eslint-plugin-react', prefix: 'react/' },
    'react-hooks': { package: 'eslint-plugin-react-hooks', prefix: 'react-hooks/' },
    next: { package: '@next/eslint-plugin-next', prefix: '@next/next/' },
    tailwind: { package: 'eslint-plugin-better-tailwindcss', prefix: 'better-tailwindcss/' },
    'eslint-comments': { package: '@eslint-community/eslint-plugin-eslint-comments', prefix: '@eslint-community/eslint-comments/' },
    boundaries: { package: 'eslint-plugin-boundaries', prefix: 'boundaries/' },
};

const { Ajv, draftFour } = (() => {
    const { required } = packageRequire('eslint');
    const ajvDirectory = NodePath.dirname(required.resolve('ajv/package.json'));
    const ajvRequire = createRequire(NodePath.join(ajvDirectory, 'package.json'));
    versions.ajv = ajvRequire('./package.json').version;
    return { Ajv: ajvRequire(ajvDirectory), draftFour: ajvRequire('./lib/refs/json-schema-draft-04.json') };
})();

// newValidator compiles a schema as ESLint 10's lib/shared/ajv.js does
function newValidator(schema) {
    const ajv = new Ajv({ meta: false, useDefaults: true, validateSchema: false, missingRefs: 'ignore', verbose: true, schemaId: 'auto' });
    ajv.addMetaSchema(draftFour);
    ajv._opts.defaultMeta = draftFour.id;
    return ajv.compile(schema);
}

// normalizedSchema is ESLint's getRuleOptionsSchema, with false kept as false
function normalizedSchema(rule) {
    const schema = rule.meta?.schema;
    if(schema === undefined) return { type: 'array', minItems: 0, maxItems: 0 };
    if(schema === false) return false;
    if(Array.isArray(schema)) {
        if(!schema.length) return { type: 'array', minItems: 0, maxItems: 0 };
        return { type: 'array', items: schema, minItems: 0, maxItems: schema.length };
    }
    return schema;
}

// Refuses a schema JSON cannot carry whole
function assertSerializable(name, value, at = '') {
    if(value === null || ['string', 'number', 'boolean'].includes(typeof value)) return;
    if(Array.isArray(value)) return value.forEach((item, index) => assertSerializable(name, item, `${at}[${index}]`));
    if(typeof value === 'object' && Object.getPrototypeOf(value) === Object.prototype) {
        for(const [key, item] of Object.entries(value)) assertSerializable(name, item, `${at}.${key}`);
        return;
    }
    throw new Error(`${name}: the schema holds a ${Object.prototype.toString.call(value)} at ${at || 'its root'}, which JSON cannot carry`);
}

function resolveReference(root, reference) {
    if(!reference.startsWith('#/')) return undefined;
    return reference.slice(2).split('/').reduce((node, key) => (node ? node[key.replace(/~1/gu, '/').replace(/~0/gu, '~')] : undefined), root);
}
function dereference(schema, root) {
    for(let hops = 0; schema && schema.$ref && hops < 10; hops++) schema = resolveReference(root, schema.$ref);
    return schema;
}

// acceptedCandidates is the first sweep's generator: values the schema describes, from its own words
function acceptedCandidates(schema, root, depth) {
    schema = dereference(schema, root);
    if(schema === undefined || schema === true || depth > 4) return ['x', true, 1, {}];
    if(schema === false) return [];
    if(schema.enum) return schema.enum.slice();
    if(schema.const !== undefined) return [schema.const];
    const values = [];
    for(const key of ['anyOf', 'oneOf']) {
        if(schema[key]) for(const branch of schema[key]) values.push(...acceptedCandidates(branch, root, depth + 1));
    }
    if(schema.allOf) for(const branch of schema.allOf) values.push(...acceptedCandidates(branch, root, depth + 1));
    let types = schema.type ? [].concat(schema.type) : [];
    if(!types.length && (schema.properties || schema.additionalProperties)) types = ['object'];
    if(!types.length && schema.items) types = ['array'];
    for(const type of types) {
        if(type === 'string') values.push(schema.default !== undefined ? schema.default : 'x', 'x', '^x$');
        else if(type === 'boolean') values.push(true, false);
        else if(type === 'integer' || type === 'number') values.push(schema.minimum !== undefined ? schema.minimum : 0, 1, 2);
        else if(type === 'null') values.push(null);
        else if(type === 'array') {
            values.push([]);
            const itemSchema = Array.isArray(schema.items) ? schema.items[0] : schema.items;
            for(const item of acceptedCandidates(itemSchema, root, depth + 1).slice(0, 3)) values.push([item]);
        } else if(type === 'object') {
            values.push({});
            for(const [key, property] of Object.entries(schema.properties ?? {})) {
                for(const value of acceptedCandidates(property, root, depth + 1).slice(0, 6)) values.push({ [key]: value });
            }
            if(schema.additionalProperties && typeof schema.additionalProperties === 'object') {
                for(const value of acceptedCandidates(schema.additionalProperties, root, depth + 1).slice(0, 2)) values.push({ someKey: value });
            }
        }
    }
    return values;
}

// validValues is a few values valid against a schema, used to fill what a refused shape keeps valid
function validValues(schema, root, depth) {
    schema = dereference(schema, root);
    if(schema === undefined || schema === true || depth > 5) return ['x', true, 1, {}];
    if(schema === false) return [];
    if(schema.enum) return schema.enum.slice(0, 3);
    if(schema.const !== undefined) return [schema.const];
    const values = [];
    for(const key of ['anyOf', 'oneOf']) if(schema[key]) for(const branch of schema[key]) values.push(...validValues(branch, root, depth + 1));
    if(schema.allOf) values.push(...validValues(Object.assign({}, ...schema.allOf.map((branch) => dereference(branch, root))), root, depth + 1));
    let types = schema.type ? [].concat(schema.type) : [];
    if(!types.length && (schema.properties || schema.additionalProperties || schema.required)) types = ['object'];
    if(!types.length && schema.items) types = ['array'];
    for(const type of types) {
        if(type === 'string') values.push(schema.default !== undefined ? schema.default : 'x', 'abc');
        else if(type === 'boolean') values.push(true);
        else if(type === 'integer' || type === 'number') values.push(schema.minimum !== undefined ? schema.minimum + (schema.exclusiveMinimum ? 1 : 0) : 1);
        else if(type === 'null') values.push(null);
        else if(type === 'array') {
            const itemSchema = Array.isArray(schema.items) ? schema.items[0] : schema.items;
            const items = validValues(itemSchema, root, depth + 1);
            const array = [];
            for(let index = 0; index < (schema.minItems ?? 0); index++) array.push(items[index % Math.max(items.length, 1)]);
            values.push(array);
            if(items.length) values.push([items[0]]);
        } else if(type === 'object') {
            const object = {};
            for(const key of schema.required ?? []) object[key] = validValues(schema.properties?.[key], root, depth + 1)[0];
            values.push(object);
        }
    }
    return values;
}

const typeWitnesses = { string: 'x', boolean: true, number: 1.5, integer: 1, array: [], object: {}, null: null };

// refusedCandidates are values invalid through one keyword at one place, each with the keyword and path
function refusedCandidates(schema, root, depth, at) {
    schema = dereference(schema, root);
    if(!schema || schema === true || depth > 5) return [];
    const out = [];
    const push = (keyword, value, where = at) => out.push({ keyword, path: where, value });

    if(schema.enum) {
        push('enum', 'notAnUpstreamValue');
        for(const value of schema.enum.filter((item) => typeof item === 'string' && item !== '').slice(0, 6)) {
            const first = value[0];
            const flipped = (first === first.toUpperCase() ? first.toLowerCase() : first.toUpperCase()) + value.slice(1);
            if(!schema.enum.includes(flipped)) push('enum', flipped);
        }
        if(schema.enum.some((item) => typeof item === 'number')) push('enum', 987654);
    }
    for(const key of ['anyOf', 'oneOf']) if(schema[key]) schema[key].forEach((branch, index) => out.push(...refusedCandidates(branch, root, depth + 1, `${at}/${key}[${index}]`)));
    if(schema.allOf) schema.allOf.forEach((branch, index) => out.push(...refusedCandidates(branch, root, depth + 1, `${at}/allOf[${index}]`)));

    const types = schema.type ? [].concat(schema.type) : [];
    if(types.length) {
        for(const [type, witness] of Object.entries(typeWitnesses)) {
            if(!types.includes(type) && !(type === 'integer' && types.includes('number'))) push('type', witness);
        }
        if(types.includes('integer') && !types.includes('number')) push('type', 1.5);
    }
    if(typeof schema.minimum === 'number') push('minimum', schema.exclusiveMinimum ? schema.minimum : schema.minimum - 1);
    if(typeof schema.maximum === 'number') push('maximum', schema.exclusiveMaximum ? schema.maximum : schema.maximum + 1);
    if(typeof schema.minLength === 'number' && schema.minLength > 0) push('minLength', 'x'.repeat(schema.minLength - 1));
    if(typeof schema.maxLength === 'number') push('maxLength', 'x'.repeat(schema.maxLength + 1));
    if(typeof schema.pattern === 'string') {
        const pattern = new RegExp(schema.pattern, 'u');
        const failing = ['', ' ', '!!!', 'x y', '0', 'a', 'A', '-', '.', 'notMatching'].find((candidate) => !pattern.test(candidate));
        if(failing !== undefined) push('pattern', failing);
    }

    if(types.includes('array') || (!types.length && schema.items)) {
        const itemSchema = Array.isArray(schema.items) ? schema.items[0] : schema.items;
        const items = validValues(itemSchema, root, depth + 1);
        const base = validValues(schema, root, depth + 1).find((value) => Array.isArray(value)) ?? [];
        if(typeof schema.minItems === 'number' && schema.minItems > 0) push('minItems', base.slice(0, schema.minItems - 1));
        if(typeof schema.maxItems === 'number') {
            const longer = [];
            for(let index = 0; index <= schema.maxItems; index++) longer.push(items[index % Math.max(items.length, 1)]);
            push('maxItems', longer);
        }
        if(schema.uniqueItems && items.length) {
            const padded = base.length ? base.slice() : [items[0]];
            push('uniqueItems', [...padded, padded[0]]);
        }
        if(itemSchema) {
            for(const refused of refusedCandidates(itemSchema, root, depth + 1, `${at}[]`)) {
                const filled = base.slice();
                filled[0] = refused.value;
                out.push({ ...refused, value: filled });
            }
        }
    }

    if(types.includes('object') || (!types.length && (schema.properties || schema.required || schema.additionalProperties))) {
        const base = {};
        for(const key of schema.required ?? []) base[key] = validValues(schema.properties?.[key], root, depth + 1)[0];
        for(const key of schema.required ?? []) {
            const without = { ...base };
            delete without[key];
            push('required', without, `${at}.${key}`);
        }
        if(typeof schema.minProperties === 'number' && schema.minProperties > 0 && !(schema.required ?? []).length) push('minProperties', {});
        if(schema.additionalProperties === false) push('additionalProperties', { ...base, notAnUpstreamKey: true }, `${at}.notAnUpstreamKey`);
        for(const [key, property] of Object.entries(schema.properties ?? {})) {
            for(const refused of refusedCandidates(property, root, depth + 1, `${at}.${key}`)) out.push({ ...refused, value: { ...base, [key]: refused.value } });
        }
        if(schema.additionalProperties && typeof schema.additionalProperties === 'object') {
            for(const refused of refusedCandidates(schema.additionalProperties, root, depth + 1, `${at}.<additional>`)) out.push({ ...refused, value: { ...base, someKey: refused.value } });
        }
    }
    return out;
}

// candidateLists turns per-element candidates into whole option lists, earlier elements filled valid
function candidateLists(schema, generate) {
    const elements = Array.isArray(schema.items) ? schema.items : null;
    const lists = [];
    if(elements) {
        const firsts = elements.map((element) => validValues(element, schema, 0)[0]);
        elements.forEach((element, index) => {
            for(const candidate of generate(element, schema, index)) lists.push({ ...candidate, list: [...firsts.slice(0, index), candidate.value] });
        });
    }
    for(const candidate of generate(schema, schema, null)) if(Array.isArray(candidate.value)) lists.push({ ...candidate, list: candidate.value });
    return lists;
}

const rows = JSON.parse(readFileSync(argumentValues.rules, 'utf8')).rules;
const pluginRules = {};
const schemas = {};
const samples = [];
for(const row of rows.sort((left, right) => (left.name < right.name ? -1 : 1))) {
    const origin = origins[row.origin];
    if(!origin) continue;
    if(!pluginRules[row.origin]) {
        pluginRules[row.origin] = row.origin === 'core'
            ? Object.fromEntries(packageRequire('eslint').required('eslint/use-at-your-own-risk').builtinRules)
            : await loadPlugin(origin.package);
    }
    if(!row.name.startsWith(origin.prefix)) throw new Error(`${row.name} is a ${row.origin} rule without the ${origin.prefix} prefix`);
    const upstream = pluginRules[row.origin][row.name.slice(origin.prefix.length)];
    if(!upstream) throw new Error(`${origin.package} has no rule ${row.name.slice(origin.prefix.length)}`);
    const schema = normalizedSchema(upstream);
    assertSerializable(row.name, schema);
    schemas[row.name] = schema;
    if(schema === false) continue;

    const validate = newValidator(schema);
    const verdict = (list) => (validate(structuredClone(list)) ? 'accepts' : 'refuses');
    const seen = new Set();
    const add = (sample) => {
        const key = JSON.stringify(sample.elements);
        if(seen.has(key)) return;
        seen.add(key);
        samples.push({ rule: row.name, ...sample, eslint: verdict(sample.elements) });
    };
    for(const { list } of candidateLists(schema, (element, root) => acceptedCandidates(element, root, 0).map((value) => ({ value })))) {
        const withDefaults = structuredClone(list);
        if(!validate(withDefaults)) continue;
        add({ elements: list });
        add({ elements: withDefaults });
    }
    for(const { list, keyword, path } of candidateLists(schema, (element, root, index) => refusedCandidates(element, root, 0, index === null ? '' : `[${index}]`))) {
        if(verdict(list) === 'refuses') add({ elements: list, keyword, path });
    }
}

// One rule per line, so a regeneration's diff names the rules whose schemas moved
const sortedVersions = Object.fromEntries(Object.entries(versions).sort());
const schemaLines = Object.entries(schemas).map(([name, schema]) => `        ${JSON.stringify(name)}: ${JSON.stringify(schema)}`);
writeFileSync(argumentValues.schemas, `{\n    "generatedFrom": ${JSON.stringify(sortedVersions)},\n    "rules": {\n${schemaLines.join(',\n')}\n    }\n}\n`);
const sampleLines = samples.map((sample) => `        ${JSON.stringify(sample)}`);
writeFileSync(argumentValues.samples, `{\n    "generatedFrom": ${JSON.stringify(sortedVersions)},\n    "samples": [\n${sampleLines.join(',\n')}\n    ]\n}\n`);

const counts = { accepts: 0, refuses: 0 };
for(const sample of samples) counts[sample.eslint]++;
console.log(`${Object.keys(schemas).length} rules, ${samples.length} samples: ajv accepts ${counts.accepts}, refuses ${counts.refuses}`);
console.log(JSON.stringify(sortedVersions));
