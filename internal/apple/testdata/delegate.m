// delegate.m: delegate.a's witness, the same classes written in Objective-C, handed to AppKit as the
// same delegates, printing the same lines (witness_test.go). Built with ARC; nothing here is Adamic.

#import <AppKit/AppKit.h>
#include <stdio.h>

static void print(NSString *line) {
	printf("%s\n", line.UTF8String);
}

@interface Keeper : NSObject <NSWindowDelegate>
@property(readonly) NSString *name;
@property long asked;
- (instancetype)initWithName:(NSString *)name;
@end

@implementation Keeper
- (instancetype)initWithName:(NSString *)name {
	if ((self = [super init])) {
		_name = name;
	}
	return self;
}

- (BOOL)windowShouldClose:(NSWindow *)sender {
	self.asked += 1;
	BOOL answer = self.asked > 1;
	print([NSString stringWithFormat:@"%@: may %@ close? %@ (asked %ld)", self.name, sender.title, answer ? @"yes" : @"not yet", self.asked]);
	return answer;
}

- (void)windowWillClose:(NSNotification *)notification {
	print([NSString stringWithFormat:@"%@: closing, %@", self.name, notification.object == nil ? @"no window" : @"its window"]);
}
@end

@interface Eager : Keeper
@end

@implementation Eager
- (BOOL)windowShouldClose:(NSWindow *)sender {
	print([NSString stringWithFormat:@"%@: %@ may close at once", self.name, sender.title]);
	return YES;
}
@end

@interface Shelf : NSObject <NSTableViewDataSource>
@property(readonly) NSMutableArray<NSString *> *names;
@end

@implementation Shelf
- (instancetype)init {
	if ((self = [super init])) {
		_names = [NSMutableArray arrayWithObjects:@"Creatine", @"Magnesium", nil];
	}
	return self;
}

- (NSInteger)numberOfRowsInTableView:(NSTableView *)tableView {
	print([NSString stringWithFormat:@"shelf: asked for rows (%ld columns)", (long)tableView.numberOfColumns]);
	return (NSInteger)self.names.count;
}
@end

static NSWindow *window(NSString *title) {
	NSWindow *made = [[NSWindow alloc] initWithContentRect:NSMakeRect(40, 40, 320, 200) styleMask:NSWindowStyleMaskTitled | NSWindowStyleMaskClosable backing:NSBackingStoreBuffered defer:YES];
	made.releasedWhenClosed = NO;
	made.title = title;
	return made;
}

int main(void) {
	@autoreleasepool {
		[NSApplication.sharedApplication setActivationPolicy:NSApplicationActivationPolicyAccessory];

		NSWindow *first = window(@"First");
		Keeper *keeper = [[Keeper alloc] initWithName:@"keeper"];
		first.delegate = keeper;
		print([NSString stringWithFormat:@"delegate: %@", first.delegate == nil ? @"none" : @"set"]);
		[first makeKeyAndOrderFront:nil];
		print([NSString stringWithFormat:@"visible: %@", first.visible ? @"true" : @"false"]);
		[first performClose:nil];
		print([NSString stringWithFormat:@"visible: %@", first.visible ? @"true" : @"false"]);
		[first performClose:nil];
		print([NSString stringWithFormat:@"visible: %@", first.visible ? @"true" : @"false"]);

		// A subclass's override answers, through a reference typed as the class it overrides.
		NSWindow *second = window(@"Second");
		Keeper *eager = [[Eager alloc] initWithName:@"eager"];
		second.delegate = eager;
		[second makeKeyAndOrderFront:nil];
		[second performClose:nil];
		print([NSString stringWithFormat:@"visible: %@", second.visible ? @"true" : @"false"]);

		// A delegate replaced is let go of; one cleared asks nothing. (AppKit holds a delegate weakly, so
		// here a local holds each; in Adamic the window does.)
		NSWindow *third = window(@"Third");
		Keeper *replaced = [[Keeper alloc] initWithName:@"replaced"];
		third.delegate = replaced;
		Keeper *kept = [[Keeper alloc] initWithName:@"kept"];
		third.delegate = kept;
		[third makeKeyAndOrderFront:nil];
		[third performClose:nil];
		third.delegate = nil;
		print([NSString stringWithFormat:@"delegate: %@", third.delegate == nil ? @"none" : @"set"]);
		[third performClose:nil];
		print([NSString stringWithFormat:@"visible: %@, kept asked %ld", third.visible ? @"true" : @"false", kept.asked]);

		NSTableView *table = [[NSTableView alloc] initWithFrame:NSMakeRect(0, 0, 200, 100)];
		Shelf *shelf = [[Shelf alloc] init];
		table.dataSource = shelf;
		print([NSString stringWithFormat:@"rows: %ld", (long)table.numberOfRows]);
		[shelf.names addObject:@"Vitamin D"];
		[table reloadData];
		print([NSString stringWithFormat:@"rows: %ld", (long)table.numberOfRows]);
		table.dataSource = nil;
	}
	return 0;
}
