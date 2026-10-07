// SwiftUI's views, through swiftui.swift's Objective-C classes (internal/native/apple): a seed
// binding, until the generator writes the shim and these from SwiftUI's interface (#7xv3pcs).

declare module 'apple/swiftui/views' {
	/** Font.TextStyle */
	export type FontStyle = 'LargeTitle' | 'Title' | 'Headline' | 'Body' | 'Caption' | 'Caption2';

	/**
	 * a SwiftUI view, made by the functions below and changed by its modifiers
	 * @objc class AdamicSwiftUIView
	 */
	export class View {
		private constructor();

		/**
		 * .padding(_:)
		 * @objc method padding: 0:double -> object
		 */
		padding(amount: number): View;

		/**
		 * .font(_:)
		 * @objc method font: 0:enum(LargeTitle=0,Title=1,Headline=2,Body=3,Caption=4,Caption2=5) -> object
		 */
		font(style: FontStyle): View;
	}

	/**
	 * Text(verbatim:)
	 * @objc send AdamicSwiftUIView text: 0:string -> object
	 */
	export function text(content: string): View;

	/**
	 * Button(_:action:): the action runs on the main thread when the button is pressed.
	 * @objc send AdamicSwiftUIView button:action: 0:string 1:block() -> object
	 */
	export function button(title: string, action: () => void): View;

	/**
	 * VStack(spacing:content:)
	 * @objc send AdamicSwiftUIView verticalStack:children: 0.spacing?:double=0 1:objects -> object
	 */
	export function verticalStack(options: { readonly spacing?: number }, children: readonly View[]): View;

	/**
	 * HStack(spacing:content:)
	 * @objc send AdamicSwiftUIView horizontalStack:children: 0.spacing?:double=0 1:objects -> object
	 */
	export function horizontalStack(options: { readonly spacing?: number }, children: readonly View[]): View;

	/**
	 * Spacer()
	 * @objc send AdamicSwiftUIView spacer -> object
	 */
	export function spacer(): View;
}
