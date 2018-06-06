package dastard

import (
	"bytes"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/usnistgov/dastard/getbytes"
	"github.com/usnistgov/dastard/ljh"
	"github.com/usnistgov/dastard/off"

	czmq "github.com/zeromq/goczmq"
)

// DataPublisher contains many optional methods for publishing data, any methods that are non-nil will be used
// in each call to PublishData
type DataPublisher struct {
	PubRecordsChan   chan []*DataRecord
	PubSummariesChan chan []*DataRecord
	LJH22            *ljh.Writer
	LJH3             *ljh.Writer3
	OFF              *off.Writer
}

// SetOFF adds an OFF writer to dp, the .file attribute is nil, and will be instantiated upon next call to dp.WriteRecord
func (dp *DataPublisher) SetOFF(ChanNum int, Timebase float64,
	NumberOfRows int, NumberOfColumns int, NumberOfBases int, FileName string) {
	w := off.Writer{ChanNum: ChanNum,
		Timebase:        Timebase,
		NumberOfRows:    NumberOfRows,
		NumberOfColumns: NumberOfColumns,
		FileName:        FileName,
		NumberOfBases:   NumberOfBases, ProjectorsPlaceHolder: "projectors go here",
		BasisPlaceHolder: "basis goes here", NoiseWhitenerPlaceHolder: "noise whitener goes here"}
	dp.OFF = &w
}

// HasLJH22 returns true if LJH22 is non-nil, used to decide if writeint to LJH22 should occur
func (dp *DataPublisher) HasOFF() bool {
	return dp.OFF != nil
}
func (dp *DataPublisher) RemoveOFF() {
	dp.OFF.Close()
	dp.OFF = nil
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

// HasLJH3 returns true if LJH3 is non-nil, eg if writing to LJH3 is occuring
func (dp *DataPublisher) HasLJH3() bool {
	return dp.LJH3 != nil
}

// RemoveLJH3 closes existing LJH3 file and assign .LJH3=nil
func (dp *DataPublisher) RemoveLJH3() {
	if dp.LJH3 != nil {
		dp.LJH3.Close()
	}
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

// RemoveLJH22 closes existing LJH22 file and assign .LJH22=nil
func (dp *DataPublisher) RemoveLJH22() {
	if dp.LJH22 != nil {
		dp.LJH22.Close()
	}
	dp.LJH22 = nil
}

// HasPubRecords return true if publishing records on PortTrigs Pub is occuring
func (dp *DataPublisher) HasPubRecords() bool {
	return dp.PubRecordsChan != nil
}

// SetPubRecords starts publishing records with czmq over tcp at port=PortTrigs
func (dp *DataPublisher) SetPubRecords() {
	if PubRecordsChan == nil {
		configurePubRecordsSocket()
	}
	if dp.PubRecordsChan == nil {
		dp.PubRecordsChan = PubRecordsChan
	}
}

// RemovePubRecords stops publing records on PortTrigs
func (dp *DataPublisher) RemovePubRecords() {
	dp.PubRecordsChan = nil
}

// HasPubSummaries return true if publishing summaries on PortSummaries Pub is occuring
func (dp *DataPublisher) HasPubSummaries() bool {
	return dp.PubSummariesChan != nil
}

// SetPubSummaries starts publishing records with czmq over tcp at port=PortSummaries
func (dp *DataPublisher) SetPubSummaries() {
	if PubSummariesChan == nil {
		configurePubSummariesSocket()
	}
	if dp.PubSummariesChan == nil {
		dp.PubSummariesChan = PubSummariesChan
	}
}

// RemovePubSummaries stop publing summaries on PortSummaires
func (dp *DataPublisher) RemovePubSummaries() {
	dp.PubSummariesChan = nil
}

// PublishData looks at each member of DataPublisher, and if it is non-nil, publishes each record into that member
func (dp DataPublisher) PublishData(records []*DataRecord) error {
	if dp.HasPubRecords() {
		dp.PubRecordsChan <- records
	}
	if dp.HasPubSummaries() {
		dp.PubSummariesChan <- records
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
			dp.LJH22.WriteRecord(int64(record.trigFrame), int64(nano)/1000, rawTypeToUint16(record.data))
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
			dp.LJH3.WriteRecord(int32(record.presamples+1), int64(record.trigFrame), int64(nano)/1000, rawTypeToUint16(record.data))
		}
	}
	if dp.HasOFF() {
		for _, record := range records {
			if !dp.OFF.HeaderWritten { // MATTER doesn't create ljh files until at least one record exists, let us do the same
				// if the file doesn't exists yet, create it and write header
				err := dp.OFF.CreateFile()
				if err != nil {
					return err
				}
				dp.OFF.WriteHeader()
			}
			nano := record.trigTime.UnixNano()
			data := make([]uint16, len(record.data))
			for i, v := range record.data {
				data[i] = uint16(v)
			}
			modelCoefs := make([]float32, len(record.modelCoefs))
			for i, v := range record.modelCoefs {
				modelCoefs[i] = float32(v)
			}
			dp.OFF.WriteRecord(record.presamples+1, int64(record.trigFrame), int64(nano)/1000, modelCoefs)
		}
	}
	return nil
}

// messageSummaries makes a message with the following format for publishing on portTrigs
// Structure of the message header is defined in BINARY_FORMATS.md
// uint16: channel number
// uint8: header version number
// uint32: bits: Presamples
// uint32: length of record
// float32: pretrigMean
// float32: peakValue
// float32: pulseRMS
// float32: residualStdDev
// uint64: UnixNano trigTime
// uint64: trigFrame
//  end of first message packet
//  modelCoefs, each coef is float32, length can vary
func messageSummaries(rec *DataRecord) [][]byte {
	const headerVersion = uint8(0)

	header := new(bytes.Buffer)
	header.Write(getbytes.FromUint16(uint16(rec.channum)))
	header.Write(getbytes.FromUint8(headerVersion))
	header.Write(getbytes.FromUint32(uint32(rec.presamples)))
	header.Write(getbytes.FromUint32(uint32(len(rec.data))))
	header.Write(getbytes.FromFloat32(float32(rec.pretrigMean)))
	header.Write(getbytes.FromFloat32(float32(rec.peakValue)))
	header.Write(getbytes.FromFloat32(float32(rec.pulseRMS)))
	header.Write(getbytes.FromFloat32(float32(rec.pulseAverage)))
	header.Write(getbytes.FromFloat32(float32(rec.residualStdDev)))
	nano := rec.trigTime.UnixNano()
	header.Write(getbytes.FromInt64(nano))
	header.Write(getbytes.FromInt64(int64(rec.trigFrame)))

	return [][]byte{header.Bytes(), getbytes.FromSliceFloat64(rec.modelCoefs)}
}

// messageRecords makes a message with the following format for publishing on portTrigs
// Structure of the message header is defined in BINARY_FORMATS.md
// uint16: channel number
// uint8: header version number
// uint8: code for data type (0-1 = 8 bits; 2-3 = 16; 4-5 = 32; 6-7 = 64; odd=uint; even=int)
// uint32: # of pre-trigger samples
// uint32: # of samples, total
// float32: sample period, in seconds (float)
// float32: volts per arb conversion (float) (not implemented yet)
// uint64: trigger time, in ns since epoch 1970
// uint64: trigger frame #
// end of first message packet
// data, each sample is uint16, length can vary (but for now is equal to # of samples which should be removed because it is redundant)
func messageRecords(rec *DataRecord) [][]byte {

	const headerVersion = uint8(0)
	const dataType = uint8(3)
	header := new(bytes.Buffer)
	header.Write(getbytes.FromUint16(uint16(rec.channum)))
	header.Write(getbytes.FromUint8(headerVersion))
	header.Write(getbytes.FromUint8(dataType))
	header.Write(getbytes.FromUint32(uint32(rec.presamples)))
	header.Write(getbytes.FromUint32(uint32(len(rec.data))))
	header.Write(getbytes.FromFloat32(rec.sampPeriod))
	header.Write(getbytes.FromFloat32(0)) // todo make this meaningful, though doesn't seem like its needed with every pulse
	nano := rec.trigTime.UnixNano()
	header.Write(getbytes.FromInt64(nano))
	header.Write(getbytes.FromUint64(uint64(rec.trigFrame)))

	data := rawTypeToBytes(rec.data)
	return [][]byte{header.Bytes(), data}
}

// PubRecordsChan is used to enable multiple different DataPublishers to publish on the same zmq pub socket
var PubRecordsChan chan []*DataRecord

// PubSummariesChan is used to enable multiple different DataPublishers to publish on the same zmq pub socket
var PubSummariesChan chan []*DataRecord

// configurePubSocket should be run exactly one time
// it initializes PubFeederChan and launches a goroutine (with no way to stop it at the moment)
// that reads from PubFeederChan and publishes records on a ZMQ PUB socket at port PortTrigs
// this way even if go routines in different threads want to publish records, they all use the same
// zmq port
// *** This looks like it could be replaced by PubChanneler, but tests show terrible perfoance with Channeler ***
func configurePubRecordsSocket() error {
	if PubRecordsChan != nil {
		panic("run configurePubSocket only one time")
	}
	const publishChannelDepth = 500
	// I think this could be a PubChanneler
	PubRecordsChan = make(chan []*DataRecord, publishChannelDepth)
	hostname := fmt.Sprintf("tcp://*:%d", PortTrigs)
	pubSocket, err := czmq.NewPub(hostname)
	if err != nil {
		return err
	}
	go func() {
		defer pubSocket.Destroy()
		for {
			records := <-PubRecordsChan
			for _, record := range records {
				message := messageRecords(record)
				err := pubSocket.SendMessage(message)
				if err != nil {
					panic("zmq send error")
				}
			}
		}
	}()
	return nil
}
func configurePubSummariesSocket() error {
	if PubSummariesChan != nil {
		panic("run configurePubSocket only one time")
	}
	const publishChannelDepth = 500
	// I think this could be a PubChanneler
	PubSummariesChan = make(chan []*DataRecord, publishChannelDepth)
	hostname := fmt.Sprintf("tcp://*:%d", PortSummaries)
	pubSocket, err := czmq.NewPub(hostname)
	if err != nil {
		return err
	}
	go func() {
		defer pubSocket.Destroy()
		for {
			records := <-PubSummariesChan
			for _, record := range records {
				message := messageSummaries(record)
				err := pubSocket.SendMessage(message)
				if err != nil {
					panic("zmq send error")
				}
			}
		}
	}()
	return nil
}

// rawTypeToBytes convert a []RawType to []byte using unsafe
// see https://stackoverflow.com/questions/11924196/convert-between-slices-of-different-types?utm_medium=organic&utm_source=google_rich_qa&utm_campaign=google_rich_qa
func rawTypeToBytes(d []RawType) []byte {
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&d))
	header.Cap *= 2 // byte takes up half the space of RawType
	header.Len *= 2
	data := *(*[]byte)(unsafe.Pointer(&header))
	return data
}

// rawTypeToUint16convert a []RawType to []uint16 using unsafe
// see https://stackoverflow.com/questions/11924196/convert-between-slices-of-different-types?utm_medium=organic&utm_source=google_rich_qa&utm_campaign=google_rich_qa
func rawTypeToUint16(d []RawType) []uint16 {
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&d))
	data := *(*[]uint16)(unsafe.Pointer(&header))
	return data
}
