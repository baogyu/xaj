package Watershed

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Watershed struct {
	Name                   string
	Area                   float64
	NumRainfallStation     int
	NumEvaporationStation  int
	NumSubWatershed        int
	AreaSubWatershed       []float64
	RateRainfallStation    [][]float64
	RateEvaporationStation [][]float64
	NameRainfallStation    []string
	NameEvaporationStation []string
	P                      [][]float64
	EM                     [][]float64
}

func (ws *Watershed) ReadFromFile(strPath string) error {
	// 打开文件
	file, err := os.Open(strPath + "watershed.txt")
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	ws.Name = lines[0]
	ws.Area, _ = strconv.ParseFloat(lines[1], 64)
	ws.NumRainfallStation, _ = strconv.Atoi(lines[2])
	ws.NumEvaporationStation, _ = strconv.Atoi(lines[3])
	ws.NumSubWatershed, _ = strconv.Atoi(lines[4])

	// 读取子流域面积
	ws.AreaSubWatershed = make([]float64, ws.NumSubWatershed)
	AreaSubWatershed := strings.Split(lines[5], "\t")
	for i := 0; i < len(AreaSubWatershed); i++ {
		ws.AreaSubWatershed[i], _ = strconv.ParseFloat(AreaSubWatershed[i], 64)
	}

	// 读取雨量站权重
	ws.RateRainfallStation = make([][]float64, ws.NumSubWatershed)
	for i := 0; i < ws.NumSubWatershed; i++ {
		rainfallStationLine := strings.Split(lines[6+i], "\t")
		ws.RateRainfallStation[i] = make([]float64, ws.NumRainfallStation)
		for j := 0; j < ws.NumRainfallStation; j++ {
			ws.RateRainfallStation[i][j], _ = strconv.ParseFloat(rainfallStationLine[j], 64)
		}
	}

	// 读取蒸发站权重
	ws.RateEvaporationStation = make([][]float64, ws.NumSubWatershed)
	for i := 0; i < ws.NumSubWatershed; i++ {
		evaporationStationLine := strings.Split(lines[6+ws.NumSubWatershed+i], "\t") // 计算正确的起始行位置
		ws.RateEvaporationStation[i] = make([]float64, ws.NumEvaporationStation)
		for j := 0; j < ws.NumEvaporationStation; j++ {
			ws.RateEvaporationStation[i][j], _ = strconv.ParseFloat(evaporationStationLine[j], 64)
		}
	}

	// 读取雨量站名称
	ws.NameRainfallStation = make([]string, ws.NumRainfallStation)
	stationName := strings.Split(lines[6+ws.NumSubWatershed*2], " ")
	for i := 0; i < len(stationName); i++ {
		ws.NameRainfallStation[i] = stationName[i] // 计算雨量站名称的起始行位置
	}

	// 读取蒸发站名称
	ws.NameEvaporationStation = make([]string, ws.NumEvaporationStation)
	EvaporationName := strings.Split(lines[6+ws.NumSubWatershed*2+1], "\t")
	for i := 0; i < len(EvaporationName); i++ {
		ws.NameEvaporationStation[i] = EvaporationName[i] // 计算蒸发站名称的起始行位置
	}

	return nil
}

func (w *Watershed) SetValues(name string, area float64, numRainfallStation, numEvaporationStation, numSubWatershed int, areaSubWatershed []float64, rateRainfallStation, rateEvaporationStation [][]float64, nameRainfallStation, nameEvaporationStation []string, P, EM [][]float64) {
	w.Name = name
	w.Area = area
	w.NumRainfallStation = numRainfallStation
	w.NumEvaporationStation = numEvaporationStation
	w.NumSubWatershed = numSubWatershed
	w.AreaSubWatershed = areaSubWatershed
	w.RateRainfallStation = rateRainfallStation
	w.RateEvaporationStation = rateEvaporationStation
	w.NameRainfallStation = nameRainfallStation
	w.NameEvaporationStation = nameEvaporationStation
	w.P = P
	w.EM = EM
}

func (w *Watershed) Calculate(io *IO) {
	nrows := io.Nrows          // 记录条数，即时段数
	ncols := w.NumSubWatershed // 单元流域个数

	// 计算各单元流域逐时段降雨量，mm
	w.P = make([][]float64, nrows)
	for r := 0; r < nrows; r++ {
		w.P[r] = make([]float64, ncols)
		for c := 0; c < ncols; c++ {
			for i := 0; i < w.NumRainfallStation; i++ {
				w.P[r][c] += io.Mp[r][i] * w.RateRainfallStation[c][i] // 按比例计算单元流域降雨量
			}
		}
	}

	// 计算各单元流域逐时段水面蒸发量，mm
	w.EM = make([][]float64, nrows)
	for r := 0; r < nrows; r++ {
		w.EM[r] = make([]float64, ncols)
		for c := 0; c < ncols; c++ {
			for i := 0; i < w.NumEvaporationStation; i++ {
				w.EM[r][c] += io.MEM[r][i] * w.RateEvaporationStation[c][i] // 按比例计算单元流域水面蒸发量
			}
		}
	}
}

func (w *Watershed) GetP(nt, nw int) float64 {
	return w.P[nt][nw]
}

func (w *Watershed) GetEM(nt, nw int) float64 {
	return w.EM[nt][nw]
}

func (w *Watershed) GetF(nw int) float64 {
	return w.AreaSubWatershed[nw]
}

func (w *Watershed) GetnW() int {
	return w.NumSubWatershed
}

func NewWatershed() *Watershed {
	return &Watershed{
		Name:                   "默认流域",
		Area:                   0.0,
		NumRainfallStation:     0,
		NumEvaporationStation:  0,
		NumSubWatershed:        0,
		AreaSubWatershed:       nil,
		RateRainfallStation:    make([][]float64, 0, 0),
		RateEvaporationStation: make([][]float64, 0, 0),
		NameRainfallStation:    nil,
		NameEvaporationStation: nil,
		P:                      make([][]float64, 0, 0),
		EM:                     make([][]float64, 0, 0),
	}
}

func (w *Watershed) Destroy() {
	if w.AreaSubWatershed != nil {
		w.AreaSubWatershed = nil
	}
	if w.NameRainfallStation != nil {
		w.NameRainfallStation = nil
	}
	if w.NameEvaporationStation != nil {
		w.NameEvaporationStation = nil
	}
}

type IO struct {
	MQ    []float64 // 流量
	Q     []float64 // 观测流量
	Nrows int
	Ncols int
	Mp    [][]float64 // 降雨
	MEM   [][]float64 // 蒸发
}

func (io *IO) ReadFromFile(strPath string) {
	// 读取降雨数据
	file, err := os.Open(strPath + "P.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		fmt.Sscanf(scanner.Text(), "%d %d", &io.Nrows, &io.Ncols)
	}
	io.Mp = make([][]float64, io.Nrows)
	for r := 0; r < io.Nrows; r++ {
		io.Mp[r] = make([]float64, io.Ncols)
		mp := make([]string, 0)
		if scanner.Scan() {
			mp = strings.Split(scanner.Text(), "\t")
		}
		for c := 0; c < io.Ncols; c++ {
			io.Mp[r][c], _ = strconv.ParseFloat(strings.TrimSpace(mp[c]), 64)
		}
	}

	// 读取蒸发数据
	file, err = os.Open(strPath + "EM.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner = bufio.NewScanner(file)
	if scanner.Scan() {
		fmt.Sscanf(scanner.Text(), "%d %d", &io.Nrows, &io.Ncols)
	}
	io.MEM = make([][]float64, io.Nrows)
	for r := 0; r < io.Nrows; r++ {
		io.MEM[r] = make([]float64, io.Ncols)
		for c := 0; c < io.Ncols; c++ {
			if scanner.Scan() {
				io.MEM[r][c], _ = strconv.ParseFloat(scanner.Text(), 64)
			}
		}
	}

	// 读取观测流量数据
	file, err = os.Open(strPath + "observed_Q.txt")
	if err != nil {
		fmt.Println("Error opening observed flow file:", err)
		return
	}
	defer file.Close()

	scanner = bufio.NewScanner(file)
	if scanner.Scan() {
		fmt.Sscanf(scanner.Text(), "%d %d", &io.Nrows, &io.Ncols)
	}

	// 初始化观测流量数组
	io.Q = make([]float64, io.Nrows)

	// 读取观测流量数据
	for i := 0; i < io.Nrows; i++ {
		if scanner.Scan() {
			io.Q[i], _ = strconv.ParseFloat(scanner.Text(), 64)
		}
	}

}

func (io *IO) WriteToFile(strPath string) {
	// 打开Q.txt输出流域出口断面流量过程，没有该文件则新建
	file, err := os.Create(strPath + "Q.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	// 输出流量过程
	for i := 0; i < io.Nrows; i++ {
		fmt.Fprintln(writer, io.MQ[i])
	}
	writer.Flush()
}

func NewIO() *IO {
	return &IO{
		MQ:    nil,
		Nrows: 0,
		Ncols: 0,
		Mp:    make([][]float64, 0, 0),
		MEM:   make([][]float64, 0, 0),
	}
}

func (io *IO) Destroy() {
	if io.MQ != nil {
		io.MQ = nil
	}
}
