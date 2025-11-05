package appctx

// PubSubResponse represents the response from Pub/Sub consumer
type PubSubResponse struct {
	Success bool
	Error   error
	Message string
}

// NewPubSubResponse creates a new PubSubResponse
func NewPubSubResponse() *PubSubResponse {
	return &PubSubResponse{
		Success: true,
	}
}

// WithSuccess sets the success status
func (r *PubSubResponse) WithSuccess(success bool) *PubSubResponse {
	r.Success = success
	return r
}

// WithError sets the error
func (r *PubSubResponse) WithError(err error) *PubSubResponse {
	r.Error = err
	r.Success = false
	return r
}

// WithMessage sets a message
func (r *PubSubResponse) WithMessage(message string) *PubSubResponse {
	r.Message = message
	return r
}
