package cache

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// CacheEntry represents a cached DNS entry for JSON serialization
type CacheEntry struct {
	Name       string         `json:"name"`
	QType      string         `json:"qtype"`
	Rcode      string         `json:"rcode"`
	TTL        int            `json:"ttl"`
	Stored     time.Time      `json:"stored"`
	Age        int            `json:"age"`
	QueryCount int            `json:"query_count"`
	SourceIPs  map[string]int `json:"source_ips"`
	Answer     []string       `json:"answer,omitempty"`
	Authority  []string       `json:"authority,omitempty"`
	Additional []string       `json:"additional,omitempty"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	PositiveEntries int `json:"positive_entries"`
	NegativeEntries int `json:"negative_entries"`
	TotalEntries    int `json:"total_entries"`
}

// HandleStats returns cache statistics as HTTP handler
func (c *Cache) HandleStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		c.handleStats(w, r)
	}
}

// HandleEntries returns cache entries as HTTP handler
func (c *Cache) HandleEntries() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		c.handleEntries(w, r)
	}
}

func (c *Cache) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := CacheStats{
		PositiveEntries: c.pcache.Len(),
		NegativeEntries: c.ncache.Len(),
		TotalEntries:    c.pcache.Len() + c.ncache.Len(),
	}
	json.NewEncoder(w).Encode(stats)
}

func (c *Cache) handleEntries(w http.ResponseWriter, r *http.Request) {
	entries := []CacheEntry{}
	now := c.now()

	// Walk positive cache
	c.pcache.Walk(func(items map[uint64]any, key uint64) bool {
		if item, ok := items[key].(*item); ok {
			queryCount, sourceIPs := item.getStats()
			entry := CacheEntry{
				Name:       item.Name,
				QType:      dns.TypeToString[item.QType],
				Rcode:      dns.RcodeToString[item.Rcode],
				TTL:        item.ttl(now),
				Stored:     item.stored,
				Age:        int(now.Sub(item.stored).Seconds()),
				QueryCount: queryCount,
				SourceIPs:  sourceIPs,
			}
			for _, rr := range item.Answer {
				entry.Answer = append(entry.Answer, strings.ReplaceAll(rr.String(), "\t", " "))
			}
			for _, rr := range item.Ns {
				entry.Authority = append(entry.Authority, strings.ReplaceAll(rr.String(), "\t", " "))
			}
			for _, rr := range item.Extra {
				entry.Additional = append(entry.Additional, strings.ReplaceAll(rr.String(), "\t", " "))
			}
			entries = append(entries, entry)
		}
		return true
	})

	// Walk negative cache
	c.ncache.Walk(func(items map[uint64]any, key uint64) bool {
		if item, ok := items[key].(*item); ok {
			queryCount, sourceIPs := item.getStats()
			entry := CacheEntry{
				Name:       item.Name,
				QType:      dns.TypeToString[item.QType],
				Rcode:      dns.RcodeToString[item.Rcode],
				TTL:        item.ttl(now),
				Stored:     item.stored,
				Age:        int(now.Sub(item.stored).Seconds()),
				QueryCount: queryCount,
				SourceIPs:  sourceIPs,
			}
			entries = append(entries, entry)
		}
		return true
	})

	json.NewEncoder(w).Encode(entries)
}