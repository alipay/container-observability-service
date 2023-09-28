package utils

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSafeMap(t *testing.T) {
	smap := NewSafeMap()

	wg := sync.WaitGroup{}
	wg.Add(5)

	go func() {
		for i := 0; i < 100000; i++ {
			smap.Set(fmt.Sprintf("%d", i), i)
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < 100000; i++ {
			smap.Set(fmt.Sprintf("%d", i), i)
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < 100000; i++ {
			smap.Set(fmt.Sprintf("%d", i), i)
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < 100000; i++ {
			s := fmt.Sprintf("%d", i)
			if val, ok := smap.Get(s); ok {
				assert.Equal(t, val, i)
				smap.Delete(s)
			}
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < 100000; i++ {
			s := fmt.Sprintf("%d", i)
			if val, ok := smap.Get(s); ok {
				assert.Equal(t, val, i)
				smap.Delete(s)
			}
		}
		wg.Done()
	}()

	wg.Wait()
}
