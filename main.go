package main

/*
Provide a metrics endpoint.
   • Provide a metrics endpoint.
   • Provide a health endpoint.
   • Periodically fetch all pods in k8s cluster and check latency to the pod via ping. Exclude pods that run with hostNetwork: true
*/

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sparrc/go-ping"
)

var (
	listenAddress = flag.String("web.listen-address", ":8080", "Address to listen on for web interface and telemetry.")
	metricsPath   = flag.String("web.metrics-path", "/metrics", "Path under which to expose metrics.")
	labels        []string
	timeout       = flag.Duration("t", time.Second*100000, "")
	interval      = flag.Duration("i", time.Second, "")
	count         = flag.Int("c", -1, "")
)

const (
	defaultTimeout         = 5 * time.Second
	defaultRefreshInterval = 15 * time.Second
)

type _labels struct {
	PodName  string
	PodIP    net.IP
	NodeName string
}

func labelsKeys(prefix string) []string {
	return []string{
		prefix + "pod_name",
		prefix + "pod_ip",
		prefix + "node_name",
	}
}

func (l *_labels) Values() []string {
	return []string{
		l.PodName,
		l.PodIP.String(),
		l.NodeName,
	}
}

type podCollector prometheus.Collector

var (
	metricPingDurations = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "monitoring",
			Name:       "ping_durations_s",
			Help:       "Ping durations in seconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		}, labels,
	)

	counter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "monitoring",
			Name:      "my_counter",
			Help:      "This is my counter",
		})

	summary = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: "monitoring",
			Name:      "my_summary",
			Help:      "This is my summary",
		})
)

func main() {
	start := time.Now()
	flag.Parse()
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	labels = append(labels, labelsKeys("source_")...)
	labels = append(labels, labelsKeys("dest_")...)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		duration := time.Since(start)
		if duration.Seconds() > 10 {
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("error: %v", duration.Seconds())))
		} else {
			w.WriteHeader(200)
			w.Write([]byte("OK."))
		}
	})

	http.Handle("/metrics", promhttp.Handler())
	prometheus.MustRegister(summary)
	prometheus.MustRegister(counter)

	go func() {
		for x := 0; x <= 50; x++ {
			counter.Add(rand.Float64() * 5)
			summary.Observe(rand.Float64() * 10)
			time.Sleep(time.Second)
		}
	}()

	pinger, err := ping.NewPinger("")
	pinger.Count = 1

	for index := 0; index <= 3; index++ {
		pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
		for _, pod := range pods.Items {
			if pod.Spec.HostNetwork == false {
				fmt.Printf("pod name: %v\n", pod.GetName())
				fmt.Printf("namespace: %s\n", pod.GetNamespace())
				fmt.Printf("pod HostNetwork: %v\n", pod.Spec.HostNetwork)
				fmt.Printf("pod IP: %v\n", pod.Status.PodIP)
				fmt.Printf("pod Phase: %v\n", pod.Status.Phase)

				pinger, err := ping.NewPinger(fmt.Sprintf("%v", pod.Status.PodIP))
				if err != nil {
					fmt.Printf("Error at ping.NewPinger: %s\n", err.Error())
					return
				}
				pinger.OnRecv = func(pkt *ping.Packet) {
					fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v\n",
						pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
				}
				pinger.OnFinish = func(stats *ping.Statistics) {
					fmt.Printf("\n--- %s ping statistics ---\n", stats.Addr)
					fmt.Printf("%d packets transmitted, %d packets received, %v%% packet loss\n",
						stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
					fmt.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
						stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
				}

				fmt.Printf("PING %s (%s):\n", pinger.Addr(), pinger.IPAddr())
				pinger.Run()
			} else {
				continue
			}
			fmt.Printf("\n")
		}
		time.Sleep(2 * time.Second)
	}
}
