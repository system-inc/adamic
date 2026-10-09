// Private to heap.c. Chunk addresses never move, including while another thread grows the table.
// Chunk number + 1 stays below the opt-in Program flag; bit 29 is reserved otherwise.
#define PAGE_SIZE 1024
#ifdef ADAMIC_PROGRAM_REGION
#define PAGE_COUNT 262144
#define ADAMIC_FIRST_FLAG ADAMIC_PROGRAM_FLAG
#else
#define PAGE_COUNT 524288
#define ADAMIC_FIRST_FLAG UINT32_C(0x20000000)
#endif
_Static_assert((uint64_t)PAGE_SIZE * PAGE_COUNT <= ADAMIC_FIRST_FLAG,
    "chunk numbers would reach the slab flags");

typedef struct remote_slot {
	struct remote_slot *next;
	void *slot;
} remote_slot;
typedef struct chunk_page { _Atomic(chunk *) slots[PAGE_SIZE]; } chunk_page;
static _Atomic(chunk_page *) chunk_pages[PAGE_COUNT];
static _Atomic size_t chunk_count;
static _Atomic size_t thread_count;
static _Thread_local size_t owner_number;

static size_t thread_number(void) {
	if (owner_number == 0) {
		owner_number = atomic_fetch_add_explicit(&thread_count, 1, memory_order_relaxed) + 1;
	}
	return owner_number;
}

static void register_chunk(chunk *each) {
	size_t number = atomic_fetch_add_explicit(&chunk_count, 1, memory_order_relaxed);
	if (number >= (size_t)ADAMIC_FIRST_FLAG - 1) { adamic_panic("too many heap chunks", sizeof "too many heap chunks" - 1); }
	chunk_page *page = atomic_load_explicit(&chunk_pages[number / PAGE_SIZE], memory_order_acquire);
	if (page == NULL) {
		chunk_page *fresh = calloc(1, sizeof *fresh);
		if (fresh == NULL) { adamic_panic("out of memory", sizeof "out of memory" - 1); }
		for (size_t slot = 0; slot < PAGE_SIZE; slot++) { atomic_init(&fresh->slots[slot], NULL); }
		chunk_page *empty = NULL;
		if (!atomic_compare_exchange_strong_explicit(&chunk_pages[number / PAGE_SIZE], &empty, fresh, memory_order_release, memory_order_acquire)) {
			free(fresh);
			page = empty;
		} else { page = fresh; }
	}
	each->number = (uint32_t)number;
	atomic_store_explicit(&page->slots[number % PAGE_SIZE], each, memory_order_release);
}

static chunk *find_chunk(uint32_t number) {
	chunk_page *page = atomic_load_explicit(&chunk_pages[number / PAGE_SIZE], memory_order_acquire);
	return atomic_load_explicit(&page->slots[number % PAGE_SIZE], memory_order_acquire);
}
