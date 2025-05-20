package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"testing"
	"user-service/internal/entity"
	"user-service/internal/model"
	"user-service/internal/repository"
	repository_mock "user-service/mocks/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestNewUserUseCase(t *testing.T) {
	newLogrus := logrus.New()
	newValidator := validator.New()
	type args struct {
		db             *gorm.DB
		logger         *logrus.Logger
		validate       *validator.Validate
		userRepository repository.UserRepositoryInterface
	}
	tests := []struct {
		name string
		args args
		want *UserUseCase
	}{
		{
			name: "success",
			args: args{
				db:             nil,
				logger:         newLogrus,
				validate:       newValidator,
				userRepository: nil,
			},
			want: &UserUseCase{
				DB:             nil,
				Log:            newLogrus,
				Validate:       newValidator,
				UserRepository: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewUserUseCase(tt.args.db, tt.args.logger, tt.args.validate, tt.args.userRepository)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUserUseCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserUseCase_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Initialize mock database
	mockDb, mock, _ := sqlmock.New()

	// Add the expected query for SELECT VERSION()
	mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("8.0.28"))

	// Expect transaction begin
	mock.ExpectBegin()

	// Optionally, expect commit or rollback depending on your code
	mock.ExpectCommit()

	mock.ExpectCommit().WillReturnError(fmt.Errorf("failed to commit transaction"))

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
		DB             *gorm.DB
		Log            *logrus.Logger
		Validate       *validator.Validate
		UserRepository repository.UserRepositoryInterface
	}
	type args struct {
		ctx     context.Context
		request *model.RegisterUserRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.UserResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				DB:       db,
				Log:      logrus.New(),
				Validate: validator.New(),
				UserRepository: func() repository.UserRepositoryInterface {
					r := repository_mock.NewMockUserRepositoryInterface(ctrl)
					r.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
					return r
				}(),
			},
			args: args{
				ctx: context.TODO(),
				request: &model.RegisterUserRequest{
					Name:     "John Doe",
					Email:    "Fg1w2@example.com",
					Phone:    "08123456789",
					Password: "password123",
				},
			},
			want: &model.UserResponse{
				ID:    "generated-uuid",
				Name:  "John Doe",
				Email: "Fg1w2@example.com",
				Phone: "08123456789",
			},
			wantErr: false,
		},
		{
			name: "email invalid",
			fields: fields{
				DB:       db,
				Log:      logrus.New(),
				Validate: validator.New(),
			},
			args: args{
				ctx: context.TODO(),
				request: &model.RegisterUserRequest{
					Name:     "John Doe",
					Email:    "invalid-email",
					Phone:    "08123456789",
					Password: "password123",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error create user",
			fields: fields{
				DB:       db,
				Log:      logrus.New(),
				Validate: validator.New(),
				UserRepository: func() repository.UserRepositoryInterface {
					r := repository_mock.NewMockUserRepositoryInterface(ctrl)
					r.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("error create user"))
					return r
				}(),
			},
			args: args{
				ctx: context.TODO(),
				request: &model.RegisterUserRequest{
					Name:     "John Doe",
					Email:    "Fg1w2@example.com",
					Phone:    "08123456789",
					Password: "password123",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error on commit",
			fields: fields{
				DB:       db,
				Log:      logrus.New(),
				Validate: validator.New(),
				UserRepository: func() repository.UserRepositoryInterface {
					r := repository_mock.NewMockUserRepositoryInterface(ctrl)
					r.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
					return r
				}(),
			},
			args: args{
				ctx: context.TODO(),
				request: &model.RegisterUserRequest{
					Name:     "John Doe",
					Email:    "Fg1w2@example.com",
					Phone:    "08123456789",
					Password: "password123",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &UserUseCase{
				DB:             tt.fields.DB,
				Log:            tt.fields.Log,
				Validate:       tt.fields.Validate,
				UserRepository: tt.fields.UserRepository,
			}
			got, err := c.Create(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserUseCase.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && !reflect.DeepEqual(got.Email, tt.want.Email) {
				t.Errorf("UserUseCase.Create() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserUseCase_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	type fields struct {
		DB             *gorm.DB
		Log            *logrus.Logger
		Validate       *validator.Validate
		UserRepository repository.UserRepositoryInterface
	}
	type args struct {
		ctx     context.Context
		request *model.LoginUserRequest
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		mockExpect func(repo *repository_mock.MockUserRepositoryInterface)
		mockTx     func()
		want       *model.UserResponse
		wantErr    bool
	}{
		{
			name: "success_login",
			fields: fields{
				DB:       db,
				Log:      logrus.New(),
				Validate: validator.New(),
				UserRepository: func() repository.UserRepositoryInterface {
					repo := repository_mock.NewMockUserRepositoryInterface(ctrl)
					repo.EXPECT().FindByEmail(gomock.Any(), gomock.Any(), "user@example.com").DoAndReturn(
						func(db *gorm.DB, user *entity.User, email string) error {
							user.Email = "user@example.com"
							user.Password = string(hashedPassword)
							return nil
						})
					repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
					return repo
				}(),
			},
			args: args{
				ctx: context.TODO(),
				request: &model.LoginUserRequest{
					Email:    "user@example.com",
					Password: "password123",
				},
			},
			mockTx: func() {
				mock.ExpectBegin()
				mock.ExpectCommit()
			},
			want: &model.UserResponse{
				Token: "token-uuid", // We won't check the exact token value
			},
			wantErr: false,
		},
		{
			name: "invalid_request_validation_fails",
			fields: fields{
				DB:             db,
				Log:            logrus.New(),
				Validate:       validator.New(),
				UserRepository: repository_mock.NewMockUserRepositoryInterface(ctrl),
			},
			args: args{
				ctx: context.TODO(),
				request: &model.LoginUserRequest{
					Email:    "", // Invalid empty email
					Password: "password123",
				},
			},
			mockTx: func() {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "user_not_found",
			fields: fields{
				DB:       db,
				Log:      logrus.New(),
				Validate: validator.New(),
				UserRepository: func() repository.UserRepositoryInterface {
					repo := repository_mock.NewMockUserRepositoryInterface(ctrl)
					repo.EXPECT().FindByEmail(gomock.Any(), gomock.Any(), "nonexistent@example.com").Return(gorm.ErrRecordNotFound)
					return repo
				}(),
			},
			args: args{
				ctx: context.TODO(),
				request: &model.LoginUserRequest{
					Email:    "nonexistent@example.com",
					Password: "password123",
				},
			},
			mockTx: func() {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "incorrect_password",
			fields: fields{
				DB:       db,
				Log:      logrus.New(),
				Validate: validator.New(),
				UserRepository: func() repository.UserRepositoryInterface {
					repo := repository_mock.NewMockUserRepositoryInterface(ctrl)
					repo.EXPECT().FindByEmail(gomock.Any(), gomock.Any(), "user@example.com").DoAndReturn(
						func(db *gorm.DB, user *entity.User, email string) error {
							user.Email = "user@example.com"
							user.Password = string(hashedPassword)
							return nil
						})
					return repo
				}(),
			},
			args: args{
				ctx: context.TODO(),
				request: &model.LoginUserRequest{
					Email:    "user@example.com",
					Password: "wrong_password", // Incorrect password
				},
			},
			mockTx: func() {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "update_user_fails",
			fields: fields{
				DB:       db,
				Log:      logrus.New(),
				Validate: validator.New(),
				UserRepository: func() repository.UserRepositoryInterface {
					repo := repository_mock.NewMockUserRepositoryInterface(ctrl)
					repo.EXPECT().FindByEmail(gomock.Any(), gomock.Any(), "user@example.com").DoAndReturn(
						func(db *gorm.DB, user *entity.User, email string) error {
							user.Email = "user@example.com"
							user.Password = string(hashedPassword)
							return nil
						})
					repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("update failed"))
					return repo
				}(),
			},
			args: args{
				ctx: context.TODO(),
				request: &model.LoginUserRequest{
					Email:    "user@example.com",
					Password: "password123",
				},
			},
			mockTx: func() {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "commit_transaction_fails",
			fields: fields{
				DB:       db,
				Log:      logrus.New(),
				Validate: validator.New(),
				UserRepository: func() repository.UserRepositoryInterface {
					repo := repository_mock.NewMockUserRepositoryInterface(ctrl)
					repo.EXPECT().FindByEmail(gomock.Any(), gomock.Any(), "user@example.com").DoAndReturn(
						func(db *gorm.DB, user *entity.User, email string) error {
							user.Email = "user@example.com"
							user.Password = string(hashedPassword)
							return nil
						})
					repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
					return repo
				}(),
			},
			args: args{
				ctx: context.TODO(),
				request: &model.LoginUserRequest{
					Email:    "user@example.com",
					Password: "password123",
				},
			},
			mockTx: func() {
				mock.ExpectBegin()
				mock.ExpectCommit().WillReturnError(errors.New("commit failed"))
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup transaction expectations
			if tt.mockTx != nil {
				tt.mockTx()
			}

			c := &UserUseCase{
				DB:             tt.fields.DB,
				Log:            tt.fields.Log,
				Validate:       tt.fields.Validate,
				UserRepository: tt.fields.UserRepository,
			}
			got, err := c.Login(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserUseCase.Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For successful login, just check that a token exists
			if !tt.wantErr && got != nil {
				if got.Token == "" {
					t.Errorf("UserUseCase.Login() expected token to be set, but it was empty")
				}
			} else if !tt.wantErr && got == nil {
				t.Errorf("UserUseCase.Login() got nil, expected non-nil response")
			}
		})
	}
}
