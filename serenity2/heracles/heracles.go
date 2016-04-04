/*
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
	quit = make(chan struct{})
)

// search for qps
func mutilateSearch(percentile, latencyUs int) (qps int) {

	cmd := exec.Command(mutilatebin,
		"--search", fmt.Sprintf("%d:%d", percentile, latencyUs),
		"--server", server,
		"--time", timeout, // just for one second
	)
	output, err := cmd.Output()
	if ee, ok := err.(*exec.ExitError); ok {
		log.Println(string(ee.Stderr))
	}
	check(err)

	re := regexp.MustCompile(`Total QPS = (\d+)`)
	qpsRaw := re.FindSubmatch(output)[1]
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
	)
	output, err := cmd.Output()
	if ee, ok := err.(*exec.ExitError); ok {
		log.Println(string(ee.Stderr))
	}
	check(err)

	sli = parseQpsSli(output)
	log.Println("qps sli =", sli)
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
	).Output()
	check(err)

	slis = parseScanSlis(output)

	return
}

func cpuset(cgroup, value string) {
	err := ioutil.WriteFile("/sys/fs/cgroup/cpuset/"+cgroup+"/cpuset.cpus", []byte(value), os.ModePerm)
	check(err)
	// err = ioutil.WriteFile("/sys/fs/cgroup/cpuset/"+cgroup+"/cpuset.mems", []byte(value), os.ModePerm)
	// check(err)
}

// memcache start memcache daemon
func memcache() {
	u, err := user.Current()
	check(err)

	log.Println("memcache starting...")
	cmd := exec.Command(memcachedbin, "-u", u.Name)
	err = cmd.Start()
	check(err)

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
	go memcache()

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
func exp3heracles() {

	// algorithm controllers
	algo := func(sli float64) {
		if sli < 15.0 {
			log.Println("set cpu 1")
			cpuset("be", "1") // one core
		} else {
			log.Println("set cpu 0")
			cpuset("be", "0") // all cores
		}
	}

	cgcreate("prod")
	cgcreate("be")
	cpuset("prod", "0")
	cpuset("be", "0")

	// prod
	go memcache()

	// load generator
	qps := mutilateSearch(99, 1000)
	fmt.Printf("target qps = %+v\n", qps)

	for {
		sli := mutilateQps(qps)

		algo(sli)

		time.Sleep(1 * time.Second)
	}
}

func main() {
	// exp1sp()
	// exp2be()
	exp3heracles()
}
