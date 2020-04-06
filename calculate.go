package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
)

func (val *HostAllInfo) calculate(screenFlg int, old *HostAllInfo) {
	// Calculate All Metrics
	metricCalc(val, old)

	// For JSON OutPut
	if OutputJSON {
		copyMap(val.Metric, val.MetricJSON)
	}
}

func metricCalc(val *HostAllInfo, old *HostAllInfo) {
	m.Lock()

	// Prometheus Metric Calc
	//CPU rate
	usr := val.PromeParam[CPUUSER] - old.PromeParam[CPUUSER]
	sys := val.PromeParam[CPUSYS] - old.PromeParam[CPUSYS]
	wait := val.PromeParam[CPUIO] - old.PromeParam[CPUIO]
	etc := val.PromeParam[CPUIRQ] - old.PromeParam[CPUIRQ]
	idle := val.PromeParam[CPUIDLE] - old.PromeParam[CPUIDLE]
	cpuTotal := usr + sys + wait + etc + idle
	var up, sp, wp, et int64
	val.MetricOS[CPUUSER], up = getPacentage(usr, cpuTotal)
	val.MetricOS[CPUSYS], sp = getPacentage(sys, cpuTotal)
	val.MetricOS[CPUIO], wp = getPacentage(wait, cpuTotal)
	val.MetricOS[CPUIRQ], et = getPacentage(etc, cpuTotal)
	val.MetricOS[CPUIDLE] = chkminus(100 - up - sp - wp - et)
	//DISK and NW and SWAP
	val.MetricOS[DISKREAD] = humanRead(uint64(val.PromeParam[DISKREAD]) - uint64(old.PromeParam[DISKREAD]))
	val.MetricOS[DISKWRITE] = humanRead(uint64(val.PromeParam[DISKWRITE]) - uint64(old.PromeParam[DISKWRITE]))
	val.MetricOS[INBOUND] = humanRead(uint64(val.PromeParam[INBOUND]) - uint64(old.PromeParam[INBOUND]))
	val.MetricOS[OUTBOUND] = humanRead(uint64(val.PromeParam[OUTBOUND]) - uint64(old.PromeParam[OUTBOUND]))
	val.MetricOS[MEMUSED] = humanRead(uint64(val.PromeParam[MEMTOL]) - uint64(val.PromeParam[MEMBUF]) - uint64(val.PromeParam[MEMCACHE]) - uint64(val.PromeParam[MEMFREE]))
	val.MetricOS[MEMBUF] = humanRead(uint64(val.PromeParam[MEMBUF]))
	val.MetricOS[MEMCACHE] = humanRead(uint64(val.PromeParam[MEMCACHE]))
	val.MetricOS[MEMFREE] = humanRead(uint64(val.PromeParam[MEMFREE]))
	val.MetricOS[MEMTOL] = humanRead(uint64(val.PromeParam[MEMTOL]))
	val.MetricOS[SWAPUSED] = humanRead(uint64(val.PromeParam[SWAPTOL] - val.PromeParam[SWAPFRE]))
	val.MetricOS[SWAPFRE] = humanRead(uint64(val.PromeParam[SWAPFRE]))
	val.MetricOS[LOADAVG1] = fmt.Sprintf("%.2f", val.PromeParam[LOADAVG1])
	val.MetricOS[LOADAVG5] = fmt.Sprintf("%.2f", val.PromeParam[LOADAVG5])
	val.MetricOS[LOADAVG15] = fmt.Sprintf("%.2f", val.PromeParam[LOADAVG15])
	// MySQL Metric Calc
	//// status
	val.Metric[THCON] = val.MyStatu[THCON]
	val.Metric[THRUN] = val.MyStatu[THRUN]
	val.Metric[ABORTCON] = chkminus(changeStr2Int(val.MyStatu[ABORTCON]) - changeStr2Int(old.MyStatu[ABORTCON]))
	csel := changeStr2Int(val.MyStatu[CSEL]) - changeStr2Int(old.MyStatu[CSEL])
	cupd := changeStr2Int(val.MyStatu[CUPD]) - changeStr2Int(old.MyStatu[CUPD]) + changeStr2Int(val.MyStatu[CUPDMUL]) - changeStr2Int(old.MyStatu[CUPDMUL])
	cins := changeStr2Int(val.MyStatu[CINS]) - changeStr2Int(old.MyStatu[CINS]) + changeStr2Int(val.MyStatu[CINSSEL]) - changeStr2Int(old.MyStatu[CINSSEL])
	cdel := changeStr2Int(val.MyStatu[CDEL]) - changeStr2Int(old.MyStatu[CDEL]) + changeStr2Int(val.MyStatu[CDELMUL]) - changeStr2Int(old.MyStatu[CDELMUL])
	crlce := changeStr2Int(val.MyStatu[CRLCE]) - changeStr2Int(old.MyStatu[CRLCE]) + changeStr2Int(val.MyStatu[CRLCESEL]) - changeStr2Int(old.MyStatu[CRLCESEL])
	cQcach := changeStr2Int(val.MyStatu[QCACHHIT]) - changeStr2Int(old.MyStatu[QCACHHIT])
	cPrce := changeStr2Int(val.MyStatu[CCALPROCEDURE]) - changeStr2Int(old.MyStatu[CCALPROCEDURE])
	cStmt := changeStr2Int(val.MyStatu[CSTMT]) - changeStr2Int(old.MyStatu[CSTMT])
	cCommit := changeStr2Int(val.MyStatu[CCOMMIT]) - changeStr2Int(old.MyStatu[CCOMMIT])
	cRollback := changeStr2Int(val.MyStatu[CROLLBACK]) - changeStr2Int(old.MyStatu[CROLLBACK])
	val.Metric[CSEL] = chkminus(csel)
	val.Metric[CUPD] = chkminus(cupd)
	val.Metric[CINS] = chkminus(cins)
	val.Metric[CDEL] = chkminus(cdel)
	val.Metric[CRLCE] = chkminus(crlce)
	val.Metric[QCACHHIT] = chkminus(cQcach)
	val.Metric[CCALPROCEDURE] = chkminus(cPrce)
	val.Metric[CSTMT] = chkminus(cStmt)
	val.Metric[QPSALL] = chkminus(csel + cupd + cins + cdel + crlce + cQcach + cPrce)
	val.Metric[CCOMMIT] = chkminus(cCommit)
	val.Metric[CROLLBACK] = chkminus(cRollback)
	digestlost := changeStr2Int(val.MyStatu[PSDIGESTLOST]) - changeStr2Int(old.MyStatu[PSDIGESTLOST])
	val.Metric[PSDIGESTLOST] = chkminus(digestlost)
	val.Metric[BUFWRITEREQ] = chkminus(changeStr2Int(val.MyStatu[BUFWRITEREQ]) - changeStr2Int(old.MyStatu[BUFWRITEREQ]))
	readRequests := changeStr2Int(val.MyStatu[BUFREADREQ]) - changeStr2Int(old.MyStatu[BUFREADREQ])
	reads := changeStr2Int(val.MyStatu[BUFREAD]) - changeStr2Int(old.MyStatu[BUFREAD])
	val.Metric[BUFREADREQ] = chkminus(readRequests)
	val.Metric[BUFREAD] = chkminus(reads)
	val.Metric[BUFFHITRATE] = calcHitRate(reads, readRequests)
	val.Metric[BUFCREATED] = chkminus(changeStr2Int(val.MyStatu[BUFCREATED]) - changeStr2Int(old.MyStatu[BUFCREATED]))
	val.Metric[BUFWRITTEN] = chkminus(changeStr2Int(val.MyStatu[BUFWRITTEN]) - changeStr2Int(old.MyStatu[BUFWRITTEN]))
	val.Metric[BUFFLUSH] = chkminus(changeStr2Int(val.MyStatu[BUFFLUSH]) - changeStr2Int(old.MyStatu[BUFFLUSH]))
	val.Metric[BUFDATAP] = val.MyStatu[BUFDATAP]
	val.Metric[BUFDIRTYP] = val.MyStatu[BUFDIRTYP]
	val.Metric[BUFFREEP] = val.MyStatu[BUFFREEP]
	val.Metric[BUFMISCP] = chkminus(changeStr2Int(val.MyStatu[BUFTOTALP]) - changeStr2Int(val.MyStatu[BUFDATAP]) - changeStr2Int(val.MyStatu[BUFFREEP]))

	var slow int64
	if old.MyStatu[SLWLOG] != "" {
		slow = changeStr2Int(val.MyStatu[SLWLOG]) - changeStr2Int(old.MyStatu[SLWLOG])
	} else {
		slow = 0
	}
	val.Metric[SLWLOG] = chkminus(slow)
	val.slowSum = slow + old.slowSum
	val.Metric[SLWSUM] = chkminus(val.slowSum)
	val.Metric[HADRRDFIRST] = chkminus(changeStr2Int(val.MyStatu[HADRRDFIRST]) - changeStr2Int(old.MyStatu[HADRRDFIRST]))
	val.Metric[HADRRDKEY] = chkminus(changeStr2Int(val.MyStatu[HADRRDKEY]) - changeStr2Int(old.MyStatu[HADRRDKEY]))
	val.Metric[HADRRDLAST] = chkminus(changeStr2Int(val.MyStatu[HADRRDLAST]) - changeStr2Int(old.MyStatu[HADRRDLAST]))
	val.Metric[HADRRDNXT] = chkminus(changeStr2Int(val.MyStatu[HADRRDNXT]) - changeStr2Int(old.MyStatu[HADRRDNXT]))
	val.Metric[HADRRDPRV] = chkminus(changeStr2Int(val.MyStatu[HADRRDPRV]) - changeStr2Int(old.MyStatu[HADRRDPRV]))
	val.Metric[HADRRDRND] = chkminus(changeStr2Int(val.MyStatu[HADRRDRND]) - changeStr2Int(old.MyStatu[HADRRDRND]))
	val.Metric[HADRRDRNDNXT] = chkminus(changeStr2Int(val.MyStatu[HADRRDRNDNXT]) - changeStr2Int(old.MyStatu[HADRRDRNDNXT]))

	val.Metric[HADRDEL] = chkminus(changeStr2Int(val.MyStatu[HADRDEL]) - changeStr2Int(old.MyStatu[HADRDEL]))
	val.Metric[HADRUPD] = chkminus(changeStr2Int(val.MyStatu[HADRUPD]) - changeStr2Int(old.MyStatu[HADRUPD]))
	val.Metric[HADRWRT] = chkminus(changeStr2Int(val.MyStatu[HADRWRT]) - changeStr2Int(old.MyStatu[HADRWRT]))
	val.Metric[HADRPREP] = chkminus(changeStr2Int(val.MyStatu[HADRPREP]) - changeStr2Int(old.MyStatu[HADRPREP]))
	val.Metric[HADRCOMT] = chkminus(changeStr2Int(val.MyStatu[HADRCOMT]) - changeStr2Int(old.MyStatu[HADRCOMT]))
	val.Metric[HADRRB] = chkminus(changeStr2Int(val.MyStatu[HADRRB]) - changeStr2Int(old.MyStatu[HADRRB]))

	val.Metric[INNOROWDEL] = chkminus(changeStr2Int(val.MyStatu[INNOROWDEL]) - changeStr2Int(old.MyStatu[INNOROWDEL]))
	val.Metric[INNOROWRD] = chkminus(changeStr2Int(val.MyStatu[INNOROWRD]) - changeStr2Int(old.MyStatu[INNOROWRD]))
	val.Metric[INNOROWINS] = chkminus(changeStr2Int(val.MyStatu[INNOROWINS]) - changeStr2Int(old.MyStatu[INNOROWINS]))
	val.Metric[INNOROWUPD] = chkminus(changeStr2Int(val.MyStatu[INNOROWUPD]) - changeStr2Int(old.MyStatu[INNOROWUPD]))

	// performanceschemas
	val.Metric[SQLCNT] = val.MyPerfomanceSchema[SQLCNT]
	val.Metric[AVGLATENCY] = formatTime(val.MyPerfomanceSchema[AVGLATENCY])
	val.Metric[MAXLATENCY] = formatTime(val.MyPerfomanceSchema[MAXLATENCY])

	val.Metric[SECBEHMAS] = val.MySlave[SECBEHMAS]
	val.Metric[RDONLY] = val.MyVariable[RDONLY]
	val.Metric[RPLSEMIMAS] = val.MyVariable[RPLSEMIMAS]
	val.Metric[RPLSEMISLV] = val.MyVariable[RPLSEMISLV]

	//InnoDB Metric
	val.Metric[BUFPOOLTOTAL] = humanRead(uint64(val.MyInnoDBBuffer[BUFPOOLTOTAL]))
	val.Metric[BUFPOOLDATA] = humanRead(uint64(val.MyInnoDBBuffer[BUFPOOLDATA]))
	val.Metric[BUFPOOLDIRTY] = humanRead(uint64(val.MyInnoDBBuffer[BUFPOOLDIRTY]))
	val.Metric[BUFPOOLFREE] = humanRead(uint64(val.MyInnoDBBuffer[BUFPOOLFREE]))
	if val.MyInnoDBBuffer[LSNMINUSCKPOINT] == -1 {
		val.Metric[LSNMINUSCKPOINT] = "-1"
	} else {
		val.Metric[LSNMINUSCKPOINT] = humanRead(uint64(val.MyInnoDBBuffer[LSNMINUSCKPOINT]))
	}

	m.Unlock()
	return

}

func calcFileIODelta(new *HostAllInfo, old *HostAllInfo) List {
	ss := List{}
	for tableNm, vStrFileIOperTable := range new.strFileIOperTable {
		if old.strFileIOperTable[tableNm] == nil {
			old.strFileIOperTable[tableNm] = &FileIOperTable{TableSchema: vStrFileIOperTable.TableSchema, TableName: vStrFileIOperTable.TableName}
		} else {
			switch SortTableFileIOPosition {
			case SORTNUM1:
				e := Entry{tableNm, vStrFileIOperTable.CountStar - old.strFileIOperTable[tableNm].CountStar}
				ss = append(ss, e)
			case SORTNUM2:
				e := Entry{tableNm, vStrFileIOperTable.FetchRows - old.strFileIOperTable[tableNm].FetchRows}
				ss = append(ss, e)
			case SORTNUM3:
				e := Entry{tableNm, vStrFileIOperTable.InsertRows - old.strFileIOperTable[tableNm].InsertRows}
				ss = append(ss, e)
			case SORTNUM4:
				e := Entry{tableNm, vStrFileIOperTable.UpdateRows - old.strFileIOperTable[tableNm].UpdateRows}
				ss = append(ss, e)
			case SORTNUM5:
				e := Entry{tableNm, vStrFileIOperTable.DeleteRows - old.strFileIOperTable[tableNm].DeleteRows}
				ss = append(ss, e)
			case SORTNUM6:
				e := Entry{tableNm, vStrFileIOperTable.ReadIOByte - old.strFileIOperTable[tableNm].ReadIOByte}
				ss = append(ss, e)
			case SORTNUM7:
				e := Entry{tableNm, vStrFileIOperTable.WriteIOByte - old.strFileIOperTable[tableNm].WriteIOByte}
				ss = append(ss, e)
			}
		}
	}
	return ss
}

func sortFileIO(new *HostAllInfo, old *HostAllInfo) {
	//Digest
	ss := calcFileIODelta(new, old)
	sort.Sort(ss)

	//new.FileIOMetric = make([]FileIOperTable, TopN)
	for idx, kv := range ss {
		if idx >= TopN {
			break
		}
		contStar := new.strFileIOperTable[kv.sKey].CountStar - old.strFileIOperTable[kv.sKey].CountStar
		if contStar == 0 {
			continue
		}
		m.Lock()
		new.FileIOMetric[idx] = FileIOperTable{
			TableSchema:    new.strFileIOperTable[kv.sKey].TableSchema,
			TableName:      new.strFileIOperTable[kv.sKey].TableName,
			CountStar:      contStar,
			TotalLatency:   new.strFileIOperTable[kv.sKey].TotalLatency - old.strFileIOperTable[kv.sKey].TotalLatency,
			FetchRows:      new.strFileIOperTable[kv.sKey].FetchRows - old.strFileIOperTable[kv.sKey].FetchRows,
			FetchLatency:   new.strFileIOperTable[kv.sKey].FetchLatency - old.strFileIOperTable[kv.sKey].FetchLatency,
			InsertRows:     new.strFileIOperTable[kv.sKey].InsertRows - old.strFileIOperTable[kv.sKey].InsertRows,
			InsertLatency:  new.strFileIOperTable[kv.sKey].InsertLatency - old.strFileIOperTable[kv.sKey].InsertLatency,
			UpdateRows:     new.strFileIOperTable[kv.sKey].UpdateRows - old.strFileIOperTable[kv.sKey].UpdateRows,
			UpdateLatency:  new.strFileIOperTable[kv.sKey].UpdateLatency - old.strFileIOperTable[kv.sKey].UpdateLatency,
			DeleteRows:     new.strFileIOperTable[kv.sKey].DeleteRows - old.strFileIOperTable[kv.sKey].DeleteRows,
			DeleteLatency:  new.strFileIOperTable[kv.sKey].DeleteLatency - old.strFileIOperTable[kv.sKey].DeleteLatency,
			ReadIORequest:  new.strFileIOperTable[kv.sKey].ReadIORequest - old.strFileIOperTable[kv.sKey].ReadIORequest,
			ReadIOByte:     new.strFileIOperTable[kv.sKey].ReadIOByte - old.strFileIOperTable[kv.sKey].ReadIOByte,
			ReadIOLatency:  new.strFileIOperTable[kv.sKey].ReadIOLatency - old.strFileIOperTable[kv.sKey].ReadIOLatency,
			WriteIORequest: new.strFileIOperTable[kv.sKey].WriteIORequest - old.strFileIOperTable[kv.sKey].WriteIORequest,
			WriteIOByte:    new.strFileIOperTable[kv.sKey].WriteIOByte - old.strFileIOperTable[kv.sKey].WriteIOByte,
			WriteIOLatency: new.strFileIOperTable[kv.sKey].WriteIOLatency - old.strFileIOperTable[kv.sKey].WriteIOLatency,
			MiscIORequest:  new.strFileIOperTable[kv.sKey].MiscIORequest - old.strFileIOperTable[kv.sKey].MiscIORequest,
			MiscIOLatency:  new.strFileIOperTable[kv.sKey].MiscIOLatency - old.strFileIOperTable[kv.sKey].MiscIOLatency,
		}
		m.Unlock()
	}
}

func (val *HostAllInfo) calcDigestDelta(base *HostAllInfo) (mapDigest, uint64) {
	var digestAllSQLCnt uint64
	digestAllSQLCnt = 0
	digestDelta := make(mapDigest, 0)
	for key, value := range val.Digest {
		digestDelta[key] = value - base.DigestBase[key]
		digestAllSQLCnt += digestDelta[key]
		base.DigestBase[key] = value
	}
	return digestDelta, digestAllSQLCnt
}

func (val *HostAllInfo) sortDigest(base *HostAllInfo) string {
	digestDelta, digestAllSQLCnt := val.calcDigestDelta(base)
	ss := List{}
	for schemaDigest, execnt := range digestDelta {
		e := Entry{schemaDigest, execnt}
		ss = append(ss, e)
	}
	sort.Sort(ss)
	topDigestQuery := ""
	len := len(ss)
	m.Lock()
	val.DigestALLSqlCnt = fmt.Sprintf("%d", digestAllSQLCnt)
	for idx, kv := range ss {
		val.DigestMetric[idx] = DigestMetric{Digest: kv.sKey, CntStar: kv.sCnt}
		if idx+1 < TopN && idx+1 < len {
			topDigestQuery += fmt.Sprintf("'%s',", kv.sKey)
		} else {
			topDigestQuery += fmt.Sprintf("'%s'", kv.sKey)
			break
		}
	}
	m.Unlock()
	return topDigestQuery
}

func (l List) Len() int {
	return len(l)
}

func (l List) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l List) Less(i, j int) bool {
	if l[i].sCnt == l[j].sCnt {
		return (l[i].sKey > l[j].sKey)
	}
	return (l[i].sCnt > l[j].sCnt)
}

func chkminus(i int64) string {
	if i < 0 {
		return "0"
	}
	return fmt.Sprintf("%d", i)
}

func changeStr2Int(v string) int64 {
	f, _ := strconv.ParseFloat(v, 64)
	i := int64(f)
	return i
}

func getPacentage(rate float64, total float64) (string, int64) {
	if rate == 0 && total == 0 {
		return "0", 0
	}
	return fmt.Sprintf("%.0f", rate/total*100), changeStr2Int(fmt.Sprintf("%.0f", rate/total*100))
}

func humanRead(v uint64) string {
	v1 := humanize.Bytes(v)
	v2 := strings.Replace(v1, " ", "", 1)
	v3 := strings.Replace(v2, "B", "", 1)
	return v3
}

func calcHitRate(reads int64, readRequests int64) string {
	if readRequests == 0 {
		return "100.0"
	}
	rate := fmt.Sprintf("%.2f", ((1.0 - (float64(reads) / float64(readRequests))) * 100.0))
	if rate == "100.00" {
		return "100.0"
	}
	return rate
}
func formatTime(v string) string {
	f, _ := strconv.ParseFloat(v, 64)
	var out string
	switch {
	case f >= 604800000000000000:
		out = fmt.Sprintf("%.1fw", math.Trunc(f/604800000000000000))
	case f >= 86400000000000000:
		out = fmt.Sprintf("%.1fd", math.Trunc(f/86400000000000000))
	case f >= 3600000000000000:
		out = fmt.Sprintf("%.1fh", math.Trunc(f/3600000000000000))
	case f >= 60000000000000:
		out = fmt.Sprintf("%.1fm", math.Trunc(f/60000000000000))
	case f >= 1000000000000:
		out = fmt.Sprintf("%.1fs", math.Trunc(f/1000000000000))
	case f >= 1000000000:
		out = fmt.Sprintf("%.1fms", math.Trunc(f/1000000000))
	case f >= 1000000:
		out = fmt.Sprintf("%.1fus", math.Trunc(f/1000000))
	case f >= 1000:
		out = fmt.Sprintf("%.1fns", math.Trunc(f/1000000))
	default:
		out = fmt.Sprintf("%.1fns", f)
	}
	return out
}
