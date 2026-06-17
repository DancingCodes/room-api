package repository

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"room-api/internal/model"
)

func TestRoomRepositoryIntegration(t *testing.T) {
	db := openIntegrationDB(t)
	prefix := testPrefix()
	cleanupIntegrationData(t, db, prefix)
	t.Cleanup(func() { cleanupIntegrationData(t, db, prefix) })

	users := NewUserRepository(db)
	rooms := NewRoomRepository(db)

	owner := createIntegrationUser(t, users, prefix, "owner")
	guest := createIntegrationUser(t, users, prefix, "guest")

	room, members, err := rooms.Create(owner, 2)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if room.ID == 0 || room.OwnerID != owner.ID || room.MaxMembers != 2 {
		t.Fatalf("Create() room = %+v", room)
	}
	if len(members) != 1 || members[0].UserID != owner.ID || !members[0].IsOwner || members[0].JoinedAt.IsZero() {
		t.Fatalf("Create() members = %+v", members)
	}

	if _, _, err := rooms.Create(owner, 2); err == nil || !strings.Contains(err.Error(), "用户已在房间内") {
		t.Fatalf("Create() duplicate owner error = %v, want 用户已在房间内", err)
	}

	joinedRoom, joinedMembers, err := rooms.Join(room.ID, guest.ID)
	if err != nil {
		t.Fatalf("Join() error = %v", err)
	}
	if joinedRoom.ID != room.ID || len(joinedMembers) != 2 {
		t.Fatalf("Join() room = %+v members = %+v", joinedRoom, joinedMembers)
	}

	isMember, err := rooms.IsMember(room.ID, guest.ID)
	if err != nil {
		t.Fatalf("IsMember() error = %v", err)
	}
	if !isMember {
		t.Fatal("IsMember() = false, want true")
	}

	micMember, err := rooms.UpdateMicStatus(room.ID, guest.ID, "on")
	if err != nil {
		t.Fatalf("UpdateMicStatus() error = %v", err)
	}
	if micMember.MicStatus != "on" {
		t.Fatalf("UpdateMicStatus() = %q, want on", micMember.MicStatus)
	}

	leaveOwner, err := rooms.Leave(room.ID, owner.ID)
	if err != nil {
		t.Fatalf("Leave(owner) error = %v", err)
	}
	if !leaveOwner.Left || !leaveOwner.OwnerChanged || leaveOwner.NewOwnerUserID != guest.ID || leaveOwner.DeletedRoom {
		t.Fatalf("Leave(owner) = %+v", leaveOwner)
	}

	detailRoom, detailMembers, err := rooms.Detail(room.ID)
	if err != nil {
		t.Fatalf("Detail() error = %v", err)
	}
	if detailRoom.OwnerID != guest.ID || len(detailMembers) != 1 || !detailMembers[0].IsOwner {
		t.Fatalf("Detail() room = %+v members = %+v", detailRoom, detailMembers)
	}

	leaveGuest, err := rooms.Leave(room.ID, guest.ID)
	if err != nil {
		t.Fatalf("Leave(guest) error = %v", err)
	}
	if !leaveGuest.Left || !leaveGuest.DeletedRoom || leaveGuest.CurrentMemberSize != 0 {
		t.Fatalf("Leave(guest) = %+v", leaveGuest)
	}

	if _, err := rooms.FindRoom(room.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("FindRoom(deleted) error = %v, want record not found", err)
	}
}

func TestMessageRepositoryIntegration(t *testing.T) {
	db := openIntegrationDB(t)
	prefix := testPrefix()
	cleanupIntegrationData(t, db, prefix)
	t.Cleanup(func() { cleanupIntegrationData(t, db, prefix) })

	users := NewUserRepository(db)
	rooms := NewRoomRepository(db)
	messages := NewMessageRepository(db)

	owner := createIntegrationUser(t, users, prefix, "owner")
	room, _, err := rooms.Create(owner, 8)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	for i := 1; i <= 3; i++ {
		message := &model.Message{
			RoomID:    room.ID,
			SenderID:  owner.ID,
			Type:      "text",
			Content:   fmt.Sprintf("message-%d", i),
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
		}
		if err := messages.Create(message); err != nil {
			t.Fatalf("Message Create(%d) error = %v", i, err)
		}
	}

	list, err := messages.List(room.ID, 2, 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("List() length = %d, want 2", len(list))
	}
	if list[0].Content != "message-2" || list[1].Content != "message-3" {
		t.Fatalf("List() contents = [%q, %q], want [message-2, message-3]", list[0].Content, list[1].Content)
	}

	beforeID := list[0].ID
	older, err := messages.List(room.ID, 2, beforeID)
	if err != nil {
		t.Fatalf("List(beforeID) error = %v", err)
	}
	if len(older) != 1 || older[0].Content != "message-1" {
		t.Fatalf("List(beforeID) = %+v, want message-1", older)
	}
}

func openIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()

	_ = godotenv.Load("../../.env")

	dsn := os.Getenv("ROOM_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set ROOM_TEST_MYSQL_DSN to run repository integration tests")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	return db
}

func createIntegrationUser(t *testing.T, users *UserRepository, prefix string, role string) *model.User {
	t.Helper()

	user := &model.User{
		Username:     prefix + "_" + role,
		Email:        prefix + "_" + role + "@example.com",
		Nickname:     role[:3] + prefix[len(prefix)-4:],
		PasswordHash: "hash",
		AvatarURL:    "https://example.com/avatar.png",
	}
	if err := users.Create(user); err != nil {
		t.Fatalf("create user %s: %v", role, err)
	}
	return user
}

func cleanupIntegrationData(t *testing.T, db *gorm.DB, prefix string) {
	t.Helper()

	var userIDs []uint64
	if err := db.Model(&model.User{}).
		Where("username LIKE ?", prefix+"_%").
		Pluck("id", &userIDs).Error; err != nil {
		t.Fatalf("find integration users: %v", err)
	}
	if len(userIDs) == 0 {
		return
	}

	var roomIDs []uint64
	if err := db.Model(&model.Room{}).
		Where("owner_id IN ?", userIDs).
		Pluck("id", &roomIDs).Error; err != nil {
		t.Fatalf("find integration rooms: %v", err)
	}

	if len(roomIDs) > 0 {
		if err := db.Delete(&model.Message{}, "room_id IN ?", roomIDs).Error; err != nil {
			t.Fatalf("cleanup messages: %v", err)
		}
		if err := db.Delete(&model.RoomMember{}, "room_id IN ?", roomIDs).Error; err != nil {
			t.Fatalf("cleanup room members by room: %v", err)
		}
		if err := db.Delete(&model.Room{}, "id IN ?", roomIDs).Error; err != nil {
			t.Fatalf("cleanup rooms: %v", err)
		}
	}

	if err := db.Delete(&model.RoomMember{}, "user_id IN ?", userIDs).Error; err != nil {
		t.Fatalf("cleanup room members by user: %v", err)
	}
	if err := db.Delete(&model.User{}, "id IN ?", userIDs).Error; err != nil {
		t.Fatalf("cleanup users: %v", err)
	}
}

func testPrefix() string {
	return fmt.Sprintf("it%08d", time.Now().UnixNano()%100000000)
}
