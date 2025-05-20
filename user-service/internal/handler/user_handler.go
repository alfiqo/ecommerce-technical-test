package handler

import (
	"errors"
	"user-service/internal/context"
	"user-service/internal/delivery/http/response"
	appErrors "user-service/internal/errors"
	"user-service/internal/model"
	"user-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	Log     *logrus.Logger
	UseCase usecase.UserUseCaseInterface
}

func NewUserHandler(useCase usecase.UserUseCaseInterface, logger *logrus.Logger) *UserHandler {
	return &UserHandler{
		Log:     logger,
		UseCase: useCase,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags Users
// @Accept json
// @Produce json
// @Param user body model.RegisterUserRequest true "User registration details"
// @Success 201 {object} model.UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users [post]
func (c *UserHandler) Register(ctx *fiber.Ctx) error {
	// Get request ID from context for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	
	request := new(model.RegisterUserRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to parse request body")
		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInvalidInput, err), c.Log)
	}

	// Add timeout to context
	timeoutCtx, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()

	userResponse, err := c.UseCase.Create(timeoutCtx, request)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"email":      request.Email,
			"error":      err.Error(),
		}).Warn("Failed to register user")
		
		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, c.Log)
		}

		// Convert known error strings to proper error types
		switch err.Error() {
		case "validation failed":
			return response.JSONError(ctx, appErrors.ErrInvalidInput, c.Log)
		case "failed to create user", "Failed create user to database":
			return response.JSONError(ctx, appErrors.ErrDuplicateEmail, c.Log)
		default:
			return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInternalServer, err), c.Log)
		}
	}

	return response.JSONSuccess(ctx, userResponse)
}

// Login godoc
// @Summary User login
// @Description Authenticate a user and return a token
// @Tags Users
// @Accept json
// @Produce json
// @Param user body model.LoginUserRequest true "User login credentials"
// @Success 200 {object} model.UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users/login [post]
func (c *UserHandler) Login(ctx *fiber.Ctx) error {
	// Get request ID from context for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	
	request := new(model.LoginUserRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to parse request body")
		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInvalidInput, err), c.Log)
	}

	// Add timeout to context
	timeoutCtx, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()

	userResponse, err := c.UseCase.Login(timeoutCtx, request)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"email":      request.Email,
			"error":      err.Error(),
		}).Warn("Failed to login user")
		
		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, c.Log)
		}

		// Convert Fiber errors to application errors
		if err == fiber.ErrUnauthorized {
			return response.JSONError(ctx, appErrors.ErrInvalidCredentials, c.Log)
		} else if err == fiber.ErrBadRequest {
			return response.JSONError(ctx, appErrors.ErrInvalidInput, c.Log)
		} else {
			return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInternalServer, err), c.Log)
		}
	}

	return response.JSONSuccess(ctx, userResponse)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Returns user details for the specified ID
// @Tags Users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} model.UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /users/{id} [get]
func (c *UserHandler) GetUser(ctx *fiber.Ctx) error {
	// Get request ID from context for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	
	// The user ID parameter from the URL
	id := ctx.Params("id")
	if id == "" {
		return response.JSONError(ctx, appErrors.WithMessage(appErrors.ErrInvalidInput, "user id is required"), c.Log)
	}

	// Get the authenticated user ID from context
	userId := ctx.Locals("userId")
	
	// Add user ID to context
	userCtx = context.WithUserID(userCtx, id)
	
	// Here you would fetch user details based on ID from the database
	// For now we'll just return the authenticated user ID
	userData := map[string]interface{}{
		"id":               id,
		"authenticated_as": userId,
		"message":          "User details fetched successfully",
	}

	return response.JSONSuccess(ctx, userData)
}