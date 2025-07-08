package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"math"
)

var rpmOpciones = []int{800, 1500, 2500, 3500, 4500, 5500, 6500}
var orden = []int{1, 3, 4, 2}

type Sensores struct {
	TPS float64 
	RPM float64
	O2 float64
	Mu sync.Mutex
}

type Inyector struct {
	ID int
	Accion chan float64
  Log []float64
	Mu sync.Mutex
}

	func main()  {
	mapa, err := cargarMapaInyeccion("mapa_inyeccion.csv")
	if err != nil {
		panic(err)
	}
	fmt.Println(mapa[1][3500])
  sensores := &Sensores{}
  go sensores.simularTPS_1()
	go sensores.simularRPMporTPS()
  
  for {
  time.Sleep(200 * time.Millisecond)
	sensores.Mu.Lock()
	fmt.Println(sensores.TPS, "TPS")
  fmt.Println(discretizar(sensores.RPM, rpmOpciones), "RPM")
	sensores.Mu.Unlock()
  }	
	select {}
	}



func (s *Sensores)simularTPS_1()  {
	for {
	for i := 0.0; i <= 100; i+=5 {
	 s.Mu.Lock()
	 s.TPS = i
	 s.Mu.Unlock()
	 time.Sleep(200 * time.Millisecond)
	}
	for i := 100.0; i >= 0; i-=5 {	
	 s.Mu.Lock()
	 s.TPS = i
	 s.Mu.Unlock()
	 time.Sleep(200 * time.Millisecond)
	}
	}
}

func (s Sensores)leerTPS()  {	
}
func (s *Sensores) simularRPMporTPS() {
for {
		s.Mu.Lock()
		tps := s.TPS
		s.RPM = 800 + (tps / 100.0 * 5700) // de 800 a 6500 según TPS
		s.Mu.Unlock()
		time.Sleep(100 * time.Millisecond)
	}
	
}
func cargarMapaInyeccion(path string) (map[int]map[int]float64, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	rpmList := []int{}
	for _, h := range headers[1:] {
		rpmStr := strings.TrimPrefix(h, "RPM")
		rpm, err := strconv.Atoi(rpmStr)
		if err != nil {
			return nil, fmt.Errorf("RPM inválido en header: %s", h)
		}
		rpmList = append(rpmList, rpm)
	}

	mapa := make(map[int]map[int]float64)

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		tps, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, fmt.Errorf("TPS inválido: %s", record[0])
		}

		mapa[tps] = make(map[int]float64)
		for i, valStr := range record[1:] {
			val, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				return nil, fmt.Errorf("valor inválido [%s] en fila TPS %d, RPM %d", valStr, tps, rpmList[i])
			}
			mapa[tps][rpmList[i]] = val
		}
	}

	return mapa, nil
}

func discretizar(valor float64, opciones []int) int {
	minDiff := math.MaxFloat64
	masCercano := opciones[0]
	for _, op := range opciones {
		diff := math.Abs(valor - float64(op))
		if diff < minDiff {
			minDiff = diff
			masCercano = op
		}
	}
	return masCercano
}

