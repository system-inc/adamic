// apple.c: Objective-C objects held by Adamic, and the conversions a call to Apple's frameworks makes
// on either side of objc_msgSend (adamic_apple.h, docs/apple.md).

#include "adamic_apple.h"

#include <CoreFoundation/CoreFoundation.h>
#include <dispatch/dispatch.h>
#include <pthread.h>
#include <objc/objc-exception.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

// libobjc's autorelease pool, as ARC uses it.
void *objc_autoreleasePoolPush(void);
void objc_autoreleasePoolPop(void *pool);

typedef void (*message_void)(id, SEL);
typedef id (*message_object)(id, SEL);
typedef BOOL (*message_boolean_object)(id, SEL, id);

#ifdef ADAMIC_COUNT
// owed is how many references the conversions made that haven't been let go of yet.
static long owed;
#define TAKEN() (owed++)
#define LET_GO() (owed--)
#else
#define TAKEN() ((void)0)
#define LET_GO() ((void)0)
#endif

void adamic_apple_let_go(id object) {
	LET_GO();
	objc_release(object);
}

// boxes is the one box each Objective-C object Adamic holds has, by the object, so two crossings of
// one object are one value (=== holds) and a crossing of an object already held makes nothing. A box
// leaves it as its last release lets the object go; until then its reference keeps the object, so
// no other object can come to have its address. Boxes are made and let go of on the main thread.
//
// The table is open addressed, probed linearly, its capacity a power of two at most half full, and
// an entry removed is a tombstone until the table is next rebuilt: a crossing costs a multiply and a
// probe or two, where a CFDictionary cost more than the message itself.
typedef struct box_entry {
	id object;
	adamic_object *box;
} box_entry;

static box_entry *boxes;
static size_t box_capacity, box_used, box_live;
static adamic_object tombstone;

static size_t box_slot(id object) {
	return (size_t)(((uintptr_t)object >> 4) * 0x9E3779B97F4A7C15ULL) & (box_capacity - 1);
}

static adamic_object *box_find(id object) {
	if (box_capacity == 0) {
		return NULL;
	}
	for (size_t slot = box_slot(object);; slot = (slot + 1) & (box_capacity - 1)) {
		if (boxes[slot].box == NULL) {
			return NULL;
		}
		if (boxes[slot].object == object && boxes[slot].box != &tombstone) {
			return boxes[slot].box;
		}
	}
}

static void box_insert(id object, adamic_object *box) {
	if (2 * (box_used + 1) > box_capacity) {
		// Rebuilt at twice the live entries (16 at least), dropping the tombstones.
		box_entry *old = boxes;
		size_t old_capacity = box_capacity;
		box_capacity = 16;
		while (box_capacity < 4 * (box_live + 1)) {
			box_capacity *= 2;
		}
		boxes = calloc(box_capacity, sizeof *boxes);
		if (boxes == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		box_used = 0;
		for (size_t index = 0; index < old_capacity; index++) {
			if (old[index].box != NULL && old[index].box != &tombstone) {
				size_t slot = box_slot(old[index].object);
				while (boxes[slot].box != NULL) {
					slot = (slot + 1) & (box_capacity - 1);
				}
				boxes[slot] = old[index];
				box_used++;
			}
		}
		free(old);
	}
	size_t slot = box_slot(object);
	while (boxes[slot].box != NULL && boxes[slot].box != &tombstone) {
		slot = (slot + 1) & (box_capacity - 1);
	}
	if (boxes[slot].box == NULL) {
		box_used++;
	}
	boxes[slot] = (box_entry){object, box};
	box_live++;
}

static void box_remove(id object) {
	for (size_t slot = box_slot(object);; slot = (slot + 1) & (box_capacity - 1)) {
		if (boxes[slot].object == object && boxes[slot].box != &tombstone) {
			boxes[slot].box = &tombstone;
			box_live--;
			return;
		}
	}
}

static void release_object(void *pointer) {
	box_remove((id)pointer);
	objc_release((id)pointer);
}

static const adamic_foreign_kind object_kind = {"Objective-C object", release_object};

// pool is the one main runs in, outside the run loop's own (the run loop drains a pool of its own per
// event). It's drained when the program finishes, once main has let go of its globals, so what was
// autoreleased into it is let go of too, and whatever Adamic gave Apple (an action's closure) comes
// back and is freed before the program's counts and leaks are taken. That's why drain is registered
// with atexit at the first object Adamic holds, inside main: exit runs what was registered last
// first, and the counts' report was registered as the program loaded.
static void *pool;
static bool draining;

static void drain(void) {
	// SwiftUI's shim keeps each button's action for tests to press, when it's linked.
	Class swiftui = objc_getClass("AdamicSwiftUIView");
	if (swiftui != Nil) {
		((message_void)objc_msgSend)((id)swiftui, sel_registerName("forgetActions"));
	}
	objc_autoreleasePoolPop(pool);
#ifdef ADAMIC_COUNT
	char line[64];
	int length = snprintf(line, sizeof line, "adamic: apple: owed %ld\n", owed);
	if (length > 0 && (size_t)length < sizeof line) {
		(void)!write(2, line, (size_t)length);
	}
#endif
}

adamic_object *adamic_apple_box(id object, bool retained) {
	if (!draining) {
		draining = true;
		atexit(drain);
	}
	if (object == nil) {
		return NULL;
	}
	adamic_object *box = box_find(object);
	if (box != NULL) {
		if (retained) {
			objc_release(object);
		}
		return adamic_retain(box);
	}
	if (!retained) {
		objc_retain(object);
	}
	box = adamic_foreign_new((void *)object, &object_kind);
	box_insert(object, box);
	return box;
}

id adamic_apple_unbox(const adamic_object *box) {
	return (id)adamic_foreign_pointer(box);
}

static void panic_text(const char *prefix, const char *detail) {
	size_t prefix_length = strlen(prefix);
	size_t detail_length = strlen(detail);
	char *message = malloc(prefix_length + detail_length + 1);
	if (message == NULL) {
		adamic_panic(prefix, prefix_length);
	}
	memcpy(message, prefix, prefix_length);
	memcpy(message + prefix_length, detail, detail_length + 1);
	adamic_panic(message, prefix_length + detail_length);
}

id adamic_apple_present(id object, const char *what) {
	if (object == nil) {
		panic_text("Apple gave nil where it promised a value: ", what);
	}
	return object;
}

id adamic_apple_class(const char *name) {
	Class class = objc_getClass(name);
	if (class == Nil) {
		panic_text("this system has no Objective-C class ", name);
	}
	return (id)class;
}

id adamic_apple_constructed(id object) {
	static Class window;
	static SEL kind, released;
	if (window == Nil) {
		window = objc_getClass("NSWindow");
		kind = sel_registerName("isKindOfClass:");
		released = sel_registerName("setReleasedWhenClosed:");
	}
	if (object != nil && window != Nil && ((message_boolean_object)objc_msgSend)(object, kind, (id)window)) {
		((void (*)(id, SEL, BOOL))objc_msgSend)(object, released, NO);
	}
	return object;
}

id adamic_apple_string(const adamic_string *text) {
	static Class string_class;
	static SEL allocate, from_bytes, from_characters;
	if (string_class == Nil) {
		string_class = (Class)adamic_apple_class("NSString");
		allocate = sel_registerName("alloc");
		from_bytes = sel_registerName("initWithBytes:length:encoding:");
		from_characters = sel_registerName("initWithCharacters:length:");
	}
	id string = ((message_object)objc_msgSend)((id)string_class, allocate);
	TAKEN();
	size_t units = adamic_string_units(text);
	if (units == text->length) {
		// ASCII: its bytes are its units. NSASCIIStringEncoding is 1.
		return ((id (*)(id, SEL, const void *, unsigned long, unsigned long))objc_msgSend)(string, from_bytes, text->bytes, text->length, 1);
	}
	uint16_t *characters = malloc(units * sizeof *characters);
	if (characters == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	for (size_t index = 0; index < units; index++) {
		characters[index] = (uint16_t)adamic_string_char_code(text, (double)index);
	}
	string = ((id (*)(id, SEL, const uint16_t *, unsigned long))objc_msgSend)(string, from_characters, characters, units);
	free(characters);
	return string;
}

adamic_string *adamic_apple_string_from(id string) {
	static SEL length_selector, characters_selector;
	if (length_selector == NULL) {
		length_selector = sel_registerName("length");
		characters_selector = sel_registerName("getCharacters:range:");
	}
	if (string == nil) {
		return adamic_string_from_char_codes(0, NULL);
	}
	unsigned long length = ((unsigned long (*)(id, SEL))objc_msgSend)(string, length_selector);
	if (length == 0) {
		return adamic_string_from_char_codes(0, NULL);
	}
	// ASCII: its units are its bytes, copied straight in, and the string says so (units is the UTF-16
	// length plus one, adamic.h), so nothing translates it again. Asking with no buffer counts the
	// characters that convert, tagged-pointer strings included, without writing.
	CFRange whole = CFRangeMake(0, (CFIndex)length);
	if (CFStringGetBytes((CFStringRef)string, whole, kCFStringEncodingASCII, 0, false, NULL, 0, NULL) == (CFIndex)length) {
		adamic_string *text = adamic_string_allocate(length);
		CFIndex written = 0;
		CFStringGetBytes((CFStringRef)string, whole, kCFStringEncodingASCII, 0, false, (UInt8 *)text->bytes, (CFIndex)length, &written);
		if (written == (CFIndex)length) {
			text->units = length + 1;
			return text;
		}
		adamic_release(text);
	}
	uint16_t *characters = malloc(length * sizeof *characters);
	double *codes = malloc(length * sizeof *codes);
	if (characters == NULL || codes == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	// An NSRange is a location and a length, two unsigned longs.
	struct {
		unsigned long location, length;
	} range = {0, length};
	((void (*)(id, SEL, uint16_t *, __typeof__(range)))objc_msgSend)(string, characters_selector, characters, range);
	for (unsigned long index = 0; index < length; index++) {
		codes[index] = characters[index];
	}
	adamic_string *result = adamic_string_from_char_codes(length, codes);
	free(characters);
	free(codes);
	return result;
}

id adamic_apple_objects(const adamic_array *array) {
	static SEL allocate, initialize;
	if (allocate == NULL) {
		allocate = sel_registerName("alloc");
		initialize = sel_registerName("initWithObjects:count:");
	}
	id *objects = adamic_apple_call_new((array->length == 0 ? 1 : array->length) * sizeof *objects);
	for (size_t index = 0; index < array->length; index++) {
		objects[index] = adamic_apple_present(adamic_apple_unbox(array->elements[index].reference), "an element of an array handed to Apple");
	}
	id result = ((message_object)objc_msgSend)(adamic_apple_class("NSArray"), allocate);
	result = ((id (*)(id, SEL, id *, unsigned long))objc_msgSend)(result, initialize, objects, array->length);
	adamic_apple_call_free(objects);
	TAKEN();
	return result;
}

static double number_field(const adamic_object *object, const char *name) {
	adamic_slot_cache cache = {NULL, 0};
	return adamic_object_field(object, name, &cache)->number;
}

adamic_apple_rectangle adamic_apple_rectangle_from(const adamic_object *object) {
	return (adamic_apple_rectangle){number_field(object, "x"), number_field(object, "y"), number_field(object, "width"), number_field(object, "height")};
}

static bool named(const adamic_string *name, const char *candidate) {
	size_t length = strlen(candidate);
	return name->length == length && memcmp(name->bytes, candidate, length) == 0;
}

long adamic_apple_enumeration(const adamic_string *name, size_t count, const char *const names[], const long values[]) {
	for (size_t index = 0; index < count; index++) {
		if (named(name, names[index])) {
			return values[index];
		}
	}
	static const char message[] = "compiler bug: a string the checker proved is one of an enumeration's names isn't";
	adamic_panic(message, sizeof message - 1);
}

unsigned long adamic_apple_options(const adamic_array *names, size_t count, const char *const names_known[], const long values[]) {
	unsigned long bits = 0;
	for (size_t index = 0; index < names->length; index++) {
		bits |= (unsigned long)adamic_apple_enumeration(names->elements[index].reference, count, names_known, values);
	}
	return bits;
}

// An action is an instance of AdamicAction, a subclass of NSObject made at runtime, whose one
// instance variable holds the closure (counted: the action holds a reference) and whose
// adamicAct: calls it.
static Ivar action_closure;

static void act(id self, SEL command, id sender) {
	(void)command;
	(void)sender;
	adamic_closure *closure = *(adamic_closure **)((char *)self + ivar_getOffset(action_closure));
	adamic_value result = closure->code(closure, NULL);
	(void)result;
	if (adamic_thrown != NULL) {
		// Nothing in Objective-C can catch what Adamic throws: it's uncaught.
		adamic_uncaught();
	}
}

static void release_closure(void *closure) {
	adamic_release(closure);
}

// An action or a block's holder may be let go of last on another thread, when Apple drops a block
// there: its closure is released on the main thread, where Adamic's counts are kept.
static void action_dealloc(id self, SEL command) {
	adamic_closure **closure = (adamic_closure **)((char *)self + ivar_getOffset(action_closure));
	if (pthread_main_np()) {
		adamic_release(*closure);
	} else {
		dispatch_async_f(dispatch_get_main_queue(), *closure, release_closure);
	}
	*closure = NULL;
	struct objc_super super = {self, class_getSuperclass(object_getClass(self))};
	((void (*)(struct objc_super *, SEL))objc_msgSendSuper)(&super, command);
}

static Class action_class(void) {
	static Class class;
	if (class != Nil) {
		return class;
	}
	class = objc_allocateClassPair((Class)adamic_apple_class("NSObject"), "AdamicAction", 0);
	if (class == Nil) {
		// Already registered by another copy of this runtime in the process.
		class = objc_getClass("AdamicAction");
		action_closure = class_getInstanceVariable(class, "closure");
		return class;
	}
	class_addIvar(class, "closure", sizeof(adamic_closure *), 3, "^v");
	class_addMethod(class, sel_registerName("adamicAct:"), (IMP)act, "v@:@");
	class_addMethod(class, sel_registerName("dealloc"), (IMP)action_dealloc, "v@:");
	objc_registerClassPair(class);
	action_closure = class_getInstanceVariable(class, "closure");
	return class;
}

id adamic_apple_action(adamic_closure *closure) {
	id action = ((message_object)objc_msgSend)((id)action_class(), sel_registerName("new"));
	*(adamic_closure **)((char *)action + ivar_getOffset(action_closure)) = adamic_retain(closure);
	TAKEN();
	return action;
}

SEL adamic_apple_action_selector(void) {
	return sel_registerName("adamicAct:");
}

// The block ABI's flags: the block has copy and dispose helpers, and a signature.
enum {
	block_has_copy_dispose = 1 << 25,
	block_has_signature = 1 << 30,
};

extern void *_NSConcreteStackBlock[32];

void adamic_apple_block_start(adamic_apple_block *block, adamic_closure *closure, const adamic_apple_block_descriptor *descriptor, void (*invoke)(void)) {
	block->isa = _NSConcreteStackBlock;
	block->flags = block_has_copy_dispose | block_has_signature;
	block->reserved = 0;
	block->invoke = invoke;
	block->descriptor = descriptor;
	block->holder = adamic_apple_action(closure);
}

void adamic_apple_block_end(adamic_apple_block *block) {
	adamic_apple_let_go(block->holder);
}

void adamic_apple_block_copy(void *destination, const void *source) {
	((adamic_apple_block *)destination)->holder = objc_retain(((const adamic_apple_block *)source)->holder);
}

void adamic_apple_block_dispose(const void *block) {
	objc_release(((const adamic_apple_block *)block)->holder);
}

void *adamic_apple_call_new(size_t size) {
	void *call = malloc(size);
	if (call == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	return call;
}

void adamic_apple_call_free(void *call) {
	free(call);
}

void adamic_apple_on_main(void *call, void (*deliver)(void *)) {
	if (pthread_main_np()) {
		deliver(call);
		return;
	}
	dispatch_async_f(dispatch_get_main_queue(), call, deliver);
}

void adamic_apple_block_call(id holder, adamic_value *arguments) {
	if (!pthread_main_np()) {
		// Adamic's counts aren't atomic: its code runs on the main thread only, and this is the door.
		static const char message[] = "runtime bug: a block called Adamic code off the main thread";
		adamic_panic(message, sizeof message - 1);
	}
	adamic_closure *closure = *(adamic_closure **)((char *)holder + ivar_getOffset(action_closure));
	adamic_value result = closure->code(closure, arguments);
	(void)result;
	if (adamic_thrown != NULL) {
		adamic_uncaught();
	}
}

// A delegate's class is made the first time an instance is: NSObject's subclass, an instance
// variable for the object, the compiler's methods, the protocols, and a dealloc that lets the
// object go.
static void delegate_dealloc(id self, SEL command) {
	Ivar held = class_getInstanceVariable(object_getClass(self), "object");
	adamic_object **object = (adamic_object **)((char *)self + ivar_getOffset(held));
	if (pthread_main_np()) {
		adamic_release(*object);
	} else {
		dispatch_async_f(dispatch_get_main_queue(), *object, release_closure);
	}
	*object = NULL;
	struct objc_super super = {self, class_getSuperclass(object_getClass(self))};
	((void (*)(struct objc_super *, SEL))objc_msgSendSuper)(&super, command);
}

static Class delegate_class(adamic_apple_delegate_class *description) {
	if (description->made != Nil) {
		return description->made;
	}
	Class class = objc_allocateClassPair((Class)adamic_apple_class("NSObject"), description->name, 0);
	if (class == Nil) {
		panic_text("an Objective-C class already has the name the compiler gave a delegate: ", description->name);
	}
	class_addIvar(class, "object", sizeof(adamic_object *), 3, "^v");
	for (size_t index = 0; index < description->method_count; index++) {
		const adamic_apple_delegate_method *method = &description->methods[index];
		class_addMethod(class, sel_registerName(method->selector), method->implementation, method->types);
	}
	class_addMethod(class, sel_registerName("dealloc"), (IMP)delegate_dealloc, "v@:");
	for (size_t index = 0; index < description->protocol_count; index++) {
		// A protocol nothing in the process has used yet isn't registered; conforming is then only
		// answering its methods, which is what Apple asks of a delegate (respondsToSelector:).
		Protocol *protocol = objc_getProtocol(description->protocols[index]);
		if (protocol != NULL) {
			class_addProtocol(class, protocol);
		}
	}
	objc_registerClassPair(class);
	description->object_offset = ivar_getOffset(class_getInstanceVariable(class, "object"));
	description->made = class;
	return class;
}

id adamic_apple_delegate(adamic_object *object, adamic_apple_delegate_class *description) {
	if (object == NULL) {
		return nil;
	}
	id delegate = ((message_object)objc_msgSend)((id)delegate_class(description), sel_registerName("new"));
	*(adamic_object **)((char *)delegate + description->object_offset) = adamic_retain(object);
	TAKEN();
	return delegate;
}

adamic_object *adamic_apple_delegate_object(id delegate, const adamic_apple_delegate_class *description) {
	if (!pthread_main_np()) {
		// Adamic's counts aren't atomic, and a delegate's method answers before Apple goes on, so it
		// can't wait for the main thread the way a block's closure does.
		panic_text("Apple called a delegate off the main thread: ", description->name);
	}
	return *(adamic_object **)((char *)delegate + description->object_offset);
}

void adamic_apple_delegate_returned(void) {
	if (adamic_thrown != NULL) {
		adamic_uncaught();
	}
}

id adamic_apple_give_back(id object) {
	LET_GO();
	return objc_autorelease(object);
}

void adamic_apple_keep(id owner, id kept, const void *key) {
	objc_setAssociatedObject(owner, key, kept, OBJC_ASSOCIATION_RETAIN_NONATOMIC);
}

// An Objective-C exception nothing catches ends the program the way an Adamic panic does, with its
// name and reason, before anything unwinds through Adamic's frames: libobjc searches for a handler
// first, finds none, and calls this.
static void uncaught(id exception) {
	SEL name = sel_registerName("name"), reason = sel_registerName("reason"), utf8 = sel_registerName("UTF8String");
	id name_string = ((message_object)objc_msgSend)(exception, name);
	id reason_string = ((message_object)objc_msgSend)(exception, reason);
	const char *name_text = name_string == nil ? "?" : ((const char *(*)(id, SEL))objc_msgSend)(name_string, utf8);
	const char *reason_text = reason_string == nil ? "" : ((const char *(*)(id, SEL))objc_msgSend)(reason_string, utf8);
	size_t length = strlen(name_text) + 2 + strlen(reason_text) + 1;
	char *detail = malloc(length);
	if (detail == NULL) {
		panic_text("uncaught Objective-C exception ", name_text);
	}
	strcpy(detail, name_text);
	strcat(detail, ": ");
	strcat(detail, reason_text);
	panic_text("uncaught Objective-C exception ", detail);
}

// An app spends its life in the main run loop, and the program's output is written when the buffer
// fills or the program exits, which an app may never do: so whatever it wrote is written each time
// the run loop is about to wait.
static void before_waiting(CFRunLoopObserverRef observer, CFRunLoopActivity activity, void *context) {
	(void)observer;
	(void)activity;
	(void)context;
	adamic_output_flush();
}

// Before main: the pool, the uncaught-exception hook, and the output's flush.
__attribute__((constructor)) static void adamic_apple_start(void) {
	pool = objc_autoreleasePoolPush();
	objc_setUncaughtExceptionHandler(uncaught);
	CFRunLoopObserverRef observer = CFRunLoopObserverCreate(NULL, kCFRunLoopBeforeWaiting, true, 0, before_waiting, NULL);
	CFRunLoopAddObserver(CFRunLoopGetMain(), observer, kCFRunLoopCommonModes);
	CFRelease(observer);
}
