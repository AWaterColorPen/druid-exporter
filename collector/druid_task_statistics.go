package collector

import (
    "github.com/prometheus/client_golang/prometheus"
)

var (
    taskStatisticsGauge = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "druid_task_statistics",
            Help: "Druid task statistics for k8s hpa",
        },
        []string{"task_status", "pod_name"},
    )
)

func init() {
    prometheus.MustRegister(taskStatisticsGauge)
}

type taskStatistics struct {
    m map[string]map[string]int
}

func (t *taskStatistics) Record(hostName string, taskStatus string) {
    if t.m == nil {
        t.m = map[string]map[string]int{}
    }

    if _, ok := t.m[hostName]; !ok {
        t.m[hostName] = map[string]int{
            "RUNNING": 0,
            "PENDING": 0,
            "SUCCESS": 0,
            "FAILED": 0,
        }
    }

    t.m[hostName][taskStatus]++
}

func (t *taskStatistics) Collect() {
    for k, v := range t.m {
        for x, y := range v {
            taskStatisticsGauge.WithLabelValues(x, k).Set(float64(y))
        }
    }
}
