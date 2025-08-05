package web

import (
	"net/http"
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

var JWTKey = []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgK")

type UserClaims struct {
	jwt.RegisteredClaims
	Uid int64
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
	//ug.POST("/login", h.Login)
	ug.POST("/login", h.JWTLogin)

	// POST /users/edit
	ug.POST("/edit", h.JWTEdit)

	// GET /users/profile
	ug.GET("/profile", h.JWTProfile)
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

	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	u, err := h.svc.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:

		// 创建一个JWT中间件，用于登录验证
		uc := UserClaims{
			// 设置用户ID
			Uid: u.Id,
			// 设置注册声明，包括过期时间
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
			},
		}
		// 使用HS512算法生成JWT
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
		// 使用密钥对JWT进行签名
		tokenStr, err := token.SignedString(JWTKey)
		// 如果签名失败，返回系统错误
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
		}
		// ■ ■ 将JWT添加到响应头中
		ctx.Header("x-jwt-token", tokenStr)
		// 返回登录成功
		ctx.String(http.StatusOK, "登录成功")

	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或者密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) JWTEdit(ctx *gin.Context) {
	//{"nickname":"huali","birthday":"2025-07-28","aboutMe":"sdfsfdfsfdfsdghhhh"}
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

	//
	uc, ok := ctx.MustGet("user").(UserClaims)
	if !ok {
		//ctx.String(http.StatusOK, "系统错误")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		//ctx.String(http.StatusOK, "系统错误")
		ctx.String(http.StatusOK, "生日格式不对")
		return
	}

	err = h.svc.UpdateNonSensitiveInfo(ctx, domain.User{
		Id:       uc.Uid,
		Nickname: req.Nickname,
		Birthday: birthday,
		AboutMe:  req.AboutMe,
	})
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	ctx.String(http.StatusOK, "更新成功")
}

func (h *UserHandler) JWTProfile(ctx *gin.Context) {

	uc, ok := ctx.MustGet("user").(UserClaims)
	if !ok {
		//ctx.String(http.StatusOK, "系统错误")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 从数据库获取用户信息
	u, err := h.svc.FindByID(ctx, uc.Uid)
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
	}

	profile := &Profile{
		Email:    u.Email,
		Phone:    u.Phone,
		Nickname: u.Nickname,
		//
		Birthday: u.Birthday.Format(time.DateOnly),
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
