package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/schemahub/backend/internal/pkg/jwt"
	projectv1 "github.com/schemahub/backend/proto/project/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	ownerID = "fbb2634a-557a-4f79-84b5-1054580ce1da" // rihankumar2004@gmail.com
	otherID = "15e65c40-8018-4002-87b2-05ba5183faa4" // fnaticritesh2004@gmail.com
)

func main() {
	ctx := context.Background()
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	client := projectv1.NewProjectServiceClient(conn)

	priv := strings.ReplaceAll(os.Getenv("JWT_PRIVATE_KEY"), `\n`, "\n")
	pub := strings.ReplaceAll(os.Getenv("JWT_PUBLIC_KEY"), `\n`, "\n")
	mgr, err := jwt.NewManager(priv, pub)
	if err != nil {
		log.Fatalf("jwt manager: %v", err)
	}
	token := func(userID, email string) string {
		claims := jwtlib.MapClaims{
			"sub":   userID,
			"email": email,
			"role":  "user",
			"exp":   time.Now().Add(15 * time.Minute).Unix(),
		}
		s, err := mgr.SignClaims(claims)
		if err != nil {
			log.Fatalf("sign: %v", err)
		}
		return s
	}
	ownerCtx := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token(ownerID, "rihankumar2004@gmail.com"))
	otherCtx := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token(otherID, "fnaticritesh2004@gmail.com"))

	fail := func(step string, err error) {
		log.Printf("FAIL %s: %v", step, err)
	}
	pass := func(step string, extra ...string) {
		log.Printf("PASS %s %s", step, strings.Join(extra, " "))
	}

	// 1. Create project as owner
	created, err := client.CreateProject(ownerCtx, &projectv1.CreateProjectRequest{
		Name:        "Invite E2E Test",
		Description: "created by inviteprobe",
		Visibility:  "private",
		Template:    "blank",
	})
	if err != nil {
		fail("CreateProject", err)
		return
	}
	projectID := created.Project.Id
	pass("CreateProject", "id="+projectID)

	// 2. Invite unregistered email → invitation created
	inv, err := client.InviteMember(ownerCtx, &projectv1.InviteMemberRequest{
		ProjectId: projectID, Email: "test-invite@example.com", Role: "member",
	})
	if err != nil {
		fail("InviteMember(unregistered)", err)
		return
	}
	if inv.InvitationId == "" {
		fail("InviteMember(unregistered) empty invitation_id", context.Canceled)
		return
	}
	pass("InviteMember(unregistered)", "invitation_id="+inv.InvitationId)

	// 3. Invite registered email → added directly, no invitation id
	inv2, err := client.InviteMember(ownerCtx, &projectv1.InviteMemberRequest{
		ProjectId: projectID, Email: "fnaticritesh2004@gmail.com", Role: "member",
	})
	if err != nil {
		fail("InviteMember(registered)", err)
		return
	}
	if inv2.InvitationId != "" {
		fail("InviteMember(registered) expected empty invitation id, got "+inv2.InvitationId, context.Canceled)
		return
	}
	pass("InviteMember(registered) added directly")

	// 4. Invite same registered email again → AlreadyExists
	_, err = client.InviteMember(ownerCtx, &projectv1.InviteMemberRequest{
		ProjectId: projectID, Email: "fnaticritesh2004@gmail.com", Role: "viewer",
	})
	if status.Code(err) != 6 {
		fail("InviteMember(duplicate)", err)
	} else {
		pass("InviteMember(duplicate) → AlreadyExists")
	}

	// 5. Invalid role → InvalidArgument
	_, err = client.InviteMember(ownerCtx, &projectv1.InviteMemberRequest{
		ProjectId: projectID, Email: "x@example.com", Role: "developer",
	})
	if status.Code(err) != 3 {
		fail("InviteMember(bad-role)", err)
	} else {
		pass("InviteMember(bad-role developer) → InvalidArgument")
	}

	// 6. Invited user (unregistered, test token is the invitation token):
	//    other user accepts a bogus token → NotFound
	_, err = client.AcceptInvitation(otherCtx, &projectv1.AcceptInvitationRequest{Token: "bogus"})
	if status.Code(err) != 5 {
		fail("AcceptInvitation(bogus token)", err)
	} else {
		pass("AcceptInvitation(bogus) → NotFound")
	}

	// 7. Accept with the real token (passed via INVITE_TOKEN env)
	if real := os.Getenv("INVITE_TOKEN"); real != "" {
		accepted, err := client.AcceptInvitation(otherCtx, &projectv1.AcceptInvitationRequest{Token: real})
		if err != nil {
			fail("AcceptInvitation(real token)", err)
		} else {
			pass("AcceptInvitation(real token)", "project="+accepted.ProjectId, "name="+accepted.ProjectName)
		}

		// 8. Reuse the same token → FailedPrecondition (already used)
		_, err = client.AcceptInvitation(otherCtx, &projectv1.AcceptInvitationRequest{Token: real})
		if status.Code(err) != 9 {
			fail("AcceptInvitation(reuse token)", err)
		} else {
			pass("AcceptInvitation(reuse) → FailedPrecondition")
		}
	}

	log.Printf("projectID=%s", projectID)
	log.Printf("DONE")
}
