package metric

import (
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/zeromicro/go-zero/core/proc"
)

// exemplarTraceKey is the key of exemplar traceID.
const exemplarTraceKey = "traceID"

type (
	// A HistogramVecOpts is a histogram vector options.
	HistogramVecOpts struct {
		Namespace   string
		Subsystem   string
		Name        string
		Help        string
		Labels      []string
		Buckets     []float64
		ConstLabels map[string]string
	}

	// A HistogramVec interface represents a histogram vector.
	HistogramVec interface {
		// Observe adds observation v to labels.
		Observe(v int64, labels ...string)
		// ObserveFloat allow to observe float64 values.
		ObserveFloat(v float64, labels ...string)
		// ObserveWithExemplar adds observation v to labels with exemplar.
		ObserveWithExemplar(value float64, exemplar prom.Labels, labels ...string)
		// ObserveWithTraceExemplar adds observation v to labels with traceID exemplar.
		ObserveWithTraceExemplar(v float64, traceID string, labels ...string)
		close() bool
	}

	promHistogramVec struct {
		histogram *prom.HistogramVec
	}
)

// NewHistogramVec returns a HistogramVec.
func NewHistogramVec(cfg *HistogramVecOpts) HistogramVec {
	if cfg == nil {
		return nil
	}

	vec := prom.NewHistogramVec(prom.HistogramOpts{
		Namespace:   cfg.Namespace,
		Subsystem:   cfg.Subsystem,
		Name:        cfg.Name,
		Help:        cfg.Help,
		Buckets:     cfg.Buckets,
		ConstLabels: cfg.ConstLabels,
	}, cfg.Labels)
	prom.MustRegister(vec)
	hv := &promHistogramVec{
		histogram: vec,
	}
	proc.AddShutdownListener(func() {
		hv.close()
	})

	return hv
}

func (hv *promHistogramVec) Observe(v int64, labels ...string) {
	update(func() {
		hv.histogram.WithLabelValues(labels...).Observe(float64(v))
	})
}

func (hv *promHistogramVec) ObserveFloat(v float64, labels ...string) {
	update(func() {
		hv.histogram.WithLabelValues(labels...).Observe(v)
	})
}

func (hv *promHistogramVec) ObserveWithExemplar(v float64, exemplar prom.Labels, labels ...string) {
	update(func() {
		hv.histogram.WithLabelValues(labels...).(prom.ExemplarObserver).ObserveWithExemplar(v, exemplar)
	})
}

func (hv *promHistogramVec) ObserveWithTraceExemplar(v float64, traceID string, labels ...string) {
	hv.ObserveWithExemplar(v, prom.Labels{exemplarTraceKey: traceID}, labels...)
}

func (hv *promHistogramVec) close() bool {
	return prom.Unregister(hv.histogram)
}
