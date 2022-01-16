package bluetooth

import "time"

type SearchTargetConfig struct {

	// DeviceAddress: The MAC address of a device.
	//
	// When this field is populated with a value,
	// DeviceRegex is ignored.
	DeviceAddress string

	// DeviceRegex: The regex-formatted search query for the target device.
	DeviceRegex string

	// ConnectRetryCoolDown: The duration that the bluetooth adapter should wait before
	// trying again after a failed connection attempt.
	ConnectRetryCoolDown time.Duration
}
