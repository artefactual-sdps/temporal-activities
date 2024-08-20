package bagvalidate

type mockValidator struct {
	err error
}

func NewMockValidator() *mockValidator {
	return &mockValidator{}
}

func (m mockValidator) Validate(_ string) error {
	return m.err
}

func (m *mockValidator) SetErr(e error) *mockValidator {
	m.err = e
	return m
}

var _ BagValidator = mockValidator{}
