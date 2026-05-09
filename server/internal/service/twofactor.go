package service

import (
	"fmt"
	"time"

	"github.com/chatagent/server/internal/model"
	"github.com/chatagent/server/internal/repository"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pquerna/otp/totp"
)

type TwoFactorService struct {
	totpRepo  *repository.UserTOTPRepository
	cfgRepo   *repository.SystemConfigRepository
	jwtSecret string
}

func NewTwoFactorService(totpRepo *repository.UserTOTPRepository, cfgRepo *repository.SystemConfigRepository, jwtSecret string) *TwoFactorService {
	return &TwoFactorService{totpRepo: totpRepo, cfgRepo: cfgRepo, jwtSecret: jwtSecret}
}

func (s *TwoFactorService) IsEnabled() (bool, error) {
	configs, err := s.cfgRepo.GetAll()
	if err != nil {
		return false, err
	}
	return configs["2fa_enabled"] == "true", nil
}

func (s *TwoFactorService) GetConfig() (map[string]string, error) {
	configs, err := s.cfgRepo.GetAll()
	if err != nil {
		return nil, err
	}
	result := map[string]string{
		"2fa_enabled": configs["2fa_enabled"],
		"2fa_issuer":  configs["2fa_issuer"],
	}
	if result["2fa_enabled"] == "" {
		result["2fa_enabled"] = "false"
	}
	if result["2fa_issuer"] == "" {
		result["2fa_issuer"] = "ChatAgent"
	}
	return result, nil
}

func (s *TwoFactorService) SetConfig(enabled bool, issuer string) error {
	v := "false"
	if enabled {
		v = "true"
	}
	if err := s.cfgRepo.Set("2fa_enabled", v); err != nil {
		return err
	}
	return s.cfgRepo.Set("2fa_issuer", issuer)
}

func (s *TwoFactorService) GenerateSecret(userID, email string) (string, string, error) {
	issuer, _ := s.getIssuer()
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: email,
	})
	if err != nil {
		return "", "", err
	}

	if err := s.totpRepo.Upsert(userID, key.Secret()); err != nil {
		return "", "", err
	}

	return key.Secret(), key.URL(), nil
}

func (s *TwoFactorService) VerifySetup(userID, code string) error {
	totpRecord, err := s.totpRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("2fa not set up")
	}

	if !totp.Validate(code, totpRecord.Secret) {
		return fmt.Errorf("invalid code")
	}

	return s.totpRepo.Enable(userID)
}

func (s *TwoFactorService) VerifyCode(userID, code string) error {
	totpRecord, err := s.totpRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("2fa not set up")
	}
	if !totpRecord.Enabled {
		return fmt.Errorf("2fa not enabled")
	}

	if !totp.Validate(code, totpRecord.Secret) {
		return fmt.Errorf("invalid code")
	}
	return nil
}

func (s *TwoFactorService) HasEnabled2FA(userID string) bool {
	t, err := s.totpRepo.GetByUserID(userID)
	if err != nil {
		return false
	}
	return t.Enabled
}

func (s *TwoFactorService) GenerateTempToken(user *model.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "2fa_pending",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *TwoFactorService) ValidateTempToken(tokenString string) (*model.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid || claims.Subject != "2fa_pending" {
		return nil, fmt.Errorf("invalid 2fa token")
	}
	return &model.User{
		ID:    claims.UserID,
		Email: claims.Email,
		Role:  claims.Role,
	}, nil
}

func (s *TwoFactorService) getIssuer() (string, error) {
	configs, err := s.cfgRepo.GetAll()
	if err != nil || configs["2fa_issuer"] == "" {
		return "ChatAgent", nil
	}
	return configs["2fa_issuer"], nil
}
