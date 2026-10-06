import { fileStatus } from 'adamic';
import { clean, dir } from '../gitignore/path.ts';

export interface LocationRequest {
    readonly workingDirectory: string;
    readonly directory: string;
    readonly engine: string;
    readonly configFileName: string;
    readonly configFileNameGiven: boolean;
    readonly lintConfigFileName: string;
    readonly lintConfigFileNameGiven: boolean;
}
export class Location {
    root = '';
    engine = '';
    configFileName = '';
    lintConfigFileName = '';
    lintConfigFileNameGiven = false;
    argumentBase = '';
    workingDirectory = '';
    error = '';
    rootNote(): string {
        return clean(this.root) === clean(this.workingDirectory)
            ? ''
            : `checked the project at ${this.root} (the run started in ${this.workingDirectory})`;
    }
}
export function absoluteFrom(directory: string, path: string): string {
    return clean(path.startsWith('/') ? path : `${directory}/${path}`);
}
function regularFile(path: string): boolean {
    const status = fileStatus(path);
    return status.kind === 'Ok' && status.type !== 'directory';
}
function engineAt(directory: string): string {
    if(regularFile(absoluteFrom(directory, 'tsconfig.json'))) {
        return 'TypeScript';
    }
    if(regularFile(absoluteFrom(directory, 'Package.swift'))) {
        return 'Swift';
    }
    return '';
}
export function locateProject(request: LocationRequest): Location {
    const location = new Location();
    const base =
        request.directory === '' ? request.workingDirectory : absoluteFrom(request.workingDirectory, request.directory);
    location.argumentBase = base;
    location.workingDirectory = request.workingDirectory;
    if(request.directory !== '') {
        location.root = base;
        location.configFileName = absoluteFrom(base, request.configFileName);
        const engine = engineAt(base);
        location.engine = request.engine !== '' ? request.engine : engine === '' ? 'TypeScript' : engine;
    }
    else if(request.configFileNameGiven) {
        location.configFileName = absoluteFrom(base, request.configFileName);
        location.root = dir(location.configFileName);
        location.engine = 'TypeScript';
    }
    else {
        let directory = clean(base);
        while(true) {
            const engine = engineAt(directory);
            if(engine !== '') {
                location.root = directory;
                location.engine = engine;
                break;
            }
            const parent = dir(directory);
            if(parent === directory) {
                const failed = new Location();
                failed.error = `no tsconfig.json or Package.swift in ${base} or any directory above it, so there is no project here to check: run from inside one, or name it with -tsconfig or -directory`;
                return failed;
            }
            directory = parent;
        }
        if(location.engine === 'TypeScript') {
            location.configFileName = clean(`${location.root}/${request.configFileName}`);
        }
    }
    location.lintConfigFileNameGiven = request.lintConfigFileNameGiven;
    location.lintConfigFileName = absoluteFrom(
        request.lintConfigFileNameGiven ? base : location.root,
        request.lintConfigFileName,
    );
    return location;
}
