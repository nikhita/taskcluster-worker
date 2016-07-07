package got

import (
	"testing"
	"time"
)

func TestDefaultBackOff(t *testing.T) {
	b := DefaultBackOff
	// Run a few iterations for good measure
	for i := 0; i < 10000; i++ {
		d := b.Delay(0)
		assert(d == 0)

		d = b.Delay(0)
		assert(d == 0)

		d = b.Delay(1)
		assert(75*time.Millisecond <= d && d <= 125*time.Millisecond,
			"delay(1) is outside range, got: ", d)

		d = b.Delay(2)
		assert(150*time.Millisecond <= d && d <= 250*time.Millisecond,
			"delay(2) is outside range, got: ", d)

		d = b.Delay(3)
		assert(300*time.Millisecond <= d && d <= 500*time.Millisecond,
			"delay(3) is outside range, got: ", d)

		d = b.Delay(4)
		assert(600*time.Millisecond <= d && d <= 1000*time.Millisecond,
			"delay(4) is outside range, got: ", d)

		d = b.Delay(5)
		assert(1200*time.Millisecond <= d && d <= 2000*time.Millisecond,
			"delay(5) is outside range, got: ", d)

		d = b.Delay(6)
		assert(2400*time.Millisecond <= d && d <= 4000*time.Millisecond,
			"delay(6) is outside range, got: ", d)

		d = b.Delay(7)
		assert(4800*time.Millisecond <= d && d <= 8000*time.Millisecond,
			"delay(7) is outside range, got: ", d)

		d = b.Delay(8)
		assert(9600*time.Millisecond <= d && d <= 16000*time.Millisecond,
			"delay(8) is outside range, got: ", d)

		d = b.Delay(9)
		assert(19200*time.Millisecond <= d && d <= 30000*time.Millisecond,
			"delay(9) is outside range, got: ", d)

		d = b.Delay(10)
		assert(30000*time.Millisecond <= d && d <= 30000*time.Millisecond,
			"delay(10) is outside range, got: ", d)
	}
}

func TestBackOffWithoutRandomization(t *testing.T) {
	b := BackOff{
		DelayFactor:         100 * time.Millisecond,
		RandomizationFactor: 0,
		MaxDelay:            30 * time.Second,
	}
	// Run a few iterations for good measure
	for i := 0; i < 10000; i++ {
		d := b.Delay(0)
		assert(d == 0)

		d = b.Delay(0)
		assert(d == 0)

		d = b.Delay(1)
		assert(d == 100*time.Millisecond,
			"delay(1) is outside range, got: ", d)

		d = b.Delay(2)
		assert(d == 200*time.Millisecond,
			"delay(2) is outside range, got: ", d)

		d = b.Delay(3)
		assert(d == 400*time.Millisecond,
			"delay(3) is outside range, got: ", d)

		d = b.Delay(4)
		assert(d == 800*time.Millisecond,
			"delay(4) is outside range, got: ", d)

		d = b.Delay(5)
		assert(d == 1600*time.Millisecond,
			"delay(5) is outside range, got: ", d)

		d = b.Delay(6)
		assert(d == 3200*time.Millisecond,
			"delay(6) is outside range, got: ", d)

		d = b.Delay(7)
		assert(d == 6400*time.Millisecond,
			"delay(7) is outside range, got: ", d)

		d = b.Delay(8)
		assert(d == 12800*time.Millisecond,
			"delay(8) is outside range, got: ", d)

		d = b.Delay(9)
		assert(d == 25600*time.Millisecond,
			"delay(9) is outside range, got: ", d)

		d = b.Delay(10)
		assert(d == 30000*time.Millisecond,
			"delay(10) is outside range, got: ", d)
	}
}
