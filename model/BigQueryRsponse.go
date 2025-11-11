package model

type BigQueryResponse struct {
	Replies      []string `json:"replies"`
	ErrorMessage string   `json:"errorMessage"`
}

func NewBigQueryResponse(replies []string) BigQueryResponse {
	return BigQueryResponse{
		Replies: replies,
	}
}

func NewBigQueryResponseError(errorMessage string) BigQueryResponse {
	return BigQueryResponse{
		ErrorMessage: errorMessage,
	}
}
