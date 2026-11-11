// tsc's semver.ts Version and its static zero, constructed in its own static initializer (#1451npc),
// with what its constructor reaches copied from TypeScript 6.0.3 as the stage 3 adaptations leave it:
// debug.ts's fail, assert, assertIsDefined and checkDefined (adaptation 41 spells (Error as
// any).captureStackTrace as Error.captureStackTrace), core.ts's every, some and isArray (adaptation 30
// reads array[i]!, adaptation 40 types isArray's value unknown), and semver.ts's tryParseComponents
// (adaptation 45 binds major as match[1]!). emptyArray drops its `as never[]`. The walk admits
// Error.captureStackTrace, a library member that runs no program code, and assert's
// verboseDebugInfo(), since no call reaching assert passes one. Node prints 0.0.0-0,
// 1.2.3-beta.1+build.5, 2.0.0, then Debug Failure. False expression: Invalid argument: major.
import 'node:process';

type AnyFunction = (...args: never[]) => void;

const emptyArray: never[] = [];

function every<T, U extends T>(array: readonly T[], callback: (element: T, index: number) => element is U): array is readonly U[];
function every<T, U extends T>(array: readonly T[] | undefined, callback: (element: T, index: number) => element is U): array is readonly U[] | undefined;
function every<T>(array: readonly T[] | undefined, callback: (element: T, index: number) => boolean): boolean;
function every<T>(array: readonly T[] | undefined, callback: (element: T, index: number) => boolean): boolean {
	if (array !== undefined) {
		for (let i = 0; i < array.length; i++) {
			if (!callback(array[i]!, i)) {
				return false;
			}
		}
	}

	return true;
}

function some<T>(array: readonly T[] | undefined): array is readonly T[];
function some<T>(array: readonly T[] | undefined, predicate: (value: T) => boolean): boolean;
function some<T>(array: readonly T[] | undefined, predicate?: (value: T) => boolean): boolean {
	if (array !== undefined) {
		if (predicate !== undefined) {
			for (let i = 0; i < array.length; i++) {
				if (predicate(array[i]!)) {
					return true;
				}
			}
		}
		else {
			return array.length > 0;
		}
	}
	return false;
}

function isArray(value: unknown): value is readonly unknown[] {
	return Array.isArray(value);
}

namespace Debug {
	export function fail(message?: string, stackCrawlMark?: AnyFunction): never {
		debugger;
		const e = new Error(message ? `Debug Failure. ${message}` : "Debug Failure.");
		if (Error.captureStackTrace) {
			Error.captureStackTrace(e, stackCrawlMark || fail);
		}
		throw e;
	}

	export function assert(expression: unknown, message?: string, verboseDebugInfo?: string | (() => string), stackCrawlMark?: AnyFunction): asserts expression {
		if (!expression) {
			message = message ? `False expression: ${message}` : "False expression.";
			if (verboseDebugInfo) {
				message += "\r\nVerbose Debug Information: " + (typeof verboseDebugInfo === "string" ? verboseDebugInfo : verboseDebugInfo());
			}
			fail(message, stackCrawlMark || assert);
		}
	}

	export function assertIsDefined<T>(value: T, message?: string, stackCrawlMark?: AnyFunction): asserts value is NonNullable<T> {
		if (value === undefined || value === null) {
			fail(message, stackCrawlMark || assertIsDefined);
		}
	}

	export function checkDefined<T>(value: T | null | undefined, message?: string, stackCrawlMark?: AnyFunction): T {
		assertIsDefined(value, message, stackCrawlMark || checkDefined);
		return value;
	}
}

const versionRegExp = /^(0|[1-9]\d*)(?:\.(0|[1-9]\d*)(?:\.(0|[1-9]\d*)(?:-([a-z0-9-.]+))?(?:\+([a-z0-9-.]+))?)?)?$/i;
const prereleaseRegExp = /^(?:0|[1-9]\d*|[a-z-][a-z0-9-]*)(?:\.(?:0|[1-9]\d*|[a-z-][a-z0-9-]*))*$/i;
const prereleasePartRegExp = /^(?:0|[1-9]\d*|[a-z-][a-z0-9-]*)$/i;
const buildRegExp = /^[a-z0-9-]+(?:\.[a-z0-9-]+)*$/i;
const buildPartRegExp = /^[a-z0-9-]+$/i;

class Version {
	static readonly zero: Version = new Version(0, 0, 0, ["0"]);

	readonly major: number;
	readonly minor: number;
	readonly patch: number;
	readonly prerelease: readonly string[];
	readonly build: readonly string[];

	constructor(text: string);
	constructor(major: number, minor?: number, patch?: number, prerelease?: string | readonly string[], build?: string | readonly string[]);
	constructor(major: number | string, minor = 0, patch = 0, prerelease: string | readonly string[] = "", build: string | readonly string[] = "") {
		if (typeof major === "string") {
			const result = Debug.checkDefined(tryParseComponents(major), "Invalid version");
			({ major, minor, patch, prerelease, build } = result);
		}

		Debug.assert(major >= 0, "Invalid argument: major");
		Debug.assert(minor >= 0, "Invalid argument: minor");
		Debug.assert(patch >= 0, "Invalid argument: patch");

		const prereleaseArray = prerelease ? isArray(prerelease) ? prerelease : prerelease.split(".") : emptyArray;
		const buildArray = build ? isArray(build) ? build : build.split(".") : emptyArray;

		Debug.assert(every(prereleaseArray, s => prereleasePartRegExp.test(s)), "Invalid argument: prerelease");
		Debug.assert(every(buildArray, s => buildPartRegExp.test(s)), "Invalid argument: build");

		this.major = major;
		this.minor = minor;
		this.patch = patch;
		this.prerelease = prereleaseArray;
		this.build = buildArray;
	}

	toString(): string {
		let result = `${this.major}.${this.minor}.${this.patch}`;
		if (some(this.prerelease)) result += `-${this.prerelease.join(".")}`;
		if (some(this.build)) result += `+${this.build.join(".")}`;
		return result;
	}
}

function tryParseComponents(text: string) {
	const match = versionRegExp.exec(text);
	if (!match) return undefined;

	const major = match[1]!;
	const [, , minor = "0", patch = "0", prerelease = "", build = ""] = match;
	if (prerelease && !prereleaseRegExp.test(prerelease)) return undefined;
	if (build && !buildRegExp.test(build)) return undefined;
	return {
		major: parseInt(major, 10),
		minor: parseInt(minor, 10),
		patch: parseInt(patch, 10),
		prerelease,
		build,
	};
}

console.log(Version.zero.toString());
console.log(new Version("1.2.3-beta.1+build.5").toString());
console.log(new Version(2).toString());
try {
	const invalid = new Version(-1);
	console.log(invalid.toString());
}
catch (error) {
	if (error instanceof Error) {
		console.log(error.message);
	}
}
