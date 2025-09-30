package cli

type (
	AcceptOrderRequest struct {
		ID             string
		ClientID       string
		ExpirationDate string
		Weight         string
		Cost           string
		Packaging      string
		AddFilm        bool
	}

	CancelOrderRequest struct {
		ID string
	}

	IssueOrderRequest struct {
		IDs []string
	}

	GetOrdersRequest struct {
		ClientID  string
		LocalOnly bool
	}

	AcceptReturnRequest struct {
		ClientID string
		OrderID  string
	}

	GetReturnedRequest struct {
		Page    int
		PerPage int
	}
)
