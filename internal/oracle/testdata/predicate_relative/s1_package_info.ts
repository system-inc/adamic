interface Contents { name: string; version: string }
interface PackageJsonInfo { packageDirectory: string; contents: Contents }
interface MissingPackageJsonInfo { packageDirectory: string; directoryExists: boolean }
type Entry = PackageJsonInfo | MissingPackageJsonInfo;
function isPackageJsonInfo(entry: Entry | undefined): entry is PackageJsonInfo {
	return entry !== undefined && entry.packageDirectory.startsWith("/info");
}
function describe(existing: Entry): string {
	if (isPackageJsonInfo(existing)) {
		return "info " + existing.packageDirectory + " " + existing.contents.name + "@" + existing.contents.version;
	}
	else {
		return "missing " + existing.packageDirectory + " " + String(existing.directoryExists);
	}
}
const entries: Entry[] = [{ packageDirectory: "/info/a", contents: { name: "a", version: "1" } }, { packageDirectory: "/b", directoryExists: false }, { packageDirectory: "/c", directoryExists: true }];
for (const entry of entries) {
	console.log(describe(entry));
}
const none: Entry | undefined = undefined;
console.log(isPackageJsonInfo(none) ? "info" : "not info");
