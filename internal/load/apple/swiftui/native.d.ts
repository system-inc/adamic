// SwiftUI's views, through swiftui.swift's Objective-C classes (internal/native/apple), written by
// hand until the generator writes the shim and these from SwiftUI's interface (#7xv3pcs). Programs
// use apple/swiftui/views, written in Adamic over these, where a view may be a class of their own.

declare module 'apple/swiftui/native' {
	/** Font.TextStyle */
	export type FontStyle = 'LargeTitle' | 'Title' | 'Headline' | 'Body' | 'Caption' | 'Caption2';

	/**
	 * a SwiftUI view, made by the functions below and changed by its modifiers
	 * @objc class AdamicSwiftUIView
	 */
	export class SwiftUIView {
		private constructor();

		/**
		 * .padding(_:)
		 * @objc method padding: 0:double -> object
		 */
		padding(amount: number): SwiftUIView;

		/**
		 * .font(_:)
		 * @objc method font: 0:enum(LargeTitle=0,Title=1,Headline=2,Body=3,Caption=4,Caption2=5) -> object
		 */
		font(style: FontStyle): SwiftUIView;

		/**
		 * .navigationTitle(_:)
		 * @objc method navigationTitle: 0:string -> object
		 */
		navigationTitle(title: string): SwiftUIView;

		/**
		 * .foregroundStyle(_:), a system color
		 * @objc method foregroundColor: 0:enum(Primary=0,Secondary=1,Accent=2,Red=3,Orange=4,Yellow=5,Green=6,Blue=7,Purple=8,Pink=9,Gray=10) -> object
		 */
		foregroundColor(color: SystemColor): SwiftUIView;

		/**
		 * .background(_:), a system color
		 * @objc method background: 0:enum(Primary=0,Secondary=1,Accent=2,Red=3,Orange=4,Yellow=5,Green=6,Blue=7,Purple=8,Pink=9,Gray=10) -> object
		 */
		background(color: SystemColor): SwiftUIView;

		/**
		 * .opacity(_:)
		 * @objc method opacity: 0:double -> object
		 */
		opacity(amount: number): SwiftUIView;

		/**
		 * .multilineTextAlignment(_:)
		 * @objc method multilineTextAlignment: 0:enum(Leading=0,Center=1,Trailing=2) -> object
		 */
		multilineTextAlignment(alignment: TextAlignment): SwiftUIView;

		/**
		 * .listStyle(_:)
		 * @objc method listStyle: 0:enum(Plain=0,Inset=1,Sidebar=2) -> object
		 */
		listStyle(style: ListStyle): SwiftUIView;

		/**
		 * .buttonStyle(_:)
		 * @objc method buttonStyle: 0:enum(Bordered=0,BorderedProminent=1,Borderless=2,Plain=3) -> object
		 */
		buttonStyle(style: ButtonStyle): SwiftUIView;
	}

	/** Color's system colors */
	export type SystemColor = 'Primary' | 'Secondary' | 'Accent' | 'Red' | 'Orange' | 'Yellow' | 'Green' | 'Blue' | 'Purple' | 'Pink' | 'Gray';

	/** TextAlignment */
	export type TextAlignment = 'Leading' | 'Center' | 'Trailing';

	/** ListStyle */
	export type ListStyle = 'Plain' | 'Inset' | 'Sidebar';

	/** ButtonStyle */
	export type ButtonStyle = 'Bordered' | 'BorderedProminent' | 'Borderless' | 'Plain';

	/**
	 * List { ForEach }
	 * @objc send AdamicSwiftUIView list: 0:objects -> object
	 */
	export function list(children: readonly SwiftUIView[]): SwiftUIView;

	/**
	 * Section(_:content:)
	 * @objc send AdamicSwiftUIView section:children: 0:string 1:objects -> object
	 */
	export function section(title: string, children: readonly SwiftUIView[]): SwiftUIView;

	/**
	 * NavigationStack { root }
	 * @objc send AdamicSwiftUIView navigationStack: 0:object -> object
	 */
	export function navigationStack(root: SwiftUIView): SwiftUIView;

	/**
	 * Image(systemName:), an SF Symbol
	 * @objc send AdamicSwiftUIView image: 0:string -> object
	 */
	export function image(systemName: string): SwiftUIView;

	/**
	 * ProgressView()
	 * @objc send AdamicSwiftUIView progressView -> object
	 */
	export function progressView(): SwiftUIView;

	/**
	 * Text(verbatim:)
	 * @objc send AdamicSwiftUIView text: 0:string -> object
	 */
	export function text(content: string): SwiftUIView;

	/**
	 * Button(_:action:): the action runs on the main thread when the button is pressed.
	 * @objc send AdamicSwiftUIView button:action: 0:string 1:block() -> object
	 */
	export function button(title: string, action: () => void): SwiftUIView;

	/**
	 * VStack(spacing:content:)
	 * @objc send AdamicSwiftUIView verticalStack:children: 0.spacing?:double=0 1:objects -> object
	 */
	export function verticalStack(options: { readonly spacing?: number }, children: readonly SwiftUIView[]): SwiftUIView;

	/**
	 * HStack(spacing:content:)
	 * @objc send AdamicSwiftUIView horizontalStack:children: 0.spacing?:double=0 1:objects -> object
	 */
	export function horizontalStack(options: { readonly spacing?: number }, children: readonly SwiftUIView[]): SwiftUIView;

	/**
	 * Spacer()
	 * @objc send AdamicSwiftUIView spacer -> object
	 */
	export function spacer(): SwiftUIView;
}
