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
id objc_autorelease(id object);

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
// Adamic's last reference goes, in the one box the object has while Adamic holds it, so the same
// object is the same value. retained says the object came with a reference the box takes over
// (alloc, new, copy), or gives back when the object already has a box; otherwise the box retains it.
// nil is undefined, a NULL box.
adamic_object *adamic_apple_box(id object, bool retained);

// adamic_apple_unbox is what a box holds, nil for undefined.
id adamic_apple_unbox(const adamic_object *box);

// adamic_apple_present is object, which Apple's header promises isn't nil (or which new made): nil
// there panics, naming what gave it, rather than reaching Adamic as a value its type says can't be.
id adamic_apple_present(id object, const char *what);

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

// adamic_apple_objects is an array of Apple's objects as an NSArray, retained for the caller to let go
// of (adamic_apple_let_go). Each element is a box the checker proved holds an object.
id adamic_apple_objects(const adamic_array *array);

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

// adamic_apple_block is a closure as an Objective-C block, laid out as the block ABI lays one out
// (clang's Block-ABI-Apple), with one captured value: holder, an object holding the closure, whose
// count is Objective-C's and so safe on any thread. The compiler declares, for each block type, its
// descriptor and its invoke (internal/native/foreign.go); a block starts on the stack of the call it's
// given to, and Apple copies it to keep it, which copy and dispose follow.
typedef struct adamic_apple_block_descriptor {
	unsigned long reserved;
	unsigned long size;
	void (*copy)(void *destination, const void *source);
	void (*dispose)(const void *block);
	const char *signature;
	const char *layout;
} adamic_apple_block_descriptor;

typedef struct adamic_apple_block {
	void *isa;
	int flags;
	int reserved;
	void (*invoke)(void);
	const adamic_apple_block_descriptor *descriptor;
	id holder;
} adamic_apple_block;

void adamic_apple_block_start(adamic_apple_block *block, adamic_closure *closure, const adamic_apple_block_descriptor *descriptor, void (*invoke)(void));
void adamic_apple_block_end(adamic_apple_block *block);
void adamic_apple_block_copy(void *destination, const void *source);
void adamic_apple_block_dispose(const void *block);

// A block's invoke may run on any thread, and Adamic's counts belong to the main thread: so invoke
// only retains what it was given into a call (adamic_apple_call_new), and adamic_apple_on_main runs
// the call's deliver on the main thread, at once when invoke is already there, otherwise when the
// main queue next runs. deliver makes the arguments Adamic values and calls adamic_apple_block_call.
void *adamic_apple_call_new(size_t size);
void adamic_apple_call_free(void *call);
void adamic_apple_on_main(void *call, void (*deliver)(void *));
void adamic_apple_block_call(id holder, adamic_value *arguments);

// A delegate is an object of the program's own class handed to Apple (docs/apple.md, "Delegates"):
// Apple gets an instance of an Objective-C class made for that class at runtime, a subclass of
// NSObject conforming to the protocols the class names, whose one instance variable holds the object
// (counted) and whose methods are the compiler's, one per protocol method the class has
// (internal/native/foreign.go). Each method calls the class's through an ordinary function.
typedef struct adamic_apple_delegate_method {
	const char *selector;
	IMP implementation;
	const char *types;
} adamic_apple_delegate_method;

typedef struct adamic_apple_delegate_class {
	const char *name;
	size_t protocol_count;
	const char *const *protocols;
	size_t method_count;
	const adamic_apple_delegate_method *methods;
	// made is the class, once the first instance needs it.
	Class made;
	ptrdiff_t object_offset;
} adamic_apple_delegate_class;

// adamic_apple_delegate is an instance of description's class holding object, retained for the caller
// to let go of, or nil for undefined.
id adamic_apple_delegate(adamic_object *object, adamic_apple_delegate_class *description);

// adamic_apple_delegate_object is the object a delegate holds, borrowed for the method Apple called.
// Apple calls a delegate on the main thread, where Adamic's counts are kept; anywhere else panics.
adamic_object *adamic_apple_delegate_object(id delegate, const adamic_apple_delegate_class *description);

// adamic_apple_delegate_returned ends a delegate's method: what Adamic threw is uncaught, since
// nothing in Objective-C can catch it.
void adamic_apple_delegate_returned(void);

// adamic_apple_give_back is an object a conversion made (adamic_apple_string), handed to Apple as a
// method's result: autoreleased, as a method's result is, and no longer owed.
id adamic_apple_give_back(id object);

#endif
