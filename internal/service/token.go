package service

import (
	ctx "context"
	"encoding/json"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/middleware/jwt"
	"github.com/ujum/dictap/internal/config"
	"github.com/ujum/dictap/internal/domain"
	"github.com/ujum/dictap/internal/dto"
	"github.com/ujum/dictap/pkg/logger"
	"net/http"
	"time"
)

const (
	privateKeyPath = "/token/rsa_private_key.pem"
	publicKeyPath  = "/token/rsa_public_key.pem"
	appName        = "dictup"
)

type TokenService interface {
	Generate(requestCtx ctx.Context, credentials *dto.UserCredentials) (*dto.TokenDTO, error)
	GenerateForUser(user *domain.User) (*dto.TokenDTO, error)
	Refresh(requestCtx ctx.Context, refreshToken json.RawMessage) (*dto.TokenDTO, error)
}

type JwtTokenService struct {
	log         logger.Logger
	cfg         *config.Config
	signer      *jwt.Signer
	verifier    *jwt.Verifier
	userService UserService
}

func newJwtSigner(cfg *config.Config) (*jwt.Signer, error) {
	privateKeyRSA, err := jwt.LoadPrivateKeyRSA(cfg.ConfigDir + privateKeyPath)
	if err != nil {
		return nil, err
	}
	min := cfg.Server.Security.ApiKeyAuth.AccessTokenMaxAgeMin
	signer := jwt.NewSigner(jwt.RS256, privateKeyRSA, time.Duration(min)*time.Minute)
	return signer, err
}

func newJwtVerifier(cfg *config.Config) (*jwt.Verifier, error) {
	publicKeyRSA, err := jwt.LoadPublicKeyRSA(cfg.ConfigDir + publicKeyPath)
	if err != nil {
		return nil, err
	}
	verifier := jwt.NewVerifier(jwt.RS256, publicKeyRSA)
	verifier.WithDefaultBlocklist()
	verifier.ErrorHandler = func(ctx *context.Context, err error) {
		ctx.StopWithJSON(http.StatusUnauthorized, map[string]string{"message": err.Error()})
	}
	return verifier, nil
}

func newJwtTokenService(cfg *config.Config, appLogger logger.Logger,
	verifier *jwt.Verifier, signer *jwt.Signer, userService UserService) *JwtTokenService {
	tokenService := &JwtTokenService{
		log:         appLogger,
		cfg:         cfg,
		userService: userService,
		signer:      signer,
		verifier:    verifier,
	}
	return tokenService
}

func (tokenSrv *JwtTokenService) Generate(requestCtx ctx.Context, credentials *dto.UserCredentials) (*dto.TokenDTO, error) {
	user, err := tokenSrv.userService.GetByCredentials(requestCtx, credentials)
	if err != nil {
		return nil, err
	}
	return tokenSrv.GenerateForUser(user)
}

func (tokenSrv *JwtTokenService) Refresh(requestCtx ctx.Context, refreshToken json.RawMessage) (*dto.TokenDTO, error) {
	verifiedToken, err := tokenSrv.verifier.VerifyToken(refreshToken)
	if err != nil {
		tokenSrv.log.Errorf("verify refresh token error: %v", err)
		return nil, err
	}
	user, err := tokenSrv.userService.GetByUID(requestCtx, verifiedToken.StandardClaims.Subject)
	if err != nil {
		return nil, err
	}

	return tokenSrv.GenerateForUser(user)
}

func (tokenSrv *JwtTokenService) GenerateForUser(user *domain.User) (*dto.TokenDTO, error) {
	refreshClaims := jwt.Claims{Subject: user.UID}
	accessClaims := &context.SimpleUser{
		Authorization: "Bearer",
		AuthorizedAt:  time.Now(),
		ID:            user.UID,
		Username:      user.Name,
		Email:         user.Email,
		Fields:        map[string]interface{}{"app": appName},
		Roles:         user.Roles,
	}
	refreshMin := tokenSrv.cfg.Server.Security.ApiKeyAuth.RefreshTokenMaxAgeMin
	tokenPair, err := tokenSrv.signer.NewTokenPair(accessClaims, refreshClaims, time.Duration(refreshMin)*time.Minute)
	if err != nil {
		tokenSrv.log.Errorf("token pair generation error: %v", err)
		return nil, err
	}
	return &dto.TokenDTO{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}
