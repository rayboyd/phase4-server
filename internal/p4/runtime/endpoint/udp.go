// SPDX-License-Identifier: Apache-2.0
package endpoint

import (
	"context"
	"log"
	"phase4/internal/p4/runtime/stage"
	"phase4/internal/p4/transport"
)

func NewUdpComponent(id string, capacity int, sender transport.Component) *UdpComponent {
	if sender == nil {
		log.Panicf("UdpComponent requires a non-nil DataSender")
	}

	a := &UdpComponent{
		sender: sender,
	}
	a.BaseActor = *stage.NewBaseActor(id, capacity, a.processMessage)

	return a
}

func (a *UdpComponent) processMessage(ctx context.Context, msg stage.Message) {
	// switch m := msg.(type) {
	// case *UdpDataMessage:
	// 	_ = a.sender.SendData(m.Payload)

	// default:
	// 	// log something about unexpected message type
	// }
}
