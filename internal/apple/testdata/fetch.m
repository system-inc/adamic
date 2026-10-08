// fetch.m: fetch.a's witness, the same requests in Objective-C, each completion hopping to the main
// queue as Adamic's do (witness_test.go). Built with ARC; nothing here is Adamic.

#import <Foundation/Foundation.h>
#include <stdio.h>

int main(int argc, char **argv) {
	@autoreleasepool {
		NSString *address = argc > 1 ? @(argv[1]) : @"http://127.0.0.1:1";
		NSURL *stack = [[NSURL alloc] initWithString:[address stringByAppendingString:@"/stack.json"]];
		printf("url: %s\n", stack.absoluteString == nil ? "none" : "made");

		__block int completions = 0;
		NSURLSessionDataTask *refused = [NSURLSession.sharedSession dataTaskWithURL:[[NSURL alloc] initWithString:@"http://127.0.0.1:1/refused"] completionHandler:^(NSData *data, NSURLResponse *response, NSError *error) {
			dispatch_async(dispatch_get_main_queue(), ^{
				completions += 1;
				printf("refused: %s, %s, %s\n", data == nil ? "no data" : "data", response == nil ? "no response" : "a response", error == nil ? "no error" : [NSString stringWithFormat:@"error %ld", (long)error.code].UTF8String);
				CFRunLoopStop(CFRunLoopGetMain());
			});
		}];

		NSURLSessionDataTask *fetched = [NSURLSession.sharedSession dataTaskWithURL:stack completionHandler:^(NSData *data, NSURLResponse *response, NSError *error) {
			dispatch_async(dispatch_get_main_queue(), ^{
				completions += 1;
				if (error != nil) {
					printf("error: %s\n", error.localizedDescription.UTF8String);
				} else if (data != nil && response != nil) {
					printf("bytes: %lu, type: %s, expected: %lld\n", (unsigned long)data.length, response.MIMEType.UTF8String, response.expectedContentLength);
					NSString *text = [[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
					printf("text: %s\n", text == nil ? "not UTF-8" : text.UTF8String);
				}
				[refused resume];
			});
		}];
		[fetched resume];
		CFRunLoopRun();
		printf("completions: %d\n", completions);
	}
	return 0;
}
