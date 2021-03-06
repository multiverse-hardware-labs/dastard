# Binary Serialization Formats for DASTARD Data Structures

## Binary Format for Triggered Data records

12/21/2017

Triggered records are published on a ZMQ PUB socket. Primary triggers appear on
port *BASE*+2, and secondary (cross-talk) triggers appear on port *BASE*+3.

### Packet Version 0

Dated 4/19/2018. Packets consist of a 2-frame ZMQ message. The first frame contains
the header, which is 36 bytes long. The second frame is the raw record data, which
is of variable length and packed in little-endian byte order.
The header also contains little-endian values:

* Byte 0 (2 bytes): channel number
* Byte 2 (1 byte):  header version number (0 in this version)
* Byte 3 (1 byte):  data type code (see below)
* Byte 4 (4 bytes): samples before trigger
* Byte 8 (4 bytes): samples in record
* Byte 12 (4 bytes): sample period in seconds (float)
* Byte 16 (4 bytes): volts per arb (float)
* Byte 20 (8 bytes): trigger time (nanoseconds since 1 Jan 1970)
* Byte 28 (8 bytes): trigger frame index

Because the channel number makes up the first 2 bytes, ZMQ subscriber sockets can
subscribe selectively to only certain channels.

Data type code: so far, only uint16 and int16 are allowed.

* 0 = int8
* 1 = uint8
* 2 = int16
* 3 = uint16
* 4 = int32
* 5 = uint32
* 6 = int64
* 7 = uint64
