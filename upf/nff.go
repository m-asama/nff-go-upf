package upf

import (
	//"fmt"
	"net"
	"sync"
	//"time"
	"encoding/binary"
	"unsafe"

	"github.com/intel-go/nff-go/flow"
	"github.com/intel-go/nff-go/packet"
	"github.com/intel-go/nff-go/types"
)

var cpuList string
var port uint16
var speed uint64

var n6VlanId uint16
var n3n9VlanId uint16

var n6DstMacAddr [types.EtherAddrLen]uint8
var n3n9DstMacAddr [types.EtherAddrLen]uint8

var localMacAddr [types.EtherAddrLen]uint8
var teAddress net.IP

var tscsec uint64

var lock sync.RWMutex

type gtp5gHdr struct {
	headerType            uint8
	messageType           uint8
	messageLength         uint16
	teid                  uint32
	sequenceNumber        uint16
	npduNumber            uint8
	nextExtensionHeader1  uint8
	extensionHeaderLength uint8
	pduType               uint8
	qfi                   uint8
	nextExtensionHeader2  uint8
}

const gtp5gHdrLen = 16

type gtp5gCork struct {
	ipv4  packet.IPv4Hdr
	udp   packet.UDPHdr
	gtp5g gtp5gHdr
}

func recalcAfterEnq(pdr *pdr) {
	//fmt.Println("→ before recalcAfterEnq:")
	//debugDump()
	//fmt.Println("← before recalcAfterEnq:")
	if pdr.pktq.qlen() == 1 {
		queuedPdrs.insert(pdr)
	}
	/* XXX: */
	//fmt.Println("→ after recalcAfterEnq:")
	//debugDump()
	//fmt.Println("← after recalcAfterEnq:")
}

func recalcAfterDeq(pdr *pdr, size uint, now uint64) {
	//fmt.Println("→ before recalcAfterDeq:")
	//debugDump()
	//fmt.Println("← before recalcAfterDeq:")
	if len(pdr.qers) == 0 {
		pdr.nextTx = now
		if pdr.pktq.qlen() != 0 {
			queuedPdrs.insert(pdr)
		}
		return
	}
	if pdr.far.destinationInterface == IV_ACCESS {
		size = size - types.EtherLen - types.VLANLen - types.IPv4MinLen - types.UDPLen - gtp5gHdrLen
	} else {
		size = size - types.EtherLen - types.VLANLen
	}
	for _, qer := range pdr.qers {
		if pdr.isUl() {
			if qer.ulBpsDelta > 0 {
				qer.nextUlTx = now + qer.ulBpsDelta*uint64(size)
			} else {
				qer.nextUlTx = now
			}
			if qer.ulPpsDelta > 0 {
				nextUlTx := now + qer.ulPpsDelta
				if int64(nextUlTx-qer.nextUlTx) > 0 {
					qer.nextUlTx = nextUlTx
				}
			}
			for _, pTmp := range qer.ulPdrs {
				pTmp.nextTxUpdated = false
			}
		} else {
			if qer.dlBpsDelta > 0 {
				qer.nextDlTx = now + qer.dlBpsDelta*uint64(size)
			} else {
				qer.nextDlTx = now
			}
			if qer.dlPpsDelta > 0 {
				nextDlTx := now + qer.dlPpsDelta
				if int64(nextDlTx-qer.nextDlTx) > 0 {
					qer.nextDlTx = nextDlTx
				}
			}
			for _, pTmp := range qer.dlPdrs {
				pTmp.nextTxUpdated = false
			}
		}
	}
	for _, qer := range pdr.qers {
		if pdr.isUl() {
			for _, pTmp := range qer.ulPdrs {
				if pTmp.nextTxUpdated {
					continue
				}
				latestNext := pTmp.qers[0].nextUlTx
				for _, qTmp := range pTmp.qers {
					if int64(latestNext-qTmp.nextUlTx) < 0 {
						latestNext = qTmp.nextUlTx
					}
				}
				if pTmp.nextTx != latestNext {
					removed := false
					if pTmp != pdr && queuedPdrs.exists(pTmp) {
						queuedPdrs.remove(pTmp)
						removed = true
					}
					pTmp.nextTx = latestNext
					if pTmp != pdr && removed {
						queuedPdrs.insert(pTmp)
					}
				}
				pTmp.nextTxUpdated = true
			}
		} else {
			for _, pTmp := range qer.dlPdrs {
				if pTmp.nextTxUpdated {
					continue
				}
				latestNext := pTmp.qers[0].nextDlTx
				for _, qTmp := range pTmp.qers {
					if int64(latestNext-qTmp.nextDlTx) < 0 {
						latestNext = qTmp.nextDlTx
					}
				}
				if pTmp.nextTx != latestNext {
					removed := false
					if pTmp != pdr && queuedPdrs.exists(pTmp) {
						queuedPdrs.remove(pTmp)
						removed = true
					}
					pTmp.nextTx = latestNext
					if pTmp != pdr && removed {
						queuedPdrs.insert(pTmp)
					}
				}
				pTmp.nextTxUpdated = true
			}
		}
	}
	if pdr.pktq.qlen() != 0 {
		queuedPdrs.insert(pdr)
	}
	/* XXX: */
	//fmt.Println("→ after recalcAfterDeq:")
	//debugDump()
	//fmt.Println("← after recalcAfterDeq:")
}

func deqable(pdr *pdr, now uint64) bool {
	return int64(now-pdr.nextTx) >= 0
}

func n6PdrLookup(pkt *packet.Packet) *pdr {
	ipv4 := pkt.GetIPv4CheckVLAN()
	var n6SessionKey n6SessionKey
	n6SessionKey = types.IPv4ToBytes(ipv4.DstAddr)
	//fmt.Println("n6PdrLookup: n6SessionKey =", n6SessionKey)
	session, ok := n6SessionMap[n6SessionKey]
	if !ok {
		return nil
	}
	if len(session.n6Pdrs) == 0 {
		return nil
	}
	//fmt.Println("n6PdrLookup: XXX")
	/* XXX: */
	return session.n6Pdrs[0]
}

func n3n9PdrLookup(pkt *packet.Packet) *pdr {
	ipv4 := pkt.GetIPv4CheckVLAN()
	gtp5g := (*gtp5gHdr)(unsafe.Pointer(pkt.Data))
	var n3n9SessionKey n3n9SessionKey
	binary.LittleEndian.PutUint32(n3n9SessionKey[0:4], gtp5g.teid)
	binary.LittleEndian.PutUint32(n3n9SessionKey[4:8], uint32(ipv4.DstAddr))
	//fmt.Println("n3n9PdrLookup: n3n9SessionKey =", n3n9SessionKey)
	session, ok := n3n9SessionMap[n3n9SessionKey]
	if !ok {
		return nil
	}
	if len(session.n3n9Pdrs) == 0 {
		return nil
	}
	//fmt.Println("n3n9PdrLookup: XXX")
	/* XXX: */
	return session.n3n9Pdrs[0]
}

func n6Handler(pkt *packet.Packet) *pdr {
	pkt.ParseL3CheckVLAN()
	ipv4 := pkt.GetIPv4CheckVLAN()
	if ipv4 == nil {
		return nil
	}
	pdr := n6PdrLookup(pkt)
	if pdr == nil {
		return nil
	}
	totalLength := packet.SwapBytesUint16(ipv4.TotalLength)
	if !pkt.EncapsulateHead(types.EtherLen+types.VLANLen, types.IPv4MinLen+types.UDPLen+gtp5gHdrLen) {
		return nil
	}
	vlan := pkt.GetVLAN()
	corkp := (*gtp5gCork)(unsafe.Pointer(uintptr(unsafe.Pointer(pkt.Ether)) + types.EtherLen + types.VLANLen))
	*corkp = pdr.far.cork
	gtp5g := (*gtp5gHdr)(unsafe.Pointer(uintptr(unsafe.Pointer(pkt.Ether)) + types.EtherLen + types.VLANLen + types.IPv4MinLen + types.UDPLen))
	//gtp5g.headerType = 0x34
	//gtp5g.messageType = 0xff
	gtp5g.messageLength = packet.SwapBytesUint16(totalLength + 8)
	//gtp5g.teid = packet.SwapBytesUint32(87)
	//gtp5g.sequenceNumber = 0x0000
	//gtp5g.npduNumber = 0x00
	//gtp5g.nextExtensionHeader1 = 0x85
	//gtp5g.extensionHeaderLength = 0x01
	//gtp5g.pduType = 0x00
	//gtp5g.qfi = 0x01
	//gtp5g.nextExtensionHeader2 = 0x00
	udp := (*packet.UDPHdr)(unsafe.Pointer(uintptr(unsafe.Pointer(pkt.Ether)) + types.EtherLen + types.VLANLen + types.IPv4MinLen))
	//udp.SrcPort = packet.SwapUDPPortGTPU
	//udp.DstPort = packet.SwapUDPPortGTPU
	udp.DgramLen = packet.SwapBytesUint16(totalLength + 8 + 16)
	//udp.DgramCksum = 0x0000
	ipv4 = (*packet.IPv4Hdr)(unsafe.Pointer(uintptr(unsafe.Pointer(pkt.Ether)) + types.EtherLen + types.VLANLen))
	//ipv4.VersionIhl = 0x45
	//ipv4.TypeOfService = 0x00
	ipv4.TotalLength = packet.SwapBytesUint16(totalLength + 8 + 16 + 20)
	pdr.far.pid += 1
	ipv4.PacketID = pdr.far.pid
	//ipv4.FragmentOffset = 0x0000
	//ipv4.TimeToLive = 64
	//ipv4.NextProtoID = 17
	//ipv4.SrcAddr = 0x010110ac
	//ipv4.DstAddr =  0x020110ac
	ipv4.HdrChecksum = packet.SwapBytesUint16(packet.CalculateIPv4Checksum(ipv4))
	vlan.SetVLANTagIdentifier(n3n9VlanId)
	pkt.Ether.DAddr = n3n9DstMacAddr
	pkt.Ether.SAddr = localMacAddr
	return pdr
}

func n3n9Handler(pkt *packet.Packet) *pdr {
	pkt.ParseL3CheckVLAN()
	ipv4 := pkt.GetIPv4CheckVLAN()
	if ipv4 == nil {
		return nil
	}
	pkt.ParseL4ForIPv4()
	udp := pkt.GetUDPForIPv4()
	if udp == nil || udp.DstPort != packet.SwapUDPPortGTPU {
		return nil
	}
	pkt.ParseL7(types.UDPNumber)
	gtp5g := (*gtp5gHdr)(unsafe.Pointer(pkt.Data))
	pdr := n3n9PdrLookup(pkt)
	if pdr == nil {
		return nil
	}
	if pdr.outerHeaderRemoval {
		if !pkt.DecapsulateHead(types.EtherLen+types.VLANLen, types.IPv4MinLen+types.UDPLen+gtp5gHdrLen) {
			return nil
		}
		vlan := pkt.GetVLAN()
		vlan.SetVLANTagIdentifier(n6VlanId)
		pkt.Ether.DAddr = n6DstMacAddr
		pkt.Ether.SAddr = localMacAddr
	} else {
		gtp5g.teid = pdr.far.cork.gtp5g.teid
		ipv4.DstAddr = pdr.far.cork.ipv4.DstAddr
		ipv4.SrcAddr = pdr.far.cork.ipv4.SrcAddr
		ipv4.HdrChecksum = packet.SwapBytesUint16(packet.CalculateIPv4Checksum(ipv4))
		vlan := pkt.GetVLAN()
		vlan.SetVLANTagIdentifier(n3n9VlanId)
		pkt.Ether.DAddr = n3n9DstMacAddr
		pkt.Ether.SAddr = localMacAddr
	}
	return pdr
}

func xlHandler(pkt *packet.Packet) *pdr {
	vlan := pkt.GetVLAN()
	if vlan == nil {
		return nil
	}
	switch vlan.GetVLANTagIdentifier() {
	case n6VlanId:
		return n6Handler(pkt)
	case n3n9VlanId:
		return n3n9Handler(pkt)
	}
	return nil
}

func xlEnq(buf uintptr, enqed *bool) {
	lock.RLock()
	defer lock.RUnlock()
	now := tsc()
	pkt := packet.ExtractPacket(buf)
	pdr := xlHandler(pkt)
	if pdr == nil {
		*enqed = false
		return
	}
	if pdr.pktq.qlen() < pdr.pktq.size-1 {
		pdr.pktq.enq(buf, now)
		*enqed = true
		recalcAfterEnq(pdr)
		return
	}
	*enqed = false
}

func xlDeq(buf *uintptr, deqed *bool) {
	lock.RLock()
	defer lock.RUnlock()
	pdr := queuedPdrs.head
	if pdr == nil {
		*deqed = false
		return
	}
	now := tsc()
	if deqable(pdr, now) {
		queuedPdrs.remove(pdr)
		*buf, _ = pdr.pktq.deq()
		*deqed = true
		pkt := packet.ExtractPacket(*buf)
		size := pkt.GetPacketLen()
		recalcAfterDeq(pdr, size, now)
		return
	}
	*deqed = false
}

func Run() {
	config := flow.Config{
		CPUList: cpuList,
	}
	flow.CheckFatal(flow.SystemInit(&config))

	xlFlow, err := flow.SetReceiver(port)
	flow.CheckFatal(err)

	flow.CheckFatal(flow.SetEnqerDeqer(xlFlow, xlEnq, xlDeq))

	flow.CheckFatal(flow.SetSender(xlFlow, port))

	flow.CheckFatal(flow.SystemStart())
}
