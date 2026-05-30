package mocks

import (
	"github.com/Bengkelin/bengkelin-service/internal/repository"
)

// Compile-time interface verification
// These variables ensure that the mock implementations properly satisfy the interfaces
var (
	_ repository.UserRepositoryInterface    = (*MockUserRepository)(nil)
	_ repository.MitraRepositoryInterface   = (*MockMitraRepository)(nil)
	_ repository.BengkelRepositoryInterface = (*MockBengkelRepository)(nil)
)
