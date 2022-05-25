package main

import (
	"fmt"

	"github.com/sharat910/edrint/common"
	"github.com/sharat910/edrint/events"
	"github.com/sharat910/edrint/telemetry"

	"github.com/rs/zerolog/log"
)

type PacketPair struct {
	upPacket   common.Packet
	downPacket common.Packet
}

type FlowPairs struct {
	telemetry.BaseFlowTelemetry
	upStream 		 []common.Packet
	buckets          map[uint64][]PacketPair
	expected_latency uint64
	bucketSize		 uint64
}

func NewFlowPairs(expectedLatency uint64, bucketSize uint64, numberOfBuckets uint64) telemetry.TeleGen {
	return func() telemetry.Telemetry {
		var f FlowPairs
		f.buckets = make(map[uint64][]PacketPair)
		f.expected_latency = expectedLatency - (expectedLatency % 5)
		// Set the buckets based on the expected latency
		// Going to have 10 buckets each of size expected_latency / 5 on each side
		f.bucketSize = bucketSize
		startBucket := f.expected_latency -  uint64(numberOfBuckets / 2) * bucketSize 
		endBucket := f.expected_latency +  uint64(numberOfBuckets / 2) * bucketSize 
		//f.buckets[5] = []PacketPair{}
		fmt.Println(startBucket)
		for i := startBucket; i <= endBucket; i+= bucketSize {
			f.buckets[i] = []PacketPair{}
		}
		return &f
	}
}

func (f *FlowPairs) Pubs() []events.Topic {
	return []events.Topic{"telemetry.flowpairs"}
}

func (f *FlowPairs) Name() string {
	return "flowpairs"
}

func (f *FlowPairs) OnFlowPacket(p common.Packet) {
	// Assuming these are all going to be upd packet paris
	if p.IsOutbound {
		f.upStream = append(f.upStream, p)		
		return
	}

	for i := range f.upStream {
		// For each downstream packet calculate the RTTtime
		upStreamPacket := f.upStream[i]
		time := uint64(p.Timestamp.Sub(upStreamPacket.Timestamp).Milliseconds())
		// If the RTT time is more than a second ignore it or the time is negative
		if (time > 1000 ) {
			continue
		}
		pair := PacketPair{
			upPacket: upStreamPacket,
			downPacket: p,
		}
		f.addToBucket(&pair, time)
	}
}

func (f *FlowPairs) addToBucket(pair *PacketPair, duration uint64) {
	duration = duration - (duration % f.bucketSize)
	if val, ok := f.buckets[duration]; ok {
		val = append(val, *pair)
		f.buckets[duration] = val
	}
}

func (f *FlowPairs) Teardown() {
	log.Debug().Str("telemetry", f.Name()).Msg("teardown")
	f.Publish("telemetry.flowpairs", struct {
		bucket map[uint64][]PacketPair
	}{
		f.buckets,
	})
}
