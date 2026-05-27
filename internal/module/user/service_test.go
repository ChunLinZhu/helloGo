// Package user — 用户 service 层单元测试（mock Repository）
package user

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"helloGo/internal/pkg/pagination"
)

// ── Mock Repository ────────────────────────────────────────

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) Create(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *mockRepo) FindByID(id string) (*User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *mockRepo) FindByUsername(username string) (*User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *mockRepo) Update(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *mockRepo) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockRepo) List(page pagination.Pagination, keyword string) ([]User, int64, error) {
	args := m.Called(page, keyword)
	return args.Get(0).([]User), args.Get(1).(int64), args.Error(2)
}

func (m *mockRepo) AssociateRoles(userID string, roleIDs []string) error {
	args := m.Called(userID, roleIDs)
	return args.Error(0)
}

// ── 测试辅助 ───────────────────────────────────────────────

func newTestService(repo *mockRepo) Service {
	logger, _ := zap.NewDevelopment()
	return NewService(repo, logger)
}

// ── Create 测试 ────────────────────────────────────────────

func TestCreate_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := newTestService(repo)

	// FindByUsername 返回 nil → 用户名不存在
	repo.On("FindByUsername", "alice").Return(nil, fmt.Errorf("not found"))

	// Create 成功
	repo.On("Create", mock.AnythingOfType("*user.User")).Return(nil)

	// FindByID 返回创建后的用户
	repo.On("FindByID", mock.AnythingOfType("string")).Return(&User{
		ID:       "uuid-1",
		Username: "alice",
		IsActive: true,
	}, nil)

	resp, err := svc.Create(&CreateUserRequest{
		Username: "alice",
		Password: "password123",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "alice", resp.Username)
	assert.Equal(t, "uuid-1", resp.ID)
	assert.True(t, resp.IsActive)
	repo.AssertExpectations(t)
}

func TestCreate_DuplicateUsername(t *testing.T) {
	repo := new(mockRepo)
	svc := newTestService(repo)

	// FindByUsername 返回已存在的用户
	repo.On("FindByUsername", "alice").Return(&User{
		ID:       "existing-uuid",
		Username: "alice",
	}, nil)

	resp, err := svc.Create(&CreateUserRequest{
		Username: "alice",
		Password: "password123",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "用户名已存在")
	repo.AssertExpectations(t)
}

// ── GetByID 测试 ───────────────────────────────────────────

func TestGetByID_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := newTestService(repo)

	repo.On("FindByID", "uuid-1").Return(&User{
		ID:       "uuid-1",
		Username: "alice",
		IsActive: true,
	}, nil)

	resp, err := svc.GetByID("uuid-1")

	assert.NoError(t, err)
	assert.Equal(t, "alice", resp.Username)
	repo.AssertExpectations(t)
}

func TestGetByID_NotFound(t *testing.T) {
	repo := new(mockRepo)
	svc := newTestService(repo)

	repo.On("FindByID", "bad-id").Return(nil, fmt.Errorf("record not found"))

	resp, err := svc.GetByID("bad-id")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "用户不存在")
	repo.AssertExpectations(t)
}

// ── Update 测试 ────────────────────────────────────────────

func TestUpdate_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := newTestService(repo)

	existingUser := &User{
		ID:       "uuid-1",
		Username: "alice",
		IsActive: true,
	}

	repo.On("FindByID", "uuid-1").Return(existingUser, nil).Times(2)
	repo.On("Update", mock.AnythingOfType("*user.User")).Return(nil)

	newEmail := "alice@example.com"
	resp, err := svc.Update("uuid-1", &UpdateUserRequest{
		Email: &newEmail,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "alice", resp.Username)
	repo.AssertExpectations(t)
}

func TestUpdate_NotFound(t *testing.T) {
	repo := new(mockRepo)
	svc := newTestService(repo)

	repo.On("FindByID", "bad-id").Return(nil, fmt.Errorf("record not found"))

	resp, err := svc.Update("bad-id", &UpdateUserRequest{})

	assert.Error(t, err)
	assert.Nil(t, resp)
	repo.AssertExpectations(t)
}

// ── Delete 测试 ────────────────────────────────────────────

func TestDelete_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := newTestService(repo)

	repo.On("FindByID", "uuid-1").Return(&User{
		ID:       "uuid-1",
		Username: "alice",
	}, nil)
	repo.On("Delete", "uuid-1").Return(nil)

	err := svc.Delete("uuid-1")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDelete_NotFound(t *testing.T) {
	repo := new(mockRepo)
	svc := newTestService(repo)

	repo.On("FindByID", "bad-id").Return(nil, fmt.Errorf("record not found"))

	err := svc.Delete("bad-id")

	assert.Error(t, err)
	repo.AssertExpectations(t)
}

// ── List 测试 ──────────────────────────────────────────────

func TestList_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := newTestService(repo)

	users := []User{
		{ID: "1", Username: "alice", IsActive: true},
		{ID: "2", Username: "bob", IsActive: true},
	}

	p := pagination.Pagination{Page: 1, Limit: 10}
	repo.On("List", p, "").Return(users, int64(2), nil)

	resp, total, err := svc.List(p, "")

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, resp, 2)
	assert.Equal(t, "alice", resp[0].Username)
	assert.Equal(t, "bob", resp[1].Username)
	repo.AssertExpectations(t)
}

func TestList_RepoError(t *testing.T) {
	repo := new(mockRepo)
	svc := newTestService(repo)

	p := pagination.Pagination{Page: 1, Limit: 10}
	repo.On("List", p, "").Return([]User{}, int64(0), fmt.Errorf("db error"))

	resp, total, err := svc.List(p, "")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, int64(0), total)
	repo.AssertExpectations(t)
}
