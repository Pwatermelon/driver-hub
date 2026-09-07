package seed

import (
	"context"
	"log"
	"time"

	"github.com/driver-hub/driver-hub/internal/auth"
	"github.com/driver-hub/driver-hub/internal/domain"
	"github.com/driver-hub/driver-hub/internal/repository/postgres"
)

const (
	DemoCompanyEmail    = "hr@logplus.ru"
	DemoCompanyPassword = "demo1234"
	DemoDriverEmail     = "driver@demo.ru"
	DemoDriverPassword  = "demo1234"
	DemoHubID           = "DH-DEMOTST2"
)

// Run создаёт тестовую компанию и водителя, если их ещё нет.
func Run(ctx context.Context, store *postgres.Store) error {
	co, err := ensureCompany(ctx, store)
	if err != nil {
		return err
	}
	d, created, err := ensureDriver(ctx, store)
	if err != nil {
		return err
	}
	if co == nil || d == nil {
		return nil
	}

	if created {
		_ = store.AddRecommendation(ctx, &domain.Recommendation{
			DriverID: d.ID, CompanyID: co.ID,
			Text: "Ответственный водитель, рекомендуем к найму", Rating: 5,
		})
		_ = store.AddComplaint(ctx, &domain.Complaint{
			DriverID: d.ID, CompanyID: co.ID,
			Category: "discipline", Text: "Единичное опоздание (демо)", Severity: "low",
		})
		cid := co.ID
		_ = store.AddAccident(ctx, &domain.Accident{
			DriverID: d.ID, CompanyID: &cid, OccurredAt: time.Now().Add(-90 * 24 * time.Hour),
			Description: "Лёгкое касание на стоянке (демо)", Fault: "mutual", DamageLevel: "minor", Location: "Москва",
		})
		issued := time.Now().Add(-30 * 24 * time.Hour)
		_ = store.AddFine(ctx, &domain.Fine{
			DriverID: d.ID, CompanyID: &cid, Article: "12.9 КоАП", Amount: 5000,
			IssuedAt: &issued, Paid: true, Description: "Превышение скорости (демо)", Source: "company",
		})
	}
	return nil
}

func ensureCompany(ctx context.Context, store *postgres.Store) (*domain.Company, error) {
	hash, err := auth.HashPassword(DemoCompanyPassword)
	if err != nil {
		return nil, err
	}
	co, err := store.GetCompanyByEmail(ctx, DemoCompanyEmail)
	if err != nil {
		return nil, err
	}
	if co != nil {
		_ = store.UpdateCompanyPassword(ctx, co.ID, hash)
		return co, nil
	}
	co = &domain.Company{
		Name: "ООО Логистика Плюс (демо)", INN: "7701234567", OGRN: "1027700132195",
		Email: DemoCompanyEmail, Phone: "+74951234567", PasswordHash: hash, Verified: true,
	}
	if err := store.CreateCompany(ctx, co); err != nil {
		return nil, err
	}
	log.Printf("seed: company %s / %s", DemoCompanyEmail, DemoCompanyPassword)
	return co, nil
}

func ensureDriver(ctx context.Context, store *postgres.Store) (*domain.Driver, bool, error) {
	hash, err := auth.HashPassword(DemoDriverPassword)
	if err != nil {
		return nil, false, err
	}
	d, _, err := store.GetDriverByEmail(ctx, DemoDriverEmail)
	if err != nil {
		return nil, false, err
	}
	if d != nil {
		_ = store.UpdateDriverPassword(ctx, d.ID, hash)
		full, err := store.GetDriverByID(ctx, d.ID)
		return full, false, err
	}
	if existing, _ := store.GetDriverByHubID(ctx, DemoHubID); existing != nil {
		_ = store.UpdateDriverPassword(ctx, existing.ID, hash)
		return existing, false, nil
	}

	birth := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	d = &domain.Driver{
		HubID: DemoHubID, Email: DemoDriverEmail, Phone: "+79007654321",
		LastName: "Иванов", FirstName: "Иван", MiddleName: "Иванович", BirthDate: &birth,
		SNILS: "123-456-789 00", INN: "500100732259", ExperienceYears: 8,
		Categories: []string{"B", "C"}, VerificationStatus: domain.VerificationVerified,
	}
	if err := store.CreateDriver(ctx, d, hash); err != nil {
		return nil, false, err
	}
	_ = store.ApplyESIAPerson(ctx, d.ID, "1000486400", d.FirstName, d.LastName, d.MiddleName,
		d.SNILS, d.INN, d.Phone, d.Email, d.BirthDate, d.Categories)

	issue := time.Date(2018, 3, 20, 0, 0, 0, 0, time.UTC)
	expiry := time.Date(2028, 3, 20, 0, 0, 0, 0, time.UTC)
	_ = store.UpsertLicense(ctx, &domain.DriverLicense{
		DriverID: d.ID, Series: "99АВ", Number: "654321",
		IssueDate: &issue, ExpiryDate: &expiry, Categories: []string{"B", "C"},
		Issuer: "ГИБДД ГУ МВД России по Московской области", Source: "esia", Verified: true,
	})
	medIssued := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	medUntil := time.Date(2027, 1, 10, 0, 0, 0, 0, time.UTC)
	_ = store.AddMedical(ctx, &domain.MedicalCertificate{
		DriverID: d.ID, Number: "МС-2025-001", IssuedAt: &medIssued, ValidUntil: &medUntil,
		Clinic: "Медцентр Транспортный", Result: "fit", Categories: []string{"B", "C"}, Source: "upload",
	})
	checked := time.Now()
	_ = store.AddCriminal(ctx, &domain.CriminalRecord{
		DriverID: d.ID, HasRecord: false, Description: "Судимостей нет (демо)",
		CheckedAt: &checked, Source: "self_declared",
	})

	full, err := store.GetDriverByID(ctx, d.ID)
	log.Printf("seed: driver %s / %s hub=%s", DemoDriverEmail, DemoDriverPassword, DemoHubID)
	return full, true, err
}
