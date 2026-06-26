package models

import "time"

type LogisticCategory struct {
	ID          int
	CatName     string
	Price       float64
	PriceMember float64
	UpdatedOn   time.Time
	CreatedOn   time.Time
}

type Logistic struct {
	ID             int
	Number         int
	StartDate      time.Time
	EndDate        time.Time
	Category       LogisticCategory
	Visitor        Visitor
	Reminder       time.Time
	Member         bool
	UnitPrice      float64
	Price          float64
	BillID         int
	BillPaid       bool
	Label          string
	Reason         string
	RequestPending bool
	UpdatedOn      time.Time
	CreatedOn      time.Time
}
