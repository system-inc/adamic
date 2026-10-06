// Bun's counterpart to oracle/node.mjs: load Adamic .a sources as TypeScript.
// Package resolution uses the independent sequential shim supplied by bench/run.go.
import { plugin } from 'bun';
import { readFileSync } from 'node:fs';

plugin({
	name: 'adamic-source',
	setup(build) {
		build.onLoad({ filter: /\.a$/ }, ({ path }) => ({
			contents: readFileSync(path, 'utf8'),
			loader: 'ts',
		}));
	},
});
