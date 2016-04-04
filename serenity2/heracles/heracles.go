/*
git submodule update --init
(cd serenity2/heracles/mutilate/; scons)
(cd serenity2/heracles/memcached/; ./autogen.sh && ./configure && make)
sudo -Es
go run ./serenity2/heracles/heracles.go
*/
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	memcachedbin = "./serenity2/heracles/memcached/memcached"
	mutilatebin  = "./serenity2/heracles/mutilate/mutilate"

	// mutilate options
	timeout = "1" // mutilate --time x
	server  = "127.0.0.1"
)

var (
	// quit brodcast signal
	quit   = make(chan struct{})
	numcpu = runtime.NumCPU()
)

// expects having heracles db a creates points in heracles measurments
func store(key string, value float64) {
	log.Printf("%s = %v\n", key, value)

	// https://docs.influxdata.com/influxdb/v0.9/write_protocols/line/
	point := fmt.Sprintf("heracles %s=%f", key, value)

	resp, err := http.Post("http://127.0.0.1:8086/write?db=heracles",
		"",
		bytes.NewBufferString(point),
	)
	check(err)
	if resp.StatusCode != http.StatusNoContent { //204
		body, err := ioutil.ReadAll(resp.Body)
		check(err)
		log.Printf("error body = %s\n", body)
	}
}

// search for qps
func mutilateSearch(percentile, latencyUs int) (qps int) {

	cmd := exec.Command(mutilatebin,
		"--search", fmt.Sprintf("%d:%d", percentile, latencyUs),
		"--server", server,
		"--time", timeout, // just for one second
		"--threads", strconv.Itoa(numcpu),
		"--connections", strconv.Itoa(numcpu),
	)
	output, err := cmd.Output()
	if ee, ok := err.(*exec.ExitError); ok {
		log.Println(string(ee.Stderr))
	}
	check(err)

	re := regexp.MustCompile(`Total QPS = (\d+)`)
	qpsRaw := re.FindSubmatch(output)[1]
	fmt.Printf("target qps = %s (for percentile=%d latency=%d)\n", qpsRaw, percentile, latencyUs)
	return atoi(string(qpsRaw))
}

func parseQpsSli(mutilateOutput []byte) (sli float64) {
	// parse mutilate --qps output
	scanner := bufio.NewScanner(bytes.NewReader(mutilateOutput))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "#type") && !strings.Contains(line, "read") {
			continue
		}
		fields := strings.Fields(line)
		p99 := atof(fields[8])
		return p99
	}
	check(scanner.Err())
	return
}

func mutilateQps(qps int) (sli float64) {

	cmd := exec.Command(mutilatebin,
		"--qps", fmt.Sprintf("%d", qps),
		"--server", server,
		"--time", timeout,
		"--threads", strconv.Itoa(numcpu),
		"--connections", strconv.Itoa(numcpu),
	)
	output, err := cmd.Output()
	if ee, ok := err.(*exec.ExitError); ok {
		log.Println(string(ee.Stderr))
	}
	check(err)

	sli = parseQpsSli(output)
	// log.Println("qps sli =", sli)
	return sli
}

func parseScanSlis(mutilateOutput []byte) (slis map[float64]float64) {
	// parse mutilate scan output
	slis = make(map[float64]float64)

	scanner := bufio.NewScanner(bytes.NewReader(mutilateOutput))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "#type") && !strings.Contains(line, "read") {
			continue
		}
		fields := strings.Fields(line)
		p99 := atof(fields[8])
		qps := atof(fields[9])

		slis[qps] = p99

	}
	check(scanner.Err())
	return
}

// mutilate scan over given load points (qps) result map qps -> 99th percentials in us
func mutilateScan(min, max, step int) (slis map[float64]float64) {
	output, err := exec.Command(mutilatebin,
		"--scan", fmt.Sprintf("%d:%d:%d", min, max, step),
		"--server", server,
		"--time", timeout, // just for one second
		"--threads", strconv.Itoa(numcpu),
		"--connections", strconv.Itoa(numcpu),
	).Output()
	check(err)

	slis = parseScanSlis(output)

	return
}

func cpucores(cgroup string, cores int) {
	store(cgroup+"_cores", float64(cores))
	cpuset(cgroup, fmt.Sprintf("0-%d", cores-1))
}

// value like 0, 1, 0-1, 0,1,2,3
// https://www.kernel.org/doc/Documentation/cgroup-v1/cpusets.txt
// zero based
func cpuset(cgroup, value string) {

	err := ioutil.WriteFile("/sys/fs/cgroup/cpuset/"+cgroup+"/cpuset.cpus", []byte(value), os.ModePerm)
	check(err)
	// err = ioutil.WriteFile("/sys/fs/cgroup/cpuset/"+cgroup+"/cpuset.mems", []byte(value), os.ModePerm)
	// check(err)
}

// memcache start memcache daemon
func memcache(threads int) {
	u, err := user.Current()
	check(err)

	log.Println("memcache starting...")
	cmd := exec.Command(memcachedbin,
		"-u", u.Name,
		"-t", strconv.Itoa(threads),
	)
	err = cmd.Start()
	check(err)
	log.Println("memcached pid =", cmd.Process.Pid, "threads =", threads)

	log.Println("memcache put into prod cgroup")
	err = ioutil.WriteFile("/sys/fs/cgroup/cpuset/prod/tasks", []byte(strconv.Itoa(cmd.Process.Pid)), os.ModePerm)
	check(err)

	select {
	case <-quit:
		err = cmd.Process.Kill()
		log.Println("memcache kill")
		check(err)
	}

	err = cmd.Wait()
	if ee, ok := err.(*exec.ExitError); ok {
		log.Println("memcache exit error: ", string(ee.Stderr))
	}
	check(err)

}

func be(cores int, bequit chan struct{}) {

	// sid, err := syscall.Setsid()
	// check(err)
	// log.Println("start be with new sid", sid)
	var err error

	cmd := exec.Command("stress", "--cpu", strconv.Itoa(cores))
	err = cmd.Start()
	// race ??? fork?
	pid := cmd.Process.Pid
	log.Println("be with pid", pid)
	time.Sleep(10 * time.Second)

	err = syscall.Setpgid(pid, pid)
	check(err)

	check(err)
	select {
	case <-bequit:
		// err = cmd.Process.Signal(syscall.SIGINT) // not just to parent process but to whole process group (negative pid)
		err = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		log.Println("sent interapt signal to ", cmd.Process.Pid)
		check(err)
	}
	err = cmd.Wait() // ignore killed signal
	if err != nil {
		log.Println(err)
	}
	log.Println("be end")
}

// utility functions
func check(err error) {
	if err != nil {
		panic(err)
	}
}

func cgcreate(cgroup string) {
	err := os.MkdirAll("/sys/fs/cgroup/cpuset/"+cgroup, os.ModePerm)
	check(err)
}

func atoi(s string) int {
	i, err := strconv.Atoi(string(s))
	check(err)
	return i
}

func atof(raw string) float64 {
	f, err := strconv.ParseFloat(raw, 64)
	check(err)
	return f
}

func avg(slis map[float64]float64) float64 {

	var sum float64
	for _, v := range slis {
		sum += v
	}
	return sum / float64(len(slis))

}

// exp 1 sensitivy profile
func exp1sp() {

	cgcreate("prod")
	cgcreate("be")
	cpuset("prod", "0")
	cpuset("be", "0")

	// prod
	go memcache(numcpu)

	// load generator
	qps := mutilateSearch(99, 1000)
	fmt.Printf("target qps = %+v\n", qps)

	repeat := 1

	// slis
	for i := 0; i < repeat; i++ {
		slis := mutilateScan(0, qps, qps/2)
		fmt.Printf("avg=%f slis=%+v\n", avg(slis), slis)
	}

	// with
	for core := 1; core <= runtime.NumCPU(); core++ {
		log.Println("be", core)
		bequit := make(chan struct{})
		go be(core, bequit)
		// slis
		for i := 0; i < repeat; i++ {
			slis := mutilateScan(0, qps, qps/2)
			fmt.Printf("avg=%f slis=%+v\n", avg(slis), slis)
		}
		close(bequit)

	}

	//
	// close all
	close(quit)
}

// exp 2 be controlling
func exp2be() {
	bequit := make(chan struct{})
	go be(1, bequit)
	log.Println("sleep")
	time.Sleep(45 * time.Second)
	log.Println("quit")
	close(bequit)
	log.Println("sleep again")
	time.Sleep(15 * time.Second)
	log.Println("exit")
}

// exp 3
func exp3prodalone() {

	// algorithm controllers
	algo := func(sli float64) {
		if sli > 6000 {
			cpucores("prod", numcpu)
		} else {
			cpucores("prod", 1)
		}
	}

	cgcreate("prod")
	cpucores("prod", numcpu)

	// prod
	go memcache(numcpu)

	// load generator
	qps := mutilateSearch(95, 1000)

	for {
		sli := mutilateQps(qps)
		store("sli", sli)

		algo(sli)

		time.Sleep(1 * time.Second)
	}
}

// exp 3
func exp3prodalone2() {

	up := true
	cores := 1

	// algorithm controllers
	algo := func(sli float64) {
		cpucores("prod", cores)
		if up {
			cores += 1
		} else {
			cores -= 1
		}
		if cores == numcpu || cores == 1 {
			up = !up
		}
	}

	cgcreate("prod")
	cpucores("prod", numcpu)

	// prod
	go memcache(numcpu)

	// load generator
	qps := mutilateSearch(95, 1000)

	for {
		sli := mutilateQps(qps)
		store("sli", sli)

		algo(sli)

		time.Sleep(1 * time.Second)
	}
}

func exp4heracles() {

	// algorithm controllers
	algo := func(sli float64) {
		if sli < 15.0 {
			cpucores("be", 1)
		} else {
			cpucores("be", numcpu/2)
		}
	}

	cgcreate("prod")
	cgcreate("be")
	cpucores("prod", numcpu)
	cpucores("be", numcpu)

	// start prod
	go memcache(numcpu)

	// start be
	go be(numcpu, quit)

	// load generator
	qps := mutilateSearch(99, 1000)
	fmt.Printf("target qps = %+v\n", qps)

	for {
		sli := mutilateQps(qps)
		store("sli", sli)

		algo(sli)

		time.Sleep(1 * time.Second)
	}
}

func main() {
	// exp1sp()
	// exp2be()
	// exp3prodalone()
	exp3prodalone2()
	// exp4heracles()
}
