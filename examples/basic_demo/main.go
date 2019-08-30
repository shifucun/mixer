// Copyright 2019 Google LLC
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

package main

import (
	"context"
	"flag"
	"log"

	pb "github.com/datacommonsorg/mixer/proto"
	"google.golang.org/grpc"
)

var (
	addr = flag.String("addr", "127.0.0.1:12345", "Address of grpc server.")
)

func main() {
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewMixerClient(conn)

	ctx := context.Background()

	// Contact the server and print out its response.
	qStrs := []string{
		`
		BASE <http://schema.org/>
		SELECT ?MeanTemp
		WHERE {
			?o typeOf WeatherObservation .
			?o measuredProperty temperature .
			?o meanValue ?MeanTemp .
			?o observationDate "2018-01" .
			?o observedNode ?place .
			?place dcid geoId/4261000
		}
		LIMIT 10`,
		`
		BASE <http://schema.org/>
		SELECT  ?Unemployment
		WHERE {
		  ?pop typeOf StatisticalPopulation .
		  ?o typeOf Observation .
		  ?pop dcid ("dc/p/qep2q2lcc3rcc" "dc/p/gmw3cn8tmsnth" "dc/p/92cxc027krdcd") .
		  ?o observedNode ?pop .
		  ?o measuredValue ?Unemployment
		}
		ORDER BY DESC(?Unemployment)
		LIMIT 10`,
		`
		SELECT ?a
		WHERE { ?a typeof USC_RaceCodeEnum}`,
	}
	for _, str := range qStrs {
		r, err := c.Query(ctx, &pb.QueryRequest{Sparql: str})
		if err != nil {
			log.Fatalf("could not Query: %v", err)
		}
		log.Printf("Query: %v", r.GetRows())
	}
}
