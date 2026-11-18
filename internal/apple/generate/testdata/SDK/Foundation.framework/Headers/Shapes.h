// Members that share a name in Objective-C or Swift but not in Adamic's shapes, each a pattern the
// real Foundation and AppKit headers hold. Original declarations, compiled only by their test.
#include <Foundation/Foundation.h>
FIXTURE_NONNULL_BEGIN
@interface NSFixtureDate : NSFixtureRoot
@end
@interface NSFixtureZone : NSFixtureRoot
// Overloaded initializers, told apart by their labels (NSAccessibilityCustomRotor).
- (instancetype)initWithLabel:(NSString *)label count:(NSInteger)count;
- (instancetype)initWithKind:(NSInteger)kind count:(NSInteger)count;
// A factory imported as init() beside a real -init (NSAffineTransform's +transform).
+ (instancetype)zone;
- (instancetype)init;
// One selector on the class and its instances (NSBundle's pathForResource:ofType:inDirectory:).
+ (nullable NSString *)pathForResource:(NSString *)name ofType:(NSString *)extension;
- (nullable NSString *)pathForResource:(NSString *)name ofType:(NSString *)extension;
// A class property beside an instance method of its name (NSColor's highlightColor).
@property(class, readonly) NSFixtureZone *highlightZone;
- (NSFixtureZone *)highlightWithLevel:(CGFloat)level;
// A class and an instance property of one name (NSDate's timeIntervalSinceReferenceDate).
@property(class, readonly) CGFloat interval;
@property(readonly) CGFloat interval;
// A property beside a method that Swift tells apart by its label (NSTimeZone's abbreviation).
@property(readonly, copy) NSString *abbreviation;
- (nullable NSString *)abbreviationForDate:(NSFixtureDate *)date;
@property(readonly, getter=isDaylightSavingTime) BOOL daylightSavingTime;
- (BOOL)isDaylightSavingTimeForDate:(NSFixtureDate *)date;
// A zero-argument initializer Swift names only by an explicit name: left out, not fatal.
- (instancetype)initToMemory;
@property(readwrite, getter=isVertical) BOOL vertical;
@end
// A category that redeclares the class's property, narrower (NSSlider's vertical).
@interface NSFixtureZone (NSFixtureZoneVerticalGetter)
@property(readonly, getter=isVertical) BOOL vertical;
@end
FIXTURE_NONNULL_END
