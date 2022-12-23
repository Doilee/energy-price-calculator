package main

import (
	"encoding/csv"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestOutput(t *testing.T) {
	main()

	assert.FileExistsf(t, "output.csv", "Unable to find output file")

	data := fetchCsv("output.csv")

	assert.Equal(t, [][]string{{"id", "cost"}, {"1", "0.02"}}, data)
}

func TestTotalPriceIsExpected(t *testing.T) {
	inputData := fetchCsv("example-input.csv")

	readings, totalPrices := organizeInputData(inputData)

	totalPrices = calculateTotalPrices(readings, totalPrices)

	assert.Equal(t, map[int]float64{1: 0.02160000000000082}, totalPrices)
}

func TestRoundingPriceToTwoDecimals(t *testing.T) {
	toRound := roundFloat(0.02160000000000082, 2)

	assert.Equal(t, 0.02, toRound)
}

// takes about ~6s on my laptop
func TestPerformance(t *testing.T) {
	CreateCsv("test-input.csv", 10000000) // ~260MB

	start := time.Now().Unix()

	generateOutputCsv(calculateTotalPrices(organizeInputData(fetchCsv("test-input.csv"))))

	end := time.Now().Unix()

	assert.LessOrEqual(t, end-start, int64(10))
}

func CreateCsv(filename string, lineCount int) {
	csvFile, err := os.Create(filename)
	defer func(csvFile *os.File) {
		err := csvFile.Close()
		if err != nil {
			log.Fatalf("Could not close %s", filename)
		}
	}(csvFile)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	if err := writer.Write([]string{"metering_point_id", "type", "reading", "created_at"}); err != nil {
		log.Fatalln("error writing header to csv file", err)
	}

	energyUsage := 0
	timestamp := 1415963700

	for i := 0; i < lineCount; i++ {
		meterpointId := strconv.Itoa(i % 10)
		energytype := strconv.Itoa(rand.Intn(2))

		line := []string{meterpointId, energytype, strconv.Itoa(energyUsage), strconv.Itoa(timestamp)}

		if err := writer.Write(line); err != nil {
			log.Fatalln("error writing record to csv file", err)
		}

		energyUsage += 100
		timestamp += 900
	}
}
