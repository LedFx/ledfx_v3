package device

type NetworkerConfig struct {
	IP   string `mapstructure:"ip" json:"ip" description:"Device IP address on the LAN" validate:"required,ip"`
	Port int    `mapstructure:"port" json:"port" description:"Port number the device is listening on" default:"21324" validate:"gte=0,lte=65536"`
}
