package timewheel

import (
	"log"
	"strconv"
	"testing"
	"time"
)

func callback(data interface{}) {
	log.Println(data)
}

// 结构体方法调用
func TestTimeWheelNew(t *testing.T) {
	tw := New(time.Second, 60, callback)
	tw.Start()
	tw.AddTimer("a", time.Second*3, 1)
	tw.AddTimer("b", time.Second*4, 1)
	tw.AddTimer("c", time.Second*7, 1)
	time.Sleep(time.Second * 10)
	tw.Stop()
}

func BenchmarkTimeWheel(b *testing.B) {
	tw := New(time.Second, 3, callback)
	tw.Start()
	for n := 0; n < b.N; n++ {
		tw.AddTimer(strconv.Itoa(n), time.Second*2, 1)
	}
	tw.Stop()
}
