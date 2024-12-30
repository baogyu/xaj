package Data

import (
	"bufio"
	"demo2/Watershed"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type State struct {
	// 外部输入
	P  float64 // 单元流域降雨量，mm
	EM float64 // 单元流域水面蒸发量，mm
	F  float64 // 单元流域面积,km2
	Dt float64 // 模型计算时段长,h

	// 模型状态
	S0  float64   // 本时段初产流面积上的平均自由水深，mm
	FR  float64   // 本时段产流面积比例
	EU  float64   // 上层蒸散发量，mm
	EL  float64   // 下层蒸散发量，mm
	ED  float64   // 深层蒸散发量，mm
	E   float64   // 总的蒸散发量，mm
	EP  float64   // 流域蒸发能力，mm
	WU  float64   // 上层张力水蓄量，mm
	WL  float64   // 下层张力水蓄量，mm
	WD  float64   // 深层张力水蓄量，mm
	W   float64   // 总的张力水蓄量，mm
	RIM float64   // 不透水面积上的产流量，mm
	R   float64   // 总径流量，mm
	RS  float64   // 地面径流，mm
	RI  float64   // 壤中流，mm
	RG  float64   // 地下径流，mm
	PE  float64   // 净雨量，mm，PE = P - KC * EM
	QS  float64   // 地面径流汇流，m3/s
	QI  float64   // 壤中流汇流，m3/s
	QG  float64   // 地下径流汇流，m3/s
	QU  float64   // 本时段末单元流域出口流量，m3/s
	QU0 float64   // 上一时段末即本时段初的单元流域出口流量，m3/s
	O   []float64 // 单元流域在各子河段出口断面形成的出流，m3/s
	O2  float64   // 单元流域在全流域出口断面形成的出流，m3/s

	// 输出外部
	Q float64 // 流域出口断面流量，m3/s
}

func NewState(P, EM, F, dt, S0, FR, EU, EL, ED, E, WU, WL, WD, W,
	RIM, R, RS, RI, RG, PE, QS, QI, QG, QU, QU0, O2, EP, Q float64, O []float64) *State {

	return &State{
		P: P, EM: EM, F: F, Dt: dt,
		S0: S0, FR: FR, EU: EU, EL: EL, ED: ED, E: E,
		WU: WU, WL: WL, WD: WD, W: W, RIM: RIM, R: R,
		RS: RS, RI: RI, RG: RG, PE: PE, QS: QS, QI: QI,
		QG: QG, QU: QU, QU0: QU0, O: O, O2: O2,
		EP: EP, Q: Q,
	}
}

// 设置外部输入
func (s *State) SetInput(nt, nw int, watershed *Watershed.Watershed) {
	s.P = watershed.GetP(nt, nw)
	s.EM = watershed.GetEM(nt, nw)
	s.F = watershed.GetF(nw)
}

// 从文件读取时段长
func (s *State) ReadFromFile(filePath string) {
	//file, err := os.Open(filePath + "time.txt")
	//if err != nil {
	//	fmt.Printf("无法打开文件: %s, 错误: %v\n", filePath, err)
	//	return
	//}
	//defer file.Close()
	// 模拟读取逻辑
	s.Dt = 24.0 // 假定读取的值
}

// 设置状态值
func (s *State) SetValues(P, EM, F, dt, S0, FR,
	EU, EL, ED, E,
	WU, WL, WD, W,
	RIM, R, RS, RI, RG,
	PE, QS, QI, QG, QU,
	QU0 float64, O []float64, O2, EP, Q float64) {

	s.P = P
	s.EM = EM
	s.F = F
	s.Dt = dt
	s.S0 = S0
	s.FR = FR
	s.EU = EU
	s.EL = EL
	s.ED = ED
	s.E = E
	s.WU = WU
	s.WL = WL
	s.WD = WD
	s.W = W
	s.RIM = RIM
	s.R = R
	s.RS = RS
	s.RI = RI
	s.RG = RG
	s.PE = PE
	s.QS = QS
	s.QI = QI
	s.QG = QG
	s.QU = QU
	s.QU0 = QU0
	s.O = O
	s.O2 = O2
	s.EP = EP
	s.Q = Q
	fmt.Println("状态值已更新")
}

type Parameter struct {
	// 蒸散发计算参数
	KC float64 // 流域蒸散发折算系数，敏感
	UM float64 // 上层张力水容量/mm，敏感，10~50
	LM float64 // 下层张力水容量/mm，敏感，60~90
	C  float64 // 深层蒸散发折算系数，不敏感，0.10~0.20

	// 产流计算参数
	WM float64 // 流域平均张力水容量/mm，不敏感，120~200
	B  float64 // 张力水蓄水容量曲线方次，不敏感，0.1~0.4
	IM float64 // 不透水面积占全流域面积的比例，不敏感

	// 水源划分参数
	SM float64 // 表层自由水蓄水容量/mm，敏感
	EX float64 // 表层自由水蓄水容量方次，不敏感，1.0~1.5
	KG float64 // 表层自由水蓄水库对地下水的日出流系数，敏感
	KI float64 // 表层自由水蓄水库对壤中流的日出流系数，敏感

	// 汇流计算参数
	CS float64 // 日模型地面径流消退系数，敏感
	CI float64 // 日模型壤中流蓄水库的消退系数，敏感
	CG float64 // 日模型地下水蓄水库的消退系数，敏感
	CR float64 // 日模型河网蓄水消退系数，敏感
	KE float64 // 马斯京根法演算参数/h，敏感，KE = N * ∆t，N为河道分段数
	XE float64 // 马斯京根法演算参数，敏感，0.0~0.5
}

func NewParameter(KC, UM, LM, C, WM, B, IM, SM, EX, KG, KI, CS, CI, CG, CR, KE, XE float64) *Parameter {
	return &Parameter{
		KC: KC, UM: UM, LM: LM, C: C,
		WM: WM, B: B, IM: IM,
		SM: SM, EX: EX, KG: KG, KI: KI,
		CS: CS, CI: CI, CG: CG, CR: CR, KE: KE, XE: XE,
	}
}

// 从文件中读取模型参数
func (p *Parameter) ReadFromFile(filePath string) {
	file, err := os.Open(filePath + "parameter.txt")
	if err != nil {
		fmt.Printf("无法打开文件: %s, 错误: %v\n", filePath, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var values []float64
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		for _, field := range fields {
			value, err := strconv.ParseFloat(field, 64)
			if err != nil {
				fmt.Printf("解析失败: %s, 错误: %v\n", field, err)
				continue
			}
			values = append(values, value)
		}
	}

	// 验证读取的数据是否足够
	if len(values) < 17 {
		fmt.Printf("参数文件数据不足，读取的数量: %d\n", len(values))
		return
	}

	// 分别设置参数
	p.KC = values[0]
	p.UM = values[1]
	p.LM = values[2]
	p.C = values[3]
	p.WM = values[4]
	p.B = values[5]
	p.IM = values[6]
	p.SM = values[7]
	p.EX = values[8]
	p.KG = values[9]
	p.KI = values[10]
	p.CS = values[11]
	p.CI = values[12]
	p.CG = values[13]
	p.CR = values[14]
	p.KE = values[15]
	p.XE = values[16]
}

// 设置参数值
func (p *Parameter) SetValues(KC, UM, LM, C, WM, B, IM, SM, EX, KG, KI, CS, CI, CG, CR, KE, XE float64) {
	p.KC = KC
	p.UM = UM
	p.LM = LM
	p.C = C
	p.WM = WM
	p.B = B
	p.IM = IM
	p.SM = SM
	p.EX = EX
	p.KG = KG
	p.KI = KI
	p.CS = CS
	p.CI = CI
	p.CG = CG
	p.CR = CR
	p.KE = KE
	p.XE = XE
	fmt.Println("模型参数已更新")
}
