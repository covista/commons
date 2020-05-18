package database

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	authKeysCreateAttempts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "commons_authorization_keys_create_attempts",
		Help: "Number of attempts to create an authorization key",
	})
	authKeysCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "commons_authorization_keys_created",
		Help: "Total number of authorization keys created successfully",
	})
	addReportAttempts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "commons_add_report_attempts",
		Help: "Number of attempts to add reports",
	})
	addReportSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "commons_add_report_success",
		Help: "Number of successfully added reports",
	})
	getDiagnosisKeysAttempts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "commons_get_diagnosis_keys_attempts",
		Help: "Number of download attempts",
	})
	getDiagnosisKeysSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "commons_get_diagnosis_keys_success",
		Help: "Number of successful download attempts",
	})
	getDiagnosisKeysTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "commons_get_diagnosis_keys_time",
		Help: "Milliseconds elapsed to download diagnosis keys",
	})
)
