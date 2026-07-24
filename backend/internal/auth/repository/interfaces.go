// Package repository re-exports domain repository interfaces for convenience.
// The actual interface definitions live in internal/auth/domain/ to follow
// the DDD pattern where domain defines the contracts.
package repository

import (
	"github.com/schemahub/backend/internal/auth/domain"
)

type UserRepository = domain.UserRepository
type RefreshTokenRepository = domain.RefreshTokenRepository
type OAuthIdentityRepository = domain.OAuthIdentityRepository
type VerificationTokenRepository = domain.VerificationTokenRepository
