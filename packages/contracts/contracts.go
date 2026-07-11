// Package contracts holds shared DTOs aligned with JSON schemas.
package contracts

import "github.com/fastygo/lab/packages/domain"

// Re-export domain report types for consumers that depend on contracts only.
type (
	Finding  = domain.Finding
	Decision = domain.Decision
	Report   = domain.Report
	Manifest = domain.Manifest
	RunEvent = domain.RunEvent
)

const APIVersion = "lab.fastygo.dev/v1"
