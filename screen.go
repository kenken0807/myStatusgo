package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	tm "github.com/buger/goterm"
	color "github.com/fatih/color"
)

func metricColoring(metrics mapPrint) {
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	magenta := color.New(color.FgMagenta).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	for key, val := range metrics {
		v := val.(string)
		switch key {
		case DISKREAD, DISKWRITE, MEMUSED, MEMCACHE, MEMTOL, MEMBUF, MEMFREE, INBOUND, OUTBOUND, SWAPUSED, SWAPFRE, BUFPOOLTOTAL, BUFPOOLDATA, BUFPOOLDIRTY, BUFPOOLFREE, LSNMINUSCKPOINT:
			switch {
			case strings.Contains(v, "k"):
				metrics[key] = yellow(v)
			case strings.Contains(v, "M"):
				metrics[key] = green(v)
			case strings.Contains(v, "G"):
				metrics[key] = magenta(v)
			default:
				metrics[key] = white(v)
			}
		case CPUUSER, CPUSYS, CPUIO, CPUIDLE, CPUIRQ:
			i, _ := strconv.Atoi(v)
			switch {
			case i == 0:
				metrics[key] = white(v)
			case 0 < i && i <= 25:
				metrics[key] = cyan(v)
			case 25 < i && i <= 50:
				metrics[key] = yellow(v)
			case 50 < i && i <= 75:
				metrics[key] = red(v)
			case 75 < i && i <= 100:
				metrics[key] = green(v)
			}
		case RDONLY, RPLSEMIMAS, RPLSEMISLV:
			if v == "OFF" {
				metrics[key] = white(v)
			} else {
				metrics[key] = cyan(v)
			}
		case SLWLOG, SECBEHMAS, PSDIGESTLOST:
			i, _ := strconv.Atoi(v)
			if i != 0 {
				metrics[key] = red(v)
			} else {
				metrics[key] = white(v)
			}
		case AVGLATENCY, MAXLATENCY:
			switch {
			case strings.Contains(v, "ns"):
				metrics[key] = white(v)
			case strings.Contains(v, "us"):
				metrics[key] = blue(v)
			case strings.Contains(v, "ms"):
				metrics[key] = green(v)
			case strings.Contains(v, "s"):
				metrics[key] = yellow(v)
			default:
				metrics[key] = red(v)
			}

		case LOADAVG1, LOADAVG5, LOADAVG15:
			i := changeStr2Int(v)
			switch {
			case i < 0:
				metrics[key] = red(v)
			case 0 <= i && i <= 1:
				metrics[key] = white(v)
			case 1 < i && i <= 3:
				metrics[key] = cyan(v)
			case 3 < i && i <= 6:
				metrics[key] = green(v)
			case 6 < i && i <= 10:
				metrics[key] = yellow(v)
			case 10 < i:
				metrics[key] = red(v)
			}
		case THRUN, SLWSUM:
			i, _ := strconv.Atoi(v)
			switch {
			case i < 0:
				metrics[key] = red(v)
			case 0 <= i && i <= 10:
				metrics[key] = white(v)
			case 10 < i && i <= 30:
				metrics[key] = cyan(v)
			case 30 < i && i <= 100:
				metrics[key] = yellow(v)
			case 100 < i:
				metrics[key] = red(v)
			}
		case BUFDATAP, BUFFREEP, BUFMISCP:
			metrics[key] = white(v)
		default:
			i, _ := strconv.Atoi(v)
			switch {
			case i < 0:
				metrics[key] = red(v)
			case i == 0:
				metrics[key] = white(v)
			case 0 < i && i <= 500:
				metrics[key] = cyan(v)
			case 500 < i && i <= 3000:
				metrics[key] = green(v)
			case 3000 < i && i <= 10000:
				metrics[key] = yellow(v)
			case 100 < i:
				metrics[key] = red(v)
			}
		}
	}
}

func formatJSON(hostInfos []HostAllInfo) string {
	var buf bytes.Buffer
	var b []byte
	var o string
	for i := 0; i < len(hostInfos); i++ {
		b, _ = json.Marshal(hostInfos[i])
		buf.Write(b)
		o += buf.String()
	}
	return o
}

func sendJSON(metric mapPrint) ([]byte, error) {
	outputJSON, err := json.Marshal(metric)
	if err != nil {
		panic(err)
	}
	return outputJSON, err
}

func (val *HostAllInfo) formatDetailEachInstance(idx int) (string, string, string, string) {
	listDigestText, listThreadText, listInnodbLockText, listFileIOTableText := "", "", "", ""
	errDigestText, errThreadText, errInnodbLockText, errFileIOTableText := "", "", "", ""
	if TopNPosition == -1 || idx == TopNPosition {
		errDigestText = formatErr(val.dbErrorPerQuery.digest, val.dberror, true)
		errThreadText = formatErr(val.dbErrorPerQuery.threads, val.dberror, true)
		errInnodbLockText = formatErr(val.dbErrorPerQuery.lock, val.dberror, true)
		errFileIOTableText = formatErr(val.dbErrorPerQuery.fileIO, val.dberror, true)
		for i := 0; i < TopN; i++ {
			listDigestText += val.DigestMetric[i].formatTopN(i)
			listThreadText += val.ThreadsInfo[i].formatTopN(i)
			listInnodbLockText += val.InnodbLockInfo[i].formatTopN(i)
			listFileIOTableText += val.FileIOMetric[i].formatTopN(i)
			if TopNPosition == -1 {
				break
			}
		}
	}
	if errDigestText != "" {
		listDigestText = errDigestText
	}
	if errThreadText != "" {
		listThreadText = errThreadText
	}
	if errInnodbLockText != "" {
		listInnodbLockText = errInnodbLockText
	}
	if errFileIOTableText != "" {
		listFileIOTableText = errFileIOTableText
	}
	return listDigestText, listThreadText, listInnodbLockText, listFileIOTableText
}

func formatErr(dbError dbError, dberror bool, instanceType bool) string {
	red := color.New(color.FgRed).SprintFunc()
	cursor := ""
	if instanceType {
		cursor = INSTANCECURSOR
	}
	if dberror == DBERROR {
		return fmt.Sprintf("%s %s", cursor, Error)
	}
	if dbError.errStatus == DBERROR {
		return fmt.Sprintf("%s %s", cursor, Error)
	}
	if dbError.errMessage != "" {
		return fmt.Sprintf("%s %s", cursor, red(dbError.errMessage))
	}
	return ""
}
func (val *HostAllInfo) formatMySQLMetric(idx int) (string, string, string, string, string, string, string, string, string) {
	listDigestText, listThreadText, listInnodbLockText, listFileIOTableText := val.formatDetailEachInstance(idx)
	// show slave status
	sInfo, normalIoTh, normalSQLTh := val.formatSlaveInfoWithColoring()
	vNormal := fmt.Sprintf("%13s %12s %14s %15s %15s %15s %15s %16s %15s %13s %15s %16s %15s %14s %13s %13s %13s %15s %12s %12s",
		val.Metric[THCON],
		val.Metric[THRUN],
		val.Metric[ABORTCON],
		val.Metric[CSEL],
		val.Metric[CUPD],
		val.Metric[CINS],
		val.Metric[CDEL],
		val.Metric[CRLCE],
		val.Metric[QCACHHIT],
		val.Metric[CCALPROCEDURE],
		val.Metric[QPSALL],
		val.Metric[CSTMT],
		val.Metric[CCOMMIT],
		val.Metric[CROLLBACK],
		val.Metric[SLWLOG],
		val.Metric[SLWSUM],
		val.Metric[RDONLY],
		val.Metric[SECBEHMAS],
		normalIoTh,
		normalSQLTh)
	vHandler := fmt.Sprintf(" %17s %17s %17s %17s %17s %17s %17s %17s %17s %17s %17s",
		val.Metric[INNOROWRD],
		val.Metric[INNOROWINS],
		val.Metric[INNOROWUPD],
		val.Metric[INNOROWDEL],
		val.Metric[HADRRDFIRST],
		val.Metric[HADRRDKEY],
		val.Metric[HADRRDLAST],
		val.Metric[HADRRDNXT],
		val.Metric[HADRRDPRV],
		val.Metric[HADRRDRND],
		val.Metric[HADRRDRNDNXT])
	vPerf := val.formatPerf()
	vBuf := val.formatInnoDBBuffer()
	return vNormal, vHandler, sInfo, vPerf, listDigestText, listThreadText, listInnodbLockText, vBuf, listFileIOTableText
}
func (val *HostAllInfo) formatInnoDBBuffer() string {
	if val.dbErrorPerQuery.bufferPool.errMessage != "" || val.dbErrorPerQuery.bufferPool.errStatus == DBERROR {
		return fmt.Sprintf("%15s %15s %15s %15s %18s %18s %18s %16s %16s %16s %16s %15s %15s %15s %15s %17s",
			Error, Error, Error, Error, Error, Error, Error, Error, Error, Error, Error, Error, Error, Error, Error, Error)
	}
	return fmt.Sprintf("%15s %15s %15s %15s %18s %18s %18s %16s %16s %16s %16s %15s %15s %15s %15s %18s",
		val.Metric[BUFPOOLTOTAL],
		val.Metric[BUFPOOLDATA],
		val.Metric[BUFPOOLDIRTY],
		val.Metric[BUFPOOLFREE],
		val.Metric[BUFDATAP],
		val.Metric[BUFDIRTYP],
		val.Metric[BUFFREEP],
		val.Metric[BUFMISCP],
		val.Metric[BUFREAD],
		val.Metric[BUFREADREQ],
		val.Metric[BUFWRITEREQ],
		val.Metric[BUFWRITTEN],
		val.Metric[BUFCREATED],
		val.Metric[BUFFLUSH],
		val.Metric[LSNMINUSCKPOINT],
		fmt.Sprintf("%v/100", val.Metric[BUFFHITRATE]))

}

func (val *HostAllInfo) formatPerf() string {
	if val.dbErrorPerQuery.perf.errMessage != "" || val.dbErrorPerQuery.perf.errStatus == DBERROR {
		return fmt.Sprintf(" %14s %16s %16s %13s ", Error, Error, Error, Error)
	}
	return fmt.Sprintf(" %14s %16s %16s %13s ", val.Metric[SQLCNT], val.Metric[AVGLATENCY], val.Metric[MAXLATENCY], val.Metric[PSDIGESTLOST])
}
func (val *HostAllInfo) formatOSMetric() string {
	printval := " %12s %12s %12s %12s %12s %13s %13s %13s %13s %13s %13s %13s %13s %13s %13s %13s %13s %13s"
	if val.promeerror == 0 && val.promPort != 0 {
		return fmt.Sprintf(printval,
			val.MetricOS[CPUUSER],
			val.MetricOS[CPUSYS],
			val.MetricOS[CPUIO],
			val.MetricOS[CPUIRQ],
			val.MetricOS[CPUIDLE],
			val.MetricOS[DISKREAD],
			val.MetricOS[DISKWRITE],
			val.MetricOS[INBOUND],
			val.MetricOS[OUTBOUND],
			val.MetricOS[LOADAVG1],
			val.MetricOS[LOADAVG5],
			val.MetricOS[LOADAVG15],
			val.MetricOS[MEMUSED],
			val.MetricOS[MEMBUF],
			val.MetricOS[MEMCACHE],
			val.MetricOS[MEMFREE],
			val.MetricOS[SWAPUSED],
			val.MetricOS[SWAPFRE])
	}
	return fmt.Sprintf(" %12s", Error)

}
func (val *HostAllInfo) metricErr() {
	// dberror
	if val.dberror == DBERROR {
		for key, _ := range val.Metric {
			val.Metric[key] = Error
		}
	}
}

func (val *HostAllInfo) screenMetric(screenFlg int, idx int) string {
	metricColoring(val.Metric)
	metricColoring(val.MetricOS)
	val.metricErr()
	myNormal, myHdlr, mySlv, myPerm, myDigest, myThreads, myInnodbLock, myInnodbBuffer, myFileIOTable := val.formatMySQLMetric(idx)
	os := val.formatOSMetric()
	switch screenFlg {
	case MODENORMAL:
		return myNormal
	case MODERESOURCE:
		return myNormal + os
	case MODEPERFORMACE:
		return myPerm + myDigest
	case MODETHREADS:
		return myThreads
	case MODESLAVE:
		return mySlv
	case MODEROW:
		return myNormal + myHdlr
	case MODEINNOLOCK:
		return myInnodbLock
	case MODEINNOBUFFER:
		return myInnodbBuffer
	case MODEFILEIOTABLE:
		return myFileIOTable
	}
	return myNormal
}

func limitString(v string, maxlen int) string {
	if maxlen == 0 {
		return v
	}
	if len(v) > maxlen {
		v = v[:maxlen-1] + "."
	}
	return v
}
func (val *DigestMetric) formatTopN(idx int) string {
	length := HostFormatSize + len(INSTANCECURSOR) + PerfFormatSize
	cursor := INSTANCECURSOR
	newline := ""
	statementText := val.DigestText
	schemaName := val.SchemaName
	full := val.FullScan
	aLty := val.LatencyAvg
	rEa := val.RowExmAvg
	if val.CntStar == 0 {
		schemaName = ""
		statementText = ""
		full = ""
		aLty = ""
		rEa = ""
	}
	schemaName = limitString(schemaName, 8)
	if idx != 0 {
		cursor = padS(HostnameSize, " ") + padS(DEFAULTPORTSIZE, " ") + DIGESTFORMAT0
		newline = "\n"
		length = 0
	}
	return getStringUntilWidth(fmt.Sprintf("%s%s %8s %5s %7s %9s %6d %-s",
		newline,
		cursor,
		schemaName,
		full,
		rEa,
		formatTime(aLty),
		val.CntStar,
		statementText,
	), length)
}
func (val *FileIOperTable) formatTopN(idx int) string {
	cursor := INSTANCECURSOR
	newline := ""
	tableSchema := limitString(val.TableSchema, 10)
	tableName := limitString(val.TableName, 20)
	totalLatency := formatTime(fmt.Sprintf("%d", val.TotalLatency))
	fetchLatency := formatTime(fmt.Sprintf("%d", val.FetchLatency))
	insertLatency := formatTime(fmt.Sprintf("%d", val.InsertLatency))
	updateLatency := formatTime(fmt.Sprintf("%d", val.UpdateLatency))
	deleteLatency := formatTime(fmt.Sprintf("%d", val.DeleteLatency))
	readIOByte := humanRead(val.ReadIOByte)
	readIOLatency := formatTime(fmt.Sprintf("%d", val.ReadIOLatency))
	writeIOByte := humanRead(val.WriteIOByte)
	writeIOLatency := formatTime(fmt.Sprintf("%d", val.WriteIOLatency))
	miscIOLatency := formatTime(fmt.Sprintf("%d", val.MiscIOLatency))

	if idx != 0 {
		cursor = padS(HostnameSize, " ") + padS(DEFAULTPORTSIZE, " ") + FILEIOTABFORMAT0
		newline = "\n"
	}
	return fmt.Sprintf("%s%s %10s %20s %10d %7s %8d %9s %7d %7s %7d %7s %7d %7s %7d %8s %8s %8d %9s %9s %7d %8s",
		newline,
		cursor,
		tableSchema,
		tableName,
		val.CountStar,
		totalLatency,
		val.FetchRows,
		fetchLatency,
		val.InsertRows,
		insertLatency,
		val.UpdateRows,
		updateLatency,
		val.DeleteRows,
		deleteLatency,
		val.ReadIORequest,
		readIOByte,
		readIOLatency,
		val.WriteIORequest,
		writeIOByte,
		writeIOLatency,
		val.MiscIORequest,
		miscIOLatency,
	)
}
func (val *InnodbLockInfo) formatTopN(idx int) string {
	length := HostFormatSize + len(INSTANCECURSOR)
	cursor := INSTANCECURSOR
	newline := ""
	waitPidlist := limitString(val.WaitPidlist, 50)
	waitCnt := val.WaitCnt
	blockPid := val.BlockPid
	waitAge := val.WaitAge
	lockedTable := limitString(val.LockedTable, 30)
	lockedIndex := limitString(val.LockedIndex, 30)

	if idx != 0 {
		cursor = padS(HostnameSize, " ") + padS(DEFAULTPORTSIZE, " ") + INNOLOCKFORMAT0
		newline = "\n"
		length = 0
	}
	return getStringUntilWidth(fmt.Sprintf("%s%s %10s %30s %30s %4s %4s %-s",
		newline,
		cursor,
		blockPid,
		lockedTable,
		lockedIndex,
		waitAge,
		waitCnt,
		waitPidlist,
	), length)
}

func (val *ThreadsInfo) formatTopN(idx int) string {
	length := HostFormatSize + len(INSTANCECURSOR)
	cursor := INSTANCECURSOR
	newline := ""
	user := limitString(val.User, 7)
	host := val.Host
	db := limitString(val.Db, 7)
	time := val.Time
	state := limitString(val.State, 20)
	vsql := removeNewline(val.Sql)
	cmd := limitString(val.Cmd, 7)

	if idx != 0 {
		cursor = padS(HostnameSize, " ") + padS(DEFAULTPORTSIZE, " ") + THREADFORMAT0
		newline = "\n"
		length = 0
	}

	return getStringUntilWidth(fmt.Sprintf("%s%s %-7s %-20s %-7s %15s %7s %5s %-s",
		newline,
		cursor,
		cmd,
		state,
		user,
		host,
		db,
		time,
		vsql,
	), length)
}

func getStringUntilWidth(str string, length int) string {
	if FullSQLStatement == true {
		return str
	}
	width := tm.Width() - length
	if width < 0 {
		return str
	}
	// "\n" count 1
	if length == 0 {
		width++
	}
	if len(str) > width {
		str = limitString(str, width)
	}
	return str
}
func (val *HostAllInfo) formatSlaveInfoWithColoring() (string, string, string) {
	red := color.New(color.FgRed).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()
	sInfo := formatErr(val.dbErrorPerQuery.slave, val.dberror, false)
	normalIoTh := white(" ")
	normalSQLTh := white(" ")
	if sInfo != "" {
		return sInfo, normalIoTh, normalSQLTh
	}
	if len(val.SlaveInfo) > 0 {
		normalIoTh = white("Yes")
		normalSQLTh = white("Yes")
	}
	for _, SlaveInfo := range val.SlaveInfo {
		iothread := ""
		sqlthread := ""
		secbhdmaster := ""
		if strings.Contains(SlaveInfo.SlaveIO, "Yes") {
			iothread = white(SlaveInfo.SlaveIO)
		} else {
			iothread = red(limitString(SlaveInfo.SlaveIO, 3))
			normalIoTh = red(limitString(SlaveInfo.SlaveIO, 3))
		}
		if strings.Contains(SlaveInfo.SlaveSQL, "Yes") {
			sqlthread = white(SlaveInfo.SlaveSQL)
		} else {
			sqlthread = red(limitString(SlaveInfo.SlaveSQL, 3))
			normalSQLTh = red(limitString(SlaveInfo.SlaveSQL, 3))
		}
		i, _ := strconv.Atoi(SlaveInfo.SecBehdMas)
		if i != 0 {
			secbhdmaster = red(SlaveInfo.SecBehdMas)
		} else {
			secbhdmaster = white(SlaveInfo.SecBehdMas)
		}
		if GtidMode {
			retvgtid := removeNewline(SlaveInfo.RetvGTID)
			exegtid := removeNewline(SlaveInfo.ExeGTID)
			if sInfo == "" {
				sInfo += fmt.Sprintf("%15s %12s %12s %15s %14s %14s %14s %-100s\n %s %-100s",
					limitString(SlaveInfo.Host, 15),
					iothread,
					sqlthread,
					secbhdmaster,
					val.Metric[RPLSEMIMAS],
					val.Metric[RPLSEMISLV],
					limitString(SlaveInfo.ChlNm, 14),
					retvgtid,
					padS(HostnameSize, " ")+padS(DEFAULTPORTSIZE, " ")+padS(58, " "),
					exegtid)
			} else {
				sInfo += fmt.Sprintf("\n%25s %5s %15s %12s %12s %15s %14s %14s %14s %-100s\n %s %-100s",
					" ", " ",
					limitString(SlaveInfo.Host, 15),
					iothread,
					sqlthread,
					secbhdmaster,
					val.Metric[RPLSEMIMAS],
					val.Metric[RPLSEMISLV],
					limitString(SlaveInfo.ChlNm, 14),
					retvgtid,
					padS(HostnameSize, " ")+padS(DEFAULTPORTSIZE, " ")+padS(58, " "),
					exegtid)
			}
		} else {
			if sInfo == "" {
				sInfo += fmt.Sprintf("%15s %12s %12s %17s %10s %15s %10s %15s %14s %14s %-17s",
					limitString(SlaveInfo.Host, 15),
					iothread,
					sqlthread,
					SlaveInfo.MsLogFile,
					SlaveInfo.RMasLogPos,
					SlaveInfo.RelMasLogFile,
					SlaveInfo.ExeMasLogPos,
					secbhdmaster,
					val.Metric[RPLSEMIMAS],
					val.Metric[RPLSEMISLV],
					SlaveInfo.ChlNm)
			} else {
				sInfo += fmt.Sprintf("\n%25s %5s %15s %12s %12s %17s %10s %15s %10s %15s %14s %14s %-17s",
					" ", " ",
					limitString(SlaveInfo.Host, 15),
					iothread,
					sqlthread,
					SlaveInfo.MsLogFile,
					SlaveInfo.RMasLogPos,
					SlaveInfo.RelMasLogFile,
					SlaveInfo.ExeMasLogPos,
					secbhdmaster,
					val.Metric[RPLSEMIMAS],
					val.Metric[RPLSEMISLV],
					SlaveInfo.ChlNm)
			}
		}
	}
	return sInfo, normalIoTh, normalSQLTh
}

func getMetricForFile(now int, old int, title *string, content *string, nowTime string, outputTitle int) string {
	var out string
	if now == old && outputTitle != 1 {
		out = fmt.Sprintf("%s\n%s", nowTime, *content)
	} else {
		out = *title + *content
	}
	out = strings.Replace(out, BLUE, "", -1)
	out = strings.Replace(out, RED, "", -1)
	out = strings.Replace(out, GREEN, "", -1)
	out = strings.Replace(out, YELLOW, "", -1)
	out = strings.Replace(out, MAZENDA, "", -1)
	out = strings.Replace(out, WHITE, "", -1)
	out = strings.Replace(out, CYAN, "", -1)
	out = strings.Replace(out, COLEND, "", -1)
	return out
}
func flush(title *string, out *string) {
	var buf string
	//tm.Clear()
	callClear()
	tm.MoveCursor(1, 1)
	// teminal.go overflow pages of Flush() does not work well
	for idx, str := range strings.SplitAfter(*title+*out, "\n") {
		if idx >= tm.Height()-2 {
			break
		}
		buf += str
	}

	//tm.Println(buf)
	fmt.Fprint(color.Output, buf)
	tm.Output = bufio.NewWriter(color.Output)
	tm.Flush()
}

func callClear() {
	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                          //if we defined a clear func for that platform:
		value() //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clear terminal screen :(")
	}
}
func fileIOTableHeaderColoring() string {
	magenta := color.New(color.FgMagenta).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	str := blue("  Schema")
	str += blue("        TableName       ")
	if SortTableFileIOPosition == SORTNUM1 {
		str += magenta("AllRequest ")
	} else {
		str += blue("AllRequest ")
	}
	str += blue("AllLtcy ")
	if SortTableFileIOPosition == SORTNUM2 {
		str += magenta("FetchRow ")
	} else {
		str += blue("FetchRow ")
	}
	str += blue("FetchLtcy ")
	if SortTableFileIOPosition == SORTNUM3 {
		str += magenta("InsRows ")
	} else {
		str += blue("InsRows ")
	}
	str += blue("InsLtcy ")
	if SortTableFileIOPosition == SORTNUM4 {
		str += magenta("UpdRows ")
	} else {
		str += blue("UpdRows ")
	}
	str += blue("UpdLtcy ")
	if SortTableFileIOPosition == SORTNUM5 {
		str += magenta("DelRows ")
	} else {
		str += blue("DelRows ")
	}
	str += blue("DelLtcy ")
	if SortTableFileIOPosition == SORTNUM6 {
		str += magenta("ReadReq ")
	} else {
		str += blue("ReadReq ")
	}
	str += blue("ReadByte ReadLtcy ")
	if SortTableFileIOPosition == SORTNUM7 {
		str += magenta("WriteReq ")
	} else {
		str += blue("WriteReq ")
	}
	str += blue("WriteByte WriteLtcy MiscReq MiscLtcy")

	return str

}

func removeNewline(str string) string {
	buf := ""
	for _, v := range strings.Split(str, "\n") {
		buf += strings.TrimSpace(v) + " "
	}
	return buf
}
func padS(no int, char string) string {
	rtn := ""
	for i := 0; i < no; i++ {
		rtn += char
	}
	return rtn
}
func hostPortFormat(no int) string {
	str := "Hostname"
	if no > DEFAULTHOSTNAMESIZE {
		for i := 0; i < no-DEFAULTHOSTNAMESIZE; i++ {
			str = str + " "
		}
	}
	str = str + "  Port"
	return str
}

func modeHeaderColoring(screenFlg int) string {
	white := color.New(color.FgWhite).SprintFunc()
	magenta := color.New(color.FgMagenta).SprintFunc()
	buf := "  Mode : "
	if screenFlg == MODEOFF || screenFlg == MODENORMAL {
		buf += magenta("F1.Normal ")
	} else {
		buf += white("F1.Normal ")
	}
	if screenFlg == MODERESOURCE {
		buf += magenta("F2.OS Metric ")
	} else {
		buf += white("F2.OS Metric ")
	}
	if screenFlg == MODETHREADS {
		buf += magenta("F3.Threads ")
	} else {
		buf += white("F3.Threads ")
	}
	if screenFlg == MODEPERFORMACE {
		buf += magenta("F4.P_S Info ")
	} else {
		buf += white("F4.P_S Info ")
	}
	if screenFlg == MODESLAVE {
		buf += magenta("F5.SlaveStatus ")
	} else {
		buf += white("F5.SlaveStatus ")
	}
	if screenFlg == MODEROW {
		buf += magenta("F6.Handler/InnoDB_Rows ")
	} else {
		buf += white("F6.Handler/InnoDB_Rows ")
	}
	if screenFlg == MODEINNOLOCK {
		buf += magenta("F7.InnoDB Lock Info ")
	} else {
		buf += white("F7.InnoDB Lock Info ")
	}
	if screenFlg == MODEINNOBUFFER {
		buf += magenta("F8.InnoDB Buffer Info ")
	} else {
		buf += white("F8.InnoDB Buffer Info ")
	}
	if screenFlg == MODEFILEIOTABLE {
		buf += magenta("F9.Table IO Statistic ")
	} else {
		buf += white("F9.Table IO Statistic ")
	}
	buf += "(LEFT[KeyArrowLeft]/RIGHT[KeyArrowRight])\n"
	return buf

}
func modeTitleFormat(screenFlg int) string {
	var drawTitle string
	drawTitle += modeHeaderColoring(screenFlg)
	switch screenFlg {
	case MODEOFF, MODENORMAL:
		drawTitle += fmt.Sprintf("%s %s\n", HostNamePortTag, MyFormat1)
		drawTitle += fmt.Sprintf("%s %s\n", HostNamePort, MyFormat2)
		drawTitle += fmt.Sprintf("%s %s \n", HostNamePortTag, MyFormat3)
	case MODERESOURCE:
		drawTitle += fmt.Sprintf("%s %s %s\n", HostNamePortTag, MyFormat1, OSFormat1)
		drawTitle += fmt.Sprintf("%s %s %s\n", HostNamePort, MyFormat2, OSFormat2)
		drawTitle += fmt.Sprintf("%s %s %s\n", HostNamePortTag, MyFormat3, OSFormat3)
	case MODETHREADS:
		drawTitle += fmt.Sprintf("%s\n", KeyArrowUpDownFormat)
		drawTitle += fmt.Sprintf("%s %s %s\n", HostNamePortTag, Cursol1, ThreadFormat1)
		drawTitle += fmt.Sprintf("%s %s %s\n", HostNamePort, Cursol2, ThreadFormat2)
		drawTitle += fmt.Sprintf("%s %s %s\n", HostNamePortTag, Cursol3, ThreadFormat3)
	case MODEPERFORMACE:
		drawTitle += fmt.Sprintf("%s\n", KeyArrowUpDownFormat)
		drawTitle += fmt.Sprintf("%s %s %s %s\n", HostNamePortTag, PerfScmFormat1, Cursol1, DigestFormat1)
		drawTitle += fmt.Sprintf("%s %s %s %s\n", HostNamePort, PerfScmFormat2, Cursol2, DigestFormat2)
		drawTitle += fmt.Sprintf("%s %s %s %s\n", HostNamePortTag, PerfScmFormat3, Cursol3, DigestFormat3)
	case MODESLAVE:
		sFormat1 := MySlaFormat1
		sFormat2 := MySlaFormat2
		sFormat3 := MySlaFormat3
		if GtidMode {
			sFormat1 = MySlaFormatGTID1
			sFormat2 = MySlaFormatGTID2
			sFormat3 = MySlaFormatGTID3
		}
		drawTitle += fmt.Sprintf("%s %s \n", HostNamePortTag, sFormat1)
		drawTitle += fmt.Sprintf("%s %s \n", HostNamePort, sFormat2)
		drawTitle += fmt.Sprintf("%s %s \n", HostNamePortTag, sFormat3)
	case MODEROW:
		drawTitle += fmt.Sprintf("%s %s %s\n", HostNamePortTag, MyFormat1, MyHadrFormat1)
		drawTitle += fmt.Sprintf("%s %s %s\n", HostNamePort, MyFormat2, MyHadrFormat2)
		drawTitle += fmt.Sprintf("%s %s %s\n", HostNamePortTag, MyFormat3, MyHadrFormat3)
	case MODEINNOLOCK:
		drawTitle += fmt.Sprintf("%s\n", KeyArrowUpDownFormat)
		drawTitle += fmt.Sprintf("%s %s %s\n", HostNamePortTag, Cursol1, InnoLockFormat1)
		drawTitle += fmt.Sprintf("%s %s %s\n", HostNamePort, Cursol2, InnoLockFormat2)
		drawTitle += fmt.Sprintf("%s %s %s\n", HostNamePortTag, Cursol3, InnoLockFormat3)
	case MODEINNOBUFFER:
		drawTitle += fmt.Sprintf("%s %s\n", HostNamePortTag, InnoBufFormat1)
		drawTitle += fmt.Sprintf("%s %s\n", HostNamePort, InnoBufFormat2)
		drawTitle += fmt.Sprintf("%s %s\n", HostNamePortTag, InnoBufFormat3)
	case MODEFILEIOTABLE:
		drawTitle += fmt.Sprintf("%s\n", FileIOTabFormat4)
		drawTitle += fmt.Sprintf("%s %s %s\n", HostNamePortTag, Cursol1, FileIOTabFormat1)
		drawTitle += fmt.Sprintf("%s %s %s\n", HostNamePort, Cursol2, fileIOTableHeaderColoring())
		drawTitle += fmt.Sprintf("%s %s %s\n", HostNamePortTag, Cursol3, FileIOTabFormat3)
	}
	return drawTitle
}
func headerColoring() {
	white := color.New(color.FgWhite).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	Error = red("-1")
	TopFormat = white(" now() : %s  [Start : %s] [Esc or SpaceBar.Exit] \n")
	FinishFormat = white(" [Start : %s] [AutoStop : %s ]\n")
	OSFormat1 = blue("------- CPU ------- -- Disk - -- NW --- ---load-avg--- ------ memory ----  - swap -- ")
	OSFormat2 = blue("usr sys wai irq idl read writ recv send  1m   5m   15m used buff cach free swap free ")
	OSFormat3 = blue("--- --- --- --- --- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ")
	MyFormat1 = blue("------------------------------------------- QPS ------------------------------------------------------------------------")
	MyFormat2 = blue("Conn Run Abort Select Update Insert Delete Replace Qcache Call QPSAll StmtExc Commit Rolbk Slow Ssum Read  Delay  IO SQL")
	MyFormat3 = blue("---- --- ----- ------ ------ ------ ------ ------- ------ ---- ------ ------- ------ ----- ---- ---- ---- ------ --- ---")
	MySlaFormat1 = blue("---------------------------------------- SLAVE STATUS -------------------------------------------------------------")
	MySlaFormat2 = blue("   MasterHost    IO SQL     Master_Log     MasterPos     Relay_M_Log    Exec_Pos   Delay SemiM SemiS  ChannelName ")
	MySlaFormat3 = blue("--------------- --- --- ----------------- ---------- ----------------- ---------- ------ ----- ----- --------------")
	MySlaFormatGTID1 = blue("---------------------------------------- SLAVE STATUS -------------------------------------------------------------")
	MySlaFormatGTID2 = blue("   MasterHost    IO SQL  Delay SemiM SemiS   ChannelName          Retrieved_Gtid_Set/Executed_Gtid_Set             ")
	MySlaFormatGTID3 = blue("--------------- --- --- ------ ----- ----- -------------- ---------------------------------------------------------")
	MyHadrFormat1 = blue("------------ InnoDB Rows ---------- ---------------------------- Handlar -------------------------")
	MyHadrFormat2 = blue("  read   inserted  updated  deleted  RdFirst   RdKey   RdLast    RdNxt   RdPrev    RdRnd  RdRndNxt")
	MyHadrFormat3 = blue("-------- -------- -------- -------- -------- -------- -------- -------- -------- -------- --------")
	KeyArrowUpDownFormat = white(fmt.Sprintf("  Display per Servers (Select Up[KeyArrowUp] or Down[KeyArrowDown] ) , Display Full SQL by pushing F12 [%v]", FullSQLStatement))
	perfBuf := "---- performace_schema ----"
	// Space +1
	PerfFormatSize = len(perfBuf) + 1
	PerfScmFormat1 = blue(perfBuf)
	PerfScmFormat2 = blue("  SQL    Avg     Max   Lost")
	PerfScmFormat3 = blue("------ ------- ------- ----")
	Cursol1 = blue("-")
	Cursol2 = blue("*")
	Cursol3 = blue("-")
	DigestFormat1 = blue("--------------------- Statement Digest -----------------------------------------------------------------------------------------------------")
	DigestFormat2 = blue(" Schema   Ful   AvgEm   AvgLtcy   ExCnt   Query")
	DigestFormat3 = blue("-------- ----- ------- --------- ------- ---------------------------------------------------------------------------------------------------")
	ThreadFormat1 = blue("------------------------------- Thread Info ----------------------------------------------------------------------------------------------------------")
	ThreadFormat2 = blue(" Cmd            State          User        Host         DB    Time   Query                                                                            ")
	ThreadFormat3 = blue("------- -------------------- ------- --------------- ------- ----- -----------------------------------------------------------------------------------")
	InnoLockFormat1 = blue("-------------------------- InnoDB Lock Info -----------------------------------------------------------------------------------------------------")
	InnoLockFormat2 = blue(" BlockPid           Locked Table                  Locked Index           time  Cnt     Waiting Pid                                               ")
	InnoLockFormat3 = blue("---------- ------------------------------ ------------------------------ ---- ---- --------------------------------------------------------------")
	InnoBufFormat1 = blue("-------------------------------------------------- InnoDB Buffer Pool Info ----------------------------------------------------")
	InnoBufFormat2 = blue(" Total  Data   Dirty  Free   DataPage DirtyPage  FreePage MiscPag   Read   RdReq  WritReq Writen Create  Flush LSN-CK  HitRate ")
	InnoBufFormat3 = blue("------ ------ ------ ------ --------- --------- --------- ------- ------- ------- ------- ------ ------ ------ ------ ---------")
	FileIOTabFormat1 = blue("---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------")
	FileIOTabFormat3 = blue("---------- -------------------- ---------- ------- -------- --------- ------- ------- ------- ------- ------- ------- ------- -------- -------- -------- --------- --------- ------- --------")
	FileIOTabFormat4 = white("     (Choose Sort Field by [Tab] key)")
	HostNamePortTag = blue(padS(HostnameSize, "-") + " " + padS(DEFAULTPORTSIZE, "-"))
	HostNamePort = blue(hostPortFormat(HostnameSize))

}
func createHeader(screenFlg int) string {
	var buf string
	headerColoring()
	buf += fmt.Sprintf(TopFormat, time.Now().Format(TIMELAYOUT), StartTime.Format(TIMELAYOUT))
	buf += modeTitleFormat(screenFlg)
	return buf
}

func (val *HostAllInfo) createHostFormat() string {
	buf := fmt.Sprintf("%-"+fmt.Sprintf("%d", HostnameSize)+"s %"+fmt.Sprintf("%d", DEFAULTPORTSIZE)+"s", val.alias, val.port)
	HostFormatSize = len(buf)
	return buf
}
func screen(hostInfos []HostAllInfo, screenFlg int, screenHostlist []string, sElaspTime time.Time) (string, string) {
	var drawTitle string
	var drawAll string
	drawTitle += createHeader(screenFlg)
	// loop all hosts

	for i := 0; i < len(screenHostlist); i++ {
		// just show the comment lines
		if strings.Contains(screenHostlist[i], "#") {
			drawAll += fmt.Sprintf("%s\n", screenHostlist[i])
			continue
		}
		ii, _ := strconv.Atoi(screenHostlist[i])
		// if host find out, formating metric
		host := hostInfos[ii].createHostFormat()
		if screenFlg != MODEOFF {
			drawAll += fmt.Sprintf("%s %s\n", host, hostInfos[ii].screenMetric(screenFlg, ii))
		} else {
			drawAll += fmt.Sprintf("%s\n", host)
		}
	}

	eElaspTime := (time.Now()).Sub(sElaspTime)
	drawAll += fmt.Sprintf("\n\n>> Elapsed time : %2v\n", eElaspTime)
	flush(&drawTitle, &drawAll)
	return drawTitle, drawAll
}
