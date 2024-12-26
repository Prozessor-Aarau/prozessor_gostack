package main

import (
	"net/http"

	"github.com/faebut/prozessor_gostack/internal/config"
	"github.com/faebut/prozessor_gostack/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)
	mux.Use(middleware.Logger)

	// fileserver
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// nonprotected routes
	mux.Get("/", handlers.Repo.Home)
	mux.Get("/register", handlers.Make(handlers.Repo.HandleRegisterIndex))
	mux.Post("/register", handlers.Make(handlers.Repo.HandleRegisterCreate))
	mux.Get("/registersuccess", handlers.Make(handlers.Repo.HandleRegisterSuccess))
	mux.Get("/login", handlers.Make(handlers.Repo.HandleLoginIndex))
	mux.Post("/login", handlers.Make(handlers.Repo.HandleLoginCreate))

	// for creating logins
	mux.Get("/createlogin", handlers.Make(handlers.Repo.HandleCreateLoginIndex))
	mux.Post("/createlogin", handlers.Make(handlers.Repo.HandleCreateLogin))

	// for creating damage reports
	mux.Get("/report", handlers.Make(handlers.Repo.HandleReportIndex))
	mux.Post("/report", handlers.Make(handlers.Repo.HandleReportCreate))
	mux.Get("/reportsuccess", handlers.Make(handlers.Repo.HandleReportSuccess))

	// protected routes
	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(Auth)

		mux.Get("/", handlers.Make(handlers.Repo.HandleAdminIndex))

		mux.Get("/logout", handlers.Repo.Logout)
		// protected routes level 1
		// routes for the profile page
		mux.Get("/profile", handlers.Make(handlers.Repo.HandleProfileIndex))
		mux.Post("/profile", handlers.Make(handlers.Repo.HandleProfileEdit))

		// routes for the checkin page
		mux.Get("/checkin", handlers.Make(handlers.Repo.HandleCheckinIndex))
		mux.Get("/checkin/visitorlist", handlers.Make(handlers.Repo.HandleCheckinList))
		mux.Post("/checkin/visitorlist", handlers.Make(handlers.Repo.HandleCheckinListFilter))

		// make new checkin
		mux.Get("/checkin/{id}", handlers.Make(handlers.Repo.HandleCheckinFormIndex))
		mux.Post("/checkin/{id}", handlers.Make(handlers.Repo.HandleCheckinForm))

		// edit checkin
		mux.Get("/checkin/{id}/{vid}", handlers.Make(handlers.Repo.HandleCheckinFormIndex))
		mux.Post("/checkin/{id}/{vid}", handlers.Make(handlers.Repo.HandleCheckinForm))
		mux.Delete("/visits/{id}", handlers.Make(handlers.Repo.HandleCheckinDelete))

		// routes for the visitor page
		mux.Get("/visitors", handlers.Make(handlers.Repo.HandleVisitorsIndex))
		mux.Get("/visitors/visitorlist", handlers.Make(handlers.Repo.HandleVisitorList))
		mux.Post("/visitors/visitorlist", handlers.Make(handlers.Repo.HandleVisitorListFilter))
		// route to get elements
		mux.Get("/visitors/selectbox/", handlers.Make(handlers.Repo.HandleVisitorsSelectBox))
		mux.Get("/visitors/selectbox/{id}", handlers.Make(handlers.Repo.HandleVisitorsSelectBox))
		mux.Post("/visitors/selectbox/", handlers.Make(handlers.Repo.HandleVisitorsSelectBox))
		mux.Post("/visitors/selectbox/{id}", handlers.Make(handlers.Repo.HandleVisitorsSelectBox))

		// routes for editing a visitor
		mux.Get("/visitors/{id}", handlers.Make(handlers.Repo.HandleProfileIndex))
		mux.Post("/visitors/{id}", handlers.Make(handlers.Repo.HandleProfileEdit))

		// routes for machines
		mux.Get("/machines", handlers.Make(handlers.Repo.HandleMachinesIndex))
		mux.Get("/machines/create", handlers.Make(handlers.Repo.HandleMachinesCreateIndex))
		mux.Post("/machines/create", handlers.Make(handlers.Repo.HandleMachinesCreate))
		mux.Get("/machines/list", handlers.Make(handlers.Repo.HandleMachinesList))
		mux.Post("/machines/list", handlers.Make(handlers.Repo.HandleMachinesListFilter))
		// edit machines
		mux.Get("/machines/edit/{id}", handlers.Make(handlers.Repo.HandleMachinesEditIndex))
		mux.Post("/machines/edit/{id}", handlers.Make(handlers.Repo.HandleMachinesEdit))
		mux.Delete("/machines/{id}", handlers.Make(handlers.Repo.HandleMachinesDelete))
		mux.Post("/machines/edit/{id}/qrcode", handlers.Make(handlers.Repo.HandleMachinesQRCreate))
		mux.Post("/machines/edit/{id}/qrcode/delete", handlers.Make(handlers.Repo.HandleMachinesQRDelete))
		mux.Post("/machines/edit/{id}/qrcode/print", handlers.Make(handlers.Repo.HandleMachinesQRPrint))

		mux.Route("/l2", func(mux chi.Router) {
			mux.Use(Level2)
			// put routes, that use at least level 2 in here
		})

		mux.Route("/l3", func(mux chi.Router) {
			mux.Use(Level3)
			// put routes, that use at least level 3 in here
			// badges
			mux.Get("/badges", handlers.Make(handlers.Repo.HandleBadgesIndex))
			mux.Get("/badges/list", handlers.Make(handlers.Repo.HandleBadgesList))
			mux.Post("/badges/list", handlers.Make(handlers.Repo.HandleBadgesListFilter))
			mux.Get("/badges/{id}", handlers.Make(handlers.Repo.HandleBadgeEditIndex))
			mux.Post("/badges/{id}", handlers.Make(handlers.Repo.HandleBadgeUpdate))
			mux.Post("/badges/reset/{id}", handlers.Make(handlers.Repo.HandleBadgeReset))
			// abonnements
			mux.Get("/abonnements", handlers.Repo.Abonnements)
			mux.Get("/abonnements/list", handlers.Repo.AbonnementsList)
			mux.Post("/abonnements/list", handlers.Repo.PostAbonnementsList)
			mux.Get("/abonnements/create", handlers.Repo.AbonnementCreate)
			mux.Post("/abonnements/create", handlers.Repo.PostAbonnementCreate)
			mux.Get("/abonnements/create/pricing/{id}", handlers.Repo.AbonnementCreatePrice)
			// create a new login
			mux.Post("/createloginmail", handlers.Repo.CreateLoginMail)
		})

		mux.Route("/l4", func(mux chi.Router) {
			mux.Use(Level4)
			// put routes, that use at least level 4 in here
			// badges
			mux.Get("/badges/create", handlers.Make(handlers.Repo.HandleBadgeCreateIndex))
			mux.Post("/badges/create", handlers.Make(handlers.Repo.HandleBadgeCreate))
			mux.Get("/badges/scan", handlers.Make(handlers.Repo.HandleBadgeRead))
			mux.Delete("/badges/{id}", handlers.Make(handlers.Repo.HandleBadgeDelete))
		})
	})

	return mux
}
