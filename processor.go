package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

type Processor struct {
	LogFileReader    *LogFileReader
	MetricsSubmitter *MetricsSubmitter
}

func NewProcessor() (*Processor, error) {
	var processor Processor
	config, err := NewConfigFromFile(os.Getenv("CONFIG_PATH"))
	if err != nil {
		return nil, err
	}
	processor.LogFileReader = NewLogFileReader(config.PathTransformingRules, config.TargetPaths)
	processor.MetricsSubmitter = NewMetricsSubmitter(config.RequestCountMetricName, config.TargetProcessingTimeMetricName, config.CustomTags)
	return &processor, nil
}

func (p *Processor) ProcessLogfile(r io.Reader, s3ObjectKey string) error {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer zr.Close()

	fmt.Println("start reading log file")

	metricsMap, err := p.LogFileReader.Read(zr)
	if err != nil {
		return err
	}

	fmt.Println("finish reading log file")

	fmt.Println("start submitting metrics")

	err = p.MetricsSubmitter.Submit(metricsMap, s3ObjectKey)
	if err != nil {
		return err
	}

	fmt.Println("finish submitting metrics")
	return nil
}
