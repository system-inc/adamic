// An untagged union's cast that keeps undefined (#axqgdck under #b5w3ycg's keepsNullishCast): undefined passes
// through as itself, and a present value takes the untagged arm's checked view, so each prints what Node prints.
interface PackageJsonInfo { readonly packageDirectory: string; readonly version: string }
interface MissingPackageJsonInfo { readonly packageDirectory: string; readonly directoryExists: boolean }
function info(index: number): PackageJsonInfo | MissingPackageJsonInfo | undefined {
	if (index === 0) {
		return { packageDirectory: '/found', version: '1.0.0' };
	}
	if (index === 1) {
		return { packageDirectory: '/missing', directoryExists: false };
	}
	return undefined;
}
const found = info(0) as PackageJsonInfo | undefined;
console.log(found === undefined ? 'absent' : found.packageDirectory + ' ' + found.version);
const none = info(2) as PackageJsonInfo | undefined;
console.log(none === undefined ? 'absent' : none.version);
console.log(typeof none);
