// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"sort"

	pb "github.com/datacommonsorg/mixer/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ObsProp represents properties for a StatObservation.
type ObsProp struct {
	Mmethod string
	Operiod string
	Unit    string
	Sfactor string
}

func tokenFn(
	keyTokens map[string]*placeStatVar) func(rowKey string) (string, error) {
	return func(rowKey string) (string, error) {
		return keyTokens[rowKey].place + "^" + keyTokens[rowKey].statVar, nil
	}
}

// Filter a list of source series given the observation properties.
func filterSeries(in []*SourceSeries, prop *ObsProp) []*SourceSeries {
	result := []*SourceSeries{}
	for _, series := range in {
		if prop.Mmethod != "" && prop.Mmethod != series.MeasurementMethod {
			continue
		}
		if prop.Operiod != "" && prop.Operiod != series.ObservationPeriod {
			continue
		}
		if prop.Unit != "" && prop.Unit != series.Unit {
			continue
		}
		if prop.Sfactor != "" && prop.Sfactor != series.ScalingFactor {
			continue
		}
		result = append(result, series)
	}
	return result
}

func (in *ObsTimeSeries) filterAndRank(prop *ObsProp) {
	if in == nil {
		return
	}
	series := filterSeries(in.SourceSeries, prop)
	sort.Sort(byRank(series))
	if len(series) > 0 {
		in.Data = series[0].Val
		in.ProvenanceURL = series[0].ProvenanceURL
	}
	in.SourceSeries = nil
}

func getBestSeries(in *pb.ObsTimeSeries) *pb.Series {
	rawSeries := in.SourceSeries
	sort.Sort(SeriesByRank(rawSeries))
	if len(rawSeries) > 0 {
		return rawSeriesToSeries(rawSeries[0])
	}
	return nil
}

func rawSeriesToSeries(in *pb.SourceSeries) *pb.Series {
	result := &pb.Series{}
	result.Val = in.Val
	result.Metadata = &pb.StatMetadata{
		ImportName:        in.ImportName,
		ProvenanceUrl:     in.ProvenanceUrl,
		MeasurementMethod: in.MeasurementMethod,
		ObservationPeriod: in.ObservationPeriod,
		ScalingFactor:     in.ScalingFactor,
		Unit:              in.Unit,
	}
	return result
}

// getValueFromBestSource get the stat value from top ranked source series.
//
// When date is given, it get the value from the highest ranked source series
// that has the date.
//
// When date is not given, it get the latest value from the highest ranked
// source series.
func getValueFromBestSource(in *ObsTimeSeries, date string) (float64, error) {
	if in == nil {
		return 0, status.Error(codes.Internal, "Nil obs time series for getValueFromBestSource()")
	}
	sourceSeries := in.SourceSeries
	sort.Sort(byRank(sourceSeries))
	if date != "" {
		for _, series := range sourceSeries {
			if value, ok := series.Val[date]; ok {
				return value, nil
			}
		}
		return 0, status.Errorf(codes.NotFound, "No data found for date %s", date)
	}
	latestDate := ""
	var result float64
	for _, series := range sourceSeries {
		for date, value := range series.Val {
			if date > latestDate {
				latestDate = date
				result = value
			}
		}
	}
	if latestDate == "" {
		return 0, status.Errorf(codes.NotFound,
			"No stat data found for %s", in.PlaceDcid)
	}
	return result, nil
}

// getValueFromBestSourcePb get the stat value from ObsTimeSeries (protobuf version)
//
// When date is given, it get the value from the highest ranked source series
// that has the date.
//
// When date is not given, it get the latest value from all the source series.
// If two sources has the same latest date, the highest ranked source is preferred.
func getValueFromBestSourcePb(
	in *pb.ObsTimeSeries, date string) (*pb.PointStat, *pb.StatMetadata) {
	if in == nil {
		return nil, nil
	}
	sourceSeries := in.SourceSeries
	sort.Sort(SeriesByRank(sourceSeries))

	// Date is given, get the value from highest ranked source that has this date.
	if date != "" {
		for _, series := range sourceSeries {
			if value, ok := series.Val[date]; ok {
				return &pb.PointStat{
						Date:  date,
						Value: value,
						Metadata: &pb.StatMetadata{
							// Each ImportName should indicate a specific source. Now this is
							// not strictly true as the MeasurementMethod encodes source information
							// as well.
							// As the source is sorted deterministically, even when an ImportName
							// contains multiple sources, the top ranked one is picked. So
							// using ImportName as key still works.
							ImportName: series.ImportName,
						},
					},
					&pb.StatMetadata{
						ImportName:        series.ImportName,
						ProvenanceUrl:     series.ProvenanceUrl,
						MeasurementMethod: series.MeasurementMethod,
						ObservationPeriod: series.ObservationPeriod,
						ScalingFactor:     series.ScalingFactor,
						Unit:              series.Unit,
					}
			}
		}
		return nil, nil
	}
	// Date is not given, get the latest value from all sources.
	latestDate := ""
	var ps *pb.PointStat
	var meta *pb.StatMetadata
	for _, series := range sourceSeries {
		for date, value := range series.Val {
			if date > latestDate {
				latestDate = date
				ps = &pb.PointStat{
					Date:  date,
					Value: value,
					Metadata: &pb.StatMetadata{
						ImportName: series.ImportName,
					},
				}
				meta = &pb.StatMetadata{
					ImportName:        series.ImportName,
					ProvenanceUrl:     series.ProvenanceUrl,
					MeasurementMethod: series.MeasurementMethod,
					ObservationPeriod: series.ObservationPeriod,
					ScalingFactor:     series.ScalingFactor,
					Unit:              series.Unit,
				}
			}
		}
	}
	if latestDate == "" {
		return nil, nil
	}
	return ps, meta
}
