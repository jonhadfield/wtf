package ping

import (
	"bytes"
	"fmt"
	"github.com/wtfutil/wtf/logger"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv6"
	"math/rand"
	"net"
	"time"
)

func checkIPV6(target string, pingTimeout int, privileged bool, logging bool) (result string) {
	dst, err := getPingDest6(target, privileged, logging)
	if err != nil {
		return msgFail
	}

	var conn6 *icmp.PacketConn

	conn6, err = getConn6(privileged, logging)
	if err != nil {
		return msgFail
	}

	defer conn6.Close()

	pSeq := rand.Intn(65535)
	pID := rand.Intn(65535)
	pData := []byte("")

	m := icmp.Message{
		Type: ipv6.ICMPTypeEchoRequest, Code: 0,
		Body: &icmp.Echo{
			ID:   pID,
			Seq:  pSeq,
			Data: pData,
		},
	}

	var b []byte

	b, err = m.Marshal(nil)
	if err != nil {
		return msgFail
	}

	var n int

	if logging {
		logger.Log(fmt.Sprintf("%s | pinging: %s", moduleName, dst.String()))
	}

	n, err = conn6.WriteTo(b, dst)
	if err != nil || n != len(b) {
		if logging {
			logger.Log(fmt.Sprintf("%s | failed to send ping to %s: %+v", moduleName, dst.String(), err))
		}

		return msgFail
	}

	reply := make([]byte, 1500)

	waitStart := time.Now()
	waitDuration := time.Duration(pingTimeout) * time.Millisecond
	waitMax := 5 * time.Second

	for {
		if time.Since(waitStart) > waitMax {
			if logging {
				logger.Log(fmt.Sprintf("%s | timed out waiting for: %s", moduleName, target))
			}

			return msgFail
		}

		conn6.IPv6PacketConn().SetControlMessage(ipv6.FlagHopLimit, true)

		if err = conn6.SetReadDeadline(time.Now().Add(waitDuration)); err != nil {
			if logging {
				logger.Log(fmt.Sprintf("%s | failed to set read deadline: %v", moduleName, err))
			}
			return msgFail
		}

		var peer net.Addr

		var cm *ipv6.ControlMessage
		n, cm, peer, err = conn6.IPv6PacketConn().ReadFrom(reply)
		logger.Log(fmt.Sprintf("CM for %s %+v", target, cm))
		if err != nil {
			return msgFail
		}

		if dst.String() != peer.String() {
			continue
		}

		var rm *icmp.Message

		rm, err = icmp.ParseMessage(ProtocolIPv6ICMP, reply[:n])
		if err != nil {
			return msgFail
		}

		if rm.Type == ipv6.ICMPTypeEchoReply {
			pe := rm.Body.(*icmp.Echo)
			if pID == pe.ID && pSeq == pe.Seq && bytes.Equal(pe.Data, pData) {
				if logging {
					logger.Log(fmt.Sprintf("%s | got reply for %s", moduleName, dst.String()))
				}
				return msgSuccess
			}

			if err = conn6.SetReadDeadline(time.Now().Add(waitDuration)); err != nil {
				if logging {
					logger.Log(fmt.Sprintf("%s | failed to set read deadline: %v", moduleName, err))
				}
				return msgFail
			}

			continue
		}

		if logging {
			logger.Log(fmt.Sprintf("%s | got reply for %s", moduleName, dst.String()))
		}

		break
	}

	return msgFail
}

func getConn6(privileged, logging bool) (conn6 *icmp.PacketConn, err error) {
	if privileged {
		conn6, err = icmp.ListenPacket("ip6:ipv6-icmp", "::")
		if err != nil {
			if logging {
				logger.Log(fmt.Sprintf(" %s | failed to listen for ip6:ipv6-icmp packets", moduleName))
			}

			return
		}
	} else {
		conn6, err = icmp.ListenPacket("udp6", "::")
		if err != nil {
			if logging {
				logger.Log(fmt.Sprintf(" %s | failed to listen for udp6 packets", moduleName))
			}

			return
		}
	}

	return
}

func getPingDest6(target string, privileged, logging bool) (dst net.Addr, err error) {
	if privileged {
		dst, err = net.ResolveIPAddr("ip6", target)
		if err != nil {
			if logging {
				logger.Log(fmt.Sprintf("%s | failed to resolve %s", moduleName, target))
			}

		}
		return
	}

	var ipaddr *net.IPAddr
	ipaddr, err = net.ResolveIPAddr("ip6", target)
	if err != nil {
		return
	}

	return &net.UDPAddr{IP: ipaddr.IP, Zone: ipaddr.Zone}, nil
}
