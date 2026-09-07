package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleDriver  Role = "driver"
	RoleCompany Role = "company"
	RoleAdmin   Role = "admin"
)

type VerificationStatus string

const (
	VerificationNone       VerificationStatus = "none"
	VerificationPending    VerificationStatus = "pending"
	VerificationVerified   VerificationStatus = "verified"
	VerificationRejected   VerificationStatus = "rejected"
	VerificationExpired    VerificationStatus = "expired"
)

// Driver — центральный профиль водителя в Hub.
type Driver struct {
	ID                 uuid.UUID          `json:"id"`
	HubID              string             `json:"hub_id"` // публичный идентификатор, напр. DH-A1B2C3D4
	ESIAOID            string             `json:"esia_oid,omitempty"`
	Email              string             `json:"email,omitempty"`
	Phone              string             `json:"phone,omitempty"`
	LastName           string             `json:"last_name"`
	FirstName          string             `json:"first_name"`
	MiddleName         string             `json:"middle_name,omitempty"`
	BirthDate          *time.Time         `json:"birth_date,omitempty"`
	SNILS              string             `json:"snils,omitempty"`
	INN                string             `json:"inn,omitempty"`
	ExperienceYears    int                `json:"experience_years"`
	Categories         []string           `json:"categories"`
	VerificationStatus VerificationStatus `json:"verification_status"`
	VerifiedAt         *time.Time         `json:"verified_at,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`

	License            *DriverLicense     `json:"license,omitempty"`
	MedicalCertificates []MedicalCertificate `json:"medical_certificates,omitempty"`
	Suspensions        []LicenseSuspension `json:"suspensions,omitempty"`
	CriminalRecords    []CriminalRecord   `json:"criminal_records,omitempty"`
}

type DriverLicense struct {
	ID           uuid.UUID  `json:"id"`
	DriverID     uuid.UUID  `json:"driver_id"`
	Series       string     `json:"series"`
	Number       string     `json:"number"`
	IssueDate    *time.Time `json:"issue_date,omitempty"`
	ExpiryDate   *time.Time `json:"expiry_date,omitempty"`
	Categories   []string   `json:"categories"`
	Issuer       string     `json:"issuer,omitempty"`
	Source       string     `json:"source"` // esia | manual | gibdd
	Verified     bool       `json:"verified"`
	RawESIAPayload string   `json:"-"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type MedicalCertificate struct {
	ID         uuid.UUID  `json:"id"`
	DriverID   uuid.UUID  `json:"driver_id"`
	Number     string     `json:"number"`
	IssuedAt   *time.Time `json:"issued_at,omitempty"`
	ValidUntil *time.Time `json:"valid_until,omitempty"`
	Clinic     string     `json:"clinic,omitempty"`
	Result     string     `json:"result"` // fit | unfit | restricted
	Categories []string   `json:"categories,omitempty"`
	FileURL    string     `json:"file_url,omitempty"`
	Source     string     `json:"source"` // upload | esia | smev
	CreatedAt  time.Time  `json:"created_at"`
}

type LicenseSuspension struct {
	ID         uuid.UUID  `json:"id"`
	DriverID   uuid.UUID  `json:"driver_id"`
	Reason     string     `json:"reason"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	Active     bool       `json:"active"`
	Authority  string     `json:"authority,omitempty"`
	CaseNumber string     `json:"case_number,omitempty"`
	Source     string     `json:"source"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CriminalRecord struct {
	ID          uuid.UUID  `json:"id"`
	DriverID    uuid.UUID  `json:"driver_id"`
	HasRecord   bool       `json:"has_record"`
	Description string     `json:"description,omitempty"`
	CheckedAt   *time.Time `json:"checked_at,omitempty"`
	ValidUntil  *time.Time `json:"valid_until,omitempty"`
	Source      string     `json:"source"` // smev_mvd | manual | self_declared
	FileURL     string     `json:"file_url,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Company struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	INN          string    `json:"inn"`
	OGRN         string    `json:"ogrn,omitempty"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone,omitempty"`
	PasswordHash string    `json:"-"`
	Verified     bool      `json:"verified"`
	CreatedAt    time.Time `json:"created_at"`
}

type AccessGrant struct {
	ID         uuid.UUID  `json:"id"`
	DriverID   uuid.UUID  `json:"driver_id"`
	CompanyID  uuid.UUID  `json:"company_id"`
	Purpose    string     `json:"purpose"` // employment | audit | ongoing
	GrantedAt  time.Time  `json:"granted_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	Active     bool       `json:"active"`
}

type Recommendation struct {
	ID          uuid.UUID `json:"id"`
	DriverID    uuid.UUID `json:"driver_id"`
	CompanyID   uuid.UUID `json:"company_id"`
	CompanyName string    `json:"company_name,omitempty"`
	Text        string    `json:"text"`
	Rating      int       `json:"rating"` // 1-5
	CreatedAt   time.Time `json:"created_at"`
}

type Complaint struct {
	ID          uuid.UUID `json:"id"`
	DriverID    uuid.UUID `json:"driver_id"`
	CompanyID   uuid.UUID `json:"company_id"`
	CompanyName string    `json:"company_name,omitempty"`
	Category    string    `json:"category"` // discipline | safety | fraud | other
	Text        string    `json:"text"`
	Severity    string    `json:"severity"` // low | medium | high
	CreatedAt   time.Time `json:"created_at"`
}

type BlacklistEntry struct {
	ID          uuid.UUID  `json:"id"`
	DriverID    uuid.UUID  `json:"driver_id"`
	CompanyID   uuid.UUID  `json:"company_id"`
	CompanyName string     `json:"company_name,omitempty"`
	Reason      string     `json:"reason"`
	Active      bool       `json:"active"`
	CreatedAt   time.Time  `json:"created_at"`
	LiftedAt    *time.Time `json:"lifted_at,omitempty"`
}

type Accident struct {
	ID          uuid.UUID  `json:"id"`
	DriverID    uuid.UUID  `json:"driver_id"`
	CompanyID   *uuid.UUID `json:"company_id,omitempty"`
	CompanyName string     `json:"company_name,omitempty"`
	OccurredAt  time.Time  `json:"occurred_at"`
	Description string     `json:"description"`
	Fault       string     `json:"fault"` // driver | other | mutual | unknown
	DamageLevel string     `json:"damage_level"` // none | minor | major | fatal
	Location    string     `json:"location,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Fine struct {
	ID          uuid.UUID  `json:"id"`
	DriverID    uuid.UUID  `json:"driver_id"`
	CompanyID   *uuid.UUID `json:"company_id,omitempty"`
	Article     string     `json:"article,omitempty"`
	Amount      float64    `json:"amount"`
	IssuedAt    *time.Time `json:"issued_at,omitempty"`
	Paid        bool       `json:"paid"`
	Description string     `json:"description,omitempty"`
	Source      string     `json:"source"` // company | gibdd | manual
	CreatedAt   time.Time  `json:"created_at"`
}

// DriverDossier — полный досье для компании при наличии согласия.
type DriverDossier struct {
	Driver              Driver           `json:"driver"`
	Recommendations     []Recommendation `json:"recommendations"`
	Complaints          []Complaint      `json:"complaints"`
	BlacklistEntries    []BlacklistEntry `json:"blacklist_entries"`
	Accidents           []Accident       `json:"accidents"`
	Fines               []Fine           `json:"fines"`
	BlacklistedByViewer bool             `json:"blacklisted_by_viewer"`
	Grant               *AccessGrant     `json:"grant,omitempty"`
}

// PublicDriverCard — публичная карточка без ПДн (для поиска по Hub ID без согласия).
type PublicDriverCard struct {
	HubID              string             `json:"hub_id"`
	FullNameMasked     string             `json:"full_name_masked"`
	Categories         []string           `json:"categories"`
	ExperienceYears    int                `json:"experience_years"`
	VerificationStatus VerificationStatus `json:"verification_status"`
	HasActiveSuspension bool              `json:"has_active_suspension"`
	BlacklistCount     int                `json:"blacklist_count"`
	ComplaintCount     int                `json:"complaint_count"`
	AccidentCount      int                `json:"accident_count"`
}
