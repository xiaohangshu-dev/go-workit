package observability

import (
	"context"
	"time"
)

type HealthRegistry struct {
	options *Options
}

type HealthReport struct {
	Status   HealthStatus           `json:"status"`
	Duration string                 `json:"duration"`
	Checks   map[string]CheckResult `json:"checks,omitempty"`
}

func NewHealthRegistry(options *Options) *HealthRegistry {
	if options == nil {
		options = NewOptions()
	}
	options.normalize()
	return &HealthRegistry{
		options: options,
	}
}

func (r *HealthRegistry) Options() *Options {
	return r.options
}

func (r *HealthRegistry) Liveness(ctx context.Context) HealthReport {
	start := time.Now()
	return HealthReport{
		Status:   HealthStatusHealthy,
		Duration: time.Since(start).String(),
	}
}

func (r *HealthRegistry) Readiness(ctx context.Context) HealthReport {
	return r.Run(ctx, "ready")
}

func (r *HealthRegistry) Run(ctx context.Context, tags ...string) HealthReport {
	start := time.Now()
	checks := r.filterChecks(tags...)
	results := make(map[string]CheckResult, len(checks))
	status := HealthStatusHealthy

	if len(checks) == 0 {
		return HealthReport{
			Status:   status,
			Duration: time.Since(start).String(),
			Checks:   results,
		}
	}

	timeout := r.options.Health.Timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	for _, registration := range checks {
		checkCtx, cancel := context.WithTimeout(ctx, timeout)
		result := r.runCheck(checkCtx, registration)
		cancel()

		results[registration.Name] = result
		status = combineHealthStatus(status, result.Status)
	}

	return HealthReport{
		Status:   status,
		Duration: time.Since(start).String(),
		Checks:   results,
	}
}

func (r *HealthRegistry) filterChecks(tags ...string) []CheckRegistration {
	if len(tags) == 0 {
		return r.options.Health.Checks
	}

	result := make([]CheckRegistration, 0, len(r.options.Health.Checks))
	for _, check := range r.options.Health.Checks {
		if hasAnyTag(check.Tags, tags) {
			result = append(result, check)
		}
	}
	return result
}

func (r *HealthRegistry) runCheck(ctx context.Context, registration CheckRegistration) (result CheckResult) {
	defer func() {
		if recovered := recover(); recovered != nil {
			result = CheckResult{
				Status: HealthStatusUnhealthy,
				Error:  "health check panic",
			}
		}
	}()

	if registration.Check == nil {
		return CheckResult{
			Status: HealthStatusUnhealthy,
			Error:  "health check is nil",
		}
	}

	result = registration.Check(ctx)
	if result.Status == "" {
		result.Status = HealthStatusHealthy
	}
	return result
}

func hasAnyTag(source, targets []string) bool {
	for _, value := range source {
		for _, target := range targets {
			if value == target {
				return true
			}
		}
	}
	return false
}

func combineHealthStatus(current, next HealthStatus) HealthStatus {
	if current == HealthStatusUnhealthy || next == HealthStatusUnhealthy {
		return HealthStatusUnhealthy
	}
	if current == HealthStatusDegraded || next == HealthStatusDegraded {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}
