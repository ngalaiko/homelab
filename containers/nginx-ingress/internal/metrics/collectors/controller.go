package collectors

import "github.com/prometheus/client_golang/prometheus"

var labelNamesController = []string{"type"}

// ControllerCollector is an interface for the metrics of the Controller
type ControllerCollector interface {
	SetIngressResources(ingressType string, count int)
	Register(registry *prometheus.Registry) error
}

// ControllerMetricsCollector implements the ControllerCollector interface and prometheus.Collector interface
type ControllerMetricsCollector struct {
	ingressResourcesTotal *prometheus.GaugeVec
}

// NewControllerMetricsCollector creates a new ControllerMetricsCollector
func NewControllerMetricsCollector() *ControllerMetricsCollector {
	cc := &ControllerMetricsCollector{
		ingressResourcesTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:      "ingress_resources_total",
				Namespace: metricsNamespace,
				Help:      "Number of handled ingress resources",
			},
			labelNamesController,
		),
	}

	return cc
}

// SetIngressResources sets the value of the ingress resources gauge for a given type
func (cc *ControllerMetricsCollector) SetIngressResources(ingressType string, count int) {
	cc.ingressResourcesTotal.WithLabelValues(ingressType).Set(float64(count))
}

// Describe implements prometheus.Collector interface Describe method
func (cc *ControllerMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	cc.ingressResourcesTotal.Describe(ch)
}

// Collect implements the prometheus.Collector interface Collect method
func (cc *ControllerMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	cc.ingressResourcesTotal.Collect(ch)
}

// Register registers all the metrics of the collector
func (cc *ControllerMetricsCollector) Register(registry *prometheus.Registry) error {
	return registry.Register(cc)
}

// ControllerFakeCollector is a fake collector that implements the ControllerCollector interface
type ControllerFakeCollector struct{}

// NewControllerFakeCollector creates a fake collector that implements the ControllerCollector interface
func NewControllerFakeCollector() *ControllerFakeCollector {
	return &ControllerFakeCollector{}
}

// Register implements a fake Register
func (cc *ControllerFakeCollector) Register(registry *prometheus.Registry) error { return nil }

// SetIngressResources implements a fake SetIngressResources
func (cc *ControllerFakeCollector) SetIngressResources(ingressType string, count int) {}
