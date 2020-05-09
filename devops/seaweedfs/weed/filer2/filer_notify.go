package filer2

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/chrislusf/seaweedfs/weed/glog"
	"github.com/chrislusf/seaweedfs/weed/notification"
	"github.com/chrislusf/seaweedfs/weed/pb/filer_pb"
	"github.com/chrislusf/seaweedfs/weed/util"
)

func (f *Filer) NotifyUpdateEvent(oldEntry, newEntry *Entry, deleteChunks bool) {
	var fullpath string
	if oldEntry != nil {
		fullpath = string(oldEntry.FullPath)
	} else if newEntry != nil {
		fullpath = string(newEntry.FullPath)
	} else {
		return
	}

	// println("fullpath:", fullpath)

	if strings.HasPrefix(fullpath, SystemLogDir) {
		return
	}

	newParentPath := ""
	if newEntry != nil {
		newParentPath, _ = newEntry.FullPath.DirAndName()
	}
	eventNotification := &filer_pb.EventNotification{
		OldEntry:      oldEntry.ToProtoEntry(),
		NewEntry:      newEntry.ToProtoEntry(),
		DeleteChunks:  deleteChunks,
		NewParentPath: newParentPath,
	}

	if notification.Queue != nil {
		glog.V(3).Infof("notifying entry update %v", fullpath)
		notification.Queue.SendMessage(fullpath, eventNotification)
	}

	f.logMetaEvent(fullpath, eventNotification)

}

func (f *Filer) logMetaEvent(fullpath string, eventNotification *filer_pb.EventNotification) {

	dir, _ := util.FullPath(fullpath).DirAndName()

	event := &filer_pb.SubscribeMetadataResponse{
		Directory:         dir,
		EventNotification: eventNotification,
		TsNs:              time.Now().UnixNano(),
	}
	data, err := proto.Marshal(event)
	if err != nil {
		glog.Errorf("failed to marshal filer_pb.SubscribeMetadataResponse %+v: %v", event, err)
		return
	}

	f.MetaLogBuffer.AddToBuffer([]byte(dir), data)

}

func (f *Filer) logFlushFunc(startTime, stopTime time.Time, buf []byte) {

	targetFile := fmt.Sprintf("%s/%04d-%02d-%02d/%02d-%02d.segment", SystemLogDir,
		startTime.Year(), startTime.Month(), startTime.Day(), startTime.Hour(), startTime.Minute(),
		// startTime.Second(), startTime.Nanosecond(),
	)

	if err := f.appendToFile(targetFile, buf); err != nil {
		glog.V(0).Infof("log write failed %s: %v", targetFile, err)
	}
}

func (f *Filer) ReadPersistedLogBuffer(startTime time.Time, eachLogEntryFn func(logEntry *filer_pb.LogEntry) error) error {

	startDate := fmt.Sprintf("%04d-%02d-%02d", startTime.Year(), startTime.Month(), startTime.Day())
	startHourMinute := fmt.Sprintf("%02d-%02d.segment", startTime.Hour(), startTime.Minute())

	sizeBuf := make([]byte, 4)
	startTsNs := startTime.UnixNano()

	dayEntries, listDayErr := f.ListDirectoryEntries(context.Background(), SystemLogDir, startDate, true, 366)
	if listDayErr != nil {
		return fmt.Errorf("fail to list log by day: %v", listDayErr)
	}
	for _, dayEntry := range dayEntries {
		// println("checking day", dayEntry.FullPath)
		hourMinuteEntries, listHourMinuteErr := f.ListDirectoryEntries(context.Background(), util.NewFullPath(SystemLogDir, dayEntry.Name()), "", false, 24*60)
		if listHourMinuteErr != nil {
			return fmt.Errorf("fail to list log %s by day: %v", dayEntry.Name(), listHourMinuteErr)
		}
		for _, hourMinuteEntry := range hourMinuteEntries {
			// println("checking hh-mm", hourMinuteEntry.FullPath)
			if dayEntry.Name() == startDate {
				if strings.Compare(hourMinuteEntry.Name(), startHourMinute) < 0 {
					continue
				}
			}
			// println("processing", hourMinuteEntry.FullPath)
			chunkedFileReader := NewChunkStreamReaderFromFiler(f.MasterClient, hourMinuteEntry.Chunks)
			if err := ReadEachLogEntry(chunkedFileReader, sizeBuf, startTsNs, eachLogEntryFn); err != nil {
				chunkedFileReader.Close()
				if err == io.EOF {
					break
				}
				return fmt.Errorf("reading %s: %v", hourMinuteEntry.FullPath, err)
			}
			chunkedFileReader.Close()
		}
	}

	return nil
}

func ReadEachLogEntry(r io.Reader, sizeBuf []byte, ns int64, eachLogEntryFn func(logEntry *filer_pb.LogEntry) error) error {
	for {
		n, err := r.Read(sizeBuf)
		if err != nil {
			return err
		}
		if n != 4 {
			return fmt.Errorf("size %d bytes, expected 4 bytes", n)
		}
		size := util.BytesToUint32(sizeBuf)
		// println("entry size", size)
		entryData := make([]byte, size)
		n, err = r.Read(entryData)
		if err != nil {
			return err
		}
		if n != int(size) {
			return fmt.Errorf("entry data %d bytes, expected %d bytes", n, size)
		}
		logEntry := &filer_pb.LogEntry{}
		if err = proto.Unmarshal(entryData, logEntry); err != nil {
			return err
		}
		if logEntry.TsNs <= ns {
			return nil
		}
		// println("each log: ", logEntry.TsNs)
		if err := eachLogEntryFn(logEntry); err != nil {
			return err
		}
	}
}