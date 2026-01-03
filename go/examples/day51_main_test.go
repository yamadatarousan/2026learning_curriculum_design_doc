package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTCreationAndValidation(t *testing.T) {
	// --- 1. テスト用のクレーム（JWTの中身）を作成 ---
	userID := 99
	userRole := "admin"
	originalClaims := AppClaims{
		userRole,
		jwt.RegisteredClaims{
			Subject:   fmt.Sprint(userID),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
		},
	}

	// --- 2. クレームを使ってトークンを生成 ---
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, originalClaims)
	tokenString, err := token.SignedString(jwtSecret)

	// トークン生成でエラーが発生してはいけない
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// --- 3. 生成したトークン文字列を検証・解析 ---
	parsedToken, err := jwt.ParseWithClaims(tokenString, &AppClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	// トークン解析でエラーが発生してはいけない
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	// トークンが無効であってはいけない
	if !parsedToken.Valid {
		t.Errorf("Token is not valid")
	}

	// --- 4. 解析したクレームの内容が、元の内容と一致するか確認 ---
	parsedClaims, ok := parsedToken.Claims.(*AppClaims)
	if !ok {
		t.Fatalf("Could not parse claims to AppClaims")
	}

	// Subject (ユーザーID) の確認
	if parsedClaims.Subject != fmt.Sprint(userID) {
		t.Errorf("Expected subject %d, but got %s", userID, parsedClaims.Subject)
	}

	// Role (役割) の確認
	if parsedClaims.Role != userRole {
		t.Errorf("Expected role %s, but got %s", userRole, parsedClaims.Role)
	}
}
