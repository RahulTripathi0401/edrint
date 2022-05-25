package main

import (
	"fmt"
	"path/filepath"


	"github.com/sharat910/edrint/events"

	"github.com/rs/zerolog/log"
	"github.com/sharat910/edrint"
	"github.com/sharat910/edrint/processor"
	"github.com/spf13/viper"
)

func main() {
	SetupConfig()
	edrint.SetupLogging(viper.GetString("log.level"))
	manager := edrint.New()
	manager.RegisterProc(processor.NewFlowProcessor(2))
	rules := GetHeaderClassificationRules()
	manager.RegisterProc(processor.NewHeaderClassifer(rules))
	//manager.RegisterProc(processor.NewSNIParser())
	//manager.RegisterProc(processor.NewSNIClassifier(viper.GetStringMapString("sniclassifier.classes")))
	teleManager := processor.NewTelemetryManager()
	teleManager.AddTFToClass("fortnite", NewFlowPairs(163, 5, 20))

	manager.RegisterProc(teleManager)

	packetPath := viper.GetString("packets.source")
	//dumpPath := fmt.Sprintf("./files/dumps/%s.json.log", filepath.Base(packetPath))
	dumpPath := fmt.Sprintf("%s/telemetry/%s.json.log",
		filepath.Dir(filepath.Dir(packetPath)), filepath.Base(packetPath))
	manager.RegisterProc(processor.NewDumper(dumpPath,
		[]events.Topic{
			"telemetry.flowpairs",
		}, false))

	err := manager.InitProcessors()
	if err != nil {
		log.Fatal().Err(err).Msg("init error")
	}
	err = manager.Run(edrint.ParserConfig{
		CapMode:    edrint.PCAPFILE,
		CapSource:  viper.GetString("packets.source"),
		DirMode:    edrint.CLIENT_IP,
		DirMatches: viper.GetStringSlice("packets.direction.client_ips"),
		BPF:        viper.GetString("packets.bpf"),
		MaxPackets: viper.GetInt("packets.maxcount"),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("some error occurred")
	}
}

func GetHeaderClassificationRules() map[string]map[string]string {
	rules := make(map[string]map[string]string)
	config := viper.GetStringMap(fmt.Sprintf("processors.header_classifier.classes"))
	for class := range config {
		rules[class] = viper.GetStringMapString(fmt.Sprintf("processors.header_classifier.classes.%s", class))
	}
	return rules
}
