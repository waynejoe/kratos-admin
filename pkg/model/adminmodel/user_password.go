package adminmodel

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"kratos-admin/pkg/toolbox/errorx"
)

// UserPassword 用户密码（DB 中以 JSON 存储）
type UserPassword struct {
	Password        string `json:"-"`
	EncryptPassword string `json:"encrypt_password"`
	Salt            string `json:"salt"`
}

func NewUserPassword(password string) *UserPassword {
	return &UserPassword{Password: password}
}

func NewUserPasswordFromModel(data string) (*UserPassword, error) {
	if data == "" {
		return &UserPassword{}, nil
	}
	var password UserPassword
	if err := json.Unmarshal([]byte(data), &password); err != nil {
		return nil, errors.WithStack(err)
	}
	return &password, nil
}

func (p *UserPassword) ToModel() (string, error) {
	if p == nil {
		return "{}", nil
	}
	data, err := json.Marshal(p)
	if err != nil {
		return "", errorx.WithStack(err)
	}
	return string(data), nil
}

func (p *UserPassword) GenerateSalt() error {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return errorx.WithStack(err)
	}
	p.Salt = hex.EncodeToString(bytes)
	return nil
}

func (p *UserPassword) Encrypt() error {
	if p.Salt == "" {
		if err := p.GenerateSalt(); err != nil {
			return err
		}
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(p.Password+p.Salt), bcrypt.DefaultCost)
	if err != nil {
		return errorx.WithStack(err)
	}
	p.EncryptPassword = string(hashed)
	return nil
}

func (p *UserPassword) Validate(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(p.EncryptPassword), []byte(password+p.Salt))
}
