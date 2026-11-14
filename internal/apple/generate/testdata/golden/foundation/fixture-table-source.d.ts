// Generated from Objective-C declaration facts. Do not edit.
// Skipped -[NSFixtureTableSource copyFixtureTable:]: a result its caller owns can't be given back by a program's class yet.
// Skipped -[NSFixtureTableSource fixtureTable:finish:]: a program's class can't be handed a block() yet.

declare module 'apple/foundation/fixture-table-source' {
	import type { FixtureRecord } from 'apple/foundation/fixture-record';

	/**
	 * NSFixtureTableSource
	 * @objc protocol NSFixtureTableSource
	 */
	export interface FixtureTableSource {

		/**
		 * -[NSFixtureTableSource fixtureTableDidReload:]
		 * @objc implement fixtureTableDidReload: 0:object -> void
		 */
		fixtureTableDidReload?(table: FixtureRecord): void;

		/**
		 * -[NSFixtureTableSource fixtureTable:shouldSelectRow:]
		 * @objc implement fixtureTable:shouldSelectRow: 0:object 1:integer -> boolean
		 */
		fixtureTableShouldSelectRow?(table: FixtureRecord, row: number): boolean;

		/**
		 * -[NSFixtureTableSource fixtureTable:textForRow:]
		 * @objc implement fixtureTable:textForRow: 0:object 1:integer -> string?
		 */
		fixtureTableTextForRow?(table: FixtureRecord, row: number): string | undefined;

		/**
		 * -[NSFixtureTableSource numberOfRowsInFixtureTable:]
		 * @objc method numberOfRowsInFixtureTable: 0:object -> integer
		 * @objc implement numberOfRowsInFixtureTable: 0:object -> integer
		 */
		numberOfRows(table: FixtureRecord): number;
	}
}
