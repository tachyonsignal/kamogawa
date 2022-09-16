package main

import (
	"encoding/csv"
	"github.com/tjarratt/babble"
	"log"
	"math/rand"
	"os"
	"strconv"
)

func main() {
	babbler := babble.NewBabbler()
	babbler.Count = 1

	iCsvFile, err := os.Create("test/instances.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	iCsvWriter := csv.NewWriter(iCsvFile)

	pCsvFile, err := os.Create("test/projects.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	pCsvWriter := csv.NewWriter(pCsvFile)

	// int, str, str, str, str
	iCsvWriter.Write([]string{"id", "name", "status", "project_id", "zone"})

	// int, int, str, str, str, 1337gamer@gmail.com
	pCsvWriter.Write([]string{"id", "project_number", "project_id", "name", "life_cycle_state", "email"})

	instanceId := 0
	for numProjects := 0; numProjects < 500; numProjects++ {
		id := strconv.Itoa(numProjects)
		project_number := strconv.Itoa(rand.Intn(100))
		project_id := babbler.Babble()
		name := babbler.Babble()
		life_cycle_state := "ACTIVE"
		email := "1337gamer@gmail.com"
		pCsvWriter.Write([]string{id, project_number, project_id, name, life_cycle_state, email})

		for numZone := 0; numZone < 5; numZone++ {
			zone := babbler.Babble()
			for numInstances := 0; numInstances < 10000; numInstances++ {
				instanceId++
				id := strconv.Itoa(instanceId)
				name := babbler.Babble()
				status := "ACTIVE"
				iCsvWriter.Write([]string{id, name, status, project_id, zone})
			}
		}
	}

	iCsvWriter.Flush()
	iCsvFile.Close()
	pCsvWriter.Flush()
	pCsvFile.Close()
}