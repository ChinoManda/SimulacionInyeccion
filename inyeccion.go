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

var rpmList = []int{}
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
  Log int
	Mu sync.Mutex
}

type ECU struct {
	Sensores *Sensores
	Inyectores []*Inyector
mapa map[int]map[int]float64 //[TPS][RPM]MS 
	OrdenInyeccion []int
}

	func main()  {
	mapa, err := cargarMapaInyeccion("mapa_inyeccion.csv")
	if err != nil {
		panic(err)
	}
	  sensores := &Sensores{}
		inyectores :=  []*Inyector{}
		for i := 1; i <= 4; i++ {
			iny := &Inyector{
				ID: i,
				Accion: make(chan float64),
			}
			inyectores = append(inyectores, iny)
		  go iny.ejecutar()
		}
		Bosch := ECU {sensores, inyectores, mapa, orden}
  go sensores.simularTPS_1()
	go sensores.simularRPMporTPS()
  go Bosch.run()
	select {}	
	}
func (s *Sensores) testing(){
	var j int
	for {
		
	  s.Mu.Lock()
		rpm := int(s.RPM)
		fmt.Println(rpm)
    s.Mu.Unlock()
		
 j = buscarRPM(rpmList, rpm)
 fmt.Println("-->", j)
 time.Sleep(100 * time.Millisecond)
 }
}
func (e *ECU) run(){
 ciclo := 1
	for {
		start := time.Now()
		e.Sensores.Mu.Lock()
		tps := int(e.Sensores.TPS)
		rpm := buscarRPM(rpmList, int(e.Sensores.RPM))
    e.Sensores.Mu.Unlock()

		delay := calcularDelay(float64(rpm))
		tiempoEsperado := delay * 4

    for _, id := range e.OrdenInyeccion {
			tiempo := e.mapa[tps][rpm]
      mostrarEstadoInyectores(id, ciclo, tps, rpm, e.Inyectores)
			go func(iny *Inyector, t float64)  {
				iny.Accion <- t
			}(e.Inyectores[id-1], tiempo)
      time.Sleep(delay)
		}
				duracion := time.Since(start)
		fmt.Printf("\n⏱️ Duración total del ciclo: %.3f ms | Esperado: %.3f ms | Δ: %.3f ms\n",
			float64(duracion.Microseconds())/1000,
			float64(tiempoEsperado.Microseconds())/1000,
			float64(duracion.Microseconds()-tiempoEsperado.Microseconds())/1000)

	time.Sleep(100 * time.Millisecond) // para ver la diferencia entre ciclos
		ciclo++
	}

	
}

func (i *Inyector) ejecutar(){
for tiempo := range i.Accion {
  fmt.Println("Inyectando")
	fmt.Println("Tiempo de Inyección:",tiempo, "ms")
	i.Mu.Lock()
	i.Log++
	i.Mu.Unlock()
	}
}

func (s *Sensores)simularTPS_1()  {
	for {
	for i := 0.0; i <= 100; i+=5 {
	 s.Mu.Lock()
	 s.TPS = i
	 s.Mu.Unlock()
	 time.Sleep(500 * time.Millisecond)
	}
	for i := 100.0; i >= 0; i-=5 {	
	 s.Mu.Lock()
	 s.TPS = i
	 s.Mu.Unlock()
	 time.Sleep(500 * time.Millisecond)
	}
	}
}
func mostrarEstadoInyectores(activo int, ciclo int, tps int, rpm int , inyectores []*Inyector) {
	// Limpiar pantalla
	fmt.Print("\033[H\033[2J") // ANSI escape: clear screen

	fmt.Printf("====== CICLO %d ======\n", ciclo)
	fmt.Printf("🚗 RPM: %d | TPS: %d\n\n", rpm, tps)
	fmt.Println("💉 Estado de Inyectores:")
  
	for _, iny := range inyectores {
		iny.Mu.Lock()
		simbolo := "○"
		if iny.ID == activo {
			simbolo = "●"
		}
		fmt.Printf("[Inyector %d] %s  (%d inyecciones)\n", iny.ID, simbolo, iny.Log)
		iny.Mu.Unlock()
	}
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

func buscarRPM(rpmList []int, rpm int) int {
	lo := 0
	hi := len(rpmList)-1

  for lo<=hi {
	mid := (hi + lo) / 2

	if rpmList[lo] == rpm {
		return rpmList[lo]
	}
	if rpmList[mid] == rpm {
		return rpmList[mid]
	}

	if rpmList[hi] == rpm {
		return rpmList[hi]
	}

	if rpmList[mid] < rpm {
    lo = mid+1
	} 
	if rpmList[mid] > rpm {
		hi = mid -1
	}
 } //si lo y hi se dan vuelta, ahi se rompe el for, si cae justo en uno se rompe antes, si termina el for hay q interpolar

 if lo >= len(rpmList){ //si lo se pasa para adelante es el mayor
	 return rpmList[len(rpmList)-1]
 }

 if hi < 0 { //si hi queda negativo es el mas chico (0)
	 return rpmList[0]
 }

 if abs(rpm - rpmList[lo]) < abs(rpm -rpmList[hi]){
	 return rpmList[lo]
 } else {
	 return rpmList[hi]
 }
  
}

func abs(a int) int {
    if a >= 0 {
        return a
    }
    return -a
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

func calcularDelay(rpm float64) time.Duration {
	segundos := (60.0 / rpm) / 2.0
	return time.Duration(segundos * float64(time.Second))
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

