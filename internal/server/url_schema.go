package server

type urlSchema struct {
	URL string `json:"url"`
}

type responseSchema struct {
	Result string `json:"result"`
}

type batchResponseSchema struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
