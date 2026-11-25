interface Contents { name: string }
interface PackageJsonInfo { packageDirectory: string; contents: Contents }
interface MissingPackageJsonInfo { packageDirectory: string; directoryExists: boolean }
type Entry = PackageJsonInfo | MissingPackageJsonInfo;
function isMissingPackageJsonInfo(entry: Entry | undefined): entry is MissingPackageJsonInfo {
	return entry !== undefined && !entry.packageDirectory.startsWith("/info");
}
function imports(cached: Entry | undefined): string {
	if (isMissingPackageJsonInfo(cached) || cached === undefined) {
		return "none";
	}
	return "imports of " + cached.contents.name;
}
console.log(imports({ packageDirectory: "/info/a", contents: { name: "a" } }));
console.log(imports({ packageDirectory: "/b", directoryExists: true }));
console.log(imports(undefined));
const missing: Entry = { packageDirectory: "/c", directoryExists: false };
if (isMissingPackageJsonInfo(missing)) {
	console.log("exists " + String(missing.directoryExists));
}
