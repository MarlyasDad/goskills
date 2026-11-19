package documents

import "<module>/internal/shared"

// CreateResponse — DTO транспорта. Доменные сущности наружу не отдаются.
type CreateResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func fromCreateOutput(out shared.CreateDocumentOutput) CreateResponse {
	return CreateResponse{
		ID:     out.ID,
		Status: out.Status,
	}
}
