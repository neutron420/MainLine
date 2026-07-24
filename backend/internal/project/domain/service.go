package domain

import (
	"context"
	"fmt"
)

type ErrProjectNotFound     struct{ ID string }
type ErrProjectSlugConflict struct{ Slug string }
type ErrMemberNotFound      struct{ ProjectID, UserID string }
type ErrLastOwner           struct{}

func (e ErrProjectNotFound) Error() string     { return fmt.Sprintf("project %s not found", e.ID) }
func (e ErrProjectSlugConflict) Error() string  { return fmt.Sprintf("project slug %s already exists", e.Slug) }
func (e ErrMemberNotFound) Error() string       { return fmt.Sprintf("member %s not found in project %s", e.UserID, e.ProjectID) }
func (e ErrLastOwner) Error() string            { return "cannot remove the last owner from a project" }

type ProjectService struct {
	projRepo ProjectRepository
}

func NewProjectService(projRepo ProjectRepository) *ProjectService {
	return &ProjectService{projRepo: projRepo}
}

func (s *ProjectService) Create(ctx context.Context, name, description, visibilityStr, userID string) (*Project, error) {
	visibility, err := ValidateVisibility(visibilityStr)
	if err != nil {
		return nil, err
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

func (s *ProjectService) AddMember(ctx context.Context, projectID, userID, roleStr, actorID string) error {
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

	member := &ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	}
	return s.projRepo.AddMember(ctx, member)
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
