// The three functions of Go's path package the gitignore Go calls: Clean, Base and Dir, for slash
// separated paths. They are written here from the package's documentation, since 0.1's library has no
// path module.

// clean is path.Clean: the shortest path naming the same file, by purely lexical processing. Runs of
// slashes are one slash, each `.` element goes, each `..` goes with the element before it, a `..` at
// the start of a rooted path goes, and the empty result is ".".
export function clean(path: string): string {
	if (path === '') {
		return '.';
	}
	const rooted = path.startsWith('/');
	const kept: string[] = [];
	for (const element of path.split('/')) {
		if (element === '' || element === '.') {
			continue;
		}
		if (element === '..') {
			const previous = kept.at(-1);
			if (previous !== undefined && previous !== '..') {
				kept.pop();
				continue;
			}
			if (rooted) {
				continue;
			}
		}
		kept.push(element);
	}
	const joined = kept.join('/');
	if (rooted) {
		return '/' + joined;
	}
	return joined === '' ? '.' : joined;
}

// base is path.Base: the last element, trailing slashes dropped first. It is "." for the empty path and
// "/" for a path of slashes.
export function base(path: string): string {
	if (path === '') {
		return '.';
	}
	let end = path.length;
	while (end > 0 && path[end - 1] === '/') {
		end--;
	}
	if (end === 0) {
		return '/';
	}
	const trimmed = path.slice(0, end);
	return trimmed.slice(trimmed.lastIndexOf('/') + 1);
}

// dir is path.Dir: everything before the last element, cleaned. It is "." when there is no slash.
export function dir(path: string): string {
	return clean(path.slice(0, path.lastIndexOf('/') + 1));
}
