package timewheel

import (
	"container/list"
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	DefaultSlotInterval = time.Second * 1
	DefaultSlotNum      = 60
)

type timer struct {
	timerID  string
	interval time.Duration
	data     interface{}
}
type timerRecord struct {
	slot  int
	wheel int
	el    *list.Element
}

var tw *TimeWheel

type TimeWheel struct {
	mu           sync.RWMutex
	ticker       *time.Ticker
	slotInterval time.Duration
	slotNum      int
	// map[wheel][slot]timerItemList
	wheelTimerList map[int]map[int]*list.List
	currentPos     int
	currentWheel   int
	callback       func(interface{})
	startChan      chan bool
	stopChan       chan bool
	timerRecordMap map[string]timerRecord
}

// New 实例化时间轮
func New(slotDuration time.Duration, slotNum int, callback func(interface{})) *TimeWheel {
	if slotDuration <= 0 || slotNum <= 0 {
		slotDuration = time.Second
		slotNum = 60
	}
	return &TimeWheel{
		slotInterval: slotDuration,
		slotNum:      60,
		// map[wheel][slot]timerItemList
		wheelTimerList: make(map[int]map[int]*list.List),
		callback:       callback,
		startChan:      make(chan bool, 1),
		stopChan:       make(chan bool, 1),
		timerRecordMap: make(map[string]timerRecord),
	}
}

// Start start timewheel
func (tw *TimeWheel) Start() {
	// 避免重复操作
	select {
	case tw.startChan <- true:
	default:
		fmt.Println("timewheel is already running, exit")
		return
	}

	// 启动时间轮的内部定时器
	tw.ticker = time.NewTicker(tw.slotInterval)
	go func() {
		for {
			select {
			case <-tw.ticker.C:
				tw.tickerHandler()
			case <-tw.stopChan:
				return
			}
		}
	}()
}

// AddTimer add timer
func (tw *TimeWheel) AddTimer(timerID string, interval time.Duration, data interface{}) (success bool) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if len(tw.startChan) == 0 {
		log.Println(" timewheel has not been started")
		return false
	}

	tm := timer{
		timerID:  timerID,
		interval: interval,
		data:     data,
	}
	// 已经存在的定时器id，将会停止旧的定时器，并创建新的定时器
	if _, ok := tw.timerRecordMap[timerID]; ok {
		tw.StopTimer(timerID)
	}
	// 计算定时器要插入的位置
	calSlotPos := (tw.currentPos + int(tm.interval)/int(tw.slotInterval)) % tw.slotNum // calculate the timer's wheel
	calWheel := (tw.currentPos+int(tm.interval)/int(tw.slotInterval))/tw.slotNum + tw.currentWheel
	if tw.wheelTimerList[calWheel] == nil {
		tw.wheelTimerList[calWheel] = make(map[int]*list.List)
	}
	// 获取定时器链表
	l := tw.wheelTimerList[calWheel][calSlotPos]
	if l == nil {
		l = list.New()
	}
	// 定时器插入链表
	el := l.PushBack(tm)
	tw.wheelTimerList[calWheel][calSlotPos] = l
	// 保存定时器信息，id及位置
	tw.timerRecordMap[tm.timerID] = timerRecord{
		slot:  calSlotPos,
		wheel: calWheel,
		el:    el,
	}
	return true
}

// StopTimer stop timer
func (tw *TimeWheel) StopTimer(timerID string) (success bool) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	// 定时器id存在检查
	if _, ok := tw.timerRecordMap[timerID]; !ok {
		log.Println("timer does not exist")
		return false
	}
	l := tw.wheelTimerList[tw.timerRecordMap[timerID].wheel][tw.timerRecordMap[timerID].slot]
	el := tw.timerRecordMap[timerID].el
	l.Remove(el)
	tw.wheelTimerList[tw.timerRecordMap[timerID].wheel][tw.timerRecordMap[timerID].slot] = l
	return true
}

// Stop stop the timewheel
func (tw *TimeWheel) Stop() (success bool) {
	// 避免重复操作
	select {
	case <-tw.startChan:
		tw.ticker.Stop()
		tw.stopChan <- true
		return true
	default:
		log.Println("timewheel has stopped,exit.")
		return false
	}
}

// tickerHandler
func (tw *TimeWheel) tickerHandler() {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	// 时间轮转动
	tw.currentPos++
	// 时间轮执行一圈后，时间轮递增，时间槽计数复位
	if tw.currentPos%tw.slotNum == 0 {
		delete(tw.wheelTimerList, tw.currentWheel)
		tw.currentWheel++
		tw.currentPos = 0
	}
	// 跳过空的时间槽
	l := tw.wheelTimerList[tw.currentWheel][tw.currentPos]
	if l == nil {
		return
	}
	// 遍历时间时间槽的定时器链表
	for e := l.Front(); e != nil; e = e.Next() {
		tm, _ := e.Value.(timer)
		delete(tw.timerRecordMap, tm.timerID)
		go tw.callback(tm.data)
	}
	// 删除时间槽的定时器链表
	delete(tw.wheelTimerList[tw.currentWheel], tw.currentPos)
}

func Start(callback func(interface{})) {
	tw = New(DefaultSlotInterval, DefaultSlotNum, callback)
	tw.Start()
}

func AddTimer(key string, interval time.Duration, data interface{}) {
	tw.AddTimer(key, interval, data)
}

func StopTimer(key string) {
	tw.StopTimer(key)
}

func Stop() {
	tw.Stop()
}
