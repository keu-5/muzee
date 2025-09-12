package repository

import (
	"github.com/google/wire"
)

// ProviderSet is the wire provider set for repository layer
var ProviderSet = wire.NewSet(
	NewUserRepository,
)