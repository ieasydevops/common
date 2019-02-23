package climit

import (
	"testing"
	"github.com/go/src/fmt"
	"sync"
	"math/rand"
	"time"
	"github.com/meixinyun/common/pkg/util/sets"
	"strings"
	"github.com/go/src/strconv"
)

func Test_ConcurrencyLimit(t *testing.T) {
	counter := NewConcurrencyLimiter(10000)
	var wg sync.WaitGroup
	msgProcessBatch := func(msg string) {
		wg.Add(1)
		counter.Increment()
		go func(msg string) {
			fmt.Println(msg)
			rand.Seed(time.Now().UnixNano())
			x := rand.Intn(5)
			time.Sleep(time.Duration(x) * time.Second)
			counter.Decrement()
			wg.Done()
		}(msg)
	}

	set := sets.NewString()
	for i := 0; i < 1000; i++ {
		set.Insert(strings.Join([]string{"a", "_",}, strconv.Itoa(i)))
	}

	for s := range set {
		msgProcessBatch(s)
	}
	wg.Wait()
}
