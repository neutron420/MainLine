package domain

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type ErrProjectNotFound struct{ ID string }
type ErrProjectSlugConflict struct{ Slug string }
type ErrMemberNotFound struct{ ProjectID, UserID string }
type ErrLastOwner struct{}
type ErrUserNotFoundByEmail struct{ Email string }
type ErrNoUserSpecified struct{}
type ErrInvitationNotFound struct{ Token string }
type ErrInvitationExpired struct{}
type ErrInvitationAlreadyUsed struct{}
type ErrAlreadyMember struct{ Email string }

func (e ErrProjectNotFound) Error() string { return fmt.Sprintf("project %s not found", e.ID) }
func (e ErrProjectSlugConflict) Error() string {
	return fmt.Sprintf("project slug %s already exists", e.Slug)
}
func (e ErrMemberNotFound) Error() string {
	return fmt.Sprintf("member %s not found in project %s", e.UserID, e.ProjectID)
}
func (e ErrLastOwner) Error() string { return "cannot remove the last owner from a project" }
func (e ErrUserNotFoundByEmail) Error() string {
	return fmt.Sprintf("no registered user found with email %s", e.Email)
}
func (e ErrNoUserSpecified) Error() string {
	return "specify either user_id or email of the user to add"
}
func (e ErrInvitationNotFound) Error() string {
	return fmt.Sprintf("invitation with token %s was not found", e.Token)
}
func (e ErrInvitationExpired) Error() string {
	return "this invitation has expired"
}
func (e ErrInvitationAlreadyUsed) Error() string {
	return "this invitation has already been used"
}
func (e ErrAlreadyMember) Error() string {
	return fmt.Sprintf("user with email %s is already a member", e.Email)
}

// UserLookup resolves a registered user ID from an email address so that
// project members can be invited by email instead of raw user IDs.
type UserLookup interface {
	GetByEmail(ctx context.Context, email string) (string, error)
}

// InviteMailer delivers project invitation emails. It is optional: when no
// mailer is configured, invitations are created but delivery is skipped
// (dev mode), matching the behaviour of the auth mailer.
type InviteMailer interface {
	SendInvitationEmail(ctx context.Context, to, projectName, token string) error
}

type ProjectService struct {
	projRepo ProjectRepository
	users    UserLookup
	mailer   InviteMailer
}

func NewProjectService(projRepo ProjectRepository, users UserLookup) *ProjectService {
	return &ProjectService{projRepo: projRepo, users: users}
}

func (s *ProjectService) SetMailer(m InviteMailer) {
	s.mailer = m
}

func newInvitationToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating invitation token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (s *ProjectService) Create(ctx context.Context, name, description, visibilityStr, template, userID string) (*Project, error) {
	visibility, err := ValidateVisibility(visibilityStr)
	if err != nil {
		return nil, err
	}

	tmpl := template
	if tmpl == "" {
		tmpl = "blank"
	}
	if !ValidTemplate(tmpl) {
		return nil, fmt.Errorf("invalid template: %s (must be blank, starter, or ecommerce)", tmpl)
	}

	slug := GenerateSlug(name)

	existing, _ := s.projRepo.GetBySlug(ctx, slug)
	if existing != nil {
		return nil, ErrProjectSlugConflict{Slug: slug}
	}

	p := &Project{
		Name:        name,
		Slug:        slug,
		Description: description,
		Visibility:  visibility,
		Template:    tmpl,
		CreatedBy:   userID,
	}

	if err := s.projRepo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("creating project: %w", err)
	}

	m := &ProjectMember{
		ProjectID: p.ID,
		UserID:    userID,
		Role:      RoleOwner,
	}
	if err := s.projRepo.AddMember(ctx, m); err != nil {
		return nil, fmt.Errorf("adding owner: %w", err)
	}

	created, err := s.projRepo.GetByID(ctx, p.ID)
	if err != nil {
		return nil, fmt.Errorf("retrieving created project: %w", err)
	}

	return created, nil
}

func (s *ProjectService) GetByID(ctx context.Context, id string) (*Project, error) {
	p, err := s.projRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrProjectNotFound{ID: id}
	}
	return p, nil
}

func (s *ProjectService) List(ctx context.Context, userID, cursor string, pageSize int32) ([]*Project, string, int32, error) {
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.projRepo.ListByUserID(ctx, userID, cursor, pageSize)
}

func (s *ProjectService) Update(ctx context.Context, id, name, description, visibilityStr, actorID string) (*Project, error) {
	p, err := s.projRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrProjectNotFound{ID: id}
	}

	m, err := s.projRepo.GetMember(ctx, id, actorID)
	if err != nil || !m.Role.CanManageConnections() {
		return nil, fmt.Errorf("permission denied")
	}

	if name != "" {
		p.Name = name
	}
	if description != "" {
		p.Description = description
	}
	if visibilityStr != "" {
		v, err := ValidateVisibility(visibilityStr)
		if err != nil {
			return nil, err
		}
		p.Visibility = v
	}

	if err := s.projRepo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("updating project: %w", err)
	}

	return p, nil
}

func (s *ProjectService) Delete(ctx context.Context, id, actorID string) error {
	p, err := s.projRepo.GetByID(ctx, id)
	if err != nil {
		return ErrProjectNotFound{ID: id}
	}

	m, err := s.projRepo.GetMember(ctx, id, actorID)
	if err != nil || m.Role != RoleOwner {
		return fmt.Errorf("permission denied: only the owner can delete a project")
	}
	_ = p

	return s.projRepo.SoftDelete(ctx, id)
}

func (s *ProjectService) AddMember(ctx context.Context, projectID, userID, email, roleStr, actorID string) error {
	m, err := s.projRepo.GetMember(ctx, projectID, actorID)
	if err != nil || !m.Role.CanManageMembers() {
		return fmt.Errorf("permission denied")
	}

	role, err := ValidateRole(roleStr)
	if err != nil {
		return err
	}

	if role == RoleOwner && m.Role != RoleOwner {
		return fmt.Errorf("only the owner can assign the owner role")
	}

	if userID == "" && email == "" {
		return ErrNoUserSpecified{}
	}
	if userID == "" {
		userID, err = s.users.GetByEmail(ctx, email)
		if err != nil {
			return ErrUserNotFoundByEmail{Email: email}
		}
	}

	member := &ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	}
	return s.projRepo.AddMember(ctx, member)
}

// InviteMember adds a registered user directly, or creates an email
// invitation (with a 7-day token) for unregistered users. Returns the
// invitation ID, or an empty string when the user was added directly.
func (s *ProjectService) InviteMember(ctx context.Context, projectID, email, roleStr, actorID string) (string, error) {
	m, err := s.projRepo.GetMember(ctx, projectID, actorID)
	if err != nil || !m.Role.CanManageMembers() {
		return "", fmt.Errorf("permission denied")
	}

	role, err := ValidateRole(roleStr)
	if err != nil {
		return "", err
	}
	if role == RoleOwner && m.Role != RoleOwner {
		return "", fmt.Errorf("only the owner can assign the owner role")
	}
	if email == "" {
		return "", ErrNoUserSpecified{}
	}

	// Registered user → add them directly, no invitation needed.
	if userID, err := s.users.GetByEmail(ctx, email); err == nil {
		if existing, _ := s.projRepo.GetMember(ctx, projectID, userID); existing != nil {
			return "", ErrAlreadyMember{Email: email}
		}
		member := &ProjectMember{
			ProjectID: projectID,
			UserID:    userID,
			Role:      role,
			InvitedBy: &actorID,
		}
		if err := s.projRepo.AddMember(ctx, member); err != nil {
			return "", fmt.Errorf("adding member: %w", err)
		}
		return "", nil
	}

	token, err := newInvitationToken()
	if err != nil {
		return "", err
	}
	inv := &ProjectInvitation{
		ProjectID: projectID,
		Email:     email,
		Role:      role,
		Token:     token,
		Status:    InvitationPending,
		InvitedBy: actorID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.projRepo.CreateInvitation(ctx, inv); err != nil {
		return "", fmt.Errorf("creating invitation: %w", err)
	}

	if s.mailer != nil {
		proj, err := s.projRepo.GetByID(ctx, projectID)
		if err != nil {
			return inv.ID, fmt.Errorf("loading project for invite email: %w", err)
		}
		if err := s.mailer.SendInvitationEmail(ctx, email, proj.Name, token); err != nil {
			return inv.ID, fmt.Errorf("sending invitation email: %w", err)
		}
	}
	return inv.ID, nil
}

// AcceptInvitation validates a token and joins the accepting user to the
// project. It is idempotent: an already-member gets the project ID back.
func (s *ProjectService) AcceptInvitation(ctx context.Context, token, userID string) (string, error) {
	inv, err := s.projRepo.GetInvitationByToken(ctx, token)
	if err != nil {
		return "", ErrInvitationNotFound{Token: token}
	}
	if inv.Status != InvitationPending {
		return "", ErrInvitationAlreadyUsed{}
	}
	if time.Now().After(inv.ExpiresAt) {
		return "", ErrInvitationExpired{}
	}

	// Idempotent join: already a member → accept and return project.
	if existing, _ := s.projRepo.GetMember(ctx, inv.ProjectID, userID); existing != nil {
		if err := s.projRepo.MarkInvitationAccepted(ctx, inv.ID, inv.ProjectID, userID); err != nil {
			return "", fmt.Errorf("marking invitation accepted: %w", err)
		}
		return inv.ProjectID, nil
	}

	member := &ProjectMember{
		ProjectID: inv.ProjectID,
		UserID:    userID,
		Role:      inv.Role,
		InvitedBy: &inv.InvitedBy,
	}
	if err := s.projRepo.AddMember(ctx, member); err != nil {
		return "", fmt.Errorf("joining project: %w", err)
	}
	if err := s.projRepo.MarkInvitationAccepted(ctx, inv.ID, inv.ProjectID, userID); err != nil {
		return "", fmt.Errorf("marking invitation accepted: %w", err)
	}
	return inv.ProjectID, nil
}

func (s *ProjectService) RemoveMember(ctx context.Context, projectID, userID, actorID string) error {
	m, err := s.projRepo.GetMember(ctx, projectID, actorID)
	if err != nil || !m.Role.CanManageMembers() {
		return fmt.Errorf("permission denied")
	}

	target, err := s.projRepo.GetMember(ctx, projectID, userID)
	if err != nil {
		return ErrMemberNotFound{ProjectID: projectID, UserID: userID}
	}

	if target.Role == RoleOwner {
		members, err := s.projRepo.ListMemberUsers(ctx, projectID)
		if err != nil {
			return fmt.Errorf("listing members: %w", err)
		}
		ownerCount := 0
		for _, mm := range members {
			if mm.Role == RoleOwner {
				ownerCount++
			}
		}
		if ownerCount <= 1 {
			return ErrLastOwner{}
		}
	}

	return s.projRepo.RemoveMember(ctx, projectID, userID)
}

func (s *ProjectService) UpdateMemberRole(ctx context.Context, projectID, userID, roleStr, actorID string) error {
	m, err := s.projRepo.GetMember(ctx, projectID, actorID)
	if err != nil || !m.Role.CanManageMembers() {
		return fmt.Errorf("permission denied")
	}

	role, err := ValidateRole(roleStr)
	if err != nil {
		return err
	}

	if role == RoleOwner && m.Role != RoleOwner {
		return fmt.Errorf("only the owner can assign the owner role")
	}

	if m.UserID == userID && m.Role == RoleOwner && role != RoleOwner {
		members, err := s.projRepo.ListMemberUsers(ctx, projectID)
		if err != nil {
			return fmt.Errorf("listing members: %w", err)
		}
		ownerCount := 0
		for _, mm := range members {
			if mm.Role == RoleOwner {
				ownerCount++
			}
		}
		if ownerCount <= 1 {
			return ErrLastOwner{}
		}
	}

	return s.projRepo.UpdateMemberRole(ctx, projectID, userID, role)
}

func (s *ProjectService) ListMembers(ctx context.Context, projectID, cursor string, pageSize int32, actorID string) ([]*ProjectMember, string, int32, error) {
	_, err := s.projRepo.GetMember(ctx, projectID, actorID)
	if err != nil {
		return nil, "", 0, fmt.Errorf("permission denied")
	}

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.projRepo.ListMembers(ctx, projectID, cursor, pageSize)
}
