package vision

import (
	"context"
)

type MockScanner struct{}

func NewMockScanner() *MockScanner { return &MockScanner{} }

func (m *MockScanner) Mode() string { return "mock" }

func (m *MockScanner) ScanLicense(ctx context.Context, front, back *ImageInput) (*LicenseDraft, error) {
	_ = ctx
	_ = front
	_ = back
	return &LicenseDraft{
		LastName:   "ИВАНОВ",
		FirstName:  "ИВАН",
		MiddleName: "ИВАНОВИЧ",
		BirthDate:  "1990-05-15",
		Series:     "99АВ",
		Number:     "654321",
		IssueDate:  "2018-03-20",
		ExpiryDate: "2028-03-20",
		Categories: []string{"B", "C"},
		Issuer:     "ГИБДД ГУ МВД России по Московской области",
		Confidence: 0.92,
		Notes:      "",
		Source:     "mock",
	}, nil
}
