// sys.ts:1597 casts what the require gives to source-map-support's type before it calls install. The
// require throws before the cast is had, so the cast checks nothing (host-any.md's Modelled row).
// Node's globals are typed only for a program that imports a node: module.
import 'node:process';
interface SourceMapSupport {
	install(): void;
}
try {
	(require("source-map-support") as SourceMapSupport).install();
}
catch {
	// Could not enable source maps.
}
