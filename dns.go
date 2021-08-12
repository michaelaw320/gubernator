package gubernator

import (
	"context"
	"time"

	"github.com/mailgun/holster/v4/setter"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type DNSPoolConfig struct {
	// (Required) The FQDN that should resolve to gubernator instance ip addresses
	FQDN string

	// (Required) Own ip address
	OwnAddress string

	// (Required) Called when the list of gubernators in the pool updates
	OnUpdate UpdateFunc

	Logger logrus.FieldLogger
}

type DNSPool struct {
	log    logrus.FieldLogger
	conf   DNSPoolConfig
	ctx    context.Context
	cancel context.CancelFunc
}

func NewDNSPool(conf DNSPoolConfig) (*DNSPool, error) {
	setter.SetDefault(&conf.Logger, logrus.WithField("category", "gubernator"))

	if conf.OwnAddress == "" {
		return nil, errors.New("Advertise.GRPCAddress is required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	pool := &DNSPool{
		log:    conf.Logger,
		conf:   conf,
		ctx:    ctx,
		cancel: cancel,
	}
	go pool.task()
	return pool, nil
}

func peer(ip string, self string, ipv6 bool) PeerInfo {

	if ipv6 {
		ip = "[" + ip + "]"
	}
	grpc := ip + ":81"
	return PeerInfo{
		DataCenter:  "",
		HTTPAddress: ip + ":80",
		GRPCAddress: ip + ":81",
		IsOwner:     grpc == self,
	}

}

func (x *DNSPool) task() {
	for {
		config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")
		c := new(dns.Client)
		c.SingleInflight = true
		var delay uint32 = 10
		var update []PeerInfo
		m4 := new(dns.Msg)
		m4.SetQuestion(dns.Fqdn(x.conf.FQDN), dns.TypeA)
		r4, _, err4 := c.Exchange(m4, config.Servers[0]+":"+config.Port)
		m6 := new(dns.Msg)
		m6.SetQuestion(dns.Fqdn(x.conf.FQDN), dns.TypeAAAA)
		r6, _, err6 := c.Exchange(m6, config.Servers[0]+":"+config.Port)
		if err4 == nil || err6 == nil {
			for _, rec := range r4.Answer {
				if rec.Header().Rrtype == dns.TypeA {
					delay = rec.Header().Ttl
					update = append(update, peer(rec.(*dns.A).A.String(), x.conf.OwnAddress, false))
				} else {
					x.log.Info("Ignored ", rec)
				}
			}
			for _, rec := range r6.Answer {
				if rec.Header().Rrtype == dns.TypeAAAA {
					delay = rec.Header().Ttl
					update = append(update, peer(rec.(*dns.AAAA).AAAA.String(), x.conf.OwnAddress, true))
				} else {
					x.log.Info("Ignored ", rec)
				}
			}
		} else {
			x.log.Error("Errors ", err4, err6)
		}
		x.log.Info("Update: ", update)
		x.conf.OnUpdate(update)
		x.log.Info("going to sleep for ", delay)
		select {
		case <-x.ctx.Done():
			return
		case <-time.After(time.Duration(delay) * time.Second):
		}
	}
}

func (x *DNSPool) Close() {
	x.cancel()
}
