// crossing.m: crossing.a's loops in Objective-C with ARC, the floor (bench/apple/run.go).

#import <AppKit/AppKit.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(int argc, char **argv) {
	@autoreleasepool {
		const char *which = argc > 1 ? argv[1] : "scalar";
		long times = argc > 2 ? atol(argv[2]) : 0;
		NSWindow *window = [[NSWindow alloc] initWithContentRect:NSMakeRect(0, 0, 100, 100) styleMask:NSWindowStyleMaskTitled backing:NSBackingStoreBuffered defer:YES];
		window.releasedWhenClosed = NO;
		window.contentView = [[NSView alloc] initWithFrame:NSMakeRect(0, 0, 100, 100)];
		NSView *view = [[NSView alloc] initWithFrame:NSMakeRect(0, 0, 10, 10)];
		long seen = 0;
		if (strcmp(which, "scalar") == 0) {
			for (long index = 0; index < times; index++) {
				if (view.hidden) seen += 1;
			}
		} else if (strcmp(which, "object") == 0) {
			for (long index = 0; index < times; index++) {
				@autoreleasepool {
					if (window.contentView != nil) seen += 1;
				}
			}
		} else {
			for (long index = 0; index < times; index++) {
				@autoreleasepool {
					window.title = @"Adamic";
					seen += (long)window.title.length;
				}
			}
		}
		printf("%s %ld: %ld\n", which, times, seen);
	}
	return 0;
}
