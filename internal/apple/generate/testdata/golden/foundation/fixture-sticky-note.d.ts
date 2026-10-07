// Generated from Objective-C declaration facts. Do not edit.

declare module 'apple/foundation/fixture-sticky-note' {
	import type { FixtureNote } from 'apple/foundation/fixture-note';
	import type { FixtureRecord } from 'apple/foundation/fixture-record';

	/**
	 * NSFixtureStickyNote
	 * @objc class NSFixtureStickyNote
	 */
	export class FixtureStickyNote extends FixtureNote {
		private constructor();

		/**
		 * -[NSFixtureStickyNote stickTo:]
		 * @objc method stickTo: 0.to:object -> void
		 */
		stick(options: { readonly to: FixtureRecord }): void;
	}
}
