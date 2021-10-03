package upf

import (
	"net"
)

type interfaceValue int

const (
	IV_ACCESS interfaceValue = iota
	IV_CORE
	IV_SGILAN_N6LAN
	IV_CP_FUNCTION
)

func (iv interfaceValue) String() string {
	switch iv {
	case IV_ACCESS:
		return "access"
	case IV_CORE:
		return "core"
	case IV_SGILAN_N6LAN:
		return "sgin6"
	case IV_CP_FUNCTION:
		return "cpf"
	}
	return "*unknown*"
}

type session struct {
	fseid    fseid
	n6Pdrs   []*pdr
	n3n9Pdrs []*pdr
	fars     []*far
	qers     []*qer
}

type fseid struct {
	seid    uint64
	address net.IP
}

type pdrColor int

const (
	PC_BLACK pdrColor = iota
	PC_RED
)

type pdr struct {
	pdrid              uint16
	precedence         uint32
	pdi                pdi
	outerHeaderRemoval bool
	far                *far
	qers               []*qer

	pktq pktq

	nextTx uint64

	parent *pdr
	left   *pdr
	right  *pdr
	color  pdrColor
}

/* XXX: */
func (pdr *pdr) isUl() bool {
	if pdr.far.destinationInterface == IV_CORE || pdr.far.destinationInterface == IV_SGILAN_N6LAN {
		return true
	}
	return false
}

/* XXX: */
func (pdr *pdr) isDl() bool {
	return !pdr.isUl()
}

type pdi struct {
	sourceInterface interfaceValue
	fteid           fteid
	ueIpAddress     net.IP
	sdfFilter       sdfFilter
}

type fteid struct {
	teid    uint32
	address net.IP
}

type sdfFilter struct {
	sourcePrefix      net.IPNet
	destinationPrefix net.IPNet
	protocol          uint16
	sourcePorts       []portRange
	destinationPorts  []portRange
}

type portRange struct {
	begin uint16
	end   uint16
}

type far struct {
	farid                uint32
	destinationInterface interfaceValue
	outerHeaderCreation  bool
	cork                 gtp5gCork
	pid                  uint16
}

type qer struct {
	qerid        uint32
	mbrUl        uint64
	mbrDl        uint64
	packetRateUl uint64
	packetRateDl uint64
	qfi          uint8

	ulPdrs []*pdr
	dlPdrs []*pdr

	nextUlTx   uint64
	nextDlTx   uint64
	ulBpsDelta uint64
	dlBpsDelta uint64
	ulPpsDelta uint64
	dlPpsDelta uint64

	internal bool
}

var sessions []*session

type n3n9SessionKey [8]byte
type n6SessionKey [4]byte

var n3n9SessionMap map[n3n9SessionKey]*session
var n6SessionMap map[n6SessionKey]*session

var queuedPdrs queuedPdrsType
