package ttlcounter

// expirationItem represents an entry in the expiration priority queue.
// It tracks when a specific key should expire from the counter.
type expirationItem struct {
	key      string // The counter key this entry refers to
	expireAt int64  // Unix timestamp when this item should expire
	index    int    // Current position in the heap (maintained by heap.Interface)
}

// expirationQueue implements heap.Interface and manages items in expiration order.
// The queue is a min-heap where the item with earliest expiration is at index 0.
type expirationQueue []*expirationItem

// Len returns the number of items in the queue.
// This satisfies the heap.Interface requirement.
func (eq expirationQueue) Len() int { return len(eq) }

// Less compares two items by their expiration time.
// Returns true if item at i expires earlier than item at j.
// This satisfies the heap.Interface requirement and defines the heap ordering.
func (eq expirationQueue) Less(i, j int) bool {
	return eq[i].expireAt < eq[j].expireAt
}

// Swap exchanges the positions of two items in the queue.
// It updates their index fields to maintain consistency.
// This satisfies the heap.Interface requirement.
func (eq expirationQueue) Swap(i, j int) {
	eq[i], eq[j] = eq[j], eq[i]
	eq[i].index = i
	eq[j].index = j
}

// Push adds an item to the queue.
// The item must be of type *expirationItem.
// This satisfies the heap.Interface requirement.
// Note: This is used by heap.Push, not typically called directly.
func (eq *expirationQueue) Push(x interface{}) {
	n := len(*eq)
	item := x.(*expirationItem)
	item.index = n
	*eq = append(*eq, item)
}

// Pop removes and returns the item at the end of the queue.
// The heap package calls this after moving the earliest item to index 0.
// This satisfies the heap.Interface requirement.
// Note: This is used by heap.Pop, not typically called directly.
func (eq *expirationQueue) Pop() interface{} {
	old := *eq
	n := len(old)
	item := old[n-1]
	item.index = -1 // Mark as removed for safety
	*eq = old[0 : n-1]

	return item
}
