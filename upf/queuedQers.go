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
		aQer := queuedQers.qersSlice[a]
		aPdr := aQer.queuedPdrs.head()
		/*
			if aPdr == nil {
				panic(errors.New("queuedPdrs empty"))
			}
		*/
		var aNext uint64
		if aPdr.isUl() {
			aNext = aQer.nextUlTx
		} else {
			aNext = aQer.nextDlTx
		}
		bQer := queuedQers.qersSlice[b]
		bPdr := bQer.queuedPdrs.head()
		/*
			if bPdr == nil {
				panic(errors.New("queuedPdrs empty"))
			}
		*/
		var bNext uint64
		if bPdr.isUl() {
			bNext = bQer.nextUlTx
		} else {
			bNext = bQer.nextDlTx
		}
		return int64(aNext-bNext) < 0
	})
}
