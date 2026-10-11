package seed

import (
	"context"
	"log"
	"time"

	"github.com/driver-hub/driver-hub/internal/auth"
	"github.com/driver-hub/driver-hub/internal/domain"
	"github.com/driver-hub/driver-hub/internal/repository/postgres"
)

const DemoPassword = "demo1234"

// Run создаёт демо-компании, водителей и насыщенные досье для презентации.
func Run(ctx context.Context, store *postgres.Store) error {
	companies, err := ensureCompanies(ctx, store)
	if err != nil {
		return err
	}
	drivers, err := ensureDrivers(ctx, store)
	if err != nil {
		return err
	}
	if len(companies) == 0 || len(drivers) == 0 {
		return nil
	}
	return seedDossiers(ctx, store, companies, drivers)
}

type companySeed struct {
	Name, INN, OGRN, Email, Phone string
}

type driverSeed struct {
	HubID, Email, Phone, Last, First, Middle, SNILS, INN, ESIAOID string
	Birth                                                         time.Time
	Experience                                                    int
	Categories                                                    []string
	Status                                                        domain.VerificationStatus
	LicenseSeries, LicenseNumber, LicenseIssuer                   string
	LicenseIssue, LicenseExpiry                                   time.Time
	WithMedical, WithCriminalClear, WithSuspension                bool
	SuspensionReason                                              string
}

func ensureCompanies(ctx context.Context, store *postgres.Store) ([]*domain.Company, error) {
	hash, err := auth.HashPassword(DemoPassword)
	if err != nil {
		return nil, err
	}
	defs := []companySeed{
		{"ООО Логистика Плюс", "7701234567", "1027700132195", "hr@logplus.ru", "+74951234567"},
		{"ООО ФастТрак Логистик", "7709876543", "1157746123456", "hr@fasttruck.ru", "+74957654321"},
		{"АО СеверТранс", "7801122334", "1027800001111", "fleet@severtrans.ru", "+78121234567"},
	}
	out := make([]*domain.Company, 0, len(defs))
	for _, def := range defs {
		co, err := store.GetCompanyByEmail(ctx, def.Email)
		if err != nil {
			return nil, err
		}
		if co != nil {
			_ = store.UpdateCompanyPassword(ctx, co.ID, hash)
			out = append(out, co)
			continue
		}
		co = &domain.Company{
			Name: def.Name, INN: def.INN, OGRN: def.OGRN,
			Email: def.Email, Phone: def.Phone, PasswordHash: hash, Verified: true,
		}
		if err := store.CreateCompany(ctx, co); err != nil {
			return nil, err
		}
		log.Printf("seed: company %s / %s", def.Email, DemoPassword)
		out = append(out, co)
	}
	return out, nil
}

func ensureDrivers(ctx context.Context, store *postgres.Store) ([]*domain.Driver, error) {
	hash, err := auth.HashPassword(DemoPassword)
	if err != nil {
		return nil, err
	}
	defs := []driverSeed{
		{
			HubID: "DH-DEMOTST2", Email: "driver@demo.ru", Phone: "+79007654321",
			Last: "Иванов", First: "Иван", Middle: "Иванович",
			Birth: time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
			SNILS: "123-456-789 00", INN: "500100732259", ESIAOID: "seed-ivanov",
			Experience: 8, Categories: []string{"B", "C"}, Status: domain.VerificationVerified,
			LicenseSeries: "99АВ", LicenseNumber: "654321",
			LicenseIssuer: "ГИБДД ГУ МВД по Московской области",
			LicenseIssue: time.Date(2018, 3, 20, 0, 0, 0, 0, time.UTC),
			LicenseExpiry: time.Date(2028, 3, 20, 0, 0, 0, 0, time.UTC),
			WithMedical: true, WithCriminalClear: true,
		},
		{
			HubID: "DH-PETROV01", Email: "petrov@demo.ru", Phone: "+79001112233",
			Last: "Петров", First: "Пётр", Middle: "Сергеевич",
			Birth: time.Date(1985, 11, 2, 0, 0, 0, 0, time.UTC),
			SNILS: "112-233-445 95", INN: "7707083893", ESIAOID: "seed-petrov",
			Experience: 12, Categories: []string{"B", "C", "CE"}, Status: domain.VerificationVerified,
			LicenseSeries: "77АА", LicenseNumber: "123456",
			LicenseIssuer: "ГИБДД ГУ МВД по г. Москве",
			LicenseIssue: time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC),
			LicenseExpiry: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			WithMedical: true, WithCriminalClear: true,
		},
		{
			HubID: "DH-SIDOROV2", Email: "sidorov@demo.ru", Phone: "+79005556677",
			Last: "Сидоров", First: "Алексей", Middle: "Викторович",
			Birth: time.Date(1988, 2, 28, 0, 0, 0, 0, time.UTC),
			SNILS: "145-667-890 12", INN: "503201234567", ESIAOID: "seed-sidorov",
			Experience: 15, Categories: []string{"C", "CE", "D"}, Status: domain.VerificationVerified,
			LicenseSeries: "50СС", LicenseNumber: "778899",
			LicenseIssuer: "ГИБДД МО",
			LicenseIssue: time.Date(2015, 1, 10, 0, 0, 0, 0, time.UTC),
			LicenseExpiry: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
			WithMedical: true, WithCriminalClear: true,
		},
		{
			HubID: "DH-KOZLOVA3", Email: "kozlova@demo.ru", Phone: "+79003334455",
			Last: "Козлова", First: "Анна", Middle: "Дмитриевна",
			Birth: time.Date(1994, 9, 8, 0, 0, 0, 0, time.UTC),
			SNILS: "198-765-432 10", INN: "771234567890", ESIAOID: "seed-kozlova",
			Experience: 5, Categories: []string{"B"}, Status: domain.VerificationVerified,
			LicenseSeries: "45КК", LicenseNumber: "112233",
			LicenseIssuer: "ГИБДД г. Санкт-Петербурга",
			LicenseIssue: time.Date(2019, 6, 15, 0, 0, 0, 0, time.UTC),
			LicenseExpiry: time.Date(2029, 6, 15, 0, 0, 0, 0, time.UTC),
			WithMedical: true, WithCriminalClear: true,
		},
		{
			HubID: "DH-NOVIKOV4", Email: "novikov@demo.ru", Phone: "+79009998877",
			Last: "Новиков", First: "Максим", Middle: "Игоревич",
			Birth: time.Date(1998, 4, 12, 0, 0, 0, 0, time.UTC),
			Experience: 1, Categories: []string{}, Status: domain.VerificationNone,
		},
		{
			HubID: "DH-ORLOV05", Email: "orlov@demo.ru", Phone: "+79002223344",
			Last: "Орлов", First: "Денис", Middle: "Андреевич",
			Birth: time.Date(1982, 12, 1, 0, 0, 0, 0, time.UTC),
			SNILS: "167-890-123 45", INN: "502309876543", ESIAOID: "seed-orlov",
			Experience: 18, Categories: []string{"B", "C"}, Status: domain.VerificationVerified,
			LicenseSeries: "77ОР", LicenseNumber: "445566",
			LicenseIssuer: "ГИБДД г. Москвы",
			LicenseIssue: time.Date(2014, 8, 5, 0, 0, 0, 0, time.UTC),
			LicenseExpiry: time.Date(2024, 8, 5, 0, 0, 0, 0, time.UTC),
			WithMedical: true, WithCriminalClear: false, WithSuspension: true,
			SuspensionReason: "Лишение права управления (ст. 12.8 КоАП), демо",
		},
	}

	out := make([]*domain.Driver, 0, len(defs))
	for _, def := range defs {
		d, err := upsertDriver(ctx, store, hash, def)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func upsertDriver(ctx context.Context, store *postgres.Store, hash string, def driverSeed) (*domain.Driver, error) {
	if existing, _ := store.GetDriverByHubID(ctx, def.HubID); existing != nil {
		_ = store.EnsureDriverLogin(ctx, def.HubID, def.Email, def.Phone, hash)
		full, err := store.GetDriverByID(ctx, existing.ID)
		return full, err
	}
	if byEmail, _, _ := store.GetDriverByEmail(ctx, def.Email); byEmail != nil {
		_ = store.UpdateDriverPassword(ctx, byEmail.ID, hash)
		full, err := store.GetDriverByID(ctx, byEmail.ID)
		return full, err
	}

	birth := def.Birth
	d := &domain.Driver{
		HubID: def.HubID, Email: def.Email, Phone: def.Phone,
		LastName: def.Last, FirstName: def.First, MiddleName: def.Middle, BirthDate: &birth,
		SNILS: def.SNILS, INN: def.INN, ExperienceYears: def.Experience,
		Categories: def.Categories, VerificationStatus: def.Status,
	}
	if err := store.CreateDriver(ctx, d, hash); err != nil {
		return nil, err
	}

	if def.Status == domain.VerificationVerified && def.ESIAOID != "" {
		_ = store.ApplyESIAPerson(ctx, d.ID, def.ESIAOID, def.First, def.Last, def.Middle,
			def.SNILS, def.INN, def.Phone, def.Email, &birth, def.Categories)
	}
	if def.LicenseNumber != "" {
		issue, expiry := def.LicenseIssue, def.LicenseExpiry
		_ = store.UpsertLicense(ctx, &domain.DriverLicense{
			DriverID: d.ID, Series: def.LicenseSeries, Number: def.LicenseNumber,
			IssueDate: &issue, ExpiryDate: &expiry, Categories: def.Categories,
			Issuer: def.LicenseIssuer, Source: "esia", Verified: true,
		})
	}
	if def.WithMedical {
		medIssued := time.Now().Add(-200 * 24 * time.Hour)
		medUntil := time.Now().Add(400 * 24 * time.Hour)
		_ = store.AddMedical(ctx, &domain.MedicalCertificate{
			DriverID: d.ID, Number: "МС-" + def.HubID, IssuedAt: &medIssued, ValidUntil: &medUntil,
			Clinic: "Медцентр Транспортный", Result: "fit", Categories: def.Categories, Source: "upload",
		})
	}
	if def.WithCriminalClear || def.WithSuspension {
		checked := time.Now().Add(-10 * 24 * time.Hour)
		_ = store.AddCriminal(ctx, &domain.CriminalRecord{
			DriverID: d.ID, HasRecord: !def.WithCriminalClear,
			Description: map[bool]string{true: "Судимостей нет (демо)", false: "Есть сведения — уточнить (демо)"}[def.WithCriminalClear],
			CheckedAt: &checked, Source: "self_declared",
		})
	}
	if def.WithSuspension {
		start := time.Now().Add(-60 * 24 * time.Hour)
		end := time.Now().Add(120 * 24 * time.Hour)
		_ = store.AddSuspension(ctx, &domain.LicenseSuspension{
			DriverID: d.ID, Reason: def.SuspensionReason,
			StartDate: &start, EndDate: &end, Active: true,
			Authority: "Суд г. Москвы", CaseNumber: "5-1234/2025", Source: "manual",
		})
	}

	log.Printf("seed: driver %s / %s hub=%s", def.Email, DemoPassword, def.HubID)
	full, err := store.GetDriverByID(ctx, d.ID)
	return full, err
}

func seedDossiers(ctx context.Context, store *postgres.Store, companies []*domain.Company, drivers []*domain.Driver) error {
	byHub := map[string]*domain.Driver{}
	for _, d := range drivers {
		byHub[d.HubID] = d
	}
	coPlus, coFast, coSever := companies[0], companies[1], companies[2]

	// Маркер: если у Петрова ещё нет жалоб — льём полный демо-набор (один раз).
	petrov := byHub["DH-PETROV01"]
	if petrov == nil {
		return nil
	}
	if n, _ := store.CountComplaints(ctx, petrov.ID); n > 0 {
		return nil
	}

	if d := byHub["DH-DEMOTST2"]; d != nil {
		seedIvanov(ctx, store, d, coPlus, coFast, coSever)
	}
	seedPetrov(ctx, store, petrov, coPlus, coFast, coSever)
	if d := byHub["DH-SIDOROV2"]; d != nil {
		seedSidorov(ctx, store, d, coPlus, coFast, coSever)
	}
	if d := byHub["DH-KOZLOVA3"]; d != nil {
		seedKozlova(ctx, store, d, coPlus, coSever)
	}
	if d := byHub["DH-ORLOV05"]; d != nil {
		seedOrlov(ctx, store, d, coFast, coSever)
	}
	log.Printf("seed: dossier demo data loaded")
	return nil
}

func seedIvanov(ctx context.Context, store *postgres.Store, d *domain.Driver, a, b, c *domain.Company) {
	_ = store.AddRecommendation(ctx, &domain.Recommendation{
		DriverID: d.ID, CompanyID: a.ID, Text: "Ответственный водитель, рекомендуем к найму на междугородние рейсы", Rating: 5,
	})
	_ = store.AddRecommendation(ctx, &domain.Recommendation{
		DriverID: d.ID, CompanyID: b.ID, Text: "Аккуратный стиль вождения, без срывов сроков за 14 месяцев", Rating: 5,
	})
	_ = store.AddRecommendation(ctx, &domain.Recommendation{
		DriverID: d.ID, CompanyID: c.ID, Text: "Хорошо знает маршруты ЦФО, коммуникабелен с клиентами", Rating: 4,
	})
	_ = store.AddComplaint(ctx, &domain.Complaint{
		DriverID: d.ID, CompanyID: a.ID, Category: "discipline",
		Text: "Единичное опоздание на погрузку (демо)", Severity: "low",
	})
	cid := a.ID
	_ = store.AddAccident(ctx, &domain.Accident{
		DriverID: d.ID, CompanyID: &cid, OccurredAt: time.Now().Add(-90 * 24 * time.Hour),
		Description: "Лёгкое касание на стоянке у склада", Fault: "mutual", DamageLevel: "minor", Location: "Москва, Южные Ворота",
	})
	issued := time.Now().Add(-30 * 24 * time.Hour)
	_ = store.AddFine(ctx, &domain.Fine{
		DriverID: d.ID, CompanyID: &cid, Article: "12.9 КоАП", Amount: 5000,
		IssuedAt: &issued, Paid: true, Description: "Превышение скорости на М-7", Source: "gibdd",
	})
	issued2 := time.Now().Add(-120 * 24 * time.Hour)
	_ = store.AddFine(ctx, &domain.Fine{
		DriverID: d.ID, CompanyID: &cid, Article: "12.6 КоАП", Amount: 1500,
		IssuedAt: &issued2, Paid: true, Description: "Непристёгнутый ремень", Source: "gibdd",
	})
}

func seedPetrov(ctx context.Context, store *postgres.Store, d *domain.Driver, a, b, c *domain.Company) {
	_ = store.AddRecommendation(ctx, &domain.Recommendation{
		DriverID: d.ID, CompanyID: a.ID, Text: "Опытный водитель CE, знает длинные плечи", Rating: 3,
	})
	_ = store.AddComplaint(ctx, &domain.Complaint{
		DriverID: d.ID, CompanyID: b.ID, Category: "safety",
		Text: "Грубое нарушение режима труда и отдыха", Severity: "high",
	})
	_ = store.AddComplaint(ctx, &domain.Complaint{
		DriverID: d.ID, CompanyID: c.ID, Category: "fraud",
		Text: "Расхождение по одометру и топливным картам (проверка)", Severity: "medium",
	})
	_ = store.UpsertBlacklist(ctx, &domain.BlacklistEntry{
		DriverID: d.ID, CompanyID: b.ID,
		Reason: "Систематические нарушения ТБ и конфликт с диспетчером", Active: true,
	})
	cid := c.ID
	_ = store.AddAccident(ctx, &domain.Accident{
		DriverID: d.ID, CompanyID: &cid, OccurredAt: time.Now().Add(-200 * 24 * time.Hour),
		Description: "Столкновение на кольце, повреждения кузова", Fault: "driver", DamageLevel: "major", Location: "МО, Одинцово",
	})
	issued := time.Now().Add(-15 * 24 * time.Hour)
	_ = store.AddFine(ctx, &domain.Fine{
		DriverID: d.ID, CompanyID: &cid, Article: "12.15 КоАП", Amount: 5000,
		IssuedAt: &issued, Paid: false, Description: "Выезд на встречную", Source: "gibdd",
	})
}

func seedSidorov(ctx context.Context, store *postgres.Store, d *domain.Driver, a, b, c *domain.Company) {
	_ = store.AddRecommendation(ctx, &domain.Recommendation{
		DriverID: d.ID, CompanyID: c.ID, Text: "Сильный водитель на фурах, знает северные направления", Rating: 4,
	})
	_ = store.AddRecommendation(ctx, &domain.Recommendation{
		DriverID: d.ID, CompanyID: a.ID, Text: "Надёжен на регулярных рейсах Москва—СПб", Rating: 5,
	})
	cid := a.ID
	_ = store.AddAccident(ctx, &domain.Accident{
		DriverID: d.ID, CompanyID: &cid, OccurredAt: time.Now().Add(-400 * 24 * time.Hour),
		Description: "ДТП на трассе из‑за гололёда, без пострадавших", Fault: "other", DamageLevel: "minor", Location: "М-10, Тверская обл.",
	})
	_ = store.AddAccident(ctx, &domain.Accident{
		DriverID: d.ID, CompanyID: &cid, OccurredAt: time.Now().Add(-40 * 24 * time.Hour),
		Description: "Наезд на бордюр при маневрировании на площадке", Fault: "driver", DamageLevel: "minor", Location: "СПб, Шушары",
	})
	for i, art := range []struct {
		Article string
		Amount  float64
		Paid    bool
		Desc    string
		DaysAgo int
	}{
		{"12.9 КоАП", 2500, true, "Превышение 20–40 км/ч", 25},
		{"12.9 КоАП", 5000, false, "Превышение 40–60 км/ч", 8},
		{"12.5 КоАП", 500, true, "Неисправность освещения", 180},
	} {
		_ = i
		issued := time.Now().Add(-time.Duration(art.DaysAgo) * 24 * time.Hour)
		_ = store.AddFine(ctx, &domain.Fine{
			DriverID: d.ID, CompanyID: &cid, Article: art.Article, Amount: art.Amount,
			IssuedAt: &issued, Paid: art.Paid, Description: art.Desc, Source: "gibdd",
		})
	}
	_ = store.AddComplaint(ctx, &domain.Complaint{
		DriverID: d.ID, CompanyID: b.ID, Category: "other",
		Text: "Жалоба клиента на грубость при выгрузке", Severity: "low",
	})
}

func seedKozlova(ctx context.Context, store *postgres.Store, d *domain.Driver, a, c *domain.Company) {
	_ = store.AddRecommendation(ctx, &domain.Recommendation{
		DriverID: d.ID, CompanyID: a.ID, Text: "Отличная дисциплина на городских доставках", Rating: 5,
	})
	_ = store.AddRecommendation(ctx, &domain.Recommendation{
		DriverID: d.ID, CompanyID: c.ID, Text: "Аккуратно работает с документами и клиентами", Rating: 5,
	})
	issued := time.Now().Add(-200 * 24 * time.Hour)
	cid := a.ID
	_ = store.AddFine(ctx, &domain.Fine{
		DriverID: d.ID, CompanyID: &cid, Article: "12.16 КоАП", Amount: 1500,
		IssuedAt: &issued, Paid: true, Description: "Остановка под знаком", Source: "gibdd",
	})
}

func seedOrlov(ctx context.Context, store *postgres.Store, d *domain.Driver, b, c *domain.Company) {
	_ = store.AddComplaint(ctx, &domain.Complaint{
		DriverID: d.ID, CompanyID: b.ID, Category: "safety",
		Text: "Управление в состоянии опьянения — основание для расторжения", Severity: "high",
	})
	_ = store.UpsertBlacklist(ctx, &domain.BlacklistEntry{
		DriverID: d.ID, CompanyID: b.ID,
		Reason: "Лишение ВУ, недопуск к рейсам", Active: true,
	})
	_ = store.UpsertBlacklist(ctx, &domain.BlacklistEntry{
		DriverID: d.ID, CompanyID: c.ID,
		Reason: "Отказ в найме по результатам проверки", Active: true,
	})
	cid := b.ID
	_ = store.AddAccident(ctx, &domain.Accident{
		DriverID: d.ID, CompanyID: &cid, OccurredAt: time.Now().Add(-70 * 24 * time.Hour),
		Description: "ДТП при выезде со двора", Fault: "driver", DamageLevel: "major", Location: "Москва",
	})
}
