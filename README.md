# myStatusgo


`myStatusgo` is a realtime monitoring tool for MySQL with status information,performance schema(MySQL5.6+) and etc..(terminal base)
This tool can monitor multiple MySQL instances.

`myStatusgo` collects metric and display them.
1. Normal.. Display QPS info from Com_xx of SHOW GLOBAL STATUS
1. OS Metric.. Display QPS info and OS metric. OS metric collect from node_expoter when the MySQL server installed node_expoter
1. Threads.. Display Running Threads from performance_schema.threads table. (MySQL5.6+)
1. P_S Info.. Display digest_text from performance_schema.statement_digest table (MySQL5.6+)
1. SlaveStatus.. Display Slave Info from SHOW SLAVE STATUS
1. Handler/InnoDB_Rows.. Display innodb_row_xx and Handler_xx from SHOW GLOBAL STATUS
1. InnoDB Lock Info.. Display InnoDB Row Lock Info from information_schema.innodb_trx and innodb_locks and innodb_lock_waits (MySQL5.6+)
1. InnoDB Buffer Info.. Display InnoDB Buffer Pool Info from information_schema.INNODB_METRICS (MySQL5.6+)
1. Table IO Statistic.. Display information of I/O request of each table and information on the number of SELECT/DML for each table from performance_schema.file_summary_by_instance and performance_schema.table_io_waits_summary_by_table (MySQL5.6+)

## Grants
myStatusgo needs following permissions user

```
GRANT SELECT on performance_schema.* to 'user'@'host';
GRANT PROCESS,REPLICATION CLIENT on *.* to 'user'@'host';
```

## Install
```
go get -u github.com/kenken0807/myStatusgo
```
The binary will be builted and installed into $GOPATH/bin/. 

## Quick Start
```
$ myStatusgo -hc Hostname:Port[:alias],Hostname:Port[:alias] -u User -p Password

Example $ myStatusgo -hc 192.168.0.1:3306:Master,192.168.0.2:3306:Slave1 -u user1 -p userpass
```
![oss_sample_qps](https://user-images.githubusercontent.com/13253434/60561472-93acc100-9d8e-11e9-843f-9dad4e564ca9.png)

## How to specify Hostname
There are three ways to specify hosts.

* -h option
JSON format
```
$ myStatusgo -h '{"ServiceList":[{"serviceName":"Sample","host":["192.168.0.1:3306","192.168.0.2:3306"]}]}'
```

* -hc option
Comma separated
```
$ myStatusgo -hc 192.168.0.1:3306,192.168.0.2:3306
```

* Specify a file as an argument
```
$ vim hostlist
# Sample
192.168.0.1:3306
192.168.0.2:3306

$ myStatusgo hostlist
```

## Options
```
Usage of myStatusgo:
  -a	Show All metrics per server
  -c int
    	Seconds of Automatic Stopping
  -d int
    	Delay between updates in seconds (default 1)
  -g	Gtid Mode at 5:SlaveStatus
  -h string
    	Hostlist written JSON
  -hc string
    	Hostlist written Comma separated
  -j	Write as JSON to file with option -o
  -m int
    	Start Mode 1:Normal 2:OSResource 3:Threads 4:PerformaceSchema 5:SlaveStatus 6:Handler/InnoDB_Rows 7:InnoDBLockInfo 8.InnoDB Buffer Info 9.Table IO Statistic (default 1)
  -n int
    	node_exporter port
  -o string
    	Write the content to file and will set autostop 3600 secs
  -p string
    	MySQL Password
  -t int
    	The number of display per instance of Threads ,Performance Schema and Table IO Statistic  (default 10)
  -u string
    	MySQL Username
  -v	Show Version
```

## Other
* -n *port* option
Also display OS metrics info if MySQL server has node exporter

```
$ myStatusgo -h '{"ServiceList":[{"serviceName":"Sample","host":["192.168.0.1:3306:Master","192.168.0.2:3306:Slave1","192.168.0.3:3306:Slave2"]}]}' -u test -p test -n 9100
```
![oss_sample qpsos](https://user-images.githubusercontent.com/13253434/60561503-b0e18f80-9d8e-11e9-8846-c3797a6fcbda.png)
* -t *number* option(default 10)
The number of display per instance of Threads ,Performance Schema and Table IO Statistic

![oss_psinfo](https://user-images.githubusercontent.com/13253434/60561522-c656b980-9d8e-11e9-9147-b2c4edee3351.png)

* -a Option
show all metrics which retrieved by myStatusgo per server
<img width="1321" alt="allmetrics" src="https://user-images.githubusercontent.com/13253434/78538796-81cfe100-782c-11ea-9730-a1f568f7442c.png">
