package services

import (
	"sort"
	"time"

	"home-provider/internal/database"
	"home-provider/internal/models"

	"github.com/google/uuid"
)

type UsageTracker struct{}

var usageTracker = &UsageTracker{}

func NewUsageTracker() *UsageTracker {
	return usageTracker
}

type UsageRecord struct {
	APIKeyID     string
	Provider     string
	Model        string
	InputTokens  int
	OutputTokens int
	LatencyMs    int
	StatusCode   int
}

func (ut *UsageTracker) Log(record UsageRecord) error {
	log := models.UsageLog{
		ID:           uuid.New().String(),
		APIKeyID:     record.APIKeyID,
		Provider:     record.Provider,
		Model:        record.Model,
		InputTokens:  record.InputTokens,
		OutputTokens: record.OutputTokens,
		LatencyMs:    record.LatencyMs,
		StatusCode:   record.StatusCode,
		CreatedAt:    time.Now(),
	}

	var logs []models.UsageLog
	database.ReadJSON("./data/usage.json", &logs)
	logs = append(logs, log)
	return database.WriteJSON("./data/usage.json", logs)
}

type UsageStats struct {
	TotalRequests     int64
	TotalInputTokens  int
	TotalOutputTokens int
	AvgLatencyMs      int
}

type GlobalStats struct {
	TotalRequests     int64 `json:"total_requests"`
	TotalInputTokens  int   `json:"total_input_tokens"`
	TotalOutputTokens int   `json:"total_output_tokens"`
	AvgLatencyMs      int   `json:"avg_latency_ms"`
}

type KeyStats struct {
	KeyID         string `json:"key_id"`
	KeyName       string `json:"key_name"`
	KeyPrefix     string `json:"key_prefix"`
	TotalRequests int64  `json:"total_requests"`
	InputTokens   int    `json:"input_tokens"`
	OutputTokens  int    `json:"output_tokens"`
	AvgLatencyMs  int    `json:"avg_latency_ms"`
}

type TimeSeriesPoint struct {
	Timestamp       string `json:"timestamp"`
	Requests        int    `json:"requests"`
	InputTokens     int    `json:"input_tokens"`
	OutputTokens    int    `json:"output_tokens"`
	CumRequests     int    `json:"cum_requests"`
	CumInputTokens  int    `json:"cum_input_tokens"`
	CumOutputTokens int    `json:"cum_output_tokens"`
}

type UsageResponse struct {
	Global     GlobalStats       `json:"global"`
	ByKey      []KeyStats        `json:"by_key"`
	TimeSeries []TimeSeriesPoint `json:"time_series"`
}

func (ut *UsageTracker) GetStats(days int) (UsageResponse, error) {
	since := time.Now().AddDate(0, 0, -days)

	var logs []models.UsageLog
	database.ReadJSON("./data/usage.json", &logs)

	keyMap := make(map[string]*KeyStatsAccum)
	keyNames := make(map[string]string)
	keyPrefixes := make(map[string]string)

	var globalTotalRequests int64
	var globalInput, globalOutput, globalLatency int

	var timeSeriesMap = make(map[string]*TimeSeriesAccum)

	for _, log := range logs {
		if !log.CreatedAt.After(since) {
			continue
		}

		globalTotalRequests++
		globalInput += log.InputTokens
		globalOutput += log.OutputTokens
		globalLatency += log.LatencyMs

		if _, ok := keyMap[log.APIKeyID]; !ok {
			keyMap[log.APIKeyID] = &KeyStatsAccum{}
		}
		acc := keyMap[log.APIKeyID]
		acc.Requests++
		acc.InputTokens += log.InputTokens
		acc.OutputTokens += log.OutputTokens
		acc.LatencyMs += log.LatencyMs

		minuteKey := log.CreatedAt.Format("2006-01-02T15:04:00Z")
		if _, ok := timeSeriesMap[minuteKey]; !ok {
			timeSeriesMap[minuteKey] = &TimeSeriesAccum{}
		}
		ts := timeSeriesMap[minuteKey]
		ts.Requests++
		ts.InputTokens += log.InputTokens
		ts.OutputTokens += log.OutputTokens
	}

	keyManager := NewKeyManager()
	keys, _ := keyManager.List()
	for _, k := range keys {
		keyNames[k.ID] = k.Name
		keyPrefixes[k.ID] = k.KeyPrefix
	}

	var byKey []KeyStats
	for keyID, acc := range keyMap {
		avgLat := 0
		if acc.Requests > 0 {
			avgLat = acc.LatencyMs / int(acc.Requests)
		}
		byKey = append(byKey, KeyStats{
			KeyID:         keyID,
			KeyName:       keyNames[keyID],
			KeyPrefix:     keyPrefixes[keyID],
			TotalRequests: acc.Requests,
			InputTokens:   acc.InputTokens,
			OutputTokens:  acc.OutputTokens,
			AvgLatencyMs:  avgLat,
		})
	}

	sort.Slice(byKey, func(i, j int) bool {
		return byKey[i].TotalRequests > byKey[j].TotalRequests
	})

	var timeSeries []TimeSeriesPoint
	times := make([]string, 0, len(timeSeriesMap))
	for t := range timeSeriesMap {
		times = append(times, t)
	}
	sort.Strings(times)

	var cumRequests, cumInputTokens, cumOutputTokens int
	for _, t := range times {
		acc := timeSeriesMap[t]
		cumRequests += acc.Requests
		cumInputTokens += acc.InputTokens
		cumOutputTokens += acc.OutputTokens
		timeSeries = append(timeSeries, TimeSeriesPoint{
			Timestamp:       t,
			Requests:        acc.Requests,
			InputTokens:     acc.InputTokens,
			OutputTokens:    acc.OutputTokens,
			CumRequests:     cumRequests,
			CumInputTokens:  cumInputTokens,
			CumOutputTokens: cumOutputTokens,
		})
	}

	avgLat := 0
	if globalTotalRequests > 0 {
		avgLat = globalLatency / int(globalTotalRequests)
	}

	return UsageResponse{
		Global: GlobalStats{
			TotalRequests:     globalTotalRequests,
			TotalInputTokens:  globalInput,
			TotalOutputTokens: globalOutput,
			AvgLatencyMs:      avgLat,
		},
		ByKey:      byKey,
		TimeSeries: timeSeries,
	}, nil
}

type KeyStatsAccum struct {
	Requests     int64
	InputTokens  int
	OutputTokens int
	LatencyMs    int
}

type TimeSeriesAccum struct {
	Requests     int
	InputTokens  int
	OutputTokens int
}

func (ut *UsageTracker) GetStatsByAPIKey(apiKeyID string, days int) (UsageStats, error) {
	since := time.Now().AddDate(0, 0, -days)

	var logs []models.UsageLog
	database.ReadJSON("./data/usage.json", &logs)

	var count int64
	var totalInput, totalOutput, totalLatency int

	for _, log := range logs {
		if log.APIKeyID == apiKeyID && log.CreatedAt.After(since) {
			count++
			totalInput += log.InputTokens
			totalOutput += log.OutputTokens
			totalLatency += log.LatencyMs
		}
	}

	avgLatency := 0
	if count > 0 {
		avgLatency = totalLatency / int(count)
	}

	return UsageStats{
		TotalRequests:     count,
		TotalInputTokens:  totalInput,
		TotalOutputTokens: totalOutput,
		AvgLatencyMs:      avgLatency,
	}, nil
}
