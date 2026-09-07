package httpapi

import (
	"net/http"

	"github.com/driver-hub/driver-hub/internal/auth"
	"github.com/driver-hub/driver-hub/internal/domain"
	"github.com/driver-hub/driver-hub/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(h *Handler, tm *auth.TokenManager) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RequestID, chimw.RealIP, chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/healthz", h.Health)

	r.Route("/api/v1", func(api chi.Router) {
		api.Use(chimw.Logger)
		api.Post("/auth/company/register", h.RegisterCompany)
		api.Post("/auth/company/login", h.LoginCompany)
		api.Post("/auth/driver/register", h.RegisterDriver)
		api.Post("/auth/driver/login", h.LoginDriver)
		api.Get("/auth/esia/start-register", h.StartESIARegister)
		api.Get("/auth/esia/callback", h.ESIACallback)
		api.Get("/mock-esia/login", h.MockESIALoginPage)

		api.Get("/drivers/{hubID}/public", h.PublicCard)

		api.Group(func(drv chi.Router) {
			drv.Use(middleware.Auth(tm, domain.RoleDriver))
			drv.Get("/me/driver", h.MeDriver)
			drv.Get("/auth/esia/start", h.StartESIA)
			drv.Post("/auth/esia/mock-verify", h.MockESIAVerify)
			drv.Post("/me/grants", h.GrantAccess)
			drv.Get("/me/grants", h.ListGrants)
			drv.Delete("/me/grants/{companyID}", h.RevokeAccess)
			drv.Post("/me/medical", h.AddMedical)
			drv.Post("/me/criminal", h.AddCriminal)
			drv.Post("/me/suspensions", h.AddSuspension)
			drv.Post("/me/license/scan", h.ScanLicense)
			drv.Post("/me/license/confirm", h.ConfirmLicense)
		})

		api.Group(func(co chi.Router) {
			co.Use(middleware.Auth(tm, domain.RoleCompany))
			co.Get("/me/company", h.MeCompany)
			co.Get("/drivers/{hubID}/dossier", h.GetDossier)
			co.Post("/drivers/{hubID}/recommendations", h.AddRecommendation)
			co.Post("/drivers/{hubID}/complaints", h.AddComplaint)
			co.Post("/drivers/{hubID}/blacklist", h.AddBlacklist)
			co.Delete("/drivers/{hubID}/blacklist", h.LiftBlacklist)
			co.Post("/drivers/{hubID}/accidents", h.AddAccident)
			co.Post("/drivers/{hubID}/fines", h.AddFine)
			co.Post("/drivers/{hubID}/suspensions", h.AddSuspension)
		})
	})

	return r
}
