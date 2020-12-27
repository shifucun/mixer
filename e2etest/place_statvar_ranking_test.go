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

package e2etest

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"
	"runtime"
	"testing"

	pb "github.com/datacommonsorg/mixer/pkg/proto"
	"github.com/datacommonsorg/mixer/pkg/server"
	"github.com/google/go-cmp/cmp"
)

type RelatedChart struct {
	Scale     bool     `json:"scale,omitempty"`
	StatsVars []string `json:"statsVars,omitempty"` // Used only for golden files
}

type Chart struct {
	Title        string       `json:"title,omitempty"`
	StatsVars    []string     `json:"statsVars,omitempty"`
	Denominator  []string     `json:"denominator,omitempty"`
	RelatedChart RelatedChart `json:"relatedChart,omitempty"`
}

func readChartConfig() ([]Chart, error) {
	var config []Chart
	resp, err := http.Get("https://raw.githubusercontent.com/datacommonsorg/website/master/server/chart_config.json")
	if err != nil {
		return config, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(body, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func getMissingStatVarRanking(client pb.MixerClient, req *pb.GetLocationsRankingsRequest) ([]string, error) {
	ctx := context.Background()
	response, err := client.GetLocationsRankings(ctx, req)
	if err != nil {
		return nil, err
	}
	statVars := req.StatVarDcids
	if len(response.Payload) == 0 {
		return statVars, nil
	}
	var missing []string
	for _, sv := range statVars {
		if response.Payload[sv] == nil {
			missing = append(missing, sv)
		}
	}
	return missing, nil
}

func TestChartConfigRankings(t *testing.T) {
	client, err := setup(server.NewMemcache(map[string][]byte{}))
	if err != nil {
		t.Fatalf("Failed to set up mixer and client")
	}
	config, err := readChartConfig()
	if err != nil {
		t.Errorf("could not read config_statvars.txt")
		return
	}
	_, filename, _, _ := runtime.Caller(0)
	goldenPath := path.Join(path.Dir(filename), "../golden_response/staging/statvar_ranking")
	for _, c := range []struct {
		placeType   string
		parentPlace string
		goldenFile  string
	}{
		{
			"Country",
			"",
			"missing_Earth_country_rankings.json",
		},
		{
			"State",
			"country/USA",
			"missing_USA_state_rankings.json",
		},
		{
			"County",
			"geoId/06",
			"missing_USA_county_rankings.json",
		},
		{
			"City",
			"geoId/06",
			"missing_USA_city_rankings.json",
		},
	} {
		var missingRankings []Chart
		for _, chart := range config {
			var missingRanking Chart
			missingRanking.Title = chart.Title

			// Test main chart rankings
			req := &pb.GetLocationsRankingsRequest{
				PlaceType:    c.placeType,
				WithinPlace:  c.parentPlace,
				StatVarDcids: chart.StatsVars,
				IsPerCapita:  len(chart.Denominator) > 0,
			}
			missingStatVars, err := getMissingStatVarRanking(client, req)
			if err != nil {
				t.Errorf("Error fetching rankings for chart %s: %s", chart.Title, c.placeType)
				t.Errorf("%s", err.Error())
				continue
			}
			missingRanking.StatsVars = missingStatVars

			// Test related chart rankings
			if chart.RelatedChart.Scale {
				req.IsPerCapita = true
				missingStatVars, err := getMissingStatVarRanking(client, req)
				if err != nil {
					t.Errorf("Error fetching rankings for chart %s: %s", chart.Title, c.placeType)
					t.Errorf("%s", err.Error())
					continue
				}
				missingRanking.RelatedChart.Scale = true
				missingRanking.RelatedChart.StatsVars = missingStatVars
			}
			if missingRanking.StatsVars != nil {
				missingRankings = append(missingRankings, missingRanking)
			}

		}
		goldenFile := path.Join(goldenPath, c.goldenFile)
		if generateGolden {
			updateGolden(missingRankings, goldenFile)
			continue
		}

		var expected []Chart
		file, _ := ioutil.ReadFile(goldenFile)
		err = json.Unmarshal(file, &expected)
		if err != nil {
			t.Errorf("Can not Unmarshal golden file %s: %v", c.goldenFile, err)
			continue
		}
		if diff := cmp.Diff(&missingRankings, &expected); diff != "" {
			t.Errorf("payload got diff: %v", diff)
			continue
		}
	}
}
