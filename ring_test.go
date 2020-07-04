package ring

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	cyclePerMinute        = 6
	minutesToCache        = 4
	reducedMinutesToCache = 2
	cycleCount            = minutesToCache * cyclePerMinute
)

func Test_RingEvictsOldValues(t *testing.T) {
	// deliberate keep look back cache to smaller than needed and ensure
	// all values are new and old ones are evicted.
	ch := NewRingChannel(reducedMinutesToCache * cyclePerMinute)
	assert.Equal(t, 0, ch.Len())

	for cycle := 1; cycle <= cycleCount; cycle++ {
		ch.In() <- cycle
		// let channel pass messages
		time.Sleep(1 * time.Millisecond)
		if cycle%cyclePerMinute == 0 {
			// 1 minute cycle over, lets consume it.
			consumer := consume(ch)
			// ensure we consumed exactly cyclePerMinute = 6
			assert.Equal(t, cyclePerMinute, len(consumer))
			// enusre the first value is fresh and newer or equal to cycle number.
			// this also means old values are evicted out.
			assert.GreaterOrEqual(t, cycle, consumer[0])
		}
	}
	ch.Close()
	consume(ch)
	assert.Equal(t, 0, ch.Len())
}

func Test_ReCycleRing(t *testing.T) {
	ch := NewRingChannel(cycleCount)
	for cycle := 1; cycle <= cycleCount; cycle++ {
		ch.In() <- cycle
		// let channel pass messages
		time.Sleep(1 * time.Millisecond)
		if cycle%cyclePerMinute == 0 {
			// 1 minute cycle over, lets consume it.
			post := consume(ch)
			// fake that metrics post failed, so put back cycles back into the ring.
			for _, putBack := range post {
				ch.In() <- putBack
			}
		}
	}
	assert.Equal(t, cycleCount, ch.Len())
	ch.Close()
	consume(ch)
	assert.Equal(t, 0, ch.Len())
}

func consume(ch *RingChannel) []int {
	var consumer []int
	for i := 0; i <= cycleCount; i++ {
		select {
		case x := <-ch.Out():
			consumer = append(consumer, x.(int))
		default:
			// nothing available
			break
		}
	}
	fmt.Printf("%v\n", consumer)
	return consumer
}
