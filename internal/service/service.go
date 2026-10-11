package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/driver-hub/driver-hub/internal/auth"
	"github.com/driver-hub/driver-hub/internal/domain"
	"github.com/driver-hub/driver-hub/internal/esia"
	"github.com/driver-hub/driver-hub/internal/pkg/hubid"
	"github.com/driver-hub/driver-hub/internal/repository/postgres"
	"github.com/driver-hub/driver-hub/internal/vision"
	"github.com/google/uuid"
)

type Service struct {
	store         *postgres.Store
	tokens        *auth.TokenManager
	esia          esia.Client
	vision        vision.Scanner
	adminPassword string
}

func New(store *postgres.Store, tokens *auth.TokenManager, esiaClient esia.Client, visionScanner vision.Scanner, adminPassword string) *Service {
	return &Service{store: store, tokens: tokens, esia: esiaClient, vision: visionScanner, adminPassword: adminPassword}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func mapDBErr(err error, fallback string) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "companies_email"), strings.Contains(msg, "drivers_email"):
		return fmt.Errorf("email уже зарегистрирован")
	case strings.Contains(msg, "companies_inn"):
		return fmt.Errorf("компания с таким ИНН уже есть")
	case strings.Contains(msg, "duplicate key"), strings.Contains(msg, "unique constraint"):
		return fmt.Errorf("такая запись уже существует")
	default:
		return fmt.Errorf("%s", fallback)
	}
}

func (s *Service) RegisterCompany(ctx context.Context, name, inn, ogrn, email, phone, password string) (*domain.Company, string, error) {
	name = strings.TrimSpace(name)
	inn = strings.TrimSpace(inn)
	email = normalizeEmail(email)
	phone = strings.TrimSpace(phone)
	if name == "" || inn == "" || email == "" || password == "" {
		return nil, "", fmt.Errorf("заполните название, ИНН, email и пароль")
	}
	if utf8.RuneCountInString(password) < 6 {
		return nil, "", fmt.Errorf("пароль не короче 6 символов")
	}
	if existing, _ := s.store.GetCompanyByEmail(ctx, email); existing != nil {
		return nil, "", fmt.Errorf("email уже зарегистрирован")
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("не удалось сохранить пароль")
	}
	c := &domain.Company{
		Name: name, INN: inn, OGRN: ogrn, Email: email, Phone: phone, PasswordHash: hash,
		Verified: false, ModerationStatus: domain.ModerationPending,
	}
	if err := s.store.CreateCompany(ctx, c); err != nil {
		return nil, "", mapDBErr(err, "не удалось создать компанию")
	}
	token, err := s.tokens.Issue(c.ID, domain.RoleCompany, c.Email)
	return c, token, err
}

func (s *Service) LoginCompany(ctx context.Context, email, password string) (*domain.Company, string, error) {
	email = normalizeEmail(email)
	if email == "" || password == "" {
		return nil, "", fmt.Errorf("введите email и пароль")
	}
	c, err := s.store.GetCompanyByEmail(ctx, email)
	if err != nil || c == nil {
		return nil, "", fmt.Errorf("неверный email или пароль")
	}
	if !auth.CheckPassword(c.PasswordHash, password) {
		return nil, "", fmt.Errorf("неверный email или пароль")
	}
	token, err := s.tokens.Issue(c.ID, domain.RoleCompany, c.Email)
	return c, token, err
}

func (s *Service) RegisterDriver(ctx context.Context, email, phone, password string) (*domain.Driver, string, error) {
	email = normalizeEmail(email)
	phone = strings.TrimSpace(phone)
	if email == "" || password == "" {
		return nil, "", fmt.Errorf("введите email и пароль")
	}
	if utf8.RuneCountInString(password) < 6 {
		return nil, "", fmt.Errorf("пароль не короче 6 символов")
	}
	if existing, _, _ := s.store.GetDriverByEmail(ctx, email); existing != nil {
		return nil, "", fmt.Errorf("email уже зарегистрирован")
	}
	hid, err := hubid.Generate()
	if err != nil {
		return nil, "", err
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("не удалось сохранить пароль")
	}
	d := &domain.Driver{
		HubID: hid, Email: email, Phone: phone,
		Categories: []string{}, VerificationStatus: domain.VerificationNone,
	}
	if err := s.store.CreateDriver(ctx, d, hash); err != nil {
		return nil, "", mapDBErr(err, "не удалось создать водителя")
	}
	token, err := s.tokens.Issue(d.ID, domain.RoleDriver, d.Email)
	return d, token, err
}

func (s *Service) LoginDriver(ctx context.Context, email, password string) (*domain.Driver, string, error) {
	email = normalizeEmail(email)
	if email == "" || password == "" {
		return nil, "", fmt.Errorf("введите email и пароль")
	}
	d, hash, err := s.store.GetDriverByEmail(ctx, email)
	if err != nil || d == nil || hash == "" || !auth.CheckPassword(hash, password) {
		return nil, "", fmt.Errorf("неверный email или пароль")
	}
	token, err := s.tokens.Issue(d.ID, domain.RoleDriver, d.Email)
	return d, token, err
}

func (s *Service) StartESIA(ctx context.Context, driverID uuid.UUID) (string, error) {
	state := uuid.NewString()
	nonce := uuid.NewString()
	id := driverID
	if err := s.store.SaveESIASession(ctx, state, nonce, &id, 15*time.Minute); err != nil {
		return "", err
	}
	return s.esia.AuthURL(state, nonce), nil
}

func (s *Service) StartESIARegistration(ctx context.Context) (string, error) {
	state := uuid.NewString()
	nonce := uuid.NewString()
	if err := s.store.SaveESIASession(ctx, state, nonce, nil, 15*time.Minute); err != nil {
		return "", err
	}
	return s.esia.AuthURL(state, nonce), nil
}

func (s *Service) CompleteESIA(ctx context.Context, code, state string) (*domain.Driver, string, error) {
	_, driverID, err := s.store.PopESIASession(ctx, state)
	if err != nil {
		return nil, "", err
	}
	tokens, err := s.esia.ExchangeCode(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("обмен кода ЕСИА: %w", err)
	}
	person, err := s.esia.FetchPerson(ctx, tokens.AccessToken)
	if err != nil {
		return nil, "", fmt.Errorf("получение профиля ЕСИА: %w", err)
	}
	if !person.Trusted {
		return nil, "", fmt.Errorf("требуется подтверждённая учётная запись Госуслуг")
	}

	var driver *domain.Driver
	if driverID != nil {
		driver, err = s.store.GetDriverByID(ctx, *driverID)
		if err != nil || driver == nil {
			return nil, "", fmt.Errorf("водитель не найден")
		}
	} else if existing, _ := s.store.GetDriverByESIAOID(ctx, person.OID); existing != nil {
		driver = existing
	} else {
		hid, err := hubid.Generate()
		if err != nil {
			return nil, "", err
		}
		driver = &domain.Driver{
			HubID: hid, Email: person.Email, Phone: person.Mobile,
			FirstName: person.FirstName, LastName: person.LastName, MiddleName: person.MiddleName,
			VerificationStatus: domain.VerificationPending,
		}
		if person.License != nil {
			driver.Categories = person.License.Categories
		}
		if err := s.store.CreateDriver(ctx, driver, ""); err != nil {
			return nil, "", err
		}
	}

	cats := driver.Categories
	if person.License != nil && len(person.License.Categories) > 0 {
		cats = person.License.Categories
	}
	if existing, _ := s.store.GetDriverByESIAOID(ctx, person.OID); existing != nil && existing.ID != driver.ID {
		if s.esia.Mode() == "mock" {
			_ = s.store.ClearESIAOID(ctx, existing.ID)
		} else {
			return nil, "", fmt.Errorf("эта учётная запись Госуслуг уже привязана к другому профилю")
		}
	}
	if err := s.store.ApplyESIAPerson(ctx, driver.ID, person.OID, person.FirstName, person.LastName, person.MiddleName,
		person.SNILS, person.INN, person.Mobile, person.Email, person.BirthDate, cats); err != nil {
		return nil, "", err
	}

	if person.License != nil {
		lic := &domain.DriverLicense{
			DriverID: driver.ID, Series: person.License.Series, Number: person.License.Number,
			IssueDate: person.License.IssueDate, ExpiryDate: person.License.ExpiryDate,
			Categories: person.License.Categories, Issuer: person.License.Issuer,
			Source: "esia", Verified: true, RawESIAPayload: person.RawJSON,
		}
		if err := s.store.UpsertLicense(ctx, lic); err != nil {
			return nil, "", err
		}
	}

	driver, err = s.store.GetDriverByID(ctx, driver.ID)
	if err != nil {
		return nil, "", err
	}
	token, err := s.tokens.Issue(driver.ID, domain.RoleDriver, driver.Email)
	aid := driver.ID
	s.store.Audit(ctx, "driver", &aid, "esia_verified", "driver", &aid, `{"oid":"`+person.OID+`"}`)
	return driver, token, err
}

// MockESIAComplete — ускоренный путь для демо без браузерного редиректа.
func (s *Service) MockESIAComplete(ctx context.Context, driverID uuid.UUID, preset string) (*domain.Driver, string, error) {
	mock, ok := s.esia.(*esia.MockClient)
	if !ok {
		return nil, "", fmt.Errorf("доступно только в ESIA_MODE=mock")
	}
	code := mock.IssueDemoCodeForDriver(preset, driverID.String())
	state := uuid.NewString()
	id := driverID
	_ = s.store.SaveESIASession(ctx, state, uuid.NewString(), &id, 5*time.Minute)
	return s.CompleteESIA(ctx, code, state)
}

func (s *Service) GetCompany(ctx context.Context, id uuid.UUID) (*domain.Company, error) {
	return s.store.GetCompanyByID(ctx, id)
}

func (s *Service) GetDriver(ctx context.Context, id uuid.UUID) (*domain.Driver, error) {
	return s.store.GetDriverByID(ctx, id)
}

func (s *Service) PublicCard(ctx context.Context, hub string) (*domain.PublicDriverCard, error) {
	hub = hubid.Normalize(hub)
	if err := hubid.Validate(hub); err != nil {
		return nil, err
	}
	d, err := s.store.GetDriverByHubID(ctx, hub)
	if err != nil || d == nil {
		return nil, fmt.Errorf("водитель не найден")
	}
	bl, _ := s.store.CountActiveBlacklist(ctx, d.ID)
	cmp, _ := s.store.CountComplaints(ctx, d.ID)
	acc, _ := s.store.CountAccidents(ctx, d.ID)
	susp, _ := s.store.HasActiveSuspension(ctx, d.ID)
	return &domain.PublicDriverCard{
		HubID: hub, FullNameMasked: maskName(d.LastName, d.FirstName, d.MiddleName),
		Categories: d.Categories, ExperienceYears: d.ExperienceYears,
		VerificationStatus: d.VerificationStatus, HasActiveSuspension: susp,
		BlacklistCount: bl, ComplaintCount: cmp, AccidentCount: acc,
	}, nil
}

func (s *Service) GrantAccess(ctx context.Context, driverID, companyID uuid.UUID, purpose string, days int) (*domain.AccessGrant, error) {
	if purpose == "" {
		purpose = "employment"
	}
	g := &domain.AccessGrant{DriverID: driverID, CompanyID: companyID, Purpose: purpose}
	if days > 0 {
		exp := time.Now().Add(time.Duration(days) * 24 * time.Hour)
		g.ExpiresAt = &exp
	}
	if err := s.store.UpsertGrant(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *Service) RevokeAccess(ctx context.Context, driverID, companyID uuid.UUID) error {
	return s.store.RevokeGrant(ctx, driverID, companyID)
}

func (s *Service) ListGrants(ctx context.Context, driverID uuid.UUID) ([]domain.AccessGrant, error) {
	return s.store.ListGrantsByDriver(ctx, driverID)
}

func (s *Service) GetDossier(ctx context.Context, companyID uuid.UUID, hub string) (*domain.DriverDossier, error) {
	hub = hubid.Normalize(hub)
	d, err := s.store.GetDriverByHubID(ctx, hub)
	if err != nil || d == nil {
		return nil, fmt.Errorf("водитель не найден")
	}
	recs, _ := s.store.ListRecommendations(ctx, d.ID)
	cmps, _ := s.store.ListComplaints(ctx, d.ID)
	bls, _ := s.store.ListBlacklist(ctx, d.ID)
	accs, _ := s.store.ListAccidents(ctx, d.ID)
	fines, _ := s.store.ListFines(ctx, d.ID)
	mine, _ := s.store.IsBlacklistedBy(ctx, d.ID, companyID)
	if recs == nil {
		recs = []domain.Recommendation{}
	}
	if cmps == nil {
		cmps = []domain.Complaint{}
	}
	if bls == nil {
		bls = []domain.BlacklistEntry{}
	}
	if accs == nil {
		accs = []domain.Accident{}
	}
	if fines == nil {
		fines = []domain.Fine{}
	}
	emp, _ := s.store.GetEmploymentByPair(ctx, d.ID, companyID)
	return &domain.DriverDossier{
		Driver: *d, Recommendations: recs, Complaints: cmps, BlacklistEntries: bls,
		Accidents: accs, Fines: fines, BlacklistedByViewer: mine, ViewerEmployment: emp,
	}, nil
}

func (s *Service) requireApprovedCompany(ctx context.Context, companyID uuid.UUID) (*domain.Company, error) {
	c, err := s.store.GetCompanyByID(ctx, companyID)
	if err != nil || c == nil {
		return nil, fmt.Errorf("компания не найдена")
	}
	if c.ModerationStatus != domain.ModerationApproved {
		return nil, fmt.Errorf("компания ещё не прошла модерацию")
	}
	return c, nil
}

func (s *Service) requireConfirmedEmployment(ctx context.Context, companyID, driverID uuid.UUID) error {
	if _, err := s.requireApprovedCompany(ctx, companyID); err != nil {
		return err
	}
	ok, err := s.store.HasConfirmedEmployment(ctx, driverID, companyID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("нужно подтверждённое трудоустройство: компания указывает водителя, водитель подтверждает")
	}
	return nil
}

func (s *Service) ClaimEmployment(ctx context.Context, companyID uuid.UUID, hub string) (*domain.Employment, error) {
	if _, err := s.requireApprovedCompany(ctx, companyID); err != nil {
		return nil, err
	}
	driverID, err := s.ResolveDriverID(ctx, hub)
	if err != nil {
		return nil, err
	}
	if existing, _ := s.store.GetEmploymentByPair(ctx, driverID, companyID); existing != nil && existing.Status == domain.EmploymentConfirmed {
		return existing, nil
	}
	e, err := s.store.UpsertEmploymentClaim(ctx, driverID, companyID)
	if err != nil {
		return nil, mapDBErr(err, "не удалось отправить запрос на трудоустройство")
	}
	full, _ := s.store.GetEmployment(ctx, e.ID)
	if full != nil {
		return full, nil
	}
	return e, nil
}

func (s *Service) ListCompanyEmployments(ctx context.Context, companyID uuid.UUID) ([]domain.Employment, error) {
	list, err := s.store.ListEmploymentsByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []domain.Employment{}
	}
	return list, nil
}

func (s *Service) ListDriverEmployments(ctx context.Context, driverID uuid.UUID) ([]domain.Employment, error) {
	list, err := s.store.ListEmploymentsByDriver(ctx, driverID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []domain.Employment{}
	}
	return list, nil
}

func (s *Service) ConfirmEmployment(ctx context.Context, driverID, employmentID uuid.UUID) (*domain.Employment, error) {
	e, err := s.store.GetEmployment(ctx, employmentID)
	if err != nil || e == nil {
		return nil, fmt.Errorf("запрос не найден")
	}
	if e.DriverID != driverID {
		return nil, fmt.Errorf("недостаточно прав")
	}
	if e.Status != domain.EmploymentPending {
		return nil, fmt.Errorf("запрос уже обработан")
	}
	return s.store.SetEmploymentStatus(ctx, employmentID, domain.EmploymentConfirmed)
}

func (s *Service) RejectEmployment(ctx context.Context, driverID, employmentID uuid.UUID) (*domain.Employment, error) {
	e, err := s.store.GetEmployment(ctx, employmentID)
	if err != nil || e == nil {
		return nil, fmt.Errorf("запрос не найден")
	}
	if e.DriverID != driverID {
		return nil, fmt.Errorf("недостаточно прав")
	}
	if e.Status != domain.EmploymentPending {
		return nil, fmt.Errorf("запрос уже обработан")
	}
	return s.store.SetEmploymentStatus(ctx, employmentID, domain.EmploymentRejected)
}

func (s *Service) EndEmployment(ctx context.Context, actorID uuid.UUID, role domain.Role, employmentID uuid.UUID) (*domain.Employment, error) {
	e, err := s.store.GetEmployment(ctx, employmentID)
	if err != nil || e == nil {
		return nil, fmt.Errorf("связь не найдена")
	}
	switch role {
	case domain.RoleDriver:
		if e.DriverID != actorID {
			return nil, fmt.Errorf("недостаточно прав")
		}
	case domain.RoleCompany:
		if e.CompanyID != actorID {
			return nil, fmt.Errorf("недостаточно прав")
		}
	default:
		return nil, fmt.Errorf("недостаточно прав")
	}
	if e.Status != domain.EmploymentConfirmed && e.Status != domain.EmploymentPending {
		return nil, fmt.Errorf("нельзя завершить эту связь")
	}
	return s.store.SetEmploymentStatus(ctx, employmentID, domain.EmploymentEnded)
}

func (s *Service) LoginAdmin(password string) (string, error) {
	if s.adminPassword == "" || password == "" || password != s.adminPassword {
		return "", fmt.Errorf("неверный пароль")
	}
	return s.tokens.Issue(uuid.Nil, domain.RoleAdmin, "admin")
}

func (s *Service) ListCompaniesAdmin(ctx context.Context, status domain.ModerationStatus) ([]domain.Company, error) {
	list, err := s.store.ListCompanies(ctx, status)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []domain.Company{}
	}
	return list, nil
}

func (s *Service) SetCompanyModeration(ctx context.Context, companyID uuid.UUID, status domain.ModerationStatus) (*domain.Company, error) {
	switch status {
	case domain.ModerationApproved, domain.ModerationRejected, domain.ModerationPending:
	default:
		return nil, fmt.Errorf("недопустимый статус модерации")
	}
	if err := s.store.SetCompanyModeration(ctx, companyID, status); err != nil {
		return nil, err
	}
	return s.store.GetCompanyByID(ctx, companyID)
}

func (s *Service) AddRecommendation(ctx context.Context, companyID, driverID uuid.UUID, text string, rating int) (*domain.Recommendation, error) {
	if err := s.requireConfirmedEmployment(ctx, companyID, driverID); err != nil {
		return nil, err
	}
	if rating < 1 || rating > 5 {
		return nil, fmt.Errorf("rating должен быть 1..5")
	}
	r := &domain.Recommendation{DriverID: driverID, CompanyID: companyID, Text: text, Rating: rating}
	if err := s.store.AddRecommendation(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Service) AddComplaint(ctx context.Context, companyID, driverID uuid.UUID, category, text, severity string) (*domain.Complaint, error) {
	if err := s.requireConfirmedEmployment(ctx, companyID, driverID); err != nil {
		return nil, err
	}
	c := &domain.Complaint{DriverID: driverID, CompanyID: companyID, Category: category, Text: text, Severity: severity}
	if err := s.store.AddComplaint(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) AddToBlacklist(ctx context.Context, companyID, driverID uuid.UUID, reason string) (*domain.BlacklistEntry, error) {
	if err := s.requireConfirmedEmployment(ctx, companyID, driverID); err != nil {
		return nil, err
	}
	b := &domain.BlacklistEntry{DriverID: driverID, CompanyID: companyID, Reason: reason}
	if err := s.store.UpsertBlacklist(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Service) LiftBlacklist(ctx context.Context, companyID, driverID uuid.UUID) error {
	if err := s.requireConfirmedEmployment(ctx, companyID, driverID); err != nil {
		return err
	}
	return s.store.LiftBlacklist(ctx, driverID, companyID)
}

func (s *Service) AddAccident(ctx context.Context, companyID, driverID uuid.UUID, occurredAt time.Time, description, fault, damage, location string) (*domain.Accident, error) {
	if err := s.requireConfirmedEmployment(ctx, companyID, driverID); err != nil {
		return nil, err
	}
	cid := companyID
	a := &domain.Accident{
		DriverID: driverID, CompanyID: &cid, OccurredAt: occurredAt,
		Description: description, Fault: fault, DamageLevel: damage, Location: location,
	}
	if err := s.store.AddAccident(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) AddFine(ctx context.Context, companyID, driverID uuid.UUID, article string, amount float64, issued *time.Time, paid bool, description string) (*domain.Fine, error) {
	if err := s.requireConfirmedEmployment(ctx, companyID, driverID); err != nil {
		return nil, err
	}
	cid := companyID
	f := &domain.Fine{
		DriverID: driverID, CompanyID: &cid, Article: article, Amount: amount,
		IssuedAt: issued, Paid: paid, Description: description, Source: "company",
	}
	if err := s.store.AddFine(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Service) AddMedical(ctx context.Context, driverID uuid.UUID, m *domain.MedicalCertificate) error {
	m.DriverID = driverID
	if m.Source == "" {
		m.Source = "upload"
	}
	return s.store.AddMedical(ctx, m)
}

func (s *Service) AddSuspension(ctx context.Context, driverID uuid.UUID, x *domain.LicenseSuspension) error {
	x.DriverID = driverID
	if x.Source == "" {
		x.Source = "manual"
	}
	return s.store.AddSuspension(ctx, x)
}

func (s *Service) AddCriminal(ctx context.Context, driverID uuid.UUID, x *domain.CriminalRecord) error {
	x.DriverID = driverID
	if x.Source == "" {
		x.Source = "manual"
	}
	return s.store.AddCriminal(ctx, x)
}

func (s *Service) ResolveDriverID(ctx context.Context, hubOrID string) (uuid.UUID, error) {
	hubOrID = strings.TrimSpace(hubOrID)
	if id, err := uuid.Parse(hubOrID); err == nil {
		return id, nil
	}
	d, err := s.store.GetDriverByHubID(ctx, hubid.Normalize(hubOrID))
	if err != nil || d == nil {
		return uuid.Nil, fmt.Errorf("водитель не найден")
	}
	return d.ID, nil
}

func (s *Service) ESIAMode() string { return s.esia.Mode() }

func (s *Service) VisionMode() string {
	if s.vision == nil {
		return "none"
	}
	return s.vision.Mode()
}

func (s *Service) ScanLicense(ctx context.Context, front, back *vision.ImageInput) (*vision.LicenseDraft, error) {
	if s.vision == nil {
		return nil, fmt.Errorf("распознавание ВУ не настроено")
	}
	return s.vision.ScanLicense(ctx, front, back)
}

func (s *Service) ConfirmLicense(ctx context.Context, driverID uuid.UUID, draft vision.LicenseDraft) (*domain.Driver, error) {
	if draft.Series == "" || draft.Number == "" {
		return nil, fmt.Errorf("укажите серию и номер ВУ")
	}
	cats := draft.Categories
	if len(cats) == 0 {
		cats = []string{}
	}
	birth := vision.ParseDate(draft.BirthDate)
	issue := vision.ParseDate(draft.IssueDate)
	expiry := vision.ParseDate(draft.ExpiryDate)

	if err := s.store.UpdateDriverProfile(ctx, driverID, draft.LastName, draft.FirstName, draft.MiddleName, birth, cats); err != nil {
		return nil, err
	}
	lic := &domain.DriverLicense{
		DriverID: driverID, Series: draft.Series, Number: draft.Number,
		IssueDate: issue, ExpiryDate: expiry, Categories: cats,
		Issuer: draft.Issuer, Source: "vision_" + draft.Source, Verified: false,
	}
	if err := s.store.UpsertLicense(ctx, lic); err != nil {
		return nil, err
	}
	return s.store.GetDriverByID(ctx, driverID)
}

func maskName(last, first, middle string) string {
	mask := func(s string) string {
		s = strings.TrimSpace(s)
		if s == "" {
			return ""
		}
		r, size := utf8.DecodeRuneInString(s)
		if r == utf8.RuneError {
			return "*"
		}
		return string(r) + strings.Repeat("*", utf8.RuneCountInString(s[size:]))
	}
	parts := []string{mask(last), mask(first)}
	if middle != "" {
		parts = append(parts, mask(middle))
	}
	return strings.Join(parts, " ")
}
