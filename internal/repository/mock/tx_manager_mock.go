package mock

import (
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockTxManager struct {
	mock.Mock
}

// Transaction immediately executes the function with nil tx
// This is fine for unit tests because repository calls are already mocked
func (m *MockTxManager) Transaction(fc func(tx *gorm.DB) error) error {
	return fc(nil)
}
