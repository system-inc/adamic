// await.m: await.a's witness, the same requests in Objective-C. Each completion hops to the main
// queue, and the program waits for one by running the main run loop until it's in, which is what an
// await of it does. Built with ARC; nothing here is Adamic.

#import <Foundation/Foundation.h>
#include <stdio.h>

@interface Fetch : NSObject
@property BOOL done;
@property long status;
@property(strong) NSString *body;
@property(strong) NSError *error;
@end

@implementation Fetch
@end

static Fetch *start(NSString *address) {
	Fetch *fetch = [Fetch new];
	NSURL *url = [[NSURL alloc] initWithString:address];
	[[NSURLSession.sharedSession dataTaskWithURL:url completionHandler:^(NSData *data, NSURLResponse *response, NSError *error) {
		dispatch_async(dispatch_get_main_queue(), ^{
			fetch.error = error;
			fetch.status = [response isKindOfClass:[NSHTTPURLResponse class]] ? ((NSHTTPURLResponse *)response).statusCode : 0;
			fetch.body = data == nil ? @"" : [[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
			fetch.done = YES;
		});
	}] resume];
	return fetch;
}

static Fetch *awaited(Fetch *fetch) {
	while (!fetch.done) {
		CFRunLoopRunInMode(kCFRunLoopDefaultMode, 1.0e10, true);
	}
	return fetch;
}

static long stack(NSString *address) {
	Fetch *response = awaited(start([address stringByAppendingString:@"/stack.json"]));
	printf("stack: %ld, %lu characters\n", response.status, (unsigned long)response.body.length);
	NSArray *decoded = [NSJSONSerialization JSONObjectWithData:[response.body dataUsingEncoding:NSUTF8StringEncoding] options:0 error:nil];
	for (NSDictionary *stacked in decoded) {
		printf("%s: %g mg, %s\n", [stacked[@"name"] UTF8String], [stacked[@"milligrams"] doubleValue], [stacked[@"timing"] UTF8String]);
	}
	return (long)decoded.count;
}

int main(int argc, char **argv) {
	@autoreleasepool {
		NSString *address = argc > 1 ? @(argv[1]) : @"http://127.0.0.1:1";
		Fetch *missing = start([address stringByAppendingString:@"/missing.json"]);
		printf("asked\n");
		long stacked = stack(address);
		Fetch *absent = awaited(missing);
		printf("missing: %ld, %lu characters\n", absent.status, (unsigned long)absent.body.length);
		Fetch *refused = awaited(start(@"http://127.0.0.1:1/refused"));
		if (refused.error != nil) {
			printf("refused: %s\n", refused.error.localizedDescription.UTF8String);
		} else {
			printf("refused: %ld\n", refused.status);
		}
		printf("stacked: %ld\n", stacked);
	}
	return 0;
}
