package observability

import (
	"github.com/msmclass/samvaad/pkg/samvaad/logger"
	"github.com/msmclass/samvaad/pkg/samvaad/observability/agentsobs"
	"github.com/msmclass/samvaad/pkg/samvaad/observability/agentsv3obs"
	"github.com/msmclass/samvaad/pkg/samvaad/observability/corecallobs"
	"github.com/msmclass/samvaad/pkg/samvaad/observability/egressobs"
	"github.com/msmclass/samvaad/pkg/samvaad/observability/gatewayobs"
	"github.com/msmclass/samvaad/pkg/samvaad/observability/ingressobs"
	"github.com/msmclass/samvaad/pkg/samvaad/observability/roomobs"
	"github.com/msmclass/samvaad/pkg/samvaad/observability/sipcallobs"
	"github.com/msmclass/samvaad/pkg/samvaad/observability/storageobs"
	"github.com/msmclass/samvaad/pkg/samvaad/observability/telephonyobs"
)

const Project = "samvaad"

type Reporter interface {
	Logger(name, projectID string) (logger.Logger, error)
	Room() roomobs.Reporter
	Agent() agentsobs.Reporter
	AgentV3() agentsv3obs.Reporter
	Gateway() gatewayobs.Reporter
	Telephony() telephonyobs.Reporter
	Egress() egressobs.Reporter
	Ingress() ingressobs.Reporter
	SIPCall() sipcallobs.Reporter
	CoreCall() corecallobs.Reporter
	Storage() storageobs.Reporter
	Close()
}

func NewReporter() Reporter {
	return reporter{}
}

type reporter struct{}

func (reporter) Logger(name, projectID string) (logger.Logger, error) {
	return logger.GetDiscardLogger(), nil
}

func (reporter) Room() roomobs.Reporter {
	return roomobs.NewNoopReporter()
}

func (reporter) Agent() agentsobs.Reporter {
	return agentsobs.NewNoopReporter()
}

func (reporter) AgentV3() agentsv3obs.Reporter {
	return agentsv3obs.NewNoopReporter()
}

func (reporter) Gateway() gatewayobs.Reporter {
	return gatewayobs.NewNoopReporter()
}

func (reporter) Telephony() telephonyobs.Reporter {
	return telephonyobs.NewNoopReporter()
}

func (reporter) Egress() egressobs.Reporter {
	return egressobs.NewNoopReporter()
}

func (reporter) Ingress() ingressobs.Reporter {
	return ingressobs.NewNoopReporter()
}

func (reporter) SIPCall() sipcallobs.Reporter {
	return sipcallobs.NewNoopReporter()
}

func (reporter) CoreCall() corecallobs.Reporter {
	return corecallobs.NewNoopReporter()
}

func (reporter) Storage() storageobs.Reporter { return storageobs.NewNoopReporter() }

func (reporter) Close() {
}
