// Generated from Objective-C declaration facts. Do not edit.

declare module 'apple/foundation/fixture-timer' {
	import type { FixtureRecord } from 'apple/foundation/fixture-record';
	import type { FixtureRoot } from 'apple/foundation/fixture-root';

	/**
	 * NSFixtureTimer
	 * @objc class NSFixtureTimer
	 */
	export class FixtureTimer extends FixtureRoot {

		/**
		 * -[NSFixtureTimer initWithTarget:]
		 * @objc init initWithTarget: 0.target:object
		 */
		constructor(options: { readonly target: FixtureRecord });

		/**
		 * -[NSFixtureTimer interval]
		 * @objc method interval -> integer
		 */
		interval(): number;
	}
}
