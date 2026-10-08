// The main thread's run loop, through Core Foundation, which isn't generated: CFRunLoopRef is an
// opaque reference the generator doesn't bind.

declare module 'apple/corefoundation/run-loop' {
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
