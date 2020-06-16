package collector

import (
    "fmt"
    "github.com/prometheus/client_golang/prometheus"
    "gopkg.in/alecthomas/kingpin.v2"
    "sort"
)

var (
    podNamePrefix       = kingpin.Flag("pod_name_prefix", "Pod name prefix to generator pod name label, Env: POD_NAME_PREFIX. (default - druid-middle-manager)").Default("druid-middle-manager").Envar("POD_NAME_PREFIX").String()
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

type taskPerPod struct {
    PodName    string
    taskCount  map[string]int
}

func (t *taskPerPod) availableCount() int {
    return t.taskCount["RUNNING"] + t.taskCount["PENDING"]
}

func (t *taskPerPod) historyCount() int {
    return t.taskCount["SUCCESS"] + t.taskCount["FAILED"]
}

func (t *taskPerPod) collect() {
    for k, v := range t.taskCount {
        taskStatisticsGauge.WithLabelValues(k, t.PodName).Set(float64(v))
    }
}

func generatorTaskPerPod(host string, tasks TasksInterface) *taskPerPod {
    m := map[string]TasksInterface{}
    for _, task := range tasks {
        k := task.Status
        m[k] = append(m[k], task)
    }

    pod := &taskPerPod{
        host,
        map[string]int{
            "RUNNING": 0,
            "PENDING": 0,
            "SUCCESS": 0,
            "FAILED": 0,
        },
    }
    for k, v := range m {
        pod.taskCount[k] = pod.taskCount[k] + len(v)
    }

    return pod
}

func Collect(tasks TasksInterface) {
    m := map[string]TasksInterface{}
    for _, task := range tasks {
        k := task.Location.Host
        m[k] = append(m[k], task)
    }

    var taskPerPods []*taskPerPod
    for k, v := range m {
        taskPerPods = append(taskPerPods, generatorTaskPerPod(k, v))
    }

    sort.SliceStable(taskPerPods, func(i, j int) bool {
        return taskPerPods[i].availableCount() > taskPerPods[j].availableCount() ||
            (taskPerPods[i].availableCount() == taskPerPods[j].availableCount() &&
                taskPerPods[i].historyCount() > taskPerPods[j].historyCount())
    })

    for k, v := range taskPerPods {
        v.PodName = fmt.Sprintf("%v-%v", *podNamePrefix, k)
        v.collect()
    }
}

