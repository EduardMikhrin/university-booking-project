package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserTestDB(t *testing.T) (*UserQ, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	sqlxDB := sqlx.NewDb(db, "postgres")
	userQ := NewUserQ(sqlxDB).(*UserQ)

	teardown := func() {
		db.Close()
	}

	return userQ, mock, teardown
}

func TestUserQ_Create(t *testing.T) {
	tests := []struct {
		name    string
		user    *types.User
		mock    func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful create with all fields",
			user: &types.User{
				ID:        uuid.New(),
				Email:     "test@example.com",
				Password:  "hashedpassword",
				Name:      "Test User",
				Phone:     stringPtr("+1234567890"),
				Photo:     stringPtr("https://example.com/photo.jpg"),
				Role:      "user",
				CreatedAt: time.Now(),
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO users`).
					WithArgs(
						sqlmock.AnyArg(), // id
						"test@example.com",
						"hashedpassword",
						"Test User",
						"+1234567890",
						"https://example.com/photo.jpg",
						"user",
						sqlmock.AnyArg(), // created_at
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "successful create with default photo",
			user: &types.User{
				ID:        uuid.New(),
				Email:     "test@example.com",
				Password:  "hashedpassword",
				Name:      "Test User",
				Phone:     nil,
				Photo:     nil,
				Role:      "user",
				CreatedAt: time.Now(),
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO users`).
					WithArgs(
						sqlmock.AnyArg(), // id
						"test@example.com",
						"hashedpassword",
						"Test User",
						nil,                    // phone
						types.DefaultUserPhoto, // default photo
						"user",
						sqlmock.AnyArg(), // created_at
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "successful create with auto-generated ID",
			user: &types.User{
				ID:        uuid.Nil,
				Email:     "test@example.com",
				Password:  "hashedpassword",
				Name:      "Test User",
				Role:      "user",
				CreatedAt: time.Now(),
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO users`).
					WithArgs(
						sqlmock.AnyArg(), // id (will be generated)
						"test@example.com",
						"hashedpassword",
						"Test User",
						nil,                    // phone
						types.DefaultUserPhoto, // default photo
						"user",
						sqlmock.AnyArg(), // created_at
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "database error",
			user: &types.User{
				ID:       uuid.New(),
				Email:    "test@example.com",
				Password: "hashedpassword",
				Name:     "Test User",
				Role:     "user",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO users`).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userQ, mock, teardown := setupUserTestDB(t)
			defer teardown()

			tt.mock(mock)

			ctx := context.Background()
			err := userQ.Create(ctx, tt.user)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify that ID was generated if it was nil
				if tt.user.ID == uuid.Nil {
					assert.NotEqual(t, uuid.Nil, tt.user.ID)
				}
				// Verify default photo was set if it was nil
				if tt.user.Photo == nil {
					assert.NotNil(t, tt.user.Photo)
					assert.Equal(t, types.DefaultUserPhoto, *tt.user.Photo)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserQ_GetByID(t *testing.T) {
	userID := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name    string
		id      uuid.UUID
		mock    func(mock sqlmock.Sqlmock)
		want    *types.User
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful get",
			id:   userID,
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "password", "name", "phone", "photo", "role", "created_at"}).
					AddRow(userID, "test@example.com", "hashedpassword", "Test User", "+1234567890", "https://example.com/photo.jpg", "user", createdAt)
				mock.ExpectQuery(`SELECT id, email, password, name, phone, photo, role, created_at FROM users WHERE id = \$1`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			want: &types.User{
				ID:        userID,
				Email:     "test@example.com",
				Password:  "hashedpassword",
				Name:      "Test User",
				Phone:     stringPtr("+1234567890"),
				Photo:     stringPtr("https://example.com/photo.jpg"),
				Role:      "user",
				CreatedAt: createdAt,
			},
			wantErr: false,
		},
		{
			name: "user not found",
			id:   userID,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, email, password, name, phone, photo, role, created_at FROM users WHERE id = \$1`).
					WithArgs(userID).
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "user not found",
		},
		{
			name: "database error",
			id:   userID,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, email, password, name, phone, photo, role, created_at FROM users WHERE id = \$1`).
					WithArgs(userID).
					WillReturnError(sql.ErrConnDone)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "user with default photo",
			id:   userID,
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "password", "name", "phone", "photo", "role", "created_at"}).
					AddRow(userID, "test@example.com", "hashedpassword", "Test User", nil, nil, "user", createdAt)
				mock.ExpectQuery(`SELECT id, email, password, name, phone, photo, role, created_at FROM users WHERE id = \$1`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			want: &types.User{
				ID:        userID,
				Email:     "test@example.com",
				Password:  "hashedpassword",
				Name:      "Test User",
				Phone:     nil,
				Photo:     stringPtr(types.DefaultUserPhoto),
				Role:      "user",
				CreatedAt: createdAt,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userQ, mock, teardown := setupUserTestDB(t)
			defer teardown()

			tt.mock(mock)

			ctx := context.Background()
			got, err := userQ.GetByID(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.EqualError(t, err, tt.errMsg)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.Email, got.Email)
				assert.Equal(t, tt.want.Name, got.Name)
				assert.Equal(t, tt.want.Role, got.Role)
				if tt.want.Photo != nil {
					assert.NotNil(t, got.Photo)
					assert.Equal(t, *tt.want.Photo, *got.Photo)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserQ_GetByEmail(t *testing.T) {
	userID := uuid.New()
	createdAt := time.Now()
	email := "test@example.com"

	tests := []struct {
		name    string
		email   string
		mock    func(mock sqlmock.Sqlmock)
		want    *types.User
		wantErr bool
		errMsg  string
	}{
		{
			name:  "successful get",
			email: email,
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "password", "name", "phone", "photo", "role", "created_at"}).
					AddRow(userID, email, "hashedpassword", "Test User", "+1234567890", "https://example.com/photo.jpg", "user", createdAt)
				mock.ExpectQuery(`SELECT id, email, password, name, phone, photo, role, created_at FROM users WHERE email = \$1`).
					WithArgs(email).
					WillReturnRows(rows)
			},
			want: &types.User{
				ID:        userID,
				Email:     email,
				Password:  "hashedpassword",
				Name:      "Test User",
				Phone:     stringPtr("+1234567890"),
				Photo:     stringPtr("https://example.com/photo.jpg"),
				Role:      "user",
				CreatedAt: createdAt,
			},
			wantErr: false,
		},
		{
			name:  "user not found",
			email: email,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, email, password, name, phone, photo, role, created_at FROM users WHERE email = \$1`).
					WithArgs(email).
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userQ, mock, teardown := setupUserTestDB(t)
			defer teardown()

			tt.mock(mock)

			ctx := context.Background()
			got, err := userQ.GetByEmail(ctx, tt.email)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.EqualError(t, err, tt.errMsg)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.want.Email, got.Email)
				assert.Equal(t, tt.want.Name, got.Name)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserQ_Update(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		user    *types.User
		mock    func(mock sqlmock.Sqlmock)
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful update",
			id:   userID,
			user: &types.User{
				Email: "updated@example.com",
				Name:  "Updated User",
				Phone: stringPtr("+9876543210"),
				Photo: stringPtr("https://example.com/new-photo.jpg"),
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users SET email = \$1, name = \$2, phone = \$3, photo = \$4 WHERE id = \$5`).
					WithArgs(
						"updated@example.com",
						"Updated User",
						stringPtr("+9876543210"),
						stringPtr("https://example.com/new-photo.jpg"),
						userID,
					).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "user not found",
			id:   userID,
			user: &types.User{
				Email: "updated@example.com",
				Name:  "Updated User",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users SET email = \$1, name = \$2, phone = \$3, photo = \$4 WHERE id = \$5`).
					WithArgs(
						"updated@example.com",
						"Updated User",
						nil,
						nil,
						userID,
					).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
			errMsg:  "user not found",
		},
		{
			name: "database error",
			id:   userID,
			user: &types.User{
				Email: "updated@example.com",
				Name:  "Updated User",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users SET email = \$1, name = \$2, phone = \$3, photo = \$4 WHERE id = \$5`).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userQ, mock, teardown := setupUserTestDB(t)
			defer teardown()

			tt.mock(mock)

			ctx := context.Background()
			err := userQ.Update(ctx, tt.id, tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.EqualError(t, err, tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
