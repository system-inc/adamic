// settled.m: settled.a's witness. What follows the await is work put on the main queue before the
// run loop runs, which the run loop does before it waits. Built with ARC; nothing here is Adamic.

#import <Foundation/Foundation.h>
#include <stdio.h>

int main(void) {
	@autoreleasepool {
		printf("started\n");
		dispatch_async(dispatch_get_main_queue(), ^{
			printf("settled\n");
			CFRunLoopStop(CFRunLoopGetMain());
		});
		printf("running\n");
		CFRunLoopRun();
		printf("stopped\n");
	}
	return 0;
}
