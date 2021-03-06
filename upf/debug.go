package upf

import (
	"fmt"
	//"time"
)

func debugDump1(pdr *pdr, i *int) {
	if pdr == nil {
		return
	}
	if pdr.left != sentinelPdr {
		debugDump1(pdr.left, i)
	}
	if pdr != sentinelPdr {
		qold := pdr.pktq.qold()
		if pdr.isUl() {
			fmt.Print("\t[UL] ")
		} else {
			fmt.Print("\t[DL] ")
		}
		fmt.Printf("me: %p parent: %p left: %p right: %p ", pdr, pdr.parent, pdr.left, pdr.right)
		switch pdr.color {
		case PC_BLACK:
			fmt.Print("*B* ")
		case PC_RED:
			fmt.Print("*R* ")
		default:
			fmt.Print("*?* ")
		}
		fmt.Println(fmt.Sprintf("%4d", *i), "pdrid:", pdr.pdrid, "nextTx:", pdr.nextTx, "qold:", qold, "pktq.qlen:", pdr.pktq.qlen())
		*i++
	}
	if pdr.right != sentinelPdr {
		debugDump1(pdr.right, i)
	}
}

func debugDump() {
	fmt.Printf("\tqueuedPdrs.root: %p queuedPdrs.qlen: %d sentinelPdr: %p\n", queuedPdrs.root, queuedPdrs.qlen, sentinelPdr)
	i := 0
	debugDump1(queuedPdrs.root, &i)
}

func debugPdr(i int, pdr *pdr) {
	if pdr == nil {
		fmt.Println("	pdr == nil")
		return
	}
	fmt.Printf("	pdr[%d]:\n", i)
	fmt.Println("		pdrid					", pdr.pdrid)
	fmt.Println("		precedence				", pdr.precedence)
	fmt.Println("		pdi:")
	fmt.Println("			sourceInterface			", pdr.pdi.sourceInterface)
	fmt.Println("			fteid				", pdr.pdi.fteid.teid, ":", pdr.pdi.fteid.address)
	fmt.Println("			ueIpAddress			", pdr.pdi.ueIpAddress)
	fmt.Println("			sdfFilter:")
	fmt.Println("				sourcePrefix		", pdr.pdi.sdfFilter.sourcePrefix)
	fmt.Println("				destinationPrefix	", pdr.pdi.sdfFilter.destinationPrefix)
	fmt.Println("				protocol		", pdr.pdi.sdfFilter.protocol)
	fmt.Print("				sourcePorts		 ")
	for _, portRange := range pdr.pdi.sdfFilter.sourcePorts {
		if portRange.begin == portRange.end {
			fmt.Printf("%d,", portRange.begin)
		} else {
			fmt.Printf("%d-%d,", portRange.begin, portRange.end)
		}
	}
	fmt.Println("")
	fmt.Print("				destinationPorts	 ")
	for _, portRange := range pdr.pdi.sdfFilter.destinationPorts {
		if portRange.begin == portRange.end {
			fmt.Printf("%d,", portRange.begin)
		} else {
			fmt.Printf("%d-%d,", portRange.begin, portRange.end)
		}
	}
	fmt.Println("")
	fmt.Println("		outerHeaderRemoval			", pdr.outerHeaderRemoval)
	fmt.Println("		far					", pdr.far.farid)
	fmt.Print("		qers					 ")
	for _, qer := range pdr.qers {
		fmt.Printf("%d,", qer.qerid)
	}
	fmt.Println("")
}

func debugFar(i int, far *far) {
	if far == nil {
		fmt.Println("	far == nil")
		return
	}
	fmt.Printf("	far[%d]:\n", i)
	fmt.Println("		farid					", far.farid)
	fmt.Println("		destinationInterface			", far.destinationInterface)
	fmt.Println("		outerHeaderCreation			", far.outerHeaderCreation)
}

func debugQer(i int, qer *qer) {
	if qer == nil {
		fmt.Println("	qer == nil")
		return
	}
	fmt.Printf("	qer[%d]:\n", i)
	fmt.Println("		qerid					", qer.qerid)
	fmt.Println("		mbrUl					", qer.mbrUl)
	fmt.Println("		mbrDl					", qer.mbrDl)
	fmt.Println("		packetRateUl				", qer.packetRateUl)
	fmt.Println("		packetRateDl				", qer.packetRateDl)
	fmt.Println("		qfi					", qer.qfi)
}

func debugSession(i int, session *session) {
	if session == nil {
		fmt.Println("session == nil")
		return
	}
	fmt.Printf("session[%d]:\n", i)
	fmt.Println("	fseid						", session.fseid.seid, ":", session.fseid.address)
	for i, pdr := range session.n6Pdrs {
		debugPdr(i, pdr)
	}
	for i, pdr := range session.n3n9Pdrs {
		debugPdr(i, pdr)
	}
	for i, far := range session.fars {
		debugFar(i, far)
	}
	for i, qer := range session.qers {
		debugQer(i, qer)
	}
}

func Debug() {
	if sessions == nil || len(sessions) == 0 {
		fmt.Println("sessions == nil || len(sessions) == 0")
		return
	}
	for i, session := range sessions {
		debugSession(i, session)
	}
	fmt.Println("n6SessionMap:")
	for key, session := range n6SessionMap {
		fmt.Println("	", key, ":", session.fseid.seid, ":", session.fseid.address)
	}
	fmt.Println("n3n9SessionMap:")
	for key, session := range n3n9SessionMap {
		fmt.Println("	", key, ":", session.fseid.seid, ":", session.fseid.address)
	}
}
