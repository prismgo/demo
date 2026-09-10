package queuedemo

import (
	"context"

	"github.com/prismgo/framework/queue"
)

// EncryptionJob records a handled sensitive payload without exposing its value.
type EncryptionJob struct {
	TraceID string `json:"trace_id"`
	Label   string `json:"label"`
	Secret  string `json:"secret"`
	Encrypt bool   `json:"encrypt"`
}

func init() {
	queue.RegisterType[*EncryptionJob]()
}

// Handle proves that the worker transparently decrypted and restored the job.
func (j *EncryptionJob) Handle(context.Context) error {
	RecordTrace(j.TraceID, j.Label+":handled")
	return nil
}

// ShouldEncrypt allows a job to request encrypted payload storage.
func (j *EncryptionJob) ShouldEncrypt() bool { return j.Encrypt }

var _ queue.EncryptedProvider = (*EncryptionJob)(nil)
