package esia

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// LiveClient — каркас боевого клиента ЕСИА (OAuth 2.0 / OpenID Connect).
// В продуктиве запросы к /aas/oauth2/ac должны подписываться алгоритмом ГОСТ Р 34.10-2012
// (см. docs/ESIA.md). Здесь реализована RSA-подпись для стендов/шлюзов, принимающих RS256.
type LiveClient struct {
	baseURL      string
	clientID     string
	clientSecret string
	redirectURI  string
	scopes       []string
	httpClient   *http.Client
	privateKey   *rsa.PrivateKey
}

func NewLiveClient(baseURL, clientID, clientSecret, redirectURI string, scopes []string, keyPath string) (*LiveClient, error) {
	c := &LiveClient{
		baseURL:      strings.TrimRight(baseURL, "/"),
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		scopes:       scopes,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
	if keyPath != "" {
		key, err := loadRSAPrivateKey(keyPath)
		if err != nil {
			return nil, fmt.Errorf("esia key: %w", err)
		}
		c.privateKey = key
	}
	return c, nil
}

func (c *LiveClient) Mode() string { return "live" }

func (c *LiveClient) AuthURL(state, nonce string) string {
	scope := strings.Join(c.scopes, " ")
	params := url.Values{}
	params.Set("client_id", c.clientID)
	params.Set("client_secret", c.clientSecret)
	params.Set("redirect_uri", c.redirectURI)
	params.Set("scope", scope)
	params.Set("response_type", "code")
	params.Set("state", state)
	params.Set("nonce", nonce)
	params.Set("access_type", "offline")
	params.Set("timestamp", time.Now().Format("2006.01.02 15:04:05 -0700"))

	clientSecret := c.clientSecret
	if c.privateKey != nil {
		if sig, err := c.sign(params.Encode()); err == nil {
			clientSecret = sig
		}
	}
	params.Set("client_secret", clientSecret)

	return fmt.Sprintf("%s/aas/oauth2/ac?%s", c.baseURL, params.Encode())
}

func (c *LiveClient) ExchangeCode(ctx context.Context, code string) (*TokenSet, error) {
	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", c.redirectURI)
	form.Set("token_type", "Bearer")
	form.Set("scope", strings.Join(c.scopes, " "))
	form.Set("timestamp", time.Now().Format("2006.01.02 15:04:05 -0700"))
	form.Set("state", "noop")
	form.Set("client_secret", c.clientSecret)

	if c.privateKey != nil {
		if sig, err := c.sign(form.Encode()); err == nil {
			form.Set("client_secret", sig)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/aas/oauth2/te", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("esia token exchange failed: %s", string(body))
	}

	var raw struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	return &TokenSet{
		AccessToken:  raw.AccessToken,
		RefreshToken: raw.RefreshToken,
		IDToken:      raw.IDToken,
		ExpiresIn:    raw.ExpiresIn,
		Scope:        raw.Scope,
	}, nil
}

func (c *LiveClient) FetchPerson(ctx context.Context, accessToken string) (*PersonData, error) {
	oid, err := extractOID(accessToken)
	if err != nil {
		return nil, err
	}

	personURL := fmt.Sprintf("%s/rs/prns/%s?embed=(contacts.elements,documents.elements)", c.baseURL, oid)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, personURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("esia person fetch failed: %s", string(body))
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	person := &PersonData{
		OID:     fmt.Sprint(raw["oid"]),
		RawJSON: string(body),
		Trusted: fmt.Sprint(raw["trusted"]) == "true" || fmt.Sprint(raw["stateFacts"]) != "",
	}
	if v, ok := raw["firstName"].(string); ok {
		person.FirstName = v
	}
	if v, ok := raw["lastName"].(string); ok {
		person.LastName = v
	}
	if v, ok := raw["middleName"].(string); ok {
		person.MiddleName = v
	}
	if v, ok := raw["snils"].(string); ok {
		person.SNILS = v
	}
	if v, ok := raw["inn"].(string); ok {
		person.INN = v
	}
	if v, ok := raw["birthDate"].(string); ok {
		if t, err := time.Parse("02.01.2006", v); err == nil {
			person.BirthDate = &t
		}
	}

	// Документы и контакты приходят во вложенных коллекциях — упрощённый разбор.
	person.License = extractLicense(raw)
	return person, nil
}

func (c *LiveClient) sign(payload string) (string, error) {
	if c.privateKey == nil {
		return "", fmt.Errorf("no private key")
	}
	h := sha256.Sum256([]byte(payload))
	sig, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, h[:])
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(sig), nil
}

func extractOID(accessToken string) (string, error) {
	parts := strings.Split(accessToken, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid access token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		payload, err = base64.URLEncoding.DecodeString(parts[1])
		if err != nil {
			return "", err
		}
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", err
	}
	if urn, ok := claims["urn:esia:sbj"].(map[string]any); ok {
		if oid, ok := urn["urn:esia:sbj:oid"].(string); ok {
			return oid, nil
		}
		if oid, ok := urn["urn:esia:sbj:oid"].(float64); ok {
			return fmt.Sprintf("%.0f", oid), nil
		}
	}
	if sub, ok := claims["sub"].(string); ok {
		return sub, nil
	}
	return "", fmt.Errorf("oid not found in token")
}

func extractLicense(raw map[string]any) *LicenseData {
	docs, ok := raw["documents"].(map[string]any)
	if !ok {
		return nil
	}
	elements, ok := docs["elements"].([]any)
	if !ok {
		return nil
	}
	for _, el := range elements {
		m, ok := el.(map[string]any)
		if !ok {
			continue
		}
		typ := fmt.Sprint(m["type"])
		if typ != "RF_DRIVING_LICENSE" && typ != "DL" {
			continue
		}
		lic := &LicenseData{}
		if v, ok := m["series"].(string); ok {
			lic.Series = v
		}
		if v, ok := m["number"].(string); ok {
			lic.Number = v
		}
		if v, ok := m["issueDate"].(string); ok {
			if t, err := time.Parse("02.01.2006", v); err == nil {
				lic.IssueDate = &t
			}
		}
		if v, ok := m["expiryDate"].(string); ok {
			if t, err := time.Parse("02.01.2006", v); err == nil {
				lic.ExpiryDate = &t
			}
		}
		return lic
	}
	return nil
}

func loadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block in %s", path)
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return key, nil
	}
	parsed, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err2 != nil {
		return nil, err
	}
	rsaKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}
	return rsaKey, nil
}
