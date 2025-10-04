package api

import (
	"go-one/internal/serializer"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UserRegister 用户注册
func (h *Handler) UserRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("参数错误", err))
		return
	}

	userService := h.serviceManager.NewUserService()
	user, err := userService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, serializer.Err(serializer.CodeError, err.Error(), nil))
		return
	}

	// 生成token
	token, err := userService.GenerateToken(strconv.FormatUint(uint64(user.ID), 10))
	if err != nil {
		c.JSON(http.StatusInternalServerError, serializer.Err(serializer.CodeError, "生成token失败", err))
		return
	}

	c.JSON(http.StatusOK, serializer.Success("注册成功", gin.H{
		"user":  user,
		"token": token,
	}))
}

// UserLogin 用户登录
func (h *Handler) UserLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("参数错误", err))
		return
	}

	userService := h.serviceManager.NewUserService()
	user, err := userService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, serializer.Err(serializer.CodeUnauthorized, err.Error(), nil))
		return
	}

	// 生成token
	token, err := userService.GenerateToken(strconv.FormatUint(uint64(user.ID), 10))
	if err != nil {
		c.JSON(http.StatusInternalServerError, serializer.Err(serializer.CodeError, "生成token失败", err))
		return
	}

	c.JSON(http.StatusOK, serializer.Success("登录成功", gin.H{
		"user":  user,
		"token": token,
	}))
}

// GetUserProfile 获取用户资料
func (h *Handler) GetUserProfile(c *gin.Context) {
	userIDStr, _ := c.Get("userID")
	userID, err := strconv.ParseUint(userIDStr.(string), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("用户ID错误", err))
		return
	}

	userService := h.serviceManager.NewUserService()
	user, err := userService.GetUserByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, serializer.Err(serializer.CodeNotFound, "用户不存在", err))
		return
	}

	c.JSON(http.StatusOK, serializer.Success("获取成功", user))
}

// UpdateUserProfile 更新用户资料
func (h *Handler) UpdateUserProfile(c *gin.Context) {
	userIDStr, _ := c.Get("userID")
	userID, err := strconv.ParseUint(userIDStr.(string), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("用户ID错误", err))
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("参数错误", err))
		return
	}

	userService := h.serviceManager.NewUserService()
	if err := userService.UpdateProfile(uint(userID), req.Nickname, req.Avatar); err != nil {
		c.JSON(http.StatusInternalServerError, serializer.Err(serializer.CodeError, "更新失败", err))
		return
	}

	c.JSON(http.StatusOK, serializer.Success("更新成功", nil))
}

// ChangePassword 修改密码
func (h *Handler) ChangePassword(c *gin.Context) {
	userIDStr, _ := c.Get("userID")
	userID, err := strconv.ParseUint(userIDStr.(string), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("用户ID错误", err))
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("参数错误", err))
		return
	}

	userService := h.serviceManager.NewUserService()
	if err := userService.ChangePassword(uint(userID), req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, serializer.Err(serializer.CodeError, err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, serializer.Success("密码修改成功", nil))
}

// ListUsers 获取用户列表
func (h *Handler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	userService := h.serviceManager.NewUserService()
	users, total, err := userService.ListUsers(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, serializer.DBErr("查询失败", err))
		return
	}

	c.JSON(http.StatusOK, serializer.Success("获取成功", gin.H{
		"list":      users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}))
}

// Ping 健康检查
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, serializer.Success("pong", nil))
}
