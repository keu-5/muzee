package database

import (
	"github.com/google/wire"
)

// ProviderSet is the wire provider set for database layer
var ProviderSet = wire.NewSet(
	ConnectDatabase,
)