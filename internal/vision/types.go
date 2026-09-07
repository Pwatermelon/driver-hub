package vision

import (
	"context"
	"time"
)

// LicenseDraft — распознанные поля ВУ до подтверждения пользователем.
type LicenseDraft struct {
	LastName   string   `json:"last_name"`
	FirstName  string   `json:"first_name"`
	MiddleName string   `json:"middle_name"`
	BirthDate  string   `json:"birth_date"` // YYYY-MM-DD
	Series     string   `json:"series"`
	Number     string   `json:"number"`
	IssueDate  string   `json:"issue_date"`
	ExpiryDate string   `json:"expiry_date"`
	Categories []string `json:"categories"`
	Issuer     string   `json:"issuer"`
	Confidence float64  `json:"confidence"` // 0..1
	Notes      string   `json:"notes,omitempty"`
	Source     string   `json:"source"` // openai | mock
}

type ImageInput struct {
	Filename    string
	ContentType string
	Data        []byte
}

type Scanner interface {
	ScanLicense(ctx context.Context, front, back *ImageInput) (*LicenseDraft, error)
	Mode() string
}

func ParseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02", "02.01.2006", "02/01/2006"} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}
