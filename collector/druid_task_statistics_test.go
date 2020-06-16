package collector_test

import (
    "druid-exporter/collector"
    "testing"
)

func TestCollect(t *testing.T) {
    ts := collector.TasksInterface{
        {
            Location: struct {
                Host    string `json:"host"`
                Port    string `json:"size"`
                TlsPort string `json:"tlsPort"`
            } {Host: "1"},
        },
    }

    collector.Collect(ts)
}
