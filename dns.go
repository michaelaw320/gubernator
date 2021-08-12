package gubernator

type DNSConfig struct {
	// (Required) The FQDN that should resolve to gubernator instance ip addresses
	FQDN string
}
