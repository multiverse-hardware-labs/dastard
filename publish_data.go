package dastard

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/usnistgov/dastard/ljh"

	czmq "github.com/zeromq/goczmq"
)

// DataPublisher contains many optional methods for publishing data, any methods that are non-nil will be used
// in each call to PublishData
type DataPublisher struct {
	PubRecords   *czmq.Channeler
	PubSummaries *czmq.Channeler
	LJH22        *ljh.Writer
	LJH3         *ljh.Writer3
}

// SetLJH3 adds an LJH3 writer to dp, the .file attribute is nil, and will be instantiated upon next call to dp.WriteRecord
func (dp *DataPublisher) SetLJH3(ChanNum int, Timebase float64,
	NumberOfRows int, NumberOfColumns int, FileName string) {
	w := ljh.Writer3{ChanNum: ChanNum,
		Timebase:        Timebase,
		NumberOfRows:    NumberOfRows,
		NumberOfColumns: NumberOfColumns,
		FileName:        FileName}
	dp.LJH3 = &w
}

// HasLJH22 returns true if LJH22 is non-nil, used to decide if writeint to LJH22 should occur
func (dp *DataPublisher) HasLJH3() bool {
	return dp.LJH3 != nil
}
func (dp *DataPublisher) RemoveLJH3() {
	dp.LJH3.Close()
	dp.LJH3 = nil
}

// SetLJH22 adds an LJH22 writer to dp, the .file attribute is nil, and will be instantiated upon next call to dp.WriteRecord
func (dp *DataPublisher) SetLJH22(ChanNum int, Presamples int, Samples int, Timebase float64, TimestampOffset float64,
	NumberOfRows int, NumberOfColumns int, FileName string) {
	w := ljh.Writer{ChanNum: ChanNum,
		Presamples:      Presamples,
		Samples:         Samples,
		Timebase:        Timebase,
		TimestampOffset: TimestampOffset,
		NumberOfRows:    NumberOfRows,
		NumberOfColumns: NumberOfColumns,
		FileName:        FileName}
	dp.LJH22 = &w
}

// HasLJH22 returns true if LJH22 is non-nil, used to decide if writeint to LJH22 should occur
func (dp *DataPublisher) HasLJH22() bool {
	return dp.LJH22 != nil
}
func (dp *DataPublisher) RemoveLJH22() {
	dp.LJH22.Close()
	dp.LJH22 = nil
}

func (dp *DataPublisher) HasPubRecords() bool {
	return dp.PubRecords != nil
}

// SetPubRecords starts publishing records with czmq over tcp at port=PortTrigs
func (dp *DataPublisher) SetPubRecords() {
	hostname := fmt.Sprintf("tcp://*:%d", PortTrigs)
	dp.SetPubRecordsWithHostname(hostname)
}
func (dp *DataPublisher) SetPubRecordsWithHostname(hostname string) {
	if dp.PubRecords != nil {
		panic("dont set this twice! Destroy first!")
	}
	dp.PubRecords = czmq.NewPubChanneler(hostname)
}
func (dp *DataPublisher) RemovePubRecords() {
	dp.PubRecords.Destroy()
	dp.PubRecords = nil
}

func (dp *DataPublisher) HasPubSummaries() bool {
	return dp.PubSummaries != nil
}

// SetPubSummaries starts publishing summaries with czmq over tcp at port=PortTrigs
func (dp *DataPublisher) SetPubSummaries() {
	hostname := fmt.Sprintf("tcp://*:%d", PortSummaries)
	dp.SetPubSummariesWithHostname(hostname)
}
func (dp *DataPublisher) SetPubSummariesWithHostname(hostname string) {
	if dp.PubSummaries != nil {
		panic("dont set this twice! Destroy first!")
	}
	dp.PubSummaries = czmq.NewPubChanneler(hostname)
}
func (dp *DataPublisher) RemovePubSummaries() {
	dp.PubSummaries.Destroy()
	dp.PubSummaries = nil
}

// PublishData looks at each member of DataPublisher, and if it is non-nil, publishes each record into that member
func (dp DataPublisher) PublishData(records []*DataRecord) error {
	if dp.HasPubRecords() {
		for _, record := range records {

			dp.PubRecords.SendChan <- messageRecords(record)
		}
	}
	if dp.HasPubSummaries() {
		for _, record := range records {

			dp.PubSummaries.SendChan <- messageSummaries(record)
		}
	}
	if dp.HasLJH22() {
		for _, record := range records {
			if !dp.LJH22.HeaderWritten { // MATTER doesn't create ljh files until at least one record exists, let us do the same
				// if the file doesn't exists yet, create it and write header
				err := dp.LJH22.CreateFile()
				if err != nil {
					return err
				}
				dp.LJH22.WriteHeader()
			}
			nano := record.trigTime.UnixNano()
			data := make([]uint16, len(record.data))
			for i, v := range record.data {
				data[i] = uint16(v)
			}
			dp.LJH22.WriteRecord(int64(record.trigFrame), int64(nano)/1000, data)
		}
	}
	if dp.HasLJH3() {
		for _, record := range records {
			if !dp.LJH3.HeaderWritten { // MATTER doesn't create ljh files until at least one record exists, let us do the same
				// if the file doesn't exists yet, create it and write header
				err := dp.LJH3.CreateFile()
				if err != nil {
					return err
				}
				dp.LJH3.WriteHeader()
			}
			nano := record.trigTime.UnixNano()
			data := make([]uint16, len(record.data))
			for i, v := range record.data {
				data[i] = uint16(v)
			}
			dp.LJH3.WriteRecord(record.presamples+1, int64(record.trigFrame), int64(nano)/1000, data)
		}
	}
	return nil
}

// messageSummaries makes a message with the following format for publishing on portTrigs
// Structure of the message header is defined in BINARY_FORMATS.md
// 16 bits: channel number
//  8 bits: header version number
//  8 bits: Presamples
//  8 bits: length of record
//  8 bits: pretrigMean
//  8 bits: peakValue
//  8 bits: residualStdDev
//  8 bits: UnixNano trigTime
//  8 bits: trigFrame
//  end of first message packet
//  modelCoefs, each coef is float32, length can vary
func messageSummaries(rec *DataRecord) [][]byte {
	const headerVersion = uint8(0)

	header := new(bytes.Buffer)
	binary.Write(header, binary.LittleEndian, uint16(rec.channum))
	binary.Write(header, binary.LittleEndian, headerVersion)
	binary.Write(header, binary.LittleEndian, uint32(rec.presamples))
	binary.Write(header, binary.LittleEndian, uint32(len(rec.data)))
	binary.Write(header, binary.LittleEndian, float32(rec.pretrigMean))    // TODO: change to volts/arb
	binary.Write(header, binary.LittleEndian, float32(rec.peakValue))      // TODO: change to volts/arb
	binary.Write(header, binary.LittleEndian, float32(rec.residualStdDev)) // TODO: change to volts/arb
	nano := rec.trigTime.UnixNano()
	binary.Write(header, binary.LittleEndian, uint64(nano))
	binary.Write(header, binary.LittleEndian, uint64(rec.trigFrame))

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, rec.modelCoefs)
	return [][]byte{header.Bytes(), buf.Bytes()}
}

// messageRecords makes a message with the following format for publishing on portTrigs
// Structure of the message header is defined in BINARY_FORMATS.md
// 16 bits: channel number
//  8 bits: header version number
//  8 bits: code for data type (0-1 = 8 bits; 2-3 = 16; 4-5 = 32; 6-7 = 64; odd=uint; even=int)
// 32 bits: # of pre-trigger samples
// 32 bits: # of samples, total
// 32 bits: sample period, in seconds (float)
// 32 bits: volts per arb conversion (float)
// 64 bits: trigger time, in ns since epoch 1970
// 64 bits: trigger frame #
func messageRecords(rec *DataRecord) [][]byte {

	const headerVersion = uint8(0)
	const dataType = uint8(3)

	header := new(bytes.Buffer)
	binary.Write(header, binary.LittleEndian, uint16(rec.channum))
	binary.Write(header, binary.LittleEndian, headerVersion)
	binary.Write(header, binary.LittleEndian, dataType)
	binary.Write(header, binary.LittleEndian, uint32(rec.presamples))
	binary.Write(header, binary.LittleEndian, uint32(len(rec.data)))
	binary.Write(header, binary.LittleEndian, rec.sampPeriod)
	binary.Write(header, binary.LittleEndian, float32(rec.peakValue)) // TODO: change to volts/arb
	nano := rec.trigTime.UnixNano()
	binary.Write(header, binary.LittleEndian, uint64(nano))
	binary.Write(header, binary.LittleEndian, uint64(rec.trigFrame))

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, rec.data)
	return [][]byte{header.Bytes(), buf.Bytes()}
}
