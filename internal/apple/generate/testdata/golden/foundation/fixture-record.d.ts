// Generated from Objective-C declaration facts. Do not edit.
// Skipped -[NSFixtureRecord consumeRecord:]: consumed parameters cannot use the bridge's borrowed arguments.
// Skipped -[NSFixtureRecord narrowFloat]: unsupported or unbound native type "float".
// Skipped -[NSFixtureRecord narrowInteger]: unsupported or unbound native type "int".
// Skipped -[NSFixtureRecord platformOnly]: unavailable on macos.
// Skipped -[NSFixtureRecord release]: manual reference-count messages bypass Adamic ownership.
// Skipped -[NSFixtureRecord takeArray:]: C arrays are not carried by the bridge.
// Skipped -[NSFixtureRecord takeCallback:]: function pointers are not carried by the bridge.
// Skipped -[NSFixtureRecord variadic:]: variadic declarations are not carried by the bridge.
// Skipped discardSelf: a consumed receiver cannot be passed borrowed.

declare module 'apple/foundation/fixture-record' {
	import type { FixtureReadable } from 'apple/foundation/fixture-readable';
	import type { FixtureRoot } from 'apple/foundation/fixture-root';
	import type { FixtureTableSource } from 'apple/foundation/fixture-table-source';

	/**
	 * NSFixtureRecord
	 * @objc class NSFixtureRecord
	 */
	export class FixtureRecord extends FixtureRoot {

		/**
		 * -[NSFixtureRecord initWithCount:]
		 * @objc init initWithCount: 0.count:integer
		 */
		constructor(options: { readonly count: number });

		/**
		 * -[NSFixtureRecord copyRecord]
		 * @objc method copyRecord -> new object
		 */
		copy(): FixtureRecord;

		/**
		 * -[NSFixtureRecord isReady]
		 * @objc get isReady -> boolean
		 */
		readonly isReady: boolean;

		/**
		 * -[NSFixtureRecord label], -[NSFixtureRecord setLabel:]
		 * @objc get label -> string
		 * @objc set setLabel: string
		 */
		label: string;

		/**
		 * -[NSFixtureRecord optionalText]
		 * @objc get optionalText -> string?
		 */
		readonly optionalText: string | undefined;

		/**
		 * -[NSFixtureRecord renamedAction]
		 * @objc method renamedAction -> void
		 */
		performAction(): void;

		/**
		 * -[NSFixtureRecord phoneOnly]
		 * @objc method phoneOnly -> void
		 */
		phoneOnly(): void;

		/**
		 * -[NSFixtureRecord rawAction]
		 * @objc method rawAction -> void
		 */
		rawAction(): void;

		/**
		 * -[NSFixtureRecord readText]
		 * @objc method readText -> string
		 */
		readText(): string;

		/**
		 * -[NSFixtureRecord source], -[NSFixtureRecord setSource:]
		 * @objc get source -> object?
		 * @objc set setSource: object?
		 */
		source: FixtureTableSource | undefined;

		/**
		 * -[NSFixtureRecord takeObject:]
		 * @objc method takeObject: 0:object -> void
		 */
		takeObject(record: FixtureRecord): void;

		/**
		 * -[NSFixtureRecord takeProtocol:]
		 * @objc method takeProtocol: 0:object -> void
		 */
		takeProtocol(reader: FixtureReadable): void;

		/**
		 * -[NSFixtureRecord takeText:]
		 * @objc method takeText: 0:string? -> void
		 */
		takeText(text: string | undefined): void;

		/**
		 * -[NSFixtureRecord displayText], -[NSFixtureRecord replaceText:]
		 * @objc get displayText -> string
		 * @objc set replaceText: string
		 */
		text: string;

		/**
		 * -[NSFixtureRecord withCompletion:]
		 * @objc method withCompletion: 0:block(string) -> void
		 */
		withCompletion(completion: (argument1: string) => void): void;
	}
}
