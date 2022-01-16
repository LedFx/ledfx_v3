package util

import (
	"fmt"
)

const (
	properMac        = "5C:49:7D:96:36:11"
	properMacNoColon = "5C497D963611"
)

func CleanMacAddress(addr string) (clean string, err error) {
	// The address length should match the MAC address spec
	if len(addr) != 17 && len(addr) != 12 {
		return "", fmt.Errorf("MAC address length invalid. Valid examples: [%s, %s]", properMac, properMacNoColon)
	}

	switch len(addr) {
	case 17:
		// Colons were provided
		for i := 2; i < len(addr); i += 3 {
			if addr[i] != ':' {
				return "", fmt.Errorf("MAC address colon order invalid. Valid example: [%s]", properMac)
			}
		}
	case 12:
		// Colons were not provided.
		// They must be added to match the format of the addresses
		// returned during discovery.
		for i := 2; i < len(addr); i += 3 {
			addr = addr[:i] + ":" + addr[i:]
		}
	}
	return addr, nil
}
