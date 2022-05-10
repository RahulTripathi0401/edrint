package telemetry

import (
	"github.com/sharat910/edrint/common"
	"github.com/sharat910/edrint/events"

	"github.com/rs/zerolog/log"
)

type PacketPair struct {
	upPacket   common.Packet
	downPacket common.Packet
}

type FlowPairs struct {
	BaseFlowTelemetry
	upstream         []common.Packet
	downstream       []common.Packet
	buckets          map[float64][]PacketPair
	expected_latency float64
}

func (f *FlowPairs) Init() {
	f.buckets = make(map[float64][]PacketPair)
	f.expected_latency = 63.0
	// Set the buckets based on the expected latency
	// Going to have 10 buckets each of size expected_latency / 5 on each side
	bucket_size := f.expected_latency / 5
	for i := 1; i <= 10; i++ {
		bucket := bucket_size * float64(i)
		f.buckets[bucket] = []PacketPair{}
	}
}

func (f *FlowPairs) Pubs() []events.Topic {
	return []events.Topic{events.TELEMETRY_FLOWPAIRS}
}

func (f *FlowPairs) Name() string {
	return "flowpairs"
}

func (f *FlowPairs) OnFlowPacket(p common.Packet) {
	// Assuming these are all going to be upd packet paris
	if p.IsOutbound {
		for i := range f.upstream {
			var upStream = f.upstream[i]
			f.addPacketPair(&upStream, &p)
		}
		f.upstream = append(f.upstream, p)
	} else {
		for i := range f.downstream {
			var downPacket = f.downstream[i]
			f.addPacketPair(&p, &downPacket)
		}
		f.downstream = append(f.downstream, p)
	}
}

func (f *FlowPairs) addPacketPair(upPacket *common.Packet, downPacket *common.Packet) {
	var duration = float64(upPacket.Timestamp.Sub(downPacket.Timestamp).Milliseconds())
	pair := PacketPair{
		upPacket:   *upPacket,
		downPacket: *downPacket,
	}
	// If the RTT can fit into the bucket then add it to the bucket
	if val, ok := f.buckets[duration]; ok {
		val = append(val, pair)
		f.buckets[duration] = val
	}
}

func (f *FlowPairs) Teardown() {
	log.Debug().Str("telemetry", f.Name()).Msg("teardown")
	f.Publish(events.TELEMETRY_FLOWPAIRS, struct {
		bucket map[float64][]PacketPair
	}{
		f.buckets,
	})
}
