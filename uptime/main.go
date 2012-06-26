package main

import (
	"encoding/csv"
	"flag"
	"github.com/papplampe/w32uptime"
	"io"
	"os"
)

const (
	DATE_FORMAT = "02.01.2006 15:04:05"
)

var (
	_file string
)

func init() {
	flag.StringVar(&_file, "o", "", "a file to write csv to")
}

func writeToFile(fd io.Writer, uptimes []w32uptime.Uptime) error {
	out := csv.NewWriter(fd)
	out.Comma = rune(';')
	out.Write([]string{"start", "end"})
	for _, uptime := range uptimes {
		out.Write([]string{uptime.Start.Format(DATE_FORMAT), uptime.End.Format(DATE_FORMAT)})
	}
	out.Flush()

	return nil
}

func main() {
	flag.Parse()

	uptimes, err := w32uptime.ReadAll()
	if err != nil {
		println(err.Error())
		return
	}

	if len(uptimes) == 0 {
		println("no events found")
		return
	}

	if _file != "" {
		fd, err := os.OpenFile(_file, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			println(err.Error())
			return
		}
		defer fd.Close()

		err = writeToFile(fd, uptimes)
	} else {
		err = writeToFile(os.Stdout, uptimes)
	}

	if err != nil {
		println(err.Error())
		return
	}
}
