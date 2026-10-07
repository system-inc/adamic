import { utf8Length } from 'adamic';
export class Alias {
    readonly directory: string; readonly alias: string;
    constructor(directory: string, alias: string) { this.directory = normalizeDirectory(directory); this.alias = alias; }
}
export function normalizeDirectory(directory: string): string {
    let result = directory;
    while(result.startsWith('./')) { result = result.slice(2); }
    while(result.endsWith('/')) { result = result.slice(0, -1); }
    return result === '.' ? '' : result;
}
export function contains(directory: string, path: string): boolean { return directory === '' || path === directory || path.startsWith(directory + '/'); }
export function relative(root: string, path: string): string | undefined {
    if(root === path) { return ''; }
    return path.startsWith(root + '/') ? path.slice(root.length + 1) : undefined;
}
export function ordered(aliases: readonly Alias[]): Alias[] {
    const result = aliases.slice(); result.sort((a, b) => utf8Length(b.directory) - utf8Length(a.directory)); return result;
}
export function aliasFor(aliases: readonly Alias[], path: string): string | undefined {
    for(const alias of aliases) {
        if(!contains(alias.directory, path)) { continue; }
        let rest = alias.directory === '' ? path : path.slice(alias.directory.length);
        if(rest.startsWith('/')) { rest = rest.slice(1); }
        return rest === '' ? alias.alias : alias.alias + '/' + rest;
    }
    return undefined;
}
export function quotedLike(original: string, replacement: string): string {
    const first = original.slice(0, 1);
    const quote = first === '"' || first === "'" || first === '`' ? first : "'";
    return quote + replacement + quote;
}

export function normalizePath(path: string): string {
    const normalized = path.split('\\').join('/');
    const parts: string[] = [];
    for(const part of normalized.split('/')) {
        if(part === '' || part === '.') { continue; }
        if(part === '..') { if(parts.length > 0 && parts[parts.length - 1] !== '..') { parts.pop(); } else if(!normalized.startsWith('/')) { parts.push(part); } }
        else { parts.push(part); }
    }
    return (normalized.startsWith('/') ? '/' : '') + parts.join('/');
}
export function resolvePath(base: string, path: string): string { return normalizePath(path.startsWith('/') || /^[A-Za-z]:/.test(path) ? path : base + '/' + path); }
export class CompilerPath {
    readonly pattern: string; readonly targets: readonly string[];
    constructor(pattern: string, targets: readonly string[]) { this.pattern = pattern; this.targets = targets; }
}
export function derivedAliases(paths: readonly CompilerPath[], base: string, root: string): Alias[] {
    const aliases: Alias[] = [];
    for(const path of paths) {
        const prefix = path.pattern.slice(0, -2);
        if(!path.pattern.endsWith('/*') || prefix === '' || prefix.includes('*') || path.targets.length !== 1) { continue; }
        const target = path.targets[0] ?? '';
        const directory = target.slice(0, -2);
        if(!target.endsWith('/*') || directory.includes('*')) { continue; }
        let resolved = resolvePath(base, directory);
        if(resolved.endsWith('/')) { resolved = resolved.slice(0, -1); }
        const local = relative(root, resolved);
        if(local !== undefined) { aliases.push(new Alias(local === '' ? '.' : local, prefix)); }
    }
    return aliases;
}
