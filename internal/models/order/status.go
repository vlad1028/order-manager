package order

type Status string

const (
	Stored        Status = "stored"
	ReachedClient Status = "reached-client"
	Returned      Status = "returned"
	Canceled      Status = "canceled"
)
