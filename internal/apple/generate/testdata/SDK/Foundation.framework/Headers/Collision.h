// Classes and protocols sharing an Objective-C name, compiled only by its test.
// The protocol comes first in one pair and second in the other, as in AppKit.
#include <Foundation/Foundation.h>
@protocol NSFixtureAmbiguous
- (void)doThing;
@end
@interface NSFixtureAmbiguous : NSFixtureRoot
- (void)doThing;
- (void)takeProtocol:(id<NSFixtureAmbiguous> _Nonnull)protocol;
- (void)takeClass:(NSFixtureAmbiguous * _Nonnull)value;
- (id<NSFixtureAmbiguous> _Nonnull)protocolValue;
@end
@interface NSFixtureElement : NSFixtureRoot
- (void)activate;
@end
@protocol NSFixtureElement
- (void)describe;
@end
@interface NSFixtureHolder : NSFixtureRoot
@property (readonly, nonnull) id<NSFixtureElement> element;
@end
