package claim

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AdminUserClaim 后台管理员 JWT
type AdminUserClaim struct {
	Exp    int64  `json:"exp"`
	Nbf    int64  `json:"nbf"`
	Iss    string `json:"iss"`
	Iat    int64  `json:"iat"`
	Jti    string `json:"jti"`
	Sub    string `json:"sub"`
	UserId int64  `json:"userId"`
}

func (uc *AdminUserClaim) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(uc.Exp, 0)), nil
}

func (uc *AdminUserClaim) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(uc.Iat, 0)), nil
}

func (uc *AdminUserClaim) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(uc.Nbf, 0)), nil
}

func (uc *AdminUserClaim) GetIssuer() (string, error) {
	return uc.Iss, nil
}

func (uc *AdminUserClaim) GetSubject() (string, error) {
	return uc.Sub, nil
}

func (uc *AdminUserClaim) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{}, nil
}

func NewAdminUserClaim(userId int64, ttl int64) *AdminUserClaim {
	now := time.Now()
	return &AdminUserClaim{
		UserId: userId,
		Sub:    "user access token",
		Iss:    "Shengshi Org",
		Iat:    now.Unix(),
		Nbf:    now.Unix(),
		Exp:    now.Add(time.Second * time.Duration(ttl)).Unix(),
	}
}
