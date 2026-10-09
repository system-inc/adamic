// window.m: window.a's witness, the same calls in Objective-C, printing the same lines
// (witness_test.go). Built with ARC; nothing here is Adamic.

#import <AppKit/AppKit.h>
#include <stdio.h>

static void print(NSString *line) {
	printf("%s\n", line.UTF8String);
}

// Presser is the button's target: its press: runs a block, as Adamic's action runs a closure.
@interface Presser : NSObject
@property(copy) void (^block)(void);
- (void)press:(id)sender;
@end

@implementation Presser
- (void)press:(id)sender {
	(void)sender;
	self.block();
}
@end

int main(void) {
	@autoreleasepool {
		NSApplication *application = NSApplication.sharedApplication;
		print([NSString stringWithFormat:@"policy: %@", [application setActivationPolicy:NSApplicationActivationPolicyAccessory] ? @"true" : @"false"]);

		NSWindow *window = [[NSWindow alloc] initWithContentRect:NSMakeRect(40, 40, 320, 200) styleMask:NSWindowStyleMaskTitled | NSWindowStyleMaskClosable | NSWindowStyleMaskResizable backing:NSBackingStoreBuffered defer:YES];
		window.releasedWhenClosed = NO;
		print([NSString stringWithFormat:@"untitled: [%@]", window.title]);
		window.title = @"Adamic";
		print([NSString stringWithFormat:@"title: %@ (%lu units, %u last)", window.title, (unsigned long)window.title.length, (unsigned)[window.title characterAtIndex:5]]);

		window.title = @"Ken ✓ \U0001D538";
		print([NSString stringWithFormat:@"title: %@ (%lu units)", window.title, (unsigned long)window.title.length]);
		unichar lone[] = {'a', 0xD800, 'b'};
		window.title = [NSString stringWithCharacters:lone length:3];
		print([NSString stringWithFormat:@"lone surrogate: %lu units, %u", (unsigned long)window.title.length, (unsigned)[window.title characterAtIndex:1]]);

		unsigned long units = 0;
		for (int index = 0; index < 64; index++) {
			window.title = [NSString stringWithFormat:@"Ken \u2713 %d", index];
			units += window.title.length;
		}
		print([NSString stringWithFormat:@"units: %lu", units]);

		NSView *content = [[NSView alloc] initWithFrame:NSMakeRect(0, 0, 320, 200)];
		print([NSString stringWithFormat:@"content hidden: %@", content.hidden ? @"true" : @"false"]);
		NSTextField *label = [NSTextField labelWithString:@"ready"];
		label.frame = NSMakeRect(20, 120, 280, 24);

		__block int presses = 0;
		Presser *presser = [[Presser alloc] init];
		presser.block = ^{
			presses += 1;
			label.stringValue = [NSString stringWithFormat:@"pressed %d", presses];
		};
		NSButton *button = [NSButton buttonWithTitle:@"Press" target:presser action:@selector(press:)];
		button.frame = NSMakeRect(20, 20, 120, 32);
		print([NSString stringWithFormat:@"button: %@", button.title]);
		button.title = @"Count";
		print([NSString stringWithFormat:@"button: %@", button.title]);

		[content addSubview:label];
		[content addSubview:button];
		window.contentView = content;
		print([NSString stringWithFormat:@"content view: %@", window.contentView == nil ? @"missing" : @"set"]);
		print([NSString stringWithFormat:@"same view: %@, same application: %@, other view: %@", window.contentView == content ? @"true" : @"false", NSApplication.sharedApplication == application ? @"true" : @"false", window.contentView == (NSView *)label ? @"true" : @"false"]);

		print([NSString stringWithFormat:@"label: %@", label.stringValue]);
		[button performClick:nil];
		print([NSString stringWithFormat:@"label: %@", label.stringValue]);
		[button performClick:nil];
		[button performClick:nil];
		print([NSString stringWithFormat:@"label: %@, presses: %d", label.stringValue, presses]);

		print([NSString stringWithFormat:@"visible: %@", window.visible ? @"true" : @"false"]);
		[window close];
		print([NSString stringWithFormat:@"visible: %@", window.visible ? @"true" : @"false"]);
	}
	return 0;
}
