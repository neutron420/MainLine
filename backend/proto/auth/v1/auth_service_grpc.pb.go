package authv1

import (
	"context"

	"google.golang.org/grpc"
)

// AuthServiceServer is the interface for the Auth service.
type AuthServiceServer interface {
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error)
	Logout(ctx context.Context, req *LogoutRequest) (*LogoutResponse, error)
	GetCurrentUser(ctx context.Context, req *GetCurrentUserRequest) (*GetCurrentUserResponse, error)
	UpdateUser(ctx context.Context, req *UpdateUserRequest) (*UpdateUserResponse, error)
	ChangePassword(ctx context.Context, req *ChangePasswordRequest) (*ChangePasswordResponse, error)
	GetOAuthURL(ctx context.Context, req *GetOAuthURLRequest) (*GetOAuthURLResponse, error)
	HandleOAuthCallback(ctx context.Context, req *HandleOAuthCallbackRequest) (*HandleOAuthCallbackResponse, error)
	LinkOAuthIdentity(ctx context.Context, req *LinkOAuthIdentityRequest) (*LinkOAuthIdentityResponse, error)
	UnlinkOAuthIdentity(ctx context.Context, req *UnlinkOAuthIdentityRequest) (*UnlinkOAuthIdentityResponse, error)
	ListLinkedIdentities(ctx context.Context, req *ListLinkedIdentitiesRequest) (*ListLinkedIdentitiesResponse, error)
	SendVerificationEmail(ctx context.Context, req *SendVerificationEmailRequest) (*SendVerificationEmailResponse, error)
	VerifyEmail(ctx context.Context, req *VerifyEmailRequest) (*VerifyEmailResponse, error)
	DeleteAccount(ctx context.Context, req *DeleteAccountRequest) (*DeleteAccountResponse, error)
}

func RegisterAuthServiceServer(s *grpc.Server, srv AuthServiceServer) {
	s.RegisterService(&AuthService_ServiceDesc, srv)
}

var AuthService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "schemahub.auth.v1.AuthService",
	Methods: []grpc.MethodDesc{
		{MethodName: "Register", Handler: _AuthService_Register_Handler},
		{MethodName: "Login", Handler: _AuthService_Login_Handler},
		{MethodName: "RefreshToken", Handler: _AuthService_RefreshToken_Handler},
		{MethodName: "Logout", Handler: _AuthService_Logout_Handler},
		{MethodName: "GetCurrentUser", Handler: _AuthService_GetCurrentUser_Handler},
		{MethodName: "UpdateUser", Handler: _AuthService_UpdateUser_Handler},
		{MethodName: "ChangePassword", Handler: _AuthService_ChangePassword_Handler},
		{MethodName: "GetOAuthURL", Handler: _AuthService_GetOAuthURL_Handler},
		{MethodName: "HandleOAuthCallback", Handler: _AuthService_HandleOAuthCallback_Handler},
		{MethodName: "LinkOAuthIdentity", Handler: _AuthService_LinkOAuthIdentity_Handler},
		{MethodName: "UnlinkOAuthIdentity", Handler: _AuthService_UnlinkOAuthIdentity_Handler},
		{MethodName: "ListLinkedIdentities", Handler: _AuthService_ListLinkedIdentities_Handler},
		{MethodName: "SendVerificationEmail", Handler: _AuthService_SendVerificationEmail_Handler},
		{MethodName: "VerifyEmail", Handler: _AuthService_VerifyEmail_Handler},
		{MethodName: "DeleteAccount", Handler: _AuthService_DeleteAccount_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "auth/v1/auth_service.proto",
}

func _AuthService_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &RegisterRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(AuthServiceServer).Register(ctx, req)
}

func _AuthService_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &LoginRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(AuthServiceServer).Login(ctx, req)
}

func _AuthService_RefreshToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &RefreshTokenRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(AuthServiceServer).RefreshToken(ctx, req)
}

func _AuthService_Logout_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &LogoutRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(AuthServiceServer).Logout(ctx, req)
}

func _AuthService_GetCurrentUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &GetCurrentUserRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(AuthServiceServer).GetCurrentUser(ctx, req)
}

func _AuthService_UpdateUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &UpdateUserRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(AuthServiceServer).UpdateUser(ctx, req)
}

func _AuthService_ChangePassword_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &ChangePasswordRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(AuthServiceServer).ChangePassword(ctx, req)
}

func _AuthService_GetOAuthURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &GetOAuthURLRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(AuthServiceServer).GetOAuthURL(ctx, req)
}

func _AuthService_HandleOAuthCallback_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &HandleOAuthCallbackRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(AuthServiceServer).HandleOAuthCallback(ctx, req)
}

func _AuthService_LinkOAuthIdentity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &LinkOAuthIdentityRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(AuthServiceServer).LinkOAuthIdentity(ctx, req)
}

func _AuthService_UnlinkOAuthIdentity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &UnlinkOAuthIdentityRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(AuthServiceServer).UnlinkOAuthIdentity(ctx, req)
}

func _AuthService_ListLinkedIdentities_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &ListLinkedIdentitiesRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(AuthServiceServer).ListLinkedIdentities(ctx, req)
}

func _AuthService_SendVerificationEmail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &SendVerificationEmailRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(AuthServiceServer).SendVerificationEmail(ctx, req)
}

func _AuthService_VerifyEmail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &VerifyEmailRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(AuthServiceServer).VerifyEmail(ctx, req)
}

func _AuthService_DeleteAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &DeleteAccountRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(AuthServiceServer).DeleteAccount(ctx, req)
}

// UnimplementedAuthServiceServer provides default implementations.
type UnimplementedAuthServiceServer struct{}

func (UnimplementedAuthServiceServer) Register(_ context.Context, _ *RegisterRequest) (*RegisterResponse, error) {
	return nil, nil
}
func (UnimplementedAuthServiceServer) Login(_ context.Context, _ *LoginRequest) (*LoginResponse, error) {
	return nil, nil
}
func (UnimplementedAuthServiceServer) RefreshToken(_ context.Context, _ *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	return nil, nil
}
func (UnimplementedAuthServiceServer) Logout(_ context.Context, _ *LogoutRequest) (*LogoutResponse, error) {
	return nil, nil
}
func (UnimplementedAuthServiceServer) GetCurrentUser(_ context.Context, _ *GetCurrentUserRequest) (*GetCurrentUserResponse, error) {
	return nil, nil
}
func (UnimplementedAuthServiceServer) UpdateUser(_ context.Context, _ *UpdateUserRequest) (*UpdateUserResponse, error) {
	return nil, nil
}
func (UnimplementedAuthServiceServer) ChangePassword(_ context.Context, _ *ChangePasswordRequest) (*ChangePasswordResponse, error) {
	return nil, nil
}
func (UnimplementedAuthServiceServer) GetOAuthURL(_ context.Context, _ *GetOAuthURLRequest) (*GetOAuthURLResponse, error) {
	return nil, nil
}
func (UnimplementedAuthServiceServer) HandleOAuthCallback(_ context.Context, _ *HandleOAuthCallbackRequest) (*HandleOAuthCallbackResponse, error) {
	return nil, nil
}
func (UnimplementedAuthServiceServer) LinkOAuthIdentity(_ context.Context, _ *LinkOAuthIdentityRequest) (*LinkOAuthIdentityResponse, error) {
	return nil, nil
}
func (UnimplementedAuthServiceServer) UnlinkOAuthIdentity(_ context.Context, _ *UnlinkOAuthIdentityRequest) (*UnlinkOAuthIdentityResponse, error) {
	return nil, nil
}
func (UnimplementedAuthServiceServer) ListLinkedIdentities(_ context.Context, _ *ListLinkedIdentitiesRequest) (*ListLinkedIdentitiesResponse, error) {
	return nil, nil
}
func (UnimplementedAuthServiceServer) SendVerificationEmail(_ context.Context, _ *SendVerificationEmailRequest) (*SendVerificationEmailResponse, error) {
	return nil, nil
}
func (UnimplementedAuthServiceServer) VerifyEmail(_ context.Context, _ *VerifyEmailRequest) (*VerifyEmailResponse, error) {
	return nil, nil
}
func (UnimplementedAuthServiceServer) DeleteAccount(_ context.Context, _ *DeleteAccountRequest) (*DeleteAccountResponse, error) {
	return nil, nil
}
