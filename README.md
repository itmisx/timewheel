# timewheel

> 时间轮，是一种实现延迟功能（定时器）的巧妙算法，在 Netty，Zookeeper，Kafka 等各种框架中，甚至 Linux 内核中都有用到。

#### 🎉 安装

`go get -u -v github.com/itmisx/timewheel`

#### ✅ 使用

```go
// 结构体方法调用
{
	// 参数1-time.Duration,时间轮精度
	// 参数2-int,时间槽数量
	// 参数3-func(interface{}),定时器过期回调函数
	tw := timewheel.New(time.Second, 60, callback)
	tw.Start()
	// timerID，定时器id，用来删除定时器
	// 参数1-string，定时器id，相同的定时器id会覆盖旧的定时器
	// 参数2-time.Duration，定时器间隔
	// 参数3-interface{}，定时器数据，将传递到回调函数
	tw.AddTimer("timerID", time.Second*3, "data")
	tw.Stop()
}
```

#### Benchmark

```
goos: darwin
goarch: amd64
pkg: 192.168.1.75/go-pkg/timewheel
cpu: Intel(R) Core(TM) i5-7360U CPU @ 2.30GHz
BenchmarkTimeWheel-4   	 1000000	      1550 ns/op
```
