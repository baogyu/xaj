package Evapotranspiration

import (
	"demo2/Data"
)

type Evapotranspiration struct {
	KC float64 // 流域蒸散发折算系数，敏感
	LM float64 // 下层张力水容量（mm），敏感，范围：60~90
	C  float64 // 深层蒸散发折算系数，不敏感，范围：0.10~0.20

	// ========模型状态======== //
	WU float64 // 上层张力水蓄量（mm）
	WL float64 // 下层张力水蓄量（mm）
	EP float64 // 单元流域蒸发能力（mm）
	E  float64 // 总的蒸散发量（mm）
	EU float64 // 上层蒸散发量（mm）
	EL float64 // 下层蒸散发量（mm）
	ED float64 // 深层蒸散发量（mm）

	// ========外部输入======== //
	P  float64 // 降雨量（mm）
	EM float64 // 水面蒸发量（mm）
}

func (e *Evapotranspiration) SetParmameter(parameter *Data.Parameter) {
	e.KC = parameter.KC
	e.LM = parameter.LM
	e.C = parameter.C
}

func (e *Evapotranspiration) SetState(state *Data.State) {
	e.WU = state.WU
	e.WL = state.WL
	e.P = state.P
	e.EM = state.EM
}

func (e *Evapotranspiration) UpdateState(state *Data.State) {
	state.EP = e.EP
	state.E = e.E
	state.EU = e.EU
	state.EL = e.EL
	state.ED = e.ED
}

func (e *Evapotranspiration) Calculate() {
	// 三层蒸散发计算
	e.EP = e.KC * e.EM // 计算流域蒸发能力

	if e.P+e.WU >= e.EP {
		e.EU = e.EP
		e.EL = 0
		e.ED = 0
	} else {
		e.EU = e.P + e.WU
		if e.WL >= e.C*e.LM {
			e.EL = (e.EP - e.EU) * e.WL / e.LM
			e.ED = 0
		} else {
			if e.WL >= e.C*(e.EP-e.EU) {
				e.EL = e.C * (e.EP - e.EU)
				e.ED = 0
			} else {
				e.EL = e.WL
				e.ED = e.C*(e.EP-e.EU) - e.EL
			}
		}
	}

	// 计算总的蒸散发量
	e.E = e.EU + e.EL + e.ED
}

func NewEvapotranspiration(kc, lm, c, wu, wl, ep, e, eu, el, ed, p, em float64) *Evapotranspiration {
	return &Evapotranspiration{
		KC: kc,
		LM: lm,
		C:  c,
		WU: wu,
		WL: wl,
		EP: ep,
		E:  e,
		EU: eu,
		EL: el,
		ED: ed,
		P:  p,
		EM: em,
	}
}

func (e *Evapotranspiration) Destroy() {
	// 析构函数
}
