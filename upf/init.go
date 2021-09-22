package upf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/intel-go/nff-go/packet"
	"github.com/intel-go/nff-go/types"

	"github.com/m-asama/nff-go-upf/config"
	//"github.com/m-asama/nff-go-upf/tsc"
)

func initGlobal(conf *config.Config) error {
	var err error
	if conf.Global == nil {
		return errors.New("global required")
	}
	if conf.Global.CpuList == nil {
		return errors.New("global.cpuList required")
	}
	cpuList = *conf.Global.CpuList
	if conf.Global.Local == nil {
		return errors.New("global.local required")
	}
	if conf.Global.Local.Port == nil {
		return errors.New("global.local.port required")
	}
	port = uint16(*conf.Global.Local.Port)
	if conf.Global.Local.Address == nil {
		return errors.New("global.local.address required")
	}
	localAddress, err := net.ParseMAC(*conf.Global.Local.Address)
	if err != nil {
		return errors.New("global.local.address invalid")
	}
	copy(localMacAddr[:], localAddress)
	if conf.Global.N6 == nil {
		return errors.New("global.n6 required")
	}
	if conf.Global.Local.TeAddress == nil {
		return errors.New("global.local.teAddress required")
	}
	teAddress = net.ParseIP(*conf.Global.Local.TeAddress)
	if teAddress == nil {
		return errors.New("global.local.teAddress invalid")
	}
	if conf.Global.N6.VlanId == nil {
		return errors.New("global.n6.vlanId required")
	}
	n6VlanId = uint16(*conf.Global.N6.VlanId)
	if conf.Global.N6.Address == nil {
		return errors.New("global.n6.address required")
	}
	n6Address, err := net.ParseMAC(*conf.Global.N6.Address)
	if err != nil {
		return errors.New("global.n6.address invalid")
	}
	copy(n6DstMacAddr[:], n6Address)
	if conf.Global.N3N9.VlanId == nil {
		return errors.New("global.n3n9.vlanId required")
	}
	n3n9VlanId = uint16(*conf.Global.N3N9.VlanId)
	if conf.Global.N3N9.Address == nil {
		return errors.New("global.n3n9.address required")
	}
	n3n9Address, err := net.ParseMAC(*conf.Global.N3N9.Address)
	if err != nil {
		return errors.New("global.n3n9.address invalid")
	}
	copy(n3n9DstMacAddr[:], n3n9Address)
	return nil
}

func initSessions(conf *config.Config) error {
	for i, confSession := range conf.Sessions {
		newSession := session{}
		if confSession.Fseid == nil {
			return fmt.Errorf("sessions[%d].fseid required", i)
		}
		newSession.fseid.seid = uint64(*confSession.Fseid.Seid)
		fseidAddress := net.ParseIP(*confSession.Fseid.Address)
		if fseidAddress == nil {
			return fmt.Errorf("sessions[%d].fseid.address invalid", i)
		}
		newSession.fseid.address = fseidAddress
		for j, confFar := range confSession.Fars {
			newFar := far{}
			if confFar.Farid == nil {
				return fmt.Errorf("sessions[%d].fars[%d].farid required", i, j)
			}
			newFar.farid = uint32(*confFar.Farid)
			if confFar.ApplyAction == nil {
				return fmt.Errorf("sessions[%d].fars[%d].applyAction required", i, j)
			}
			if *confFar.ApplyAction != 2 {
				return fmt.Errorf("sessions[%d].fars[%d].applyAction invalid", i, j)
			}
			if confFar.ForwardingParameters == nil {
				return fmt.Errorf("sessions[%d].fars[%d].forwardingParameters required", i, j)
			}
			if confFar.ForwardingParameters.DestinationInterface == nil {
				return fmt.Errorf("sessions[%d].fars[%d].forwardingParameters.destinationInterface required", i, j)
			}
			switch *confFar.ForwardingParameters.DestinationInterface {
			case "acccess":
				newFar.destinationInterface = IV_ACCESS
			case "core":
				newFar.destinationInterface = IV_CORE
			case "sgin6":
				newFar.destinationInterface = IV_SGILAN_N6LAN
			case "cpf":
				newFar.destinationInterface = IV_CP_FUNCTION
			default:
				return fmt.Errorf("sessions[%d].fars[%d].forwardingParameters.destinationInterface invalid", i, j)
			}
			if confFar.ForwardingParameters.OuterHeaderCreation != nil {
				if confFar.ForwardingParameters.OuterHeaderCreation.Teid == nil {
					return fmt.Errorf("sessions[%d].fars[%d].forwardingParameters.outerHeaderCreation.teid required", i, j)
				}
				if confFar.ForwardingParameters.OuterHeaderCreation.Address == nil {
					return fmt.Errorf("sessions[%d].fars[%d].forwardingParameters.outerHeaderCreation.address required", i, j)
				}
				ohcAddress := net.ParseIP(*confFar.ForwardingParameters.OuterHeaderCreation.Address)
				if ohcAddress == nil {
					return fmt.Errorf("sessions[%d].fars[%d].forwardingParameters.outerHeaderCreation.address invalid", i, j)
				}
				newFar.outerHeaderCreation = true
				newFar.cork.gtp5g.headerType = 0x34
				newFar.cork.gtp5g.messageType = 0xff
				//newFar.cork.gtp5g.messageLength = packet.SwapBytesUint16(totalLength + 8)
				//newFar.cork.gtp5g.teid = packet.SwapBytesUint32(87)
				newFar.cork.gtp5g.teid = packet.SwapBytesUint32(uint32(*confFar.ForwardingParameters.OuterHeaderCreation.Teid))
				newFar.cork.gtp5g.sequenceNumber = 0x0000
				newFar.cork.gtp5g.npduNumber = 0x00
				newFar.cork.gtp5g.nextExtensionHeader1 = 0x85
				newFar.cork.gtp5g.extensionHeaderLength = 0x01
				newFar.cork.gtp5g.pduType = 0x00
				newFar.cork.gtp5g.qfi = 0x00
				newFar.cork.gtp5g.nextExtensionHeader2 = 0x00
				newFar.cork.udp.SrcPort = packet.SwapUDPPortGTPU
				newFar.cork.udp.DstPort = packet.SwapUDPPortGTPU
				//newFar.cork.udp.DgramLen = packet.SwapBytesUint16(totalLength + 8 + 16)
				newFar.cork.udp.DgramCksum = 0x0000
				newFar.cork.ipv4.VersionIhl = 0x45
				newFar.cork.ipv4.TypeOfService = 0x00
				//newFar.cork.ipv4.TotalLength = packet.SwapBytesUint16(totalLength + 8 + 16 + 20)
				//pid += 1
				//newFar.cork.ipv4.PacketID = pid
				newFar.cork.ipv4.FragmentOffset = 0x0000
				newFar.cork.ipv4.TimeToLive = 64
				newFar.cork.ipv4.NextProtoID = 17
				//newFar.cork.ipv4.SrcAddr = 0x010110ac
				newFar.cork.ipv4.SrcAddr = types.SliceToIPv4(teAddress.To4())
				//newFar.cork.ipv4.DstAddr = 0x020110ac
				newFar.cork.ipv4.DstAddr = types.SliceToIPv4(ohcAddress.To4())
				//newFar.cork.ipv4.HdrChecksum = packet.SwapBytesUint16(packet.CalculateIPv4Checksum(ipv4))
			}
			newSession.fars = append(newSession.fars, &newFar)
		}
		for j, confQer := range confSession.Qers {
			newQer := qer{}
			if confQer.Qerid == nil {
				return fmt.Errorf("sessions[%d].qers[%d].qerid required", i, j)
			}
			newQer.qerid = uint32(*confQer.Qerid)
			if confQer.GateStatus == nil {
				return fmt.Errorf("sessions[%d].qers[%d].gateStatus required", i, j)
			}
			if *confQer.GateStatus != "open" {
				return fmt.Errorf("sessions[%d].qers[%d].gateStatus invalid", i, j)
			}
			if confQer.Mbr != nil {
				if confQer.Mbr.Ul != nil {
					newQer.mbrUl = *confQer.Mbr.Ul
					newQer.ulDelta = tscsec * 8 / newQer.mbrUl
				}
				if confQer.Mbr.Dl != nil {
					newQer.mbrDl = *confQer.Mbr.Dl
					newQer.dlDelta = tscsec * 8 / newQer.mbrDl
				}
			}
			if confQer.Qfi != nil {
				newQer.qfi = uint8(*confQer.Qfi)
			}
			newQer.queuedPdrs.init()
			newQer.nextUlTx = tsc()
			newQer.nextDlTx = tsc()
			newSession.qers = append(newSession.qers, &newQer)
		}
		for j, confPdr := range confSession.Pdrs {
			newPdr := pdr{}
			if confPdr.Pdrid == nil {
				return fmt.Errorf("sessions[%d].pdrs[%d].pdrid required", i, j)
			}
			newPdr.pdrid = uint16(*confPdr.Pdrid)
			if confPdr.Precedence == nil {
				return fmt.Errorf("sessions[%d].pdrs[%d].precedence required", i, j)
			}
			newPdr.precedence = uint32(*confPdr.Precedence)
			if confPdr.Pdi == nil {
				return fmt.Errorf("sessions[%d].pdrs[%d].pdi required", i, j)
			}
			if confPdr.Pdi.SourceInterface == nil {
				return fmt.Errorf("sessions[%d].pdrs[%d].pdi.sourceInterface required", i, j)
			}
			switch *confPdr.Pdi.SourceInterface {
			case "access":
				newPdr.pdi.sourceInterface = IV_ACCESS
			case "core":
				newPdr.pdi.sourceInterface = IV_CORE
			case "sgin6":
				newPdr.pdi.sourceInterface = IV_ACCESS
			case "cpf":
				newPdr.pdi.sourceInterface = IV_CP_FUNCTION
			default:
				return fmt.Errorf("sessions[%d].pdrs[%d].pdi.sourceInterface invalid", i, j)
			}
			if *confPdr.Pdi.SourceInterface == "access" && confPdr.Pdi.Fteid == nil {
				return fmt.Errorf("sessions[%d].pdrs[%d].pdi.fteid required when source access", i, j)
			}
			if confPdr.Pdi.Fteid != nil {
				if confPdr.Pdi.Fteid.Teid == nil {
					return fmt.Errorf("sessions[%d].pdrs[%d].pdi.fteid.teid required", i, j)
				}
				newPdr.pdi.fteid.teid = uint32(*confPdr.Pdi.Fteid.Teid)
				if confPdr.Pdi.Fteid.Address == nil {
					return fmt.Errorf("sessions[%d].pdrs[%d].pdi.fteid.address required", i, j)
				}
				fteidAddress := net.ParseIP(*confPdr.Pdi.Fteid.Address)
				if fteidAddress == nil {
					return fmt.Errorf("sessions[%d].pdrs[%d].pdi.fteid.address invalid", i, j)
				}
				newPdr.pdi.fteid.address = fteidAddress
			}
			if confPdr.Pdi.UeIpAddress == nil {
				return fmt.Errorf("sessions[%d].pdrs[%d].pdi.ueIpAddress required", i, j)
			}
			ueIpAddress := net.ParseIP(*confPdr.Pdi.UeIpAddress)
			if ueIpAddress == nil {
				return fmt.Errorf("sessions[%d].pdrs[%d].pdi.ueIpAddress invalid", i, j)
			}
			newPdr.pdi.ueIpAddress = ueIpAddress
			if confPdr.Pdi.SdfFilter != nil {
				if confPdr.Pdi.SdfFilter.SourcePrefix != nil {
					_, sourcePrefix, err := net.ParseCIDR(*confPdr.Pdi.SdfFilter.SourcePrefix)
					if err != nil || sourcePrefix == nil {
						return fmt.Errorf("sessions[%d].pdrs[%d].pdi.sdfFilter.sourcePrefix invalid")
					}
					newPdr.pdi.sdfFilter.sourcePrefix = *sourcePrefix
				}
				if confPdr.Pdi.SdfFilter.DestinationPrefix != nil {
					_, destinationPrefix, err := net.ParseCIDR(*confPdr.Pdi.SdfFilter.DestinationPrefix)
					if err != nil || destinationPrefix == nil {
						return fmt.Errorf("sessions[%d].pdrs[%d].pdi.sdfFilter.destinationPrefix invalid")
					}
					newPdr.pdi.sdfFilter.destinationPrefix = *destinationPrefix
				}
				if confPdr.Pdi.SdfFilter.Protocol != nil {
					var protocol uint16
					switch *confPdr.Pdi.SdfFilter.Protocol {
					case "tcp":
						protocol = 6
					case "udp":
						protocol = 17
					default:
					}
					newPdr.pdi.sdfFilter.protocol = protocol
				}
				if confPdr.Pdi.SdfFilter.SourcePorts != nil {
					sourcePorts := parsePorts(*confPdr.Pdi.SdfFilter.SourcePorts)
					if sourcePorts == nil {
						return fmt.Errorf("sessions[%d].pdrs[%d].pdi.sdfFilter.sourcePorts invalid")
					}
					newPdr.pdi.sdfFilter.sourcePorts = sourcePorts
				}
				if confPdr.Pdi.SdfFilter.DestinationPorts != nil {
					destinationPorts := parsePorts(*confPdr.Pdi.SdfFilter.DestinationPorts)
					if destinationPorts == nil {
						return fmt.Errorf("sessions[%d].pdrs[%d].pdi.sdfFilter.destinationPorts invalid")
					}
					newPdr.pdi.sdfFilter.destinationPorts = destinationPorts
				}
			}
			if confPdr.OuterHeaderRemoval != nil && *confPdr.OuterHeaderRemoval {
				newPdr.outerHeaderRemoval = true
			}
			if confPdr.Farid == nil {
				return fmt.Errorf("sessions[%d].pdrs[%d].farid required", i, j)
			}
			farFound := false
			for _, far := range newSession.fars {
				if far.farid == uint32(*confPdr.Farid) {
					farFound = true
					newPdr.far = far
				}
			}
			if !farFound {
				return fmt.Errorf("sessions[%d].pdrs[%d] far not found", i, j)
			}
			if confPdr.Qerids == nil || len(confPdr.Qerids) == 0 {
				return fmt.Errorf("sessions[%d].pdrs[%d].qerids required", i, j)
			}
			newPdr.qers = make([]*qer, 0)
			for _, qerid := range confPdr.Qerids {
				qerFound := false
				for _, qer := range newSession.qers {
					if qer.qerid == uint32(qerid) {
						qerFound = true
						newPdr.qers = append(newPdr.qers, qer)
					}
				}
				if !qerFound {
					return fmt.Errorf("sessions[%d].pdrs[%d] qer not found", i, j)
				}
			}
			newPdr.pktq = newPktq(1024)
			if newPdr.pdi.fteid.teid == 0 && newPdr.pdi.fteid.address == nil {
				newSession.n6Pdrs = append(newSession.n6Pdrs, &newPdr)
				var n6SessionKey n6SessionKey
				ueIpAddress := newPdr.pdi.ueIpAddress.To4()
				copy(n6SessionKey[0:4], ueIpAddress)
				if session, ok := n6SessionMap[n6SessionKey]; ok {
					if session != &newSession {
						return fmt.Errorf("sessions[%d].pdrs[%d] n6SessionKey dup", i, j)
					}
				} else {
					n6SessionMap[n6SessionKey] = &newSession
				}
			} else {
				newSession.n3n9Pdrs = append(newSession.n3n9Pdrs, &newPdr)
				var n3n9SessionKey n3n9SessionKey
				binary.BigEndian.PutUint32(n3n9SessionKey[0:4], newPdr.pdi.fteid.teid)
				fteidAddress := newPdr.pdi.fteid.address.To4()
				copy(n3n9SessionKey[4:8], fteidAddress)
				if session, ok := n3n9SessionMap[n3n9SessionKey]; ok {
					if session != &newSession {
						return fmt.Errorf("sessions[%d].pdrs[%d] n3n9SessionKey dup", i, j)
					}
				} else {
					n3n9SessionMap[n3n9SessionKey] = &newSession
				}
			}
		}
		sort.Slice(newSession.n6Pdrs, func(i, j int) bool { return newSession.n6Pdrs[i].precedence > newSession.n6Pdrs[j].precedence })
		sort.Slice(newSession.n3n9Pdrs, func(i, j int) bool { return newSession.n3n9Pdrs[i].precedence > newSession.n3n9Pdrs[j].precedence })
		sessions = append(sessions, &newSession)
	}
	return nil
}

func parsePorts(s string) []portRange {
	portRanges := make([]portRange, 0)
	a1 := strings.Split(s, ",")
	for _, r1 := range a1 {
		portRange := portRange{}
		a2 := strings.Split(r1, "-")
		if len(a2) == 1 {
			port, err := strconv.Atoi(a2[0])
			if err != nil || port <= 0 {
				return nil
			}
			portRange.begin = uint16(port)
			portRange.end = uint16(port)
		} else if len(a2) == 2 {
			beg, err := strconv.Atoi(a2[0])
			if err != nil || beg <= 0 {
				return nil
			}
			end, err := strconv.Atoi(a2[1])
			if err != nil || end <= 0 {
				return nil
			}
			if beg > end {
				tmp := beg
				beg = end
				end = tmp
			}
			portRange.begin = uint16(beg)
			portRange.end = uint16(end)
		} else {
			return nil
		}
		portRanges = append(portRanges, portRange)
	}
	return portRanges
}

func Init(conf *config.Config) error {
	var err error
	if conf == nil {
		return errors.New("conf is null")
	}
	err = initGlobal(conf)
	if err != nil {
		return err
	}
	err = initSessions(conf)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	sessions = make([]*session, 0)
	n6SessionMap = make(map[n6SessionKey]*session)
	n3n9SessionMap = make(map[n3n9SessionKey]*session)
	queuedQers.init()

	t1 := tsc()
	time.Sleep(time.Second)
	t2 := tsc()
	tscsec = (t2 - t1) / 1000000 * 1000000
}
