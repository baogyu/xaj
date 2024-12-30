package Calibration

import (
	"bufio"
	"demo2/Confluence"
	"demo2/Data"
	"demo2/Evapotranspiration"
	"demo2/Muskingum"
	"demo2/Source"
	"demo2/Watershed"
	Runoff "demo2/runoff"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
)

// SCEUA 结构体定义
type SCEUA struct {
	// 优化参数
	nopt  int       // 待优化的参数数量
	xname []string  // 变量名
	a     []float64 // 参数初始值
	bl    []float64 // 参数下限
	bu    []float64 // 参数上限

	// SCE控制参数
	ngs   int // 复形数量
	npg   int // 每个复形中的点的个数
	nps   int // 子复形中的点数
	alpha int // CCE步骤中重复次数
	beta  int // CCE步骤3-5重复次数
	npt   int // 初始种群中的总点数

	// 收敛检查参数
	maxn   int     // 最大试验次数
	kstop  int     // 收敛判断循环数
	pcento float64 // 函数值变化百分比
	peps   float64 // 最低变化率

	// 标志变量
	iniflg bool // 是否包含初始点
	iprint bool // 打印控制标志
	ideflt bool // 是否使用默认参数

	// 优化结��
	bestx [][]float64 // 每次洗牌循环的最佳点
	bestf []float64   // 每次循环的最佳点函数值

	// 收敛判据
	icall  []int     // 模型调用次数
	timeou []float64 // 函数值变化率
	gnrng  []float64 // 参数范围归一化几何平均值

	// 模型数据
	measuredValues  []float64 // 实测值
	simulatedValues []float64 // 模拟值

	// 文件路径
	filePath string // 工作目录路径
}

// NewSCEUA 创建新的SCEUA优化器实例
func NewSCEUA() *SCEUA {
	s := &SCEUA{
		nopt:  7,
		xname: []string{"x1", "x2", "x3", "x4", "x5", "x6", "x7"},
		a:     make([]float64, 7),
		bl:    []float64{-2, -2, -2, -2, -2, -2, -2},
		bu:    []float64{2, 2, 2, 2, 2, 2, 2},

		ngs:    10,
		alpha:  1,
		maxn:   5000,
		kstop:  15,
		pcento: 0.1,
		peps:   0.001,

		iniflg: true,
		iprint: false,
		ideflt: false,
	}

	// 设置依赖参数
	s.npg = 2*s.nopt + 1
	s.nps = s.nopt + 1
	s.beta = 2*s.nopt + 1
	s.npt = s.ngs * s.npg

	return s
}

// Optimize 执行SCE-UA优化
func (s *SCEUA) Optimize() {
	s.scemain()
}

// scemain 主要优化流程
func (s *SCEUA) scemain() {
	s.scein()  // 设置优化参数
	s.sceua()  // SCE-UA算法
	s.sceout() // 输出优化结果
}

// scein 读取优化参数
func (s *SCEUA) scein() {
	// 打开输入文件
	file, err := os.Open(s.filePath + "scein.txt")
	if err != nil {
		fmt.Printf("无法打开输入文件: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// 读取参数个数
	if scanner.Scan() {
		fmt.Sscanf(scanner.Text(), "%d", &s.nopt)
	}

	// 初始化参数数组
	s.xname = make([]string, s.nopt)
	s.a = make([]float64, s.nopt)
	s.bl = make([]float64, s.nopt)
	s.bu = make([]float64, s.nopt)

	// 读取每个变量的参数名、初始值、下限、上限
	for i := 0; i < s.nopt; i++ {
		if scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) >= 4 {
				s.xname[i] = fields[0]
				fmt.Sscanf(fields[1], "%f", &s.a[i])
				fmt.Sscanf(fields[2], "%f", &s.bl[i])
				fmt.Sscanf(fields[3], "%f", &s.bu[i])
			}
		}
	}

	// 读取是否使用默认参数标识
	if scanner.Scan() {
		fmt.Sscanf(scanner.Text(), "%t", &s.ideflt)
	}

	// 如果不使用默认参数,则读取SCE控制参数
	if s.ideflt {
		if scanner.Scan() {
			fmt.Sscanf(scanner.Text(), "%d %d %d %d %d",
				&s.ngs, &s.npg, &s.nps, &s.alpha, &s.beta)
		}
		if scanner.Scan() {
			fmt.Sscanf(scanner.Text(), "%d %d %f %f",
				&s.maxn, &s.kstop, &s.pcento, &s.peps)
		}
		if scanner.Scan() {
			fmt.Sscanf(scanner.Text(), "%t %t",
				&s.iniflg, &s.iprint)
		}
	} else {
		// 使用默认参数
		s.ngs = 10
		s.npg = 2*s.nopt + 1
		s.nps = s.nopt + 1
		s.alpha = 1
		s.beta = 2*s.nopt + 1
		s.maxn = 10000
		s.kstop = 15
		s.pcento = 0.05
		s.peps = 0.0005
		s.iniflg = true
		s.iprint = false
	}

	// 计算初始种群中的总点数
	s.npt = s.ngs * s.npg

	// 读取实测值
	s.measuredValues = s.ReadValues(s.filePath + "observe.txt")
}

// sceout 输出优化结果
func (s *SCEUA) sceout() {
	// 控制台输出
	fmt.Println("===================参数初始值及上下限===================")

	fmt.Print("参数名：")
	for _, name := range s.xname {
		fmt.Printf("%s  ", name)
	}
	fmt.Println()

	fmt.Print("参数初始值：")
	for _, val := range s.a {
		fmt.Printf("%f  ", val)
	}
	fmt.Println()

	fmt.Print("参数下限：")
	for _, val := range s.bl {
		fmt.Printf("%f  ", val)
	}
	fmt.Println()

	fmt.Print("参数上限：")
	for _, val := range s.bu {
		fmt.Printf("%f  ", val)
	}
	fmt.Println()

	fmt.Printf("初始参数函数值：%f\n", s.functn(s.a))

	// 打印优化结果
	fmt.Println("===================SCE-UA搜索的结果===================")

	fmt.Println("最优点:")
	for _, name := range s.xname {
		fmt.Printf("%s  ", name)
	}
	fmt.Println()

	for _, val := range s.bestx[len(s.bestx)-1] {
		fmt.Printf("%f  ", val)
	}
	fmt.Println()

	fmt.Printf("最优函数值: %f\n", s.bestf[len(s.bestf)-1])

	fmt.Printf("%d次洗牌演化的最优点函数值：\n", len(s.bestf))
	for _, val := range s.bestf {
		fmt.Printf("%f  ", val)
	}
	fmt.Println()

	// 文件输出
	file, err := os.Create(s.filePath + "sceout.txt")
	if err != nil {
		fmt.Printf("无法创建输出文件: %v\n", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	// ... 将上述相同的输出内容写入文件 ...
	writer.Flush()
}

// ReadValues 从文件中读取数值
func (s *SCEUA) ReadValues(fileName string) []float64 {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("无法打开文件 %s: %v\n", fileName, err)
		return nil
	}
	defer file.Close()

	var values []float64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var val float64
		fmt.Sscanf(scanner.Text(), "%f", &val)
		values = append(values, val)
	}

	return values
}

// sceua SCE-UA主优化算法
func (s *SCEUA) sceua() {
	fmt.Println("==================================================")
	fmt.Println("                  进入SCE-UA全局搜索                ")
	fmt.Println("==================================================")

	// 1. 初始局部变量
	x := make([][]float64, s.npt) // 种群中点的坐标
	for i := range x {
		x[i] = make([]float64, s.nopt)
	}
	xf := make([]float64, s.npt) // 种群中点的函数值

	cx := make([][]float64, s.npg) // 复形中点的坐标
	for i := range cx {
		cx[i] = make([]float64, s.nopt)
	}
	cf := make([]float64, s.npg) // 复形中点的函数值

	xnstd := make([]float64, s.nopt) // 种群中参数的标准差
	for i := range xnstd {
		xnstd[i] = 0.1
	}

	bound := make([]float64, s.nopt) // 优化变量的约束范围
	for j := 0; j < s.nopt; j++ {
		bound[j] = s.bu[j] - s.bl[j]
	}

	icall := 0        // 模型调用次数
	timeou := 10000.0 // 函数值变化率
	gnrng := 10000.0  // 参数范围归一化几何平均值
	nloop := 0        // 主循环次数

	// 2. 生成初始样本
	s.GenerateSample(x, xf, &icall)

	// 3. 样本点排序
	s.RankPoints(x, xf)

	fmt.Println("目标函数调用次数  函数值变化率  参数变化范围")

	// 主循环
	for icall < s.maxn && timeou > s.pcento && gnrng > s.peps {
		nloop++

		// 对每个复形进行独立演化
		for k := 0; k < s.ngs; k++ {
			// 4. 划分复形群体
			s.Partition2Complexes(k, x, xf, cx, cf)

			// 5. 复形演化
			s.cce(cx, cf, xnstd, &icall)

			// 6. 复形牌
			s.ShuffleComplexes(k, x, xf, cx, cf)
		}

		// 3. 样本点排序
		s.RankPoints(x, xf)

		// 7. 收敛判断
		s.CheckConvergence(x, xnstd, bound, nloop, &icall, &timeou, &gnrng)

		fmt.Printf("      %d            %f        %f\n", icall, timeou, gnrng)
	}
}

// GenerateSample 在参数空间内生成初始样本点
func (s *SCEUA) GenerateSample(x [][]float64, xf []float64, icall *int) {
	xx := make([]float64, s.nopt)

	// 生成随机点
	for i := 0; i < s.npt; i++ {
		s.getpnt(xx)
		copy(x[i], xx)
	}

	// 如果需要包含初始点
	if s.iniflg {
		copy(x[0], s.a)
	}

	// 计算函数值
	for i := 0; i < s.npt; i++ {
		xf[i] = s.functn(x[i])
		*icall++
	}
}

// getpnt 在可行区域内生成一个随机点
func (s *SCEUA) getpnt(snew []float64) {
	rand.Seed(time.Now().UnixNano())
	ibound := true

	for ibound {
		for j := 0; j < s.nopt; j++ {
			snew[j] = s.bl[j] + rand.Float64()*(s.bu[j]-s.bl[j])
		}
		s.chkcst(snew, &ibound)
	}
}

// getpntNormal 根据正态分布生成新点
func (s *SCEUA) getpntNormal(xi, std, snew []float64) {
	rand.Seed(time.Now().UnixNano())
	ibound := true

	for ibound {
		for j := 0; j < s.nopt; j++ {
			snew[j] = rand.NormFloat64()*std[j] + xi[j]
		}
		s.chkcst(snew, &ibound)
	}
}

// chkcst 检查点是否满足约束条件
func (s *SCEUA) chkcst(xx []float64, ibound *bool) {
	*ibound = false
	for i := 0; i < s.nopt; i++ {
		if xx[i] < s.bl[i] || xx[i] > s.bu[i] {
			*ibound = true
			return
		}
	}
}

// RankPoints 对样本点按函数值升序排序
func (s *SCEUA) RankPoints(x [][]float64, xf []float64) {
	s.sort(x, xf)

	// 记录最优点
	bestx := make([]float64, s.nopt)
	copy(bestx, x[0])
	s.bestx = append(s.bestx, bestx)
	s.bestf = append(s.bestf, xf[0])
}

// sort 按函数值升序排序
func (s *SCEUA) sort(x [][]float64, xf []float64) {
	// 使用冒泡排序
	for i := 0; i < len(xf)-1; i++ {
		for j := 0; j < len(xf)-i-1; j++ {
			if xf[j] > xf[j+1] {
				// 交换函数值
				xf[j], xf[j+1] = xf[j+1], xf[j]
				// 交换对应的点
				x[j], x[j+1] = x[j+1], x[j]
			}
		}
	}
}

// Partition2Complexes 将样本点划分到ngs个复形中
func (s *SCEUA) Partition2Complexes(k int, x [][]float64, xf []float64, cx [][]float64, cf []float64) {
	// 遍历复形中的每个点，为其从种群中赋值
	for j := 0; j < s.npg; j++ {
		index := k + s.ngs*j // 计算下标

		// 划分样本点到复形
		copy(cx[j], x[index])
		cf[j] = xf[index]
	}
}

// ShuffleComplexes 将复形中的点放回种群
func (s *SCEUA) ShuffleComplexes(k int, x [][]float64, xf []float64, cx [][]float64, cf []float64) {
	// 遍历复形中的每个点，将其值赋给种群
	for j := 0; j < s.npg; j++ {
		index := k + s.ngs*j // 计算下标

		// 样本点从复形赋值到种群
		copy(x[index], cx[j])
		xf[index] = cf[j]
	}
}

// cce 复形演化算法
func (s *SCEUA) cce(cx [][]float64, cf []float64, xnstd []float64, icall *int) {
	// 初始化局部变量
	ss := make([][]float64, s.nps) // 当前单纯形中点的坐标
	for i := range ss {
		ss[i] = make([]float64, s.nopt)
	}
	sf := make([]float64, s.nps) // 单纯形中点的函数值

	sb := make([]float64, s.nopt)   // 单纯形的最佳点
	sw := make([]float64, s.nopt)   // 单纯形的最差点
	ce := make([]float64, s.nopt)   // 单纯形排除最差点的形心
	snew := make([]float64, s.nopt) // 从单纯形生成的新点

	// 遍历beta次
	for ibeta := 0; ibeta < s.beta; ibeta++ {
		// 选择父辈群体
		lcs := s.selectParents(s.npg, s.nps)

		// 构建单纯形
		for i := 0; i < s.nps; i++ {
			copy(ss[i], cx[lcs[i]])
			sf[i] = cf[lcs[i]]
		}

		// 对单纯形进行alpha次演化
		for ialpha := 0; ialpha < s.alpha; ialpha++ {
			// 获取最佳点���最差点
			copy(sb, ss[0])
			copy(sw, ss[s.nps-1])
			fw := sf[s.nps-1]

			// 计算心(不包括最差点)
			for j := 0; j < s.nopt; j++ {
				sum := 0.0
				for i := 0; i < s.nps-1; i++ {
					sum += ss[i][j]
				}
				ce[j] = sum / float64(s.nps-1)
			}

			// 反射步骤
			for j := 0; j < s.nopt; j++ {
				snew[j] = 2.0*ce[j] - sw[j]
			}

			var ibound bool
			s.chkcst(snew, &ibound)

			if ibound {
				s.getpntNormal(sb, xnstd, snew)
			}

			// 计算新点函数值
			fnew := s.functn(snew)
			*icall++

			// 比较并更新
			if fnew < fw {
				copy(ss[s.nps-1], snew)
				sf[s.nps-1] = fnew
			} else {
				// 收缩步骤
				for j := 0; j < s.nopt; j++ {
					snew[j] = (ce[j] + sw[j]) / 2.0
				}

				fnew = s.functn(snew)
				*icall++

				if fnew < fw {
					copy(ss[s.nps-1], snew)
					sf[s.nps-1] = fnew
				} else {
					// 突变步骤
					s.getpntNormal(sb, xnstd, snew)
					fnew = s.functn(snew)
					*icall++

					copy(ss[s.nps-1], snew)
					sf[s.nps-1] = fnew
				}
			}

			// 对单纯形重新排序
			s.sort(ss, sf)
		}

		// 将演化后的点放回复形
		for i := 0; i < s.nps; i++ {
			copy(cx[lcs[i]], ss[i])
			cf[lcs[i]] = sf[i]
		}

		// 对复形重新排序
		s.sort(cx, cf)
	}
}

// selectParents 从复形中选择父代点
func (s *SCEUA) selectParents(npg, nps int) []int {
	// 计算每个点的权重
	wts := make([]float64, npg)
	for i := 0; i < npg; i++ {
		wts[i] = float64(npg - i)
	}

	// 随机扰动权重
	rand.Seed(time.Now().UnixNano())
	vals := make([]struct {
		idx int
		val float64
	}, npg)

	for i := 0; i < npg; i++ {
		vals[i].idx = i
		vals[i].val = math.Pow(rand.Float64(), 1.0/wts[i])
	}

	// 按扰动后的权重排序
	sort.Slice(vals, func(i, j int) bool {
		return vals[i].val > vals[j].val
	})

	// 选择前nps个点
	lcs := make([]int, nps)
	for i := 0; i < nps; i++ {
		lcs[i] = vals[i].idx
	}

	// 按索引升序排序
	sort.Ints(lcs)

	return lcs
}

// CheckConvergence 检查是否满足收敛条件
func (s *SCEUA) CheckConvergence(x [][]float64, xnstd []float64, bound []float64, nloop int, icall *int, timeou *float64, gnrng *float64) {
	// 检查是否超过最大试验次数
	if *icall >= s.maxn {
		fmt.Printf("经过%d次洗牌演化，优化搜索已经终止，因为超过了最大试验次数%d次的限制\n",
			nloop, s.maxn)
		return
	}

	// 检查函数值改进
	if nloop >= s.kstop {
		// 计算最近kstop次数的函数值平均值
		sum := 0.0
		for i := len(s.bestf) - s.kstop; i < len(s.bestf); i++ {
			sum += math.Abs(s.bestf[i])
		}
		denomi := sum / float64(s.kstop)

		// 计算变化率
		*timeou = math.Abs(s.bestf[len(s.bestf)-1]-s.bestf[len(s.bestf)-s.kstop]) / denomi

		if *timeou < s.pcento {
			fmt.Printf("最佳点在最近%d次循环中函数值变化率小于阈值%f\n",
				s.kstop, s.pcento)
			fmt.Printf("经过%d次洗牌演化，基于目标函数标准已经实现了收敛！\n", nloop)
		}
	}

	// 检查参数空间收敛性
	xmax := make([]float64, s.nopt)
	xmin := make([]float64, s.nopt)
	xmean := make([]float64, s.nopt)

	// 计算各参数的统计值
	for j := 0; j < s.nopt; j++ {
		// 提取第j个参数的所有值
		par := make([]float64, s.npt)
		for i := 0; i < s.npt; i++ {
			par[i] = x[i][j]
		}

		// 计算最大值、最小值、平均值
		xmax[j] = maxSlice(par)
		xmin[j] = minSlice(par)
		xmean[j] = meanSlice(par)

		// 计算标准差
		sum2 := 0.0
		for _, v := range par {
			sum2 += v * v
		}
		xnstd[j] = math.Sqrt(sum2/float64(s.npt) - xmean[j]*xmean[j])
		xnstd[j] /= bound[j] // 归一化
	}

	// 计算几何平均值
	const delta = 1e-20
	gsum := 0.0
	for j := 0; j < s.nopt; j++ {
		gsum += math.Log(delta + (xmax[j]-xmin[j])/bound[j])
	}
	*gnrng = math.Exp(gsum / float64(s.nopt))

	if *gnrng < s.peps {
		fmt.Printf("经过%d次洗牌演化，种群已经收敛到一个预先指定的小参数空间\n", nloop)
	}

	// 记录收敛判据
	s.icall = append(s.icall, *icall)
	s.timeou = append(s.timeou, *timeou)
	s.gnrng = append(s.gnrng, *gnrng)
}

// functn 计算目标函数值
func (s *SCEUA) functn(x []float64) float64 {
	// 1. 前处理
	s.PreProcessing(x)

	// 2. 运行模型
	s.RunModel()

	// 3. 后处理
	return s.PostProcessing()
}

// PreProcessing 前处理，将参数写入模型输入文件
func (s *SCEUA) PreProcessing(x []float64) {
	// 打开模板文件
	fin, err := os.Open(s.filePath + "parameter.tpl")
	if err != nil {
		fmt.Printf("无法打开参数模板文件: %v\n", err)
		return
	}
	defer fin.Close()

	// 创建输出文件
	fout, err := os.Create(s.filePath + "parameter.txt")
	if err != nil {
		fmt.Printf("无法创建参数文件: %v\n", err)
		return
	}
	defer fout.Close()

	scanner := bufio.NewScanner(fin)
	for scanner.Scan() {
		parameter := scanner.Text()

		// 查找参数名是否在待优化参数中
		found := false
		for i, name := range s.xname {
			if parameter == name {
				fmt.Fprintf(fout, "%f\n", x[i])
				found = true
				break
			}
		}

		// 如果不是待优化参数，直接写入
		if !found {
			fmt.Fprintln(fout, parameter)
		}
	}
}

// RunModel 运行水文模型
func (s *SCEUA) RunModel() {
	// 调用实际的水文模型
	// 获取可执行文件所在路径
	path := "/Users/baogy/goProject/owner/demo2/datas/"

	// 读取流域分块信息
	var watershed Watershed.Watershed
	watershed.ReadFromFile(path)

	var parameter Data.Parameter
	parameter.ReadFromFile(path)

	var io Watershed.IO
	io.ReadFromFile(path)
	watershed.Calculate(&io)
	// 首先进行参数率定

	// 设置模型实例

	// 创建状态数组
	nw := watershed.GetnW()
	states := make([]*Data.State, nw)
	for i := 0; i < nw; i++ {
		states[i] = &Data.State{
			WU: 0.085292130191392,
			WL: 25.3537665495867,
			WD: 36.38818002,
			W:  61.82723869977809,
			Dt: 24.0,
		}
		states[i].ReadFromFile(path)
	}

	//流域蒸散发
	var evapotranspiration Evapotranspiration.Evapotranspiration
	evapotranspiration.WL = 25.3537665495867
	evapotranspiration.WU = 0.085292130191392
	evapotranspiration.SetParmameter(&parameter)

	//流域产流
	var runoff Runoff.Runoff
	runoff.WU = 0.085292130191392
	runoff.WL = 25.3537665495867
	runoff.WD = 36.38818002
	runoff.WDM = 20
	runoff.SetParmameter(&parameter)

	//流域分水源
	var source Source.Source
	source.N = 1
	source.SetParmameter(&parameter)

	//呈村流域汇流
	var confluence Confluence.Confluence
	confluence.Dt = states[0].Dt
	confluence.SetParmameter(&parameter)

	//呈村流域河道汇流
	var muskingum Muskingum.Muskingum
	muskingum.Dt = states[0].Dt
	muskingum.N = 1
	muskingum.SetParmameter(&parameter)

	// 逐时段逐单元流域计算
	nT := io.Nrows
	io.MQ = make([]float64, nT)
	for t := 0; t < nT; t++ {
		states[0].Q = 0.0
		for w := 0; w < nw; w++ {
			states[w].SetInput(t, w, &watershed)
			evapotranspiration.SetState(states[w])
			evapotranspiration.Calculate()
			evapotranspiration.UpdateState(states[w])
			runoff.SetState(states[w])
			runoff.Calculate()
			runoff.UpdateState(states[w])
			source.SetState(states[w])
			source.Calculate()
			source.UpdateState(states[w])
			confluence.SetState(states[w])
			confluence.Calculate()
			confluence.UpdateState(states[w])
			muskingum.SetState(states[w])
			muskingum.Calculate()
			muskingum.UpdateState(states[w])
			states[0].Q += states[w].O2
		}
		io.MQ[t] = states[0].Q
	}

	// 输出流域出口断面流量过程到文本Q.txt中
	io.WriteToFile(path)
}

// PostProcessing 后处理，计算目标函数值
func (s *SCEUA) PostProcessing() float64 {
	// 读取模拟值
	s.simulatedValues = s.ReadValues(s.filePath + "Q.txt")

	// 计算NSE
	nse := s.CalculateNSE(s.simulatedValues, s.measuredValues)

	// 返回1-NSE作为优化目标
	return 1 - nse
}

// CalculateNSE 计算Nash-Sutcliffe效率系数
func (s *SCEUA) CalculateNSE(simulatedValues, measuredValues []float64) float64 {
	var sumSquaredError, sumSquaredDeviation float64

	// 计算观测流量平均值
	measuredMean := meanSlice(measuredValues)

	// 计算NSE
	for i := range measuredValues {
		sumSquaredError += math.Pow(measuredValues[i]-simulatedValues[i], 2)
		sumSquaredDeviation += math.Pow(measuredValues[i]-measuredMean, 2)
	}

	return 1 - (sumSquaredError / sumSquaredDeviation)
}

// 辅助函数
func maxSlice(s []float64) float64 {
	if len(s) == 0 {
		return 0
	}
	max := s[0]
	for _, v := range s {
		if v > max {
			max = v
		}
	}
	return max
}

func minSlice(s []float64) float64 {
	if len(s) == 0 {
		return 0
	}
	min := s[0]
	for _, v := range s {
		if v < min {
			min = v
		}
	}
	return min
}

func meanSlice(s []float64) float64 {
	if len(s) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range s {
		sum += v
	}
	return sum / float64(len(s))
}

// SetFilePath 设置工作目录路径
func (s *SCEUA) SetFilePath(path string) {
	s.filePath = path
}
