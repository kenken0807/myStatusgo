package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
)

func (val *HostAllInfo) retrieve(screenFlg int, old *HostAllInfo, baseInfo *HostAllInfo) {
	val.dberror = DBOK
	if err := val.db.Ping(); err != nil {
		val.dberror = DBERROR
	}

	wg := &sync.WaitGroup{}
	wg.Add(8)
	// OS metric
	go func() {
		if val.promPort != 0 {
			val.getOsMetric()
		}
		wg.Done()
	}()

	// Thread Info
	go func() {
		val.doExecute(&val.dbErrorPerQuery.threads, 7)
		wg.Done()
	}()

	// Show Subordinate Status
	// Show Global Status
	// Show Variables
	go func() {
		val.doExecute(&val.dbErrorPerQuery.subordinate, 1)
		val.doExecute(&val.dbErrorPerQuery.status, 2)
		val.doExecute(&val.dbErrorPerQuery.variables, 3)
		wg.Done()
	}()

	// events_statements_current  Over MySQL5.6
	go func() {
		val.doExecute(&val.dbErrorPerQuery.perf, 4)
		wg.Done()
	}()

	// InnoDB Buffer Pool Info
	go func() {
		val.doExecute(&val.dbErrorPerQuery.bufferPool, 5)
		wg.Done()
	}()
	// InnoDB Row Lock Info Over MySQL5.6
	//// session can not kill if lots of locks
	go func() {
		if screenFlg == MODEINNOLOCK && baseInfo.topNPosition == TopNPosition {
			val.doExecute(&val.dbErrorPerQuery.lock, 6)
		}
		wg.Done()
	}()

	// events_statements_summary_by_digest Over MySQL5.6
	//// At First, retrieve events_statements_summary_by_digest limit 5000
	go func() {
		if screenFlg == MODEOFF || len(baseInfo.DigestBase) == 0 {
			if ok := baseInfo.doExecute(&val.dbErrorPerQuery.digest, 8); ok == DBERROR {
				initializeStructIfdbErrorDigest(old, baseInfo)
			}
		}
		//// events_statements_summary_by_digest WHERE LAST_SEEN > now() - 1
		if ok := val.doExecute(&val.dbErrorPerQuery.digest, 9); ok == DBERROR {
			initializeStructIfdbErrorDigest(old, baseInfo)
		}
		// Success
		topDigestQuery := val.sortDigest(baseInfo)
		if topDigestQuery != "" {
			s, _ := val.sqlExecDigestText(topDigestQuery)
			if s != DBOK {
				initializeStructIfdbErrorDigest(old, baseInfo)
			}
		}
		wg.Done()
	}()
	// Table IO Stastics Info
	go func() {
		if screenFlg == MODEFILEIOTABLE && baseInfo.topNPosition == TopNPosition {
			if ok := val.doExecute(&val.dbErrorPerQuery.fileIO, 10); ok == DBERROR {
				initializeStructIfdbErrorFileIO(old, baseInfo)
			} else {
				// Calculate FileIOTable
				sortFileIO(val, old)
			}
		}
		wg.Done()
	}()
	wg.Wait()
}

func (val *HostAllInfo) doExecute(dbError *dbError, id int) bool {
	// IF DB Error means down mysqld , Nothing to do
	if val.dberror == DBERROR {
		return DBERROR
	}
	// Error count >= wait , reset counter. do SQL again
	if dbError.count >= WaitCountDBError {
		dbError.setErrCount(0)
	}
	// Error count >= try , count++ and nothing to do
	if dbError.count >= TryCountDBError {
		dbError.setErrCount(1)
		return DBERROR
	}
	// if status == DBERROR, check SQL is still running.
	// if running, count++ and nothing to do
	// if not running, count++
	if dbError.errStatus == DBERROR {
		if ok, _ := val.sqlCheckSlowQuery(dbError); ok == DBERROR {
			dbError.setErrCount(1)
			return DBERROR
		}
	}
	var errStatus bool
	var errMessage string
	switch id {
	case 1:
		errStatus, errMessage = val.sqlExecSubordinateStatus()
	case 2:
		errStatus, errMessage = val.sqlExecShowStatus()
	case 3:
		errStatus, errMessage = val.sqlExecShowVariables()
	case 4:
		errStatus, errMessage = val.sqlExecPerfSchema()
	case 5:
		errStatus, errMessage = val.sqlExecBufferPool()
	case 6:
		errStatus, errMessage = val.sqlExecInnoDBLock()
	case 7:
		errStatus, errMessage = val.sqlExecThreads()
	case 8:
		errStatus, errMessage = val.sqlExecDigestBase()
	case 9:
		errStatus, errMessage = val.sqlExecDigestRepeat()
	case 10:
		errStatus, errMessage = val.sqlExecFileIOTable()
	}

	dbError.addStatus(errStatus, errMessage)
	if errStatus == DBOK {
		dbError.resetStatus()
		dbError.setErrCount(0)
	} else {
		dbError.setErrCount(1)
		return DBERROR
	}

	return DBOK

}

func (dbError *dbError) setErrCount(setNum int) {
	m.Lock()
	defer m.Unlock()
	if setNum == 0 {
		dbError.count = 0
	} else {
		dbError.count++
	}
}

func (dbError *dbError) addStatus(s bool, v string) {
	m.Lock()
	defer m.Unlock()
	dbError.errStatus = s
	dbError.errMessage = v
}

func (dbError *dbError) resetStatus() {
	m.Lock()
	defer m.Unlock()
	dbError.errStatus = false
	dbError.errMessage = ""
	dbError.count = 0
}

func (val *HostAllInfo) dbPing() bool {
	if err := val.db.Ping(); err != nil {
		return DBERROR
	}
	return DBOK
}

func copyMap(base interface{}, new interface{}) {
	m.Lock()
	defer m.Unlock()
	switch base.(type) {
	case mapProme:
		b := base.(mapProme)
		n := new.(mapProme)
		for key, value := range b {
			n[key] = value
		}
	case mapPrint:
		b := base.(mapPrint)
		n := new.(mapPrint)
		for key, value := range b {
			n[key] = value
		}
	case mapMy:
		b := base.(mapMy)
		n := new.(mapMy)
		for key, value := range b {
			n[key] = value
		}
	}
	return
}

func (val *HostAllInfo) getOsMetric() {
	uri := fmt.Sprintf("http://%s:%d/metrics", val.Hname, val.promPort)
	c := http.Client{
		Timeout: 300 * time.Millisecond,
	}
	res, err := c.Get(uri)
	if err != nil {
		val.promeerror = 1
		return
	}
	val.promeerror = 0
	n, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	defer res.Body.Close()
	promeparam := make(mapProme, 0)
	for _, v := range strings.Split(string(n), "\n") {
		// ignore #
		if len(v) > 0 && v[:1] == "#" {
			continue
		}
		// split by space
		arr := strings.Split(v, " ")
		if len(arr[0]) == 0 || len(arr[1]) == 0 {
			continue
		}
		mtrcname := arr[0] // metoric name
		mtrcval := arr[1]  // metoric value
		switch {
		case mtrcname == SWAPTOL:
			promeparam[SWAPTOL] = rtnInt2Stg(mtrcval, promeparam[SWAPTOL])
		case mtrcname == SWAPFRE:
			promeparam[SWAPFRE] = rtnInt2Stg(mtrcval, promeparam[SWAPFRE])
		case mtrcname == MEMTOL:
			promeparam[MEMTOL] = rtnInt2Stg(mtrcval, promeparam[MEMTOL])
		case mtrcname == MEMFREE:
			promeparam[MEMFREE] = rtnInt2Stg(mtrcval, promeparam[MEMFREE])
		case mtrcname == MEMBUF:
			promeparam[MEMBUF] = rtnInt2Stg(mtrcval, promeparam[MEMBUF])
		case mtrcname == MEMCACHE:
			promeparam[MEMCACHE] = rtnInt2Stg(mtrcval, promeparam[MEMCACHE])
		case mtrcname == LOADAVG1:
			promeparam[LOADAVG1] = rtnInt2Stg(mtrcval, promeparam[LOADAVG1])
		case mtrcname == LOADAVG5:
			promeparam[LOADAVG5] = rtnInt2Stg(mtrcval, promeparam[LOADAVG5])
		case mtrcname == LOADAVG15:
			promeparam[LOADAVG15] = rtnInt2Stg(mtrcval, promeparam[LOADAVG15])
		case mtrcname[:8] == "node_cpu":
			if strings.LastIndex(mtrcname, "user") > 0 {
				promeparam[CPUUSER] = rtnInt2Stg(mtrcval, promeparam[CPUUSER])
			} else if strings.LastIndex(mtrcname, "system") > 0 {
				promeparam[CPUSYS] = rtnInt2Stg(mtrcval, promeparam[CPUSYS])
			} else if strings.LastIndex(mtrcname, "iowait") > 0 {
				promeparam[CPUIO] = rtnInt2Stg(mtrcval, promeparam[CPUIO])
			} else if strings.LastIndex(mtrcname, "idle") > 0 {
				promeparam[CPUIDLE] = rtnInt2Stg(mtrcval, promeparam[CPUIDLE])
			} else if strings.LastIndex(mtrcname, "irq") > 0 {
				promeparam[CPUIRQ] = rtnInt2Stg(mtrcval, promeparam[CPUIRQ])
			}

		case len(mtrcname) > 25 && mtrcname[:26] == INBOUND:
			promeparam[INBOUND] = rtnInt2Stg(mtrcval, promeparam[INBOUND])

		case len(mtrcname) > 26 && mtrcname[:27] == OUTBOUND:
			promeparam[OUTBOUND] = rtnInt2Stg(mtrcval, promeparam[OUTBOUND])

		case len(mtrcname) > 19 && mtrcname[:20] == DISKREAD:
			promeparam[DISKREAD] = rtnInt2Stg(mtrcval, promeparam[DISKREAD])

		case len(mtrcname) > 22 && mtrcname[:23] == DISKWRITE:
			promeparam[DISKWRITE] = rtnInt2Stg(mtrcval, promeparam[DISKWRITE])
		}
	}
	copyMap(promeparam, val.PromeParam)
	return
}

func columnIndex(cols []string, colName string) int {
	for idx := range cols {
		if cols[idx] == colName {
			return idx
		}
	}
	return -1
}

func columnValue(scanArgs []interface{}, cols []string, colName string) string {
	var columnIndex = columnIndex(cols, colName)
	if columnIndex == -1 {
		return ""
	}
	return string(*scanArgs[columnIndex].(*sql.RawBytes))
}

func (val *HostAllInfo) sqlExecBufferPool() (bool, string) {
	var rows *sql.Rows
	var err error
	db := val.db
	rows, err = db.Query(InnodbBufferPoolSQL)
	if err != nil {
		return checkMysqlErr(err)
	}
	defer rows.Close()
	m.Lock()
	var myStatusgoQuery, name, status string
	var count int64
	for rows.Next() {
		if err := rows.Scan(&myStatusgoQuery, &name, &count, &status); err != nil {
			continue
		}
		switch name {
		case BUFPOOLTOTAL:
			val.MyInnoDBBuffer[BUFPOOLTOTAL] = count
		case BUFPOOLDATA:
			val.MyInnoDBBuffer[BUFPOOLDATA] = count
		case BUFPOOLDIRTY:
			val.MyInnoDBBuffer[BUFPOOLDIRTY] = count
		case BUFPOOLFREE:
			val.MyInnoDBBuffer[BUFPOOLFREE] = count
		case LSNMINUSCKPOINT:
			if status == "disabled" {
				val.MyInnoDBBuffer[LSNMINUSCKPOINT] = -1
			} else {
				val.MyInnoDBBuffer[LSNMINUSCKPOINT] = count
			}
		}
	}
	m.Unlock()
	return DBOK, ""
}

func (val *HostAllInfo) sqlExecPerfSchema() (bool, string) {
	var rows *sql.Rows
	var err error
	db := val.db
	rows, err = db.Query(PerfSchemaSQL)
	if err != nil {
		return checkMysqlErr(err)
	}
	defer rows.Close()
	m.Lock()
	var myStatusgoQuery, sqlCnt, avgLatency, maxLatency string
	for rows.Next() {
		if err := rows.Scan(&myStatusgoQuery, &sqlCnt, &avgLatency, &maxLatency); err != nil {
			continue
		}
		val.MyPerfomanceSchema[SQLCNT] = sqlCnt
		val.MyPerfomanceSchema[AVGLATENCY] = avgLatency
		val.MyPerfomanceSchema[MAXLATENCY] = maxLatency
	}
	m.Unlock()
	return DBOK, ""
}

func showStatusCompare(variableName string, value string) bool {
	switch variableName {
	case THCON, THRUN, ABORTCON, SLWLOG, CSEL, CUPDMUL, CUPD, CINSSEL, CINS, CDELMUL, CDEL, CRLCESEL, CRLCE, QCACHHIT, CCALPROCEDURE, CSTMT, CCOMMIT, CROLLBACK:
		return true
	case HADRRDFIRST, HADRRDKEY, HADRRDLAST, HADRRDNXT, HADRRDPRV, HADRRDRND, HADRRDRNDNXT, HADRDEL, HADRUPD, HADRWRT, HADRPREP, HADRCOMT, HADRRB, INNOROWDEL, INNOROWRD, INNOROWINS, INNOROWUPD, PSDIGESTLOST:
		return true
	case BUFWRITEREQ, BUFREADREQ, BUFREAD, BUFCREATED, BUFWRITTEN, BUFFLUSH, BUFDATAP, BUFDIRTYP, BUFFREEP, BUFMISCP, BUFTOTALP:
		return true
	default:
		return false
	}
}
func (val *HostAllInfo) sqlExecShowStatus() (bool, string) {
	var rows *sql.Rows
	var err error
	db := val.db
	rows, err = db.Query(ShowStatusSQL)
	if err != nil {
		return checkMysqlErr(err)
	}
	defer rows.Close()
	var vName, value string
	mapBuf := make(mapMy, 0)
	for rows.Next() {
		if err := rows.Scan(&vName, &value); err != nil {
			continue
		}
		// if add status val, count the value
		if len(mapBuf) == 47 {
			break
		}
		if showStatusCompare(vName, value) {
			mapBuf[vName] = value
		}
	}
	copyMap(mapBuf, val.MyStatu)
	return DBOK, ""
}

func (val *HostAllInfo) sqlExecShowVariables() (bool, string) {
	var rows *sql.Rows
	var err error
	db := val.db
	rows, err = db.Query(ShowVariableSQL)
	if err != nil {
		return checkMysqlErr(err)
	}
	defer rows.Close()
	m.Lock()
	var VariableName, Value string
	for rows.Next() {
		if err := rows.Scan(&VariableName, &Value); err != nil {
			continue
		}
		switch VariableName {
		case RDONLY:
			val.MyVariable[RDONLY] = Value
		case RPLSEMIMAS:
			val.MyVariable[RPLSEMIMAS] = Value
		case RPLSEMISLV:
			val.MyVariable[RPLSEMISLV] = Value
		}
	}
	m.Unlock()
	return DBOK, ""
}

func (val *HostAllInfo) sqlExecSubordinateStatus() (bool, string) {
	var rows *sql.Rows
	var err error
	db := val.db
	rows, err = db.Query(ShowSubordinateStatusSQL)
	if err != nil {
		return checkMysqlErr(err)
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	printNormalModeSecndBhMas := 0
	m.Lock()
	val.SubordinateInfo = make([]SubordinateInfo, 0)
	for rows.Next() {
		scanArgs := make([]interface{}, len(cols))
		for i := range scanArgs {
			scanArgs[i] = &sql.RawBytes{}
		}
		if err := rows.Scan(scanArgs...); err != nil {
			continue
		}
		buf := columnValue(scanArgs, cols, SECBEHMAS)
		i, _ := strconv.Atoi(buf)
		if i > printNormalModeSecndBhMas {
			printNormalModeSecndBhMas = i
		}
		val.MySubordinate[SECBEHMAS] = fmt.Sprintf("%d", printNormalModeSecndBhMas)
		val.SubordinateInfo = append(val.SubordinateInfo, SubordinateInfo{
			Host:          columnValue(scanArgs, cols, MHOST),
			SubordinateIO:       columnValue(scanArgs, cols, SLAVEIO),
			SubordinateSQL:      columnValue(scanArgs, cols, SLAVESQL),
			MsLogFile:     columnValue(scanArgs, cols, MASLOGFILE),
			RMasLogPos:    columnValue(scanArgs, cols, RMASLOGPOS),
			RelMasLogFile: columnValue(scanArgs, cols, RELMASLOGFILE),
			ExeMasLogPos:  columnValue(scanArgs, cols, EXEMASLOGPOS),
			SecBehdMas:    columnValue(scanArgs, cols, SECBEHMAS),
			ChlNm:         columnValue(scanArgs, cols, CHNLNM),
			RetvGTID:      columnValue(scanArgs, cols, RETVGTID),
			ExeGTID:       columnValue(scanArgs, cols, EXEGTID),
			AutoPos:       columnValue(scanArgs, cols, AUTOPOS),
		})
	}
	m.Unlock()
	return DBOK, ""
}

func (val *HostAllInfo) sqlExecDigestBase() (bool, string) {
	var rows *sql.Rows
	var err error
	db := val.db
	rows, err = db.Query(DigestFirstSQL)
	if err != nil {
		return checkMysqlErr(err)
	}
	defer rows.Close()
	m.Lock()
	val.DigestBase = make(mapDigest, 0)
	for rows.Next() {
		var myStatusgoQuery, schemaDigest string
		var execnt uint64
		if err := rows.Scan(&myStatusgoQuery, &schemaDigest, &execnt); err != nil {
			continue
		}
		val.DigestBase[schemaDigest] = execnt
	}
	m.Unlock()
	return DBOK, ""
}

func (val *HostAllInfo) sqlExecDigestRepeat() (bool, string) {
	var rows *sql.Rows
	var err error
	db := val.db
	rows, err = db.Query(DigestRepeatSQL)
	if err != nil {
		return checkMysqlErr(err)
	}
	defer rows.Close()
	m.Lock()
	val.Digest = make(mapDigest, 0)
	var myStatusgoQuery, schemaDigest string
	var execnt uint64
	for rows.Next() {
		if err := rows.Scan(&myStatusgoQuery, &schemaDigest, &execnt); err != nil {
			continue
		}
		val.Digest[schemaDigest] = execnt
	}
	m.Unlock()
	return DBOK, ""
}

func (val *HostAllInfo) sqlExecDigestText(topDigestQuery string) (bool, string) {
	var rows *sql.Rows
	var err error
	db := val.db
	rows, err = db.Query(fmt.Sprintf(DigestTextSQL, topDigestQuery, topDigestQuery))
	if err != nil {
		return checkMysqlErr(err)
	}
	defer rows.Close()
	cnt := 0
	m.Lock()
	for rows.Next() {
		var myStatusgoQuery, digText, schemadigest, digest, schema, full, avglacy, avgExm string
		if err := rows.Scan(&myStatusgoQuery, &digText, &schemadigest, &digest, &schema, &full, &avglacy, &avgExm); err != nil {
			continue
		}
		if val.DigestMetric[cnt].Digest == schemadigest {
			val.DigestMetric[cnt].SchemaDigest = schemadigest
			val.DigestMetric[cnt].SchemaName = schema
			val.DigestMetric[cnt].FullScan = full
			val.DigestMetric[cnt].DigestText = digText
			val.DigestMetric[cnt].LatencyAvg = avglacy
			val.DigestMetric[cnt].RowExmAvg = avgExm
			cnt++
		}
		if cnt == TopN {
			break
		}
	}
	m.Unlock()
	return DBOK, ""
}

func (val *HostAllInfo) sqlExecFileIOTable() (bool, string) {
	var rows *sql.Rows
	var err error
	db := val.db
	rows, err = db.Query(TableFileIOSQL)
	if err != nil {
		return checkMysqlErr(err)
	}
	defer rows.Close()
	m.Lock()
	val.strFileIOperTable = make(mapStrFileIOperTable, 0)
	for rows.Next() {
		var myStatusgoQuery, tableSchema, tableName string
		var countStar, totalLatency, fetchRows, fetchLatency, insertRows, insertLatency, updateRows, updateLatency, deleteRows, deleteLatency uint64
		var readIORequest, readIOByte, readIOLatency, writeIORequest, writeIOByte, writeIOLatency, miscIORequest, miscIOLatency uint64
		if err := rows.Scan(&myStatusgoQuery, &tableSchema, &tableName, &countStar, &totalLatency, &fetchRows, &fetchLatency, &insertRows, &insertLatency, &updateRows, &updateLatency, &deleteRows, &deleteLatency,
			&readIORequest, &readIOByte, &readIOLatency, &writeIORequest, &writeIOByte, &writeIOLatency, &miscIORequest, &miscIOLatency); err != nil {
			continue
		}
		key := fmt.Sprintf("%s.%s", tableSchema, tableName)
		val.strFileIOperTable[key] = &FileIOperTable{
			TableSchema:    tableSchema,
			TableName:      tableName,
			CountStar:      countStar,
			TotalLatency:   totalLatency,
			FetchRows:      fetchRows,
			FetchLatency:   fetchLatency,
			InsertRows:     insertRows,
			InsertLatency:  insertLatency,
			UpdateRows:     updateRows,
			UpdateLatency:  updateLatency,
			DeleteRows:     deleteRows,
			DeleteLatency:  deleteLatency,
			ReadIORequest:  readIORequest,
			ReadIOByte:     readIOByte,
			ReadIOLatency:  readIOLatency,
			WriteIORequest: writeIORequest,
			WriteIOByte:    writeIOByte,
			WriteIOLatency: writeIOLatency,
			MiscIORequest:  miscIORequest,
			MiscIOLatency:  miscIOLatency,
		}
	}
	m.Unlock()
	return DBOK, ""
}

func (val *HostAllInfo) sqlExecThreads() (bool, string) {
	var rows *sql.Rows
	var err error
	db := val.db
	rows, err = db.Query(fmt.Sprintf(ThreadsSQL, TopN))
	if err != nil {
		return checkMysqlErr(err)
	}
	defer rows.Close()
	cnt := 0
	m.Lock()
	val.ThreadsInfo = make([]ThreadsInfo, TopN)
	for rows.Next() {
		var myStatusgoQuery, user, host, db, time, cmd, sql, state string
		if err := rows.Scan(&myStatusgoQuery, &user, &host, &db, &time, &cmd, &sql, &state); err != nil {
		}
		if sql == "" {
			continue
		}
		if len(sql) < 1 {
			if state == "Subordinate has read all relay log; waiting for more updates" {
				continue
			}
		}
		val.ThreadsInfo[cnt].User = user
		val.ThreadsInfo[cnt].Host = host
		val.ThreadsInfo[cnt].Db = db
		val.ThreadsInfo[cnt].Time = time
		val.ThreadsInfo[cnt].State = state
		val.ThreadsInfo[cnt].Sql = sql
		val.ThreadsInfo[cnt].Cmd = cmd
		cnt++
		if cnt == TopN {
			break
		}
	}
	m.Unlock()
	return DBOK, ""
}

func (val *HostAllInfo) sqlExecInnoDBLock() (bool, string) {
	var rows *sql.Rows
	var err error
	db := val.db
	rows, err = db.Query(InnodbLockSQL)
	if err != nil {
		return checkMysqlErr(err)
	}
	defer rows.Close()
	cnt := 0
	m.Lock()
	val.InnodbLockInfo = make([]InnodbLockInfo, TopN)
	for rows.Next() {
		var myStatusgoQuery, waitPidlist, blockPid, waitAge, lockedTable, lockedIndex, blockingQuery, waitCnt string
		if err := rows.Scan(&myStatusgoQuery, &waitPidlist, &waitCnt, &blockPid, &waitAge, &lockedTable, &lockedIndex, &blockingQuery); err != nil {
		}
		val.InnodbLockInfo[cnt].WaitPidlist = waitPidlist
		val.InnodbLockInfo[cnt].WaitCnt = waitCnt
		val.InnodbLockInfo[cnt].BlockPid = blockPid
		val.InnodbLockInfo[cnt].WaitAge = waitAge
		val.InnodbLockInfo[cnt].LockedTable = lockedTable
		val.InnodbLockInfo[cnt].LockedIndex = lockedIndex
		val.InnodbLockInfo[cnt].BlockingQuery = blockingQuery
		cnt++
		if cnt == TopN {
			break
		}
	}
	m.Unlock()
	return DBOK, ""
}

func (val *HostAllInfo) sqlCheckSlowQuery(dbError *dbError) (bool, string) {
	var rows *sql.Rows
	var err error

	db := val.db
	rows, err = db.Query(fmt.Sprintf(SLOWCHECKSQL, dbError.slowCheckKey))
	if err != nil {
		return checkMysqlErr(err)
	}
	defer rows.Close()
	for rows.Next() {
		var cnt int
		if err := rows.Scan(&cnt); err != nil {
		}
		if cnt != 0 {
			return DBERROR, ""
		}
	}
	return DBOK, ""
}

func checkMysqlErr(err error) (bool, string) {
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		return DBOK, mysqlErr.Message
	}
	return DBERROR, ""
}

func rtnInt2Stg(v string, nval float64) float64 {
	f, _ := strconv.ParseFloat(v, 64)
	//i := int64(f)
	return nval + f
}
