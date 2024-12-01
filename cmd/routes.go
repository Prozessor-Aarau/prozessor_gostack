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
	// TODO enable again for csrf handling
	// mux.Use(NoSurf)
	mux.Use(SessionLoad)
	mux.Use(middleware.Logger)

	// fileserver
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// nonprotected routes
	mux.Get("/", handlers.Repo.Home)
	mux.Get("/register", handlers.Repo.Register)
	mux.Post("/register", handlers.Repo.PostRegister)
	mux.Get("/login", handlers.Make(handlers.Repo.HandleLoginIndex))
	mux.Post("/login", handlers.Make(handlers.Repo.HandleLoginCreate))

	// for creating logins
	mux.Get("/createlogin", handlers.Repo.CreateLogin)
	mux.Post("/createlogin", handlers.Repo.PostCreateLogin)

	// protected routes
	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(Auth)

		mux.Get("/", handlers.Repo.Admin)

		mux.Get("/logout", handlers.Repo.Logout)
		// protected routes level 1
		// routes for the profile page
		mux.Get("/profile", handlers.Repo.EditVisitor)
		mux.Post("/profile", handlers.Repo.PostEditVisitor)

		// routes for the checkin page
		mux.Get("/checkin", handlers.Repo.Checkin)

		// make new checkin
		mux.Get("/checkin/{id}", handlers.Repo.CheckinVisitorForm)
		mux.Post("/checkin/{id}", handlers.Repo.PostCheckin)
		// edit checkin
		mux.Get("/checkin/{id}/{vid}", handlers.Repo.CheckinVisitorForm)
		mux.Post("/checkin/{id}/{vid}", handlers.Repo.PostCheckin)

		mux.Get("/checkinvisitorlist", handlers.Repo.GetCheckinVisitors)
		mux.Post("/checkinvisitorlist", handlers.Repo.PostGetCheckinVisitors)
		mux.Get("/checkedinvisitslist", handlers.Repo.GetVisitsByDate)
		mux.Delete("/visits/{id}", handlers.Repo.DeleteCheckin)

		// routes for the visitor page
		mux.Get("/visitors", handlers.Repo.Visitors)
		mux.Get("/visitorlist", handlers.Repo.GetVisitors)
		mux.Post("/visitorlist", handlers.Repo.PostGetVisitors)
		// route to get elements
		mux.Get("/visitors/selectbox/", handlers.Repo.VisitorsSelectBox)
		mux.Get("/visitors/selectbox/{id}", handlers.Repo.VisitorsSelectBox)
		mux.Post("/visitors/selectbox/", handlers.Repo.VisitorsSelectBox)
		mux.Post("/visitors/selectbox/{id}", handlers.Repo.VisitorsSelectBox)

		// routes for editing a visitor
		mux.Get("/visitors/{id}", handlers.Repo.EditVisitor)
		mux.Post("/visitors/{id}", handlers.Repo.PostEditVisitor)

		mux.Route("/l2", func(mux chi.Router) {
			mux.Use(Level2)
			// put routes, that use at least level 2 in here
		})

		mux.Route("/l3", func(mux chi.Router) {
			mux.Use(Level3)
			// put routes, that use at least level 3 in here
			// badges
			mux.Get("/badges", handlers.Repo.Badges)
			mux.Get("/badges/list", handlers.Repo.BadgesList)
			mux.Post("/badges/list", handlers.Repo.BadgesFilteredList)
			mux.Get("/badges/{id}", handlers.Repo.Badge)
			mux.Post("/badges/{id}", handlers.Repo.BadgeUpdate)
			mux.Post("/badges/reset/{id}", handlers.Repo.BadgeReset)
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
			mux.Get("/badges/create", handlers.Repo.BadgeCreate)
			mux.Post("/badges/create", handlers.Repo.PostBadgeCreate)
			mux.Get("/badges/scan", handlers.Repo.BadgeRead)
			mux.Delete("/badges/{id}", handlers.Repo.BadgeDelete)
		})
	})

	return mux
}
