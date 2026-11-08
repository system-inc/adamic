// sys.ts:1619 returns what a computed require gives as the module of a ModuleImportResult. The require
// throws before the object is built, so no member typed any reaches the return (host-any.md's
// Modelled row). Node's globals are typed only for a program that imports a node: module.
import 'node:process';
type ModuleImportResult<T = {}> =
	| { module: T; modulePath?: string; error: undefined; }
	| { module: undefined; modulePath?: undefined; error: { stack?: string; message?: string; }; };
function requirePlugin(modulePath: string): ModuleImportResult {
	try {
		return { module: require(modulePath), modulePath, error: undefined };
	}
	catch (error) {
		return { module: undefined, modulePath: undefined, error: { message: error instanceof Error ? error.message : "unknown" } };
	}
}
requirePlugin("./plugin.js");
