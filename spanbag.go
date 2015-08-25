package monitor

// spanBag is a bag data structure (can add 0 or more references to a span,
// where every add needs to be matched with an equivalent remove). spanBag has
// a fast path for dealing with cases where the bag only has one element (the
// common case). spanBag is not threadsafe
type spanBag struct {
	first *Span
	rest  map[*Span]int32
}

func (b *spanBag) Add(s *Span) {
	if b.first == nil {
		b.first = s
		return
	}
	if b.rest == nil {
		b.rest = map[*Span]int32{}
	}
	b.rest[s] += 1
}

func (b *spanBag) Remove(s *Span) {
	if b.first == s {
		b.first = nil
		return
	}
	// okay it must be in b.rest
	count := b.rest[s]
	if count <= 1 {
		delete(b.rest, s)
	} else {
		b.rest[s] = count - 1
	}
}

// Iterate returns all elements
func (b *spanBag) Iterate(cb func(*Span)) {
	if b.first != nil {
		cb(b.first)
	}
	for s := range b.rest {
		cb(s)
	}
}
