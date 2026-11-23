// The lie: a MissingPackageJsonInfo cast as a PackageJsonInfo, an untagged union's cast, a checked view with no test
// at the cast (#axqgdck). It stops at the first read of a field the value lacks, with the view's own message. Node
// runs on and prints undefined.
interface PackageJsonInfo { readonly packageDirectory: string; readonly version: string }
interface MissingPackageJsonInfo { readonly packageDirectory: string; readonly directoryExists: boolean }
function info(entry: PackageJsonInfo | MissingPackageJsonInfo): PackageJsonInfo { return entry as PackageJsonInfo; }
console.log('casting');
console.log(typeof info({ packageDirectory: '/missing', directoryExists: false }).version);
