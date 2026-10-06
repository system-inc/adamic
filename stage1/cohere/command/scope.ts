// Named paths use WalkDir, not the formatter walk: keep duplicates and symlink entries, and
// never follow a directory link. Dot means everything only at the project root.
import { fileStatus, readDirectory, realPath } from 'adamic';
import { lstat } from '../formatfiles/disk.ts';
import { compareCodePoints } from '../formatfiles/golang.ts';
import type { Enumeration } from '../formatfiles/enumerate.ts';
import { clean } from '../gitignore/path.ts';
import { absoluteFrom } from './location.ts';

export class Scope {
    fileNames: string[] = [];
    description = '';
    everything = false;
    dependentCount = 0;
    requestDescription = '';
    error = '';
    index = new Map<string, boolean>();
    includes(path: string): boolean {
        return this.index.has(path);
    }
    copy(): Scope {
        const result = new Scope();
        result.fileNames = this.fileNames;
        result.description = this.description;
        result.everything = this.everything;
        result.dependentCount = this.dependentCount;
        result.requestDescription = this.requestDescription;
        result.error = this.error;
        result.index = this.index;
        return result;
    }
    narrowTo(population: string[]): Scope {
        if(this.everything || this.fileNames.length === 0) {
            return this.copy();
        }
        const set = new Map<string, boolean>();
        for(const name of population) {
            set.set(name, true);
        }
        let count = 0;
        for(const name of this.fileNames) {
            if(set.has(name)) {
                count++;
            }
        }
        const result = this.copy();
        result.description += `, ${count} of them in the program`;
        return result;
    }
    narrowToEnumeration(enumeration: Enumeration): Scope {
        if(this.everything) {
            const result = new Scope();
            result.fileNames = enumeration.files;
            for(const name of result.fileNames) {
                result.index.set(name, true);
            }
            result.description = describeEnumeration(enumeration, result.fileNames.length);
            return result;
        }
        if(this.fileNames.length === 0) {
            return this.copy();
        }
        const handled = new Map<string, boolean>();
        for(const name of enumeration.files) {
            handled.set(name, true);
        }
        let count = 0;
        for(const name of this.fileNames) {
            if(handled.has(name)) {
                count++;
            }
        }
        const result = this.copy();
        result.description += `, ${count} the formatter handles · ${describeEnumeration(enumeration, count)}`;
        return result;
    }
}
export function wholeTreeScope(): Scope {
    const result = new Scope();
    result.everything = true;
    result.description = 'every file (--format-all)';
    return result;
}
function walk(path: string, names: string[]): string {
    const status = lstat(path);
    if(!status.directory) {
        names.push(clean(path));
        return '';
    }
    const listing = readDirectory(path);
    if(listing.kind === 'Error') {
        const reason = listing.message.endsWith(': permission denied')
            ? 'permission denied'
            : 'unrepresented filesystem error';
        return `open ${path}: ${reason}`;
    }
    const ordered = listing.names.slice();
    ordered.sort(compareCodePoints);
    for(const name of ordered) {
        const error = walk(absoluteFrom(path, name), names);
        if(error !== '') {
            return error;
        }
    }
    return '';
}
export function namedPathsScope(workingDirectory: string, root: string, names: string[]): Scope {
    const scope = new Scope();
    let cwd = workingDirectory;
    if(cwd === '') {
        const current = realPath('.');
        if(current.kind === 'Error') {
            scope.error = 'resolving the working directory: unrepresented filesystem error';
            return scope;
        }
        cwd = current.path;
    }
    const projectRoot = root === '' ? cwd : root;
    for(const name of names) {
        const path = absoluteFrom(cwd, name);
        const status = fileStatus(path);
        if(status.kind === 'Error') {
            const reason = status.message.endsWith(': no such file')
                ? 'no such file or directory'
                : status.message.endsWith(': permission denied')
                  ? 'permission denied'
                  : status.message.endsWith(': not a directory')
                    ? 'not a directory'
                    : 'unrepresented filesystem error';
            const failed = new Scope();
            failed.error = `resolving ${name}: stat ${path}: ${reason}`;
            return failed;
        }
        if(status.type !== 'directory') {
            scope.fileNames.push(path);
            continue;
        }
        if(path === clean(projectRoot)) {
            return wholeTreeScope();
        }
        const error = walk(path, scope.fileNames);
        if(error !== '') {
            const failed = new Scope();
            failed.error = `walking ${name}: ${error}`;
            return failed;
        }
    }
    scope.fileNames.sort(compareCodePoints);
    for(const name of scope.fileNames) {
        scope.index.set(name, true);
    }
    scope.description =
        scope.fileNames.length === 0
            ? `0 files under the named paths in ${cwd}`
            : `${scope.fileNames.length} named path${scope.fileNames.length === 1 ? '' : 's'}`;
    scope.requestDescription =
        names.length > 3 ? `${names.slice(0, 3).join(', ')} and ${names.length - 3} more` : names.join(', ');
    return scope;
}
export function describeEnumeration(enumeration: Enumeration, formattable: number): string {
    let description = `walked ${enumeration.root}, ${enumeration.walked} files, ${formattable} the formatter handles`;
    const layers: string[] = [];
    for(const [layer, count] of enumeration.ignoredByLayer) {
        if(count > 0) {
            layers.push(`${count} by ${layer}`);
        }
    }
    layers.sort(compareCodePoints);
    if(layers.length > 0) {
        description += ', ignored ' + layers.join(', ');
    }
    if(enumeration.nestedRepositories.length > 0) {
        const repositories = enumeration.nestedRepositories.slice();
        repositories.sort(compareCodePoints);
        description += ', skipped nested repositories ' + repositories.join(', ');
    }
    if(enumeration.symbolicLinks > 0) {
        description += `, skipped ${enumeration.symbolicLinks} symbolic links, which are never followed`;
    }
    const extensions: string[] = [];
    for(const extension of enumeration.declinedExtensions.keys()) {
        extensions.push(extension);
    }
    extensions.sort(compareCodePoints);
    const declined: string[] = [];
    for(const extension of extensions) {
        declined.push(`${enumeration.declinedExtensions.get(extension) ?? 0} ${extension}`);
    }
    if(declined.length > 0) {
        description += ', declined ' + declined.join(', ');
    }
    return description;
}
