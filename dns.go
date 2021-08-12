package gubernator

import (
	"time"

	"github.com/mailgun/holster/v4/setter"
	"github.com/sirupsen/logrus"
)

type DNSPoolConfig struct {
	// (Required) The FQDN that should resolve to gubernator instance ip addresses
	FQDN string

	// (Required) Called when the list of gubernators in the pool updates
	OnUpdate UpdateFunc

	Logger logrus.FieldLogger
}

type DNSPool struct {
	peers map[string]PeerInfo
	log   logrus.FieldLogger
	conf  DNSPoolConfig
}

func NewDNSPool(conf DNSPoolConfig) (*DNSPool, error) {
	setter.SetDefault(&conf.Logger, logrus.WithField("category", "gubernator"))

	//ctx, cancel := context.WithCancel(context.Background())
	pool := &DNSPool{
		peers: make(map[string]PeerInfo),
		log:   conf.Logger,
		//cancelCtx: cancel,
		conf: conf,
		//ctx:       ctx,
	}
	go pool.task()
	return pool, nil
}

func (x *DNSPool) task() {
	for {
		time.Sleep(10 * time.Millisecond)
	}
}

func (x *DNSPool) Close() {

}
