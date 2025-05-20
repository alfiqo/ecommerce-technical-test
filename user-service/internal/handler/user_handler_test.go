package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"user-service/internal/model"
	usecase_mock "user-service/mocks/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_Register(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        any
		mockExpectations   func(mockUseCase *usecase_mock.MockUserUseCaseInterface)
		expectedStatusCode int
	}{
		{
			name: "success",
			requestBody: model.RegisterUserRequest{
				Name:     "testuser",
				Email:    "testemail@mail.com",
				Phone:    "1234567890",
				Password: "password123",
			},
			mockExpectations: func(mockUseCase *usecase_mock.MockUserUseCaseInterface) {
				mockUseCase.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&model.UserResponse{}, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "invalid request body",
			requestBody:        `{}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "err on create user",
			requestBody: map[string]any{
				"name":     "testuser",
				"email":    "testemail@mail.com",
				"phone":    "1234567890",
				"password": "password123",
			},
			mockExpectations: func(mockUseCase *usecase_mock.MockUserUseCaseInterface) {
				mockUseCase.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, fiber.ErrInternalServerError)
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Disable logger output during tests
			logger := logrus.New()
			logger.SetOutput(&bytes.Buffer{})

			// Create a new fiber.Ctx
			app := fiber.New()
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(ctx)
			
			mockUseCase := usecase_mock.NewMockUserUseCaseInterface(ctrl)
			handler := NewUserHandler(mockUseCase, logger)

			// Create request body
			jsonBody, _ := json.Marshal(tt.requestBody)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")

			// Set request to context
			ctx.Request().SetRequestURI("/register")
			ctx.Request().Header.Set("Content-Type", req.Header.Get("Content-Type"))
			ctx.Request().Header.Set("Accept", req.Header.Get("Accept"))
			ctx.Request().SetBody(jsonBody)

			// Set mock expectations
			if tt.mockExpectations != nil {
				tt.mockExpectations(mockUseCase)
			}

			// Call the handler
			handler.Register(ctx)

			// Check the response status code
			assert.Equal(t, tt.expectedStatusCode, ctx.Response().StatusCode())

			// Additional assertions based on response body
			var response map[string]interface{}
			err := json.Unmarshal(ctx.Response().Body(), &response)
			assert.NoError(t, err)
			
			if tt.expectedStatusCode == http.StatusOK {
				assert.True(t, response["success"].(bool))
				assert.NotNil(t, response["data"])
			} else {
				assert.False(t, response["success"].(bool))
				assert.NotNil(t, response["error"])
			}
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        any
		mockExpectations   func(mockUseCase *usecase_mock.MockUserUseCaseInterface)
		expectedStatusCode int
	}{
		{
			name: "success",
			requestBody: model.LoginUserRequest{
				Email:    "testemail@mail.com",
				Password: "password123",
			},
			mockExpectations: func(mockUseCase *usecase_mock.MockUserUseCaseInterface) {
				mockUseCase.EXPECT().Login(gomock.Any(), gomock.Any()).Return(&model.UserResponse{
					Token: "sample-token",
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "invalid request body",
			requestBody:        `{}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid credentials",
			requestBody: model.LoginUserRequest{
				Email:    "testemail@mail.com",
				Password: "wrong-password",
			},
			mockExpectations: func(mockUseCase *usecase_mock.MockUserUseCaseInterface) {
				mockUseCase.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, fiber.ErrUnauthorized)
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "internal server error",
			requestBody: model.LoginUserRequest{
				Email:    "testemail@mail.com",
				Password: "password123",
			},
			mockExpectations: func(mockUseCase *usecase_mock.MockUserUseCaseInterface) {
				mockUseCase.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, fiber.ErrInternalServerError)
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Disable logger output during tests
			logger := logrus.New()
			logger.SetOutput(&bytes.Buffer{})

			// Create a new fiber.Ctx
			app := fiber.New()
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(ctx)
			
			mockUseCase := usecase_mock.NewMockUserUseCaseInterface(ctrl)
			handler := NewUserHandler(mockUseCase, logger)

			// Create request body
			jsonBody, _ := json.Marshal(tt.requestBody)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")

			// Set request to context
			ctx.Request().SetRequestURI("/login")
			ctx.Request().Header.Set("Content-Type", req.Header.Get("Content-Type"))
			ctx.Request().Header.Set("Accept", req.Header.Get("Accept"))
			ctx.Request().SetBody(jsonBody)

			// Set mock expectations
			if tt.mockExpectations != nil {
				tt.mockExpectations(mockUseCase)
			}

			// Call the handler
			handler.Login(ctx)

			// Check the response status code
			assert.Equal(t, tt.expectedStatusCode, ctx.Response().StatusCode())

			// Additional assertions based on response body
			var response map[string]interface{}
			err := json.Unmarshal(ctx.Response().Body(), &response)
			assert.NoError(t, err)
			
			if tt.expectedStatusCode == http.StatusOK {
				assert.True(t, response["success"].(bool))
				data := response["data"].(map[string]interface{})
				assert.NotEmpty(t, data["token"])
			} else {
				assert.False(t, response["success"].(bool))
				assert.NotNil(t, response["error"])
			}
		})
	}
}