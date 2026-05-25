package adminuserbiz

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"kratos-admin/internal/conf"
	"kratos-admin/internal/data/adminrepo"
	pb "kratos-admin/pb/admin/v1"
	"kratos-admin/pkg/model/adminmodel"
	"kratos-admin/pkg/toolbox/claim"
)

const AuthAccessKey = "admin:auth:access:%d"

type LoginUsecase struct {
	adminUserRepo *adminrepo.AdminUserRepo
	securityConf  *conf.Security
}

func NewLoginUsecase(securityConf *conf.Security, adminUserRepo *adminrepo.AdminUserRepo) *LoginUsecase {
	return &LoginUsecase{
		adminUserRepo: adminUserRepo,
		securityConf:  securityConf,
	}
}

func (uc *LoginUsecase) Login(ctx context.Context, username, password string) (string, error) {
	adminUser, err := uc.adminUserRepo.GetAdminUserByAccount(ctx, username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", pb.ErrorUserLoginFailed("user or password error")
	}
	if err != nil {
		return "", err
	}
	if adminUser.Status != adminmodel.UserStatusNormal {
		return "", pb.ErrorUserLoginFailed("user disabled")
	}

	userPassword, err := adminmodel.NewUserPasswordFromModel(adminUser.Password)
	if err != nil {
		return "", err
	}
	if err = userPassword.Validate(password); err != nil {
		return "", pb.ErrorUserLoginFailed("user or password error")
	}

	return uc.issueToken(ctx, adminUser)
}

func (uc *LoginUsecase) issueToken(ctx context.Context, user *adminmodel.AdminUser) (string, error) {
	if user == nil {
		return "", pb.ErrorUserLoginFailed("用户或密码错误")
	}

	accessTokenKey := fmt.Sprintf(AuthAccessKey, user.Id)
	newAccessToken := uc.newAccessToken(user)

	_ = uc.adminUserRepo.GetCache().Del(ctx, accessTokenKey)
	if err := uc.adminUserRepo.GetCache().Set(ctx, accessTokenKey, newAccessToken, uc.securityConf.JwtTTL); err != nil {
		return "", errors.New("set access token failed")
	}

	return newAccessToken, nil
}

func (uc *LoginUsecase) newAccessToken(user *adminmodel.AdminUser) string {
	tokenClaim := claim.NewAdminUserClaim(user.Id, uc.securityConf.JwtTTL)
	accessToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaim).SignedString([]byte(uc.securityConf.JwtSecret))
	return accessToken
}
