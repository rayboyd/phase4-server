// SPDX-License-Identifier: Apache-2.0
package transport

func NewUdpTransport(addr, path string) (*UdpTransport, error) {
	udp := &UdpTransport{}

	return udp, nil
}

func (udp *UdpTransport) SendData(data []byte) error {
	return nil
}

func (udp *UdpTransport) Close() error {
	return nil
}
