package api

import (
	"go-one/internal/serializer"
	"go-one/internal/service"
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

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UserRegister 用户注册
func (h *Handler) UserRegister(c *gin.Context) {
	// 1. 获取BusinessContext（与HTTP层解耦）
	bizCtx := GetBusinessContext(c)

	// 2. 绑定并验证请求参数
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("参数错误", err))
		return
	}

	// 3. 转换为Service层DTO
	dto := &service.RegisterDTO{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	// 4. 调用Service层（传入BusinessContext）
	userService := h.serviceManager.NewUserService()
	result, serviceErr := userService.Register(bizCtx, dto)
	if serviceErr != nil {
		HandleServiceError(c, serviceErr)
		return
	}

	// 5. 返回成功响应
	ResponseWithMessage(c, "注册成功", gin.H{
		"user":          result.User,
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
	})
}

// UserLogin 用户登录
func (h *Handler) UserLogin(c *gin.Context) {
	// 1. 获取BusinessContext
	bizCtx := GetBusinessContext(c)

	// 2. 绑定并验证请求参数
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("参数错误", err))
		return
	}

	// 3. 转换为Service层DTO
	dto := &service.LoginDTO{
		Username: req.Username,
		Password: req.Password,
	}

	// 4. 调用Service层
	userService := h.serviceManager.NewUserService()
	result, serviceErr := userService.Login(bizCtx, dto)
	if serviceErr != nil {
		HandleServiceError(c, serviceErr)
		return
	}

	// 5. 返回成功响应
	ResponseWithMessage(c, "登录成功", gin.H{
		"user":          result.User,
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
	})
}

// GetUserProfile 获取用户资料
func (h *Handler) GetUserProfile(c *gin.Context) {
	// 1. 获取BusinessContext
	bizCtx := GetBusinessContext(c)

	// 2. 从JWT claims获取用户ID
	if !bizCtx.IsAuthenticated() {
		c.JSON(http.StatusUnauthorized, serializer.Err(serializer.CodeUnauthorized, "未认证", nil))
		return
	}

	// 从Account获取用户ID
	if bizCtx.Account == nil {
		c.JSON(http.StatusBadRequest, serializer.Err(serializer.CodeError, "无法获取用户信息", nil))
		return
	}

	// 3. 调用Service层
	userService := h.serviceManager.NewUserService()
	user, serviceErr := userService.GetUserByID(bizCtx)
	if serviceErr != nil {
		HandleServiceError(c, serviceErr)
		return
	}

	// 4. 返回成功响应
	ResponseWithMessage(c, "获取成功", user)
}

// UpdateUserProfile 更新用户资料
func (h *Handler) UpdateUserProfile(c *gin.Context) {
	// 1. 获取BusinessContext
	bizCtx := GetBusinessContext(c)

	// 2. 验证认证状态
	if !bizCtx.IsAuthenticated() || bizCtx.Account == nil {
		c.JSON(http.StatusUnauthorized, serializer.Err(serializer.CodeUnauthorized, "未认证", nil))
		return
	}

	// 3. 绑定请求参数
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("参数错误", err))
		return
	}

	// 4. 转换为Service层DTO
	dto := &service.UpdateProfileDTO{
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
	}

	// 5. 调用Service层
	userService := h.serviceManager.NewUserService()
	serviceErr := userService.UpdateProfile(bizCtx, dto)
	if serviceErr != nil {
		HandleServiceError(c, serviceErr)
		return
	}

	// 6. 返回成功响应
	ResponseWithMessage(c, "更新成功", nil)
}

// ChangePassword 修改密码
func (h *Handler) ChangePassword(c *gin.Context) {
	// 1. 获取BusinessContext
	bizCtx := GetBusinessContext(c)

	// 2. 验证认证状态
	if !bizCtx.IsAuthenticated() || bizCtx.Account == nil {
		c.JSON(http.StatusUnauthorized, serializer.Err(serializer.CodeUnauthorized, "未认证", nil))
		return
	}

	// 3. 绑定请求参数
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("参数错误", err))
		return
	}

	// 4. 转换为Service层DTO
	dto := &service.ChangePasswordDTO{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}

	// 5. 调用Service层
	userService := h.serviceManager.NewUserService()
	serviceErr := userService.ChangePassword(bizCtx, dto)
	if serviceErr != nil {
		HandleServiceError(c, serviceErr)
		return
	}

	// 6. 返回成功响应
	ResponseWithMessage(c, "密码修改成功", nil)
}

// ListUsers 获取用户列表
func (h *Handler) ListUsers(c *gin.Context) {
	// 1. 获取BusinessContext
	bizCtx := GetBusinessContext(c)

	// 2. 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 3. 构建查询对象
	query := &service.ListUsersQuery{
		Page:     page,
		PageSize: pageSize,
	}

	// 4. 调用Service层
	userService := h.serviceManager.NewUserService()
	result, serviceErr := userService.ListUsers(bizCtx, query)
	if serviceErr != nil {
		HandleServiceError(c, serviceErr)
		return
	}

	// 5. 返回成功响应
	ResponseWithMessage(c, "获取成功", gin.H{
		"list":      result.List,
		"total":     result.Total,
		"page":      result.Page,
		"page_size": result.PageSize,
	})
}

// RefreshToken 刷新访问令牌
func (h *Handler) RefreshToken(c *gin.Context) {
	// 1. 获取BusinessContext
	bizCtx := GetBusinessContext(c)

	// 2. 绑定请求参数
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializer.ParamErr("参数错误", err))
		return
	}

	// 3. 转换为Service层DTO
	dto := &service.RefreshTokenDTO{
		RefreshToken: req.RefreshToken,
	}

	// 4. 调用Service层
	userService := h.serviceManager.NewUserService()
	result, serviceErr := userService.RefreshToken(bizCtx, dto)
	if serviceErr != nil {
		HandleServiceError(c, serviceErr)
		return
	}

	// 5. 返回成功响应
	ResponseWithMessage(c, "刷新成功", gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
	})
}

// Ping 健康检查
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, serializer.Success("pong", nil))
}
