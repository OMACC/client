package teams

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/keybase/client/go/kbtest"
	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/go/protocol/keybase1"
)

func memberSetup(t *testing.T) (libkb.TestContext, *kbtest.FakeUser, string) {
	tc := libkb.SetupTest(t, "team", 1)
	tc.Tp.UpgradePerUserKey = true

	u, err := kbtest.CreateAndSignupFakeUser("team", tc.G)
	if err != nil {
		t.Fatal(err)
	}

	name := createTeam(tc)

	return tc, u, name
}

func memberSetupMultiple(t *testing.T) (tc libkb.TestContext, owner, other *kbtest.FakeUser, name string) {
	tc = libkb.SetupTest(t, "team", 1)
	tc.Tp.UpgradePerUserKey = true

	other, err := kbtest.CreateAndSignupFakeUser("team", tc.G)
	if err != nil {
		t.Fatal(err)
	}
	tc.G.Logout()

	owner, err = kbtest.CreateAndSignupFakeUser("team", tc.G)
	if err != nil {
		t.Fatal(err)
	}

	name = createTeam(tc)

	return tc, owner, other, name
}

func TestMemberOwner(t *testing.T) {
	tc, u, name := memberSetup(t)
	defer tc.Cleanup()

	assertRole(tc, name, u.Username, keybase1.TeamRole_OWNER)
	assertRole(tc, name, "t_alice", keybase1.TeamRole_NONE)
}

type setRoleTest struct {
	name        string
	setRoleFunc func(ctx context.Context, g *libkb.GlobalContext, teamname, username string) error
	afterRole   keybase1.TeamRole
}

var setRoleTests = []setRoleTest{
	setRoleTest{name: "owner", setRoleFunc: SetRoleOwner, afterRole: keybase1.TeamRole_OWNER},
	setRoleTest{name: "admin", setRoleFunc: SetRoleAdmin, afterRole: keybase1.TeamRole_ADMIN},
	setRoleTest{name: "writer", setRoleFunc: SetRoleWriter, afterRole: keybase1.TeamRole_WRITER},
	setRoleTest{name: "reader", setRoleFunc: SetRoleReader, afterRole: keybase1.TeamRole_READER},
}

func TestMemberSetRole(t *testing.T) {
	for _, test := range setRoleTests {
		testMemberSetRole(t, test)
	}
}

func testMemberSetRole(t *testing.T, test setRoleTest) {
	tc, owner, other, name := memberSetupMultiple(t)
	defer tc.Cleanup()

	if err := test.setRoleFunc(context.TODO(), tc.G, name, other.Username); err != nil {
		t.Fatal(err)
	}

	assertRole(tc, name, owner.Username, keybase1.TeamRole_OWNER)
	assertRole(tc, name, other.Username, test.afterRole)
}

func TestMemberAdd(t *testing.T) {
	tc, _, other, name := memberSetupMultiple(t)
	defer tc.Cleanup()

	assertRole(tc, name, other.Username, keybase1.TeamRole_NONE)

	if err := AddMember(context.TODO(), tc.G, name, other.Username, keybase1.TeamRole_READER); err != nil {
		t.Fatal(err)
	}

	assertRole(tc, name, other.Username, keybase1.TeamRole_READER)

	// second AddMember should return err
	if err := AddMember(context.TODO(), tc.G, name, other.Username, keybase1.TeamRole_WRITER); err == nil {
		t.Errorf("second AddMember succeeded, should have failed since user already a member")
	}

	assertRole(tc, name, other.Username, keybase1.TeamRole_READER)
}

func TestMemberAddInvalidRole(t *testing.T) {
	tc, _, other, name := memberSetupMultiple(t)
	defer tc.Cleanup()

	if err := AddMember(context.TODO(), tc.G, name, other.Username, keybase1.TeamRole(8888)); err == nil {
		t.Errorf("AddMember worked with invalid role")
	}

	assertRole(tc, name, other.Username, keybase1.TeamRole_NONE)
}

func TestMemberRemove(t *testing.T) {
	tc, owner, other, name := memberSetupMultiple(t)
	defer tc.Cleanup()

	if err := SetRoleWriter(context.TODO(), tc.G, name, other.Username); err != nil {
		t.Fatal(err)
	}

	assertRole(tc, name, owner.Username, keybase1.TeamRole_OWNER)
	assertRole(tc, name, other.Username, keybase1.TeamRole_WRITER)

	if err := RemoveMember(context.TODO(), tc.G, name, other.Username); err != nil {
		t.Fatal(err)
	}

	assertRole(tc, name, owner.Username, keybase1.TeamRole_OWNER)
	assertRole(tc, name, other.Username, keybase1.TeamRole_NONE)
}

func TestMemberChangeRole(t *testing.T) {
	tc, owner, other, name := memberSetupMultiple(t)
	defer tc.Cleanup()

	if err := SetRoleWriter(context.TODO(), tc.G, name, other.Username); err != nil {
		t.Fatal(err)
	}

	assertRole(tc, name, owner.Username, keybase1.TeamRole_OWNER)
	assertRole(tc, name, other.Username, keybase1.TeamRole_WRITER)

	if err := SetRoleReader(context.TODO(), tc.G, name, other.Username); err != nil {
		t.Fatal(err)
	}

	assertRole(tc, name, owner.Username, keybase1.TeamRole_OWNER)
	assertRole(tc, name, other.Username, keybase1.TeamRole_READER)
}

func assertRole(tc libkb.TestContext, name, username string, expected keybase1.TeamRole) {
	role, err := MemberRole(context.TODO(), tc.G, name, username)
	if err != nil {
		tc.T.Fatal(err)
	}
	if role != expected {
		tc.T.Errorf("role: %s, expected %s", role, expected)
	}
}
