package main

import (
	"encoding/csv"
	"errors"
	"log"
	"math"
	"os"
	"strconv"
	"time"
)

const (
	ELECTRICITY_TYPE = 1
	GAS_TYPE         = 2
)

// Assumption 1: the listed price (0.18 kWh) is kWh per euro
// Assumption 2: the total price is rounded to 2 decimals
// Assumption 3: The CSV input is clean

const (
	DEFAULT_ELECTRICITY_PRICE = .18
	WEEKDAY_ELECTRICITY_PRICE = .20
	GAS_PRICE                 = .6
)

type Reading struct {
	energyType int
	usage      float64
	createdAt  time.Time
}

func main() {
	inputData := fetchCsv("test-input.csv")

	readings, totalPrices := organizeInputData(inputData)

	totalPrices = calculateTotalPrices(readings, totalPrices)

	generateOutputCsv(totalPrices)
}

func fetchCsv(filename string) [][]string {
	// open file
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	// remember to close the file at the end of the program
	defer f.Close()

	// read csv values using csv.Reader
	csvReader := csv.NewReader(f)
	inputData, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	return inputData
}

func generateOutputCsv(totalPrices map[int]float64) {
	csvFile, err := os.Create("output.csv")
	defer func(csvFile *os.File) {
		err := csvFile.Close()
		if err != nil {
			log.Fatal("Could not close output.csv")
		}
	}(csvFile)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	if err := writer.Write([]string{"id", "cost"}); err != nil {
		log.Fatalln("error writing header to csv file", err)
	}

	for id, price := range totalPrices {
		line := []string{strconv.Itoa(id), strconv.FormatFloat(roundFloat(price, 2), 'f', -1, 64)}

		if err := writer.Write(line); err != nil {
			log.Fatalln("error writing record to csv file", err)
		}
	}
}

func calculateTotalPrices(readings map[int][]Reading, totalPrices map[int]float64) map[int]float64 {
	for meterPointId, rs := range readings {
		for index, reading := range rs {
			pricePerHour := getPricePerHour(reading)

			kilowattHour, _ := convertWattHourToKilowattHour(reading.usage, reading.energyType)

			// To avoid undefined meterPointId error
			if index+1 < len(readings[meterPointId]) {
				nextReading := readings[meterPointId][index+1]
				nextKilowattHour, _ := convertWattHourToKilowattHour(nextReading.usage, nextReading.energyType)

				usage := nextKilowattHour - kilowattHour

				// ignore incorrect usages
				if math.Signbit(usage) || usage > 100 {
					continue
				}

				totalPrices[meterPointId] += usage * pricePerHour
			}
		}
	}

	return totalPrices
}

func getPricePerHour(reading Reading) float64 {
	pricePerHour := DEFAULT_ELECTRICITY_PRICE

	if reading.energyType == ELECTRICITY_TYPE {
		isWeekday := reading.createdAt.Weekday() != time.Sunday && reading.createdAt.Weekday() != time.Saturday

		if isWeekday && reading.createdAt.Hour() > 7 && reading.createdAt.Hour() < 23 {
			pricePerHour = WEEKDAY_ELECTRICITY_PRICE
		}
	} else if reading.energyType == GAS_TYPE {
		pricePerHour = GAS_PRICE
	}
	return pricePerHour
}

func organizeInputData(inputData [][]string) (map[int][]Reading, map[int]float64) {

	// [ meterPointId => [...], meterPointId => [...].. ]
	readings := make(map[int][]Reading)
	totalPrices := make(map[int]float64)

	// skip the header row
	for _, r := range inputData[1:] {
		timestamp, _ := strconv.ParseInt(r[3], 10, 64)

		createdAt := time.Unix(timestamp, 0)
		energyType, _ := strconv.Atoi(r[1])
		usage, _ := strconv.ParseFloat(r[2], 64)
		reading := Reading{energyType: energyType, usage: usage, createdAt: createdAt}

		meteringPointId, _ := strconv.Atoi(r[0])

		if _, ok := readings[meteringPointId]; !ok {
			readings[meteringPointId] = make([]Reading, 0)
		}

		readings[meteringPointId] = append(readings[meteringPointId], reading)

		totalPrices[meteringPointId] = 0.0
	}
	return readings, totalPrices
}

func convertWattHourToKilowattHour(reading float64, energyType int) (float64, error) {
	var err error
	kilowattHour := 0.0

	if energyType == ELECTRICITY_TYPE {
		kilowattHour = reading / 1000
	} else if energyType == GAS_TYPE {
		kilowattHour = reading * 9.769
	} else {
		err = errors.New("Found a line with unknown energy type!")
	}

	return kilowattHour, err
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
