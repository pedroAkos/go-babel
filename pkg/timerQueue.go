package pkg

import (
	"container/heap"
	"math"
	"math/rand"
	"reflect"
	"time"

	"github.com/nm-morais/go-babel/pkg/dataStructures"
	"github.com/nm-morais/go-babel/pkg/errors"
	"github.com/nm-morais/go-babel/pkg/logs"
	"github.com/nm-morais/go-babel/pkg/protocol"
	"github.com/nm-morais/go-babel/pkg/timer"
	"github.com/sirupsen/logrus"
)

const timerQueueCaller = "timerQueue"

type cancelTimerReq struct {
	key     int
	removed chan int
}

type pqItemValue struct {
	protoID protocol.ID
	timer   timer.Timer
}

type TimerQueue interface {
	AddTimer(timer timer.Timer, protocolId protocol.ID) int
	CancelTimer(int) errors.Error
	Logger() *logrus.Logger
}

type timerQueueImpl struct {
	pq              dataStructures.PriorityQueue
	addTimerChan    chan *dataStructures.Item
	cancelTimerChan chan *cancelTimerReq
	logger          *logrus.Logger
}

func NewTimerQueue() TimerQueue {
	tq := &timerQueueImpl{
		pq:              make(dataStructures.PriorityQueue, 0),
		addTimerChan:    make(chan *dataStructures.Item),
		cancelTimerChan: make(chan *cancelTimerReq),
		logger:          logs.NewLogger(timerQueueCaller),
	}
	go tq.start()
	return tq
}

func (tq *timerQueueImpl) AddTimer(timer timer.Timer, protocolId protocol.ID) int {
	pqItem := &dataStructures.Item{
		Value: &pqItemValue{
			protoID: protocolId,
			timer:   timer,
		},
		Priority: timer.Deadline().UnixNano(),
		Key:      rand.Int(),
	}
	tq.addTimerChan <- pqItem
	return pqItem.Key
}

func (tq *timerQueueImpl) removeItem(timerID int) int {
	tq.logger.Infof("Canceling timer with ID %d", timerID)
	removed := tq.pq.Remove(timerID)
	if removed == nil {
		return -1
	}
	return removed.Key
}

func (tq *timerQueueImpl) CancelTimer(timerID int) errors.Error {
	responseChan := make(chan int)
	defer close(responseChan)
	tq.cancelTimerChan <- &cancelTimerReq{key: timerID, removed: responseChan}
	response := <-responseChan
	if response == -1 {
		return errors.NonFatalError(404, "timer not found", timerQueueCaller)
	}
	return nil
}

func (tq *timerQueueImpl) Logger() *logrus.Logger {
	return tq.logger
}

func (tq *timerQueueImpl) start() {

	for {
		var nextItem *dataStructures.Item
		var waitTime time.Duration
		var currTimer *time.Timer

		if tq.pq.Len() > 0 {
			// tq.pq.LogEntries(tq.logger)
			nextItem = heap.Pop(&tq.pq).(*dataStructures.Item)
			value := nextItem.Value.(*pqItemValue)
			waitTime = time.Until(value.timer.Deadline())
			currTimer = time.NewTimer(waitTime)
			tq.logger.Infof("Waiting %s for timer of type %s with id %d", waitTime, reflect.TypeOf(value.timer), nextItem.Key)
		} else {
			currTimer = time.NewTimer(math.MaxInt64)
		}

		select {
		case req := <-tq.cancelTimerChan:
			tq.logger.Infof("Received cancel timer signal...")
			req.removed <- tq.removeItem(req.key)
			if req.key == nextItem.Key {
				currTimer.Stop()
				tq.logger.Infof("Canceled timer")
			}
			tq.logger.Infof("Removed timer %d", req.key)
		case newItem := <-tq.addTimerChan:
			tq.logger.Infof("Received add timer signal...")
			if nextItem != nil {
				heap.Push(&tq.pq, nextItem)
			}
			heap.Push(&tq.pq, newItem)
		case <-currTimer.C:
			tq.logger.Info()
			tq.logger.Infof("----------------------Processing %+v------------------", *nextItem)
			value := nextItem.Value.(*pqItemValue)
			if proto, ok := p.protocols.Load(value.protoID); ok {
				proto.(protocolValueType).DeliverTimer(value.timer)
			}
		}
	}
}
