// Members a subclass shares with what it inherits, and the other shapes the real Foundation, AppKit
// and CoreGraphics held. Original declarations, compiled only by their test.
#include <Foundation/Foundation.h>
#define FIXTURE_OBSOLETE(x, y, z) __attribute__((availability(macos, introduced = x, deprecated = y, obsoleted = z)))
#define DEPRECATED_IN_FIXTURE_VERSION_10_1_AND_LATER __attribute__((availability(macos, deprecated = 10.1)))
FIXTURE_NONNULL_BEGIN
@interface NSObject : NSFixtureRoot
+ (void)initialize;
@end
// An informal protocol: a delegate's methods, declared on the root class.
@interface NSObject (NSFixtureInformal)
- (void)fixtureDidFinish:(id)sender;
@end
@interface NSFixtureCell : NSObject
@end
// An enumeration over NSUInteger, and option bits over unsigned long long (NSBackingStoreType, NSEventMask).
typedef enum NSFixtureBacking : NSUInteger NSFixtureBacking;
enum NSFixtureBacking : NSUInteger { NSFixtureBackingRetained = 0, NSFixtureBackingBuffered = 2 };
typedef enum __attribute__((flag_enum)) NSFixtureMask : unsigned long long NSFixtureMask;
enum __attribute__((flag_enum)) NSFixtureMask : unsigned long long { NSFixtureMaskDown = 1ULL << 1, NSFixtureMaskUp = 1ULL << 2 };
@interface NSFixtureBase : NSObject
- (void)removeItem:(NSString *)item NS_SWIFT_NAME(remove(_:));
- (nullable NSFixtureCell *)selectedCell;
@property(readonly, copy) NSString *title;
@property(class, readonly) NSFixtureBase *shared;
@property(readonly) NSFixtureCell *cell;
// Any object, as every action's sender is.
- (void)add:(nullable id)sender;
- (void)addObject:(id)object;
- (void)storeBacking:(NSFixtureBacking)backing mask:(NSFixtureMask)mask;
- (void)captureOld FIXTURE_OBSOLETE(10.5, 14.0, 15.0);
- (void)drawOld DEPRECATED_IN_FIXTURE_VERSION_10_1_AND_LATER;
@end
// Named to sort before its superclass, so it is written after it only because ancestors come first.
@interface NSFixtureAdvanced : NSFixtureBase
// An unlabeled overload of an inherited name: the class re-declares the inherited one beside it.
- (void)removeCount:(NSInteger)count NS_SWIFT_NAME(remove(_:));
// The ancestor's method stands over this property, and its nonnull title over this nullable one.
@property(readonly, nullable) NSFixtureCell *selectedCell;
@property(readonly, copy, nullable) NSString *title;
// The ancestor's class property stands over this class method.
+ (NSFixtureAdvanced *)shared;
// An inherited property's name: this method folds its first label in.
- (nullable NSFixtureCell *)cellAtRow:(NSInteger)row column:(NSInteger)column;
@end
FIXTURE_NONNULL_END
