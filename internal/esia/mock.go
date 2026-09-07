package esia

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MockClient эмулирует ЕСИА для локальной разработки без сертификатов ГОСТ.
type MockClient struct {
	publicBase string
	mu         sync.Mutex
	codes      map[string]*PersonData
}

func NewMockClient(publicBase string) *MockClient {
	return &MockClient{
		publicBase: strings.TrimRight(publicBase, "/"),
		codes:      make(map[string]*PersonData),
	}
}

func (m *MockClient) Mode() string { return "mock" }

func (m *MockClient) AuthURL(state, nonce string) string {
	u := fmt.Sprintf("%s/mock-esia/login?state=%s&nonce=%s", m.publicBase, url.QueryEscape(state), url.QueryEscape(nonce))
	return u
}

// IssueDemoCode выдаёт код авторизации с предустановленным профилем (для UI/тестов).
func (m *MockClient) IssueDemoCode(preset string) string {
	return m.IssueDemoCodeForDriver(preset, "")
}

// IssueDemoCodeForDriver — уникальный OID на водителя, чтобы не бить unique esia_oid.
func (m *MockClient) IssueDemoCodeForDriver(preset, driverID string) string {
	person := demoPerson(preset)
	if driverID != "" {
		p := *person
		p.OID = "mock-" + driverID
		if p.License != nil {
			lic := *p.License
			p.License = &lic
		}
		person = &p
	}
	code := "mock-" + uuid.NewString()
	m.mu.Lock()
	m.codes[code] = person
	m.mu.Unlock()
	return code
}

func (m *MockClient) ExchangeCode(ctx context.Context, code string) (*TokenSet, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.codes[code]; !ok {
		// Автосоздание для прямых callback-ов с demo-кодом.
		m.codes[code] = demoPerson("ivanov")
	}
	return &TokenSet{
		AccessToken:  "mock-access-" + code,
		RefreshToken: "mock-refresh-" + code,
		ExpiresIn:    3600,
		Scope:        strings.Join(RecommendedScopes, " "),
	}, nil
}

func (m *MockClient) FetchPerson(ctx context.Context, accessToken string) (*PersonData, error) {
	_ = ctx
	code := strings.TrimPrefix(accessToken, "mock-access-")
	m.mu.Lock()
	defer m.mu.Unlock()
	person, ok := m.codes[code]
	if !ok {
		person = demoPerson("ivanov")
	}
	raw, _ := json.Marshal(person)
	clone := *person
	clone.RawJSON = string(raw)
	return &clone, nil
}

func demoPerson(preset string) *PersonData {
	birth := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	issue := time.Date(2018, 3, 20, 0, 0, 0, 0, time.UTC)
	expiry := time.Date(2028, 3, 20, 0, 0, 0, 0, time.UTC)

	switch strings.ToLower(preset) {
	case "petrov":
		birth = time.Date(1985, 11, 2, 0, 0, 0, 0, time.UTC)
		return &PersonData{
			OID: "1000486401", FirstName: "Пётр", LastName: "Петров", MiddleName: "Сергеевич",
			BirthDate: &birth, Gender: "M", SNILS: "112-233-445 95", INN: "7707083893",
			Mobile: "+79001234567", Email: "petrov@example.ru", Trusted: true,
			License: &LicenseData{
				Series: "77АА", Number: "123456", IssueDate: &issue, ExpiryDate: &expiry,
				Categories: []string{"B", "C", "CE"}, Issuer: "ГИБДД ГУ МВД России по г. Москве",
			},
			Passport: &PassportData{Series: "4509", Number: "123456", IssueDate: &issue, IssuedBy: "ОВД Тверской"},
		}
	default:
		return &PersonData{
			OID: "1000486400", FirstName: "Иван", LastName: "Иванов", MiddleName: "Иванович",
			BirthDate: &birth, Gender: "M", SNILS: "123-456-789 00", INN: "500100732259",
			Mobile: "+79007654321", Email: "ivanov@example.ru", Trusted: true,
			License: &LicenseData{
				Series: "99АВ", Number: "654321", IssueDate: &issue, ExpiryDate: &expiry,
				Categories: []string{"B", "C"}, Issuer: "ГИБДД ГУ МВД России по Московской области",
			},
			Passport: &PassportData{Series: "4510", Number: "654321", IssueDate: &issue, IssuedBy: "ОВД Мытищи"},
		}
	}
}
