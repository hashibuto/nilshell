package ns

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFileLockSeq(t *testing.T) {
	l := NewFileLock("/tmp/test.lock")
	err := l.Lock()
	assert.NoError(t, err)
	err = l.Unlock()
	assert.NoError(t, err)

	err = l.Lock()
	assert.NoError(t, err)
	err = l.Unlock()
	assert.NoError(t, err)
}

func TestFileLockConcur(t *testing.T) {
	l := NewFileLock("/tmp/test.lock")
	err := l.Lock()
	assert.NoError(t, err)
	go func() {
		time.Sleep(1 * time.Second)
		err = l.Unlock()
		assert.NoError(t, err)
	}()

	l2 := NewFileLock("/tmp/test.lock")
	err = l2.Lock()
	assert.NoError(t, err)
	err = l2.Unlock()
	assert.NoError(t, err)
}
