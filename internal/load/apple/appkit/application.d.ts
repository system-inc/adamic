// AppKit's application: a seed binding, until the generator writes these (#qxe07rq).

declare module 'apple/appkit/application' {
	/**
	 * NSApplication
	 * @objc class NSApplication
	 */
	export class Application {
		private constructor();

		/**
		 * +[NSApplication sharedApplication]
		 * @objc get sharedApplication -> object
		 */
		static readonly shared: Application;

		/**
		 * -[NSApplication setActivationPolicy:]
		 * @objc method setActivationPolicy: 0:enum(Regular=0,Accessory=1,Prohibited=2) -> boolean
		 */
		setActivationPolicy(policy: 'Regular' | 'Accessory' | 'Prohibited'): boolean;

		/**
		 * -[NSApplication activate]
		 * @objc method activate -> void
		 */
		activate(): void;

		/**
		 * -[NSApplication run]
		 * @objc method run -> void
		 */
		run(): void;

		/**
		 * -[NSApplication terminate:]
		 * @objc method terminate: nil:object? -> void
		 */
		terminate(): void;
	}
}
