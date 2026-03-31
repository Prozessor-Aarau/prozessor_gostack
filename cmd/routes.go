package main

import (
	"net/http"

	"github.com/faebut/prozessor_gostack/internal/config"
	"github.com/faebut/prozessor_gostack/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://db.prozessor-aarau.ch", "http://localhost", "https://prozessor-aarau.ch", "https://www.prozessor-aarau.ch"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
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
	mux.Get("/unsubscribe/{id}", handlers.Make(handlers.Repo.HandleNewsletterUnsubscribe))
	mux.Get("/subscribe", handlers.Make(handlers.Repo.HandleNewsletterSubscribeIndex))
	mux.Post("/subscribe", handlers.Make(handlers.Repo.HandleNewsletterSubscribe))

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
		mux.Post("/expertise", handlers.Make(handlers.Repo.HandleExpertiseUpdate))

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

		// membership
		mux.Group(func(mux chi.Router) {
			mux.Use(CheckPermission("EditVisitorMembership"))
			// create a new login
			mux.Post("/createloginmail", handlers.Repo.CreateLoginMail)
		})

		// machines
		mux.Group(func(mux chi.Router) {
			mux.Use(CheckPermission("ViewMachines"))
			mux.Get("/machines", handlers.Make(handlers.Repo.HandleMachinesIndex))
			mux.Get("/machines/create", handlers.Make(handlers.Repo.HandleMachinesCreateIndex))
			mux.Post("/machines/create", handlers.Make(handlers.Repo.HandleMachinesCreate))
			mux.Get("/machines/list", handlers.Make(handlers.Repo.HandleMachinesList))
			mux.Post("/machines/list", handlers.Make(handlers.Repo.HandleMachinesListFilter))
			// edit machines
			mux.Get("/machines/edit/{id}", handlers.Make(handlers.Repo.HandleMachinesEditIndex))
			mux.Post("/machines/edit/{id}", handlers.Make(handlers.Repo.HandleMachinesEdit))
			mux.Delete("/machines/{id}", handlers.Make(handlers.Repo.HandleMachinesDelete))
			mux.Post(
				"/machines/edit/{id}/qrcode",
				handlers.Make(handlers.Repo.HandleMachinesQRCreate),
			)
			mux.Post(
				"/machines/edit/{id}/qrcode/delete",
				handlers.Make(handlers.Repo.HandleMachinesQRDelete),
			)
			mux.Post(
				"/machines/edit/{id}/qrcode/print",
				handlers.Make(handlers.Repo.HandleMachinesQRPrint),
			)
			// edit reports
			mux.Post("/machines/{id}/reports/{reportid}/{solved}", handlers.Make(handlers.Repo.HandleMachineUpdateReportStatus))
			mux.Delete("/machines/{id}/reports/{reportid}", handlers.Make(handlers.Repo.HandleMachineDeleteReport))
		})

		// supervisions calendar
		mux.Group(func(mux chi.Router) {
			mux.Use(CheckPermission("ViewSupervisions"))
			mux.Get("/supervisions", handlers.Make(handlers.Repo.HandleSupervisionsIndex))
			mux.Get("/supervisions/opening_hours/{startdate}/{daycount}", handlers.Make(handlers.Repo.HandleSupervisionsOpeningHours))

			// add opening hours
			mux.Group(func(mux chi.Router) {
				mux.Use(CheckPermission("EditOpeningHours"))
				mux.Get("/supervisions/opening_hours/new", handlers.Make(handlers.Repo.HandleSupervisionsOpeningHoursNew))
				mux.Post("/supervisions/opening_hours", handlers.Make(handlers.Repo.HandleSupervisionsOpeningHoursCreate))
				mux.Group(func(mux chi.Router) {
					mux.Delete("/supervisions/opening_hours/{oh_id}", handlers.Make(handlers.Repo.HandleSuperVisionsOpeningHoursDelete))
				})
			})

			// supervision modal
			mux.Group(func(mux chi.Router) {
				mux.Use(CheckPermission("EditOwnSupervisions", "EditAllSupervisions"))
				mux.Get("/supervisions/opening_hours/{id}/supervision", handlers.Make(handlers.Repo.HandleSupervisionsSupervisionModal))
				// supervisions creation and deletion
				mux.Post("/supervisions/{date_id}", handlers.Make(handlers.Repo.HandleSuperVisionsPost))
				mux.Delete("/supervisions/{date_id}/{visitor_id}", handlers.Make(handlers.Repo.HandleSuperVisionsDelete))
			})
		})

		// logistics
		mux.Group(func(mux chi.Router) {
			mux.Use(CheckPermission("ViewLogistics"))
			mux.Get("/logistics", handlers.Make(handlers.Repo.HandleLogisticsIndex))
			mux.Get("/logistics/list", handlers.Make(handlers.Repo.HandleLogisticsList))
			mux.Post("/logistics/list", handlers.Make(handlers.Repo.HandleLogisticsListFilter))
			mux.Group(func(mux chi.Router) {
				mux.Use(CheckPermission("EditLogistics"))
				mux.Get(
					"/logistics/create",
					handlers.Make(handlers.Repo.HandleLogisticsCreateIndex),
				)
				mux.Post(
					"/logistics/create",
					handlers.Make(handlers.Repo.HandleLogisticsCreateAndEdit),
				)
				mux.Delete("/logistics/{id}", handlers.Make(handlers.Repo.HandleLogisticsDelete))
				mux.Get("/logistics/{id}", handlers.Make(handlers.Repo.HandleLogisticsEditIndex))
				mux.Post(
					"/logistics/{id}",
					handlers.Make(handlers.Repo.HandleLogisticsCreateAndEdit),
				)
				// labels
				mux.Post(
					"/logistics/{id}/label",
					handlers.Make(handlers.Repo.HandleLogisticsLabelCreate),
				)
				mux.Post(
					"/logistics/{id}/label/delete",
					handlers.Make(handlers.Repo.HandleLogisticsLabelDelete),
				)
				mux.Post(
					"/logistics/{id}/label/print",
					handlers.Make(handlers.Repo.HandleLogisticsLabelPrint),
				)
				mux.Get("/logistics/billing", handlers.Make(handlers.Repo.HandleLogisticsCreateBillIndex))
				mux.Post("/logistics/billing", handlers.Make(handlers.Repo.HandleLogisticsCreateBillPost))
				mux.Get("/logistics/billing/list", handlers.Make(handlers.Repo.HandleLogisticsBillList))
				mux.Get("/logistics/billing/list/{id}", handlers.Make(handlers.Repo.HandleLogisticsBillList))
			})
		})

		// badges
		mux.Group(func(mux chi.Router) {
			mux.Use(CheckPermission("ViewBadges"))
			mux.Get("/badges", handlers.Make(handlers.Repo.HandleBadgesIndex))
			mux.Get("/badges/list", handlers.Make(handlers.Repo.HandleBadgesList))
			mux.Post("/badges/list", handlers.Make(handlers.Repo.HandleBadgesListFilter))
			// assign badge and view detail
			mux.Group(func(mux chi.Router) {
				mux.Use(CheckPermission("AssignBadge"))
				mux.Get("/badges/{id}", handlers.Make(handlers.Repo.HandleBadgeEditIndex))
				mux.Post("/badges/{id}", handlers.Make(handlers.Repo.HandleBadgeUpdate))
				mux.Post("/badges/reset/{id}", handlers.Make(handlers.Repo.HandleBadgeReset))
				// create, edit and delete badges
				mux.Group(func(mux chi.Router) {
					mux.Use(CheckPermission("EditBadge"))
					mux.Get("/badges/create", handlers.Make(handlers.Repo.HandleBadgeCreateIndex))
					mux.Post("/badges/create", handlers.Make(handlers.Repo.HandleBadgeCreate))
					mux.Get("/badges/scan", handlers.Make(handlers.Repo.HandleBadgeRead))
					mux.Delete("/badges/{id}", handlers.Make(handlers.Repo.HandleBadgeDelete))
				})
			})
		})

		// newsletters
		mux.Group(func(mux chi.Router) {
			mux.Use(CheckPermission("ViewNewsletters"))
			mux.Get("/newsletter", handlers.Make(handlers.Repo.HandleNewsletterIndex))
			mux.Get("/newsletter/list", handlers.Make(handlers.Repo.HandleNewsletterList))
			mux.Group(func(mux chi.Router) {
				mux.Use(CheckPermission("AddNewsletter"))
				mux.Get("/newsletter/create", handlers.Make(handlers.Repo.HandleNewsletterCreateIndex))
				mux.Post("/newsletter/create", handlers.Make(handlers.Repo.HandleNewsletterCreateEdit))
				mux.Post("/newsletter/create/element", handlers.Make(handlers.Repo.HandleNewsletterElementCreate))
				mux.Get("/newsletter/create/loadevents", handlers.Make(handlers.Repo.HandleNewsletterEventsLoad))
				mux.Get("/newsletter/edit/{id}", handlers.Make(handlers.Repo.HandleNewsletterEditIndex))
				mux.Post("/newsletter/edit/{id}", handlers.Make(handlers.Repo.HandleNewsletterCreateEdit))
				mux.Delete("/newsletter/image/{id}", handlers.Make(handlers.Repo.HandleNewsletterImageDelete))
				mux.Get("/newsletter/preview/{id}", handlers.Make(handlers.Repo.HandleNewsletterPreview))
				mux.Get("/newsletter/sendlist/{id}", handlers.Make(handlers.Repo.HandleNewsletterSendList))
				mux.Get("/newsletter/send/{id}", handlers.Make(handlers.Repo.HandleNewsletterSend))
				mux.Get("/newsletter/progress", handlers.Repo.HandleNewsletterProgressBar)
			})
		})

		// abonnements
		mux.Group(func(mux chi.Router) {
			mux.Use(CheckPermission("ViewAbonnements"))
			mux.Get("/abonnements", handlers.Make(handlers.Repo.HandleAbonnementsIndex))
			mux.Get(
				"/abonnements/abonnementlist",
				handlers.Make(handlers.Repo.HandleAbonnementsList),
			)
			mux.Post(
				"/abonnements/abonnementlist",
				handlers.Make(handlers.Repo.HandleAbonnementsListFilter),
			)
			// add abonnements
			mux.Group(func(mux chi.Router) {
				mux.Use(CheckPermission("AddAbonnements"))
				mux.Get(
					"/abonnements/create",
					handlers.Make(handlers.Repo.HandleAbonnementCreateIndex),
				)
				mux.Post("/abonnements/create", handlers.Make(handlers.Repo.HandleAbonnementCreate))
			})
		})
		// this is only basic info for selectboxes, so no permission needed
		mux.Get("/elements/pricing/{id}", handlers.Repo.AbonnementCreatePrice)
		mux.Get("/elements/member/{id}", handlers.Repo.HandleSelectMembership)

		// reminder
		mux.Group(func(mux chi.Router) {
			mux.Use(CheckPermission("ViewReminders"))
			mux.Get("/reminder", handlers.Make(handlers.Repo.HandleReminderIndex))
			mux.Get("/reminder/list", handlers.Make(handlers.Repo.HandleReminderList))
			mux.Post("/reminder/list", handlers.Make(handlers.Repo.HandleReminderListPost))
			mux.Post("/reminder/confirm", handlers.Make(handlers.Repo.HandleReminderSend))
			mux.Get("/reminder/progress", handlers.Repo.HandleReminderProgressBar)
			mux.Get("/reminder/preview/{id}", handlers.Make(handlers.Repo.HandleReminderPreview))
		})

		// statistics
		mux.Group(func(mux chi.Router) {
			mux.Use(CheckPermission("ViewStatistics"))
			mux.Get("/statistics", handlers.Make(handlers.Repo.HandleStatisticsIndex))
			mux.Get("/statistics/charts/keyfigures", handlers.Make(handlers.Repo.HandleKeyFigures))
			mux.Post("/statistics/charts/filter", handlers.Make(handlers.Repo.HandleStatisticsFilter))
			mux.Get("/statistics/charts/visitors", handlers.Make(handlers.Repo.HandleVisitorChart))
			mux.Get("/statistics/charts/weekdays", handlers.Make(handlers.Repo.HandleWeekdayChart))
			mux.Get("/statistics/charts/ateliers", handlers.Make(handlers.Repo.HandleAtelierChart))
			mux.Get("/statistics/charts/ages", handlers.Make(handlers.Repo.HandleAgesChart))
		})

		// finances
		mux.Group(func(mux chi.Router) {
			mux.Use(CheckPermission("ViewFinances"))
			mux.Get("/billing", handlers.Make(handlers.Repo.HandleBillingIndex))
			mux.Get("/billing/list", handlers.Make(handlers.Repo.HandleBillingList))
			mux.Post("/billing/list", handlers.Make(handlers.Repo.HandleBillingListFilter))
			mux.Post("/billing/payment", handlers.Make(handlers.Repo.HandleBillingPayment))
			mux.Delete("/billing/{id}", handlers.Make(handlers.Repo.HandleBillingDelete))
			// create/edit billing
			mux.Group(func(mux chi.Router) {
				mux.Use(CheckPermission("AddFinances"))
				mux.Get("/billing/create", handlers.Make(handlers.Repo.HandleBillCreateIndex))
				mux.Get("/billing/create/item", handlers.Make(handlers.Repo.HandleBillCreateNewItem))
				mux.Post("/billing/create", handlers.Make(handlers.Repo.HandleBillCreateEditPost))
				mux.Get("/billing/edit/{id}", handlers.Make(handlers.Repo.HandleBillEditIndex))
				mux.Post("/billing/edit/{id}", handlers.Make(handlers.Repo.HandleBillCreateEditPost))
			})
		})

		// logs
		mux.Group(func(mux chi.Router) {
			mux.Use(CheckPermission("ViewLogs"))
			mux.Get("/logs", handlers.Make(handlers.Repo.HandleLogIndex))
			mux.Get("/logs/list", handlers.Make(handlers.Repo.HandleLogList))
			mux.Post("/logs/filter", handlers.Make(handlers.Repo.HandleLogsFilter))
		})
	})

	return mux
}
