package api

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"

	"github.com/blastjax/golang-redis/internal/cache"
	"github.com/blastjax/golang-redis/internal/github"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handlers contains all the dependencies needed for API handlers
type Handlers struct {
	userRepo     UserRepository
	redisClient  RedisClient
	githubClient GitHubClient
}

// NewHandlers creates a new Handlers instance
func NewHandlers(userRepo UserRepository, redisClient RedisClient, githubClient GitHubClient) *Handlers {
	return &Handlers{
		userRepo:     userRepo,
		redisClient:  redisClient,
		githubClient: githubClient,
	}
}

// UserCreateRequest represents the payload for creating a new user
type UserCreateRequest struct {
	ID        int     `json:"id" binding:"required"`
	Login     string  `json:"login" binding:"required"`
	AvatarURL string  `json:"avatar_url,omitempty"`
	URL       string  `json:"url,omitempty"`
	Type      string  `json:"type" binding:"required"`
	Name      *string `json:"name,omitempty"`
	Company   *string `json:"company,omitempty"`
	Blog      *string `json:"blog,omitempty"`
	Location  *string `json:"location,omitempty"`
	Email     *string `json:"email,omitempty"`
	Bio       *string `json:"bio,omitempty"`
}

// UserUpdateRequest represents the JSON payload for updating user details
type UserUpdateRequest struct {
	Name     *string `json:"name,omitempty"`
	Blog     *string `json:"blog,omitempty"`
	Location *string `json:"location,omitempty"`
	Company  *string `json:"company,omitempty"`
	Bio      *string `json:"bio,omitempty"`
}

// loginPattern matches GitHub logins, including bot accounts like "dependabot[bot]".
// Anything else is rejected before it can reach the GitHub URL or a cache key.
var loginPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,39}(\[bot\])?$`)

// usernameParam returns the :username path parameter, responding 400 if it isn't a valid GitHub login
func usernameParam(c *gin.Context) (string, bool) {
	username := c.Param("username")
	if !loginPattern.MatchString(username) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid username",
		})
		return "", false
	}
	return username, true
}

// ListUsers handles GET /users - List all users from the database
func (h *Handlers) ListUsers(c *gin.Context) {
	// Parse query parameters for pagination
	limitStr := c.DefaultQuery("limit", "30")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid limit parameter. Must be between 1 and 1000",
		})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid offset parameter. Must be >= 0",
		})
		return
	}

	// Get users from database
	users, err := h.userRepo.GetUsers(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve users from database",
		})
		return
	}

	// Get total count for pagination info
	total, err := h.userRepo.GetTotalUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get total user count",
		})
		return
	}

	// Return paginated response
	c.JSON(http.StatusOK, gin.H{
		"users":  users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"count":  len(users),
	})
}

// GetUser handles GET /users/:username - Get a user from Redis, then the database, then GitHub
func (h *Handlers) GetUser(c *gin.Context) {
	username, ok := usernameParam(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()

	// Step 1: Check Redis cache first
	cachedUser, err := h.redisClient.GetUser(ctx, username)
	switch {
	case errors.Is(err, cache.ErrNotFound):
		// GitHub reported this user missing within the cache TTL
		c.Header("X-Cache-Status", "hit")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	case err != nil:
		// Flag the cache error but continue to the database
		c.Header("X-Cache-Error", "true")
	case cachedUser != nil:
		// Cache hit - return cached data
		c.Header("X-Cache-Status", "hit")
		c.JSON(http.StatusOK, gin.H{
			"user":   cachedUser,
			"source": "cache",
		})
		return
	}

	c.Header("X-Cache-Status", "miss")

	// Step 2: Cache miss - check the database, then fall back to GitHub
	source := "database"
	user, err := h.userRepo.GetUserByUsername(username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		source = "github"
		user, err = h.githubClient.GetUserByUsername(ctx, username)
	}

	if errors.Is(err, github.ErrNotFound) {
		// Cache the miss so repeated lookups don't use up the GitHub rate limit
		if err := h.redisClient.SetUser(ctx, username, nil); err != nil {
			c.Header("X-Cache-Store-Error", "true")
		}
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch user from " + source,
		})
		return
	}

	// Step 3: Store in Redis cache (TTL: 30s by default)
	if err := h.redisClient.SetUser(ctx, username, user); err != nil {
		// Flag the error but don't fail the request
		c.Header("X-Cache-Store-Error", "true")
	}

	// Return user data
	c.JSON(http.StatusOK, gin.H{
		"user":   user,
		"source": source,
	})
}

// UpdateUser handles PUT /users/:username - Update user details
func (h *Handlers) UpdateUser(c *gin.Context) {
	username, ok := usernameParam(c)
	if !ok {
		return
	}

	// Parse JSON payload
	var updateReq UserUpdateRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON payload",
		})
		return
	}

	ctx := c.Request.Context()

	// Step 1: Find user in database by username
	user, err := h.userRepo.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found in database",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user from database",
		})
		return
	}

	// Step 2: Update user fields
	updated := false
	if updateReq.Name != nil {
		user.Name = updateReq.Name
		updated = true
	}
	if updateReq.Blog != nil {
		user.Blog = updateReq.Blog
		updated = true
	}
	if updateReq.Location != nil {
		user.Location = updateReq.Location
		updated = true
	}
	if updateReq.Company != nil {
		user.Company = updateReq.Company
		updated = true
	}
	if updateReq.Bio != nil {
		user.Bio = updateReq.Bio
		updated = true
	}

	if !updated {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No valid fields provided for update",
		})
		return
	}

	// Step 3: Update in the database
	if err := h.userRepo.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user in database",
		})
		return
	}

	// Step 4: Drop the cached copy so the next GET reads the update from the database
	if err := h.redisClient.DeleteUser(ctx, username); err != nil {
		// Flag the error but don't fail the request
		c.Header("X-Cache-Delete-Error", "true")
	}

	// Return updated user
	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"message": "User updated successfully",
	})
}

// DeleteUser handles DELETE /users/:username - Delete user from database and cache
func (h *Handlers) DeleteUser(c *gin.Context) {
	username, ok := usernameParam(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()

	// Step 1: Check if user exists in database
	user, err := h.userRepo.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found in database",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user from database",
		})
		return
	}

	// Step 2: Delete from the database
	if err := h.userRepo.DeleteUser(username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete user from database",
		})
		return
	}

	// Step 3: Delete from Redis cache (if cached)
	if err := h.redisClient.DeleteUser(ctx, username); err != nil {
		// Flag the error but don't fail the request since database deletion succeeded
		c.Header("X-Cache-Delete-Error", "true")
	}

	// Return success response with deleted user info
	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
		"user": gin.H{
			"id":    user.ID,
			"login": user.Login,
		},
	})
}

// HealthCheck handles GET /health - Health check endpoint
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "github-users-api",
	})
}

// CreateUser handles POST /users - Create user and store to database
func (h *Handlers) CreateUser(c *gin.Context) {
	// Parse JSON payload (binding requires id, login and type)
	var newUser UserCreateRequest
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON payload",
		})
		return
	}

	if !loginPattern.MatchString(newUser.Login) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid login",
		})
		return
	}

	// Step 1: Build new user struct
	user := &github.User{
		ID:        newUser.ID,
		Login:     newUser.Login,
		AvatarURL: newUser.AvatarURL,
		URL:       newUser.URL,
		Type:      newUser.Type,
		Name:      newUser.Name,
		Company:   newUser.Company,
		Blog:      newUser.Blog,
		Location:  newUser.Location,
		Email:     newUser.Email,
		Bio:       newUser.Bio,
	}

	// Step 2: Insert into the database; the primary key and unique login reject duplicates
	if err := h.userRepo.CreateUser(user); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "User already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to insert user into database",
		})
		return
	}

	// Step 3: Drop any cached GitHub copy or "not found" entry for this login
	if err := h.redisClient.DeleteUser(c.Request.Context(), user.Login); err != nil {
		c.Header("X-Cache-Delete-Error", "true")
	}

	// Return created user
	c.JSON(http.StatusCreated, gin.H{
		"user":    user,
		"message": "User created successfully",
	})
}
