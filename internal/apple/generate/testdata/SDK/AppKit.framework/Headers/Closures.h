// Closures Apple calls, setters whose getters can't cross, and names a subclass must keep: each a
// pattern the real AppKit and Foundation held. Original declarations, compiled only by their test.
#include <AppKit/AppKit.h>
#define FIXTURE_SENDABLE __attribute__((swift_attr("@Sendable")))
typedef long long NSFixtureCount;
FIXTURE_NONNULL_BEGIN
@interface NSFixtureRequest : NSFixtureRoot
@end
@interface NSFixtureTask : NSFixtureRoot
@end
@interface NSFixtureSession : NSFixtureRoot
// A completion block, its parameters named in the header (dataTaskWithURL:completionHandler:), and
// the same task without one: the overload that takes the closure is written first.
- (NSFixtureTask *)taskWithRequest:(NSFixtureRequest *)request;
- (NSFixtureTask *)taskWithRequest:(NSFixtureRequest *)request completionHandler:(void (FIXTURE_SENDABLE ^)(NSString * _Nullable text, NSFixtureRequest * _Nullable request, long long length))completionHandler;
// A block returning a value isn't carried.
- (void)sortWith:(long (^)(NSString *left, NSString *right))comparator;
// A 64-bit count spelled long long (NSURLResponse's expectedContentLength).
@property(readonly) long long expectedLength;
// A typedef over long long (int64_t, NSProgress's counts).
@property(readonly) NSFixtureCount completedCount;
@end
@interface NSFixtureControl : NSFixturePanel
// A target and its action are one closure (buttonWithTitle:target:action:).
+ (instancetype)controlWithTitle:(NSString *)title target:(nullable id)target action:(nullable SEL)action;
// A setter for a getter whose rectangle can't come back (NSView's frame).
@property NSRect frame;
// An unlabeled method whose shortened name an ancestor holds under another selector keeps its
// words (NSControl's drawCell: beside NSView's drawRect:).
- (void)paintRecord:(NSFixtureRecord *)record;
@end
FIXTURE_NONNULL_END
