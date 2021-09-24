package upf

import (
	"sort"
	"unsafe"
)

type queuedPdrsType struct {
	pdrsSlice     []*pdr
	pdrsSliceData []*pdr
}

func (queuedPdrs *queuedPdrsType) init() {
	queuedPdrs.pdrsSliceData = make([]*pdr, 1024)
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
	//queuedPdrs.sort()
}

func (queuedPdrs *queuedPdrsType) deq() *pdr {
	if len(queuedPdrs.pdrsSlice) == 0 {
		return nil
	}
	p := queuedPdrs.pdrsSlice[0]
	queuedPdrs.pdrsSlice[0] = queuedPdrs.pdrsSlice[len(queuedPdrs.pdrsSlice)-1]
	queuedPdrs.pdrsSlice = queuedPdrs.pdrsSliceData[0 : len(queuedPdrs.pdrsSlice)-1]
	//queuedPdrs.sort()
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
		/*
			amin := queuedPdrs.pdrsSlice[a].pktq.qold()
			bmin := queuedPdrs.pdrsSlice[b].pktq.qold()
			return int64(amin-bmin) < 0
		*/
		aNext := queuedPdrs.pdrsSlice[a].nextTx
		bNext := queuedPdrs.pdrsSlice[b].nextTx
		if aNext != bNext {
			return int64(aNext-bNext) < 0
		}
		aQold := queuedPdrs.pdrsSlice[a].pktq.qold()
		bQold := queuedPdrs.pdrsSlice[b].pktq.qold()
		if aQold != bQold {
			return int64(aQold-bQold) < 0
		}
		aUintptr := uintptr(unsafe.Pointer(queuedPdrs.pdrsSlice[a]))
		bUintptr := uintptr(unsafe.Pointer(queuedPdrs.pdrsSlice[b]))
		return aUintptr < bUintptr
	})
}
