// SPDX-License-Identifier: Apache-2.0
package transport

type MockTransportComponent struct {
	SendDataFunc func(data []byte) error
	CloseFunc    func() error
	ID           string
	Closed       bool
}

func (m *MockTransportComponent) SendData(data []byte) error {
	if m.SendDataFunc != nil {
		return m.SendDataFunc(data)
	}
	return nil
}

func (m *MockTransportComponent) Close() error {
	m.Closed = true
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
