package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/driver-hub/driver-hub/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) CreateCompany(ctx context.Context, c *domain.Company) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO companies (name, inn, ogrn, email, phone, password_hash, verified)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, created_at`,
		c.Name, c.INN, c.OGRN, c.Email, c.Phone, c.PasswordHash, c.Verified,
	).Scan(&c.ID, &c.CreatedAt)
}

func (s *Store) GetCompanyByEmail(ctx context.Context, email string) (*domain.Company, error) {
	var c domain.Company
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, inn, COALESCE(ogrn,''), email, COALESCE(phone,''), password_hash, verified, created_at
		FROM companies WHERE email=$1`, email,
	).Scan(&c.ID, &c.Name, &c.INN, &c.OGRN, &c.Email, &c.Phone, &c.PasswordHash, &c.Verified, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &c, err
}

func (s *Store) GetCompanyByID(ctx context.Context, id uuid.UUID) (*domain.Company, error) {
	var c domain.Company
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, inn, COALESCE(ogrn,''), email, COALESCE(phone,''), password_hash, verified, created_at
		FROM companies WHERE id=$1`, id,
	).Scan(&c.ID, &c.Name, &c.INN, &c.OGRN, &c.Email, &c.Phone, &c.PasswordHash, &c.Verified, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &c, err
}

func (s *Store) UpdateCompanyPassword(ctx context.Context, id uuid.UUID, hash string) error {
	_, err := s.pool.Exec(ctx, `UPDATE companies SET password_hash=$2, verified=TRUE WHERE id=$1`, id, hash)
	return err
}

func (s *Store) UpdateDriverPassword(ctx context.Context, id uuid.UUID, hash string) error {
	_, err := s.pool.Exec(ctx, `UPDATE drivers SET password_hash=$2, updated_at=NOW() WHERE id=$1`, id, hash)
	return err
}

func (s *Store) UpdateDriverProfile(ctx context.Context, id uuid.UUID, last, first, middle string, birth *time.Time, categories []string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE drivers SET
			last_name=COALESCE(NULLIF($2,''), last_name),
			first_name=COALESCE(NULLIF($3,''), first_name),
			middle_name=COALESCE(NULLIF($4,''), middle_name),
			birth_date=COALESCE($5, birth_date),
			categories=CASE WHEN cardinality($6::text[])>0 THEN $6 ELSE categories END,
			updated_at=NOW()
		WHERE id=$1`, id, last, first, middle, birth, categories)
	return err
}

func (s *Store) CreateDriver(ctx context.Context, d *domain.Driver, passwordHash string) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO drivers (hub_id, email, phone, last_name, first_name, middle_name, experience_years, categories, verification_status, password_hash)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, created_at, updated_at`,
		d.HubID, nullStr(d.Email), nullStr(d.Phone), d.LastName, d.FirstName, nullStr(d.MiddleName),
		d.ExperienceYears, d.Categories, d.VerificationStatus, nullStr(passwordHash),
	).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
}

func (s *Store) GetDriverByEmail(ctx context.Context, email string) (*domain.Driver, string, error) {
	var d domain.Driver
	var pass *string
	err := s.pool.QueryRow(ctx, `
		SELECT id, hub_id, COALESCE(esia_oid,''), COALESCE(email,''), COALESCE(phone,''),
		       last_name, first_name, COALESCE(middle_name,''), birth_date, COALESCE(snils,''), COALESCE(inn,''),
		       experience_years, categories, verification_status, verified_at, password_hash, created_at, updated_at
		FROM drivers WHERE email=$1`, email,
	).Scan(&d.ID, &d.HubID, &d.ESIAOID, &d.Email, &d.Phone, &d.LastName, &d.FirstName, &d.MiddleName,
		&d.BirthDate, &d.SNILS, &d.INN, &d.ExperienceYears, &d.Categories, &d.VerificationStatus,
		&d.VerifiedAt, &pass, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", nil
	}
	hash := ""
	if pass != nil {
		hash = *pass
	}
	return &d, hash, err
}

func (s *Store) GetDriverByID(ctx context.Context, id uuid.UUID) (*domain.Driver, error) {
	return s.scanDriver(ctx, `WHERE id=$1`, id)
}

func (s *Store) GetDriverByHubID(ctx context.Context, hubID string) (*domain.Driver, error) {
	return s.scanDriver(ctx, `WHERE hub_id=$1`, hubID)
}

func (s *Store) GetDriverByESIAOID(ctx context.Context, oid string) (*domain.Driver, error) {
	return s.scanDriver(ctx, `WHERE esia_oid=$1`, oid)
}

func (s *Store) scanDriver(ctx context.Context, where string, arg any) (*domain.Driver, error) {
	var d domain.Driver
	err := s.pool.QueryRow(ctx, `
		SELECT id, hub_id, COALESCE(esia_oid,''), COALESCE(email,''), COALESCE(phone,''),
		       last_name, first_name, COALESCE(middle_name,''), birth_date, COALESCE(snils,''), COALESCE(inn,''),
		       experience_years, categories, verification_status, verified_at, created_at, updated_at
		FROM drivers `+where, arg,
	).Scan(&d.ID, &d.HubID, &d.ESIAOID, &d.Email, &d.Phone, &d.LastName, &d.FirstName, &d.MiddleName,
		&d.BirthDate, &d.SNILS, &d.INN, &d.ExperienceYears, &d.Categories, &d.VerificationStatus,
		&d.VerifiedAt, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = s.attachDriverExtras(ctx, &d)
	return &d, nil
}

func (s *Store) attachDriverExtras(ctx context.Context, d *domain.Driver) error {
	lic, err := s.GetLicense(ctx, d.ID)
	if err != nil {
		return err
	}
	d.License = lic
	meds, err := s.ListMedical(ctx, d.ID)
	if err != nil {
		return err
	}
	d.MedicalCertificates = meds
	susp, err := s.ListSuspensions(ctx, d.ID)
	if err != nil {
		return err
	}
	d.Suspensions = susp
	crim, err := s.ListCriminal(ctx, d.ID)
	if err != nil {
		return err
	}
	d.CriminalRecords = crim
	return nil
}

func (s *Store) ApplyESIAPerson(ctx context.Context, driverID uuid.UUID, oid, first, last, middle, snils, inn, phone, email string, birth *time.Time, categories []string) error {
	now := time.Now()
	_, err := s.pool.Exec(ctx, `
		UPDATE drivers SET
			esia_oid=$2, first_name=$3, last_name=$4, middle_name=$5, snils=$6, inn=$7,
			phone=COALESCE(NULLIF($8,''), phone), email=COALESCE(NULLIF($9,''), email),
			birth_date=$10, categories=CASE WHEN cardinality($11::text[])>0 THEN $11 ELSE categories END,
			verification_status='verified', verified_at=$12, updated_at=$12
		WHERE id=$1`,
		driverID, oid, first, last, nullStr(middle), snils, inn, phone, email, birth, categories, now,
	)
	return err
}

func (s *Store) ClearESIAOID(ctx context.Context, driverID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `UPDATE drivers SET esia_oid=NULL, updated_at=NOW() WHERE id=$1`, driverID)
	return err
}

func (s *Store) UpsertLicense(ctx context.Context, lic *domain.DriverLicense) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO driver_licenses (driver_id, series, number, issue_date, expiry_date, categories, issuer, source, verified, raw_esia_payload)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (driver_id) DO UPDATE SET
			series=EXCLUDED.series, number=EXCLUDED.number, issue_date=EXCLUDED.issue_date,
			expiry_date=EXCLUDED.expiry_date, categories=EXCLUDED.categories, issuer=EXCLUDED.issuer,
			source=EXCLUDED.source, verified=EXCLUDED.verified, raw_esia_payload=EXCLUDED.raw_esia_payload,
			updated_at=NOW()
		RETURNING id, updated_at`,
		lic.DriverID, lic.Series, lic.Number, lic.IssueDate, lic.ExpiryDate, lic.Categories,
		lic.Issuer, lic.Source, lic.Verified, nullStr(lic.RawESIAPayload),
	).Scan(&lic.ID, &lic.UpdatedAt)
}

func (s *Store) GetLicense(ctx context.Context, driverID uuid.UUID) (*domain.DriverLicense, error) {
	var lic domain.DriverLicense
	err := s.pool.QueryRow(ctx, `
		SELECT id, driver_id, series, number, issue_date, expiry_date, categories, COALESCE(issuer,''), source, verified, updated_at
		FROM driver_licenses WHERE driver_id=$1`, driverID,
	).Scan(&lic.ID, &lic.DriverID, &lic.Series, &lic.Number, &lic.IssueDate, &lic.ExpiryDate,
		&lic.Categories, &lic.Issuer, &lic.Source, &lic.Verified, &lic.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &lic, err
}

func (s *Store) AddMedical(ctx context.Context, m *domain.MedicalCertificate) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO medical_certificates (driver_id, number, issued_at, valid_until, clinic, result, categories, file_url, source)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, created_at`,
		m.DriverID, m.Number, m.IssuedAt, m.ValidUntil, nullStr(m.Clinic), m.Result, m.Categories, nullStr(m.FileURL), m.Source,
	).Scan(&m.ID, &m.CreatedAt)
}

func (s *Store) ListMedical(ctx context.Context, driverID uuid.UUID) ([]domain.MedicalCertificate, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, driver_id, number, issued_at, valid_until, COALESCE(clinic,''), result, categories, COALESCE(file_url,''), source, created_at
		FROM medical_certificates WHERE driver_id=$1 ORDER BY created_at DESC`, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.MedicalCertificate
	for rows.Next() {
		var m domain.MedicalCertificate
		if err := rows.Scan(&m.ID, &m.DriverID, &m.Number, &m.IssuedAt, &m.ValidUntil, &m.Clinic, &m.Result, &m.Categories, &m.FileURL, &m.Source, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) AddSuspension(ctx context.Context, x *domain.LicenseSuspension) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO license_suspensions (driver_id, reason, start_date, end_date, active, authority, case_number, source)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, created_at`,
		x.DriverID, x.Reason, x.StartDate, x.EndDate, x.Active, nullStr(x.Authority), nullStr(x.CaseNumber), x.Source,
	).Scan(&x.ID, &x.CreatedAt)
}

func (s *Store) ListSuspensions(ctx context.Context, driverID uuid.UUID) ([]domain.LicenseSuspension, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, driver_id, reason, start_date, end_date, active, COALESCE(authority,''), COALESCE(case_number,''), source, created_at
		FROM license_suspensions WHERE driver_id=$1 ORDER BY created_at DESC`, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.LicenseSuspension
	for rows.Next() {
		var x domain.LicenseSuspension
		if err := rows.Scan(&x.ID, &x.DriverID, &x.Reason, &x.StartDate, &x.EndDate, &x.Active, &x.Authority, &x.CaseNumber, &x.Source, &x.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *Store) AddCriminal(ctx context.Context, x *domain.CriminalRecord) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO criminal_records (driver_id, has_record, description, checked_at, valid_until, source, file_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at`,
		x.DriverID, x.HasRecord, nullStr(x.Description), x.CheckedAt, x.ValidUntil, x.Source, nullStr(x.FileURL),
	).Scan(&x.ID, &x.CreatedAt)
}

func (s *Store) ListCriminal(ctx context.Context, driverID uuid.UUID) ([]domain.CriminalRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, driver_id, has_record, COALESCE(description,''), checked_at, valid_until, source, COALESCE(file_url,''), created_at
		FROM criminal_records WHERE driver_id=$1 ORDER BY created_at DESC`, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.CriminalRecord
	for rows.Next() {
		var x domain.CriminalRecord
		if err := rows.Scan(&x.ID, &x.DriverID, &x.HasRecord, &x.Description, &x.CheckedAt, &x.ValidUntil, &x.Source, &x.FileURL, &x.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *Store) UpsertGrant(ctx context.Context, g *domain.AccessGrant) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO access_grants (driver_id, company_id, purpose, expires_at, active)
		VALUES ($1,$2,$3,$4,TRUE)
		ON CONFLICT (driver_id, company_id) DO UPDATE SET
			purpose=EXCLUDED.purpose, expires_at=EXCLUDED.expires_at, active=TRUE, revoked_at=NULL, granted_at=NOW()
		RETURNING id, granted_at, active`,
		g.DriverID, g.CompanyID, g.Purpose, g.ExpiresAt,
	).Scan(&g.ID, &g.GrantedAt, &g.Active)
}

func (s *Store) RevokeGrant(ctx context.Context, driverID, companyID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE access_grants SET active=FALSE, revoked_at=NOW()
		WHERE driver_id=$1 AND company_id=$2`, driverID, companyID)
	return err
}

func (s *Store) GetActiveGrant(ctx context.Context, driverID, companyID uuid.UUID) (*domain.AccessGrant, error) {
	var g domain.AccessGrant
	err := s.pool.QueryRow(ctx, `
		SELECT id, driver_id, company_id, purpose, granted_at, expires_at, revoked_at, active
		FROM access_grants
		WHERE driver_id=$1 AND company_id=$2 AND active=TRUE
		  AND (expires_at IS NULL OR expires_at > NOW())`, driverID, companyID,
	).Scan(&g.ID, &g.DriverID, &g.CompanyID, &g.Purpose, &g.GrantedAt, &g.ExpiresAt, &g.RevokedAt, &g.Active)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &g, err
}

func (s *Store) ListGrantsByDriver(ctx context.Context, driverID uuid.UUID) ([]domain.AccessGrant, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, driver_id, company_id, purpose, granted_at, expires_at, revoked_at, active
		FROM access_grants WHERE driver_id=$1 ORDER BY granted_at DESC`, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.AccessGrant
	for rows.Next() {
		var g domain.AccessGrant
		if err := rows.Scan(&g.ID, &g.DriverID, &g.CompanyID, &g.Purpose, &g.GrantedAt, &g.ExpiresAt, &g.RevokedAt, &g.Active); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (s *Store) AddRecommendation(ctx context.Context, r *domain.Recommendation) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO recommendations (driver_id, company_id, text, rating)
		VALUES ($1,$2,$3,$4) RETURNING id, created_at`,
		r.DriverID, r.CompanyID, r.Text, r.Rating,
	).Scan(&r.ID, &r.CreatedAt)
}

func (s *Store) ListRecommendations(ctx context.Context, driverID uuid.UUID) ([]domain.Recommendation, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT r.id, r.driver_id, r.company_id, c.name, r.text, r.rating, r.created_at
		FROM recommendations r JOIN companies c ON c.id=r.company_id
		WHERE r.driver_id=$1 ORDER BY r.created_at DESC`, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Recommendation
	for rows.Next() {
		var r domain.Recommendation
		if err := rows.Scan(&r.ID, &r.DriverID, &r.CompanyID, &r.CompanyName, &r.Text, &r.Rating, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) AddComplaint(ctx context.Context, c *domain.Complaint) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO complaints (driver_id, company_id, category, text, severity)
		VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`,
		c.DriverID, c.CompanyID, c.Category, c.Text, c.Severity,
	).Scan(&c.ID, &c.CreatedAt)
}

func (s *Store) ListComplaints(ctx context.Context, driverID uuid.UUID) ([]domain.Complaint, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT x.id, x.driver_id, x.company_id, c.name, x.category, x.text, x.severity, x.created_at
		FROM complaints x JOIN companies c ON c.id=x.company_id
		WHERE x.driver_id=$1 ORDER BY x.created_at DESC`, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Complaint
	for rows.Next() {
		var c domain.Complaint
		if err := rows.Scan(&c.ID, &c.DriverID, &c.CompanyID, &c.CompanyName, &c.Category, &c.Text, &c.Severity, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) UpsertBlacklist(ctx context.Context, b *domain.BlacklistEntry) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO blacklist_entries (driver_id, company_id, reason, active)
		VALUES ($1,$2,$3,TRUE)
		ON CONFLICT (driver_id, company_id) DO UPDATE SET
			reason=EXCLUDED.reason, active=TRUE, lifted_at=NULL, created_at=NOW()
		RETURNING id, created_at, active`,
		b.DriverID, b.CompanyID, b.Reason,
	).Scan(&b.ID, &b.CreatedAt, &b.Active)
}

func (s *Store) LiftBlacklist(ctx context.Context, driverID, companyID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE blacklist_entries SET active=FALSE, lifted_at=NOW()
		WHERE driver_id=$1 AND company_id=$2`, driverID, companyID)
	return err
}

func (s *Store) ListBlacklist(ctx context.Context, driverID uuid.UUID) ([]domain.BlacklistEntry, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT b.id, b.driver_id, b.company_id, c.name, b.reason, b.active, b.created_at, b.lifted_at
		FROM blacklist_entries b JOIN companies c ON c.id=b.company_id
		WHERE b.driver_id=$1 ORDER BY b.created_at DESC`, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.BlacklistEntry
	for rows.Next() {
		var b domain.BlacklistEntry
		if err := rows.Scan(&b.ID, &b.DriverID, &b.CompanyID, &b.CompanyName, &b.Reason, &b.Active, &b.CreatedAt, &b.LiftedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) IsBlacklistedBy(ctx context.Context, driverID, companyID uuid.UUID) (bool, error) {
	var active bool
	err := s.pool.QueryRow(ctx, `
		SELECT active FROM blacklist_entries WHERE driver_id=$1 AND company_id=$2`, driverID, companyID,
	).Scan(&active)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return active, err
}

func (s *Store) AddAccident(ctx context.Context, a *domain.Accident) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO accidents (driver_id, company_id, occurred_at, description, fault, damage_level, location)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at`,
		a.DriverID, a.CompanyID, a.OccurredAt, a.Description, a.Fault, a.DamageLevel, nullStr(a.Location),
	).Scan(&a.ID, &a.CreatedAt)
}

func (s *Store) ListAccidents(ctx context.Context, driverID uuid.UUID) ([]domain.Accident, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.id, a.driver_id, a.company_id, COALESCE(c.name,''), a.occurred_at, a.description, a.fault, a.damage_level, COALESCE(a.location,''), a.created_at
		FROM accidents a LEFT JOIN companies c ON c.id=a.company_id
		WHERE a.driver_id=$1 ORDER BY a.occurred_at DESC`, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Accident
	for rows.Next() {
		var a domain.Accident
		if err := rows.Scan(&a.ID, &a.DriverID, &a.CompanyID, &a.CompanyName, &a.OccurredAt, &a.Description, &a.Fault, &a.DamageLevel, &a.Location, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) AddFine(ctx context.Context, f *domain.Fine) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO fines (driver_id, company_id, article, amount, issued_at, paid, description, source)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, created_at`,
		f.DriverID, f.CompanyID, nullStr(f.Article), f.Amount, f.IssuedAt, f.Paid, nullStr(f.Description), f.Source,
	).Scan(&f.ID, &f.CreatedAt)
}

func (s *Store) ListFines(ctx context.Context, driverID uuid.UUID) ([]domain.Fine, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, driver_id, company_id, COALESCE(article,''), amount, issued_at, paid, COALESCE(description,''), source, created_at
		FROM fines WHERE driver_id=$1 ORDER BY created_at DESC`, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Fine
	for rows.Next() {
		var f domain.Fine
		if err := rows.Scan(&f.ID, &f.DriverID, &f.CompanyID, &f.Article, &f.Amount, &f.IssuedAt, &f.Paid, &f.Description, &f.Source, &f.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (s *Store) CountActiveBlacklist(ctx context.Context, driverID uuid.UUID) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM blacklist_entries WHERE driver_id=$1 AND active`, driverID).Scan(&n)
	return n, err
}

func (s *Store) CountComplaints(ctx context.Context, driverID uuid.UUID) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM complaints WHERE driver_id=$1`, driverID).Scan(&n)
	return n, err
}

func (s *Store) CountAccidents(ctx context.Context, driverID uuid.UUID) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM accidents WHERE driver_id=$1`, driverID).Scan(&n)
	return n, err
}

func (s *Store) HasActiveSuspension(ctx context.Context, driverID uuid.UUID) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM license_suspensions WHERE driver_id=$1 AND active AND (end_date IS NULL OR end_date >= CURRENT_DATE))`,
		driverID,
	).Scan(&exists)
	return exists, err
}

func (s *Store) SaveESIASession(ctx context.Context, state, nonce string, driverID *uuid.UUID, ttl time.Duration) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO esia_sessions (state, nonce, driver_id, expires_at)
		VALUES ($1,$2,$3,$4)`, state, nonce, driverID, time.Now().Add(ttl))
	return err
}

func (s *Store) PopESIASession(ctx context.Context, state string) (nonce string, driverID *uuid.UUID, err error) {
	err = s.pool.QueryRow(ctx, `
		DELETE FROM esia_sessions WHERE state=$1 AND expires_at > NOW()
		RETURNING nonce, driver_id`, state,
	).Scan(&nonce, &driverID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil, fmt.Errorf("сессия ЕСИА не найдена или истекла")
	}
	return nonce, driverID, err
}

func (s *Store) Audit(ctx context.Context, actorType string, actorID *uuid.UUID, action, entityType string, entityID *uuid.UUID, meta string) {
	_, _ = s.pool.Exec(ctx, `
		INSERT INTO audit_log (actor_type, actor_id, action, entity_type, entity_id, meta)
		VALUES ($1,$2,$3,$4,$5,$6::jsonb)`, actorType, actorID, action, entityType, entityID, nullJSON(meta))
}

func nullStr(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

func nullJSON(s string) string {
	if strings.TrimSpace(s) == "" {
		return "{}"
	}
	return s
}
