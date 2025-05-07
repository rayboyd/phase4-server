// SPDX-License-Identifier: Apache-2.0
package endpoint

import (
	"phase4/internal/p4/runtime/stage"
	"phase4/internal/p4/transport"
)

type UdpComponent struct {
	sender transport.Component
	stage.BaseActor
}

/*
Mesage types
- UdpDataMessage: Contains the data payload to be sent via UDP. "udp.data"
*/

type UdpDataMessage struct {
	Payload any
}

func (m *UdpDataMessage) Type() string {
	return "udp.data"
}
