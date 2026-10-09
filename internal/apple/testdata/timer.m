// timer.m: timer.a's witness, the same timer in Objective-C, its block handed the timer and nothing
// else, so it was handed one argument.
// Built with ARC; nothing here is Adamic.

#import <Foundation/Foundation.h>
#include <stdio.h>

int main(void) {
	@autoreleasepool {
		__block int fired = 0;
		NSTimer *timer = [NSTimer scheduledTimerWithTimeInterval:0 repeats:NO block:^(NSTimer *given) {
			fired += 1;
			printf("fired: %s, handed 1\n", given == nil ? "no timer" : "a timer");
			CFRunLoopStop(CFRunLoopGetMain());
		}];
		CFRunLoopRun();
		[timer invalidate];
		printf("fired %d\n", fired);
	}
	return 0;
}
