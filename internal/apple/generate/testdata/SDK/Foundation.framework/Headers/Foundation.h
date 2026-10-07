#ifndef FIXTURE_FOUNDATION_H
#define FIXTURE_FOUNDATION_H
// Original test declarations, not an excerpt from an Apple SDK.
typedef signed char BOOL;
typedef long NSInteger;
typedef unsigned long NSUInteger;
typedef double CGFloat;
#define NS_SWIFT_NAME(name) __attribute__((swift_name(#name)))
#define API_UNAVAILABLE(platform) __attribute__((availability(platform,unavailable)))
#define FIXTURE_NONNULL_BEGIN _Pragma("clang assume_nonnull begin")
#define FIXTURE_NONNULL_END _Pragma("clang assume_nonnull end")
__attribute__((objc_root_class))
@interface NSFixtureRoot
+ (instancetype _Nonnull)alloc;
@end
@interface NSString : NSFixtureRoot
@end
@interface NSError : NSFixtureRoot
@property(readonly) NSInteger code;
@end
FIXTURE_NONNULL_BEGIN
@protocol NSFixtureTableSource;
@protocol NSFixtureReadable
- (NSString *)readText;
@optional
- (void)optionalPing;
@end
@interface NSFixtureRecord : NSFixtureRoot <NSFixtureReadable>
- (instancetype)initWithCount:(NSInteger)count NS_SWIFT_NAME(init(count:));
@property(copy, getter=displayText, setter=replaceText:) NSString *text;
@property(readonly, getter=isReady) BOOL ready;
@property(readonly) NSString * _Nullable optionalText;
- (void)takeText:(NSString * _Nullable)text;
- (void)takeObject:(NSFixtureRecord *)record;
- (void)consumeRecord:(NSFixtureRecord * __attribute__((ns_consumed)))record;
- (void)release;
- (void)discardSelf __attribute__((ns_consumes_self));
- (NSString *)readText;
- (NSFixtureRecord *)copyRecord;
- (void)platformOnly API_UNAVAILABLE(macos);
- (void)phoneOnly API_UNAVAILABLE(ios);
- (NSString *)label;
- (void)setLabel:(NSString *)label;
- (void)rawAction __attribute__((swift_private));
- (void)renamedAction NS_SWIFT_NAME(performAction());
- (void)takeProtocol:(id<NSFixtureReadable>)reader;
- (int)narrowInteger;
- (float)narrowFloat;
- (void)withCompletion:(void (^)(NSString *))completion;
- (void)variadic:(NSString *)format, ...;
- (void)takeCallback:(void (*)(NSInteger))callback;
- (void)takeArray:(double[4])values;
@property(assign, nullable) id<NSFixtureTableSource> source;
@end
@protocol NSFixtureTableSource
- (NSInteger)numberOfRowsInFixtureTable:(NSFixtureRecord *)table;
@optional
- (nullable NSString *)fixtureTable:(NSFixtureRecord *)table textForRow:(NSInteger)row;
- (BOOL)fixtureTable:(NSFixtureRecord *)table shouldSelectRow:(NSInteger)row;
- (void)fixtureTableDidReload:(NSFixtureRecord *)table;
- (void)fixtureTable:(NSFixtureRecord *)table finish:(void (^)(void))done;
- (NSFixtureRecord *)copyFixtureTable:(NSFixtureRecord *)table;
@end
FIXTURE_NONNULL_END
@interface NSFixtureLoose : NSFixtureRoot
- (void)acceptText:(NSString *)text;
@end
NSInteger NSFixtureAdd(NSInteger left, NSInteger right);
NSString * _Nonnull NSFixtureCreateText(void) __attribute__((ns_returns_retained));
extern const NSInteger NSFixtureLimit;
#endif
