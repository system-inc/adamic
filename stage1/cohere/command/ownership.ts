// ownership.go's repository planning. The checker still sees yielded files in its own graph;
// this decides which project reports them and which directory each formatter leaves to another.
import { fileStatus, readTextFile, panic } from 'adamic';
import { readProjectConfig, type ProjectResult } from '../config/tsconfig.ts';
import { discoverConfiguredProjects } from '../config/discovery.ts';
import { compareCodePoints } from '../formatfiles/golang.ts';
import { clean, dir, base } from '../gitignore/path.ts';
import { absoluteFrom } from './location.ts';
export interface Project {
    readonly directory: string;
    readonly engine: string;
    readonly configFile: string;
}
export class Yield {
    readonly files: string[] = [];
    readonly directories: string[] = [];
}
export class OwnedProjects {
    projects: Project[] = [];
    readonly solutions: string[] = [];
    readonly yielding: string[] = [];
    readonly yields = new Map<string, Yield>();
    readonly sharedDirectory = new Map<string, boolean>();
    error = '';
}
export function projectLabel(project: Project): string {
    return project.configFile === '' ? project.directory : clean(`${project.directory}/${project.configFile}`);
}
function configPath(root: string, project: Project): string {
    return absoluteFrom(
        root,
        `${project.directory}/${project.configFile === '' ? 'tsconfig.json' : project.configFile}`,
    );
}
function below(directory: string, ancestor: string): boolean {
    return directory !== ancestor && directory !== '.' && (ancestor === '.' || directory.startsWith(ancestor + '/'));
}
function shared(left: string, right: string): number {
    if(left === '.' || right === '.') {
        return 0;
    }
    const a = left.split('/');
    const b = right.split('/');
    let count = 0;
    while(count < a.length && count < b.length && a[count] === b[count]) {
        count++;
    }
    return count;
}
function relative(root: string, path: string): string {
    return path === root
        ? '.'
        : path.startsWith(root === '/' ? '/' : root + '/')
          ? path.slice(root === '/' ? 1 : root.length + 1)
          : '..';
}
function order(left: Project, right: Project): number {
    if(left.directory === right.directory) {
        if(left.engine !== right.engine) {
            return left.engine === 'TypeScript' ? -1 : 1;
        }
        return compareCodePoints(left.configFile, right.configFile);
    }
    if(left.directory === '.' || right.directory === '.') {
        return left.directory === '.' ? -1 : 1;
    }
    const a = left.directory.split('/');
    const b = right.directory.split('/');
    for(let at = 0; at < a.length && at < b.length; at++) {
        const compared = compareCodePoints(a[at] ?? '', b[at] ?? '');
        if(compared !== 0) {
            return compared;
        }
    }
    return a.length - b.length;
}
export function resolveOwnership(
    root: string,
    projects: Project[],
    refused: string[],
    repositories: string[],
): OwnedProjects {
    const result = new OwnedProjects();
    result.projects = projects.slice();
    const typed = projects.filter((project) => project.engine === 'TypeScript');
    if(typed.length === 0) {
        return result;
    }
    if(typed.length === 1) {
        const read = readTextFile(configPath(root, typed[0] ?? panic('missing project')));
        if(read.kind === 'Ok' && !read.text.includes('"references"')) {
            return result;
        }
    }
    const reads = new Map<string, ProjectResult>();
    const known = new Map<string, boolean>();
    for(const project of typed) {
        const path = configPath(root, project);
        known.set(path, true);
        reads.set(path, readProjectConfig(path));
    }
    let queue = projects;
    while(queue.length > 0) {
        const added: Project[] = [];
        for(const project of queue) {
            const read = reads.get(configPath(root, project));
            if(project.engine !== 'TypeScript' || read === undefined || read.kind === 'Error') {
                continue;
            }
            for(const reference of read.project.references) {
                if(known.has(reference)) {
                    continue;
                }
                known.set(reference, true);
                const path = relative(root, reference);
                const directory = dir(path);
                if(
                    path === '..' ||
                    directory
                        .split('/')
                        .some((segment) =>
                            ['node_modules', '.build', '.cache', '.git', 'testdata'].includes(segment),
                        ) ||
                    repositories.some((repository) => directory === repository || below(directory, repository)) ||
                    refused.includes(path)
                ) {
                    continue;
                }
                const status = fileStatus(reference);
                if(status.kind === 'Error' || status.type === 'directory') {
                    continue;
                }
                added.push({
                    directory,
                    engine: 'TypeScript',
                    configFile: base(path) === 'tsconfig.json' ? '' : base(path),
                });
            }
        }
        for(const project of added) {
            const path = configPath(root, project);
            reads.set(path, readProjectConfig(path));
            result.projects.push(project);
        }
        queue = added;
    }
    result.projects.sort(order);
    const running: Project[] = [];
    for(const project of result.projects) {
        const read = reads.get(configPath(root, project));
        if(
            project.engine === 'TypeScript' &&
            read !== undefined &&
            read.kind === 'Ok' &&
            read.project.files.length === 0 &&
            read.project.references.length > 0
        ) {
            result.solutions.push(projectLabel(project));
        }
        else {
            running.push(project);
        }
    }
    result.projects = running;
    const includers = new Map<string, number[]>();
    for(let at = 0; at < running.length; at++) {
        const project = running[at] ?? panic('missing project');
        const read = reads.get(configPath(root, project));
        if(project.engine !== 'TypeScript' || read === undefined || read.kind === 'Error') {
            continue;
        }
        for(const file of read.project.files) {
            const indexes: number[] = includers.get(file) ?? [];
            indexes.push(at);
            includers.set(file, indexes);
        }
    }
    const yielded = running.map(() => new Yield());
    for(const [file, indexes] of includers) {
        if(indexes.length < 2) {
            continue;
        }
        const directory = dir(relative(root, file));
        let owner = indexes[0] ?? 0;
        let ownerHolds = false;
        let ownerShared = -1;
        for(const index of indexes) {
            const project = running[index] ?? panic('missing includer');
            const holds =
                project.directory === '.' || directory === project.directory || below(directory, project.directory);
            const common = shared(directory, project.directory);
            if((holds && !ownerHolds) || (holds === ownerHolds && common > ownerShared)) {
                owner = index;
                ownerHolds = holds;
                ownerShared = common;
            }
        }
        for(const index of indexes) {
            if(index !== owner) {
                (yielded[index] ?? panic('missing yield')).files.push(file);
            }
        }
    }
    const kept: Project[] = [];
    const keeps: Yield[] = [];
    for(let at = 0; at < running.length; at++) {
        const project = running[at] ?? panic('missing project');
        const read = reads.get(configPath(root, project));
        const yieldTo = yielded[at] ?? panic('missing yield');
        if(
            read !== undefined &&
            read.kind === 'Ok' &&
            read.project.files.length > 0 &&
            yieldTo.files.length === read.project.files.length
        ) {
            result.yielding.push(projectLabel(project));
        }
        else {
            kept.push(project);
            keeps.push(yieldTo);
        }
    }
    result.projects = kept;
    const cacheDirectories = new Map<string, boolean>();
    for(let at = 0; at < kept.length; at++) {
        const project = kept[at] ?? panic('missing project');
        if(project.engine !== 'TypeScript') {
            continue;
        }
        const yieldTo = keeps[at] ?? panic('missing yield');
        const directory = absoluteFrom(root, project.directory);
        const label = projectLabel(project);
        if(cacheDirectories.has(directory)) {
            result.sharedDirectory.set(label, true);
            yieldTo.directories.push(directory);
        }
        else {
            cacheDirectories.set(directory, true);
            for(const other of kept) {
                if(other.engine === 'TypeScript' && below(other.directory, project.directory)) {
                    yieldTo.directories.push(absoluteFrom(root, other.directory));
                }
            }
        }
        if(yieldTo.files.length > 0 || yieldTo.directories.length > 0) {
            yieldTo.files.sort(compareCodePoints);
            yieldTo.directories.sort(compareCodePoints);
            result.yields.set(label, yieldTo);
        }
    }
    return result;
}
export function discoverOwnedProjects(root: string): OwnedProjects {
    const discovered = discoverConfiguredProjects(root);
    if(discovered.kind === 'Error') {
        const result = new OwnedProjects();
        result.error = discovered.message;
        return result;
    }
    return resolveOwnership(
        root,
        discovered.found.projects.map((project) => ({
            directory: project.directory,
            engine: project.engine,
            configFile: '',
        })),
        discovered.found.refused,
        discovered.found.nestedRepositories,
    );
}
export function childArguments(argumentsFromCaller: string[], absoluteLintConfig: string): string[] {
    if(absoluteLintConfig === '') {
        return argumentsFromCaller;
    }
    const result: string[] = [];
    for(let at = 0; at < argumentsFromCaller.length; at++) {
        const argument = argumentsFromCaller[at] ?? '';
        let withoutHyphens = argument;
        while(withoutHyphens.startsWith('-')) {
            withoutHyphens = withoutHyphens.slice(1);
        }
        const equals = withoutHyphens.indexOf('=');
        const name = equals < 0 ? withoutHyphens : withoutHyphens.slice(0, equals);
        if(name === 'lint-config' && argument.startsWith('-')) {
            if(equals < 0) {
                at++;
            }
            continue;
        }
        result.push(argument);
    }
    result.push('--lint-config');
    result.push(absoluteLintConfig);
    return result;
}
