package main

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/AV-IO/DynAWSDNS/pkg/DynR53"
	"github.com/AV-IO/DynAWSDNS/pkg/DynSG"
	"github.com/AV-IO/DynAWSDNS/pkg/DynService"
	ipr "github.com/AV-IO/DynAWSDNS/pkg/IPRetriever"
	"github.com/alexflint/go-arg"
	"github.com/rs/zerolog/log"
)

type config struct {
	DelayHours int `json:"Delay"`
	R53confs   []struct {
		DomainName    string `json:"Domain"`
		SubDomainName string `json:"Subdomain"`
	} `json:"Route53_Configs"`
	SGconfs []struct {
		ID string `json:"ID"`
	} `json:"SecurityGroup_Configs"`
}

func main() {
	var args struct {
		configPath string `arg:"-c,--config,required" help:"configuration file path"`
	}
	arg.MustParse(&args)

	conf := load(args.configPath)

	delay := time.Duration(conf.DelayHours) * time.Hour
	// ticker distributes requests across the delay evenly
	ticker := time.NewTicker(
		time.Duration(delay.Seconds()/float64(len(conf.R53confs))) * time.Second,
	)

	var wg sync.WaitGroup
	for _, rconf := range conf.R53confs {
		wg.Add(1)
		go setR53Record(rconf.DomainName, rconf.SubDomainName, &delay, ipr.AWS, &wg)
		<-ticker.C
	}
	for _, sgconf := range conf.SGconfs {
		wg.Add(1)
		go setSGRecord(sgconf.ID, &delay, ipr.AWS, &wg)
		<-ticker.C
	}

	// Exits after all service handlers have failed
	wg.Wait()
}

func load(path string) (c config) {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not read configuration file")
	}

	if err = json.Unmarshal(content, &c); err != nil {
		log.Fatal().Err(err).Msg("Could not parse the configuration file")
	}
	return
}

func setR53Record(domainName, subDomainName string, delay *time.Duration, retriever ipr.IpRetriever, wg *sync.WaitGroup) {
	r53, err := DynR53.New("", "")
	if err != nil {
		log.Fatal().Msg("Could not initialize route53 handler")
	}

	loopUpdate(r53, delay, retriever)
	wg.Done()
}

func setSGRecord(id string, delay *time.Duration, retriever ipr.IpRetriever, wg *sync.WaitGroup) {
	sg, err := DynSG.New(id)
	if err != nil {
		log.Fatal().Msg("Could not initialize security group handler")
	}

	loopUpdate(sg, delay, retriever)
	wg.Done()
}

func loopUpdate(service DynService.Service, delay *time.Duration, retriever ipr.IpRetriever) {
	retry := 3

	for {
		// Wait
		time.Sleep(*delay)

		// Get IP, failing up to `retry` times
		ip, err := retriever()
		if err != nil {
			if retry == 0 {
				return
			}
			retry--
			continue
		}
		retry = 3

		// Update service
		err = service.Update(ip)
		if err != nil {
			return
		}
	}
}
