// Generated from Objective-C declaration facts. Do not edit.
// Skipped -[NSFixturePanel bounds]: rectangle results are not carried by the bridge.
// Skipped -[NSFixturePanel takeUnion:]: unions are not carried by the bridge.

declare module 'apple/appkit/fixture-panel' {
	import type { FixtureMode } from 'apple/appkit/fixture-mode';
	import type { FixtureRecord } from 'apple/foundation/fixture-record';
	import type { FixtureRoot } from 'apple/foundation/fixture-root';
	import type { FixtureStyle } from 'apple/appkit/fixture-style';
	import type { Rectangle } from 'apple/appkit/rectangle';

	/**
	 * NSFixturePanel
	 * @objc class NSFixturePanel
	 */
	export class FixturePanel extends FixtureRoot {

		/**
		 * -[NSFixturePanel configureMode:style:]
		 * @objc method configureMode:style: 0:enum(Calm=0,Loud=7) 1.style:options(Plain=0,Bright=2,Quiet=8,Flipped=9223372036854775808) -> void
		 */
		configureMode(mode: FixtureMode, options: { readonly style: readonly FixtureStyle[] }): void;

		/**
		 * -[NSFixturePanel initWithTitle:]
		 * @objc init initWithTitle: 0.title:string
		 */
		constructor(options: { readonly title: string });

		/**
		 * -[NSFixturePanel initWithRecord:]
		 * @objc init initWithRecord: 0:object
		 */
		constructor(record: FixtureRecord);

		/**
		 * +[NSFixturePanel currentPanel]
		 * @objc static currentPanel -> object
		 */
		static current(): FixturePanel;

		/**
		 * -[NSFixturePanel paint]
		 * @objc method paint -> void
		 */
		paint(): void;

		/**
		 * -[NSFixturePanel scale]
		 * @objc method scale -> double
		 */
		scale(): number;

		/**
		 * -[NSFixturePanel setFrame:display:animate:]
		 * @objc method setFrame:display:animate: 0:rectangle 1.display:boolean 1.animate:boolean -> void
		 */
		setFrame(rectangle: Rectangle, options: { readonly display: boolean; readonly animate: boolean }): void;

		/**
		 * -[NSFixturePanel showCount:]
		 * @objc method showCount: 0:integer -> void
		 */
		show(count: number): void;

		/**
		 * -[NSFixturePanel showText:]
		 * @objc method showText: 0:string -> void
		 */
		show(text: string): void;

		/**
		 * -[NSFixturePanel takeRecord:]
		 * @objc method takeRecord: 0:object? -> void
		 */
		take(record: FixtureRecord | undefined): void;
	}
}
