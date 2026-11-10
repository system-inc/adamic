function run(read: (version: Version) => number, version: Version): number {
	return read(version);
}
function reader(version: Version): number {
	return version.size;
}
class Version {
	static later = 7;
	static zero = run(reader, new Version());
	get size(): number {
		return Version.later;
	}
}
console.log(String(Version.zero));
