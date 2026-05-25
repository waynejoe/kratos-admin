package claim

import (
	"context"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/transport"
	"github.com/golang-jwt/jwt/v5"

	"kratos-admin/pkg/toolbox/errorx"
)

type UserClaim struct {
	Exp           int64  `json:"exp"`
	Nbf           int64  `json:"nbf"`
	Iss           string `json:"iss"`
	Iat           int64  `json:"iat"`
	Jti           string `json:"jti"`
	Sub           string `json:"sub"`
	UserId        int64  `json:"userId"`
	RegionId      string `json:"regionId"`
	LoginRegionId string `json:"loginRegionId"`
	IsGuest       int32  `json:"isGuest"` // 1: guest 2: user
}

func (uc *UserClaim) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(uc.Exp, 0)), nil
}

func (uc *UserClaim) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(uc.Iat, 0)), nil
}
func (uc *UserClaim) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(uc.Nbf, 0)), nil
}
func (uc *UserClaim) GetIssuer() (string, error) {
	return uc.Iss, nil
}
func (uc *UserClaim) GetSubject() (string, error) {
	return uc.Sub, nil
}
func (uc *UserClaim) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{}, nil
}

func ParseWithClaims(token, jwtSecret string) (claims *UserClaim, err error) {
	claims = &UserClaim{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, errorx.WithStack(err)
	}
	return claims, nil
}

// ParseClaimFromCtx 从ctx中解析claims
func ParseClaimFromCtx(ctx context.Context, secret string) (string, *UserClaim, error) {
	header, ok := transport.FromServerContext(ctx)
	if !ok {
		return "", nil, errorx.New("wrong context")
	}

	token := strings.TrimPrefix(header.RequestHeader().Get("Authorization"), "Bearer ")
	if token == "" {
		return "", nil, errorx.New("invalid access token")
	}

	claims, err := ParseWithClaims(token, secret)

	return token, claims, err
}

func NewUserClaim(userId int64, regionId string, loginRegionId string, isGuest int32, ttl int64) *UserClaim {
	now := time.Now()
	claim := &UserClaim{}
	claim.UserId = userId
	claim.IsGuest = isGuest
	claim.RegionId = regionId
	claim.LoginRegionId = loginRegionId
	claim.Sub = "user access token"
	claim.Iss = "Shengshi Org"
	claim.Iat = now.Unix()
	claim.Nbf = now.Unix()
	claim.Exp = now.Add(time.Second * time.Duration(ttl)).Unix()
	//claim.Jti = ""
	return claim
}
