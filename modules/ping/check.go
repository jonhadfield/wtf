package ping

import (
	"fmt"
	"github.com/wtfutil/wtf/logger"
	"net"
)

const (
	ProtocolICMP     = 1
	ProtocolIPv6ICMP = 58
	msgFail          = "fail"
	msgSuccess       = "success"
)

func checkTarget(t *net.IP, pingTimeout int, privileged bool, logging bool) (result string) {
	if t.To4() != nil {
		result = checkIPV4(t.String(), pingTimeout, privileged, logging)
		return
	}

	result = checkIPV6(t.String(), pingTimeout, false, logging)

	return
}

func ipsFromTarget(t string) (ips []*net.IP, isIP bool, err error) {
	// try to parse target as an IP
	if i := net.ParseIP(t); i != nil {
		// return the target parsed as an IP
		return []*net.IP{&i}, true, nil
	}

	// try to parse target as an FQDN
	var pIPs []string

	pIPs, err = net.LookupHost(t)
	if err != nil {
		logger.Log(fmt.Sprintf("%s | lookup failed for: %s", moduleName, t))
		return
	}

	for x := 0; x < len(pIPs); x++ {
		pIP := net.ParseIP(pIPs[x])
		ips = append(ips, &pIP)
	}

	return
}

type target struct {
	raw string
	ips []*net.IP
	err error
}

func parseTargets(ts []string) (targets []target) {
	for _, t := range ts {
		res, isIP, err := ipsFromTarget(t)
		// if parsing shortens ipv6 address then use shorter version for checking and display
		if isIP && res[0].String() != t {
			t = res[0].String()
		}

		targets = append(targets, target{
			raw: t,
			ips: res,
			err: err,
		})
	}

	return
}
