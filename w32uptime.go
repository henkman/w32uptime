package w32uptime

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/AllenDang/w32"
	"syscall"
	"time"
	"unsafe"
)

const (
	EVENT_ID_SHUTDOWN = 6006
	EVENT_ID_STARTUP  = 6009
)

type Uptime struct {
	Start time.Time
	End   time.Time
}

func ReadAll() ([]Uptime, error) {
	var read, req, offset, recordsize uint32
	var lastEventDate time.Time
	var record w32.EVENTLOGRECORD
	var curUptime Uptime
	findstart := true
	uptimes := make([]Uptime, 0)
	buffer := make([]byte, 1024*64)
	recordsize = uint32(unsafe.Sizeof(record))

	eventlog := w32.OpenEventLog(nil, syscall.StringToUTF16Ptr("system"))
	if eventlog == 0 {
		return nil, errors.New("could not open event log")
	}
	defer w32.CloseEventLog(eventlog)

	for w32.ReadEventLog(eventlog, w32.EVENTLOG_SEQUENTIAL_READ|w32.EVENTLOG_FORWARDS_READ,
		0, buffer, uint32(len(buffer)), &read, &req) {

		offset = 0
		for offset < read {
			in := bytes.NewBuffer(buffer[offset : offset+recordsize])
			err := binary.Read(in, binary.LittleEndian, &record)
			if err != nil {
				return nil, err
			}

			eventid := record.EventID & 0xFFFF
			tm := time.Unix(int64(record.TimeGenerated), 0)
			if findstart && eventid == EVENT_ID_STARTUP {
				findstart = false
				curUptime.Start = tm
			} else if !findstart {
				// there is no corresponding shutdown entry in the log
				// we just take the time of the last event in the log as the shutdown date
				if eventid == EVENT_ID_STARTUP {
					curUptime.End = lastEventDate
					uptimes = append(uptimes, curUptime)
					findstart = true
				} else if eventid == EVENT_ID_SHUTDOWN {
					curUptime.End = tm
					uptimes = append(uptimes, curUptime)
					findstart = true
				}
			}
			lastEventDate = tm

			offset += record.Length
		}
	}

	// catch the current uptime
	if !findstart {
		curUptime.End = time.Now()
		uptimes = append(uptimes, curUptime)
	}

	return uptimes, nil
}
