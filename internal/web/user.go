package web

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/liliang-cn/webook/internal/domain"
	"github.com/liliang-cn/webook/internal/service"
)

const (
	emailRegx    = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
	passwordRegx = "^[a-zA-Z0-9]{8,}$"
)

type UserHandler struct {
	svc             *service.UserService
	emailRegxExp    *regexp.Regexp
	passwordRegxExp *regexp.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		svc:             svc,
		emailRegxExp:    regexp.MustCompile(emailRegx, regexp.None),
		passwordRegxExp: regexp.MustCompile(passwordRegx, regexp.None),
	}
}

type SignUpRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

func (u *UserHandler) RegisterRoutesV1(ug *gin.RouterGroup) {
	ug.POST("/signup", u.SignUp)
	//ug.POST("/login", u.Login)
	ug.POST("/login", u.LoginJWT)
	ug.POST("/edit", u.EditJWT)
	//ug.GET("/profile", u.Profile)
	ug.GET("/profile", u.ProfileJWT)
	ug.GET("/status", u.Status)
	ug.GET("logout", u.Logout)
}

func (u *UserHandler) Status(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"status": "ok"})
}

func (u *UserHandler) SignUp(ctx *gin.Context) {

	var req SignUpRequest

	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	fmt.Println(req)

	if req.Email == "" {
		ctx.JSON(400, gin.H{"error": "username is required"})
		return
	}

	if req.Password == "" {
		ctx.JSON(400, gin.H{"error": "password is required"})
		return

	}

	if req.ConfirmPassword == "" {
		ctx.JSON(400, gin.H{"error": "confirm password is required"})
		return

	}

	ok, err := u.emailRegxExp.MatchString(req.Email)

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"error": "internal error"})
		return
	}

	if !ok {
		ctx.JSON(http.StatusOK, gin.H{"error": "invalid email"})
		return
	}

	ok, err = u.passwordRegxExp.MatchString(req.Password)

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"error": "internal error"})
		return

	}

	if !ok {
		ctx.JSON(http.StatusOK, gin.H{"error": "invalid password"})
		return

	}

	if req.Password != req.ConfirmPassword {
		ctx.JSON(400, gin.H{"error": "password and confirm password do not match"})
		return

	}

	// create user
	err = u.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	if errors.Is(err, service.ErrDuplicateEmail) {
		ctx.JSON(400, gin.H{"error": "duplicate email"})
		return

	}

	if err != nil {
		ctx.JSON(500, gin.H{"error": "internal error"})
		return

	}

	ctx.JSON(200, gin.H{"message": "user created"})
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq

	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := u.svc.Login(ctx, req.Email, req.Password)

	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.JSON(http.StatusOK, gin.H{
			"error": "invalid email or password",
		})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"error": "internal error",
		})
		return
	}

	sess := sessions.Default(ctx)
	sess.Set("uid", user.ID)
	sess.Options(sessions.Options{
		MaxAge: 60,
	})
	err = sess.Save()
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"error": "internal error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "login success"})
}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq

	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := u.svc.Login(ctx, req.Email, req.Password)

	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.JSON(http.StatusOK, gin.H{
			"error": "invalid email or password",
		})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"error": "internal error",
		})
		return
	}

	claims := &UserClaims{
		Uid: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("J5gqVj2LTDm82PPVvRcRzU7m6uQwzj5E"))

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"error": "internal error",
		})
		return
	}

	ctx.Header("x-jwt-token", tokenStr)

	ctx.JSON(http.StatusOK, gin.H{"message": "login success"})
}

func (u *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Options(sessions.Options{MaxAge: -1})
	err := sess.Save()

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"error": "internal error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "logout success"})
}

type Profile struct {
	NickName    string `json:"nickName"`
	BirthDate   string `json:"birthDate"`
	Description string `json:"description"`
}

func (u *UserHandler) Edit(ctx *gin.Context) {
	var req Profile

	if err := ctx.Bind(&req); err != nil {
		return
	}

	sess := sessions.Default(ctx)
	uid := sess.Get("uid")

	if uid == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err := u.svc.EditProfile(ctx, uid.(int64), domain.User{
		ID:          uid.(int64),
		NickName:    req.NickName,
		Description: req.Description,
		BirthDate:   req.BirthDate,
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

func (u *UserHandler) EditJWT(ctx *gin.Context) {
	var req Profile

	if err := ctx.Bind(&req); err != nil {
		return
	}

	c, ok := ctx.Get("claims")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	claims, ok := c.(*UserClaims)

	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if claims.Uid == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err := u.svc.EditProfile(ctx, claims.Uid, domain.User{
		ID:          claims.Uid,
		NickName:    req.NickName,
		Description: req.Description,
		BirthDate:   req.BirthDate,
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

func (u *UserHandler) Profile(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	uid := sess.Get("uid")

	if uid == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := u.svc.GetProfile(ctx, uid.(int64))
	if err != nil {
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user": Profile{
			BirthDate:   user.BirthDate,
			Description: user.Description,
			NickName:    user.NickName,
		},
	})
}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	c, ok := ctx.Get("claims")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	claims, ok := c.(*UserClaims)

	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	user, err := u.svc.GetProfile(ctx, claims.Uid)
	if err != nil {
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user": Profile{
			BirthDate:   user.BirthDate,
			Description: user.Description,
			NickName:    user.NickName,
		},
	})
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid int64
}
