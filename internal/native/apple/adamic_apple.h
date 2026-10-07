// adamic_apple.h: what a program that calls Apple's frameworks needs, beside the runtime (apple.c,
// docs/apple.md). The compiler writes each call as a typed objc_msgSend (internal/native/foreign.go);
// these are the conversions on either side of it.

#ifndef ADAMIC_APPLE_H
#define ADAMIC_APPLE_H

#include "adamic.h"

#include <objc/message.h>
#include <objc/runtime.h>

// libobjc's own retain and release, the ones ARC calls: stable, and cheaper than a message.
id objc_retain(id object);
void objc_release(id object);

// adamic_apple_rectangle is a CGRect, laid out as CGRect is on a 64-bit Apple platform, where CGFloat
// is a double: so it crosses objc_msgSend exactly as one, without Core Graphics' header.
typedef struct adamic_apple_rectangle {
	double x, y, width, height;
} adamic_apple_rectangle;

// adamic_apple_integer and adamic_apple_unsigned are a number as an NSInteger or an NSUInteger:
// truncated toward zero, NaN as 0, and anything past the type's range as its nearest end, where C's
// own conversion would be undefined.
static inline long adamic_apple_integer(double value) {
	if (value != value) {
		return 0;
	}
	if (value >= 9223372036854775807.0) {
		return 9223372036854775807L;
	}
	if (value <= -9223372036854775808.0) {
		return -9223372036854775807L - 1;
	}
	return (long)value;
}

static inline unsigned long adamic_apple_unsigned(double value) {
	if (value != value || value <= 0) {
		return 0;
	}
	if (value >= 18446744073709551615.0) {
		return 18446744073709551615UL;
	}
	return (unsigned long)value;
}

// adamic_apple_let_go releases a reference the conversions made for one call (an NSString made for an
// argument, an action once its owner keeps it). A counted build counts each one made and each let go
// of, and a finished program must owe none (apple.c): macOS's leaks tool can't see a leaked object
// inside an AppKit process, so this is what holds the conversions to their releases.
void adamic_apple_let_go(id object);

// adamic_apple_box is an Objective-C object as Adamic holds it: one strong reference, let go of when
// Adamic's last reference goes. retained says the object came with a reference the box takes over
// (alloc, new, copy); otherwise the box retains it. nil is undefined, a NULL box.
adamic_object *adamic_apple_box(id object, bool retained);

// adamic_apple_unbox is what a box holds, nil for undefined.
id adamic_apple_unbox(const adamic_object *box);

// adamic_apple_constructed is what an init made, made safe to hold by count: a window that would
// release itself when closed is told not to, since Adamic's count owns it.
id adamic_apple_constructed(id object);

// adamic_apple_class is a class by name, and panics when the running system has no such class.
id adamic_apple_class(const char *name);

// adamic_apple_string is a string as an NSString, retained, for the caller to release; and
// adamic_apple_string_from is an NSString as a string the caller owns, "" for nil. Both go by UTF-16
// units, so every string crosses exactly, a lone surrogate included.
id adamic_apple_string(const adamic_string *text);
adamic_string *adamic_apple_string_from(id string);

// adamic_apple_rectangle_from reads { x, y, width, height }.
adamic_apple_rectangle adamic_apple_rectangle_from(const adamic_object *object);

// adamic_apple_enumeration is the value a string literal names, and adamic_apple_options the bits an
// array of them names. The checker proved each is one of the names.
long adamic_apple_enumeration(const adamic_string *name, size_t count, const char *const names[], const long values[]);
unsigned long adamic_apple_options(const adamic_array *names, size_t count, const char *const names_known[], const long values[]);

// adamic_apple_action is a closure as an Objective-C control's target, retained for the caller, with
// adamic_apple_action_selector its action. adamic_apple_keep makes owner hold kept for as long as owner
// lives, under key: a control holds its target only weakly, so something has to.
id adamic_apple_action(adamic_closure *closure);
SEL adamic_apple_action_selector(void);
void adamic_apple_keep(id owner, id kept, const void *key);

#endif
