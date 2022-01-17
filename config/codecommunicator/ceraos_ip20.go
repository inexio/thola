package codecommunicator

import (
	"context"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/deviceclass/groupproperty"
	"github.com/rs/zerolog/log"
	"strings"
)

type ceraosIP20Communicator struct {
	codeCommunicator
}

// GetInterfaces returns the interfaces of ceraos/ip20 devices.
func (c *ceraosIP20Communicator) GetInterfaces(ctx context.Context, filter ...groupproperty.Filter) ([]device.Interface, error) {
	interfaces, err := c.deviceClass.GetInterfaces(ctx, filter...)
	if err != nil {
		return nil, err
	}

	model, err := c.deviceClass.GetModel(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to get model of IP20 device")
		return interfaces, nil
	}
	if !strings.HasSuffix(model, "2 radio") {
		return interfaces, nil
	}

	var maxbitrateIn *uint64
	var maxbitrateOut *uint64
	for _, interf := range interfaces {
		if interf.Radio != nil && interf.Radio.MaxbitrateIn != nil {
			rate := *interf.Radio.MaxbitrateIn
			if maxbitrateIn != nil {
				rate += *maxbitrateIn
			}
			maxbitrateIn = &rate
		}
		if interf.Radio != nil && interf.Radio.MaxbitrateOut != nil {
			rate := *interf.Radio.MaxbitrateOut
			if maxbitrateOut != nil {
				rate += *maxbitrateOut
			}
			maxbitrateOut = &rate
		}
	}

	if maxbitrateIn != nil || maxbitrateOut != nil {
		for i, interf := range interfaces {
			if interf.IfName != nil && strings.HasPrefix(*interf.IfName, "Multi Carrier ABC") {
				if interf.Radio != nil {
					interfaces[i].Radio.MaxbitrateIn = maxbitrateIn
					interfaces[i].Radio.MaxbitrateIn = maxbitrateOut
				} else {
					interfaces[i].Radio = &device.RadioInterface{
						MaxbitrateIn:  maxbitrateIn,
						MaxbitrateOut: maxbitrateOut,
					}
				}
			}
		}
		return filterInterfaces(ctx, interfaces, filter)
	}

	return interfaces, nil
}
