// Which run discovers a repository rather than checking the one project it names.
import { fileStatus } from 'adamic';
import { clean, dir } from '../gitignore/path.ts';
import { absoluteFrom, type Location } from './location.ts';
export function discoveryApplies(given: Map<string, boolean>, names: string[]): boolean {
    if(names.length > 0) {
        return false;
    }
    for(const name of [
        'directory',
        'tsconfig',
        'rules',
        'rules-enabled',
        'print-config',
        'version',
        'cache-dump',
        'explain',
        'stdin-filepath',
        'profile',
    ]) {
        if(given.get(name) === true) {
            return false;
        }
    }
    return true;
}
export function discoveryRoot(location: Location, workingDirectory: string): string {
    let repository = '';
    for(let directory = clean(workingDirectory); ; directory = dir(directory)) {
        const status = fileStatus(absoluteFrom(directory, '.git'));
        if(status.kind === 'Ok' && status.type === 'directory') {
            repository = directory;
            break;
        }
        if(dir(directory) === directory) {
            break;
        }
    }
    if(location.error === '') {
        if(repository !== '' && repository.startsWith(location.root + '/')) {
            return repository;
        }
        return location.root;
    }
    return repository === '' ? workingDirectory : repository;
}
