package yaml

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

// TestParserConcurrentParse 多个 agollo client 会并发解析各自的 namespace，
// 每次 Parse 的结果必须只来自本次 content，不能出现 key 在但值为 nil 或串到别的 namespace。
func TestParserConcurrentParse(t *testing.T) {
	const workers, rounds, keys = 8, 20, 100
	contents := make([]string, workers)
	for w := 0; w < workers; w++ {
		var b strings.Builder
		fmt.Fprintf(&b, "ns%d:\n", w)
		for k := 0; k < keys; k++ {
			fmt.Fprintf(&b, "  k%d: v%d-%d\n", k, w, k)
		}
		contents[w] = b.String()
	}

	p := &Parser{}
	var wg sync.WaitGroup
	errs := make(chan string, workers)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for r := 0; r < rounds; r++ {
				m, err := p.Parse(contents[w])
				if err != nil {
					errs <- err.Error()
					return
				}
				if len(m) != keys {
					errs <- fmt.Sprintf("worker %d: got %d keys, want %d", w, len(m), keys)
					return
				}
				for k := 0; k < keys; k++ {
					key, want := fmt.Sprintf("ns%d.k%d", w, k), fmt.Sprintf("v%d-%d", w, k)
					if got := m[key]; got != want {
						errs <- fmt.Sprintf("worker %d: %s = %v, want %s", w, key, got, want)
						return
					}
				}
			}
		}(w)
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}
