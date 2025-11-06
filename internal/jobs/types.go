package jobs

// Job type constants
const (
	// JobTypeSendEmail is the job type for sending emails
	JobTypeSendEmail = "email:send"

	// JobTypeGenerateReport is the job type for generating reports
	JobTypeGenerateReport = "report:generate"

	// JobTypeSyncData is the job type for syncing data with external services
	JobTypeSyncData = "data:sync"

	// JobTypeCleanupExpiredData is the job type for cleaning up expired data
	JobTypeCleanupExpiredData = "cleanup:expired"

	// JobTypeProcessWebhook is the job type for processing webhooks
	JobTypeProcessWebhook = "webhook:process"
)
