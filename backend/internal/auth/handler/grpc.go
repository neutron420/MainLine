package handler

import (
	"context"
	"errors"
	"time"

	"github.com/schemahub/backend/internal/auth/domain"
	"github.com/schemahub/backend/internal/pkg/interceptor"
	authv1 "github.com/schemahub/backend/proto/auth/v1"
	commonv1 "github.com/schemahub/backend/proto/common/v1"
)

func userIDFromContext(ctx context.Context) string {
	id, _ := interceptor.UserIDFromContext(ctx)
	return id
}

type AuthHandler struct {
	authv1.UnimplementedAuthServiceServer
	svc *domain.AuthService
}

func NewAuthHandler(svc *domain.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	user, accessToken, refreshToken, err := h.svc.Register(ctx, req.Email, req.Password, req.DisplayName)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.RegisterResponse{
		User:         toProtoUser(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	user, accessToken, refreshToken, err := h.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.LoginResponse{
		User:         toProtoUser(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
	}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	accessToken, refreshToken, err := h.svc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if err := h.svc.Logout(ctx, req.RefreshToken); err != nil {
		return nil, mapError(err)
	}
	return &authv1.LogoutResponse{}, nil
}

func (h *AuthHandler) GetCurrentUser(ctx context.Context, req *authv1.GetCurrentUserRequest) (*authv1.GetCurrentUserResponse, error) {
	user, err := h.svc.GetUserByID(ctx, userIDFromContext(ctx))
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.GetCurrentUserResponse{User: toProtoUser(user)}, nil
}

func (h *AuthHandler) UpdateUser(ctx context.Context, req *authv1.UpdateUserRequest) (*authv1.UpdateUserResponse, error) {
	user, err := h.svc.UpdateUser(ctx, userIDFromContext(ctx), req.DisplayName, req.AvatarUrl)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.UpdateUserResponse{User: toProtoUser(user)}, nil
}

func (h *AuthHandler) ChangePassword(ctx context.Context, req *authv1.ChangePasswordRequest) (*authv1.ChangePasswordResponse, error) {
	if err := h.svc.ChangePassword(ctx, userIDFromContext(ctx), req.CurrentPassword, req.NewPassword); err != nil {
		return nil, mapError(err)
	}
	return &authv1.ChangePasswordResponse{}, nil
}

func (h *AuthHandler) GetOAuthURL(ctx context.Context, req *authv1.GetOAuthURLRequest) (*authv1.GetOAuthURLResponse, error) {
	AuthUrl, stateToken, err := h.svc.GetOAuthURL(req.Provider, req.RedirectTo, req.Linking, req.CodeChallenge)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.GetOAuthURLResponse{AuthUrl: AuthUrl, StateToken: stateToken}, nil
}

func (h *AuthHandler) HandleOAuthCallback(ctx context.Context, req *authv1.HandleOAuthCallbackRequest) (*authv1.HandleOAuthCallbackResponse, error) {
	user, accessToken, refreshToken, isNew, needsLinking, err := h.svc.HandleOAuthCallback(ctx, req.Provider, req.Code, req.State, req.CodeVerifier)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.HandleOAuthCallbackResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
		User:         toProtoUser(user),
		IsNewUser:    isNew,
		NeedsLinking: needsLinking,
	}, nil
}

func (h *AuthHandler) LinkOAuthIdentity(ctx context.Context, req *authv1.LinkOAuthIdentityRequest) (*authv1.LinkOAuthIdentityResponse, error) {
	if err := h.svc.LinkOAuthIdentity(ctx, userIDFromContext(ctx), req.Provider, req.Code, req.State); err != nil {
		return nil, mapError(err)
	}
	return &authv1.LinkOAuthIdentityResponse{}, nil
}

func (h *AuthHandler) UnlinkOAuthIdentity(ctx context.Context, req *authv1.UnlinkOAuthIdentityRequest) (*authv1.UnlinkOAuthIdentityResponse, error) {
	if err := h.svc.UnlinkOAuthIdentity(ctx, userIDFromContext(ctx), req.Provider); err != nil {
		return nil, mapError(err)
	}
	return &authv1.UnlinkOAuthIdentityResponse{}, nil
}

func (h *AuthHandler) ListLinkedIdentities(ctx context.Context, req *authv1.ListLinkedIdentitiesRequest) (*authv1.ListLinkedIdentitiesResponse, error) {
	identities, err := h.svc.ListLinkedIdentities(ctx, userIDFromContext(ctx))
	if err != nil {
		return nil, mapError(err)
	}
	var protoIdentities []*authv1.OAuthIdentity
	for _, id := range identities {
		protoIdentities = append(protoIdentities, &authv1.OAuthIdentity{
			Id:            id.ID,
			Provider:      id.Provider,
			ProviderEmail: id.ProviderEmail,
			CreatedAt:     id.CreatedAt.Format(time.RFC3339),
			LastUsedAt:    optionalTime(id.LastUsedAt),
		})
	}
	return &authv1.ListLinkedIdentitiesResponse{Identities: protoIdentities}, nil
}

func (h *AuthHandler) SendVerificationEmail(ctx context.Context, req *authv1.SendVerificationEmailRequest) (*authv1.SendVerificationEmailResponse, error) {
	if err := h.svc.SendVerificationEmail(ctx, req.Email); err != nil {
		return nil, mapError(err)
	}
	return &authv1.SendVerificationEmailResponse{}, nil
}

func (h *AuthHandler) VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) (*authv1.VerifyEmailResponse, error) {
	if err := h.svc.VerifyEmail(ctx, req.Token); err != nil {
		return nil, mapError(err)
	}
	return &authv1.VerifyEmailResponse{}, nil
}

func (h *AuthHandler) DeleteAccount(ctx context.Context, req *authv1.DeleteAccountRequest) (*authv1.DeleteAccountResponse, error) {
	if err := h.svc.DeleteAccount(ctx, userIDFromContext(ctx), req.Password); err != nil {
		return nil, mapError(err)
	}
	return &authv1.DeleteAccountResponse{}, nil
}

func (h *AuthHandler) ForgotPassword(ctx context.Context, req *authv1.ForgotPasswordRequest) (*authv1.ForgotPasswordResponse, error) {
	if err := h.svc.ForgotPassword(ctx, req.Email); err != nil {
		return nil, mapError(err)
	}
	return &authv1.ForgotPasswordResponse{}, nil
}

func (h *AuthHandler) ResetPassword(ctx context.Context, req *authv1.ResetPasswordRequest) (*authv1.ResetPasswordResponse, error) {
	if err := h.svc.ResetPassword(ctx, req.Token, req.Password); err != nil {
		return nil, mapError(err)
	}
	return &authv1.ResetPasswordResponse{}, nil
}

func toProtoUser(u *domain.User) *commonv1.User {
	if u == nil {
		return nil
	}
	return &commonv1.User{
		Id:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		AvatarUrl:   u.AvatarURL,
		Role:        u.Role,
		IsActive:    u.IsActive,
		CreatedAt:   u.CreatedAt.Format(time.RFC3339),
	}
}

func optionalTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		return grpcUnauthenticated("invalid email or password")
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		return grpcAlreadyExists("email already registered")
	case errors.Is(err, domain.ErrUserNotFound):
		return grpcNotFound("user not found")
	case errors.Is(err, domain.ErrInvalidRefreshToken):
		return grpcUnauthenticated("invalid or expired refresh token")
	case errors.Is(err, domain.ErrTokenRevoked):
		return grpcUnauthenticated("token revoked â€” all sessions invalidated")
	case errors.Is(err, domain.ErrPasswordMismatch):
		return grpcInvalidArgument("current password is incorrect")
	case errors.Is(err, domain.ErrWeakPassword):
		return grpcInvalidArgument("password must be 8+ chars with uppercase, lowercase, and digit")
	default:
		return grpcInternal("internal server error")
	}
}
