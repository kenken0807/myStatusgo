package main

import (
	"flag"
	"os"
	"os/exec"
)

func init() {
	clear = make(map[string]func())
	clear["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["darwin"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func main() {
	Opt := Options{}
	parseOptions(&Opt)
	exitcode := run(&Opt)
	os.Exit(exitcode)
}

func parseOptions(Opt *Options) {
	flag.StringVar(&Opt.dbUser, "u", "", "MySQL Username")
	flag.StringVar(&Opt.dbPasswd, "p", "", "MySQL Password")
	flag.StringVar(&Opt.hostJSON, "h", "", "Hostlist written JSON")
	flag.StringVar(&Opt.hostComma, "hc", "", "Hostlist written Comma separated")
	flag.IntVar(&Opt.promPort, "n", 0, "node_exporter port")
	flag.IntVar(&Opt.modeflg, "m", 1, "Start Mode 1:Normal 2:OSResource 3:Threads 4:PerformaceSchema 5:SlaveStatus 6:Handler/InnoDB_Rows 7:InnoDBLockInfo 8.InnoDB Buffer Info 9.Table IO Statistic")
	flag.IntVar(&Opt.digest, "t", 10, "The number of display per instance of Threads ,Performance Schema and Table IO Statistic ")
	flag.Int64Var(&Opt.endSec, "c", 0, "Seconds of Automatic Stopping")
	flag.Int64Var(&Opt.interval, "d", 1, "Delay between updates in seconds")
	flag.StringVar(&Opt.file, "o", "", "Write the content to file and will set autostop 3600 secs")
	flag.BoolVar(&Opt.outputJSON, "j", false, "Write as JSON to file with option -o")
	flag.BoolVar(&Opt.gtid, "g", false, "Gtid Mode at 5:SlaveStatus")
	flag.BoolVar(&Opt.all, "a", false, "Show All metrics per server")
	flag.BoolVar(&Opt.version, "v", false, "Show Version")
	flag.Parse()
}
