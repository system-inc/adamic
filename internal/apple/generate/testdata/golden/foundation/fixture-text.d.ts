// Generated from Objective-C declaration facts. Do not edit.
// Skipped -[NSFixtureText oldWay]: deprecated on macos.

declare module 'apple/foundation/fixture-text' {
	import type { FixtureRoot } from 'apple/foundation/fixture-root';

	/**
	 * NSFixtureText
	 * @objc class NSFixtureText
	 */
	export class FixtureText extends FixtureRoot {

		/**
		 * -[NSFixtureText initWithText:]
		 * @objc init initWithText: 0.text:string
		 */
		constructor(options: { readonly text: string });

		/**
		 * -[NSFixtureText futureWay]
		 * @objc method futureWay -> void
		 */
		futureWay(): void;

		/**
		 * -[NSFixtureText phoneWay]
		 * @objc method phoneWay -> void
		 */
		phoneWay(): void;

		/**
		 * -[NSFixtureText text]
		 * @objc get text -> string
		 */
		readonly text: string;
	}
}
