package Muskingum

import "demo2/Data"

type Muskingum struct {
	// ========模型参数======== //
	KE float64 // 马斯京根法演算参数（h），敏感，KE = N * ∆t
	XE float64 // 马斯京根法演算参数，敏感，取值范围：0.0~0.5

	// ========模型状态======== //
	KL float64   // 子河段的马斯京根法演算参数（h）
	XL float64   // 子河段的马斯京根法演算参数
	N  int       // 单元河段数，即分段数
	C0 float64   // 马斯京根流量演算公式 I2 系数
	C1 float64   // 马斯京根流量演算公式 I1 系数
	C2 float64   // 马斯京根流量演算公式 O1 系数
	I1 float64   // 时段初的河段入流量（m³/s）
	I2 float64   // 时段末的河段入流量（m³/s）
	O1 float64   // 时段初的河段出流量（m³/s）
	O2 float64   // 时段末的河段出流量（m³/s）
	O  []float64 // 各子河段出流量（m³/s）
	Dt float64   // 模型计算时段长（h）
}

func (m *Muskingum) SetParmameter(parameter *Data.Parameter) {
	m.KE = parameter.KE
	m.XE = parameter.XE
}

func (m *Muskingum) SetState(state *Data.State) {
	m.I1 = state.QU0
	m.I2 = state.QU
	m.O = state.O
	m.Dt = state.Dt
}

func (m *Muskingum) UpdateState(state *Data.State) {
	state.O = m.O
	state.O2 = m.O2
}

func (m *Muskingum) Calculate() {
	m.KL = m.Dt                            // 为了保证马斯京根法的两个线性条件，每个单元河取 KL = ∆t
	m.N = int(m.KE / m.KL)                 // 单元河段数
	m.XL = 0.5 - float64(m.N)*(1-2*m.XE)/2 // 计算单元河段XL

	denominator := 0.5*m.Dt + m.KL - m.KL*m.XL
	m.C0 = (0.5*m.Dt - m.KL*m.XL) / denominator
	m.C1 = (0.5*m.Dt + m.KL*m.XL) / denominator
	m.C2 = (-0.5*m.Dt + m.KL - m.KL*m.XL) / denominator

	if m.O == nil {
		m.O = make([]float64, m.N) // 创建存储单元流域在子河段出口断面的出流量的动态数组
		for n := 0; n < m.N; n++ {
			m.O[n] = 0.0 // 单元流域在子河段出口断面的出流量为0
		}
	}

	for n := 0; n < m.N; n++ {
		m.O1 = m.O[n]                            // 子河段时段初出流量
		m.O2 = m.C0*m.I2 + m.C1*m.I1 + m.C2*m.O1 // 计算时段末单元流域在子河段出口断面的出流量，m3/s
		m.O[n] = m.O2                            // 更新子河段时段初出流量
		m.I1 = m.O1                              // 上一河段时段初出流为下一河段时段初入流
		m.I2 = m.O2                              // 上一河段时段末出流为下一河段时段末入流
	}
}

func NewMuskingum(ke, xe, kl, xl float64, n int, c0, c1, c2, i1, i2, o1, o2 float64, o []float64, dt float64) *Muskingum {
	return &Muskingum{
		KE: ke,
		XE: xe,
		KL: kl,
		XL: xl,
		N:  n,
		C0: c0,
		C1: c1,
		C2: c2,
		I1: i1,
		I2: i2,
		O1: o1,
		O2: o2,
		O:  o,
		Dt: dt,
	}
}

func (m *Muskingum) Destroy() {
	if m.O != nil {
		m.O = nil
	}
}
