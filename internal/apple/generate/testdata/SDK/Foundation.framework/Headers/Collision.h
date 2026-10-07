// An original class/protocol namespace collision, compiled only by its test.
#include <Foundation/Foundation.h>
@protocol NSFixtureAmbiguous
- (void)doThing;
@end
@interface NSFixtureAmbiguous : NSFixtureRoot
- (void)doThing;
@end
