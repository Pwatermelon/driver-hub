package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/driver-hub/driver-hub/internal/domain"
	"github.com/driver-hub/driver-hub/internal/esia"
	"github.com/driver-hub/driver-hub/internal/middleware"
	"github.com/driver-hub/driver-hub/internal/service"
	"github.com/driver-hub/driver-hub/internal/vision"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	svc  *service.Service
	mock *esia.MockClient
}

func NewHandler(svc *service.Service, mock *esia.MockClient) *Handler {
	return &Handler{svc: svc, mock: mock}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":      "ok",
		"service":     "driver-hub",
		"esia_mode":   h.svc.ESIAMode(),
		"vision_mode": h.svc.VisionMode(),
	})
}

func (h *Handler) RegisterCompany(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		INN      string `json:"inn"`
		OGRN     string `json:"ogrn"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	c, token, err := h.svc.RegisterCompany(r.Context(), req.Name, req.INN, req.OGRN, req.Email, req.Phone, req.Password)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 201, map[string]any{"company": c, "token": token})
}

func (h *Handler) LoginCompany(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	c, token, err := h.svc.LoginCompany(r.Context(), req.Email, req.Password)
	if err != nil {
		writeErr(w, 401, err)
		return
	}
	writeJSON(w, 200, map[string]any{"company": c, "token": token})
}

func (h *Handler) RegisterDriver(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	d, token, err := h.svc.RegisterDriver(r.Context(), req.Email, req.Phone, req.Password)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 201, map[string]any{"driver": d, "token": token})
}

func (h *Handler) LoginDriver(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	d, token, err := h.svc.LoginDriver(r.Context(), req.Email, req.Password)
	if err != nil {
		writeErr(w, 401, err)
		return
	}
	writeJSON(w, 200, map[string]any{"driver": d, "token": token})
}

func (h *Handler) MeDriver(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFrom(r.Context())
	d, err := h.svc.GetDriver(r.Context(), claims.SubjectID)
	if err != nil || d == nil {
		writeJSON(w, 404, map[string]string{"error": "не найден"})
		return
	}
	writeJSON(w, 200, d)
}

func (h *Handler) StartESIA(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFrom(r.Context())
	url, err := h.svc.StartESIA(r.Context(), claims.SubjectID)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]string{"auth_url": url})
}

func (h *Handler) StartESIARegister(w http.ResponseWriter, r *http.Request) {
	url, err := h.svc.StartESIARegistration(r.Context())
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]string{"auth_url": url})
}

func (h *Handler) ESIACallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		writeJSON(w, 400, map[string]string{"error": "отсутствует code/state"})
		return
	}
	d, token, err := h.svc.CompleteESIA(r.Context(), code, state)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	// Для браузерного демо — редирект на UI с токеном.
	http.Redirect(w, r, "/driver?esia_ok=1&token="+token+"&hub_id="+d.HubID, http.StatusFound)
}

func (h *Handler) MockESIAVerify(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFrom(r.Context())
	var req struct {
		Preset string `json:"preset"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Preset == "" {
		req.Preset = "ivanov"
	}
	d, token, err := h.svc.MockESIAComplete(r.Context(), claims.SubjectID, req.Preset)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"driver": d, "token": token})
}

func (h *Handler) MockESIALoginPage(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	preset := r.URL.Query().Get("preset")
	if preset == "" {
		preset = "ivanov"
	}
	if h.mock == nil {
		http.Error(w, "mock ESIA disabled", 400)
		return
	}
	code := h.mock.IssueDemoCode(preset)
	http.Redirect(w, r, "/api/v1/auth/esia/callback?code="+code+"&state="+state, http.StatusFound)
}

func (h *Handler) PublicCard(w http.ResponseWriter, r *http.Request) {
	card, err := h.svc.PublicCard(r.Context(), chi.URLParam(r, "hubID"))
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	writeJSON(w, 200, card)
}

func (h *Handler) GrantAccess(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFrom(r.Context())
	var req struct {
		CompanyID string `json:"company_id"`
		Purpose   string `json:"purpose"`
		Days      int    `json:"days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	cid, err := uuid.Parse(req.CompanyID)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	g, err := h.svc.GrantAccess(r.Context(), claims.SubjectID, cid, req.Purpose, req.Days)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 201, g)
}

func (h *Handler) RevokeAccess(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFrom(r.Context())
	cid, err := uuid.Parse(chi.URLParam(r, "companyID"))
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := h.svc.RevokeAccess(r.Context(), claims.SubjectID, cid); err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "revoked"})
}

func (h *Handler) ListGrants(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFrom(r.Context())
	list, err := h.svc.ListGrants(r.Context(), claims.SubjectID)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, list)
}

func (h *Handler) GetDossier(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFrom(r.Context())
	dossier, err := h.svc.GetDossier(r.Context(), claims.SubjectID, chi.URLParam(r, "hubID"))
	if err != nil {
		writeErr(w, 403, err)
		return
	}
	writeJSON(w, 200, dossier)
}

func (h *Handler) resolveTarget(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "hubID")
	if raw == "" {
		raw = r.URL.Query().Get("hub_id")
	}
	return h.svc.ResolveDriverID(r.Context(), raw)
}

func (h *Handler) AddRecommendation(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFrom(r.Context())
	driverID, err := h.resolveTarget(r)
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	var req struct {
		Text   string `json:"text"`
		Rating int    `json:"rating"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	item, err := h.svc.AddRecommendation(r.Context(), claims.SubjectID, driverID, req.Text, req.Rating)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 201, item)
}

func (h *Handler) AddComplaint(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFrom(r.Context())
	driverID, err := h.resolveTarget(r)
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	var req struct {
		Category string `json:"category"`
		Text     string `json:"text"`
		Severity string `json:"severity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if req.Category == "" {
		req.Category = "other"
	}
	if req.Severity == "" {
		req.Severity = "medium"
	}
	item, err := h.svc.AddComplaint(r.Context(), claims.SubjectID, driverID, req.Category, req.Text, req.Severity)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 201, item)
}

func (h *Handler) AddBlacklist(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFrom(r.Context())
	driverID, err := h.resolveTarget(r)
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	item, err := h.svc.AddToBlacklist(r.Context(), claims.SubjectID, driverID, req.Reason)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 201, item)
}

func (h *Handler) LiftBlacklist(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFrom(r.Context())
	driverID, err := h.resolveTarget(r)
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	if err := h.svc.LiftBlacklist(r.Context(), claims.SubjectID, driverID); err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "lifted"})
}

func (h *Handler) AddAccident(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFrom(r.Context())
	driverID, err := h.resolveTarget(r)
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	var req struct {
		OccurredAt  string `json:"occurred_at"`
		Description string `json:"description"`
		Fault       string `json:"fault"`
		DamageLevel string `json:"damage_level"`
		Location    string `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	occurred := time.Now()
	if req.OccurredAt != "" {
		if t, err := time.Parse(time.RFC3339, req.OccurredAt); err == nil {
			occurred = t
		}
	}
	if req.Fault == "" {
		req.Fault = "unknown"
	}
	if req.DamageLevel == "" {
		req.DamageLevel = "minor"
	}
	item, err := h.svc.AddAccident(r.Context(), claims.SubjectID, driverID, occurred, req.Description, req.Fault, req.DamageLevel, req.Location)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 201, item)
}

func (h *Handler) AddFine(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFrom(r.Context())
	driverID, err := h.resolveTarget(r)
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	var req struct {
		Article     string  `json:"article"`
		Amount      float64 `json:"amount"`
		IssuedAt    string  `json:"issued_at"`
		Paid        bool    `json:"paid"`
		Description string  `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	var issued *time.Time
	if req.IssuedAt != "" {
		if t, err := time.Parse("2006-01-02", req.IssuedAt); err == nil {
			issued = &t
		}
	}
	item, err := h.svc.AddFine(r.Context(), claims.SubjectID, driverID, req.Article, req.Amount, issued, req.Paid, req.Description)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 201, item)
}

func (h *Handler) AddMedical(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFrom(r.Context())
	var req domain.MedicalCertificate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := h.svc.AddMedical(r.Context(), claims.SubjectID, &req); err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 201, req)
}

func (h *Handler) AddSuspension(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFrom(r.Context())
	var req domain.LicenseSuspension
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	// Компания или сам водитель (admin-путь упрощён — водитель/компания через роль).
	driverID := claims.SubjectID
	if claims.Role == domain.RoleCompany {
		id, err := h.resolveTarget(r)
		if err != nil {
			writeErr(w, 404, err)
			return
		}
		driverID = id
	}
	if err := h.svc.AddSuspension(r.Context(), driverID, &req); err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 201, req)
}

func (h *Handler) AddCriminal(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFrom(r.Context())
	var req domain.CriminalRecord
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := h.svc.AddCriminal(r.Context(), claims.SubjectID, &req); err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 201, req)
}

func (h *Handler) MeCompany(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFrom(r.Context())
	c, err := h.svc.GetCompany(r.Context(), claims.SubjectID)
	if err != nil || c == nil {
		writeJSON(w, 404, map[string]string{"error": "не найден"})
		return
	}
	writeJSON(w, 200, c)
}

func (h *Handler) ScanLicense(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		writeErr(w, 400, err)
		return
	}
	front, err := readUpload(r, "front")
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	back, err := readUpload(r, "back")
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	draft, err := h.svc.ScanLicense(r.Context(), front, back)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, draft)
}

func (h *Handler) ConfirmLicense(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFrom(r.Context())
	var draft vision.LicenseDraft
	if err := json.NewDecoder(r.Body).Decode(&draft); err != nil {
		writeErr(w, 400, err)
		return
	}
	if draft.Source == "" {
		draft.Source = "manual"
	}
	d, err := h.svc.ConfirmLicense(r.Context(), claims.SubjectID, draft)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, d)
}

func readUpload(r *http.Request, field string) (*vision.ImageInput, error) {
	file, hdr, err := r.FormFile(field)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 12<<20))
	if err != nil {
		return nil, err
	}
	ct := hdr.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/jpeg"
	}
	if !strings.HasPrefix(ct, "image/") {
		// некоторые мобильные браузеры шлют application/octet-stream
		name := strings.ToLower(hdr.Filename)
		switch {
		case strings.HasSuffix(name, ".png"):
			ct = "image/png"
		case strings.HasSuffix(name, ".webp"):
			ct = "image/webp"
		default:
			ct = "image/jpeg"
		}
	}
	return &vision.ImageInput{Filename: hdr.Filename, ContentType: ct, Data: data}, nil
}
