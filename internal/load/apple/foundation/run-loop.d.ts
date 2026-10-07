// The main thread's run loop, through Core Foundation: a seed binding, until the generator writes
// these (#qxe07rq).

declare module 'apple/foundation/run-loop' {
	/**
	 * CFRunLoopRef
	 * @objc class __NSCFType
	 */
	export class RunLoop {
		private constructor();

		/**
		 * CFRunLoopGetMain()
		 * @objc function CFRunLoopGetMain -> object
		 */
		static readonly main: RunLoop;

		/**
		 * CFRunLoopRun(): runs the current thread's loop until it's stopped.
		 * @objc function CFRunLoopRun -> void
		 */
		static run(): void;

		/**
		 * CFRunLoopStop(loop)
		 * @objc function CFRunLoopStop -> void
		 */
		stop(): void;
	}
}
