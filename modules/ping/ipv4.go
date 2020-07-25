package ping

import (
	"bytes"
	"fmt"
	"github.com/wtfutil/wtf/logger"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"math/rand"
	"net"
	"time"
)

func getPingDest4(target string, privileged bool) (dst net.Addr, err error) {
	if privileged {
		return net.ResolveIPAddr("ip4", target)
	}

	var ipaddr *net.IPAddr

	ipaddr, err = net.ResolveIPAddr("ip4", target)
	if err != nil {
		return
	}

	return &net.UDPAddr{IP: ipaddr.IP, Zone: ipaddr.Zone}, nil
}

func checkIPV4(target string, pingTimeout int, privileged bool, logging bool) (result string) {
	dst, err := getPingDest4(target, privileged)

	var conn *icmp.PacketConn

	conn, err = getConn4(privileged, logging)
	if err != nil {
		return msgFail
	}

	logger.Log(conn.LocalAddr().String())

	defer conn.Close()

	pSeq := rand.Intn(65535)
	pID := rand.Intn(65535)
	pData := []byte("wtf")

	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
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

	n, err = conn.WriteTo(b, dst)
	if err != nil || n != len(b) {
		if logging {
			logger.Log(fmt.Sprintf("%s | failed to send ping to %s", moduleName, dst.String()))
		}

		return msgFail
	}

	reply := make([]byte, 1500)

	waitStart := time.Now()
	waitDuration := time.Duration(pingTimeout) * time.Millisecond
	maxWait := 1 * time.Second

	for {
		if time.Since(waitStart) > maxWait {
			if logging {
				logger.Log(fmt.Sprintf("%s | timed out after %ds waiting for: %s",
					moduleName, pingTimeout, target))
			}

			return msgFail
		}

		if err = conn.SetReadDeadline(time.Now().Add(waitDuration)); err != nil {

			if logging {
				logger.Log(fmt.Sprintf("%s | failed to set read deadline: %v", moduleName, err))
			}
			return msgFail
		}

		var cm *ipv4.ControlMessage
		conn.IPv4PacketConn().SetControlMessage(ipv4.FlagTTL, true)
		conn.IPv4PacketConn().SetControlMessage(ipv4.FlagInterface, true)
		conn.IPv4PacketConn().SetControlMessage(ipv4.FlagSrc, true)


		conn.IPv4PacketConn().SetControlMessage(ipv4.FlagDst, true)

		n, cm, _, err = conn.IPv4PacketConn().ReadFrom(reply)

		if err != nil {

			if logging {
				logger.Log(fmt.Sprintf("%s | failed to read reply for target: %s %v", moduleName, target, err))
			}

			return msgFail
		}

		if cm.Src != nil && cm.Src.String() != target {
			// we've read reply to another ping so try again
			logger.Log(fmt.Sprintf("%s | got reply for wrong target %s", moduleName, dst.String()))
			continue
		}


		// we received a reply from the intended recipient
		var rm *icmp.Message
		rm, err = icmp.ParseMessage(ProtocolICMP, reply[:n])
		if err != nil {
			return msgFail
		}

		// check the reply matches the request
		if rm.Type == ipv4.ICMPTypeEchoReply {
			pe := rm.Body.(*icmp.Echo)
			if pID == pe.ID && pSeq == pe.Seq && bytes.Equal(pe.Data, pData) {
				if logging {
					logger.Log(fmt.Sprintf("%s | got reply for %s", moduleName, dst.String()))
				}
				return msgSuccess
			}

			// we've read from the correct peer, just not the correct response
			logger.Log(fmt.Sprintf("%s | got reply for %s but packet doesn't match", moduleName, dst.String()))
			continue
		}
		logger.Log("f")
		break
	}

	return msgFail
}

func getConn4(privileged, logging bool) (conn *icmp.PacketConn, err error) {
	if privileged {
		conn, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
		if err != nil {
			if logging {
				logger.Log(fmt.Sprintf("%s | failed to listen for ip4:icmp packets", moduleName))
			}
		}

		return
	}

	// get unprivileged connection
	conn, err = icmp.ListenPacket("udp4", "")
	if err != nil {
		if logging {
			logger.Log(fmt.Sprintf("%s | failed to listen for udp4 packets: %+v", "ping", err))
		}
	}

	return
}
