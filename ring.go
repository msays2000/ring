package ring

// RingChannel uses two channels input and output in a way that never blocks the input.
// Specifically, if a value is written to a RingChannel's input when its buffer is full then the oldest
// value in the output channel is discarded to make room (just like a standard ring-buffer).

type RingChannel struct {
	input  chan interface{}
	output chan interface{}
	size   int
}

// Ring Channel helps us to keep the latest N samples in memory
// If new samples arrive, oldest ones are evicted out.
func NewRingChannel(size int) *RingChannel {
	if size <= 0 {
		panic("invalid negative or empty size in NewRingChannel")
	}

	ch := &RingChannel{
		input:  make(chan interface{}),
		output: make(chan interface{}, size),
		size:   size,
	}
	go ch.run()
	return ch
}

func (ch *RingChannel) run() {
	// range on input will get elements or block until input channel is closed
	for i := range ch.input {
		select {
		// output has some space
		case ch.output <- i:
		default:
			// output doesn't have space, discard oldest one to put one
			<-ch.output
			ch.output <- i
		}
	}
	// if we are here implies input channel is closed
	// so close the output channel as well.
	close(ch.output)
}

// In exposes the channel, where you can only send to.
func (ch *RingChannel) In() chan<- interface{} {
	return ch.input
}

// Out exposes the channel, where you can only receive from.
func (ch *RingChannel) Out() <-chan interface{} {
	return ch.output
}

// the number of elements queued (unread) in the output channel
func (ch *RingChannel) Len() int {
	return len(ch.output)
}

func (ch *RingChannel) Size() int {
	return ch.size
}

func (ch *RingChannel) Close() {
	// close input first
	close(ch.input)
}
