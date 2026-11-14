#ifndef FIXTURE_APPKIT_H
#define FIXTURE_APPKIT_H
#include <Foundation/Foundation.h>
// Enum expansions are written directly so no vendor macro text is copied.
typedef enum NSFixtureMode : NSInteger NSFixtureMode;
enum NSFixtureMode : NSInteger {
 NSFixtureModeCalm = 0,
 NSFixtureModeLoud = 7
};
typedef enum __attribute__((flag_enum)) NSFixtureStyle : NSUInteger NSFixtureStyle;
enum __attribute__((flag_enum)) NSFixtureStyle : NSUInteger {
 NSFixtureStylePlain = 0,
 NSFixtureStyleBright = 1 << 1,
 NSFixtureStyleQuiet = 1 << 3,
 NSFixtureStyleFlipped = 1UL << 63
};
typedef struct CGPoint { CGFloat x; CGFloat y; } CGPoint;
typedef struct CGSize { CGFloat width; CGFloat height; } CGSize;
typedef struct CGRect { CGPoint origin; CGSize size; } CGRect;
typedef CGRect NSRect;
union NSFixtureUnion { long count; double ratio; };
@interface NSFixturePanel : NSFixtureRoot
- (instancetype _Nonnull)initWithRecord:(NSFixtureRecord * _Nonnull)record NS_SWIFT_NAME(init(_:));
- (void)setFrame:(NSRect)rectangle display:(BOOL)display animate:(BOOL)animate;
- (void)configureMode:(NSFixtureMode)mode style:(NSFixtureStyle)style;
- (void)takeRecord:(NSFixtureRecord * _Nullable)record;
- (CGFloat)scale;
- (CGRect)bounds;
- (void)takeUnion:(union NSFixtureUnion)value;
+ (NSFixturePanel * _Nonnull)currentPanel;
// Overloads Swift imports under one name, told apart by their argument types.
- (instancetype _Nonnull)initWithTitle:(NSString * _Nonnull)title NS_SWIFT_NAME(init(title:));
- (void)showText:(NSString * _Nonnull)text NS_SWIFT_NAME(show(_:));
- (void)showCount:(NSInteger)count NS_SWIFT_NAME(show(_:));
@end
@interface NSFixturePanel (Painting)
- (void)paint;
@end
#include <Other/Other.h>
@interface NSFixtureAfterInclude : NSFixtureRoot
- (void)afterInclude;
@end
#endif
