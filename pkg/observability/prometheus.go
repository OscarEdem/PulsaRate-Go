package observability

import (
	"sync/atomic"
)

// ------------------------------------------------------------------------------------------------------------------                                                                                                                                                                                #*eddiere

// MetricsCollector tracks rate limiting metrics using lightweight thread-safe atomic counters.
type MetricsCollector struct {
	allowedTotal  int64
	rejectedTotal int64
	leasesTotal   int64
}

// Global default metrics collector instance
var DefaultMetrics = NewMetricsCollector()

// NewMetricsCollector instantiates a new MetricsCollector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

// IncAllowed increments the count of allowed requests.
func (m *MetricsCollector) IncAllowed() {
	atomic.AddInt64(&m.allowedTotal, 1)
}

// IncRejected increments the count of throttled/rejected requests.
func (m *MetricsCollector) IncRejected() {
	atomic.AddInt64(&m.rejectedTotal, 1)
}

// IncLeaseRequests increments the count of Redis batch lease requests.
func (m *MetricsCollector) IncLeaseRequests() {
	atomic.AddInt64(&m.leasesTotal, 1)
}

// Stats returns a snapshot of current metric counters.
func (m *MetricsCollector) Stats() (allowed, rejected, leases int64) {
	return atomic.LoadInt64(&m.allowedTotal), atomic.LoadInt64(&m.rejectedTotal), atomic.LoadInt64(&m.leasesTotal)
}
