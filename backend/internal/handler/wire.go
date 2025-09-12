package handler

import (
	"github.com/google/wire"
)

// ProviderSet is the wire provider set for handler layer
var ProviderSet = wire.NewSet(
	NewHandler,
)