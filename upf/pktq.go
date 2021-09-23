package upf

type pkt struct {
	pktp uintptr
	qdat uint64
}

type pktq struct {
	pktq []pkt
	head int
	tail int
	size int
}

func newPktq(size int) pktq {
	pktq := pktq{}
	pktq.pktq = make([]pkt, size, size)
	pktq.head = 0
	pktq.tail = 0
	pktq.size = size
	return pktq
}

func (pktq *pktq) enq(pktp uintptr, qdat uint64) error {
	/*
	   if pktq.qlen() == pktq.size-1 {
	           return errors.New("pktq full")
	   }
	*/
	pktq.pktq[pktq.tail].pktp = pktp
	pktq.pktq[pktq.tail].qdat = qdat
	if pktq.tail == pktq.size-1 {
		pktq.tail = 0
	} else {
		pktq.tail += 1
	}
	return nil
}

func (pktq *pktq) deq() (uintptr, error) {
	/*
	   if pktq.qlen() == 0 {
	           return 0, errors.New("pktq empty")
	   }
	*/
	pktp := pktq.pktq[pktq.head].pktp
	if pktq.head == pktq.size-1 {
		pktq.head = 0
	} else {
		pktq.head += 1
	}
	return pktp, nil
}

func (pktq *pktq) qlen() int {
	if pktq.head <= pktq.tail {
		return pktq.tail - pktq.head
	}
	return pktq.size + pktq.tail - pktq.head
}

/*
func (pktq *pktq) qold() (uint64, error) {
	if pktq.qlen() == 0 {
		return 0, errors.New("pktq empty")
	}
	return pktq.pktq[pktq.head].qdat, nil
}
*/
func (pktq *pktq) qold() uint64 {
	return pktq.pktq[pktq.head].qdat
}
