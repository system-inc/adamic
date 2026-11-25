interface Contents { name: string }
interface PackageJsonInfo { packageDirectory: string; contents: Contents }
interface MissingPackageJsonInfo { packageDirectory: string; directoryExists: boolean }
type Entry = PackageJsonInfo | MissingPackageJsonInfo;
function isInfo(entry: Entry): entry is PackageJsonInfo {
	return entry.packageDirectory === "never";
}
function describe(entry: Entry): string {
	if (isInfo(entry)) {
		return "info";
	}
	return "missing at " + entry.packageDirectory;
}
console.log(describe({ packageDirectory: "/a", contents: { name: "a" } }));
