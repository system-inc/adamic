// The three functions of Go's path package the gitignore Go calls: Clean, Base and Dir, for slash
// separated paths. They are written here from the package's documentation, since 0.1's library has no
// path module.

// isClean is whether clean would give the path back unchanged: not empty; no empty element (`//`) and
// no trailing slash but the root's; no `.` element but a path that is `.` alone; and no `..` element
// but in a leading run of them in a path that isn't rooted. Each test is one of the string library's
// own scans, which stage 0's runtime does in C, so a path already clean is checked without being read a
// character at a time into new strings, and comes back as itself, uncopied, as Go's path.Clean gives it.
function isClean(path: string): boolean {
	if (path === '' || path.includes('//') || (path.endsWith('/') && path !== '/')) {
		return false;
	}
	if (path === '.') {
		return true;
	}
	if (path.startsWith('./') || path.includes('/./') || path.endsWith('/.')) {
		return false;
	}
	if (path.startsWith('/')) {
		return !path.includes('/../') && !path.endsWith('/..');
	}
	// A relative path keeps its leading `..` elements; past them, none may follow.
	let rest = path;
	while (rest.startsWith('../')) {
		rest = rest.slice(3);
	}
	if (rest === '..') {
		return true;
	}
	return rest !== '.' && !rest.startsWith('./') && !rest.includes('/../') && !rest.endsWith('/..') && !rest.startsWith('../');
}

// clean is path.Clean: the shortest path naming the same file, by purely lexical processing. Runs of
// slashes are one slash, each `.` element goes, each `..` goes with the element before it, a `..` at
// the start of a rooted path goes, and the empty result is ".". A path already clean comes back as
// itself, as Go's does; any other is taken apart at its slashes and put back together.
//
// Go writes Clean as one pass over the bytes into a buffer that copies nothing until the output first
// differs from the input. Written that way here, unit by unit with charCodeAt, it was the slower of the
// two in stage 0 today: every charCodeAt is a call that, on a short string holding anything but ASCII,
// walks its UTF-8 from the start, and the format walk took 4.7 s where this takes 2.7 s or less (GAPS.md,
// "Path cleaning").
export function clean(path: string): string {
	if (isClean(path)) {
		return path;
	}
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

// dir is path.Dir: everything before the last element, cleaned. It is "." when there is no slash. Go
// cleans the part up to and with the last slash; cleaning it without that slash gives the same for
// anything but the root's slash alone, and an already-clean directory then comes back uncopied.
export function dir(path: string): string {
	const lastSlash = path.lastIndexOf('/');
	if (lastSlash < 0) {
		return '.';
	}
	if (lastSlash === 0) {
		return '/';
	}
	return clean(path.slice(0, lastSlash));
}
