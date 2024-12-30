package Confluence

import (
	"demo2/Data"
	"math"
)

type Confluence struct {
	// ========模型参数======== //
	CS float64 // 地面径流消退系数，敏感
	CI float64 // 壤中流消退系数，敏感
	CG float64 // 地下水消退系数，敏感
	CR float64 // 河网蓄水消退系数，敏感
	IM float64 // 不透水面积占全流域面积的比例，不敏感

	// ========模型状态======== //
	QS  float64 // 单元流域地面径流（m³/s）
	QI  float64 // 单元流域壤中流（m³/s）
	QG  float64 // 单元流域地下径流（m³/s）
	QT  float64 // 单元流域河网总入流（m³/s），即进入单元面积的地面径流、壤中流和地下径流之和
	QU  float64 // 单元流域出口流量（m³/s）
	RS  float64 // 地面径流量（mm）
	RI  float64 // 壤中流径流量（mm）
	RG  float64 // 地下径流量（mm）
	RIM float64 // 不透水面积上的产流量（mm）
	QI0 float64 // 前一时刻壤中流 QI(t-1)（m³/s）
	QG0 float64 // 前一时刻地下径流 QG(t-1)（m³/s）
	QU0 float64 // 前一时刻单元流域出口流量 QU(t-1)（m³/s）
	F   float64 // 单元流域面积（km²）
	U   float64 // 单位转换系数

	// ========计算参数======== //
	M   float64 // 一天划分的计算时段数
	CSD float64 // 计算时段内地面径流蓄水库的消退系数
	CID float64 // 计算时段内壤中流蓄水库的消退系数
	CGD float64 // 计算时段内地下水蓄水库的消退系数
	CRD float64 // 计算时段内河网蓄水消退系数
	Dt  float64 // 模型计算时段长（h）
}

func (c *Confluence) SetParmameter(parameter *Data.Parameter) {
	c.CS = parameter.CS
	c.CI = parameter.CI
	c.CG = parameter.CG
	c.CR = parameter.CR
	c.IM = parameter.IM
}

func (c *Confluence) SetState(state *Data.State) {
	c.RIM = state.RIM
	c.RS = state.RS
	c.RI = state.RI
	c.RG = state.RG
	c.QS = state.QS
	c.QI = state.QI
	c.QG = state.QG
	c.QU0 = state.QU
	c.F = state.F
	c.Dt = state.Dt
}

func (c *Confluence) UpdateState(state *Data.State) {
	state.QS = c.QS
	state.QI = c.QI
	state.QG = c.QG
	state.QU = c.QU
	state.QU0 = c.QU0
}

func (c *Confluence) Calculate() {
	// 计算时段系数
	c.M = 24.0 / c.Dt
	c.CSD = math.Pow(c.CS, 1.0/c.M)
	c.CID = math.Pow(c.CI, 1.0/c.M)
	c.CGD = math.Pow(c.CG, 1.0/c.M)
	c.CRD = math.Pow(c.CR, 1.0/c.M)
	c.U = c.F / 3.6 / c.Dt
	// 总地面径流深度
	totalSurfaceRunoff := c.RS*(1-c.IM) + c.RIM

	// 地面径流汇流计算
	c.QS = c.CSD*c.QS + (1-c.CSD)*totalSurfaceRunoff*c.U
	c.QI = c.CID*c.QI + (1-c.CID)*c.RI*(1-c.IM)*c.U
	c.QG = c.CGD*c.QG + (1-c.CGD)*c.RG*(1-c.IM)*c.U

	// 确保流量非负
	c.QS = math.Max(0, c.QS)
	c.QI = math.Max(0, c.QI)
	c.QG = math.Max(0, c.QG)

	// 计算总入流
	c.QT = c.QS + c.QI + c.QG

	// 河网汇流
	if c.F < 200 {
		c.QU = c.QT
	} else {
		c.QU = c.CRD*c.QU0 + (1-c.CRD)*c.QT
	}
}

func NewConfluence(cs, ci, cg, cr, im, qs, qi, qg, qt, qu, rs, ri, rg, rim, qi0, qg0, qu0, f, u, m, csd, cid, cgd, crd, dt float64) *Confluence {
	return &Confluence{
		CS:  cs,
		CI:  ci,
		CG:  cg,
		CR:  cr,
		IM:  im,
		QS:  qs,
		QI:  qi,
		QG:  qg,
		QT:  qt,
		QU:  qu,
		RS:  rs,
		RI:  ri,
		RG:  rg,
		RIM: rim,
		QI0: qi0,
		QG0: qg0,
		QU0: qu0,
		F:   f,
		U:   u,
		M:   m,
		CSD: csd,
		CID: cid,
		CGD: cgd,
		CRD: crd,
		Dt:  dt,
	}
}

func (c *Confluence) Destroy() {
	// 析构函数
}
