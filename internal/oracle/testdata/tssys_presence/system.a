/*
 * tsc's ts.sys (src/compiler/sys.ts at 050880ce), for the presence tests read off it (#p2n17c4 T1): an
 * interface of required members and optional ones, two of each optional form present on the system and
 * the rest absent, and an exported let made by an arrow called where it's written. The arrow nests
 * getNodeSystem (sys.ts:1468), whose const literal holds method shorthands, shorthand properties of
 * functions declared after its return, and one shorthand that reads the literal itself (createDirectory,
 * sys.ts:1546). The arrow's own let is undefined until getNodeSystem fills it, a later call stores a
 * field onto it (patchWriteFileEnsuringDirectory, sys.ts:1376-1381), and it's handed back with !
 * (sys.ts:1955-1963). setSys writes the export, and nothing calls it (sys.ts:1966-1969).
 */
export interface System {
	newLine: string;
	useCaseSensitiveFileNames: boolean;
	write(s: string): void;
	writeOutputIsTTY?(): boolean;
	readFile(path: string): string | undefined;
	writeFile(path: string, data: string): void;
	directoryExists(path: string): boolean;
	createDirectory(path: string): void;
	getMemoryUsage?(): number;
	tryEnableSourceMapsForHost?(): void;
	setBlocking?(): void;
	clearScreen?: () => void;
	base64encode?: (input: string) => string;
}

function isNodeLikeSystem(): boolean {
	return true;
}

export function patchWriteFileEnsuringDirectory(sys: System): void {
	sys.writeFile = (path: string, data: string): void => {
		console.log(`patched writeFile ${path} ${data}`);
	};
}

export let sys: System = (() => {
	function getNodeSystem(): System {
		const useCaseSensitiveFileNames = true;
		const nodeSystem: System = {
			newLine: '\n',
			useCaseSensitiveFileNames,
			write(s: string): void {
				console.log(s);
			},
			writeOutputIsTTY() {
				return false;
			},
			readFile,
			writeFile(path: string, data: string): void {
				console.log(`writeFile ${path} ${data}`);
			},
			directoryExists,
			createDirectory(directoryName: string) {
				if (!nodeSystem.directoryExists(directoryName)) {
					console.log(`createDirectory ${directoryName}`);
				}
			},
			tryEnableSourceMapsForHost() {
				console.log('tryEnableSourceMapsForHost');
			},
			setBlocking() {
				console.log('setBlocking');
			},
			base64encode: (input: string): string => `encoded ${input}`,
		};
		return nodeSystem;

		function readFile(path: string): string | undefined {
			return path === 'tsconfig.json' ? '{}' : undefined;
		}

		function directoryExists(path: string): boolean {
			return path === 'src';
		}
	}

	let sys: System | undefined;
	if (isNodeLikeSystem()) {
		sys = getNodeSystem();
	}
	if (sys) {
		patchWriteFileEnsuringDirectory(sys);
	}
	return sys!;
})();

export function setSys(s: System): void {
	sys = s;
}
