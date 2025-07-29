package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"

	"gochuji/webook/internal/domain"
	"gochuji/webook/internal/service"
)

type UserHandler struct {
	svc         *service.UserService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	const (
		emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		// 和上面比起来，用 ` 看起来就比较清爽
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	return &UserHandler{
		svc:         svc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	// REST 风格 但是现在不用, 一般看公司用什么你就用什么
	//server.POST("/user", h.SignUp)
	//server.PUT("/user", h.SignUp)
	//server.GET("/users/:username", h.Profile)

	ug := server.Group("/users")

	// POST /users/signup
	ug.POST("/signup", h.SignUp)

	// POST /users/login
	ug.POST("/login", h.Login)
	//ug.POST("/login", h.JWTLogin)

	// POST /users/edit
	ug.POST("/edit", h.Edit)

	// GET /users/profile
	ug.GET("/profile", h.Profile)
}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	isEmail, err := h.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "非法邮箱格式")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入密码不对")
		return
	}

	isPassword, err := h.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含字母、数字、特殊字符，并且不少于八位")
		return
	}

	err = h.svc.Signup(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	switch err {
	case nil:
		ctx.String(http.StatusOK, "注册成功")
	case service.ErrDuplicateEmail:
		ctx.String(http.StatusOK, "邮箱冲突，请换一个")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) Login(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	u, err := h.svc.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		// 获取当前会话
		sess := sessions.Default(ctx)
		// 设置会话中的userId为u.Id
		sess.Set("userId", u.Id)
		// 设置会话的选项
		// 设置session的选项
		sess.Options(sessions.Options{
			// 设置session的路径
			Path: "",
			// 设置session的域名
			Domain: "",
			// 设置session的最大生命周期
			MaxAge: 900,
			// 设置session是否只在https下传输
			Secure: false,
			// 设置session是否只能通过http传输
			HttpOnly: false,
			// 设置session的SameSite属性
			SameSite: 0,
		})
		err = sess.Save()
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或者密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) JWTLogin(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	// 声明JWT的Claims结构
	type Claims struct {
		Username string `json:"username"`
		jwt.RegisteredClaims
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	_, err := h.svc.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		//jwt登录实现
		context.TODO()

		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或者密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) Edit(ctx *gin.Context) {
	//{"nickname":"huali","birthday":"2025-07-28","aboutMe":"sdfsfdfsfdfsdghhhh"}
	sess := sessions.Default(ctx)
	// 如果 session 中没有 userId，则表示用户未登录
	if sess.Get("userId") == nil {
		// 中断，不要往后执行，也就是不要执行后面的业务逻辑
		ctx.String(http.StatusOK, "用户未登录")
		return
	}
	userIDAnyBtIsInt64 := sess.Get("userId")

	// 从上下文中获取用户ID
	//userID, exists := ctx.Get("user_id")
	//if !exists {
	//	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取用户信息"})
	//	return
	//}

	userIDInt64, ok := userIDAnyBtIsInt64.(int64)
	if !ok {
		// 处理类型不匹配的情况
		ctx.String(http.StatusOK, "系统错误")
		log.Fatal("类型断言失败：puserID2 不是 int64 类型")
	}
	userIDStr := strconv.FormatInt(userIDInt64, 10)

	//处理传入的json
	type Req struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	u, err := h.svc.Edit(ctx, userIDStr, req.Nickname, req.Birthday, req.AboutMe)
	fmt.Println(u, time.Now())
	switch err {
	case nil:
		ctx.String(http.StatusOK, "")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) Profile(ctx *gin.Context) {

	// 获取 session
	sess := sessions.Default(ctx)
	// 如果 session 中没有 userId，则表示用户未登录
	if sess.Get("userId") == nil {
		// 中断，不要往后执行，也就是不要执行后面的业务逻辑
		ctx.String(http.StatusOK, "用户未登录")
		return
	}
	userIDAnyBtIsInt64 := sess.Get("userId")

	// 从上下文中获取用户ID
	//userID, exists := ctx.Get("user_id")
	//if !exists {
	//	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取用户信息"})
	//	return
	//}

	userIDInt64, ok := userIDAnyBtIsInt64.(int64)
	if !ok {
		// 处理类型不匹配的情况
		ctx.String(http.StatusOK, "系统错误")
		log.Fatal("类型断言失败：puserID2 不是 int64 类型")
	}
	userIDStr := strconv.FormatInt(userIDInt64, 10)

	// 从数据库获取用户信息
	u, err := h.svc.FindByID(ctx, userIDStr)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "无法获取用户信息")
	}

	profile := &Profile{
		Email:    u.Email,
		Phone:    u.Phone,
		Nickname: u.Nickname,
		Birthday: u.Birthday,
		AboutMe:  u.AboutMe,
	}

	// 返回JSON响应
	ctx.JSON(http.StatusOK, profile)
}

type Profile struct {
	Email    string `json:"Email"`
	Phone    string `json:"Phone"`
	Nickname string `json:"Nickname"`
	Birthday string `json:"Birthday"`
	AboutMe  string `json:"AboutMe"`
}
