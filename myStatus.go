package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	termbox "github.com/nsf/termbox-go"
)

func run(Opt *Options) int {
	if Opt.version {
		fmt.Println("myStatusgo version " + VERSION)
		return 0
	}
	if Opt.dbUser == "" || Opt.dbPasswd == "" {
		fmt.Println("Set Username[-u] and Password[-p]")
		return -1
	}

	if Opt.file == "" && Opt.outputJSON {
		fmt.Println("-j option must be with -o option")
		return -1
	}

	var err error
	// Global Variables
	OutputJSON = Opt.outputJSON
	MaxTopNPosition = 0
	GtidMode = Opt.gtid
	HostnameSize = DEFAULTHOSTNAMESIZE
	AllShowMetricFlg = Opt.all
	if Opt.interval < 1 {
		Opt.interval = 1
	}
	// create hostlist
	bufText, exitcode := createHostList(Opt)
	if exitcode == -1 {
		return exitcode
	}

	// Set to struct from Hostlist
	// screenHostlist conteins only comment line and sequence number of host
	hostInfos1 := make([]HostAllInfo, 0)
	hostInfos2 := make([]HostAllInfo, 0)
	screenHostlist := make([]string, 0)
	for _, v := range bufText {
		if v == "" {
			continue
		}
		hArr := strings.Split(v, ":")
		if strings.Contains(hArr[0], "#") {
			screenHostlist = append(screenHostlist, v)
			continue
		}
		if len(hArr) < 2 {
			fmt.Printf("Check Format The Line [%s]\n", v)
			return -1
		}
		alias := aliasNameFind(hArr)
		if len(alias) > HostnameSize {
			HostnameSize = len(alias)
		}
		db, err := dbconn(hArr[0], hArr[1], Opt.dbUser, Opt.dbPasswd)
		if err != nil {
			panic(err.Error())
		}
		defer db.Close()
		db.SetMaxOpenConns(2)
		db.SetMaxIdleConns(1)
		db.SetConnMaxLifetime(0)

		hostInfos1 = append(hostInfos1, HostAllInfo{
			Hname:      hArr[0],
			port:       hArr[1],
			alias:      alias,
			dbuser:     Opt.dbUser,
			dbpasswd:   Opt.dbPasswd,
			promPort:   Opt.promPort,
			db:         db,
			PromeParam: make(mapProme, 0),
			MySlave:    make(mapMy, 0), MyStatu: make(mapMy, 0), MyVariable: make(mapMy, 0), MyPerfomanceSchema: make(mapMy, 0), MyInnoDBBuffer: make(mapInnoDBBuffer, 0),
			DigestBase: make(mapDigest, 0), Metric: make(mapPrint, 0), MetricOS: make(mapPrint, 0), MetricJSON: make(mapPrint, 0),
		})
		hostInfos2 = append(hostInfos2, HostAllInfo{
			Hname:      hArr[0],
			port:       hArr[1],
			alias:      alias,
			dbuser:     Opt.dbUser,
			dbpasswd:   Opt.dbPasswd,
			promPort:   Opt.promPort,
			db:         db,
			PromeParam: make(mapProme, 0),
			MySlave:    make(mapMy, 0), MyStatu: make(mapMy, 0), MyVariable: make(mapMy, 0), MyPerfomanceSchema: make(mapMy, 0), MyInnoDBBuffer: make(mapInnoDBBuffer, 0),
			Metric: make(mapPrint, 0), MetricOS: make(mapPrint, 0), MetricJSON: make(mapPrint, 0),
		})
		screenHostlist = append(screenHostlist, fmt.Sprintf("%d", MaxTopNPosition))
		MaxTopNPosition++
	}

	// Create Timer
	t := time.NewTicker(time.Duration(Opt.interval) * time.Second)
	defer t.Stop()

	// Input Keybord
	termbox.Init()
	eventQueue := make(chan termbox.Event)
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()
	defer termbox.Close()

	// initial val
	StartTime = time.Now()
	screenFlg := MODEOFF
	beforeScreenFlg := beforeScreenFlg()
	outputTitleforFile := outputforTitle()
	execCounter := 1
	if AllShowMetricFlg == false {
		TopNPosition = -1
		TopN = Opt.digest
	} else {
		TopNPosition = 0
		TopN = 5
	}
	SortTableFileIOPosition = 1

	// Create OutPut File
	var outputFileName string
	var outputFile *os.File
	if Opt.file != "" {
		outputFileName = Opt.file
		outputFile, err = createOutputFile(outputFileName)
		if err != nil {
			fmt.Println(fmt.Sprintf("Failed Create File[%v]", err))
			return -1
		}
		defer outputFile.Close()
		if Opt.endSec == 0 {
			Opt.endSec = 3600
		}
	}

	// Main Loop
	baseInfo := hostInfos1
	var hostInfos, hostInfosOld []HostAllInfo
	for {
		select {
		case <-t.C:
			if ok := checkAutoStop(Opt.endSec, StartTime); ok {
				return 0
			}
			// switching
			if execCounter%2 == 1 {
				hostInfos = hostInfos1
				hostInfosOld = hostInfos2
			} else {
				hostInfos = hostInfos2
				hostInfosOld = hostInfos1
			}
			wg := &sync.WaitGroup{}
			sElaspTime := time.Now()
			for i := 0; i < len(hostInfos); i++ {
				wg.Add(1)
				go func(ii int) {
					hostInfos[ii].initializeStruct(&hostInfosOld[ii])
					hostInfos[ii].retrieve(screenFlg, &hostInfosOld[ii], &baseInfo[ii])
					hostInfos[ii].calculate(screenFlg, &hostInfosOld[ii])
					hostInfos[ii].ExecTime = time.Now().Format(TIMELAYOUT)
					wg.Done()
				}(i)
			}
			wg.Wait()
			title, out := screen(hostInfos, screenFlg, screenHostlist, sElaspTime)
			// Write File
			if outputFileName != "" {
				if OutputJSON {
					fmt.Fprintln(outputFile, formatJSON(hostInfos))
				} else {
					fmt.Fprintln(outputFile, getMetricForFile(screenFlg, beforeScreenFlg(screenFlg), &title, &out, sElaspTime.Format(TIMELAYOUT), outputTitleforFile(1)))
				}
			}

			if screenFlg == MODEOFF {
				if AllShowMetricFlg == false {
					screenFlg = Opt.modeflg
				} else {
					screenFlg = MODEFILEIOTABLE
				}
			}
			execCounter++
		case ev := <-eventQueue:
			screenFlg = switchModeByKeyEvent(ev, screenFlg)
			switchEachInstanceByKeyEvent(ev, screenFlg)
			// Stop myStatusgo
			if (ev.Type == termbox.EventKey) && (ev.Key == termbox.KeyEsc || ev.Key == termbox.KeySpace) {
				callClear()
				return 0
			}
		}
	}
}
func dbconn(host string, port string, dbuser string, dbpasswd string) (*sql.DB, error) {
	return sql.Open("mysql", dbuser+":"+dbpasswd+"@tcp("+host+":"+port+")/information_schema?readTimeout=300ms&timeout=250ms")
}

func checkAutoStop(endSec int64, startTime time.Time) bool {
	if endSec != 0 && (time.Now().Unix()-startTime.Unix()) > endSec {
		fmt.Printf(FinishFormat, startTime.Format(TIMELAYOUT), time.Now().Format(TIMELAYOUT))
		return true
	}
	return false
}

func (val *HostAllInfo) initializeStruct(old *HostAllInfo) {
	m.Lock()
	defer m.Unlock()
	val.SlaveInfo = make([]SlaveInfo, 0)
	val.DigestMetric = make([]DigestMetric, TopN)
	val.strFileIOperTable = make(mapStrFileIOperTable, 0)
	val.FileIOMetric = make([]FileIOperTable, TopN)
	val.ThreadsInfo = make([]ThreadsInfo, TopN)
	val.InnodbLockInfo = make([]InnodbLockInfo, TopN)
	val.dbErrorPerQuery.addSlowCheckKey()
	val.dbErrorPerQuery = old.dbErrorPerQuery
}
func initializeStructIfdbErrorFileIO(old *HostAllInfo, baseInfo *HostAllInfo) {
	m.Lock()
	defer m.Unlock()
	old.strFileIOperTable = make(mapStrFileIOperTable, 0)
	old.FileIOMetric = make([]FileIOperTable, TopN)
}

func initializeStructIfdbErrorDigest(old *HostAllInfo, baseInfo *HostAllInfo) {
	m.Lock()
	defer m.Unlock()
	baseInfo.DigestBase = make(mapDigest, 0)
	old.DigestMetric = make([]DigestMetric, TopN)
}

func beforeScreenFlg() func(int) int {
	var v int
	return func(now int) int {
		before := v
		v = now
		return before
	}
}

func outputforTitle() func(int) int {
	sum := 0
	return func(v int) int {
		sum = sum + v
		if sum == 20 {
			sum = 1
		}
		return sum
	}
}

func createOutputFile(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func createHostList(Opt *Options) ([]string, int) {
	bufText := make([]string, 0)
	var err error
	//File Type
	if Opt.hostJSON == "" && Opt.hostComma == "" {
		//Args
		if flag.NArg() != 1 {
			fmt.Println("Set HostFile or -h JSON or -hc hostname1:3306,hostname2:3306")
			return nil, -1
		}
		//Open File
		bufText, err = createHostListByFile(flag.Arg(0))
		if err != nil {
			fmt.Println(fmt.Sprintf("Can't Open File [%s]", flag.Arg(0)))
			return nil, -1
		}
	} else if Opt.hostJSON != "" {
		// JSON Type
		bufText = createHostListByJSON(Opt.hostJSON)
		if len(bufText) == 0 {
			fmt.Println(fmt.Sprintf("Check JSON Format [%s]", Opt.hostJSON))
			return nil, -1
		}
	} else if Opt.hostComma != "" {
		// Comma Type
		bufText = createHostListByComma(Opt.hostComma)

	} else {
		fmt.Println("Set HostFile or -h JSON or -hc hostname1:3306,hostname2:3306")
		return nil, -1
	}
	return bufText, 0
}

func createHostListByComma(hostList string) []string {
	bufText := strings.Split(hostList, ",")
	return bufText
}

func createHostListByFile(fName string) ([]string, error) {
	bufText := make([]string, 0)
	fp, err := os.Open(fName)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	// Read line
	vScan := bufio.NewScanner(fp)
	for vScan.Scan() {
		bufText = append(bufText, vScan.Text())
	}
	return bufText, nil
}

func createHostListByJSON(hostList string) []string {
	bufText := make([]string, 0)
	b := []byte(hostList)
	var m hostListJson
	json.Unmarshal(b, &m)
	for _, v := range m.ServiceList {
		bufText = append(bufText, "#"+v.ServiceName)
		for _, v2 := range v.Host {
			bufText = append(bufText, v2)
		}
	}
	return bufText
}

func (dbErrorPerQuery *dbErrorPerQuery) addSlowCheckKey() {
	if dbErrorPerQuery.fileIO.slowCheckKey != "" {
		return
	}
	dbErrorPerQuery.fileIO.slowCheckKey = FILEIOCHECKSLOW
	dbErrorPerQuery.digest.slowCheckKey = DIGESTCHECKSLOW
	dbErrorPerQuery.bufferPool.slowCheckKey = BUFFERPOOLCHECKSLOW
	dbErrorPerQuery.lock.slowCheckKey = LOCKCHECKSLOW
	dbErrorPerQuery.threads.slowCheckKey = THREADCHECKSLOW
	dbErrorPerQuery.perf.slowCheckKey = PERFCHECKSLOW
	dbErrorPerQuery.slave.slowCheckKey = ShowSlaveStatusSQL
	dbErrorPerQuery.status.slowCheckKey = ShowStatusSQL
	dbErrorPerQuery.variables.slowCheckKey = ShowVariableSQL
}

func switchEachInstanceByKeyEvent(ev termbox.Event, screenFlg int) {
	if screenFlg == MODEOFF {
		return
	}
	switch {
	case ev.Type == termbox.EventKey && (screenFlg == MODETHREADS || screenFlg == MODEPERFORMACE || screenFlg == MODEINNOLOCK || screenFlg == MODEFILEIOTABLE || AllShowMetricFlg) && ev.Key == termbox.KeyArrowDown:
		if TopNPosition < MaxTopNPosition-1 {
			TopNPosition++
		}
	case ev.Type == termbox.EventKey && (screenFlg == MODETHREADS || screenFlg == MODEPERFORMACE || screenFlg == MODEINNOLOCK || screenFlg == MODEFILEIOTABLE || AllShowMetricFlg) && ev.Key == termbox.KeyArrowUp:
		if TopNPosition > -1 {
			TopNPosition--
		}
	case ev.Type == termbox.EventKey && ev.Key == termbox.KeyTab && (screenFlg == MODEFILEIOTABLE || AllShowMetricFlg):
		if SortTableFileIOPosition == SORTNUMMAX {
			SortTableFileIOPosition = SORTNUM1
		} else {
			SortTableFileIOPosition++
		}
	case ev.Type == termbox.EventKey && ev.Key == termbox.KeyF12:
		FullSQLStatement = !FullSQLStatement
	}
	if (screenFlg == MODEINNOLOCK || screenFlg == MODEFILEIOTABLE) && TopNPosition == -1 {
		TopNPosition = 0
	}
}
func switchModeByKeyEvent(ev termbox.Event, screenFlg int) int {
	if screenFlg == MODEOFF {
		return MODEOFF
	}
	if AllShowMetricFlg == true {
		return MODEFILEIOTABLE
	}
	switch {
	case ev.Type == termbox.EventKey && ev.Key == termbox.KeyF1:
		return MODENORMAL
	case ev.Type == termbox.EventKey && ev.Key == termbox.KeyF2:
		return MODERESOURCE
	case ev.Type == termbox.EventKey && ev.Key == termbox.KeyF3:
		return MODETHREADS
	case ev.Type == termbox.EventKey && ev.Key == termbox.KeyF4:
		return MODEPERFORMACE
	case ev.Type == termbox.EventKey && ev.Key == termbox.KeyF5:
		return MODESLAVE
	case ev.Type == termbox.EventKey && ev.Key == termbox.KeyF6:
		return MODEROW
	case ev.Type == termbox.EventKey && ev.Key == termbox.KeyF7:
		return MODEINNOLOCK
	case ev.Type == termbox.EventKey && ev.Key == termbox.KeyF8:
		return MODEINNOBUFFER
	case ev.Type == termbox.EventKey && ev.Key == termbox.KeyF9:
		return MODEFILEIOTABLE
	case ev.Type == termbox.EventKey && ev.Key == termbox.KeyArrowLeft:
		if screenFlg > 1 {
			return screenFlg - 1
		}
		return MODEMAX
	case ev.Type == termbox.EventKey && ev.Key == termbox.KeyArrowRight:
		if screenFlg < MODEMAX {
			return screenFlg + 1
		}
		return MODENORMAL
	default:
		return screenFlg
	}
}

func aliasNameFind(hArr []string) string {
	str := hArr[0]
	if len(hArr[0]) > MAXHOSTNAMESIZE {
		str = hArr[0][:MAXHOSTNAMESIZE]
	}
	if len(hArr) == 3 {
		if len(hArr[2]) > MAXHOSTNAMESIZE {
			str = hArr[2][:MAXHOSTNAMESIZE]
		} else {
			str = hArr[2]
		}
	}
	return str
}
