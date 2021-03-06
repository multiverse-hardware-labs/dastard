# DASTARD TCP PORTS

(Last revision: 14 June 2018.)

DASTARD uses a series of TCP ports for communications with its GUI control
clients (DASTARD-Commander) and to publish triggered pulse data for plotting
(Microscope).

These ports are numbered sequencially from a *base port* number. By default,
this number is **BASE=5500**. We might allow this number to be set at the
DASTARD command-line, but for now it's a constant.  The TCP ports are:

* **5500** (base+0): **Control**. JSON-RPC port for controlling DASTARD.  (JSON-RPC = "Remote Procedure Calls" specificed by JSON data format). Message format is defined by [json-rpc version 1.0](http://www.jsonrpc.org/specification_v1).
* **5501** (base+1): **Status**. ZMQ PUB port where DASTARD reports its status to all control GUIs.
* **5502** (base+2): **Pulses**. ZMQ PUB port where DASTARD puts all pulse records. Subscribe by 4-byte channel number. These are for Microscope to use, so it can plot data.
* **5503** (base+3): **Secondary records**. ZMQ PUB port, same as BASE+2, except that here we put only the secondary triggered records (i.e from a group trigger).
* **5504** (base+4): **Pulse Summaries**. ZMQ PUB port. Just has summary info and model fit coefficients.

### JSON-RPC commands (BASE+0)

Hmm. Should document these.

### Status messages (BASE+1)
Format is a text message-key (as a ZMQ frame) then a status block in JSON format. The messages are meant to be adequate to inform all Dastard control clients (the `dastard-commander` GUI, or others) everything they need to know about the Dastard internal state. Message keys include:

* **STATUS**: what data source or sources; idling or running; what is the data rate in bytes/sec (publish every 1-2 sec). What # of rows, columns, channels, and whether there are Error channels, too.
* **TRIGGER**: contains the trigger configuration (publish only when commander changes something). Possibly this can be a partial configuration, so for example if you change the trigger state for a subset of channels, the message contains their new state. But make one command exist that can request the full trigger state. Even then, we can be efficient by sending only 1 message per unique state, along with a list of the channel numbers that are in that specific state.
* **SIMPULSE**: contains the configuration of the Simulated Pulse data source.
* **TRIANGLE**: contains the configuration of the Triangle Wave data source.
* **LANCERO**: contains the configuration of the Lancero data source (e.g., which cards to use, fiber mask, etc.)

_The following are not implemented yet:_
* **RATE**: contains array-wide trigger rate and per-TES rates (publish regularly, every 1-2 sec)
* **WRITING**: contains output file information (type, filename, writing status stop/go/pause) (publish on change)
* **DECIMATION**: decimation state. This is universal to all channels.
* **MIXING**: TDM mixing state. Like TRIGGER, publish all values that match as a block of identically mixed channels.

### Primary and secondary pulse records (BASE+2 or 3)

This message format has a tentative definition, which we need to add here.

### Pulse summaries (BASE+4)

This message format has a tentative definition, which we need to add here.
