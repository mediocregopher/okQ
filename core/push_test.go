package core

import (
	. "testing"
	"time"

	"github.com/levenlabs/golib/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyWait(t *T) {
	// Make sure notifying on Keys which don't have waiters is fine
	base := testutil.RandStr()
	k1 := randKey(base)
	k2 := randKey(base)
	testCore.KeyNotify(k1)
	testCore.KeyNotify(k2)

	// This test shouldn't take too long
	go func() {
		time.Sleep(5 * time.Second)
		panic("test took too long")
	}()

	ch1 := make(chan bool, 1)
	go func() {
		<-testCore.KeyWait(k1, nil)
		ch1 <- true
	}()

	ch2 := make(chan bool, 1)
	go func() {
		<-testCore.KeyWait(k1, nil)
		ch2 <- true
	}()

	ch3 := make(chan bool, 1)
	go func() {
		<-testCore.KeyWait(k2, nil)
		ch3 <- true
	}()

	ch4 := make(chan bool, 1)
	ch4stop := make(chan struct{})
	go func() {
		<-testCore.KeyWait(randKey(base), ch4stop)
		ch4 <- true
	}()

	time.Sleep(100 * time.Millisecond)

	assertBlocking := func(ch chan bool) {
		select {
		case <-ch:
			assert.Fail(t, "channel should be blocking")
		default:
		}
	}

	assertNotBlocking := func(ch chan bool) {
		select {
		case b := <-ch:
			assert.True(t, b)
		case <-time.After(100 * time.Millisecond):
			assert.Fail(t, "channel should not be blocking")
		}
	}

	assertBlocking(ch1)
	assertBlocking(ch2)
	assertBlocking(ch3)
	assertBlocking(ch4)

	testCore.KeyNotify(k1)
	assertNotBlocking(ch1)
	assertNotBlocking(ch2)
	assertBlocking(ch3)
	assertBlocking(ch4)

	testCore.KeyNotify(k2)
	assertNotBlocking(ch3)
	assertBlocking(ch4)

	close(ch4stop)
	assertNotBlocking(ch4)
}

func TestStoreWaiters(t *T) {
	id1 := testutil.RandStr()
	id2 := testutil.RandStr()
	es := randKey(testutil.RandStr())

	now := NewTS(time.Now())
	assertCount := func(c int) {
		cc, err := testCore.KeyCountWaiters(es, now)
		require.Nil(t, err)
		assert.Equal(t, c, cc)
	}
	assertCount(0)

	now = NewTS(time.Now())
	err := testCore.KeyStoreWaiter(es, id1, now, NewTS(time.Now().Add(1*time.Second)))
	require.Nil(t, err)
	err = testCore.KeyStoreWaiter(es, id2, now, NewTS(time.Now().Add(5*time.Second)))
	require.Nil(t, err)
	assertCount(2)

	now = NewTS(time.Now().Add(2 * time.Second))
	assertCount(1)

	now = NewTS(time.Now().Add(10 * time.Second))
	assertCount(0)
}
