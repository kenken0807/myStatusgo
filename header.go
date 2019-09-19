package main

import (
	"database/sql"
	"sync"
	"time"
)

// StartTime ..
var StartTime time.Time

// OutputJSON is whether JSON output or not
var OutputJSON bool

// TopN ... are Some Counters
var TopN int

// TopNPosition ..
var TopNPosition int

// MaxTopNPosition ..
var MaxTopNPosition int

// SortTableFileIOPosition ..
var SortTableFileIOPosition int

// HostnameSize ..
var HostnameSize int

// HostFormatSpace ..
var HostFormatSpace string

// GtidMode ..
var GtidMode bool

// FullSQLStatement
var FullSQLStatement bool

// Global Mutex
var (
	m = sync.Mutex{}
)

var clear map[string]func()

var (
	Error                string
	TopFormat            string
	FinishFormat         string
	OSFormat1            string
	OSFormat2            string
	OSFormat3            string
	MyFormat1            string
	MyFormat2            string
	MyFormat3            string
	MySlaFormat1         string
	MySlaFormat2         string
	MySlaFormat3         string
	MySlaFormatGTID1     string
	MySlaFormatGTID2     string
	MySlaFormatGTID3     string
	MyHadrFormat1        string
	MyHadrFormat2        string
	MyHadrFormat3        string
	KeyArrowUpDownFormat string
	PerfScmFormat1       string
	PerfScmFormat2       string
	PerfScmFormat3       string
	Cursol1              string
	Cursol2              string
	Cursol3              string
	DigestFormat1        string
	DigestFormat2        string
	DigestFormat3        string
	ThreadFormat1        string
	ThreadFormat2        string
	ThreadFormat3        string
	InnoLockFormat1      string
	InnoLockFormat2      string
	InnoLockFormat3      string
	InnoBufFormat1       string
	InnoBufFormat2       string
	InnoBufFormat3       string
	FileIOTabFormat1     string
	FileIOTabFormat3     string
	FileIOTabFormat4     string
	HostNamePort         string
	HostNamePortTag      string
	HostFormatSize       int
	PerfFormatSize       int
)

const (
	// TIMESTAMP ..
	TIMESTAMP = "timestamp"
	// OS Resource
	SWAPTOL  = "node_memory_SwapTotal"
	SWAPFRE  = "node_memory_SwapFree"
	SWAPUSED = "node_memory_SwapUsed"
	INBOUND  = "node_network_receive_bytes"
	OUTBOUND = "node_network_transmit_bytes"
	CPUUSER  = "node_cpu_user"
	CPUSYS   = "node_cpu_system"
	CPUIO    = "node_cpu_iowait"
	CPUIDLE  = "node_cpu_idle"
	CPUIRQ   = "node_cpu_irq"
	/*
		CPUIRQ is...
		node_cpu{cpu="cpu1",mode="irq"} 0
		node_cpu{cpu="cpu1",mode="softirq"} 66.51
	*/
	DISKREAD  = "node_disk_bytes_read"
	DISKWRITE = "node_disk_bytes_written"
	MEMTOL    = "node_memory_MemTotal"
	MEMFREE   = "node_memory_MemFree"
	MEMBUF    = "node_memory_Buffers"
	MEMCACHE  = "node_memory_Cached"
	MEMUSED   = "node_memory_used"
	LOADAVG1  = "node_load1"
	LOADAVG5  = "node_load5"
	LOADAVG15 = "node_load15"
	// SHOW SLAVE STATUS
	MHOST         = "Master_Host"
	SLAVEIO       = "Slave_IO_Running"
	SLAVESQL      = "Slave_SQL_Running"
	MASLOGFILE    = "Master_Log_File"
	RMASLOGPOS    = "Read_Master_Log_Pos"
	RELMASLOGFILE = "Relay_Master_Log_File"
	EXEMASLOGPOS  = "Exec_Master_Log_Pos"
	SECBEHMAS     = "Seconds_Behind_Master"
	CHNLNM        = "Channel_Name"
	RETVGTID      = "Retrieved_Gtid_Set"
	EXEGTID       = "Executed_Gtid_Set"
	AUTOPOS       = "Auto_Position"

	// SHOW GLOBAL STATUS
	THCON         = "Threads_connected"
	THRUN         = "Threads_running"
	ABORTCON      = "Aborted_connects"
	SLWLOG        = "Slow_queries"
	SLWSUM        = "Slowlog_sum"
	CSEL          = "Com_select"
	CUPDMUL       = "Com_update_multi"
	CUPD          = "Com_update"
	CINSSEL       = "Com_insert_select"
	CINS          = "Com_insert"
	CDELMUL       = "Com_delete_multi"
	CDEL          = "Com_delete"
	CRLCESEL      = "Com_replace_select"
	CRLCE         = "Com_replace"
	QCACHHIT      = "Qcache_hits"
	CCALPROCEDURE = "Com_call_procedure"
	CSTMT         = "Com_stmt_execute"
	CCOMMIT       = "Com_commit"
	CROLLBACK     = "Com_rollback"
	QPSALL        = "QPSALL"
	HADRRDFIRST   = "Handler_read_first"
	HADRRDKEY     = "Handler_read_key"
	HADRRDLAST    = "Handler_read_last"
	HADRRDNXT     = "Handler_read_next"
	HADRRDPRV     = "Handler_read_prev"
	HADRRDRND     = "Handler_read_rnd"
	HADRRDRNDNXT  = "Handler_read_rnd_next"
	INNOROWDEL    = "Innodb_rows_deleted"
	INNOROWRD     = "Innodb_rows_read"
	INNOROWINS    = "Innodb_rows_inserted"
	INNOROWUPD    = "Innodb_rows_updated"
	PSDIGESTLOST  = "Performance_schema_digest_lost"
	BUFWRITEREQ   = "Innodb_buffer_pool_write_requests"
	BUFREADREQ    = "Innodb_buffer_pool_read_requests"
	BUFREAD       = "Innodb_buffer_pool_reads"
	BUFCREATED    = "Innodb_pages_created"
	BUFWRITTEN    = "Innodb_pages_written"
	BUFFLUSH      = "Innodb_buffer_pool_pages_flushed"
	BUFDATAP      = "Innodb_buffer_pool_pages_data"
	BUFDIRTYP     = "Innodb_buffer_pool_pages_dirty"
	BUFFREEP      = "Innodb_buffer_pool_pages_free"
	BUFMISCP      = "Innodb_buffer_pool_pages_misc"
	BUFTOTALP     = "Innodb_buffer_pool_pages_total"

	// InnoDB Metric
	BUFPOOLTOTAL    = "buffer_pool_pages_total"
	BUFPOOLDATA     = "buffer_pool_bytes_data"
	BUFPOOLDIRTY    = "buffer_pool_bytes_dirty"
	BUFPOOLFREE     = "buffer_pool_pages_free"
	LSNMINUSCKPOINT = "log_lsn_checkpoint_age"
	//SHOW GLOBAL VARIABLES
	RDONLY     = "read_only"
	RPLSEMIMAS = "rpl_semi_sync_master_enabled"
	RPLSEMISLV = "rpl_semi_sync_slave_enabled"
	// Other
	PSDIGESTALLSQL     = "Performance_schema_digest_allquery"
	SQLCNT             = "Performance_schema.sqlcnt"
	AVGLATENCY         = "Performance_schema.avglatency"
	MAXLATENCY         = "Performance_schema.maxlatency"
	BUFFHITRATE        = "Bufferpool_hitrate"
	ShowStatusSQL      = "SHOW GLOBAL STATUS"
	ShowSlaveStatusSQL = "SHOW SLAVE STATUS"

	ShowVariableSQL = `SHOW VARIABLES WHERE Variable_name='read_only' or Variable_name='rpl_semi_sync_slave_enabled' or Variable_name='rpl_semi_sync_master_enabled'`
	DigestFirstSQL  = `/*!50601 SELECT 'myStatusgo.DigestTextSQL' as mystatusgoquery,  
						concat(schema_name,'|',DIGEST) schema_digest, COUNT_STAR FROM performance_schema.events_statements_summary_by_digest 
						WHERE SCHEMA_NAME not in ('mysql','sys','information_schema','performance_schema') 
						AND DIGEST_TEXT not in ('BEGIN') 
						AND DIGEST_TEXT not like 'SET%' 
						AND DIGEST_TEXT not like 'USE%' 
						AND DIGEST_TEXT not like 'SHOW%' 
						ORDER BY LAST_SEEN DESC,COUNT_STAR DESC,DIGEST DESC LIMIT 5000 */`
	DigestRepeatSQL = `/*!50601 SELECT 'myStatusgo.DigestTextSQL' as mystatusgoquery, 
						concat(schema_name,'|',DIGEST) schema_digest, 
						COUNT_STAR 
						FROM performance_schema.events_statements_summary_by_digest 
						WHERE SCHEMA_NAME not in ('mysql','sys','information_schema','performance_schema') 
						AND DIGEST_TEXT not in ('BEGIN') 
						AND DIGEST_TEXT not like 'SET%' 
						AND DIGEST_TEXT not like 'USE%' 
						AND DIGEST_TEXT not like 'SHOW%' 
						 AND LAST_SEEN > now() - 1 */`
	DigestTextSQL = `/*!50601 SELECT 'myStatusgo.DigestTextSQL' as mystatusgoquery,
						DIGEST_TEXT,
						concat(schema_name,'|',DIGEST) schema_digest,
						DIGEST,
						SCHEMA_NAME,
						IF(((SUM_NO_GOOD_INDEX_USED > 0) OR (SUM_NO_INDEX_USED > 0)),'Yes','No') AS FULL_SCAN,
						AVG_TIMER_WAIT AVG_LATENCY, 
						ROUND(IFNULL((SUM_ROWS_EXAMINED / NULLIF(COUNT_STAR,0)),0),0) AS rows_examined_avg
						FROM  performance_schema.events_statements_summary_by_digest 
						WHERE concat(schema_name,'|',DIGEST) in (%s) ORDER BY FIELD(concat(schema_name,'|',DIGEST), %s) */`
	PerfSchemaSQL = `/*!50601 SELECT 'myStatusgo.PerfSchemaSQL' as mystatusgoquery,IFNULL(count(*), 0) SQLCNT, IFNULL(avg(TIMER_END-TIMER_START),0) AVGLATENCY, IFNULL(max(TIMER_END-TIMER_START),0) MAXLATENCY 
		FROM performance_schema.threads t1 
		INNER JOIN performance_schema.events_statements_current t2 
		ON t2.thread_id = t1.thread_id  WHERE t1.PROCESSLIST_COMMAND != 'Sleep' AND EVENT_NAME not in ('statement/com/Binlog Dump','statement/com/Binlog Dump GTID')*/`
	ThreadsSQL = `/*!50601 SELECT 'myStatusgo.ThreadsSQL' as mystatusgoquery,
						CASE NAME WHEN 'thread/sql/slave_sql' THEN 'slave' ELSE PROCESSLIST_USER END as PROCESSLIST_USER,
						PROCESSLIST_HOST ,
						ifnull(PROCESSLIST_DB,'NULL') ,
						PROCESSLIST_TIME ,
						PROCESSLIST_COMMAND ,
						PROCESSLIST_INFO  ,
						PROCESSLIST_STATE 
						 FROM performance_schema.threads 
						WHERE TYPE='FOREGROUND' 
						      AND NAME in ('thread/sql/one_connection','thread/thread_pool/tp_one_connection','thread/sql/slave_sql')
							  AND ifnull(PROCESSLIST_DB,'none') not in ('mysql')
							  AND PROCESSLIST_COMMAND not in('Sleep','Binlog Dump') 
							  AND PROCESSLIST_ID != @@pseudo_thread_id 
							  AND (PROCESSLIST_INFO not like '%%myStatusgo%%'
							  OR PROCESSLIST_TIME > 0)
							  ORDER BY PROCESSLIST_TIME DESC LIMIT %d */`
	InnodbLockSQL = `/*!50601 SELECT 'myStatusgo.InnodbLockSQL' as mystatusgoquery,
						SUBSTRING(GROUP_CONCAT(waiting_pid),1,50) waitPidlist,
						count(waiting_pid) waitCnt ,
						blocking_pid blockPid,
						MAX(wait_age_secs) waitAge,
						max(locked_table) lockedTable,
						max(locked_index) lockedIndex,
						max(blocking_query) blockingQuery
						FROM
						(SELECT 
							TIMESTAMPDIFF(SECOND,r.trx_wait_started, NOW()) AS wait_age_secs,
							rl.lock_table AS locked_table,
							rl.lock_index AS locked_index,
							r.trx_mysql_thread_id AS waiting_pid,
							r.trx_query AS waiting_query,
							b.trx_mysql_thread_id AS blocking_pid,
							b.trx_query AS blocking_query
						FROM
							(((( information_schema.innodb_lock_waits w
							JOIN information_schema.innodb_trx b ON ((b.trx_id = w.blocking_trx_id)))
							JOIN information_schema.innodb_trx r ON ((r.trx_id = w.requesting_trx_id)))
							JOIN information_schema.innodb_locks bl ON ((bl.lock_id = w.blocking_lock_id)))
							JOIN information_schema.innodb_locks rl ON ((rl.lock_id = w.requested_lock_id)))
						) a
						LEFT JOIN (SELECT DISTINCT waiting_pid AS wpid FROM (SELECT 
							TIMESTAMPDIFF(SECOND,r.trx_wait_started, NOW()) AS wait_age_secs,
							rl.lock_table AS locked_table,
							rl.lock_index AS locked_index,
							r.trx_mysql_thread_id AS waiting_pid,
							r.trx_query AS waiting_query,
							b.trx_mysql_thread_id AS blocking_pid,
							b.trx_query AS blocking_query
						FROM
							(((( information_schema.innodb_lock_waits w
							JOIN information_schema.innodb_trx b ON ((b.trx_id = w.blocking_trx_id)))
							JOIN information_schema.innodb_trx r ON ((r.trx_id = w.requesting_trx_id)))
							JOIN information_schema.innodb_locks bl ON ((bl.lock_id = w.blocking_lock_id)))
							JOIN information_schema.innodb_locks rl ON ((rl.lock_id = w.requested_lock_id)))
						) a) b 
						ON a.blocking_pid = b.wpid WHERE wpid IS NULL
						GROUP BY blocking_pid ORDER BY waitCnt desc */`
	InnodbBufferPoolSQL = `/*!50601 SELECT 'myStatusgo.InnodbBufferPoolSQL' as mystatusgoquery,
							NAME,
							CASE NAME 
							WHEN 'buffer_pool_pages_total' THEN COUNT*@@innodb_page_size 
							WHEN 'buffer_pool_pages_free'  THEN COUNT*@@innodb_page_size 
							ELSE COUNT END as COUNT,
							STATUS
							from information_schema.INNODB_METRICS 
							WHERE NAME in ('buffer_pool_pages_total','buffer_pool_bytes_dirty','buffer_pool_bytes_data','buffer_pool_pages_free', 'log_lsn_checkpoint_age') */`

	TableFileIOSQL = `/*!50601 SELECT 'myStatusgo.TableFileIOSQL' as mystatusgoquery ,
	                    pst.OBJECT_SCHEMA AS tableSchema,
						pst.OBJECT_NAME AS tableName,
						ifnull(pst.COUNT_STAR,0) + ifnull(fsbi.cntstar,0) as countStar,
						pst.SUM_TIMER_WAIT AS totalLatency,
						pst.COUNT_FETCH AS fetchRows,
						pst.SUM_TIMER_FETCH AS fetchLatency,
						pst.COUNT_INSERT AS insertRows,
						pst.SUM_TIMER_INSERT AS insertLatency,
						pst.COUNT_UPDATE AS updateRows,
						pst.SUM_TIMER_UPDATE AS updateLatency,
						pst.COUNT_DELETE AS deleteRows,
						pst.SUM_TIMER_DELETE AS deleteLatency,
						fsbi.count_read AS readIORequest,
						fsbi.sum_number_of_bytes_read AS readIOByte,
						fsbi.sum_timer_read AS readIOLatency,
						fsbi.count_write AS writeIORequest,
						fsbi.sum_number_of_bytes_write AS writeIOByte,
						fsbi.sum_timer_write AS writeIOLatency,
						fsbi.count_misc AS miscIORequest,
						fsbi.sum_timer_misc AS miscIOLatency
						from
						(
							performance_schema.table_io_waits_summary_by_table pst
							left join (select
								LEFT(SUBSTRING_INDEX(SUBSTRING_INDEX(REPLACE(fsbi.FILE_NAME, '\\', '/'), '/', -2), '/', 1), 64) AS table_schema,
								-- LEFT(SUBSTRING_INDEX(REPLACE(SUBSTRING_INDEX(REPLACE(fsbi.FILE_NAME, '\\', '/'), '/', -1), '@0024', '$'), '.', 1), 64) AS table_name,
								LEFT(SUBSTRING_INDEX(SUBSTRING_INDEX(REPLACE(SUBSTRING_INDEX(REPLACE(fsbi.FILE_NAME, '\\', '/'), '/', -1), '@0024', '$'), '.', 1),'#',1), 64) AS table_name,
								sum(fsbi.COUNT_STAR) as cntstar,
								sum(fsbi.COUNT_READ) AS count_read,
								sum(fsbi.SUM_NUMBER_OF_BYTES_READ) AS sum_number_of_bytes_read,
								sum(fsbi.SUM_TIMER_READ) AS sum_timer_read,
								sum(fsbi.COUNT_WRITE) AS count_write,
								sum(fsbi.SUM_NUMBER_OF_BYTES_WRITE) AS sum_number_of_bytes_write,
								sum(fsbi.SUM_TIMER_WRITE) AS sum_timer_write,
								sum(fsbi.COUNT_MISC) AS count_misc,
								sum(fsbi.SUM_TIMER_MISC) AS sum_timer_misc
								from
								performance_schema.file_summary_by_instance fsbi
								WHERE COUNT_STAR > 0 
								and LEFT(SUBSTRING_INDEX(SUBSTRING_INDEX(REPLACE(fsbi.FILE_NAME, '\\', '/'), '/', -2), '/', 1), 64) not in ('performance_schema','sys','mysql')
								group by
								table_schema,
								table_name) fsbi 
								on(
									(
										(pst.OBJECT_SCHEMA = fsbi.table_schema)
										and (pst.OBJECT_NAME = fsbi.table_name)
									)
								)
							)
							WHERE pst.OBJECT_SCHEMA not in ('performance_schema','sys','mysql') */`
	SLOWCHECKSQL = `/*!50601 SELECT count(1) FROM performance_schema.threads 
	                      WHERE PROCESSLIST_COMMAND='QUERY' and PROCESSLIST_TIME > 0 and PROCESSLIST_INFO like '%%%s%%' */`
	FILEIOCHECKSLOW     = "myStatusgo.TableFileIOSQL"
	DIGESTCHECKSLOW     = "myStatusgo.DigestTextSQL"
	BUFFERPOOLCHECKSLOW = "myStatusgo.InnodbBufferPoolSQL"
	LOCKCHECKSLOW       = "myStatusgo.InnodbLockSQL"
	THREADCHECKSLOW     = "myStatusgo.ThreadsSQL"
	PERFCHECKSLOW       = "myStatusgo.PerfSchemaSQL"

	// Format
	RED                 = "\x1b[31m"
	GREEN               = "\x1b[32m"
	YELLOW              = "\x1b[33m"
	BLUE                = "\x1b[34m"
	MAZENDA             = "\x1b[35m"
	CYAN                = "\x1b[36m"
	WHITE               = "\x1b[37m"
	COLEND              = "\x1b[0m"
	MODEOFF             = 0
	MODENORMAL          = 1
	MODERESOURCE        = 2
	MODETHREADS         = 3
	MODEPERFORMACE      = 4
	MODESLAVE           = 5
	MODEROW             = 6
	MODEINNOLOCK        = 7
	MODEINNOBUFFER      = 8
	MODEFILEIOTABLE     = 9
	MODEMAX             = 9
	SORTNUM1            = 1
	SORTNUM2            = 2
	SORTNUM3            = 3
	SORTNUM4            = 4
	SORTNUM5            = 5
	SORTNUM6            = 6
	SORTNUM7            = 7
	SORTNUMMAX          = 7
	TIMELAYOUT          = "2006-01-02 15:04:05"
	MAXHOSTNAMESIZE     = 25
	DEFAULTHOSTNAMESIZE = 8
	DEFAULTPORTSIZE     = 5
	DBERROR             = true
	DBOK                = false
	MYSLAFORMATSPACE    = "                                                                                                                                                                                      "
	DIGESTFORMAT0       = "                              *"
	THREADFORMAT0       = "  *"
	INNOLOCKFORMAT0     = "  *"
	FILEIOTABFORMAT0    = "  *"
	INSTANCECURSOR      = "*"
	TryCountDBError     = 2
	WaitCountDBError    = TryCountDBError + 60
)

type mapMy map[string]string
type mapInnoDBBuffer map[string]int64
type mapDigest map[string]uint64
type mapFileIO map[string]uint64
type mapProme map[string]float64
type mapPrint map[string]interface{}
type mapStrFileIOperTable map[string]*FileIOperTable

type HostAllInfo struct {
	Hname                          string
	port                           string
	alias                          string
	dbuser                         string
	dbpasswd                       string
	promPort                       int
	dberror                        bool
	promeerror                     int
	db                             *sql.DB
	dbErrorPerQuery                dbErrorPerQuery
	topNPosition                   int
	PromeParam                     mapProme `json:"-"`
	MySlave                        mapMy    `json:"-"`
	MyStatu                        mapMy    `json:"-"`
	MyVariable, MyPerfomanceSchema mapMy
	MyInnoDBBuffer                 mapInnoDBBuffer
	DigestBase                     mapDigest `json:"-"` // 最初に全件取得するやつ
	Digest                         mapDigest `json:"-"` // WHERE LAST_SEEN > now() - 1 を保管
	DigestALLSqlCnt                string
	DigestMetric                   []DigestMetric
	ThreadsInfo                    []ThreadsInfo
	InnodbLockInfo                 []InnodbLockInfo
	strFileIOperTable              mapStrFileIOperTable
	FileIOMetric                   []FileIOperTable
	Metric                         mapPrint `json:"-"`
	MetricOS                       mapPrint `json:"-"`
	MetricJSON                     mapPrint
	slowSum                        int64
	SlaveInfo                      []SlaveInfo
}

type Options struct {
	dbUser      string
	dbPasswd    string
	promPort    int
	modeflg     int
	digest      int
	endSec      int64
	interval    int64
	file        string
	hostJSON    string
	hostComma   string
	queryLength int
	outputJSON  bool
	gtid        bool
}

type sortDigestCnt struct {
	sDigest string
	sCnt    int64
}

type DigestMetric struct {
	SchemaDigest string
	Digest       string
	CntStar      uint64
	SchemaName   string
	FullScan     string
	DigestText   string
	LatencyAvg   string
	RowExmAvg    string
}

type ThreadsInfo struct {
	User  string `json:"User"`
	Host  string `json:"Host"`
	Db    string `json:"Db"`
	Time  string `json:"Time"`
	State string `json:"State"`
	Sql   string `json:"Sql"`
	Cmd   string `json:"Cmd"`
}

type InnodbLockInfo struct {
	WaitPidlist   string
	BlockPid      string
	WaitAge       string
	LockedTable   string
	LockedIndex   string
	BlockingQuery string
	WaitCnt       string
}

type SlaveInfo struct {
	Host          string `json:"Host"`
	SlaveIO       string `json:"SlaveIO"`
	SlaveSQL      string `json:"SlaveSQL"`
	MsLogFile     string `json:"MsLogFile"`
	RMasLogPos    string `json:"RMasLogPos"`
	RelMasLogFile string `json:"RelMasLogFile"`
	ExeMasLogPos  string `json:"ExeMasLogPos"`
	SecBehdMas    string `json:"SecBehdMas"`
	ChlNm         string `json:"ChlNm"`
	RetvGTID      string `json:"RetvGTID"`
	ExeGTID       string `json:"ExeGTID"`
	AutoPos       string `json:"AutoPos"`
}

type hostListJson struct {
	ServiceList []struct {
		ServiceName string   `json:"serviceName"`
		Host        []string `json:"host"`
	} `json:"serviceList"`
}

type FileIOperTable struct {
	TableSchema    string
	TableName      string
	CountStar      uint64
	TotalLatency   uint64
	FetchRows      uint64
	FetchLatency   uint64
	InsertRows     uint64
	InsertLatency  uint64
	UpdateRows     uint64
	UpdateLatency  uint64
	DeleteRows     uint64
	DeleteLatency  uint64
	ReadIORequest  uint64
	ReadIOByte     uint64
	ReadIOLatency  uint64
	WriteIORequest uint64
	WriteIOByte    uint64
	WriteIOLatency uint64
	MiscIORequest  uint64
	MiscIOLatency  uint64
}

type dbErrorPerQuery struct {
	slave      dbError
	status     dbError
	variables  dbError
	perf       dbError
	bufferPool dbError
	lock       dbError
	digest     dbError
	threads    dbError
	fileIO     dbError
}
type dbError struct {
	errStatus    bool
	errMessage   string
	slowCheckKey string
	count        int
}
type Entry struct {
	sKey string
	sCnt uint64
}
type List []Entry
