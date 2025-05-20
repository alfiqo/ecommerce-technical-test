package repository

import (
	"log"
	"reflect"
	"testing"
	"time"
	"user-service/internal/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestNewUserRepository(t *testing.T) {
	type args struct {
		log *logrus.Logger
		db  *gorm.DB
	}
	tests := []struct {
		name string
		args args
		want UserRepositoryInterface
	}{
		{
			name: "success",
			args: args{
				log: nil,
				db:  nil,
			},
			want: &UserRepository{
				DB:  nil,
				Log: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewUserRepository(tt.args.log, tt.args.db); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUserRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserRepository_Create(t *testing.T) {

	// uuid := uuid.New()
	// Initialize mock database
	mockDb, mock, _ := sqlmock.New()

	// Add the expected query for SELECT VERSION()
	mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("8.0.28"))

	// Proceed with the GORM setup
	dialector := mysql.New(mysql.Config{
		Conn:       mockDb,
		DriverName: "mysql",
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatal("Error opening DB connection: ", err)
	}

	type fields struct {
		DB  *gorm.DB
		Log *logrus.Logger
	}
	type args struct {
		db   *gorm.DB
		user *entity.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "failed",
			fields: fields{
				DB:  db,
				Log: logrus.New(),
			},
			args: args{
				db: db,
				user: &entity.User{
					ID:        uuid.New(),
					Name:      "Test User",
					Email:     "nK4e0@example.com",
					Phone:     "1234567890",
					Password:  "hashedpassword",
					Token:     "sometoken",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &UserRepository{
				DB:  tt.fields.DB,
				Log: tt.fields.Log,
			}
			if err := r.Create(tt.args.db, tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("UserRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	// Initialize mock database
	mockDb, mock, _ := sqlmock.New()

	// Add the expected query for SELECT VERSION()
	mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("8.0.28"))

	// Proceed with the GORM setup
	dialector := mysql.New(mysql.Config{
		Conn:       mockDb,
		DriverName: "mysql",
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatal("Error opening DB connection: ", err)
	}

	userID := uuid.New()
	testEmail := "test@example.com"
	testUser := &entity.User{
		ID:        userID,
		Name:      "Test User",
		Email:     testEmail,
		Phone:     "1234567890",
		Password:  "hashedpassword",
		Token:     "sometoken",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test cases
	type fields struct {
		DB  *gorm.DB
		Log *logrus.Logger
	}
	type args struct {
		db    *gorm.DB
		user  *entity.User
		email string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mockFn  func()
		wantErr bool
	}{
		{
			name: "success_find_by_email",
			fields: fields{
				DB:  db,
				Log: logrus.New(),
			},
			args: args{
				db:    db,
				user:  &entity.User{},
				email: testEmail,
			},
			mockFn: func() {
				rows := sqlmock.NewRows([]string{"uuid", "name", "email", "phone", "password", "token", "created_at", "updated_at"}).
					AddRow(userID.String(), testUser.Name, testUser.Email, testUser.Phone, testUser.Password, testUser.Token, testUser.CreatedAt, testUser.UpdatedAt)
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE email = \\? LIMIT \\?").
					WithArgs(testEmail, 1).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "failed_find_by_email_not_found",
			fields: fields{
				DB:  db,
				Log: logrus.New(),
			},
			args: args{
				db:    db,
				user:  &entity.User{},
				email: "nonexistent@example.com",
			},
			mockFn: func() {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE email = \\? LIMIT \\?").
					WithArgs("nonexistent@example.com", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			tt.mockFn()

			r := &UserRepository{
				DB:  tt.fields.DB,
				Log: tt.fields.Log,
			}
			err := r.FindByEmail(tt.args.db, tt.args.user, tt.args.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserRepository.FindByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Additional validation for success case
			if tt.name == "success_find_by_email" && err == nil {
				if tt.args.user.Email != testEmail {
					t.Errorf("UserRepository.FindByEmail() got user email = %v, want %v", tt.args.user.Email, testEmail)
				}
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	// Initialize mock database
	mockDb, mock, _ := sqlmock.New()

	// Add the expected query for SELECT VERSION()
	mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("8.0.28"))

	// Proceed with the GORM setup
	dialector := mysql.New(mysql.Config{
		Conn:       mockDb,
		DriverName: "mysql",
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatal("Error opening DB connection: ", err)
	}

	userID := uuid.New()
	testUser := &entity.User{
		ID:        userID,
		Name:      "Updated User",
		Email:     "updated@example.com",
		Phone:     "9876543210",
		Password:  "updatedpassword",
		Token:     "updatedtoken",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	type fields struct {
		DB  *gorm.DB
		Log *logrus.Logger
	}
	type args struct {
		db   *gorm.DB
		user *entity.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mockFn  func()
		wantErr bool
	}{
		{
			name: "success_update",
			fields: fields{
				DB:  db,
				Log: logrus.New(),
			},
			args: args{
				db:   db,
				user: testUser,
			},
			mockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `users` SET").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "failed_update",
			fields: fields{
				DB:  db,
				Log: logrus.New(),
			},
			args: args{
				db:   db,
				user: testUser,
			},
			mockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `users` SET").
					WillReturnError(gorm.ErrInvalidData)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			tt.mockFn()

			r := &UserRepository{
				DB:  tt.fields.DB,
				Log: tt.fields.Log,
			}
			err := r.Update(tt.args.db, tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserRepository_FindByToken(t *testing.T) {
	// Initialize mock database
	mockDb, mock, _ := sqlmock.New()

	// Add the expected query for SELECT VERSION()
	mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("8.0.28"))

	// Proceed with the GORM setup
	dialector := mysql.New(mysql.Config{
		Conn:       mockDb,
		DriverName: "mysql",
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatal("Error opening DB connection: ", err)
	}

	userID := uuid.New()
	testToken := "test-token-123456"
	testUser := &entity.User{
		ID:        userID,
		Name:      "Test User",
		Email:     "token-test@example.com",
		Phone:     "1234567890",
		Password:  "hashedpassword",
		Token:     testToken,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	type fields struct {
		DB  *gorm.DB
		Log *logrus.Logger
	}
	type args struct {
		db    *gorm.DB
		token string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mockFn  func()
		want    *entity.User
		wantErr bool
	}{
		{
			name: "success_find_by_token",
			fields: fields{
				DB:  db,
				Log: logrus.New(),
			},
			args: args{
				db:    db,
				token: testToken,
			},
			mockFn: func() {
				rows := sqlmock.NewRows([]string{"uuid", "name", "email", "phone", "password", "token", "created_at", "updated_at"}).
					AddRow(userID.String(), testUser.Name, testUser.Email, testUser.Phone, testUser.Password, testUser.Token, testUser.CreatedAt, testUser.UpdatedAt)
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE token = \\? ORDER BY `users`.`uuid` LIMIT \\?").
					WithArgs(testToken, 1).
					WillReturnRows(rows)
			},
			want:    testUser,
			wantErr: false,
		},
		{
			name: "failed_find_by_token_not_found",
			fields: fields{
				DB:  db,
				Log: logrus.New(),
			},
			args: args{
				db:    db,
				token: "nonexistent-token",
			},
			mockFn: func() {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE token = \\? ORDER BY `users`.`uuid` LIMIT \\?").
					WithArgs("nonexistent-token", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			tt.mockFn()

			r := &UserRepository{
				DB:  tt.fields.DB,
				Log: tt.fields.Log,
			}
			got, err := r.FindByToken(tt.args.db, tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserRepository.FindByToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// For the success case, check if token matches
			if !tt.wantErr && got.Token != tt.args.token {
				t.Errorf("UserRepository.FindByToken() got = %v, want token = %v", got.Token, tt.args.token)
			}
			
			// For failure case, check if we got nil
			if tt.wantErr && got != nil {
				t.Errorf("UserRepository.FindByToken() got = %v, want nil", got)
			}
		})
	}
}
