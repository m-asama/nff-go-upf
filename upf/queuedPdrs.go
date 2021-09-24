package upf

import (
	"sort"
)

type queuedPdrsType struct {
	pdrsSlice     []*pdr
	pdrsSliceData []*pdr
}

func (queuedPdrs *queuedPdrsType) init() {
	queuedPdrs.pdrsSliceData = make([]*pdr, 8)
	queuedPdrs.pdrsSlice = queuedPdrs.pdrsSliceData[0:0]
}

func (queuedPdrs *queuedPdrsType) head() *pdr {
	if len(queuedPdrs.pdrsSlice) == 0 {
		return nil
	}
	return queuedPdrs.pdrsSlice[0]
}

func (queuedPdrs *queuedPdrsType) enq(p *pdr) {
	if len(queuedPdrs.pdrsSlice) == len(queuedPdrs.pdrsSliceData) {
		queuedPdrs.pdrsSliceData = append(queuedPdrs.pdrsSliceData, make([]*pdr, 8)...)
	}
	queuedPdrs.pdrsSlice = queuedPdrs.pdrsSliceData[0 : len(queuedPdrs.pdrsSlice)+1]
	queuedPdrs.pdrsSlice[len(queuedPdrs.pdrsSlice)-1] = p
	queuedPdrs.sort()
}

func (queuedPdrs *queuedPdrsType) deq() *pdr {
	if len(queuedPdrs.pdrsSlice) == 0 {
		return nil
	}
	p := queuedPdrs.pdrsSlice[0]
	queuedPdrs.pdrsSlice[0] = queuedPdrs.pdrsSlice[len(queuedPdrs.pdrsSlice)-1]
	queuedPdrs.pdrsSlice = queuedPdrs.pdrsSliceData[0 : len(queuedPdrs.pdrsSlice)-1]
	queuedPdrs.sort()
	return p
}

func (queuedPdrs *queuedPdrsType) exists(pdr *pdr) bool {
	for _, p := range queuedPdrs.pdrsSlice {
		if p == pdr {
			return true
		}
	}
	return false
}

func (queuedPdrs *queuedPdrsType) qlen() int {
	return len(queuedPdrs.pdrsSlice)
}

func (queuedPdrs *queuedPdrsType) sort() {
	sort.Slice(queuedPdrs.pdrsSlice, func(a, b int) bool {
		amin := queuedPdrs.pdrsSlice[a].pktq.qold()
		bmin := queuedPdrs.pdrsSlice[b].pktq.qold()
		return int64(amin-bmin) < 0
	})
}
