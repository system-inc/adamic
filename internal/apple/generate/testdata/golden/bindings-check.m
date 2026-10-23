// Generated header witness. Compile with clang -fsyntax-only -Werror.
#include "umbrella.h"
#include <stddef.h>

typedef struct adamic_apple_rectangle { double x, y, width, height; } adamic_apple_rectangle;
_Static_assert(sizeof(long) == 8 && sizeof(void *) == 8, "the bridge requires a 64-bit Apple ABI");
_Static_assert(NSFixtureModeCalm == 0, "NSFixtureModeCalm enum value");
_Static_assert(NSFixtureModeLoud == 7, "NSFixtureModeLoud enum value");
_Static_assert(NSFixtureStylePlain == 0UL, "NSFixtureStylePlain enum value");
_Static_assert(NSFixtureStyleBright == 2UL, "NSFixtureStyleBright enum value");
_Static_assert(NSFixtureStyleQuiet == 8UL, "NSFixtureStyleQuiet enum value");
_Static_assert(NSFixtureStyleFlipped == 9223372036854775808UL, "NSFixtureStyleFlipped enum value");

void adamic_binding_check_0(long argument0, long argument1) {
	_Static_assert(__builtin_types_compatible_p(long, NSInteger) || __builtin_types_compatible_p(long long, NSInteger), "NSFixtureAdd parameter 0 ABI");
	_Static_assert(__builtin_types_compatible_p(long, NSInteger) || __builtin_types_compatible_p(long long, NSInteger), "NSFixtureAdd parameter 1 ABI");
	_Static_assert(__builtin_types_compatible_p(long, __typeof__(NSFixtureAdd(argument0, argument1))) || __builtin_types_compatible_p(long long, __typeof__(NSFixtureAdd(argument0, argument1))), "NSFixtureAdd result ABI");
	long result = NSFixtureAdd(argument0, argument1);
	(void)result;
}

void adamic_binding_check_1(void) {
	id result = NSFixtureCreateText();
	(void)result;
}

void adamic_binding_check_2(NSError * _Nonnull receiver) {
	_Static_assert(__builtin_types_compatible_p(long, __typeof__([receiver code])) || __builtin_types_compatible_p(long long, __typeof__([receiver code])), "code result ABI");
	long result = [receiver code];
	(void)result;
}

void adamic_binding_check_3(NSFixtureAfterInclude * _Nonnull receiver) {
	[receiver afterInclude];
}

void adamic_binding_check_4(NSFixtureLoose * _Nonnull receiver, id argument0) {
	[receiver acceptText:argument0];
}

void adamic_binding_check_5(NSFixtureMutableText * _Nonnull receiver, id argument0) {
	[receiver appendText:argument0];
}

void adamic_binding_check_6(NSFixtureNote * _Nonnull receiver) {
	id result = [receiver body];
	(void)result;
}

void adamic_binding_check_7(void) {
	id result = [NSFixturePanel currentPanel];
	(void)result;
}

void adamic_binding_check_8(NSFixturePanel * _Nonnull receiver, long argument0, unsigned long argument1) {
	_Static_assert(__builtin_types_compatible_p(long, NSFixtureMode) || __builtin_types_compatible_p(unsigned long, NSFixtureMode), "configureMode:style: parameter 0 ABI");
	_Static_assert(__builtin_types_compatible_p(unsigned long, NSFixtureStyle) || __builtin_types_compatible_p(unsigned long long, NSFixtureStyle), "configureMode:style: parameter 1 ABI");
	[receiver configureMode:argument0 style:argument1];
}

void adamic_binding_check_9(NSFixturePanel * _Nonnull receiver, id argument0) {
	id result = [receiver initWithRecord:argument0];
	(void)result;
}

void adamic_binding_check_10(NSFixturePanel * _Nonnull receiver, id argument0) {
	id result = [receiver initWithTitle:argument0];
	(void)result;
}

void adamic_binding_check_11(NSFixturePanel * _Nonnull receiver) {
	[receiver paint];
}

void adamic_binding_check_12(NSFixturePanel * _Nonnull receiver) {
	_Static_assert(__builtin_types_compatible_p(double, __typeof__([receiver scale])), "scale result ABI");
	double result = [receiver scale];
	(void)result;
}

void adamic_binding_check_13(NSFixturePanel * _Nonnull receiver, adamic_apple_rectangle argument0, BOOL argument1, BOOL argument2) {
	_Static_assert(sizeof(NSRect) == sizeof(adamic_apple_rectangle) && _Alignof(NSRect) == _Alignof(adamic_apple_rectangle), "rectangle ABI");
	_Static_assert(offsetof(NSRect, origin.x) == offsetof(adamic_apple_rectangle, x) && __builtin_types_compatible_p(__typeof__(((NSRect *)0)->origin.x), double), "rectangle x");
	_Static_assert(offsetof(NSRect, origin.y) == offsetof(adamic_apple_rectangle, y) && __builtin_types_compatible_p(__typeof__(((NSRect *)0)->origin.y), double), "rectangle y");
	_Static_assert(offsetof(NSRect, size.width) == offsetof(adamic_apple_rectangle, width) && __builtin_types_compatible_p(__typeof__(((NSRect *)0)->size.width), double), "rectangle width");
	_Static_assert(offsetof(NSRect, size.height) == offsetof(adamic_apple_rectangle, height) && __builtin_types_compatible_p(__typeof__(((NSRect *)0)->size.height), double), "rectangle height");
	NSRect header0;
	__builtin_memcpy(&header0, &argument0, sizeof(header0));
	_Static_assert(__builtin_types_compatible_p(BOOL, BOOL), "setFrame:display:animate: parameter 1 ABI");
	_Static_assert(__builtin_types_compatible_p(BOOL, BOOL), "setFrame:display:animate: parameter 2 ABI");
	[receiver setFrame:header0 display:argument1 animate:argument2];
}

void adamic_binding_check_14(NSFixturePanel * _Nonnull receiver, long argument0) {
	_Static_assert(__builtin_types_compatible_p(long, NSInteger) || __builtin_types_compatible_p(long long, NSInteger), "showCount: parameter 0 ABI");
	[receiver showCount:argument0];
}

void adamic_binding_check_15(NSFixturePanel * _Nonnull receiver, id argument0) {
	[receiver showText:argument0];
}

void adamic_binding_check_16(NSFixturePanel * _Nonnull receiver, id argument0) {
	[receiver takeRecord:argument0];
}

void adamic_binding_check_17(id<NSFixtureReadable> _Nonnull receiver) {
	id result = [receiver readText];
	(void)result;
}

void adamic_binding_check_18(NSFixtureRecord * _Nonnull receiver) {
	id result = [receiver copyRecord];
	(void)result;
}

void adamic_binding_check_19(NSFixtureRecord * _Nonnull receiver) {
	id result = [receiver displayText];
	(void)result;
}

void adamic_binding_check_20(NSFixtureRecord * _Nonnull receiver, long argument0) {
	_Static_assert(__builtin_types_compatible_p(long, NSInteger) || __builtin_types_compatible_p(long long, NSInteger), "initWithCount: parameter 0 ABI");
	id result = [receiver initWithCount:argument0];
	(void)result;
}

void adamic_binding_check_21(NSFixtureRecord * _Nonnull receiver) {
	_Static_assert(__builtin_types_compatible_p(BOOL, __typeof__([receiver isReady])), "isReady result ABI");
	BOOL result = [receiver isReady];
	(void)result;
}

void adamic_binding_check_22(NSFixtureRecord * _Nonnull receiver) {
	id result = [receiver label];
	(void)result;
}

void adamic_binding_check_23(NSFixtureRecord * _Nonnull receiver) {
	id result = [receiver optionalText];
	(void)result;
}

void adamic_binding_check_24(NSFixtureRecord * _Nonnull receiver) {
	[receiver phoneOnly];
}

void adamic_binding_check_25(NSFixtureRecord * _Nonnull receiver) {
	[receiver rawAction];
}

void adamic_binding_check_26(NSFixtureRecord * _Nonnull receiver) {
	id result = [receiver readText];
	(void)result;
}

void adamic_binding_check_27(NSFixtureRecord * _Nonnull receiver) {
	[receiver renamedAction];
}

void adamic_binding_check_28(NSFixtureRecord * _Nonnull receiver, id argument0) {
	[receiver replaceText:argument0];
}

void adamic_binding_check_29(NSFixtureRecord * _Nonnull receiver, id argument0) {
	[receiver setLabel:argument0];
}

void adamic_binding_check_30(NSFixtureRecord * _Nonnull receiver, id argument0) {
	[receiver setSource:argument0];
}

void adamic_binding_check_31(NSFixtureRecord * _Nonnull receiver) {
	id result = [receiver source];
	(void)result;
}

void adamic_binding_check_32(NSFixtureRecord * _Nonnull receiver, id argument0) {
	[receiver takeObject:argument0];
}

void adamic_binding_check_33(NSFixtureRecord * _Nonnull receiver, id argument0) {
	[receiver takeProtocol:argument0];
}

void adamic_binding_check_34(NSFixtureRecord * _Nonnull receiver, id argument0) {
	[receiver takeText:argument0];
}

void adamic_binding_check_35(NSFixtureRecord * _Nonnull receiver, void (^argument0)(id)) {
	[receiver withCompletion:argument0];
}

void adamic_binding_check_36(void) {
	id result = [NSFixtureRoot alloc];
	(void)result;
}

void adamic_binding_check_37(NSFixtureStickyNote * _Nonnull receiver, id argument0) {
	[receiver stickTo:argument0];
}

void adamic_binding_check_38(id<NSFixtureTableSource> _Nonnull receiver, id argument0) {
	_Static_assert(__builtin_types_compatible_p(long, __typeof__([receiver numberOfRowsInFixtureTable:argument0])) || __builtin_types_compatible_p(long long, __typeof__([receiver numberOfRowsInFixtureTable:argument0])), "numberOfRowsInFixtureTable: result ABI");
	long result = [receiver numberOfRowsInFixtureTable:argument0];
	(void)result;
}

void adamic_binding_check_39(NSFixtureText * _Nonnull receiver) {
	[receiver futureWay];
}

void adamic_binding_check_40(NSFixtureText * _Nonnull receiver, id argument0) {
	id result = [receiver initWithText:argument0];
	(void)result;
}

void adamic_binding_check_41(NSFixtureText * _Nonnull receiver) {
	[receiver phoneWay];
}

void adamic_binding_check_42(NSFixtureText * _Nonnull receiver) {
	id result = [receiver text];
	(void)result;
}

void adamic_binding_check_43(NSFixtureTimer * _Nonnull receiver, id argument0) {
	id result = [receiver initWithTarget:argument0];
	(void)result;
}

void adamic_binding_check_44(NSFixtureTimer * _Nonnull receiver) {
	_Static_assert(__builtin_types_compatible_p(long, __typeof__([receiver interval])) || __builtin_types_compatible_p(long long, __typeof__([receiver interval])), "interval result ABI");
	long result = [receiver interval];
	(void)result;
}

void adamic_binding_check_45(NSString * _Nonnull receiver) {
	_Static_assert(__builtin_types_compatible_p(long, __typeof__([receiver length])) || __builtin_types_compatible_p(long long, __typeof__([receiver length])), "length result ABI");
	long result = [receiver length];
	(void)result;
}

#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wprotocol"
#pragma clang diagnostic ignored "-Wobjc-protocol-property-synthesis"
#pragma clang diagnostic ignored "-Wobjc-property-implementation"
__attribute__((objc_root_class))
@interface AdamicImplementCheck0 <NSFixtureReadable>
@end
@implementation AdamicImplementCheck0
- (void)optionalPing { }
- (id)readText { return 0; }
@end
_Static_assert(__builtin_types_compatible_p(long, NSInteger) || __builtin_types_compatible_p(long long, NSInteger), "fixtureTable:shouldSelectRow: implemented ABI");
_Static_assert(__builtin_types_compatible_p(long, NSInteger) || __builtin_types_compatible_p(long long, NSInteger), "fixtureTable:textForRow: implemented ABI");
_Static_assert(__builtin_types_compatible_p(long, NSInteger) || __builtin_types_compatible_p(long long, NSInteger), "numberOfRowsInFixtureTable: implemented ABI");
__attribute__((objc_root_class))
@interface AdamicImplementCheck1 <NSFixtureTableSource>
@end
@implementation AdamicImplementCheck1
- (BOOL)fixtureTable:(id)argument0 shouldSelectRow:(NSInteger)argument1 { return 0; }
- (id)fixtureTable:(id)argument0 textForRow:(NSInteger)argument1 { return 0; }
- (void)fixtureTableDidReload:(id)argument0 { }
- (NSInteger)numberOfRowsInFixtureTable:(id)argument0 { return 0; }
@end
#pragma clang diagnostic pop
