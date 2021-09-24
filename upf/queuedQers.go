package upf

import (
	//"errors"
	"sort"
)

type queuedQersType struct {
	qersSlice     []*qer
	qersSliceData []*qer
}

func (queuedQers *queuedQersType) init() {
	queuedQers.qersSliceData = make([]*qer, 1024)
	queuedQers.qersSlice = queuedQers.qersSliceData[0:0]
}

func (queuedQers *queuedQersType) head() *qer {
	if len(queuedQers.qersSlice) == 0 {
		return nil
	}
	return queuedQers.qersSlice[0]
}

func (queuedQers *queuedQersType) enq(q *qer) {
	if len(queuedQers.qersSlice) == len(queuedQers.qersSliceData) {
		queuedQers.qersSliceData = append(queuedQers.qersSliceData, make([]*qer, 1024)...)
	}
	queuedQers.qersSlice = queuedQers.qersSliceData[0 : len(queuedQers.qersSlice)+1]
	queuedQers.qersSlice[len(queuedQers.qersSlice)-1] = q
	queuedQers.sort()
}

func (queuedQers *queuedQersType) deq() *qer {
	if len(queuedQers.qersSlice) == 0 {
		return nil
	}
	q := queuedQers.qersSlice[0]
	queuedQers.qersSlice[0] = queuedQers.qersSlice[len(queuedQers.qersSlice)-1]
	queuedQers.qersSlice = queuedQers.qersSliceData[0 : len(queuedQers.qersSlice)-1]
	queuedQers.sort()
	return q
}

func (queuedQers *queuedQersType) exists(qer *qer) bool {
	for _, q := range queuedQers.qersSlice {
		if q == qer {
			return true
		}
	}
	return false
}

func (queuedQers *queuedQersType) qlen() int {
	return len(queuedQers.qersSlice)
}

func (queuedQers *queuedQersType) sort() {
	sort.Slice(queuedQers.qersSlice, func(a, b int) bool {
		aNext := queuedQers.qersSlice[a].nextTx()
		bNext := queuedQers.qersSlice[b].nextTx()
		return int64(aNext-bNext) < 0
	})
}
