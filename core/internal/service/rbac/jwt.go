package rbac

import (
	"billionmail-core/internal/consts"
	"billionmail-core/internal/service/public"
	"context"
	"fmt"
	"github.com/gogf/gf/v2/util/gconv"
	"strings"
	"time"

	"github.com/gogf/gf/util/guid"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/golang-jwt/jwt/v5"
)

// JWTCustomClaims defines the custom claims for JWT
type JWTCustomClaims struct {
	AccountId int64    `json:"accountId"`
	Username  string   `json:"username"`
	Roles     []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTService provides methods for JWT operations
type JWTService struct {
	Secret        string        // Secret key for signing JWT
	AccessExpiry  time.Duration // Duration for access token expiry
	RefreshExpiry time.Duration // Duration for refresh token expiry
}

// newJWTService creates a new JWTService instance
func newJWTService() *JWTService {
	defaultJwtSecret := ""

	if dbpass, err := public.DockerEnv("DBPASS"); err == nil {
		defaultJwtSecret += dbpass
	}

	if redispass, err := public.DockerEnv("REDISPASS"); err == nil {
		defaultJwtSecret += redispass
	}

	if defaultJwtSecret == "" {
		panic("no default jwt secret found")
	}

	return &JWTService{
		Secret:        g.Cfg().MustGet(context.Background(), "jwt.secret", defaultJwtSecret).String(),
		AccessExpiry:  time.Duration(g.Cfg().MustGet(context.Background(), "jwt.accessExpiry", 86400).Int()) * time.Second,
		RefreshExpiry: time.Duration(g.Cfg().MustGet(context.Background(), "jwt.refreshExpiry", 86400*7).Int()) * time.Second,
	}
}

// GenerateToken generates a new JWT token
func (s *JWTService) GenerateToken(accountId int64, username string, roles []string) (string, int64, error) {
	expiryTime := time.Now().Add(s.AccessExpiry)
	claims := &JWTCustomClaims{
		AccountId: accountId,
		Username:  username,
		Roles:     roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    consts.DEFAULT_SERVER_NAME,
			Subject:   "access_token",
			ID:        guid.S() + guid.S(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.Secret))
	if err != nil {
		return "", 0, err
	}

	return signedToken, expiryTime.Unix(), nil
}

// GenerateRefreshToken generates a refresh token
func (s *JWTService) GenerateRefreshToken(accountId int64, username string) (string, error) {
	expiryTime := time.Now().Add(s.RefreshExpiry)
	claims := &JWTCustomClaims{
		AccountId: accountId,
		Username:  username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    consts.DEFAULT_SERVER_NAME,
			Subject:   "refresh_token",
			ID:        guid.S() + guid.S(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.Secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ParseToken parses and validates a JWT token
func (s *JWTService) ParseToken(tokenString string) (*JWTCustomClaims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &JWTCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTCustomClaims); ok && token.Valid && !s.IsTokenBlacklisted(claims) {
		return claims, nil
	}

	return nil, gerror.New("invalid token")
}

// InvalidateToken invalidates a JWT token
func (s *JWTService) InvalidateToken(token *JWTCustomClaims) error {
	// Using redis to invalidate the token
	now := time.Now()
	if token.RegisteredClaims.ExpiresAt == nil || token.RegisteredClaims.ExpiresAt.Time.Before(now) {
		return nil
	}

	return g.Redis().SetEX(context.Background(), consts.JWT_BLACK_LIST_KEY_PREFIX+token.RegisteredClaims.ID, 0, gconv.Int64(token.RegisteredClaims.ExpiresAt.Sub(now).Seconds()))
}

// IsTokenBlacklisted checks if a token is blacklisted
func (s *JWTService) IsTokenBlacklisted(token *JWTCustomClaims) bool {
	if token.RegisteredClaims.ExpiresAt == nil {
		return false
	}

	exists, err := g.Redis().Exists(context.Background(), consts.JWT_BLACK_LIST_KEY_PREFIX+token.RegisteredClaims.ID)
	if err != nil {
		return false
	}
	return exists > 0
}

// JWTAuthMiddleware provides JWT authentication middleware
func (s *JWTService) JWTAuthMiddleware(r *ghttp.Request) {
	// Skip authentication for login and refresh token endpoints
	if r.URL.Path == "/api/login" || r.URL.Path == "/api/refresh-token" || r.URL.Path == "/api/unsubscribe/user_group" || r.URL.Path == "/api/get_validate_code" || r.URL.Path == "/api/unsubscribe" || r.URL.Path == "/api/languages/set" || r.URL.Path == "/api/languages/get" || r.URL.Path == "/api/batch_mail/api/send" {
		r.Middleware.Next()
		return
	}

	resp := public.CodeMap[401]

	// no message
	resp.Msg = ""

	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		resp.Msg = "no authorization"
		r.Response.WriteJson(resp)
		r.Exit()
		return
	}

	// Parse token
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := s.ParseToken(tokenString)
	if err != nil {
		//resp.Msg = fmt.Sprintf("invalid token: %s", err.Error())
		r.Response.WriteJson(resp)
		r.Exit()
		return
	}

	// Set account info in context
	r.SetCtxVar("accountId", claims.AccountId)
	r.SetCtxVar("username", claims.Username)

	// Retrieve roles from cache or database
	cacheKey := fmt.Sprintf("ACCOUNT_ROLES_%d", claims.AccountId)
	roles := public.GetCache(cacheKey)

	if roles == nil {
		roles, err = Account().GetAccountRoles(r.GetCtx(), claims.AccountId)
		if err != nil {
			resp.Msg = "failed to get account roles"
			r.Response.WriteJson(resp)
			r.Exit()
			return
		}

		public.SetCache(cacheKey, roles, 20)
	} else {
		r.SetCtxVar("roles", roles)
	}

	// Update Session
	err = r.Session.Set("UserLogin", true)

	if err != nil {
		g.Log().Warning(r.GetCtx(), "set UserLogin failed ", err)
	}

	r.Middleware.Next()
}
