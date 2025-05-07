// SPDX-License-Identifier: Apache-2.0
package analysis

type MockDSPComponent struct {
	ProcessFunc func(input []int32)
	CloseFunc   func() error
	ID          string
	Closed      bool
}

func (m *MockDSPComponent) Process(input []int32) {
	if m.ProcessFunc != nil {
		m.ProcessFunc(input)
	}
}

func (m *MockDSPComponent) Close() error {
	m.Closed = true
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
