package esia

import (
	"context"
	"time"
)

// Client — абстракция над ЕСИА / Цифровым профилем.
type Client interface {
	// AuthURL формирует URL редиректа на Госуслуги.
	AuthURL(state, nonce string) string
	// ExchangeCode обменивает authorization code на токены и профиль.
	ExchangeCode(ctx context.Context, code string) (*TokenSet, error)
	// FetchPerson загружает персональные данные по access_token.
	FetchPerson(ctx context.Context, accessToken string) (*PersonData, error)
	Mode() string
}

type TokenSet struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	ExpiresIn    int
	Scope        string
}

type PersonData struct {
	OID        string
	FirstName  string
	LastName   string
	MiddleName string
	BirthDate  *time.Time
	Gender     string
	SNILS      string
	INN        string
	Mobile     string
	Email      string
	Trusted    bool // подтверждённая УЗ
	License    *LicenseData
	Passport   *PassportData
	RawJSON    string
}

type LicenseData struct {
	Series     string
	Number     string
	IssueDate  *time.Time
	ExpiryDate *time.Time
	Categories []string
	Issuer     string
}

type PassportData struct {
	Series    string
	Number    string
	IssueDate *time.Time
	IssuedBy  string
}

// Scopes, которые нужны Driver Hub для верификации водителя.
// Полный перечень запрашивается в заявке на подключение ИС к ЕСИА.
var RecommendedScopes = []string{
	"openid",
	"fullname",
	"birthdate",
	"gender",
	"snils",
	"inn",
	"mobile",
	"email",
	"id_doc",              // паспорт
	"drivers_licence_doc", // водительское удостоверение
}

// DigitalProfileDocuments — документы через ЦПГ (при наличии правовых оснований).
var DigitalProfileDocuments = []string{
	"GIBDD_DRIVER_LICENSE", // сведения о ВУ из ГИБДД
	"RF_DRIVING_LICENSE",   // ВУ из цифрового профиля
	"RF_PASSPORT",
}
